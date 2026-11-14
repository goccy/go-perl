//go:build darwin || linux

package xs

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/goccy/go-perl/internal"
)

// Pointers of native code's own in the interpreter's arrays and hashes.
//
// perl lets a module keep a pointer to a structure of its own where an
// array or a hash keeps its values, as if it were a value (Object::Pad
// keeps the descriptions of its methods and fields that way, and never
// lets perl treat them as values). A slot of the interpreter's holds 32
// bits; a pointer of this process is wider. Such a pointer is kept in the
// slot as a number of a range at the top of the 32-bit space, which is no
// address of the interpreter's memory (its growth stops below the range)
// and none of this process's (the range is taken here, inaccessible, for
// as long as the process runs). The number foreignBase + 4*i stands for the
// i-th pointer kept for an instance: the SDK reads the pointers from the
// instance's table (goperl_foreign_t, C heap) and asks the loader for the
// number of a pointer it stores (foreignNumber).

// foreignSpan is the size of the range: as many pointers as a quarter of
// it, kept per instance.
const foreignSpan = 16 << 20

var (
	foreignOnce sync.Once
	foreignAt   uint64
)

// foreignBase is where the range starts, or 0 when it cannot be had in
// this process.
func foreignBase() uint64 {
	foreignOnce.Do(func() { foreignAt = reserveForeign() })
	return foreignAt
}

func reserveForeign() uint64 {
	const base = 1<<32 - foreignSpan
	// On macOS on arm64 the main executable's __PAGEZERO holds the first
	// 4 GiB of the address space: nothing of the process is ever there.
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return base
	}
	mmap, err := purego.Dlsym(purego.RTLD_DEFAULT, "mmap")
	if err != nil {
		return 0
	}
	munmap, err := purego.Dlsym(purego.RTLD_DEFAULT, "munmap")
	if err != nil {
		return 0
	}
	const protNone, mapPrivate = 0, 2
	mapAnon := uintptr(0x20)
	if runtime.GOOS == "darwin" {
		mapAnon = 0x1000
	}
	// the range where it was asked for, or else elsewhere when some of it
	// is in use
	got, _, _ := purego.SyscallN(mmap, base, foreignSpan, protNone, mapPrivate|mapAnon, ^uintptr(0), 0)
	if got == base {
		return base
	}
	if got != ^uintptr(0) {
		purego.SyscallN(munmap, got, foreignSpan)
	}
	return 0
}

// foreignTable is the C layout of goperl_foreign_t.
type foreignTable struct {
	base uint64
	n    uint64
	ptrs uintptr
}

// newForeignTable makes the table of an instance whose memory is p's,
// copying the pointers of from when it is given (the table of the instance
// p is a copy of). The memory of p stays below the range; when it is
// already bigger, the table has no range, and no pointer can be kept.
func newForeignTable(p *internal.Perl, from *state) (tab uintptr, idx map[uint64]uint32, capacity int) {
	t := foreignTable{base: foreignBase()}
	if t.base != 0 && !p.LimitMemory(t.base) {
		t.base = 0
	}
	idx = map[uint64]uint32{}
	if from != nil && from.foreignTab != 0 {
		var src foreignTable
		cMemcpy(uintptr(unsafe.Pointer(&src)), from.foreignTab, int(unsafe.Sizeof(src)))
		if t.base == src.base && src.n > 0 {
			t.n = src.n
			capacity = int(src.n)
			t.ptrs = cMalloc(capacity * 8)
			cMemcpy(t.ptrs, src.ptrs, capacity*8)
			for k, v := range from.foreignIdx {
				idx[k] = v
			}
		}
	}
	tab = cMalloc(int(unsafe.Sizeof(t)))
	cMemcpy(tab, uintptr(unsafe.Pointer(&t)), int(unsafe.Sizeof(t)))
	runtime.KeepAlive(&t)
	return tab, idx, capacity
}

// foreignNumber returns the number that stands for ptr in the instance's
// arrays and hashes, keeping ptr in the table when it is not there yet;
// 0 when it cannot be kept.
func (s *state) foreignNumber(ptr uint64) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	var t foreignTable
	cMemcpy(uintptr(unsafe.Pointer(&t)), s.foreignTab, int(unsafe.Sizeof(t)))
	if t.base == 0 {
		return 0
	}
	if i, ok := s.foreignIdx[ptr]; ok {
		return t.base + 4*uint64(i)
	}
	if t.n >= foreignSpan/4 {
		return 0
	}
	if t.n == uint64(s.foreignCap) {
		ncap := 2 * s.foreignCap
		if ncap == 0 {
			ncap = 64
		}
		grown := cMalloc(ncap * 8)
		cMemcpy(grown, t.ptrs, int(t.n)*8)
		old := t.ptrs
		t.ptrs = grown
		s.foreignCap = ncap
		pokeU64(s.foreignTab+unsafe.Offsetof(t.ptrs), uint64(grown))
		cFree(old)
	}
	pokeU64(t.ptrs+uintptr(t.n)*8, ptr)
	s.foreignIdx[ptr] = uint32(t.n)
	pokeU64(s.foreignTab+unsafe.Offsetof(t.n), t.n+1)
	return t.base + 4*t.n
}
