package internal

import "sync/atomic"

// servingInInit is set when the process runs a perl command line inside the
// Go program's package initialization (gperl's re-entry: a program started
// as the perl command, see gperl/reexec.go). Go holds a call into Go from a
// thread it did not start until the initialization ends, which there is the
// end of the program, so a native module's own thread that calls Perl would
// wait for ever; the native loader has the SDK tell such a module instead.
// Nothing else can tell: Go has no way to ask whether initialization ended.
var servingInInit atomic.Bool

// SetServingInInit records that the process runs a perl command line inside
// package initialization.
func SetServingInInit() { servingInInit.Store(true) }

// ServingInInit reports whether the process runs a perl command line inside
// package initialization.
func ServingInInit() bool { return servingInInit.Load() }
