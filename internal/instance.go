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
	"encoding/json"
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

// InstanceOptions is the resolved per-instance configuration. The public package
// maps its Config here after applying defaults (the FS arrives non-nil and
// already carrying a working /dev/null).
type InstanceOptions struct {
	StdlibDir string
	Env       []string
	FS        goperlfs.FS
	FSAccess  func(path string, write bool) bool
	NetAccess func(op string) bool
	Dial      func(network, host, ip string, port int) bool
	Resolve   func(host string) bool
	Exec      func(path string, argv []string) bool
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer

	MaxMemoryBytes     int
	MemoryReserveBytes int
}

// WireNode is one tagged value on the bridge (JSON is only the carrier; the
// tag is the semantics):
//
//	k="u"                 undef / nil
//	k="d" v=<scalar>      plain scalar by value
//	k="j" v=<structure>   composite data by value (fresh structures on decode)
//	k="r" h=<id> t/c      a Perl reference by handle (t=reftype, c=class)
//	k="f" h=<id>          a host function by id (decodes to a Perl closure)
type WireNode struct {
	K string          `json:"k"`
	V json.RawMessage `json:"v,omitempty"`
	H uint64          `json:"h,omitempty"`
	T string          `json:"t,omitempty"`
	C string          `json:"c,omitempty"`
}

// EvalResult is the decoded perl_eval response envelope.
type EvalResult struct {
	Ok     bool   `json:"ok"`
	Result string `json:"result"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error  string `json:"error"`
}

// callEnvelope is the decoded perl_call response document. Exit is set (and
// Ok false) when the guest called exit() and the bridge caught the unwind.
type callEnvelope struct {
	Ok     bool       `json:"ok"`
	Result []WireNode `json:"result"`
	Exit   *int       `json:"exit"`
	Error  string     `json:"error"`
}

// PerlDied reports a Perl-level die/croak (its Message is $@'s text). The
// public package converts it to its user-facing error type.
type PerlDied struct {
	Message string
}

func (e *PerlDied) Error() string { return e.Message }

// ExitError is a guest exit() the perl_call bridge caught cleanly: the guest
// unwound back to the call frame, so the interpreter is still flushable and
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
func (p *Perl) drainReleases(ctx context.Context) {
	p.relMu.Lock()
	ids := p.pendingRel
	p.pendingRel = nil
	p.relMu.Unlock()
	if len(ids) == 0 || p.closed.Load() {
		return
	}
	nodes := make([]WireNode, len(ids))
	for i, id := range ids {
		raw, err := json.Marshal(id)
		if err != nil {
			return
		}
		nodes[i] = WireNode{K: "d", V: raw}
	}
	_, _ = p.call(ctx, "__plwasm_release_all", nodes, false)
}

// ReleaseHandle releases one registry pin now instead of waiting for the
// next drain. A no-op on a closed instance.
func (p *Perl) ReleaseHandle(id uint64) error {
	if p.closed.Load() {
		return nil
	}
	raw, err := json.Marshal(id)
	if err != nil {
		return err
	}
	_, err = p.call(context.Background(), "__plwasm_release", []WireNode{{K: "d", V: raw}}, false)
	return err
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
	if opts.FSAccess != nil {
		wasi.SetFSAccessHook(opts.FSAccess)
	}
	if opts.NetAccess != nil {
		wasi.SetNetAccessHook(opts.NetAccess)
	}
	if opts.Dial != nil {
		wasi.SetDialHook(opts.Dial)
	}
	if opts.Resolve != nil {
		wasi.SetResolveHook(opts.Resolve)
	}
	if opts.Exec != nil {
		wasi.SetExecHook(opts.Exec)
	}
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

// Eval compiles and runs src in the interpreter's persistent package and
// returns the decoded envelope. A Perl-level die is reported inside the
// envelope (Ok=false, Error=$@); an error return is a host/transport failure
// or a context cancellation.
func (p *Perl) Eval(ctx context.Context, src string) (EvalResult, error) {
	if err := ctx.Err(); err != nil {
		return EvalResult{}, err
	}
	if p.closed.Load() {
		return EvalResult{}, ErrClosed
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
	var env EvalResult
	if err := json.Unmarshal([]byte(js), &env); err != nil {
		return EvalResult{}, fmt.Errorf("decode eval result %q: %w", js, err)
	}
	if interrupted && !env.Ok {
		// We tripped the interrupt and the eval died: report the
		// cancellation, not the croak text it surfaced as.
		return EvalResult{}, ctx.Err()
	}
	return env, nil
}

// Call invokes the named Perl subroutine in list context with the given
// argument nodes and returns the result nodes. A Perl-level die comes back
// as *PerlDied, a caught guest exit() as *ExitError; other errors are
// host/transport failures or a context cancellation.
func (p *Perl) Call(ctx context.Context, name string, args []WireNode) ([]WireNode, error) {
	return p.call(ctx, name, args, true)
}

// call is Call's engine; drain=false skips the release-queue flush (the
// flush itself and handle releases use that path, so a drain cannot recurse).
func (p *Perl) call(ctx context.Context, name string, args []WireNode, drain bool) ([]WireNode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.closed.Load() {
		return nil, ErrClosed
	}
	if drain {
		p.drainReleases(ctx)
	}
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendString(buf, 2, name)
	buf = pbAppendString(buf, 3, string(argsJSON))

	disarm := p.armInterrupt(ctx)
	resp, invokeErr := p.m.invoke(0, midCall, buf, wasm2go.Inv_0_0)
	interrupted := disarm()

	if invokeErr != nil {
		return nil, invokeErr
	}
	if e := pbExtractError(resp); e != nil {
		return nil, e
	}
	js := readScalarAtField(resp, 1, (*pbReader).readString)
	var env callEnvelope
	if err := json.Unmarshal([]byte(js), &env); err != nil {
		return nil, fmt.Errorf("decode call result %q: %w", js, err)
	}
	if !env.Ok {
		if env.Exit != nil {
			return nil, &ExitError{Code: *env.Exit}
		}
		if interrupted {
			return nil, ctx.Err()
		}
		return nil, &PerlDied{Message: env.Error}
	}
	return env.Result, nil
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
