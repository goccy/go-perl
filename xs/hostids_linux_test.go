//go:build linux

package xs

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

// A change of the process's ids the system refuses reports the system's
// error: an unprivileged process may not take another user's id (EPERM).
func TestHostIDsReportTheErrorOfTheCLibrary(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root may change its user")
	}
	other := os.Getuid() + 1
	if err := (libcIDs{}).Setuid(other); !errors.Is(err, syscall.EPERM) {
		t.Errorf("setuid(%d) as uid %d: %v, want EPERM", other, os.Getuid(), err)
	}
	if err := (libcIDs{}).Setreuid(other, other); !errors.Is(err, syscall.EPERM) {
		t.Errorf("setreuid(%d, %d): %v, want EPERM", other, other, err)
	}
	if err := (libcIDs{}).Setuid(os.Getuid()); err != nil {
		t.Errorf("setuid to its own uid: %v", err)
	}
}
