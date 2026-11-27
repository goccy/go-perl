package internal

import (
	"errors"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"unsafe"

	"github.com/goccy/perlwasm2go/base"
)

// setKernelAction sets the kernel's action of sig behind the Go runtime's
// back, as native code does, and puts the action back when the test ends.
func setKernelAction(t *testing.T, sig syscall.Signal, act kernelSigaction) {
	t.Helper()
	before, err := hostSignalAction(sig)
	if errors.Is(err, errors.ErrUnsupported) {
		t.Skipf("the kernel's sigaction is not read on %s", runtime.GOARCH)
	}
	if err != nil {
		t.Fatal(err)
	}
	set := func(act kernelSigaction) error {
		if _, _, errno := syscall.RawSyscall6(syscall.SYS_RT_SIGACTION, uintptr(sig),
			uintptr(unsafe.Pointer(&act)), 0, unsafe.Sizeof(act.mask), 0, 0); errno != 0 {
			return errno
		}
		return nil
	}
	if err := set(act); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := set(before); err != nil {
			t.Error(err)
		}
	})
}

// TestSignalRelayAdoptsAnIgnoringOfNativeCode pins that a signal native
// code set SIG_IGN for goes nowhere, for any instance, and that the Go
// runtime is told it is ignored.
func TestSignalRelayAdoptsAnIgnoringOfNativeCode(t *testing.T) {
	// Run last: the runtime handles SIGUSR2 again.
	t.Cleanup(func() {
		if err := restoreHostSignal(syscall.SIGUSR2); err != nil {
			t.Error(err)
		}
	})
	setKernelAction(t, syscall.SIGUSR2, kernelSigaction{handler: sigIGN})
	root := bindFake(t)
	for _, target := range []base.SignalTarget{root, newFakeRoot()} {
		if taken, err := hostRelay.TakeSignal(target, 12); !taken || err != nil {
			t.Fatalf("SIGUSR2 native code ignores: taken %v (%v), want taken", taken, err)
		}
	}
	if !signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("the Go runtime was not told that SIGUSR2 is ignored")
	}
}

// TestSignalRelayGivesSIGCHLDBackWhenTheGuestChangesIt pins that native
// code's handler of SIGCHLD stays until the guest changes what it does
// with SIGCHLD, which it then does through the runtime (its waits are the
// program's), without a subscription.
func TestSignalRelayGivesSIGCHLDBackWhenTheGuestChangesIt(t *testing.T) {
	own, err := hostSignalAction(syscall.SIGURG)
	if errors.Is(err, errors.ErrUnsupported) {
		t.Skipf("the kernel's sigaction is not read on %s", runtime.GOARCH)
	}
	if err != nil {
		t.Fatal(err)
	}
	// A handler that is not the runtime's; no child ends in this test.
	native := own
	native.handler = own.handler + 1
	setKernelAction(t, syscall.SIGCHLD, native)
	root := bindFake(t)

	hostRelay.Dispositions(root, guestSignalBit(10), 0, 0)
	if act, err := hostSignalAction(syscall.SIGCHLD); err != nil || act.handler != native.handler {
		t.Fatalf("SIGCHLD's handler after a change of SIGUSR1: %#x (%v), want native code's %#x", act.handler, err, native.handler)
	}
	hostRelay.Dispositions(root, guestSignalBit(10)|guestSignalBit(guestSIGCHLD), 0, 0)
	if act, err := hostSignalAction(syscall.SIGCHLD); err != nil || act.handler != own.handler {
		t.Fatalf("SIGCHLD's handler once the guest handles it: %#x (%v), want the runtime's %#x", act.handler, err, own.handler)
	}
	hostRelay.mu.Lock()
	relayed := hostRelay.relayed
	hostRelay.mu.Unlock()
	if relayed&guestSignalBit(guestSIGCHLD) != 0 || signal.Ignored(syscall.SIGCHLD) {
		t.Fatal("SIGCHLD is relayed or ignored")
	}
}

// TestSignalRelayKeepsHandlersOfNativeCode pins that a handler native code
// installed stays while the guest changes other signals, and goes when the
// guest changes its signal, which then reaches the guest.
func TestSignalRelayKeepsHandlersOfNativeCode(t *testing.T) {
	own, err := hostSignalAction(syscall.SIGURG)
	if errors.Is(err, errors.ErrUnsupported) {
		t.Skipf("the kernel's sigaction is not read on %s", runtime.GOARCH)
	}
	if err != nil {
		t.Fatal(err)
	}
	// A handler that is not the runtime's; the signal is never sent.
	native := own
	native.handler = own.handler + 1
	setKernelAction(t, syscall.SIGTERM, native)
	root := bindFake(t)

	hostRelay.Dispositions(root, guestSignalBit(10), 0, 0)
	if act, err := hostSignalAction(syscall.SIGTERM); err != nil || act.handler != native.handler {
		t.Fatalf("SIGTERM's handler after a change of SIGUSR1: %#x (%v), want native code's %#x", act.handler, err, native.handler)
	}
	hostRelay.Dispositions(root, guestSignalBit(10)|guestSignalBit(15), 0, 0)
	if act, err := hostSignalAction(syscall.SIGTERM); err != nil || act.handler != own.handler {
		t.Fatalf("SIGTERM's handler once the guest handles it: %#x (%v), want the runtime's %#x", act.handler, err, own.handler)
	}
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if got, want := <-root.delivered, (delivery{sig: 15, reach: base.ReachProcess}); got != want {
		t.Fatalf("delivered %+v, want %+v", got, want)
	}
}

// TestNativeSignalActionTellsNativeCodesHandler pins that a handler native
// code made the action of a signal is told from the Go runtime's, and that
// the relay raises no such signal for another instance than the root (the
// stand-in handler would run).
func TestNativeSignalActionTellsNativeCodesHandler(t *testing.T) {
	own, err := hostSignalAction(syscall.SIGURG)
	if errors.Is(err, errors.ErrUnsupported) {
		t.Skipf("the kernel's sigaction is not read on %s", runtime.GOARCH)
	}
	if err != nil {
		t.Fatal(err)
	}
	// A handler that is not the runtime's; the signal is never sent.
	native := own
	native.handler = own.handler + 1
	setKernelAction(t, syscall.SIGTERM, native)
	if got, err := nativeSignalAction(syscall.SIGTERM); !got || err != nil {
		t.Fatalf("SIGTERM with native code's handler: %v (%v), want native", got, err)
	}
	if got, err := nativeSignalAction(syscall.SIGUSR1); got || err != nil {
		t.Fatalf("SIGUSR1 with the runtime's handler: %v (%v), want not native", got, err)
	}
	bindFake(t)
	if taken, err := hostRelay.TakeSignal(newFakeRoot(), 15); taken || err != nil {
		t.Fatalf("SIGTERM of another instance: taken %v (%v), want not", taken, err)
	}
}

// TestSignalRelayKeepsSIGPIPEIgnoredAcrossALocal pins a SIGPIPE native code
// ignores across a program's local $SIG{PIPE}: taken (EPIPE alone) before,
// the guest's while its handler is the action, and taken again once the
// local ends, as perl puts IGNORE back; an explicit default keeps it
// ignored as well.
func TestSignalRelayKeepsSIGPIPEIgnoredAcrossALocal(t *testing.T) {
	setKernelAction(t, syscall.SIGPIPE, kernelSigaction{handler: sigIGN})
	root := bindFake(t)
	if taken, err := hostRelay.TakeSignal(root, 13); !taken || err != nil {
		t.Fatalf("SIGPIPE native code ignores: taken %v (%v), want taken", taken, err)
	}
	hostRelay.Dispositions(root, guestSignalBit(13), 0, 0)
	if taken, err := hostRelay.TakeSignal(root, 13); taken || err != nil {
		t.Fatalf("SIGPIPE the guest handles: taken %v (%v), want not", taken, err)
	}
	hostRelay.Dispositions(root, 0, 0, 0)
	if taken, err := hostRelay.TakeSignal(root, 13); !taken || err != nil {
		t.Fatalf("SIGPIPE after the guest's local ends: taken %v (%v), want taken", taken, err)
	}
	if act, err := hostSignalAction(syscall.SIGPIPE); err != nil || act.handler != sigIGN {
		t.Fatalf("SIGPIPE's action after the local: %#x (%v), want SIG_IGN", act.handler, err)
	}
	hostRelay.Dispositions(root, guestSignalBit(13), 0, 0)
	hostRelay.Dispositions(root, 0, guestSignalBit(13), 0)
	if taken, err := hostRelay.TakeSignal(newFakeRoot(), 13); !taken || err != nil {
		t.Fatalf("SIGPIPE after a local IGNORE: taken %v (%v), want taken", taken, err)
	}
}
