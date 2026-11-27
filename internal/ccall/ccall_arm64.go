//go:build (darwin || linux) && arm64

package ccall

// callbackEntrySize is the length of a trampoline: two instructions.
const callbackEntrySize = 8
