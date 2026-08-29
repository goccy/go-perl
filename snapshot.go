package perl

// Process-wide, copy-on-write snapshots of started Perl interpreters.
//
// Once perl is up, an instance's linear memory is several MB of which almost
// nothing is about THIS instance: the data segments and the heap that perl_new
// leaves behind are identical in every instance. So start ONE interpreter per
// process, all the way through perl_new (plus a warmup eval), image its memory
// and wasm globals, and give every later instance a private copy-on-write map
// of it. Instances re-run nothing and inherit a perl that is already
// initialized; they pay only for the pages they actually write.
//
// Same machinery go-spidermonkey uses (base.NewSharedSnapshot). Two values
// are baked into the image and so key the snapshot cache:
//   - the stdlib dir: perl_new(stdlibDir) writes it into @INC;
//   - the environment: wasi libc caches environ at startup, so %ENV reflects
//     what the snapshotted instance was booted with.
// Everything else in Config (FS backend, hooks, stdio, memory caps) lives on
// the host side of the WASI boundary and is freely per-instance.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// snapshotCeiling is the linear-memory ceiling an instance built on the
// snapshot is mapped at, in bytes — the growth cap too (NewFromSnapshot sets
// MaxMem to the mapping length, and MemoryGrow refuses to grow past it before
// it could reallocate, which keeps the mmap and its sharing intact for the
// instance's whole life). Virtual, sparse: unused pages never become resident.
const snapshotCeiling = 512 << 20

type perlSnapshot struct {
	img    *base.SharedImage
	handle uint64
	err    error
}

var (
	snapMu    sync.Mutex
	snapshots = map[string]*perlSnapshot{}
)

// sharedSnapshot returns the process-wide snapshot for (stdlibDir, env),
// building and caching it on first use. A failure is remembered, not retried:
// New then initializes each instance privately, which is correct, only larger.
func sharedSnapshot(stdlibDir string, env []string) *perlSnapshot {
	if os.Getenv("GO_PERL_NO_SHARED_IMAGE") != "" {
		return &perlSnapshot{err: fmt.Errorf("disabled by GO_PERL_NO_SHARED_IMAGE")}
	}
	key := fmt.Sprintf("lib=%q env=%q", stdlibDir, env)
	snapMu.Lock()
	defer snapMu.Unlock()
	if s, ok := snapshots[key]; ok {
		return s
	}
	s := buildSnapshot(stdlibDir, env)
	snapshots[key] = s
	return s
}

func buildSnapshot(stdlibDir string, env []string) *perlSnapshot {
	var handle uint64
	img := base.NewSharedSnapshot(func() (g *base.Module, err error) {
		defer func() {
			if r := recover(); r != nil {
				g, err = nil, fmt.Errorf("starting the interpreter to snapshot panicked: %v", r)
			}
		}()
		wasi := base.DefaultWASI()
		wasi.SetEnv(env)
		wasi.SetStdin(bytes.NewReader(nil))
		wasi.SetStdout(io.Discard)
		wasi.SetStderr(io.Discard)

		mod := &Module{}
		mod.g = wasm2go.NewWithWASI(wasi, envStubs{m: mod})
		wasm2go.Initialize(mod.g)
		_ = wasm2go.WasmInit(mod.g)

		tmp := &Perl{m: mod, wasi: wasi}
		h, err := tmp.perlNew(stdlibDir)
		if err != nil {
			return nil, fmt.Errorf("perl_new: %w", err)
		}
		if h == 0 {
			return nil, fmt.Errorf("perl_new returned 0")
		}
		tmp.h = h
		// Warm up before snapshotting: like CPython, perl defers work to the
		// first eval (loading the opcode tables, compiling the boot ops). Doing
		// it once here bakes that state into the shared image, so every
		// instance's own first eval re-does — and privately dirties — far less.
		// A bare expression leaves no package-level state; the isolation tests
		// guard that it stays so.
		if _, err := tmp.Eval(context.Background(), "1 + 1;"); err != nil {
			return nil, fmt.Errorf("snapshot warmup: %w", err)
		}
		handle = h
		return mod.g, nil
	})
	if e := img.Err(); e != nil {
		return &perlSnapshot{err: e}
	}
	return &perlSnapshot{img: img, handle: handle}
}
