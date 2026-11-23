//go:build darwin || linux

package internal

// The host process's signals, for a program that is the host process
// (InstanceOptions.HostProcess): the process has one set of signal actions,
// which follow the dispositions of one program, the root (bindHostSignals).

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"

	"github.com/goccy/perlwasm2go/base"
)

// The guest's numbers (Linux numbering) of the signals the relay treats
// apart.
const (
	guestSIGINT   = 2
	guestSIGQUIT  = 3
	guestSIGKILL  = 9
	guestSIGPIPE  = 13
	guestSIGCHLD  = 17
	guestSIGSTOP  = 19
	guestSIGTSTP  = 20
	guestSIGURG   = 23
	guestSIGWINCH = 28
)

// guestSignalBit returns the mask bit of a guest signal number (signal n
// is bit n-1, as in the dispositions).
func guestSignalBit(sig int) uint64 { return uint64(1) << uint(sig-1) }

// uncatchableSignals can be neither caught nor ignored.
var uncatchableSignals = guestSignalBit(guestSIGKILL) | guestSignalBit(guestSIGSTOP)

// unrelayedSignals are never relayed to the guest: SIGCHLD the runtime
// posts itself when a process of the guest ends, and the host process's own
// reports either the same end or one the guest has no part in; SIGURG is the
// Go runtime's, which preempts goroutines with it.
var unrelayedSignals = uncatchableSignals | guestSignalBit(guestSIGCHLD) | guestSignalBit(guestSIGURG)

// defaultIgnoredSignals are the signals whose default action is to do
// nothing to the process (the stop signals among them: stopping has no
// meaning for a guest). The host process does not ignore them for a guest
// that does: ignoring SIGCHLD would have the kernel reap the commands the
// runtime waits for.
var defaultIgnoredSignals = guestSignalBit(17) | guestSignalBit(18) | guestSignalBit(19) |
	guestSignalBit(20) | guestSignalBit(21) | guestSignalBit(22) | guestSignalBit(23) |
	guestSignalBit(guestSIGWINCH)

// signalsHeldByRuntime are the signals whose default action ends the
// process, but which a Go program that has not asked for them takes
// otherwise: the Go runtime catches them and carries on, or, for SIGQUIT
// and SIGABRT, prints its goroutines and exits with status 2 instead of
// dying of the signal. They are relayed to a guest that leaves them at
// their default action, which then ends the guest by the signal.
var signalsHeldByRuntime = guestSignalBit(3) | guestSignalBit(6) | guestSignalBit(10) |
	guestSignalBit(12) | guestSignalBit(14) | guestSignalBit(24) | guestSignalBit(25) |
	guestSignalBit(26) | guestSignalBit(29)

// guestSignals maps the host's signals to the guest's numbers.
var guestSignals = func() map[syscall.Signal]int {
	m := make(map[syscall.Signal]int, len(hostSignals))
	for sig, hostSig := range hostSignals {
		m[hostSig] = sig
	}
	return m
}()

// signalReach is where a signal the host process receives goes once the
// guest has become a command (exec): the signals a terminal sends to the
// whole foreground process group reach the command by themselves, and go
// to the guest alone; any other, sent to the process's pid, is the
// command's as well.
func signalReach(sig int) base.SignalReach {
	switch sig {
	case guestSIGINT, guestSIGQUIT, guestSIGTSTP, guestSIGWINCH:
		return base.ReachGuest
	}
	return base.ReachProcess
}

// signalRoot is the host state of the root: the instance's WasiStubs.
type signalRoot interface {
	base.SignalTarget
}

// hostRelay is the process's relay.
var hostRelay = &signalRelay{}

// signalRelay makes the host process's signal actions follow the
// dispositions of its root's guest:
//
//   - a signal the guest handles is caught and delivered to the guest;
//   - a signal the guest ignores explicitly is ignored by the host process;
//   - a signal the Go runtime would swallow (signalsHeldByRuntime) is caught
//     and delivered as well when the guest leaves it at its default action,
//     which then ends the guest by that signal;
//   - any other signal keeps the action a Go program has for it.
//
// A handler that native code sharing the process installed (an event
// loop's signal watcher) is left in place for every signal the guest does
// not change (Linux). It is also the guest's Kill of a process of the host.
// mu guards everything; nothing of the runtime is called with it held,
// since the runtime calls Dispositions with its own locks held.
type signalRelay struct {
	mu sync.Mutex
	// root is the instance whose guest the host process's signals are
	// for; nil for none.
	root signalRoot
	// errs are the failures to change the host process's actions for the
	// root's reports, which unbindHostSignals returns.
	errs firstError
	// relay receives the signals of relayed. ignored are the signals the
	// host process ignores for the guest; adopted the ones the process
	// ignored by itself (inherited, or by native code), which stay ignored
	// whenever the guest does not handle them (adoptIgnoredSignals). chld is the guest's last
	// reported disposition of SIGCHLD, which is neither relayed nor ignored.
	relay            chan os.Signal
	relayed, ignored uint64
	adopted          uint64
	chld             childDisposition
}

// childDisposition is whether the guest handles SIGCHLD and whether it
// ignores it explicitly.
type childDisposition struct{ handled, ignored bool }

// Kill sends the guest's signal sig to the process pid of the host, or to
// the process group -pid.
func (r *signalRelay) Kill(pid, sig int32) error {
	if sig == 0 {
		return syscall.Kill(int(pid), 0)
	}
	hostSig, known := hostSignals[int(sig)]
	if !known {
		return syscall.EINVAL
	}
	return syscall.Kill(int(pid), hostSig)
}

// TakeSignal takes a signal the runtime raises for a guest itself (the
// interval timer's SIGALRM, a child's SIGCHLD, a kill of the guest's own
// pid, a raise, the SIGPIPE of a write that fails with EPIPE) as the host
// process's action for it has the signal: the last sigaction wins.
//
//   - One the process ignores by itself (inherited, or set by native code)
//     goes nowhere, for any instance (a fork child too). The ignorings the
//     relay makes for the root's guest are that guest's.
//   - One native code made a handler of its own for is raised in the host
//     process, for the root alone; the handler has run when it returns.
//   - Of any other, the guest has it. The root's child-end SIGCHLD is sent
//     the host process as well where native code's handlers are not read
//     (signalHostOfChild).
//
// The actions are read at every raise: native code changes them at any time.
func (r *signalRelay) TakeSignal(target base.SignalTarget, sig int32) (bool, error) {
	hostSig, known := hostSignals[int(sig)]
	if !known {
		return false, nil
	}
	if ignored, err := r.ignoredByProcess(sig); err != nil || ignored {
		return ignored, err
	}
	r.mu.Lock()
	root := r.root
	r.mu.Unlock()
	if root == nil || target != root {
		return false, nil
	}
	native, err := nativeSignalAction(hostSig)
	if err != nil {
		return false, err
	}
	if !native {
		if sig == guestSIGCHLD {
			return false, signalHostOfChild()
		}
		return false, nil
	}
	if err := raiseOnThisThread(hostSig); err != nil {
		return false, fmt.Errorf("raise %v for the handler of native code: %w", hostSig, err)
	}
	return true, nil
}

// SignalIgnored reports whether the host process ignores the guest's
// signal sig by itself (inherited, or set by native code), for any
// instance (a fork child too): the guest takes the action as its own.
func (r *signalRelay) SignalIgnored(target base.SignalTarget, sig int32) (bool, error) {
	return r.ignoredByProcess(sig)
}

// ignoredByProcess reports whether the host process ignores the guest's
// signal sig by itself (inherited, or set by native code). The ignorings
// the relay makes for the root's guest are that guest's own.
func (r *signalRelay) ignoredByProcess(sig int32) (bool, error) {
	hostSig, known := hostSignals[int(sig)]
	if !known {
		return false, nil
	}
	r.mu.Lock()
	relayIgnored := r.ignored
	r.mu.Unlock()
	if relayIgnored&guestSignalBit(int(sig)) != 0 {
		return false, nil
	}
	return processIgnores(hostSig)
}

// processIgnores reports whether the host process ignores the host signal:
// the Go runtime does (signal.Ignored), or the kernel's action is SIG_IGN,
// whoever set it.
func processIgnores(hostSig syscall.Signal) (bool, error) {
	if signal.Ignored(hostSig) {
		return true, nil
	}
	return hostSignalIgnored(hostSig)
}

// Dispositions makes the host process's actions follow the root's guest.
// Reports of other instances (a fork child, a second program that is the
// host process) are not the process's.
func (r *signalRelay) Dispositions(target base.SignalTarget, handled, ignored, blocked uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.root == nil || target != r.root {
		return
	}
	if err := r.syncLocked(handled, ignored); err != nil {
		r.errs.add(err)
	}
}

// syncLocked sets the host process's actions for the guest's dispositions;
// without a root, the actions of a Go program. Caller holds r.mu.
func (r *signalRelay) syncLocked(handled, ignored uint64) error {
	var relay, ignore uint64
	var chld childDisposition
	chldBit := guestSignalBit(guestSIGCHLD)
	if r.root != nil {
		// A signal the process ignored by itself (adopted) is relayed only
		// while the guest handles it.
		relay = (handled | signalsHeldByRuntime&^(ignored|r.adopted)) &^ unrelayedSignals
		ignore = ignored &^ (uncatchableSignals | defaultIgnoredSignals | r.adopted)
		chld = childDisposition{handled: handled&chldBit != 0, ignored: ignored&chldBit != 0}
	}

	var errs []error
	outside, err := hostActionsOutside()
	if err != nil {
		errs = append(errs, err)
	}
	// The handlers of native code for the signals the guest changes go:
	// the guest's change is the process's. SIGCHLD counts as changed when
	// the guest changes what it does with it, though it is neither relayed
	// nor ignored: the guest's waits are the program's.
	changed := (relay ^ r.relayed) | (ignore ^ r.ignored)
	if r.root != nil && chld != r.chld {
		changed |= chldBit
	}
	r.chld = chld
	taken := map[syscall.Signal]bool{}
	for hostSig := range outside {
		if sig, known := guestSignals[hostSig]; known && changed&guestSignalBit(sig) != 0 {
			delete(outside, hostSig)
			taken[hostSig] = true
		}
	}
	// The Go runtime's handler goes back where native code's goes, which
	// os/signal does not do for a signal it takes for its own already:
	// before signal.Ignore, which makes the runtime let go of the signal,
	// and after restoreHostSignal, which may give it its default action.
	give := func(sigs ...syscall.Signal) {
		var given []syscall.Signal
		for _, hostSig := range sigs {
			if taken[hostSig] {
				delete(taken, hostSig)
				given = append(given, hostSig)
			}
		}
		if err := giveSignalsToRuntime(given); err != nil {
			errs = append(errs, err)
		}
	}

	for sig, hostSig := range hostSignals {
		bit := guestSignalBit(sig)
		if (ignore^r.ignored)&bit == 0 {
			continue
		}
		if ignore&bit != 0 {
			give(hostSig)
			signal.Ignore(hostSig)
			continue
		}
		if err := restoreHostSignal(hostSig); err != nil {
			errs = append(errs, err)
		}
		give(hostSig)
	}
	r.ignored = ignore
	// The rest (a change of what is relayed, SIGCHLD) before os/signal is
	// asked for them.
	give(slices.Collect(maps.Keys(taken))...)

	if relay != r.relayed {
		// The new subscription is in place before the old one is dropped,
		// so a signal that stays handled is caught throughout.
		var relayTo chan os.Signal
		var sigs []os.Signal
		for sig, hostSig := range hostSignals {
			if relay&guestSignalBit(sig) != 0 {
				sigs = append(sigs, hostSig)
			}
		}
		if len(sigs) > 0 {
			// os/signal drops a signal that finds the channel full; forward
			// empties it as fast as DeliverSignal records a signal. As the
			// final runtime had it.
			relayTo = make(chan os.Signal, 64)
			signal.Notify(relayTo, sigs...)
			go r.forward(relayTo)
		}
		if r.relay != nil {
			signal.Stop(r.relay)
			close(r.relay)
		}
		// An adopted signal the guest no longer handles is ignored again,
		// whether the guest's action is now the default or ignore: it is
		// the action the guest's handler replaced, as perl's %SIG shows it
		// (IGNORE), which a local of it puts back.
		for sig, hostSig := range hostSignals {
			if (r.relayed&^relay)&r.adopted&guestSignalBit(sig) != 0 {
				signal.Ignore(hostSig)
			}
		}
		r.relay, r.relayed = relayTo, relay
		// An adopted signal the guest ignores again after handling it has
		// the Go runtime's handler since its subscription was dropped: the
		// process ignores it again.
		for sig, hostSig := range hostSignals {
			if guestSignalBit(sig)&r.adopted&ignored&^relay != 0 {
				signal.Ignore(hostSig)
			}
		}
	}

	// os/signal installs the Go runtime's handler again for a signal it
	// is asked for or stops being asked for.
	if err := restoreHostActionsOutside(outside); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// forward delivers the signals arriving on from to the root's guest, until
// from is closed. SIGPIPE is not delivered: the host process's SIGPIPE is
// the report of a write that failed with EPIPE, for which the runtime posts
// the guest its own at the write. It is still caught while the guest handles
// it, which keeps the Go runtime from ending the process over it.
func (r *signalRelay) forward(from <-chan os.Signal) {
	for received := range from {
		// os/signal gives a syscall.Signal on this system.
		r.deliver(received.(syscall.Signal))
	}
}

// deliver delivers the host signal to the root's guest (see forward).
func (r *signalRelay) deliver(hostSig syscall.Signal) {
	sig, known := guestSignals[hostSig]
	if !known || sig == guestSIGPIPE {
		return
	}
	r.mu.Lock()
	target := r.root
	r.mu.Unlock()
	if target != nil {
		target.DeliverSignal(int32(sig), signalReach(sig))
	}
}

// restoreHostSignal gives a signal the host process ignores (signal.Ignore)
// the action it has in a Go program that never touched it. Package signal
// has no way back from Ignore: Reset leaves the signal ignored. A
// subscription that is taken and dropped again puts the runtime's handler
// back, for every signal but SIGHUP and SIGINT, which the runtime leaves
// ignored when it finds them so: for those two the kernel is told first
// that the signal has its default action. No other signal goes through its
// default action on the way: for some it ends the process, which may get
// the signal at any moment.
func restoreHostSignal(sig syscall.Signal) error {
	if sig == syscall.SIGHUP || sig == syscall.SIGINT {
		if err := defaultSignalAction(sig); err != nil {
			return fmt.Errorf("give %v its default action: %w", sig, err)
		}
	}
	taken := make(chan os.Signal, 1)
	signal.Notify(taken, sig)
	signal.Stop(taken)
	return nil
}

var sigpipeForwardedToIgnore sync.Once

// forwardSIGPIPEToIgnore makes the Go runtime's action for a SIGPIPE it is
// not handling at that moment "ignore". The runtime forwards such a signal
// to the action the process had when the runtime took the signal over —
// the default action, which ends the process. signal.Ignore stops the
// handling before it has the kernel ignore the signal, so a SIGPIPE in
// between ends the process. A subscription taken after an Ignore records
// "ignore" as the action it replaces, and every Ignore after that forwards
// there. So it is done once, before the first program that is the host
// process runs, and SIGPIPE is given back the handling a Go program has.
func forwardSIGPIPEToIgnore() error {
	var err error
	sigpipeForwardedToIgnore.Do(func() {
		signal.Ignore(syscall.SIGPIPE)
		err = restoreHostSignal(syscall.SIGPIPE)
	})
	return err
}

// hostSignalsConfig is the Signals of a program that is the host process.
func hostSignalsConfig() base.Signals { return hostRelay }

// hostSignalTaker is the TakeSignal of a program that is the host process.
func hostSignalTaker() func(base.SignalTarget, int32) (bool, error) { return hostRelay.TakeSignal }

// hostSignalIgnoredFunc is the SignalIgnored of a program that is the host
// process.
func hostSignalIgnoredFunc() func(base.SignalTarget, int32) (bool, error) {
	return hostRelay.SignalIgnored
}

// bindHostSignals makes w, the host state of a program that is the host
// process, the root of the host process's signals when no other program
// is, before any code runs in it: the root is the first such program, for
// as long as it lives. The signals the process ignores by itself are left
// so (adoptIgnoredSignals).
func bindHostSignals(w signalRoot) error {
	if err := forwardSIGPIPEToIgnore(); err != nil {
		return err
	}
	hostRelay.mu.Lock()
	root := hostRelay.root == nil
	if root {
		hostRelay.root = w
	}
	hostRelay.mu.Unlock()
	if !root {
		return nil
	}
	return adoptIgnoredSignals()
}

// adoptIgnoredSignals records the signals the host process ignores by
// itself, inherited (a program started by nohup ignores SIGHUP) or set by
// native code (a network library ignores SIGPIPE), for the relay to keep
// ignored whenever the guest does not handle them, and tells the Go
// runtime of them. The guest's raises of them
// go nowhere (TakeSignal). It is done when the root is bound and after a
// native library is loaded.
func adoptIgnoredSignals() error {
	hostRelay.mu.Lock()
	target := hostRelay.root
	mine := hostRelay.ignored | hostRelay.adopted
	hostRelay.mu.Unlock()
	if target == nil {
		return nil
	}
	var adopted uint64
	var errs []error
	for sig, hostSig := range hostSignals {
		bit := guestSignalBit(sig)
		if bit&(mine|uncatchableSignals) != 0 {
			continue
		}
		ignored, err := processIgnores(hostSig)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if ignored {
			adopted |= bit
		}
	}
	// The root and what the relay ignores for it may have changed since
	// they were read.
	hostRelay.mu.Lock()
	if hostRelay.root != target {
		hostRelay.mu.Unlock()
		return errors.Join(errs...)
	}
	adopted &^= hostRelay.ignored | hostRelay.adopted
	hostRelay.adopted |= adopted
	hostRelay.mu.Unlock()
	for sig, hostSig := range hostSignals {
		if adopted&guestSignalBit(sig) == 0 {
			continue
		}
		// The kernel ignores it already; the Go runtime is told so as
		// well, which otherwise ends the process for a SIGPIPE of a write
		// to its standard output or error.
		signal.Ignore(hostSig)
	}
	return errors.Join(errs...)
}

// unbindHostSignals gives the host process's signals back, when w is the
// root: the actions are those of a Go program again, the relay ends, and a
// later program that is the host process can be the root. It returns the
// failures of the root's life too.
func unbindHostSignals(w signalRoot) error {
	hostRelay.mu.Lock()
	defer hostRelay.mu.Unlock()
	if hostRelay.root == nil || hostRelay.root != w {
		return nil
	}
	hostRelay.root = nil
	// An adopted signal the guest handles has the Go runtime's handler
	// once the subscription goes: the process ignores it again, as it did
	// before the guest.
	handledAdopted := hostRelay.adopted & hostRelay.relayed
	err := errors.Join(hostRelay.errs.take(), hostRelay.syncLocked(0, 0))
	for sig, hostSig := range hostSignals {
		if guestSignalBit(sig)&handledAdopted != 0 {
			signal.Ignore(hostSig)
		}
	}
	hostRelay.adopted = 0
	return err
}
