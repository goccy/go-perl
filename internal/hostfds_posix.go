//go:build darwin || linux

package internal

// The descriptors of the host process, for a program that is the host
// process (InstanceOptions.HostProcess): the runtime gives the guest the
// host's numbers (WASIConfig.HostDescriptors); this keeps the Go runtime's
// own descriptors off the low numbers and hands the program the
// descriptors the process was started with (InstanceOptions.
// InheritDescriptors).

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

// lowDescriptors is how many of the lowest descriptor numbers are kept
// free of the Go runtime's poller: see KeepRuntimeDescriptorsHigh.
const lowDescriptors = 64

// KeepRuntimeDescriptorsHigh has the Go runtime make the descriptors it
// needs for itself now, above the numbers a program's descriptors get. A
// program that is the host process has the host's descriptor numbers,
// lowest free first, and expects the first file it opens to be descriptor
// 3: the runtime's network poller, made when it is first needed, would take
// that number. It is done once, by the first instance whose program is the
// host process; a command built around such a program calls this while
// its packages are initialized as well: the earlier, the fewer descriptors
// the runtime has made already.
func KeepRuntimeDescriptorsHigh() {
	// The error is returned by that first instance's New, which asks for
	// the same result.
	_ = keepRuntimeHigh()
}

// keepRuntimeHigh is keepRuntimeDescriptorsHigh, done once.
var keepRuntimeHigh = sync.OnceValue(keepRuntimeDescriptorsHigh)

// keepRuntimeDescriptorsHigh fills the low numbers, has the runtime make
// its descriptors, and frees the numbers again.
func keepRuntimeDescriptorsHigh() (err error) {
	var plugs []int
	defer func() {
		for _, fd := range plugs {
			if cerr := syscall.Close(fd); cerr != nil {
				err = errors.Join(err, fmt.Errorf("close descriptor %d: %w", fd, cerr))
			}
		}
	}()
	// syscall.Dup makes a descriptor that is not close-on-exec: package
	// syscall's ForkLock keeps commands from starting until it is.
	syscall.ForkLock.RLock()
	root, err := syscall.Open("/", syscall.O_RDONLY|syscall.O_CLOEXEC, 0)
	if err != nil {
		syscall.ForkLock.RUnlock()
		return fmt.Errorf("open /: %w", err)
	}
	plugs = append(plugs, root)
	full := false
	for root < lowDescriptors {
		fd, err := syscall.Dup(root)
		if errors.Is(err, syscall.EMFILE) {
			// Every number the process may have (RLIMIT_NOFILE) is
			// taken, the low ones too: there is nothing more to plug.
			full = true
			break
		}
		if err != nil {
			syscall.ForkLock.RUnlock()
			return fmt.Errorf("dup: %w", err)
		}
		syscall.CloseOnExec(fd)
		plugs = append(plugs, fd)
		if fd >= lowDescriptors {
			break
		}
	}
	syscall.ForkLock.RUnlock()
	if full {
		// The limit is below lowDescriptors: the highest numbers it allows
		// go to the pipe and the poller (two for the pipe, which darwin's
		// signal pipe takes after it; the poller's queue, and on Linux its
		// eventfd), as high as they can be.
		const room = 4
		for range min(room, len(plugs)) {
			fd := plugs[len(plugs)-1]
			plugs = plugs[:len(plugs)-1]
			if err := syscall.Close(fd); err != nil {
				return fmt.Errorf("close descriptor %d: %w", fd, err)
			}
		}
	}
	// A pipe of package os is a descriptor the poller watches: making one
	// makes the poller.
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("make the network poller: %w", err)
	}
	if err := errors.Join(r.Close(), w.Close()); err != nil {
		return err
	}
	// On darwin the runtime hands signals to package os/signal through a
	// pipe it makes when a signal is first asked for: asking for one now
	// makes it, under the numbers the pipe above had. SIGWINCH is ignored
	// unless a program handles it, so asking for it and stopping changes
	// nothing else.
	asked := make(chan os.Signal, 1)
	signal.Notify(asked, syscall.SIGWINCH)
	signal.Stop(asked)
	return nil
}

var inheritedTaken sync.Once

// inheritedFiles returns the descriptors from inheritedFrom(opts) on that
// the process was started with, as files under their numbers, the first time it is
// called (InstanceOptions.InheritDescriptors: the first instance gets
// them); none afterwards. They are the ones open and not close-on-exec: the
// Go runtime makes every descriptor of its own close-on-exec. They are
// marked close-on-exec, so that a command the program starts does not get
// them besides the ones it is given. The ones below are left as they are.
func inheritedFiles(opts InstanceOptions) (map[int32]*os.File, error) {
	var (
		files map[int32]*os.File
		err   error
	)
	inheritedTaken.Do(func() { files, err = takeInheritedFiles(inheritedFrom(opts)) })
	return files, err
}

// inheritedFrom is the lowest descriptor the program gets of those the
// process was started with: 3 for a program that is the host process,
// whose descriptors are the host's; 4 for any other, whose descriptor 3 is
// its preopened root.
func inheritedFrom(opts InstanceOptions) int {
	if opts.HostProcess {
		return 3
	}
	return 4
}

// takeInheritedFiles is the body of inheritedFiles.
func takeInheritedFiles(lowest int) (map[int32]*os.File, error) {
	fds, err := openDescriptors()
	if err != nil {
		return nil, err
	}
	var files map[int32]*os.File
	for _, fd := range fds {
		if fd < lowest {
			continue
		}
		flags, err := descriptorFlags(fd)
		if errors.Is(err, syscall.EBADF) {
			// The listing's own descriptor, closed by now.
			continue
		}
		if err != nil {
			return nil, errors.Join(err, closeFiles(files))
		}
		if flags&syscall.FD_CLOEXEC != 0 {
			continue
		}
		syscall.CloseOnExec(fd)
		if files == nil {
			files = map[int32]*os.File{}
		}
		files[int32(fd)] = os.NewFile(uintptr(fd), "/dev/fd/"+strconv.Itoa(fd))
	}
	return files, nil
}

// closeFiles closes the files.
func closeFiles(files map[int32]*os.File) error {
	var errs []error
	for _, f := range files {
		errs = append(errs, f.Close())
	}
	return errors.Join(errs...)
}

// openDescriptors returns the numbers in /dev/fd: the descriptors of the
// process that are open, and the one the listing itself was read through,
// which is closed when it returns.
func openDescriptors() ([]int, error) {
	entries, err := os.ReadDir("/dev/fd")
	if err != nil {
		return nil, fmt.Errorf("list the descriptors of the process: %w", err)
	}
	fds := make([]int, 0, len(entries))
	for _, entry := range entries {
		fd, err := strconv.Atoi(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("list the descriptors of the process: %q in /dev/fd: %w", entry.Name(), err)
		}
		fds = append(fds, fd)
	}
	return fds, nil
}

// descriptorFlags returns the descriptor flags of fd (F_GETFD). Package
// syscall has no fcntl of its own.
func descriptorFlags(fd int) (int, error) {
	flags, _, errno := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), syscall.F_GETFD, 0)
	if errno != 0 {
		return 0, errno
	}
	return int(flags), nil
}
