package internal

import (
	"context"
	"testing"
	"unsafe"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// The first instance of a key starts on memory of its own: nothing is
// imaged for it, and the snapshot is built when a second instance of the key
// is made, which starts from it. The first instance's memory keeps its
// address when it grows, as native code that holds pointers into it needs.
func TestFirstInstanceStartsWithoutTheSnapshot(t *testing.T) {
	opts := InstanceOptions{Env: []string{"GO_PERL_TEST_KEY=" + t.Name()}}
	key := snapshotKey(opts.StdlibDir, opts.SiteDir, opts.Env)
	built := func() *perlSnapshot {
		snapMu.Lock()
		defer snapMu.Unlock()
		return snapshots[key]
	}

	first, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer first.Close()
	if s := built(); s != nil {
		t.Fatal("the snapshot was built for the first instance")
	}
	if first.mapped == nil || first.image == nil {
		t.Fatal("the first instance is not on memory of its own that a fork can image")
	}
	at := unsafe.SliceData(wasm2go.Memory(first.m.g))
	before := first.m.g.MemSize.Load()
	resp, _, err := first.EvalOp(context.Background(), `my @x = (1) x 3_000_000; scalar @x`)
	if err != nil || len(resp) == 0 || resp[0] != 0 {
		t.Fatalf("eval in the first instance: %v (%q)", err, resp)
	}
	if after := first.m.g.MemSize.Load(); after <= before {
		t.Fatalf("the memory did not grow (%d bytes, then %d)", before, after)
	}
	if unsafe.SliceData(wasm2go.Memory(first.m.g)) != at {
		t.Fatal("the first instance's memory moved when it grew")
	}

	second, err := New(opts)
	if err != nil {
		t.Fatalf("second New: %v", err)
	}
	defer second.Close()
	s := built()
	if s == nil || s.err != nil {
		t.Fatalf("no snapshot for the second instance (%v)", s)
	}
	if second.image != s.image {
		t.Fatal("the second instance is not built on the snapshot")
	}
	resp, _, err = second.EvalOp(context.Background(), `1 + 2`)
	if err != nil || len(resp) == 0 || resp[0] != 0 {
		t.Fatalf("eval in the second instance: %v (%q)", err, resp)
	}
}

// An instance that boots on private memory (no shared image to map) keeps
// one address for its memory up to its limit, also a limit above the room a
// reservation has by default (8 GiB): native code holds pointers into the
// memory. Before, the memory moved to a larger reservation when it grew
// past 8 GiB.
func TestPrivateMemoryKeepsItsAddressUpToItsLimit(t *testing.T) {
	if !base.MemReserved {
		t.Skip("linear memory is no reservation in this build")
	}
	t.Setenv("GO_PERL_NO_SHARED_IMAGE", "1")
	const limit = 10 << 30
	p, err := New(InstanceOptions{MaxMemoryBytes: limit})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	if p.mapped != nil {
		t.Fatal("the instance did not boot on private memory")
	}
	g := p.m.g
	at := unsafe.SliceData(wasm2go.Memory(g))
	want := uint64(9 << 30)
	if prev := base.MemoryGrow64(g, int64((want-g.MemSize.Load())/wasmPage)); prev < 0 {
		t.Fatal("the memory could not grow to 9 GiB")
	}
	if got := g.MemSize.Load(); got != want {
		t.Fatalf("memory size %d after the growth, want %d", got, want)
	}
	if unsafe.SliceData(wasm2go.Memory(g)) != at {
		t.Fatal("the memory moved when it grew past 8 GiB")
	}
	if prev := base.MemoryGrow64(g, int64((limit-want)/wasmPage)+1); prev >= 0 {
		t.Fatal("the memory grew past its limit")
	}
	resp, _, err := p.EvalOp(context.Background(), `my @x = (1) x 300_000; scalar @x`)
	if err != nil || len(resp) == 0 || resp[0] != 0 {
		t.Fatalf("eval after the growth: %v (%q)", err, resp)
	}
}
