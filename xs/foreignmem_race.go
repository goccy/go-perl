//go:build (darwin || linux) && race

package xs

import (
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Under -race, checkptr fatals on any Go dereference of memory it cannot
// attribute to a Go allocation, which foreign memory is. So the C library
// does the touching, driven through purego (memcpy to move bytes across,
// strlen to size a C string), and Go only reads its own buffers.

func foreignRead(dst []byte, src uintptr) {
	if len(dst) == 0 {
		return
	}
	purego.SyscallN(memcpyFn, uintptr(unsafe.Pointer(&dst[0])), src, uintptr(len(dst)))
	runtime.KeepAlive(dst)
}

func foreignWrite(dst uintptr, b []byte) {
	if len(b) == 0 {
		return
	}
	purego.SyscallN(memcpyFn, dst, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	runtime.KeepAlive(b)
}

func foreignMove(dst, src uintptr, n int) {
	if n > 0 {
		purego.SyscallN(memcpyFn, dst, src, uintptr(n))
	}
}

func foreignLoadU64(src uintptr) uint64 {
	var buf [8]byte
	foreignRead(buf[:], src)
	return binary.LittleEndian.Uint64(buf[:])
}

func foreignStoreU64(dst uintptr, v uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], v)
	foreignWrite(dst, buf[:])
}

func foreignStrlen(p uintptr) int {
	n, _, _ := purego.SyscallN(strlenFn, p)
	return int(n)
}

// frameAt is the frame pinned at addr: checkptr rejects a conversion of the
// address back into a pointer to the Go heap.
func frameAt(addr uintptr) *cFrame { return pinnedFrame(addr) }
