//go:build !darwin && !linux

package internal

import "github.com/goccy/perlwasm2go/base"

// DieOfSignal ends the host process the way the guest's signal sig ends a
// process. This platform's signals are not mapped, so it returns and the
// caller falls back to an exit status.
func DieOfSignal(sig int) {}

// hostSignalsConfig is the Signals of a program that is the host process:
// none on this platform, whose signals are not mapped (the guest's are its
// own).
func hostSignalsConfig() base.Signals { return nil }

// hostSignalTaker is the TakeSignal of a program that is the host process:
// none on this platform, whose guest has every signal the runtime raises.
func hostSignalTaker() func(base.SignalTarget, int32) (bool, error) { return nil }

// hostSignalIgnoredFunc is the SignalIgnored of a program that is the host
// process: none on this platform, whose guest ignores no signal it did not
// set.
func hostSignalIgnoredFunc() func(base.SignalTarget, int32) (bool, error) { return nil }

// bindHostSignals has no signals of the host process to bind.
func bindHostSignals(w *base.WasiStubs) error { return nil }

// adoptIgnoredSignals has no signals of the host process to read.
func adoptIgnoredSignals() error { return nil }

// unbindHostSignals has no signals of the host process to give back.
func unbindHostSignals(w *base.WasiStubs) error { return nil }
