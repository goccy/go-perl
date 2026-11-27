package xs_test

import (
	"os"
	"runtime"
	"syscall"
)

// signalThisThread sends sig to the thread the caller runs on.
func signalThisThread(sig syscall.Signal) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return syscall.Tgkill(os.Getpid(), syscall.Gettid(), sig)
}
