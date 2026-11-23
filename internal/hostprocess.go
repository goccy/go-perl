package internal

// The host process, for a program that is it (InstanceOptions.HostProcess).

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/goccy/perlwasm2go/base"
)

// HostIDs changes the user and group ids of the host process: the setuid
// and setgid families and setgroups, with the meaning of the C library's
// calls (-1 leaves an id as it is).
type HostIDs interface {
	Setuid(uid int) error
	Seteuid(euid int) error
	Setreuid(ruid, euid int) error
	Setresuid(ruid, euid, suid int) error
	Setgid(gid int) error
	Setegid(egid int) error
	Setregid(rgid, egid int) error
	Setresgid(rgid, egid, sgid int) error
	Setgroups(gids []int) error
}

// hostIDs is how a program that is the host process changes the process's
// user and group ids when a package linked into the binary knows a way the
// runtime does not have; see RegisterHostIDs. nil is the runtime's own way.
var hostIDs HostIDs

// RegisterHostIDs sets the way the ids of the host process are changed for
// every instance made afterwards whose program is the host process. The
// native module loader registers the C library's: a process that loads
// native code is linked with a C runtime, and has to change its ids through
// it. It is called from package initialization.
func RegisterHostIDs(ids HostIDs) { hostIDs = ids }

// hostProcess returns the Process of a program that is the host process:
// the runtime's, which reads the host's ids and accounting. On a system
// whose umask cannot be read without changing it, the program starts with
// the umask the host process had as this package was initialized
// (startUmask).
func hostProcess() (base.Process, error) {
	proc, err := base.HostProcess()
	if errors.Is(err, errors.ErrUnsupported) {
		return base.HostProcessUmask(startUmask), nil
	}
	return proc, err
}

// hostChanges is the ProcessChanges of a program that is the host process:
// the runtime's, with the id and group changes made through ids when a
// package registered a way (RegisterHostIDs).
type hostChanges struct {
	base.ProcessChanges
	ids HostIDs
}

// newHostChanges returns the ProcessChanges of a program that is the host
// process.
func newHostChanges() base.ProcessChanges {
	changes := base.HostProcessChanges()
	if hostIDs == nil {
		return changes
	}
	return hostChanges{ProcessChanges: changes, ids: hostIDs}
}

// SetIDs makes the change through the registered way.
func (c hostChanges) SetIDs(change base.IDChange) error {
	a, b, d := int(change.IDs[0]), int(change.IDs[1]), int(change.IDs[2])
	switch change.Call {
	case base.IDSetuid:
		return c.ids.Setuid(a)
	case base.IDSeteuid:
		return c.ids.Seteuid(a)
	case base.IDSetreuid:
		return c.ids.Setreuid(a, b)
	case base.IDSetresuid:
		return c.ids.Setresuid(a, b, d)
	case base.IDSetgid:
		return c.ids.Setgid(a)
	case base.IDSetegid:
		return c.ids.Setegid(a)
	case base.IDSetregid:
		return c.ids.Setregid(a, b)
	case base.IDSetresgid:
		return c.ids.Setresgid(a, b, d)
	}
	return fmt.Errorf("id change %d: %w", change.Call, errors.ErrUnsupported)
}

// SetGroups sets the groups through the registered way.
func (c hostChanges) SetGroups(groups []int32) error {
	gids := make([]int, len(groups))
	for i, g := range groups {
		gids[i] = int(g)
	}
	return c.ids.Setgroups(gids)
}

// hostSetenv makes a change of the guest's environment the host process's.
func hostSetenv(name, value string, unset bool) error {
	if unset {
		return os.Unsetenv(name)
	}
	return os.Setenv(name, value)
}

// hostChdirAt is the directory go-perl last made the host process's working
// directory (hostChdir), for when getcwd(2) cannot name the directory the
// process is in: darwin's fails in one the process may not search. nil when
// none is known.
var hostChdirAt struct {
	mu sync.Mutex
	at os.FileInfo
}

// hostChdir makes dir the host process's working directory, where it is
// already when it may not search dir (Internals::start_directory).
func hostChdir(dir string) error {
	if err := os.Chdir(dir); err != nil && !isHostDir(dir) {
		return err
	}
	at, err := os.Stat(dir)
	if err != nil {
		at = nil // not known: isHostDir then takes no directory for the host's
	}
	hostChdirAt.mu.Lock()
	defer hostChdirAt.mu.Unlock()
	hostChdirAt.at = at
	return nil
}

// hostCommandDir is the WASIConfig.CommandDir of the host process: "" for
// dir when it is the host process's working directory, which a command then
// inherits as a child of fork and exec does instead of changing into it
// (which fails for a directory the user may no longer search), dir
// otherwise, as without the hook.
func hostCommandDir(dir string) (string, error) {
	if isHostDir(dir) {
		return "", nil
	}
	return dir, nil
}

// isHostDir reports whether dir is known to be the host process's working
// directory: the one getcwd(2) names, or hostChdirAt where it names none.
// It is not known to be when it or that directory cannot be stat'ed.
func isHostDir(dir string) bool {
	at, err := os.Stat(dir)
	if err != nil {
		return false
	}
	var cwd os.FileInfo
	if wd, err := syscall.Getwd(); err == nil {
		if cwd, err = os.Stat(wd); err != nil {
			return false
		}
	} else {
		hostChdirAt.mu.Lock()
		cwd = hostChdirAt.at
		hostChdirAt.mu.Unlock()
		if cwd == nil {
			return false
		}
	}
	return os.SameFile(cwd, at)
}

// hostClock reads the host's CLOCK_MONOTONIC for a program that is the
// host process where a package linked into the binary knows a way the
// runtime does not have; see RegisterHostMonotonicClock. nil is the
// runtime's reading (base.HostMonotonicClock).
var hostClock func() uint64

// RegisterHostMonotonicClock sets the reader of the host's CLOCK_MONOTONIC,
// in nanoseconds, for every instance made afterwards whose program is the
// host process. The native module loader registers the C library's on
// darwin, where the clock the runtime reads (the Go runtime's) stands
// still while the system sleeps and CLOCK_MONOTONIC does not. It is called
// from package initialization.
func RegisterHostMonotonicClock(read func() uint64) { hostClock = read }

// hostMonotonicClock returns the MonotonicClock of a program that is the
// host process: the host's CLOCK_MONOTONIC.
func hostMonotonicClock() func() uint64 {
	if hostClock != nil {
		return hostClock
	}
	return base.HostMonotonicClock()
}
