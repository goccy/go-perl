package internal

import (
	"errors"
	"syscall"
	"unsafe"
)

// kernelSigaction is the action of a signal that code outside the Go
// runtime installed; this system's actions are not read (the structure the
// kernel gives back differs from the one it takes), so there are none.
type kernelSigaction struct{}

// hostSignalIgnored reports false: an action set outside the Go runtime is
// not read on this system. One the runtime keeps (an ignoring of SIGHUP or
// SIGINT the process inherited) signal.Ignored reports.
func hostSignalIgnored(sig syscall.Signal) (bool, error) { return false, nil }

// hostActionsOutside returns none (see kernelSigaction).
func hostActionsOutside() (map[syscall.Signal]kernelSigaction, error) { return nil, nil }

// nativeSignalAction reports false (see kernelSigaction).
func nativeSignalAction(sig syscall.Signal) (bool, error) { return false, nil }

// signalHostOfChild sends the host process the SIGCHLD of a child of the
// root whose status can be waited for, as the kernel sends the one of a
// child of its own: native code's handler of SIGCHLD (an event loop's child
// watcher) is not read here, and the guest has the signal as well.
func signalHostOfChild() error {
	return syscall.Kill(syscall.Getpid(), syscall.SIGCHLD)
}

// raiseOnThisThread is never called: no signal has a handler of native code
// here (nativeSignalAction).
func raiseOnThisThread(sig syscall.Signal) error { return errors.ErrUnsupported }

// restoreHostActionsOutside has none to put back (see kernelSigaction).
func restoreHostActionsOutside(outside map[syscall.Signal]kernelSigaction) error { return nil }

// giveSignalsToRuntime is never given a signal (see kernelSigaction).
func giveSignalsToRuntime(sigs []syscall.Signal) error { return nil }

// defaultSignalAction gives a signal its default action in the kernel,
// behind the Go runtime's back: sigaction with an action of zeros, which is
// SIG_DFL with no flags and an empty mask. Package syscall has the call's
// number and not the call.
func defaultSignalAction(sig syscall.Signal) error {
	var action [64]byte
	if _, _, errno := syscall.Syscall(syscall.SYS_SIGACTION, uintptr(sig),
		uintptr(unsafe.Pointer(&action[0])), 0); errno != 0 {
		return errno
	}
	return nil
}
