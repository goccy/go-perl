//go:build linux && !(arm64 || amd64)

package ccall

import "github.com/ebitengine/purego"

// On the other architectures purego serves both directions: its callback
// takes eight integer arguments and hands them on as Args.

// newCallback returns the address of a C function that calls fn with its
// arguments, through purego, which makes more callbacks than MaxCallbacks.
func newCallback(_ int, fn func(args *Args) uintptr) uintptr {
	return purego.NewCallback(func(a0, a1, a2, a3, a4, a5, a6, a7 uintptr) uintptr {
		args := Args{a0, a1, a2, a3, a4, a5, a6, a7}
		return fn(&args)
	})
}

// invoke runs c's function. Of SyscallN's results only the first is the C
// function's: the second is a second return register, which a function
// returning one integer leaves unspecified, and the third is errno, which
// the C function sets only when its own interface says so and its caller
// reads from the C library.
func invoke(c *Call) {
	c.ret, _, _ = purego.SyscallN(c.fn, c.args[:]...)
}
