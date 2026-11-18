//go:build (darwin || linux) && amd64

package ccall

// callbackEntrySize is the length of a trampoline: a CALL with a 32-bit
// displacement.
const callbackEntrySize = 5
