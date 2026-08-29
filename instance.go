package perl

// Perl is a hand-written, multi-instance API layered on top of the generated
// single-global binding in perl.go. Each Perl owns its own wasm module
// (independent linear memory + WASI host) and one Perl runtime, so several
// instances run concurrently and in isolation. Construction goes through the
// process-wide copy-on-write snapshot when possible (see snapshot.go), so an
// instance boots by mapping an already-initialized interpreter image instead
// of re-running perl_new.

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// Method IDs of the perl bridge service (service 0), in the order the proto
// declares them (alphabetical over the pl.h exports): perl_call, perl_close,
// perl_eval, perl_interrupt_addr, perl_new, perl_register_native_xs,
// perl_set_go_dispatcher, perl_xs_helper.
const (
	midCall             = 0
	midClose            = 1
	midEval             = 2
	midInterruptAddr    = 3
	midNew              = 4
	midRegisterNativeXS = 5
	midSetGoDispatcher  = 6
	midXSHelper         = 7
)

// EvalResult is the decoded form of the JSON document perl_eval returns.
type EvalResult struct {
	// Ok is false when the eval died (a Perl-level die/croak, a compile
	// error, or a cancellation); Error then holds $@.
	Ok bool `json:"ok"`
	// Result is the stringified value of the evaluated expression (the value
	// of the last statement), valid only when Ok is true.
	Result string `json:"result"`
	// Stdout / Stderr capture what the eval printed to STDOUT / STDERR (the
	// bridge redirects both onto in-memory scalars for the duration).
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	// Error is $@ when Ok is false.
	Error string `json:"error"`
}

// Perl is one isolated Perl interpreter instance.
type Perl struct {
	m    *Module
	wasi *base.WasiStubs
	h    uint64
	// intrAddr is the linear-memory offset of the interpreter's interrupt-flag
	// word, resolved once at New so a context cancellation can fire without
	// taking the instance lock a running Eval holds.
	intrAddr uint32
	// mapped is the copy-on-write mapping this instance lives in when it was
	// built from the process snapshot; unmapped on Close. nil for instances
	// that booted privately.
	mapped []byte

	// Go-function bridge state (bridge.go). funcs maps the ids baked into the
	// bound Perl subs to their Go functions; dispatcherSet records whether the
	// per-instance callback handler has been registered with the guest.
	funcsMu       sync.RWMutex
	funcs         map[int32]GoFunc
	nextFuncID    int32
	dispatcherSet bool
	// nativeXS handles native-XSUB dispatch (the reserved callback method
	// id); see native.go and the xsnative subpackage.
	nativeXS func(req []byte) []byte
	// magicFree handles teardown of host-side MAGIC mirrors when the guest
	// frees an anchored SV (reserved callback method id -2).
	magicFree func(id uint32)
	// ppHook handles pp-hook dispatch for op types a native module claimed
	// (reserved callback method id -3).
	ppHook func(req []byte) []byte
	// dtorFire handles save-stack destructors registered by native modules
	// (reserved callback method id -4).
	dtorFire func(id uint32)

	// closed flips when Close runs; every public entry point checks it so a
	// straggler — including a *Ref Free racing a Close — errors out instead
	// of calling into a destroyed (and possibly unmapped) instance.
	closed atomic.Bool
	// pendingRel collects handle ids whose *Ref wrappers were collected by
	// the Go GC. Finalizers ONLY append here (a pure host-side operation —
	// invoking the guest from the finalizer goroutine could block behind a
	// long Eval or hit a closed instance); the queue drains through one
	// batched guest call at the start of the next Eval/Call.
	relMu      sync.Mutex
	pendingRel []uint64
}

// queueRelease records a handle whose Go wrapper is gone. Safe from any
// goroutine, including the runtime's finalizer goroutine.
func (p *Perl) queueRelease(id uint64) {
	p.relMu.Lock()
	p.pendingRel = append(p.pendingRel, id)
	p.relMu.Unlock()
}

// drainReleases releases every queued handle in one guest call. Runs on
// user-initiated entry points only. Best-effort: a failure re-queues nothing
// (the registry dies with the instance anyway).
func (p *Perl) drainReleases(ctx context.Context) {
	p.relMu.Lock()
	ids := p.pendingRel
	p.pendingRel = nil
	p.relMu.Unlock()
	if len(ids) == 0 || p.closed.Load() {
		return
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, _ = p.call(ctx, "__plwasm_release_all", args, false)
}

// New builds a fresh Perl instance, applies the sandbox config, and returns
// it ready to Eval. When the config permits (see snapshot.go) the instance is
// mapped copy-on-write from a process-wide snapshot of an already-initialized
// interpreter, so construction is cheap and the instances share the read-only
// bulk of their memory.
func New(cfg Config) (p *Perl, err error) {
	// Resolve the standard library location:
	//   - custom FS backend: the stdlib must live inside it; default the search
	//     path to its root ("/"). (Use NewStdlibMemFS to pre-load it.)
	//   - no FS and no StdlibDir: extract the embedded stdlib to a temp dir.
	if cfg.FS != nil {
		if cfg.StdlibDir == "" {
			cfg.StdlibDir = "/"
		}
	} else if cfg.StdlibDir == "" {
		dir, sErr := ExtractStdlib()
		if sErr != nil {
			return nil, fmt.Errorf("extract embedded stdlib: %w", sErr)
		}
		cfg.StdlibDir = dir
	}

	m := &Module{}
	wasi := buildWASI(cfg)
	p = &Perl{m: m, wasi: wasi}

	if err := p.boot(cfg, wasi); err != nil {
		return nil, err
	}

	// Resolve the interrupt-flag address up front so Eval's context watchdog
	// can fire without calling into a busy instance.
	addr, err := p.interruptAddr()
	if err != nil {
		p.releaseMapping()
		return nil, fmt.Errorf("resolve interrupt address: %w", err)
	}
	var addrErr error
	base.AccessMemory(m.g, func(mem []byte) {
		if uint64(addr)+4 > uint64(len(mem)) {
			addrErr = fmt.Errorf("interrupt address %#x out of range (memory is %d bytes)", addr, len(mem))
		}
	})
	if addrErr != nil {
		p.releaseMapping()
		return nil, addrErr
	}
	p.intrAddr = addr
	return p, nil
}

// buildWASI assembles the per-instance WASI host from the config. Sandbox by
// default: no host environment, no host stdio.
func buildWASI(cfg Config) *base.WasiStubs {
	wasi := base.DefaultWASI()
	wasi.SetEnv(cfg.Env)
	if cfg.FS != nil {
		// Perl cannot boot without a /dev/null (the -e bootstrap opens the
		// bit bucket as its script filehandle); synthesize one over any
		// custom backend.
		wasi.SetFS(withDevNull(cfg.FS))
	} else if cfg.PreopenDir != "" {
		wasi.SetPreopenDir(cfg.PreopenDir)
	}
	if cfg.FSAccess != nil {
		wasi.SetFSAccessHook(cfg.FSAccess)
	}
	if cfg.NetAccess != nil {
		wasi.SetNetAccessHook(cfg.NetAccess)
	}
	if cfg.Dial != nil {
		wasi.SetDialHook(cfg.Dial)
	}
	if cfg.Resolve != nil {
		wasi.SetResolveHook(cfg.Resolve)
	}
	if cfg.Exec != nil {
		wasi.SetExecHook(cfg.Exec)
	}
	// Sandbox stdio by default: an unset stream does NOT fall through to the
	// host process stdio. Stdin defaults to empty (immediate EOF), stdout and
	// stderr to discard. (Perl-level print output is still captured into
	// EvalResult by the bridge; these back the raw guest fds 0/1/2.)
	if cfg.Stdin != nil {
		wasi.SetStdin(cfg.Stdin)
	} else {
		wasi.SetStdin(bytes.NewReader(nil))
	}
	if cfg.Stdout != nil {
		wasi.SetStdout(cfg.Stdout)
	} else {
		wasi.SetStdout(io.Discard)
	}
	if cfg.Stderr != nil {
		wasi.SetStderr(cfg.Stderr)
	} else {
		wasi.SetStderr(io.Discard)
	}
	return wasi
}

// boot brings the instance's wasm module up and obtains the interpreter
// handle — via the copy-on-write snapshot when possible, privately otherwise.
func (p *Perl) boot(cfg Config, wasi *base.WasiStubs) error {
	if snap := sharedSnapshot(cfg.StdlibDir, cfg.Env); snap.err == nil {
		ceiling := snapshotCeiling
		if cfg.MaxMemoryBytes > 0 {
			const wasmPage = 65536
			max := cfg.MaxMemoryBytes / wasmPage * wasmPage
			if max < ceiling && uint64(max) >= snap.img.Size() {
				ceiling = max
			}
		}
		mem, err := snap.img.Memory(ceiling)
		if err == nil {
			p.m.g = wasm2go.NewFromSnapshot(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m}, mem, snap.img.Size(), snap.img.Globals())
			p.h = snap.handle
			p.mapped = mem
			return nil
		}
		// Mapping failed (exotic platform / fd exhaustion): boot privately.
	}
	return p.bootPrivate(cfg, wasi)
}

// bootPrivate is the no-snapshot path: run the reactor initializer and
// perl_new on a private linear memory.
func (p *Perl) bootPrivate(cfg Config, wasi *base.WasiStubs) (err error) {
	if cfg.MemoryReserveBytes > 0 {
		p.m.g = wasm2go.NewWithWASIReserve(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m}, cfg.MemoryReserveBytes)
	} else {
		p.m.g = wasm2go.NewWithWASI(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m})
	}
	// Cap linear-memory growth so a runaway allocation fails in the guest
	// instead of growing the host process unbounded. Round down to a wasm
	// page; ignore values below the module's initial memory.
	if cfg.MaxMemoryBytes > 0 {
		const wasmPage = 65536
		max := uint64(cfg.MaxMemoryBytes) / wasmPage * wasmPage
		if max >= uint64(len(wasm2go.Memory(p.m.g))) {
			wasm2go.SetMaxMemory(p.m.g, max)
		}
	}

	// Run the reactor _initialize + wasmify init under a recover so a C++
	// static-initializer trap surfaces as an error.
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("instance init panicked: %v", r)
			}
		}()
		wasm2go.Initialize(p.m.g)
		_ = wasm2go.WasmInit(p.m.g)
	}()
	if err != nil {
		return err
	}

	h, err := p.perlNew(cfg.StdlibDir)
	if err != nil {
		return fmt.Errorf("perl_new: %w", err)
	}
	if h == 0 {
		return fmt.Errorf("perl_new returned 0 (interpreter init failed; check StdlibDir=%q)", cfg.StdlibDir)
	}
	p.h = h
	return nil
}

// Eval compiles and runs src in this instance's persistent package (its
// main::) and returns the structured result. A Perl-level die is reported via
// EvalResult.Ok=false / .Error, not as a Go error; a Go error indicates a
// host/transport failure (a wasm trap, encoding problem, ...) or a context
// cancellation.
//
// Cancelling ctx while the eval runs stops it at the next Perl opcode (the
// interpreter runs a pluggable run loop that tests a host-writable flag; see
// perl-wasm's bridge) and Eval returns ctx.Err(). A single long-running
// opcode — a pathological regex, a big sort — is not preempted until it
// returns to the run loop.
func (p *Perl) Eval(ctx context.Context, src string) (EvalResult, error) {
	if err := ctx.Err(); err != nil {
		return EvalResult{}, err
	}
	if p.closed.Load() {
		return EvalResult{}, errClosed
	}
	p.drainReleases(ctx)
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendString(buf, 2, src)

	disarm := p.armInterrupt(ctx)
	resp, invokeErr := p.m.invoke(0, midEval, buf, wasm2go.Inv_0_2)
	interrupted := disarm()

	if invokeErr != nil {
		return EvalResult{}, invokeErr
	}
	if e := pbExtractError(resp); e != nil {
		return EvalResult{}, e
	}
	js := readScalarAtField(resp, 1, (*pbReader).readString)
	var r EvalResult
	if err := json.Unmarshal([]byte(js), &r); err != nil {
		return EvalResult{}, fmt.Errorf("decode eval result %q: %w", js, err)
	}
	if interrupted && !r.Ok {
		// We tripped the interrupt and the eval died: report the
		// cancellation, not the croak text it surfaced as.
		return EvalResult{}, ctx.Err()
	}
	return r, nil
}

// armInterrupt starts the cancellation watchdog for ctx: on ctx cancellation
// it trips the interrupt flag with a plain store into linear memory — no call
// into the (busy) instance. The returned disarm function MUST be called
// exactly once, after the guarded invoke returns; it reports whether the
// watchdog fired and, when it did, clears the flag (the eval may have
// finished before testing it, and a lingering flag would poison the next
// call). Receiving after close(stop) is deterministic: the watchdog sends
// exactly one value, and a true arrives strictly after its store.
func (p *Perl) armInterrupt(ctx context.Context) (disarm func() bool) {
	done := ctx.Done()
	if done == nil {
		return func() bool { return false }
	}
	stop := make(chan struct{})
	fired := make(chan bool, 1)
	go func() {
		select {
		case <-done:
			p.fireInterrupt()
			fired <- true
		case <-stop:
			fired <- false
		}
	}()
	return func() bool {
		close(stop)
		if <-fired {
			p.clearInterrupt()
			return true
		}
		return false
	}
}

// errClosed is returned by every entry point once Close has run.
var errClosed = fmt.Errorf("perl: instance is closed")

// Close finalizes the instance. The Perl must not be used afterward: every
// later Eval/Call errors, and outstanding *Ref handles become inert (their
// finalizers only touch host-side state). Idempotent.
func (p *Perl) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil
	}
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	resp, err := p.m.invoke(0, midClose, buf, wasm2go.Inv_0_1)
	if err == nil {
		err = pbExtractError(resp)
	}
	p.releaseMapping()
	return err
}

// releaseMapping returns the instance's copy-on-write mapping to the OS. Must
// only run when the module will never be entered again.
func (p *Perl) releaseMapping() {
	if p.mapped != nil {
		base.UnmapMemory(p.mapped)
		p.mapped = nil
	}
}

func (p *Perl) perlNew(stdlibDir string) (uint64, error) {
	var buf []byte
	buf = pbAppendString(buf, 1, stdlibDir)
	resp, err := p.m.invoke(0, midNew, buf, wasm2go.Inv_0_4)
	if err != nil {
		return 0, err
	}
	if e := pbExtractError(resp); e != nil {
		return 0, e
	}
	return readScalarAtField(resp, 1, (*pbReader).readUint64), nil
}

// interruptAddr fetches the address of this instance's interrupt-flag word in
// linear memory.
func (p *Perl) interruptAddr() (uint32, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	resp, err := p.m.invoke(0, midInterruptAddr, buf, wasm2go.Inv_0_3)
	if err != nil {
		return 0, err
	}
	if e := pbExtractError(resp); e != nil {
		return 0, e
	}
	return readScalarAtField(resp, 1, (*pbReader).readUint32), nil
}

// fireInterrupt stores 1 into the interrupt-flag word. It does not take the
// instance lock — that is the point: it runs concurrently with a busy Eval.
//
// The write happens inside base.AccessMemory, which holds the same lock the
// runtime's memory.grow takes to mutate the memory slice header or relocate
// its backing array, so the flag lands in the array the guest observes. The
// run loop reads the word with a plain single-word load on its own goroutine,
// which is exactly the pluggable PL_runops contract.
func (p *Perl) fireInterrupt() {
	base.AccessMemory(p.m.g, func(mem []byte) {
		binary.LittleEndian.PutUint32(mem[p.intrAddr:], 1)
	})
}

// clearInterrupt resets the interrupt-flag word. Called after an Eval whose
// watchdog fired: if the eval finished before the flag was tested, the guest
// never cleared it, and a lingering flag would abort the next Eval.
func (p *Perl) clearInterrupt() {
	base.AccessMemory(p.m.g, func(mem []byte) {
		binary.LittleEndian.PutUint32(mem[p.intrAddr:], 0)
	})
}
