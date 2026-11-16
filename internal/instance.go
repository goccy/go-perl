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
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

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
	midForkChild        = 9
	midHashDelete       = 10
	midHashGet          = 11
	midHashKeys         = 12
	midHashSet          = 13
	midInterruptAddr    = 14
	midInvoke           = 15
	midMethodCall       = 16
	midNew              = 17
	midNewArray         = 18
	midNewHash          = 19
	midRegisterNativeXS = 20
	midRelease          = 21
	midRunProgram       = 22
	midSetGoDispatcher  = 23
	midXSHelper         = 24
	midXSOp             = 25
	midXSOpBox          = 26
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
	// SiteDir is the directory the interpreter's site directories are
	// under (perl-wasm's perl_new lays them out below it); empty for an
	// interpreter without any.
	SiteDir string
	// Fork lets the guest fork: each fork runs a copy of the instance as a
	// pseudo-process inside this host process (see fork.go).
	Fork bool
	// HostProcess makes the program the host process: its ids, user
	// database and signals are the process's own.
	HostProcess bool
	// InheritDescriptors gives the program the descriptors above 2 that
	// the host process was started with (the first instance that asks).
	InheritDescriptors bool
	// Syslog gives the program the host's system logger (see syslog.go).
	Syslog bool
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	MaxMemoryBytes     int
	MemoryReserveBytes int
}

// SignalError is a guest that ended by a signal: one it neither handles nor
// ignores, whose default action ends the process. Signal is the guest's
// number for it (Linux numbering).
type SignalError struct{ Signal int }

func (e *SignalError) Error() string { return fmt.Sprintf("perl: terminated by signal %d", e.Signal) }

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
	intrAddr uint64
	// mapped is the copy-on-write mapping this instance lives in when it was
	// built from the process snapshot; unmapped on Close. nil for instances
	// that booted privately.
	mapped []byte
	// image is the image mapped is a private mapping of, when a fork can
	// start from it: the copy made for a fork child is then limited to the
	// pages this instance changed since the image was made (see fork.go).
	// It cannot be found from the mapping, hence the field. nil when the
	// memory is no such mapping.
	image *memoryImage

	// callbackDepth counts the callbacks of the guest that are running (one
	// inside the other), and jump is a non-local exit of the guest that
	// one of them has to let through when it returns: see guestJump. Both
	// belong to the goroutine that runs the interpreter.
	callbackDepth int
	jump          guestJump

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
	// magicLocal gives the SV `local` puts in a variable's place the
	// host-side magic of the variable (reserved callback method id -8).
	magicLocal func(id uint32, sv uint64)
	// dtorFire handles save-stack destructors registered by native modules
	// (reserved callback method id -4); its result is the response.
	dtorFire func(id uint32) []byte
	// perlioHook handles PerlIO layer-slot dispatch for layers a native
	// module registered (reserved callback method id -6).
	perlioHook func(req []byte) []byte
	// keywordHook handles keyword/infix plugin forwarding while the guest
	// parser compiles (reserved callback method id -7).
	keywordHook func(req []byte) []byte
	// hostCall handles the calls of native code that share one channel
	// (reserved callback method id -9).
	hostCall func(req []byte) []byte
	// processMemory reads memory of the process outside the guest's
	// (reserved callback method id -10): unpack 'p' and 'P' of such an
	// address.
	processMemory func(addr uint64, n int64) ([]byte, error)

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

	// OnFork runs on every fork child of this instance before it starts,
	// so the public package can give the child its own callback routing
	// (UserHandler) and OnFork.
	OnFork func(child *Perl)
	// family is shared by an instance and every pseudo-process forked from
	// it, directly or not. The instance that was not itself forked is the
	// family's first process: when it has ended, the pseudo-processes that
	// are still running are ended with it (see forkFamily.end).
	family *forkFamily
	// killed is the signal this fork child died of (0 = none).
	killed atomic.Int32
}

// forkFamily knows the running pseudo-processes forked from one instance,
// directly or not.
type forkFamily struct {
	origin *Perl

	mu sync.Mutex
	// running holds the fork children whose program has not ended. Which
	// ones those are can be learnt nowhere else: the process table of the
	// host state knows a child only once the fork that made it is over.
	running map[*Perl]struct{}
	// over records that the first process has ended: no pseudo-process is
	// started from then on. A child that forks while the others are being
	// ended would otherwise outlive them.
	over bool
}

// start runs child's program as the pseudo-process pid, on a goroutine of
// its own; exited is told the status it ends with. It reports false, and
// starts nothing, when the family is over.
func (f *forkFamily) start(child *Perl, pid int32, exited func(status int32)) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.over {
		return false
	}
	if f.running == nil {
		f.running = map[*Perl]struct{}{}
	}
	f.running[child] = struct{}{}
	go func() {
		status := child.runForkChild(pid)
		f.mu.Lock()
		delete(f.running, child)
		f.mu.Unlock()
		exited(status)
	}()
	return true
}

// isOver reports whether the first process of the family has ended.
func (f *forkFamily) isOver() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.over
}

// end ends the pseudo-processes that are still running, when the first
// process of the family has ended. To the program it is what the end of a
// process group is: every one of them is killed as by SIGKILL — it runs no
// Perl code from its next opcode on, no END block and no destructor — the
// commands it started and has not waited for are killed and reaped, and its
// descriptors are closed.
//
// end does not wait for the goroutine of a pseudo-process to notice. One
// that is in a call the host cannot interrupt (native code of an XS module,
// a Go function the embedder registered) would hold the first process's end
// up for as long as that call takes, and a process does not wait for what
// it kills. Every blocking call of the host's own is ended by the signal,
// so the goroutine is normally gone a moment later; it gives back the
// child's memory itself.
func (f *forkFamily) end() {
	f.mu.Lock()
	f.over = true
	children := make([]*Perl, 0, len(f.running))
	for child := range f.running {
		children = append(children, child)
	}
	f.mu.Unlock()
	// Every one is killed before anything of any of them is taken away: a
	// child that sees a command die, or a descriptor close, must not get
	// to act on it.
	for _, child := range children {
		_ = child.signalForkChild(sigKILL)
	}
	// A child that has noticed by now has killed its commands itself
	// (releaseHostState); this is for the ones that have not.
	for _, child := range children {
		child.wasi.KillCommands()
	}
	for _, child := range children {
		_ = child.wasi.CloseDescriptors()
	}
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
	_, _ = p.m.invoke(0, midRelease, buf, wasm2go.Inv_0_21)
}

// New boots a fresh instance from opts, mapping the process-wide
// copy-on-write snapshot when possible (see snapshot.go).
func New(opts InstanceOptions) (p *Perl, err error) {
	m := &Module{}
	wasi := buildWASI(opts)
	p = &Perl{m: m, wasi: wasi, opts: opts}
	p.family = &forkFamily{origin: p}

	if err := p.boot(opts, wasi); err != nil {
		return nil, err
	}
	p.installForkHook()

	// Resolve the interrupt-flag address up front so Eval's context watchdog
	// can fire without calling into a busy instance.
	addr, err := p.interruptAddr()
	if err != nil {
		p.releaseMapping()
		return nil, fmt.Errorf("resolve interrupt address: %w", err)
	}
	var addrErr error
	base.AccessMemory(m.g, func(mem []byte) {
		if addr+4 > uint64(len(mem)) {
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

// The words after the interpreter's interrupt flag (perl-wasm's
// g_hostwords): a marker that says they are there, and where this process
// has the instance's memory, which pack 'p' and 'P' add to an address of
// the memory and unpack 'p' and 'P' take away again (the pointers are for
// native code, which reads the process's memory).
const (
	hostWordsMagic       = 0x676f7077
	hostWordsMagicOffset = 4
	hostWordsBaseOffset  = 8
)

// MarkMemoryBase writes where this process has the instance's memory into
// the interpreter's word for it, from when native code runs in the
// instance: the native loader marks an instance when it loads the first
// module into it, and each copy of such an instance (a Clone, a fork
// child), whose memory is in a place of its own. Until then pack 'p' makes
// addresses of the interpreter's memory, and unpack 'p' reads nothing but
// that memory, as before. The memory does not move (it is a reservation as
// large as it can grow).
func (p *Perl) MarkMemoryBase() {
	base.AccessMemory(p.m.g, func(mem []byte) {
		at := p.intrAddr
		if len(mem) == 0 || at+hostWordsBaseOffset+8 > uint64(len(mem)) ||
			binary.LittleEndian.Uint32(mem[at+hostWordsMagicOffset:]) != hostWordsMagic {
			return
		}
		binary.LittleEndian.PutUint64(mem[at+hostWordsBaseOffset:], uint64(uintptr(unsafe.Pointer(&mem[0]))))
	})
}

// KeepRuntimeDescriptorsHigh has the Go runtime make the descriptors it
// needs for itself now, above the numbers a program's descriptors get. A
// program that is the host process (InstanceOptions.HostProcess) has the
// host's descriptor numbers, lowest free first, and expects the first file
// it opens to be descriptor 3. A command built around such a program calls
// this while its packages are initialized: the earlier, the fewer
// descriptors the runtime has made already.
func KeepRuntimeDescriptorsHigh() { base.KeepRuntimeDescriptorsHigh() }

// hostIDs is how a program that is the host process changes the process's
// user and group ids when a package linked into the binary knows a way the
// runtime does not have; see RegisterHostIDs. nil is the runtime's own way.
var hostIDs base.HostIDs

// RegisterHostIDs sets the way the ids of the host process are changed for
// every instance made afterwards whose program is the host process. The
// native module loader registers the C library's: a process that loads
// native code is linked with a C runtime, and has to change its ids through
// it. It is called from package initialization.
func RegisterHostIDs(ids base.HostIDs) { hostIDs = ids }

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
	wasi.SetHostProcess(opts.HostProcess)
	if opts.HostProcess && hostIDs != nil {
		wasi.SetHostIDs(hostIDs)
	}
	// A program that is the host process is where the host is: when it
	// names no time zone (TZ), its local time is the host's. An isolated
	// program is in UTC.
	if opts.HostProcess {
		wasi.SetLocalZone(time.Local)
	}
	if opts.InheritDescriptors {
		wasi.AdoptInheritedDescriptors()
	}
	if opts.Syslog {
		wasi.SetSyslogger(NewSyslogger())
	}
	// Taking a local address to accept connections on makes the program a
	// server on the host's network: only a program that is the host process
	// may. Traffic on sockets it already has is not gated here (outbound
	// connections have the Dial hook).
	mayListen := opts.HostProcess
	wasi.SetNetAccessHook(func(op string) bool { return op != "listen" || mayListen })
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
// The first instance of a process starts on memory of its own (see
// sharedSnapshot).
func (p *Perl) boot(opts InstanceOptions, wasi *base.WasiStubs) error {
	snap := sharedSnapshot(opts.StdlibDir, opts.SiteDir, opts.Env)
	if snap == nil {
		return p.bootOwn(opts, wasi)
	}
	if snap.err == nil {
		ceiling := memoryCeiling(opts, snap.img.Size())
		mem, err := snap.img.Memory(ceiling)
		if err == nil {
			p.m.g = wasm2go.NewFromSnapshot(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m}, mem, snap.img.Size(), snap.img.Globals())
			p.h = snap.handle
			p.mapped = mem
			p.image = snap.image.retain()
			return nil
		}
		// Mapping failed (exotic platform / fd exhaustion): boot privately.
	}
	return p.bootPrivate(opts, wasi)
}

// memoryCeiling is the limit of an instance's linear memory: the memory
// cap of opts rounded down to a wasm page, the host's memory for
// HostMemoryLimit (when it can be told), or defaultMemoryCeiling without
// one (or with one below least, the memory the instance starts with). It
// is the length of the mapping the instance's memory is, and its growth
// cap, however the instance gets its memory.
func memoryCeiling(opts InstanceOptions, least uint64) uint64 {
	switch {
	case opts.MaxMemoryBytes == HostMemoryLimit:
		if size, ok := hostMemory(); ok && size/wasmPage*wasmPage >= least {
			return size / wasmPage * wasmPage
		}
	case opts.MaxMemoryBytes > 0:
		if max := uint64(opts.MaxMemoryBytes) / wasmPage * wasmPage; max >= least {
			return max
		}
	}
	return defaultMemoryCeiling
}

// bootOwn starts the instance on memory of its own: a private mapping that
// keeps one address for the instance's life (native code holds pointers
// into it) and holds the data segments of a module constructed for the
// purpose. Nothing is imaged first; the instance's first fork images its
// memory (base.ForkMemory with the image AnonymousMemory returned).
func (p *Perl) bootOwn(opts InstanceOptions, wasi *base.WasiStubs) error {
	from := wasm2go.NewWithWASI(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m})
	img, mem, err := base.AnonymousMemory(from, memoryCeiling(opts, wasm2go.InitialMemoryBytes))
	if err != nil {
		// No mapping to be had: memory of the Go heap, which moves when
		// it grows past what was reserved.
		return p.bootPrivate(opts, wasi)
	}
	p.m.g = wasm2go.NewWithMemory(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m}, mem, wasm2go.InitialMemoryBytes)
	p.mapped = mem
	p.image = (&memoryImage{img: img}).retain()
	return p.start(opts)
}

// bootPrivate is the no-snapshot path: run the reactor initializer and
// perl_new on a private linear memory. Where linear memory is a
// reservation (base.MemReserved), the reservation holds the memory's whole
// ceiling, as the mapping of an instance on memory of its own does: a
// memory that outgrows its reservation moves to a larger one, and native
// code holds the address it had. A memory of the Go heap gets the room
// MemoryReserveBytes asks for, and moves when it grows past it.
func (p *Perl) bootPrivate(opts InstanceOptions, wasi *base.WasiStubs) (err error) {
	ceiling := memoryCeiling(opts, wasm2go.InitialMemoryBytes)
	reserve := opts.MemoryReserveBytes
	if base.MemReserved && uint64(reserve) < ceiling {
		reserve = int(ceiling)
	}
	if reserve > 0 {
		p.m.g = wasm2go.NewWithWASIReserve(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m}, reserve)
	} else {
		p.m.g = wasm2go.NewWithWASI(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m})
	}
	// Cap linear-memory growth so a runaway allocation fails in the guest
	// instead of growing the host process unbounded: the same limit as an
	// instance on a mapping has.
	wasm2go.SetMaxMemory(p.m.g, ceiling)
	return p.start(opts)
}

// start runs the reactor initializer and perl_new on the instance's fresh
// module.
func (p *Perl) start(opts InstanceOptions) (err error) {
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

	h, err := p.perlNew(opts.StdlibDir, opts.SiteDir)
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
	return p.valueOp(ctx, midInvoke, wasm2go.Inv_0_15, buf)
}

// MethodCallOp invokes $obj->method(args...) on the reference behind obj.
func (p *Perl) MethodCallOp(ctx context.Context, obj uint64, method string, args []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, obj)
	buf = pbAppendString(buf, 3, method)
	buf = pbAppendBytes(buf, 4, args)
	buf = pbAppendUint64(buf, 5, uint64(len(args)))
	return p.valueOp(ctx, midMethodCall, wasm2go.Inv_0_16, buf)
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
	return p.valueOp(ctx, midHashGet, wasm2go.Inv_0_11, buf)
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
	return p.valueOp(ctx, midHashSet, wasm2go.Inv_0_13, buf)
}

// HashDeleteOp performs delete $hv->{key}.
func (p *Perl) HashDeleteOp(ctx context.Context, hv uint64, key []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, hv)
	buf = pbAppendBytes(buf, 3, key)
	buf = pbAppendUint64(buf, 4, uint64(len(key)))
	return p.valueOp(ctx, midHashDelete, wasm2go.Inv_0_10, buf)
}

// HashKeysOp returns the envelope carrying keys %{$hv} as a node list.
func (p *Perl) HashKeysOp(ctx context.Context, hv uint64) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendUint64(buf, 2, hv)
	return p.valueOp(ctx, midHashKeys, wasm2go.Inv_0_12, buf)
}

// NewArrayOp materialises a fresh guest array from the node list and returns
// the envelope carrying its ref node.
func (p *Perl) NewArrayOp(ctx context.Context, vals []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendBytes(buf, 2, vals)
	buf = pbAppendUint64(buf, 3, uint64(len(vals)))
	return p.valueOp(ctx, midNewArray, wasm2go.Inv_0_18, buf)
}

// NewHashOp materialises a fresh guest hash from the alternating key/value
// node list and returns the envelope carrying its ref node.
func (p *Perl) NewHashOp(ctx context.Context, pairs []byte) ([]byte, bool, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendBytes(buf, 2, pairs)
	buf = pbAppendUint64(buf, 3, uint64(len(pairs)))
	return p.valueOp(ctx, midNewHash, wasm2go.Inv_0_19, buf)
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
	return errors.Join(err, p.releaseHostState())
}

// releaseHostState gives back what the host holds for an instance that
// will never be entered again: its open descriptors and its memory mapping.
// The end of the family's first process is the end of the pseudo-processes
// forked from it that are still running.
func (p *Perl) releaseHostState() error {
	if xsDynamicLoader != nil && xsDynamicLoader.Release != nil {
		xsDynamicLoader.Release(p)
	}
	first := p.family == nil || p.family.origin == p
	if !first && p.family.isOver() {
		// A pseudo-process that goes because the family is over takes its
		// commands with it. It does that itself, here, before its process
		// table goes: whoever ended it may find the table gone already.
		p.wasi.KillCommands()
	}
	err := p.wasi.CloseAll()
	if first && p.family != nil {
		p.family.end()
	}
	p.releaseMapping()
	return err
}

// stderr is where the host reports failures of a pseudo-process that has no
// caller to return them to.
func (p *Perl) stderr() io.Writer {
	if p.opts.Stderr != nil {
		return p.opts.Stderr
	}
	return io.Discard
}

// releaseMapping returns the instance's copy-on-write mapping to the OS. Must
// only run when the module will never be entered again.
func (p *Perl) releaseMapping() {
	if p.mapped != nil {
		// Under the lock of base.AccessMemory: another goroutine that is
		// setting a flag in this memory (a signal for a fork child) is
		// done first, and one that comes later finds the instance closed.
		base.AccessMemory(p.m.g, func([]byte) { base.UnmapMemory(p.mapped) })
		p.mapped = nil
	}
	p.image.release()
	p.image = nil
}

// memoryImage is a memory image that instances map privately, counted by
// the instances that may still fork from it: the image is closed, and its
// pages go once the mappings have gone, when the last of them is done.
type memoryImage struct {
	img  *base.SharedImage
	refs atomic.Int32
}

// retain counts one more instance that maps the image, and returns it.
func (i *memoryImage) retain() *memoryImage {
	if i != nil {
		i.refs.Add(1)
	}
	return i
}

// release counts an instance out.
func (i *memoryImage) release() {
	if i != nil && i.refs.Add(-1) == 0 {
		i.img.Close()
	}
}

// ArchTag names the kind of machine this process runs on, the way the
// directories of architecture-specific files are named: the native modules
// a process loads are built for it.
func ArchTag() string { return runtime.GOOS + "_" + runtime.GOARCH }

// perlNew starts the interpreter with its library in stdlibDir and, when
// siteDir is not empty, its site directories under siteDir, telling it the
// name of the architecture (ArchName): the host's, for which the native
// modules a program loads into this process are built.
func (p *Perl) perlNew(stdlibDir, siteDir string) (uint64, error) {
	var buf []byte
	buf = pbAppendString(buf, 1, stdlibDir)
	buf = pbAppendString(buf, 2, siteDir)
	buf = pbAppendString(buf, 3, ArchName())
	resp, err := p.m.invoke(0, midNew, buf, wasm2go.Inv_0_17)
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
func (p *Perl) interruptAddr() (uint64, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	resp, err := p.m.invoke(0, midInterruptAddr, buf, wasm2go.Inv_0_14)
	if err != nil {
		return 0, err
	}
	if e := pbExtractError(resp); e != nil {
		return 0, e
	}
	return readScalarAtField(resp, 1, (*pbReader).readUint64), nil
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
