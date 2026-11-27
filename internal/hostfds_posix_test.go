//go:build darwin || linux

package internal

import (
	"errors"
	"os"
	"os/exec"
	"slices"
	"syscall"
	"testing"
)

// TestInheritedFilesAreTheOnesNotCloseOnExec pins which descriptors are
// the program's: the ones from the lowest on that are not close-on-exec,
// which are close-on-exec afterwards.
func TestInheritedFilesAreTheOnesNotCloseOnExec(t *testing.T) {
	var pipe [2]int
	if err := syscall.Pipe(pipe[:]); err != nil {
		t.Fatal(err)
	}
	own, err := os.Open("/")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := own.Close(); err != nil {
			t.Error(err)
		}
	}()

	files, err := takeInheritedFiles(3)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := closeFiles(files); err != nil {
			t.Error(err)
		}
	}()
	for _, fd := range pipe {
		f, ok := files[int32(fd)]
		if !ok {
			t.Fatalf("descriptor %d of a pipe is not among %v", fd, files)
		}
		if int(f.Fd()) != fd {
			t.Fatalf("descriptor %d is given as the file of %d", fd, f.Fd())
		}
		flags, err := descriptorFlags(fd)
		if err != nil {
			t.Fatal(err)
		}
		if flags&syscall.FD_CLOEXEC == 0 {
			t.Fatalf("descriptor %d is not close-on-exec once taken", fd)
		}
	}
	if _, ok := files[int32(own.Fd())]; ok {
		t.Fatalf("a close-on-exec descriptor %d of the Go runtime is among them", own.Fd())
	}
	for fd := range files {
		if fd <= 2 {
			t.Fatalf("standard stream %d is among them", fd)
		}
	}
}

// TestInheritedFilesLeaveTheOnesBelowAlone pins that a descriptor below
// the lowest is neither taken nor changed, and that the lowest is 3 for a
// program that is the host process only: any other program's descriptor 3
// is its preopened root.
func TestInheritedFilesLeaveTheOnesBelowAlone(t *testing.T) {
	if got := inheritedFrom(InstanceOptions{}); got != 4 {
		t.Fatalf("an isolated program's inherited descriptors start at %d, want 4", got)
	}
	if got := inheritedFrom(InstanceOptions{HostProcess: true}); got != 3 {
		t.Fatalf("the host process's inherited descriptors start at %d, want 3", got)
	}

	var pipe [2]int
	if err := syscall.Pipe(pipe[:]); err != nil {
		t.Fatal(err)
	}
	below, lowest := min(pipe[0], pipe[1]), max(pipe[0], pipe[1])
	defer func() {
		if err := syscall.Close(below); err != nil {
			t.Error(err)
		}
	}()
	files, err := takeInheritedFiles(lowest)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := closeFiles(files); err != nil {
			t.Error(err)
		}
	}()
	if _, ok := files[int32(lowest)]; !ok {
		t.Fatalf("descriptor %d, the lowest, is not among %v", lowest, files)
	}
	if _, ok := files[int32(below)]; ok {
		t.Fatalf("descriptor %d below the lowest is among them", below)
	}
	flags, err := descriptorFlags(below)
	if err != nil {
		t.Fatal(err)
	}
	if flags&syscall.FD_CLOEXEC != 0 {
		t.Fatalf("descriptor %d below the lowest was made close-on-exec", below)
	}
}

// keepHighChild is set in the environment of the process
// TestKeepRuntimeDescriptorsHigh starts, which runs
// TestKeepRuntimeDescriptorsHighInAFreshProcess.
const keepHighChild = "GO_PERL_TEST_KEEP_RUNTIME_HIGH"

// TestKeepRuntimeDescriptorsHigh runs the tests of keeping the runtime's
// descriptors high in processes of their own, whose Go runtime has made no
// network poller yet: this one has.
func TestKeepRuntimeDescriptorsHigh(t *testing.T) {
	for _, name := range []string{
		"TestKeepRuntimeDescriptorsHighInAFreshProcess",
		"TestKeepRuntimeDescriptorsHighUnderASmallLimit",
	} {
		cmd := exec.Command(os.Args[0], "-test.run=^"+name+"$", "-test.v")
		cmd.Env = append(os.Environ(), keepHighChild+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v\n%s", name, err, out)
		}
	}
}

// TestKeepRuntimeDescriptorsHighUnderASmallLimit pins that a limit of
// descriptors below the low numbers does not fail it: the pipe and the
// poller get the highest numbers the limit allows, and the program then
// has descriptors to make.
func TestKeepRuntimeDescriptorsHighUnderASmallLimit(t *testing.T) {
	if os.Getenv(keepHighChild) == "" {
		t.Skip("run in a process of its own by TestKeepRuntimeDescriptorsHigh")
	}
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		t.Fatal(err)
	}
	limit.Cur = 40
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		t.Fatal(err)
	}
	if err := keepRuntimeHigh(); err != nil {
		t.Fatal(err)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := errors.Join(r.Close(), w.Close()); err != nil {
		t.Fatal(err)
	}
}

// lowOpen returns the descriptors below lowDescriptors that are open,
// asking each (F_GETFD): package os would make the network poller.
func lowOpen(t *testing.T) []int {
	t.Helper()
	var open []int
	for fd := range lowDescriptors {
		_, err := descriptorFlags(fd)
		if errors.Is(err, syscall.EBADF) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		open = append(open, fd)
	}
	return open
}

// TestKeepRuntimeDescriptorsHighInAFreshProcess pins that the network
// poller is made above the low numbers, which stay as they were, and that
// the first descriptor the program makes then gets the lowest free number.
func TestKeepRuntimeDescriptorsHighInAFreshProcess(t *testing.T) {
	if os.Getenv(keepHighChild) == "" {
		t.Skip("run in a process of its own by TestKeepRuntimeDescriptorsHigh")
	}
	before := lowOpen(t)
	if err := keepRuntimeHigh(); err != nil {
		t.Fatal(err)
	}
	if after := lowOpen(t); !slices.Equal(after, before) {
		t.Fatalf("open below %d: %v, before %v", lowDescriptors, after, before)
	}
	// A pipe of package os uses the poller, which exists by now.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := errors.Join(r.Close(), w.Close()); err != nil {
			t.Error(err)
		}
	}()
	lowestFree := 0
	for slices.Contains(before, lowestFree) {
		lowestFree++
	}
	if got := int(r.Fd()); got != lowestFree {
		t.Fatalf("the first descriptor the program makes is %d, want %d, the lowest free (open below %d: %v)", got, lowestFree, lowDescriptors, lowOpen(t))
	}
	// Nothing of the runtime's came with the pipe: the poller is above.
	want := append(slices.Clone(before), int(r.Fd()), int(w.Fd()))
	slices.Sort(want)
	if got := lowOpen(t); !slices.Equal(got, want) {
		t.Fatalf("open below %d after the program's pipe: %v, want %v", lowDescriptors, got, want)
	}
}
