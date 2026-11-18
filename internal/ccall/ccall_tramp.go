//go:build (darwin || linux) && (arm64 || amd64)

package ccall

import (
	"sync"
	"sync/atomic"
	"unsafe"

	// fakecgo, the runtime support of calls between Go and C when cgo is
	// off (crosscall2, the runtime's thread hooks), comes with purego.
	_ "github.com/ebitengine/purego"
)

// maxCallbacks is the number of trampolines ccall_<arch>.s provides.
const maxCallbacks = 64

// callbackFrame is what a trampoline hands to callbackWrap (through
// cgocallback): the trampoline's number, where it saved the call's
// arguments, and the result to return to C. The trampolines know its
// layout.
type callbackFrame struct {
	index  uintptr
	args   *Args
	result uintptr
}

var (
	callbacksMu sync.Mutex
	callbackN   int
	callbacks   [maxCallbacks]atomic.Pointer[func(*Args) uintptr]
)

// callbackWrap is the Go side of every trampoline: it calls the function
// registered for the trampoline with the call's arguments.
func callbackWrap(f *callbackFrame) {
	fn := callbacks[f.index].Load()
	if fn == nil {
		panic("ccall: call of a callback that was never registered")
	}
	f.result = (*fn)(f.args)
}

// callbackWrapCall is callbackWrap as a function value: the trampolines read
// its code pointer, which is the function's Go entry, from it.
var callbackWrapCall = callbackWrap

// callbackEntries is the address of the first trampoline (ccall_<arch>.s);
// trampoline i is callbackEntrySize*i bytes after it.
var callbackEntries = callbackTableAddr()

func callbackTableAddr() uintptr

// NewCallback returns the address of a C function that calls fn with its
// arguments (at most MaxArgs integers) and returns fn's result. A program
// has a fixed number of callbacks (64); registrations are permanent.
func NewCallback(fn func(args *Args) uintptr) uintptr {
	callbacksMu.Lock()
	defer callbacksMu.Unlock()
	if callbackN >= maxCallbacks {
		panic("ccall: no callback left")
	}
	i := callbackN
	callbackN++
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
