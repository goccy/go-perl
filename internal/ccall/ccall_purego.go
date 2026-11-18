//go:build (darwin || linux) && !(arm64 || amd64)

package ccall

import "github.com/ebitengine/purego"

// On the other architectures purego serves both directions: its callback
// takes eight integer arguments and hands them on as Args.

// NewCallback returns the address of a C function that calls fn with its
// arguments (at most MaxArgs integers) and returns fn's result.
func NewCallback(fn func(args *Args) uintptr) uintptr {
	return purego.NewCallback(func(a0, a1, a2, a3, a4, a5, a6, a7 uintptr) uintptr {
		args := Args{a0, a1, a2, a3, a4, a5, a6, a7}
		return fn(&args)
	})
}

// invoke runs c's function.
func invoke(c *Call) {
	c.ret, _, _ = purego.SyscallN(c.fn, c.args[:]...)
}
