package xs_test

import (
	"os"
	"syscall"
)

// signalThisThread sends sig to the process: this host gives a Go program
// no way to choose the thread.
func signalThisThread(sig syscall.Signal) error {
	return syscall.Kill(os.Getpid(), sig)
}
