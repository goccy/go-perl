package internal

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"testing"
	"unsafe"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// TestGoHeapIsAFailureForNativeCodeOnly pins that an instance whose memory
// is of the Go heap for a failure reports it only once native code holds
// addresses of the memory (here MarkMemoryBase), and never where the
// system has no mappings.
func TestGoHeapIsAFailureForNativeCodeOnly(t *testing.T) {
	instance := func() *Perl {
		g := wasm2go.NewWithMemory(nil, nil, nil, make([]byte, wasm2go.InitialMemoryBytes), wasm2go.InitialMemoryBytes)
		p := &Perl{m: &Module{g: g}}
		p.family = &forkFamily{origin: p}
		return p
	}
	reason := errors.New("no image")

	plain := instance()
	plain.onGoHeap(reason)
	if err := plain.family.end(); err != nil {
		t.Fatalf("an instance without native code reports %v", err)
	}

	native := instance()
	native.onGoHeap(reason)
	native.MarkMemoryBase()
	if err := native.family.end(); !errors.Is(err, reason) {
		t.Fatalf("an instance with native code reports %v, want %v", err, reason)
	}

	unsupported := instance()
	unsupported.onGoHeap(fmt.Errorf("no mappings: %w", errors.ErrUnsupported))
	unsupported.MarkMemoryBase()
	if err := unsupported.family.end(); err != nil {
		t.Fatalf("a system without mappings reports %v", err)
	}
}

// TestHeapForkCopiesMemoryAndGlobals pins the fork of an instance whose
// memory is of the Go heap: the child has the parent's memory as far as
// its size and the parent's globals, may grow to the ceiling, and neither
// sees the other's writes afterwards.
func TestHeapForkCopiesMemoryAndGlobals(t *testing.T) {
	const size = wasm2go.InitialMemoryBytes
	parent := wasm2go.NewWithMemory(nil, nil, nil, make([]byte, size+wasmPage), size)
	base.AccessMemory(parent, func(mem []byte) {
		copy(mem[100:], "parent's")
		copy(mem[size-8:], "the end!")
	})
	globals := base.SaveGlobals(parent)
	globals[0] += 4096
	base.RestoreGlobals(parent, globals)

	const ceiling = 1 << 30
	child := heapFork(parent, nil, &Module{}, ceiling)

	if got := child.MemSize.Load(); got != size {
		t.Fatalf("child's memory size: %d, want %d", got, size)
	}
	if child.MaxMem != ceiling {
		t.Fatalf("child's growth cap: %d, want %d", child.MaxMem, ceiling)
	}
	if got := base.SaveGlobals(child); !slices.Equal(got, globals) {
		t.Fatalf("child's globals: %v, want %v", got, globals)
	}
	base.AccessMemory(child, func(mem []byte) {
		if !bytes.Equal(mem[100:108], []byte("parent's")) || !bytes.Equal(mem[size-8:size], []byte("the end!")) {
			t.Fatal("the child's memory is not the parent's")
		}
		copy(mem[100:], "child's!")
	})
	base.AccessMemory(parent, func(mem []byte) {
		if !bytes.Equal(mem[100:108], []byte("parent's")) {
			t.Fatal("the child's write reached the parent")
		}
	})

	// The child has the parent's room: it grows in place as far as the
	// parent's memory does, under the pointers of native code.
	at := unsafe.SliceData(wasm2go.Memory(child))
	if prev := base.MemoryGrow64(child, 1); prev < 0 {
		t.Fatal("the child's memory could not grow by a page")
	}
	if unsafe.SliceData(wasm2go.Memory(child)) != at {
		t.Fatal("the child's memory moved when it grew within the parent's room")
	}
}
