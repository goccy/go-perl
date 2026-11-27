package internal

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

// kernelSigaction is struct sigaction as the kernel's rt_sigaction takes it
// on amd64 and arm64.
type kernelSigaction struct {
	handler  uintptr
	flags    uint64
	restorer uintptr
	mask     uint64
}

// sigIGN is the handler of an ignored signal.
const sigIGN = 1

// hostSignalAction returns the kernel's action for the host signal:
// errors.ErrUnsupported where the layout of the kernel's structure is not
// known.
func hostSignalAction(sig syscall.Signal) (kernelSigaction, error) {
	var act kernelSigaction
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
		return act, errors.ErrUnsupported
	}
	if _, _, errno := syscall.RawSyscall6(syscall.SYS_RT_SIGACTION, uintptr(sig), 0,
		uintptr(unsafe.Pointer(&act)), unsafe.Sizeof(act.mask), 0, 0); errno != 0 {
		return act, fmt.Errorf("read the action of %v: %w", sig, errno)
	}
	return act, nil
}

// hostSignalIgnored reports whether the host process ignores the host signal
// (its action is SIG_IGN), whoever set it; false where the kernel cannot be
// asked.
func hostSignalIgnored(sig syscall.Signal) (bool, error) {
	act, err := hostSignalAction(sig)
	if errors.Is(err, errors.ErrUnsupported) {
		return false, nil
	}
	return err == nil && act.handler == sigIGN, err
}

// hostActionsOutside returns the actions of the host process's signals that
// code outside the Go runtime made its own: a handler of a native library
// (libev's, an rsignal of a module), installed after the Go runtime
// installed its own for every signal. They are told apart by the handler:
// the Go runtime has one for all the signals it takes, which is its handler
// of SIGURG (the signal it preempts goroutines with, always its own). None
// where the kernel cannot be asked.
func hostActionsOutside() (map[syscall.Signal]kernelSigaction, error) {
	own, err := runtimeSignalAction()
	if own == nil || err != nil {
		return nil, err
	}
	var outside map[syscall.Signal]kernelSigaction
	for _, hostSig := range hostSignals {
		if hostSig == syscall.SIGKILL || hostSig == syscall.SIGSTOP {
			continue
		}
		act, err := hostSignalAction(hostSig)
		if err != nil {
			return nil, err
		}
		if act.handler > sigIGN && act.handler != own.handler {
			if outside == nil {
				outside = map[syscall.Signal]kernelSigaction{}
			}
			outside[hostSig] = act
		}
	}
	return outside, nil
}

// runtimeSignalAction returns the Go runtime's action for the signals it
// takes, its action of SIGURG (see hostActionsOutside); nil where the
// kernel cannot be asked.
func runtimeSignalAction() (*kernelSigaction, error) {
	own, err := hostSignalAction(syscall.SIGURG)
	if errors.Is(err, errors.ErrUnsupported) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if own.handler <= sigIGN {
		return nil, fmt.Errorf("the Go runtime has no handler of %v", syscall.SIGURG)
	}
	return &own, nil
}

// nativeSignalAction reports whether code outside the Go runtime made a
// handler of its own the host process's action of the signal (see
// hostActionsOutside); false where the kernel cannot be asked.
func nativeSignalAction(sig syscall.Signal) (bool, error) {
	own, err := runtimeSignalAction()
	if own == nil || err != nil {
		return false, err
	}
	act, err := hostSignalAction(sig)
	if err != nil {
		return false, err
	}
	return act.handler > sigIGN && act.handler != own.handler, nil
}

// signalHostOfChild sends nothing: native code's handler of SIGCHLD is read
// here, and raised as any other (TakeSignal).
func signalHostOfChild() error { return nil }

// raiseOnThisThread sends the host process the signal on the calling
// thread, as perl's kill of its own pid has it delivered in a process of
// one thread: the handler has run when it returns. Sent to the process, any
// of the runtime's threads would take it, some time later.
func raiseOnThisThread(sig syscall.Signal) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return syscall.Tgkill(syscall.Getpid(), syscall.Gettid(), sig)
}

// restoreHostActionsOutside puts the actions hostActionsOutside returned
// back, after the Go runtime has installed its own handler for a signal
// again.
func restoreHostActionsOutside(outside map[syscall.Signal]kernelSigaction) error {
	var errs []error
	for sig, act := range outside {
		if _, _, errno := syscall.RawSyscall6(syscall.SYS_RT_SIGACTION, uintptr(sig),
			uintptr(unsafe.Pointer(&act)), 0, unsafe.Sizeof(act.mask), 0, 0); errno != 0 {
			errs = append(errs, fmt.Errorf("give %v back the handler of native code: %w", sig, errno))
		}
	}
	return errors.Join(errs...)
}

// giveSignalsToRuntime makes the Go runtime's handler the action of the
// signals again, over the handlers of native code. os/signal cannot: the
// runtime installs its handler only for a signal it does not handle
// already, and it handles every signal from its start, whatever was
// installed over its handler since. Its action is the one it has for
// SIGURG (see hostActionsOutside).
func giveSignalsToRuntime(sigs []syscall.Signal) error {
	if len(sigs) == 0 {
		return nil
	}
	own, err := hostSignalAction(syscall.SIGURG)
	if err != nil {
		return err
	}
	var errs []error
	for _, sig := range sigs {
		if _, _, errno := syscall.RawSyscall6(syscall.SYS_RT_SIGACTION, uintptr(sig),
			uintptr(unsafe.Pointer(&own)), 0, unsafe.Sizeof(own.mask), 0, 0); errno != 0 {
			errs = append(errs, fmt.Errorf("give %v back to the Go runtime: %w", sig, errno))
		}
	}
	return errors.Join(errs...)
}

// defaultSignalAction gives a signal its default action in the kernel,
// behind the Go runtime's back: rt_sigaction with an action of zeros, which
// is SIG_DFL with no flags and an empty mask whatever the layout of the
// structure. Package syscall has the call's number and not the call.
func defaultSignalAction(sig syscall.Signal) error {
	var action [64]byte
	const maskSize = 8 // the kernel's signal set
	if _, _, errno := syscall.RawSyscall6(syscall.SYS_RT_SIGACTION, uintptr(sig),
		uintptr(unsafe.Pointer(&action[0])), 0, maskSize, 0, 0); errno != 0 {
		return errno
	}
	return nil
}
