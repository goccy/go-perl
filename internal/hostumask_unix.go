//go:build unix

package internal

import (
	"os"
	"syscall"
)

// startUmask is the umask of the host process as this package is
// initialized. syscall.Umask sets a mask and returns the one before, so
// reading it changes it for a moment: at initialization no goroutine of the
// program creates files yet.
var startUmask = func() os.FileMode {
	mask := syscall.Umask(0o022)
	syscall.Umask(mask)
	return os.FileMode(mask)
}()
