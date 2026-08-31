// Package internal holds the machinery under the public go-perl API: the
// generated wasm2go binding (perl.go), the per-instance bridge transport,
// the copy-on-write snapshot/clone plumbing, and the native-XS hook surface
// the xs subpackage drives. Nothing here is application API — the public
// package wraps a *Perl from this package and exposes only what users
// should call.
package internal

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	goperlfs "github.com/goccy/go-perl/fs"
	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// Method IDs of the perl bridge service (service 0), in the order the proto
// declares them (alphabetical over the pl.h exports). ONE added export
// renumbers every later id — re-check this table against the regenerated
// perl.go on every pl.h change.
const (
	midArrayGet         = 0
	midArrayLen         = 1
	midArrayPush        = 2
	midArraySet         = 3
	midArrayValues      = 4
	midCall             = 5
	midClose            = 6
	midDeref            = 7
	midEval             = 8
	midHashDelete       = 9
	midHashGet          = 10
	midHashKeys         = 11
	midHashSet          = 12
	midInterruptAddr    = 13
	midInvoke           = 14
	midMethodCall       = 15
	midNew              = 16
	midNewArray         = 17
	midNewHash          = 18
	midRegisterNativeXS = 19
	midRelease          = 20
	midSetGoDispatcher  = 21
	midXSHelper         = 22
)

// InstanceOptions is the resolved per-instance configuration. The public package
// maps its Config here after applying defaults (the FS arrives non-nil and
// already carrying a working /dev/null).
type InstanceOptions struct {
	StdlibDir string
	Env       []string
	FS        goperlfs.FS
	Dial      func(network, host, ip string, port int) bool
	Resolve   func(host string) bool
	Exec      func(path string, argv []string) bool
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer

	MaxMemoryBytes     int
	MemoryReserveBytes int
}

// ExitError is a guest exit() the bridge caught cleanly: the guest unwound
// back to the call frame, so the interpreter is still flushable and
// destructible.
type ExitError struct{ Code int }

func (e *ExitError) Error() string { return fmt.Sprintf("perl: exit(%d)", e.Code) }

// Perl is one isolated Perl interpreter instance's machinery: the wasm
// module, the interpreter handle, the interrupt watchdog, the handle-release
// queue, and the callback dispatch tables.
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

	// UserHandler receives every callback whose method id is not one of the
	// reserved hook ids: the Perl->Go function bridge. The public package
	// sets it once at construction, before any dispatcher registration.
	UserHandler func(methodID int32, req []byte) ([]byte, error)

	// hookMu guards the reserved-id hook fields and dispatcherSet.
	hookMu        sync.RWMutex
	dispatcherSet bool
	// nativeXS handles native-XSUB dispatch (the reserved callback method
	// id); see native.go and the xs subpackage.
	nativeXS func(req []byte) []byte
	// magicFree handles teardown of host-side MAGIC mirrors when the guest
	// frees an anchored SV (reserved callback method id -2).
	magicFree func(id uint32)
	// ppHook handles pp-hook dispatch for op types a native module claimed
	// (reserved callback method id -3).
	ppHook func(req []byte) []byte
	// magicSet runs mirror svt_set hooks when the guest reports an
	// assignment to an SV whose anchor magic was upgraded to fire them.
	magicSet func(id uint32)
	// dtorFire handles save-stack destructors registered by native modules
	// (reserved callback method id -4).
	dtorFire func(id uint32)
	// perlioHook handles PerlIO layer-slot dispatch for layers a native
	// module registered (reserved callback method id -6).
	perlioHook func(req []byte) []byte
	// keywordHook handles keyword/infix plugin forwarding while the guest
	// parser compiles (reserved callback method id -7).
	keywordHook func(req []byte) []byte

	// closed flips when Close runs; every entry point checks it so a
	// straggler — including a handle release racing a Close — errors out
	// instead of calling into a destroyed (and possibly unmapped) instance.
	closed atomic.Bool
	// pendingRel collects handle ids whose host wrappers were collected by
	// the Go GC. Finalizers ONLY append here (a pure host-side operation —
	// invoking the guest from the finalizer goroutine could block behind a
	// long Eval or hit a closed instance); the queue drains through one
	// batched guest call at the start of the next Eval/Call.
	relMu      sync.Mutex
	pendingRel []uint64

	// opts is the resolved configuration the instance booted with; Clone
	// builds its workers' WASI stubs from it.
	opts InstanceOptions
	// cloneImg is the lazily built copy-on-write image Clone maps new
	// instances from (see clone.go).
	cloneImg cloneImage
	// cloneHooks run on every Clone (AddCloneHook).
	cloneHooks []func(clone *Perl) error
}

// Closed reports whether Close has run.
func (p *Perl) Closed() bool { return p.closed.Load() }

// QueueRelease records a handle whose host wrapper is gone. Safe from any
// goroutine, including the runtime's finalizer goroutine.
func (p *Perl) QueueRelease(id uint64) {
	p.relMu.Lock()
	p.pendingRel = append(p.pendingRel, id)
	p.relMu.Unlock()
}

// drainReleases releases every queued handle in one guest call. Runs on
// user-initiated entry points only. Best-effort: a failure re-queues nothing
// (the registry dies with the instance anyway).
func (p *Perl) drainReleases() {
	p.relMu.Lock()
	ids := p.pendingRel
	p.pendingRel = nil
	p.relMu.Unlock()
	if len(ids) == 0 || p.closed.Load() {
		return
	}
	packed := make([]byte, 0, len(ids)*8)
	for _, id := range ids {
		packed = binary.LittleEndian.AppendUint64(packed, id)
	}
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendBytes(buf, 2, packed)
	buf = pbAppendUint64(buf, 3, uint64(len(packed)))
	_, _ = p.m.invoke(0, midRelease, buf, wasm2go.Inv_0_20)
}

// New boots a fresh instance from opts, mapping the process-wide
// copy-on-write snapshot when possible (see snapshot.go).
func New(opts InstanceOptions) (p *Perl, err error) {
	m := &Module{}
	wasi := buildWASI(opts)
	p = &Perl{m: m, wasi: wasi, opts: opts}

	if err := p.boot(opts, wasi); err != nil {
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

// buildWASI assembles the per-instance WASI host from the options. Sandbox
// by default: no host environment, no host stdio.
func buildWASI(opts InstanceOptions) *base.WasiStubs {
	wasi := base.DefaultWASI()
	wasi.SetEnv(opts.Env)
	if opts.FS != nil {
		wasi.SetFS(opts.FS)
	}
	// The capability hooks are FAIL-CLOSED: the runtime allows when a hook
	// is unset, so a nil hook installs an explicit deny-all here — the zero
	// configuration grants nothing.
	dial := opts.Dial
	if dial == nil {
		dial = func(string, string, string, int) bool { return false }
	}
	wasi.SetDialHook(dial)
	resolve := opts.Resolve
	if resolve == nil {
		resolve = func(string) bool { return false }
	}
	wasi.SetResolveHook(resolve)
	exec := opts.Exec
	if exec == nil {
		exec = func(string, []string) bool { return false }
	}
	wasi.SetExecHook(exec)
	// Sandbox stdio by default: an unset stream does NOT fall through to the
	// host process stdio. Stdin defaults to empty (immediate EOF), stdout and
	// stderr to discard. (Perl-level print output is still captured into the
	// eval envelope by the bridge; these back the raw guest fds 0/1/2.)
	if opts.Stdin != nil {
		wasi.SetStdin(opts.Stdin)
	} else {
		wasi.SetStdin(bytes.NewReader(nil))
	}
	if opts.Stdout != nil {
		wasi.SetStdout(opts.Stdout)
	} else {
		wasi.SetStdout(io.Discard)
	}
	if opts.Stderr != nil {
		wasi.SetStderr(opts.Stderr)
	} else {
		wasi.SetStderr(io.Discard)
	}
	return wasi
}

// boot brings the instance's wasm module up and obtains the interpreter
// handle — via the copy-on-write snapshot when possible, privately otherwise.
func (p *Perl) boot(opts InstanceOptions, wasi *base.WasiStubs) error {
	if snap := sharedSnapshot(opts.StdlibDir, opts.Env); snap.err == nil {
		ceiling := snapshotCeiling
		if opts.MaxMemoryBytes > 0 {
			const wasmPage = 65536
			max := opts.MaxMemoryBytes / wasmPage * wasmPage
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
	return p.bootPrivate(opts, wasi)
}

// bootPrivate is the no-snapshot path: run the reactor initializer and
// perl_new on a private linear memory.
func (p *Perl) bootPrivate(opts InstanceOptions, wasi *base.WasiStubs) (err error) {
	if opts.MemoryReserveBytes > 0 {
		p.m.g = wasm2go.NewWithWASIReserve(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m}, opts.MemoryReserveBytes)
	} else {
		p.m.g = wasm2go.NewWithWASI(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m})
	}
	// Cap linear-memory growth so a runaway allocation fails in the guest
	// instead of growing the host process unbounded. Round down to a wasm
	// page; ignore values below the module's initial memory.
	if opts.MaxMemoryBytes > 0 {
		const wasmPage = 65536
		max := uint64(opts.MaxMemoryBytes) / wasmPage * wasmPage
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

	h, err := p.perlNew(opts.StdlibDir)
	if err != nil {
		return fmt.Errorf("perl_new: %w", err)
	}
	if h == 0 {
		return fmt.Errorf("perl_new returned 0 (interpreter init failed; check StdlibDir=%q)", opts.StdlibDir)
	}
	p.h = h
	return nil
}

// ErrClosed is returned by every entry point once Close has run.
var ErrClosed = fmt.Errorf("perl: instance is closed")

// invoker matches the generated per-method entry points (wasm2go.Inv_0_N).
type invoker = func(*base.Module, wptr, wptr) (int64, error)

// valueOp sends one bridge operation and returns the raw response envelope
// bytes (the typed-value protocol documented in perl-wasm's pl.h; the public
// package decodes them). interrupted reports whether the ctx watchdog fired
// while the guest ran — a die envelope then means the interruption, not a
// user-level die.
func (p *Perl) valueOp(ctx context.Context, mid int32, inv invoker, buf []byte) (resp []byte, interrupted bool, err error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if p.closed.Load() {
		return nil, false, ErrClosed
	}
	p.drainReleases()

	disarm := p.armInterrupt(ctx)
	out, invokeErr := p.m.invoke(0, mid, buf, inv)
	interrupted = disarm()

	if invokeErr != nil {
		return nil, interrupted, invokeErr
	}
	if e := pbExtractError(out); e != nil {
		return nil, interrupted, e
	}
	return readScalarAtField(out, 1, (*pbReader).readBytes), interrupted, nil
}

// EvalOp evaluates src (string eval, scalar context) and returns the raw
// eval envelope (result node + captured stdout/stderr).
func (p *Perl) EvalOp(ctx context.Context, src string) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendString(buf, 2, src)
	return p.valueOp(ctx, midEval, wasm2go.Inv_0_8, buf)
}

// CallOp invokes the named sub in list context with an encoded node list.
func (p *Perl) CallOp(ctx context.Context, name string, args []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendString(buf, 2, name)
	buf = pbAppendBytes(buf, 3, args)
	buf = pbAppendUint64(buf, 4, uint64(len(args)))
	return p.valueOp(ctx, midCall, wasm2go.Inv_0_5, buf)
}

// InvokeOp calls the CODE reference behind a handle; scalarCtx selects
// scalar calling context.
func (p *Perl) InvokeOp(ctx context.Context, code uint64, scalarCtx bool, args []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, code)
	if scalarCtx {
		buf = pbAppendUint64(buf, 3, 1)
	}
	buf = pbAppendBytes(buf, 4, args)
	buf = pbAppendUint64(buf, 5, uint64(len(args)))
	return p.valueOp(ctx, midInvoke, wasm2go.Inv_0_14, buf)
}

// MethodCallOp invokes $obj->method(args...) on the reference behind obj.
func (p *Perl) MethodCallOp(ctx context.Context, obj uint64, method string, args []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, obj)
	buf = pbAppendString(buf, 3, method)
	buf = pbAppendBytes(buf, 4, args)
	buf = pbAppendUint64(buf, 5, uint64(len(args)))
	return p.valueOp(ctx, midMethodCall, wasm2go.Inv_0_15, buf)
}

// DerefOp dereferences the SCALAR/REF reference behind ref ($$ref).
func (p *Perl) DerefOp(ctx context.Context, ref uint64) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, ref)
	return p.valueOp(ctx, midDeref, wasm2go.Inv_0_7, buf)
}

// ArrayLenOp returns the envelope carrying scalar @{$av}.
func (p *Perl) ArrayLenOp(ctx context.Context, av uint64) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, av)
	return p.valueOp(ctx, midArrayLen, wasm2go.Inv_0_1, buf)
}

// ArrayGetOp returns the envelope carrying $av->[idx].
func (p *Perl) ArrayGetOp(ctx context.Context, av uint64, idx int64) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, av)
	buf = pbAppendInt64(buf, 3, idx)
	return p.valueOp(ctx, midArrayGet, wasm2go.Inv_0_0, buf)
}

// ArraySetOp performs $av->[idx] = val (val: one encoded node).
func (p *Perl) ArraySetOp(ctx context.Context, av uint64, idx int64, val []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, av)
	buf = pbAppendInt64(buf, 3, idx)
	buf = pbAppendBytes(buf, 4, val)
	buf = pbAppendUint64(buf, 5, uint64(len(val)))
	return p.valueOp(ctx, midArraySet, wasm2go.Inv_0_3, buf)
}

// ArrayPushOp appends the encoded node list to @{$av}.
func (p *Perl) ArrayPushOp(ctx context.Context, av uint64, vals []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, av)
	buf = pbAppendBytes(buf, 3, vals)
	buf = pbAppendUint64(buf, 4, uint64(len(vals)))
	return p.valueOp(ctx, midArrayPush, wasm2go.Inv_0_2, buf)
}

// ArrayValuesOp returns the envelope carrying @{$av} as a node list.
func (p *Perl) ArrayValuesOp(ctx context.Context, av uint64) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, av)
	return p.valueOp(ctx, midArrayValues, wasm2go.Inv_0_4, buf)
}

// HashGetOp returns the envelope carrying (exists, $hv->{key}); key is one
// encoded string node.
func (p *Perl) HashGetOp(ctx context.Context, hv uint64, key []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, hv)
	buf = pbAppendBytes(buf, 3, key)
	buf = pbAppendUint64(buf, 4, uint64(len(key)))
	return p.valueOp(ctx, midHashGet, wasm2go.Inv_0_10, buf)
}

// HashSetOp performs $hv->{key} = val.
func (p *Perl) HashSetOp(ctx context.Context, hv uint64, key, val []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, hv)
	buf = pbAppendBytes(buf, 3, key)
	buf = pbAppendUint64(buf, 4, uint64(len(key)))
	buf = pbAppendBytes(buf, 5, val)
	buf = pbAppendUint64(buf, 6, uint64(len(val)))
	return p.valueOp(ctx, midHashSet, wasm2go.Inv_0_12, buf)
}

// HashDeleteOp performs delete $hv->{key}.
func (p *Perl) HashDeleteOp(ctx context.Context, hv uint64, key []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, hv)
	buf = pbAppendBytes(buf, 3, key)
	buf = pbAppendUint64(buf, 4, uint64(len(key)))
	return p.valueOp(ctx, midHashDelete, wasm2go.Inv_0_9, buf)
}

// HashKeysOp returns the envelope carrying keys %{$hv} as a node list.
func (p *Perl) HashKeysOp(ctx context.Context, hv uint64) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, hv)
	return p.valueOp(ctx, midHashKeys, wasm2go.Inv_0_11, buf)
}

// NewArrayOp materialises a fresh guest array from the node list and returns
// the envelope carrying its ref node.
func (p *Perl) NewArrayOp(ctx context.Context, vals []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendBytes(buf, 2, vals)
	buf = pbAppendUint64(buf, 3, uint64(len(vals)))
	return p.valueOp(ctx, midNewArray, wasm2go.Inv_0_17, buf)
}

// NewHashOp materialises a fresh guest hash from the alternating key/value
// node list and returns the envelope carrying its ref node.
func (p *Perl) NewHashOp(ctx context.Context, pairs []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendBytes(buf, 2, pairs)
	buf = pbAppendUint64(buf, 3, uint64(len(pairs)))
	return p.valueOp(ctx, midNewHash, wasm2go.Inv_0_18, buf)
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

// Close finalizes the instance. It must not be used afterward: every later
// Eval/Call errors, and outstanding handles become inert. Idempotent.
func (p *Perl) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil
	}
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	resp, err := p.m.invoke(0, midClose, buf, wasm2go.Inv_0_6)
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
	resp, err := p.m.invoke(0, midNew, buf, wasm2go.Inv_0_16)
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
	resp, err := p.m.invoke(0, midInterruptAddr, buf, wasm2go.Inv_0_13)
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

// WasiExitCode reports whether err is the guest terminating itself — a
// caught Perl-level exit() (*ExitError) or a raw wasi proc_exit — and
// returns the exit status.
func WasiExitCode(err error) (int, bool) {
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee.Code, true
	}
	var we *base.WasiExitError
	if errors.As(err, &we) {
		return int(we.Code), true
	}
	return 0, false
}
