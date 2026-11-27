package internal

import (
	"context"
	"testing"
	"unsafe"

	goperlfs "github.com/goccy/go-perl/fs"
	wasm2go "github.com/goccy/perlwasm2go"
)

// The first instance of a key starts on memory of its own: nothing is
// imaged for it, and the snapshot is built when a second instance of the key
// is made, which starts from it. The first instance's memory keeps its
// address when it grows, as native code that holds pointers into it needs.
func TestFirstInstanceStartsWithoutTheSnapshot(t *testing.T) {
	// The host's filesystem has the /dev/null perl needs.
	opts := InstanceOptions{Env: []string{"GO_PERL_TEST_KEY=" + t.Name()}, FS: goperlfs.NewHostFS()}
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
	if first.mapped == nil {
		t.Fatal("the first instance is not on memory of its own")
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
	if second.mapped == nil || second.h != s.handle {
		t.Fatal("the second instance is not built on the snapshot")
	}
	resp, _, err = second.EvalOp(context.Background(), `1 + 2`)
	if err != nil || len(resp) == 0 || resp[0] != 0 {
		t.Fatalf("eval in the second instance: %v (%q)", err, resp)
	}
}
