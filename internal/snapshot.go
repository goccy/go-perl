package internal

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
// Built on the bundle's base.NewSharedSnapshot. Three values are baked into
// the image and so key the snapshot cache:
//   - the stdlib dir: perl_new(stdlibDir) writes it into @INC;
//   - the site dir: perl_new lays the site directories out below it, in
//     @INC and for %Config;
//   - the environment: wasi libc caches environ at startup, so %ENV reflects
//     what the snapshotted instance was booted with.
// Everything else in Options (FS backend, hooks, stdio, memory caps) lives on
// the host side of the WASI boundary and is freely per-instance.
//
// The snapshot is built for the SECOND instance of a key, not the first.
// Most processes start one interpreter (the gperl command, a program built
// with gperl build), and for that one building the snapshot first — start an
// interpreter, write its memory into an image, map the image again — costs
// as much as starting it. The first instance starts on memory of its own
// that keeps one address (base.AnonymousMemory); its first fork images its
// memory as any fork does.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"sync"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// snapshotCeiling is the linear-memory ceiling an instance built on the
// snapshot is mapped at, in bytes — the growth cap too (NewFromSnapshot sets
// MaxMem to the mapping length, and MemoryGrow refuses to grow past it before
// it could reallocate, which keeps the mmap and its sharing intact for the
// instance's whole life). It is as much as a wasm32 memory may grow to (the
// runtime's hard cap, just under 4 GiB): perl on the host has no limit of
// its own below the machine's memory, and a program's memory must not run
// out earlier here. Virtual and sparse: unused pages never become resident,
// and the runtime maps them without charging them against the commit limit
// until they are written. An instance whose options cap its memory lower is
// mapped at that cap instead (memoryCeiling).
var snapshotCeiling = int(min(base.WasmMemHardCap(), uint64(math.MaxInt)) / wasmPageBytes * wasmPageBytes)

// wasmPageBytes is the size of a page of wasm linear memory.
const wasmPageBytes = 65536

type perlSnapshot struct {
	img *base.SharedImage
	// image is img as the instances built on the snapshot hold it. The
	// snapshot holds it too, for the life of the process.
	image  *memoryImage
	handle uint64
	err    error
}

var (
	snapMu    sync.Mutex
	snapshots = map[string]*perlSnapshot{}
	// started holds the keys an instance has been made for without a
	// snapshot: the first instance of a key starts on its own memory.
	started = map[string]bool{}
)

// sharedSnapshot returns the process-wide snapshot for (stdlibDir, siteDir,
// env), building and caching it — from the second instance of the key on:
// for the first it returns nil, and that instance starts on memory of its
// own. A failure is remembered, not retried: New then initializes each
// instance privately, which is correct, only larger.
func sharedSnapshot(stdlibDir, siteDir string, env []string) *perlSnapshot {
	if os.Getenv("GO_PERL_NO_SHARED_IMAGE") != "" {
		return &perlSnapshot{err: fmt.Errorf("disabled by GO_PERL_NO_SHARED_IMAGE")}
	}
	key := snapshotKey(stdlibDir, siteDir, env)
	snapMu.Lock()
	defer snapMu.Unlock()
	if s, ok := snapshots[key]; ok {
		return s
	}
	if !started[key] {
		started[key] = true
		return nil
	}
	s := buildSnapshot(stdlibDir, siteDir, env)
	snapshots[key] = s
	return s
}

// snapshotKey names what a snapshot bakes in (see the top of this file).
func snapshotKey(stdlibDir, siteDir string, env []string) string {
	return fmt.Sprintf("lib=%q site=%q env=%q", stdlibDir, siteDir, env)
}

func buildSnapshot(stdlibDir, siteDir string, env []string) *perlSnapshot {
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
		mod.g = wasm2go.NewWithWASI(wasi, envStubs{m: mod}, wasmifyStubs{m: mod})
		wasm2go.Initialize(mod.g)
		_ = wasm2go.WasmInit(mod.g)

		tmp := &Perl{m: mod, wasi: wasi}
		h, err := tmp.perlNew(stdlibDir, siteDir)
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
		resp, _, err := tmp.EvalOp(context.Background(), "1 + 1;")
		if err != nil {
			return nil, fmt.Errorf("snapshot warmup: %w", err)
		}
		// The envelope's leading status byte is 0 on success (the public
		// package owns the full decoder; success is all the warmup needs).
		if len(resp) == 0 || resp[0] != 0 {
			return nil, fmt.Errorf("snapshot warmup died")
		}
		handle = h
		return mod.g, nil
	})
	if e := img.Err(); e != nil {
		return &perlSnapshot{err: e}
	}
	return &perlSnapshot{img: img, image: (&memoryImage{img: img}).retain(), handle: handle}
}
