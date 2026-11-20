//go:build darwin || linux

package gperl_test

import (
	"syscall"
	"testing"
)

// keepStandardDescriptors copies descriptors 0, 1 and 2 of the test process
// as they are now, and returns the function that puts them back. The
// copies are made at 1024 or above, where the runtime keeps descriptors of
// its own and a program numbers none of its. A descriptor that is closed
// now is left as the program leaves it.
func keepStandardDescriptors(tb testing.TB) (putBack func()) {
	tb.Helper()
	saved := [3]int{-1, -1, -1}
	for fd := range saved {
		copied, _, errno := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), syscall.F_DUPFD, 1024)
		if errno == syscall.EBADF {
			continue
		}
		if errno != 0 {
			tb.Fatalf("copy of descriptor %d: %v", fd, errno)
		}
		saved[fd] = int(copied)
		syscall.CloseOnExec(saved[fd])
	}
	return func() {
		for fd, copied := range saved {
			if copied < 0 {
				continue
			}
			if err := dupOnto(copied, fd); err != nil {
				tb.Errorf("descriptor %d put back: %v", fd, err)
			}
			_ = syscall.Close(copied)
		}
	}
}
