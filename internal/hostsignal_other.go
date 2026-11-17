//go:build !darwin && !linux

package internal

// DieOfSignal ends the host process the way the guest's signal sig ends a
// process. This platform's signals are not mapped, so it returns and the
// caller falls back to an exit status.
func DieOfSignal(sig int) {}

// hostChildEnded would send the host process the SIGCHLD of the end of a
// pseudo-process; this platform has no such signal.
func hostChildEnded() {}
