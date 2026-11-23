//go:build darwin || linux

package internal

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/goccy/perlwasm2go/base"
)

// delivery is a signal the relay delivered.
type delivery struct {
	sig   int32
	reach base.SignalReach
}

// fakeRoot records what the relay does to the root's host state.
type fakeRoot struct {
	delivered chan delivery
}

func newFakeRoot() *fakeRoot { return &fakeRoot{delivered: make(chan delivery, 8)} }

func (f *fakeRoot) PostSignal(sig int32) { f.delivered <- delivery{sig: sig, reach: -1} }
func (f *fakeRoot) DeliverSignal(sig int32, reach base.SignalReach) {
	f.delivered <- delivery{sig: sig, reach: reach}
}

// bindFake makes a fakeRoot the root for the test; the failures of its
// life come back from the unbinding.
func bindFake(t *testing.T) *fakeRoot {
	t.Helper()
	root := newFakeRoot()
	if err := bindHostSignals(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := unbindHostSignals(root); err != nil {
			t.Error(err)
		}
	})
	return root
}

// TestSignalRelayDeliversWhatTheGuestHandles pins that a signal the root's
// guest handles reaches it, with the reach of its kind: a terminal's
// signal the guest alone, any other the process.
func TestSignalRelayDeliversWhatTheGuestHandles(t *testing.T) {
	root := bindFake(t)
	hostRelay.Dispositions(root, guestSignalBit(10)|guestSignalBit(guestSIGINT), 0, 0)
	for _, tc := range []struct {
		hostSig syscall.Signal
		want    delivery
	}{
		{syscall.SIGUSR1, delivery{sig: 10, reach: base.ReachProcess}},
		{syscall.SIGINT, delivery{sig: guestSIGINT, reach: base.ReachGuest}},
	} {
		if err := syscall.Kill(os.Getpid(), tc.hostSig); err != nil {
			t.Fatal(err)
		}
		if got := <-root.delivered; got != tc.want {
			t.Fatalf("%v: delivered %+v, want %+v", tc.hostSig, got, tc.want)
		}
	}
}

// TestSignalRelayLeavesOtherInstancesAlone pins that only the root's
// reports set the host process's actions.
func TestSignalRelayLeavesOtherInstancesAlone(t *testing.T) {
	root := bindFake(t)
	other := newFakeRoot()
	hostRelay.Dispositions(other, 0, guestSignalBit(12), 0)
	if signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("a report of another instance made the host process ignore SIGUSR2")
	}
	hostRelay.Dispositions(root, 0, guestSignalBit(12), 0)
	if !signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("the root's report did not make the host process ignore SIGUSR2")
	}
	if err := unbindHostSignals(root); err != nil {
		t.Fatal(err)
	}
	if signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("SIGUSR2 is still ignored once the root is gone")
	}
	if hostRelay.relay != nil || hostRelay.relayed != 0 {
		t.Fatalf("the relay is still subscribed to %#x once the root is gone", hostRelay.relayed)
	}
}

// TestSignalRelayDoesNotIgnoreDefaultIgnoredSignals pins that a guest that
// ignores SIGCHLD does not have the host process ignore it (the kernel
// would reap the commands the runtime waits for).
func TestSignalRelayDoesNotIgnoreDefaultIgnoredSignals(t *testing.T) {
	root := bindFake(t)
	hostRelay.Dispositions(root, 0, guestSignalBit(guestSIGCHLD), 0)
	if signal.Ignored(syscall.SIGCHLD) {
		t.Fatal("the host process ignores SIGCHLD")
	}
}

// TestSignalRelayDropsSIGPIPE pins that the host process's SIGPIPE is not
// delivered (the runtime posts the guest's own at the failed write), and
// that the signals the relay does not know are not either.
func TestSignalRelayDropsSIGPIPE(t *testing.T) {
	root := bindFake(t)
	hostRelay.deliver(syscall.SIGPIPE)
	hostRelay.deliver(syscall.Signal(64))
	hostRelay.deliver(syscall.SIGUSR2)
	if got, want := <-root.delivered, (delivery{sig: 12, reach: base.ReachProcess}); got != want {
		t.Fatalf("delivered %+v, want %+v", got, want)
	}
}

// TestSignalRelayAdoptsWhatTheProcessIgnores pins that a signal the host
// process ignores by itself goes nowhere, for the root and for any other
// instance, and is left ignored when the root goes.
func TestSignalRelayAdoptsWhatTheProcessIgnores(t *testing.T) {
	signal.Ignore(syscall.SIGUSR2)
	t.Cleanup(func() {
		if err := restoreHostSignal(syscall.SIGUSR2); err != nil {
			t.Error(err)
		}
	})
	root := bindFake(t)
	// The guest's first report: SIGUSR2 at its default action, which the
	// relay would catch for a signal of the Go runtime's.
	hostRelay.Dispositions(root, 0, 0, 0)
	for _, target := range []base.SignalTarget{root, newFakeRoot()} {
		if taken, err := hostRelay.TakeSignal(target, 12); !taken || err != nil {
			t.Fatalf("SIGUSR2 the process ignores: taken %v (%v), want taken", taken, err)
		}
	}
	// Handled, and back to the default: ignored again.
	hostRelay.Dispositions(root, guestSignalBit(12), 0, 0)
	if taken, err := hostRelay.TakeSignal(root, 12); taken || err != nil {
		t.Fatalf("SIGUSR2 the guest handles: taken %v (%v), want not", taken, err)
	}
	hostRelay.Dispositions(root, 0, 0, 0)
	if !signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("SIGUSR2 is not ignored again once the guest's handler goes")
	}
	if err := unbindHostSignals(root); err != nil {
		t.Fatal(err)
	}
	if !signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("the process's own ignoring of SIGUSR2 was undone")
	}
}

// TestSignalRelayKeepsAnAdoptedIgnoringAfterTheGuestsHandler pins that a
// signal the process ignored by itself is ignored again when the guest,
// having handled it, ignores it again.
func TestSignalRelayKeepsAnAdoptedIgnoringAfterTheGuestsHandler(t *testing.T) {
	signal.Ignore(syscall.SIGUSR2)
	t.Cleanup(func() {
		if err := restoreHostSignal(syscall.SIGUSR2); err != nil {
			t.Error(err)
		}
	})
	root := bindFake(t)
	hostRelay.Dispositions(root, 0, guestSignalBit(12), 0)
	if !signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("SIGUSR2 is not ignored once adopted")
	}
	hostRelay.Dispositions(root, guestSignalBit(12), 0, 0)
	if signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("SIGUSR2 is ignored while the guest handles it")
	}
	hostRelay.Dispositions(root, 0, guestSignalBit(12), 0)
	if !signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("the guest ignores SIGUSR2 again, but the host process no longer ignores it")
	}
}

// TestSignalRelayLeavesAnAdoptedIgnoringWhenTheRootGoes pins that a signal
// the process ignored by itself is ignored again when the root goes while
// its guest handles the signal.
func TestSignalRelayLeavesAnAdoptedIgnoringWhenTheRootGoes(t *testing.T) {
	signal.Ignore(syscall.SIGUSR2)
	t.Cleanup(func() {
		if err := restoreHostSignal(syscall.SIGUSR2); err != nil {
			t.Error(err)
		}
	})
	root := bindFake(t)
	hostRelay.Dispositions(root, guestSignalBit(12), 0, 0)
	if signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("SIGUSR2 is ignored while the guest handles it")
	}
	if err := unbindHostSignals(root); err != nil {
		t.Fatal(err)
	}
	if !signal.Ignored(syscall.SIGUSR2) {
		t.Fatal("the process no longer ignores SIGUSR2 once the root has gone")
	}
}

// TestSignalRelayKill pins the guest's kill of a process of the host.
func TestSignalRelayKill(t *testing.T) {
	if err := hostRelay.Kill(int32(os.Getpid()), 0); err != nil {
		t.Fatalf("probe of the host process: %v", err)
	}
	if err := hostRelay.Kill(int32(os.Getpid()), 99); !errors.Is(err, syscall.EINVAL) {
		t.Fatalf("kill with no such signal: %v, want EINVAL", err)
	}
}

// TestSignalRelayTakesWhatTheProcessIgnoresAtTheRaise pins that an
// ignoring is read at the raise, and that the relay's own ignoring for the
// root's guest is that guest's: another instance has the signal, and the
// guest's change back to the default action gives the root the signal too.
func TestSignalRelayTakesWhatTheProcessIgnoresAtTheRaise(t *testing.T) {
	t.Cleanup(func() {
		if err := restoreHostSignal(syscall.SIGUSR2); err != nil {
			t.Error(err)
		}
	})
	root := bindFake(t)
	other := newFakeRoot()
	signal.Ignore(syscall.SIGUSR2)
	if taken, err := hostRelay.TakeSignal(other, 12); !taken || err != nil {
		t.Fatalf("SIGUSR2 ignored after the root was bound: taken %v (%v), want taken", taken, err)
	}
	hostRelay.Dispositions(root, 0, guestSignalBit(12), 0)
	if taken, err := hostRelay.TakeSignal(other, 12); taken || err != nil {
		t.Fatalf("SIGUSR2 the relay ignores for the root's guest: taken %v (%v), want not", taken, err)
	}
	hostRelay.Dispositions(root, 0, 0, 0)
	if taken, err := hostRelay.TakeSignal(root, 12); taken || err != nil {
		t.Fatalf("SIGUSR2 back at its default action: taken %v (%v), want not", taken, err)
	}
}

// TestSignalRelayTakesNoSignalOfTheGuests pins that the relay leaves to the
// guest a signal whose action is the Go runtime's, one for another instance
// than the root, one the host has no number for, and any without a root.
func TestSignalRelayTakesNoSignalOfTheGuests(t *testing.T) {
	other := newFakeRoot()
	if taken, err := hostRelay.TakeSignal(other, 14); taken || err != nil {
		t.Fatalf("without a root: taken %v (%v), want not", taken, err)
	}
	root := bindFake(t)
	for _, tc := range []struct {
		name   string
		target base.SignalTarget
		sig    int32
	}{
		{"the runtime's action", root, 14},
		{"another instance", other, 14},
		{"no signal of the host", root, 16},
	} {
		if taken, err := hostRelay.TakeSignal(tc.target, tc.sig); taken || err != nil {
			t.Errorf("%s: taken %v (%v), want not", tc.name, taken, err)
		}
	}
}
