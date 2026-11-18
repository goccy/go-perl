//go:build (darwin || linux) && !race

package xs

import (
	"bytes"
	"unsafe"
)

// Foreign memory (the C heap, a native caller's stack, a loaded library's
// data) is read and written in place. A Go dereference of memory the Go
// heap does not hold is sound: the collector does not look at such
// addresses. Only checkptr, which a -race build turns on, rejects it, so a
// race build goes through the C library instead (foreignmem_race.go).

// foreignPtr is the address p as a pointer. (Through a variable: the
// conversion is of an address the program holds as a number, which is what
// it is, not of a pointer that went through uintptr arithmetic.)
func foreignPtr(p uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&p)) }

// The copies are not inlined: the compiler (go1.26's memmove inlining)
// fails on a copy between a slice and memory at an address it cannot
// type.

// foreignRead copies len(dst) bytes at src into dst.
//go:noinline
func foreignRead(dst []byte, src uintptr) {
	copy(dst, unsafe.Slice((*byte)(foreignPtr(src)), len(dst)))
}

// foreignWrite copies b to dst.
//go:noinline
func foreignWrite(dst uintptr, b []byte) {
	copy(unsafe.Slice((*byte)(foreignPtr(dst)), len(b)), b)
}

// foreignMove copies n bytes from src to dst, both foreign.
//go:noinline
func foreignMove(dst, src uintptr, n int) {
	copy(unsafe.Slice((*byte)(foreignPtr(dst)), n), unsafe.Slice((*byte)(foreignPtr(src)), n))
}

// foreignLoadU64 and foreignStoreU64 read and write a word.
func foreignLoadU64(src uintptr) uint64 { return *(*uint64)(foreignPtr(src)) }

func foreignStoreU64(dst uintptr, v uint64) { *(*uint64)(foreignPtr(dst)) = v }

// foreignStrlen is the length of the NUL-terminated string at p. It looks
// at one page at a time, up to the end of the page it is in: the bytes up
// to the NUL are there, and a page holding one of them is all there.
func foreignStrlen(p uintptr) int {
	const page = 4096
	n := 0
	for {
		chunk := unsafe.Slice((*byte)(foreignPtr(p)), int(page-p%page))
		if i := bytes.IndexByte(chunk, 0); i >= 0 {
			return n + i
		}
		n += len(chunk)
		p += uintptr(len(chunk))
	}
}

// frameAt is the frame of a native call in progress at addr, which C was
// given (pinFrame keeps it reachable until the call ends).
func frameAt(addr uintptr) *cFrame { return (*cFrame)(foreignPtr(addr)) }
