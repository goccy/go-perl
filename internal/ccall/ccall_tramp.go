//go:build (darwin || linux) && (arm64 || amd64)

package ccall

import (
	"sync/atomic"
	"unsafe"

	// fakecgo, the runtime support of calls between Go and C when cgo is
	// off (crosscall2, the runtime's thread hooks), comes with purego.
	_ "github.com/ebitengine/purego"
)

// callbackFrame is what a trampoline hands to callbackWrap (through
// cgocallback): the trampoline's number, where it saved the call's
// arguments, and the result to return to C. The trampolines know its
// layout.
type callbackFrame struct {
	index  uintptr
	args   *Args
	result uintptr
}

// callbacks holds the function of each trampoline (ccall_<arch>.s has
// MaxCallbacks of them).
var callbacks [MaxCallbacks]atomic.Pointer[func(*Args) uintptr]

// callbackWrap is the Go side of every trampoline: it calls the function
// registered for the trampoline with the call's arguments.
func callbackWrap(f *callbackFrame) {
	// newCallback stores the function before the trampoline's address is
	// handed out, so a trampoline C can call has one.
	f.result = (*callbacks[f.index].Load())(f.args)
}

// callbackWrapCall is callbackWrap as a function value: the trampolines read
// its code pointer, which is the function's Go entry, from it.
var callbackWrapCall = callbackWrap

// callbackEntries is the address of the first trampoline (ccall_<arch>.s);
// trampoline i is callbackEntrySize*i bytes after it.
var callbackEntries = callbackTableAddr()

func callbackTableAddr() uintptr

// newCallback makes trampoline i call fn and returns its address.
func newCallback(i int, fn func(args *Args) uintptr) uintptr {
	callbacks[i].Store(&fn)
	return callbackEntries + uintptr(i)*callbackEntrySize
}

// callC is the address of the trampoline (ccall_<arch>.s) runtime.cgocall
// runs on the system stack with a *Call: it calls the Call's function with
// its arguments and stores the result.
var callC = callCAddr()

func callCAddr() uintptr

//go:linkname runtime_cgocall runtime.cgocall
func runtime_cgocall(fn uintptr, arg unsafe.Pointer) int32

// invoke runs c's function on the system stack.
func invoke(c *Call) {
	runtime_cgocall(callC, unsafe.Pointer(c))
}
