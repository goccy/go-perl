// Package internal holds the machinery under the public go-perl API: the
// generated wasm2go binding (perl.go), the per-instance bridge transport,
// the copy-on-write snapshot/clone plumbing, and the native-XS hook surface
// the xs subpackage drives. Nothing here is application API — the public
// package wraps a *Perl from this package and exposes only what users
// should call.
package internal

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/netip"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
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
	midRunProgram       = 21
	midSetGoDispatcher  = 22
	midXSHelper         = 23
	midXSOp             = 24
	midXSOpBox          = 25
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
	// OwnState keeps a program that is the host process (HostProcess) a
	// process of its own inside it: the host process does not follow its
	// working directory or environment, a command it starts runs in its
	// working directory, its pid is its own (ownPIDProcess), and the
	// signals the host process receives are not relayed to it.
	OwnState bool
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
	// mapped is the mapping of the runtime's the instance's memory is
	// (base.MapMemory, a SharedImage's Memory, base.ForkMemory); unmapped on
	// Close. nil for memory of the Go heap (bootPrivate).
	mapped []byte
	// onHeap is why the memory is of the Go heap when it was to be a
	// mapping (onGoHeap), until native code holds addresses of it; nil
	// for none.
	onHeap error

	// callbackDepth counts the callbacks of the guest that are running (one
	// inside the other), and jump is a non-local exit of the guest that
	// one of them has to let through when it returns: see guestJump. Both
	// belong to the goroutine that runs the interpreter.
	callbackDepth int
	jump          guestJump
	// xsOpCall is XSOpRun's call of the guest, made once (a closure made
	// for every operation native code runs was allocated each time), and
	// xsOpParked whether the operation it ran last parked a jump.
	xsOpCall   invoker
	xsOpParked bool

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
	magicFree func(id uint32) []byte
	// ppHook handles pp-hook dispatch for op types a native module claimed
	// (reserved callback method id -3).
	ppHook func(req []byte) []byte
	// magicLocal gives the SV `local` puts in a variable's place the
	// host-side magic of the variable (reserved callback method id -8).
	magicLocal func(id uint32, sv uint64) []byte
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
	// thread is the thread the instance runs its guest code on, when it
	// runs native code (see instanceThread); nil runs it on the callers'
	// goroutines.
	thread atomic.Pointer[instanceThread]
}

// forkFamily knows the running pseudo-processes forked from one instance,
// directly or not.
type forkFamily struct {
	origin *Perl

	mu sync.Mutex
	// running holds the fork children whose program has not ended, with
	// the pid each is known by. Which ones those are can be learnt nowhere
	// else: the process table of the host state knows a child only once
	// the fork that made it is over.
	running map[*Perl]int32
	// over records that the first process has ended: no pseudo-process is
	// started from then on. A child that forks while the others are being
	// ended would otherwise outlive them.
	over bool
	// errs are failures of the family's that no caller takes, by where they
	// come from, which end returns to the first process's Close.
	errs map[failureKind]*firstError
}

// failureKind is where a failure the family records (fail) comes from: the
// first of each kind is kept, so that one that repeats does not hide one of
// another kind.
type failureKind string

const (
	failHostSIGCHLD failureKind = "the host process's SIGCHLD for a fork child"
	failAdoption    failureKind = "signals the host process ignores after a library loaded"
	failChildEnd    failureKind = "the release of a fork child"
	failGoHeap      failureKind = "the instance's memory is of the Go heap, where native code holds addresses of it"
)

// firstError keeps the first of the failures it is given and counts the
// others: a failure that repeats (at every fork, at every report of the
// guest's) holds no more than one.
type firstError struct {
	err  error
	more int
}

// add records err.
func (e *firstError) add(err error) {
	if e.err == nil {
		e.err = err
		return
	}
	e.more++
}

// take returns the failures recorded, and forgets them.
func (e *firstError) take() error {
	err := e.err
	if e.more > 0 {
		err = fmt.Errorf("%w (and %d more failures)", err, e.more)
	}
	*e = firstError{}
	return err
}

// fail records err, of kind, for end to return. One that comes once the
// family is over has no Close left to go to: it is said on the first
// process's standard error.
func (f *forkFamily) fail(kind failureKind, err error) {
	f.mu.Lock()
	over := f.over
	if !over {
		if f.errs == nil {
			f.errs = map[failureKind]*firstError{}
		}
		if f.errs[kind] == nil {
			f.errs[kind] = &firstError{}
		}
		f.errs[kind].add(err)
	}
	f.mu.Unlock()
	if over {
		// A failure to write it has nowhere further to go.
		_, _ = fmt.Fprintf(f.origin.stderr(), "perl: %s: %v\n", kind, err)
	}
}

// start records child as the running pseudo-process pid. It reports false,
// and records nothing, when the family is over.
func (f *forkFamily) start(child *Perl, pid int32) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.over {
		return false
	}
	if f.running == nil {
		f.running = map[*Perl]int32{}
	}
	f.running[child] = pid
	return true
}

// member returns the instance of the family whose module is m: the first
// process or a running pseudo-process; nil for none.
func (f *forkFamily) member(m *base.Module) *Perl {
	if f.origin.m.g == m {
		return f.origin
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for child := range f.running {
		if child.m.g == m {
			return child
		}
	}
	return nil
}

// ended takes child, whose program has ended, off the running ones.
func (f *forkFamily) ended(child *Perl) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.running, child)
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
// descriptors are closed. Unlike the end of a process group, it is said on
// the first process's standard error, with the pids: a process that a
// program forks and leaves running (a daemon, a child that writes after
// its parent exited) outlives its parent elsewhere, and here it runs in
// the program's own process, which cannot go on without the program.
//
// end does not wait for the goroutine of a pseudo-process to notice. One
// that is in a call the host cannot interrupt (native code of an XS module,
// a Go function the embedder registered) would hold the first process's end
// up for as long as that call takes, and a process does not wait for what
// it kills. Every blocking call of the host's own is ended by the signal,
// so the goroutine is normally gone a moment later; it gives back the
// child's memory itself.
func (f *forkFamily) end() error {
	f.mu.Lock()
	f.over = true
	children := make([]*Perl, 0, len(f.running))
	pids := make([]int32, 0, len(f.running))
	for child, pid := range f.running {
		children = append(children, child)
		pids = append(pids, pid)
	}
	var errs []error
	for _, kind := range slices.Sorted(maps.Keys(f.errs)) {
		errs = append(errs, fmt.Errorf("%s: %w", kind, f.errs[kind].take()))
	}
	f.errs = nil
	err := errors.Join(errs...)
	f.mu.Unlock()
	if len(pids) > 0 {
		slices.Sort(pids)
		names := make([]string, len(pids))
		for i, pid := range pids {
			names[i] = strconv.Itoa(int(pid))
		}
		fmt.Fprintf(f.origin.stderr(), "perl: the program exited while processes it forked were still running;"+
			" they run inside its process and end with it: %s\n", strings.Join(names, ", "))
	}
	// Every one is killed before anything of any of them is taken away: a
	// child that sees a command die, or a descriptor close, must not get
	// to act on it.
	for _, child := range children {
		err = errors.Join(err, child.signalForkChild(sigKILL))
	}
	// A child that has noticed by now has killed its commands itself
	// (releaseHostState); this is for the ones that have not.
	for _, child := range children {
		err = errors.Join(err, child.wasi.KillCommands())
	}
	for _, child := range children {
		err = errors.Join(err, child.wasi.CloseDescriptors())
	}
	return err
}

// Closed reports whether Close has run.
func (p *Perl) Closed() bool { return p.closed.Load() }

// Environment is the environment the instance was made with: the program's
// as it starts, before the program changes %ENV.
func (p *Perl) Environment() []string { return p.opts.Env }

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
	_, _ = p.invoke(0, midRelease, buf, wasm2go.Inv_0_20)
}

// New boots a fresh instance from opts, mapping the process-wide
// copy-on-write snapshot when possible (see snapshot.go).
func New(opts InstanceOptions) (_ *Perl, err error) {
	m := &Module{}
	// A local, not the result: the cleanup below needs the instance on
	// every failure, which returns no instance.
	p := &Perl{m: m, opts: opts}
	p.family = &forkFamily{origin: p}
	wasi, err := buildWASI(opts, p.family)
	if err != nil {
		return nil, err
	}
	p.wasi = wasi
	defer func() {
		if err != nil {
			err = errors.Join(err, unbindHostSignals(wasi), wasi.CloseAll(), p.releaseMapping(), p.family.end())
		}
	}()

	if err := p.boot(opts, wasi); err != nil {
		return nil, err
	}
	// The guest asks the dispatcher for the host's socket addresses
	// (sockaddrMethodID) whenever it is shown the host's numbers.
	if opts.HostProcess {
		if err := p.EnsureDispatcher(); err != nil {
			return nil, err
		}
	}

	// Resolve the interrupt-flag address up front so Eval's context watchdog
	// can fire without calling into a busy instance.
	addr, err := p.interruptAddr()
	if err != nil {
		return nil, fmt.Errorf("resolve interrupt address: %w", err)
	}
	var addrErr error
	base.AccessMemory(m.g, func(mem []byte) {
		if addr+4 > uint64(len(mem)) {
			addrErr = fmt.Errorf("interrupt address %#x out of range (memory is %d bytes)", addr, len(mem))
		}
	})
	if addrErr != nil {
		return nil, addrErr
	}
	p.intrAddr = addr
	p.writeHostABI()
	return p, nil
}

// hostWordsBaseOffset is where, after the interpreter's interrupt flag
// (perl-wasm's g_hostwords), the word lies that says where this process has
// the instance's memory: pack 'p' and 'P' add it to an address of the
// memory and unpack 'p' and 'P' take it away again (the pointers are for
// native code, which reads the process's memory). 0 means no base: the
// addresses are the memory's own.
const hostWordsBaseOffset = 8

// hostWordsABIOffset is where, after the interpreter's interrupt flag
// (perl-wasm's g_hostwords), the word lies of the facts of the host's C ABI
// (HostABI), which the interpreter is told before any code runs in it: a
// copy of an instance (a Clone, a fork child) has the word with the memory.
const hostWordsABIOffset = 4

// writeHostABI writes the facts of the host's C ABI into the word.
func (p *Perl) writeHostABI() {
	base.AccessMemory(p.m.g, func(mem []byte) {
		at := p.intrAddr
		if at+hostWordsABIOffset+4 > uint64(len(mem)) {
			return
		}
		binary.LittleEndian.PutUint32(mem[at+hostWordsABIOffset:], HostABI())
	})
}

// MarkMemoryBase gives an instance that runs native code its base, before
// any code runs in it: the configuration decides it once, for the life of
// the instance (a pointer pack 'p' made before a base appeared would be
// taken for an address of the process afterwards). Memory that is a
// mapping of the runtime's does not move: it is as long as the memory can
// grow. Memory of the Go heap (bootPrivate) moves when it grows past its
// room (MemoryReserveBytes), and the base is then out of date.
func (p *Perl) MarkMemoryBase() {
	p.nativeHoldsAddresses()
	p.writeMemoryBase(false)
}

// writeMemoryBase writes the base of the instance's memory into the word;
// with ifSet, only when the word has a base already: a copy of an instance
// (a Clone, a fork child) has its memory in a place of its own, and its word
// came with the original's memory.
func (p *Perl) writeMemoryBase(ifSet bool) {
	base.AccessMemory(p.m.g, func(mem []byte) {
		at := p.intrAddr
		if len(mem) == 0 || at+hostWordsBaseOffset+8 > uint64(len(mem)) {
			return
		}
		if ifSet && binary.LittleEndian.Uint64(mem[at+hostWordsBaseOffset:]) == 0 {
			return
		}
		binary.LittleEndian.PutUint64(mem[at+hostWordsBaseOffset:], uint64(uintptr(unsafe.Pointer(&mem[0]))))
	})
}

// buildWASI assembles the per-instance WASI host from the options. Sandbox
// by default: no host environment, no host stdio (a nil Stdin reads as
// empty, a nil Stdout or Stderr discards; Perl-level print output is still
// captured into the eval envelope by the bridge, these back the raw guest
// fds 0/1/2), and a nil Dial, Resolve or Exec denies every connection,
// lookup or spawn. Sockets the guest holds may accept, receive and send:
// what it may reach is decided by Dial, and taking a local address to
// accept connections on (Listen) makes the program a server on the host's
// network, which only a program that is the host process may; its
// descriptors are the host's. Any other program's datagram socket is bound
// at a port the host picks when it first sends without having been bound
// (sendsFromAPickedPort). fork is the fork family's (forker).
func buildWASI(opts InstanceOptions, family *forkFamily) (*base.WasiStubs, error) {
	cfg := base.WASIConfig{
		Env:       opts.Env,
		Stdin:     opts.Stdin,
		Stdout:    opts.Stdout,
		Stderr:    opts.Stderr,
		FS:        opts.FS,
		NetAccess: base.AllowAllNetAccess,
		Dial:      opts.Dial,
		Resolve:   opts.Resolve,
		Listen:    sendsFromAPickedPort,
	}
	if opts.Exec != nil {
		cfg.Exec = func(r base.ExecRequest) bool { return opts.Exec(r.Path, r.Argv) }
	}
	if opts.HostProcess {
		if err := keepRuntimeHigh(); err != nil {
			return nil, fmt.Errorf("perl: keep the Go runtime's descriptors high: %w", err)
		}
		proc, err := hostProcess()
		if err != nil {
			return nil, fmt.Errorf("perl: the host process: %w", err)
		}
		cfg.HostDescriptors = true
		cfg.Process = proc
		cfg.ProcessChanges = newHostChanges()
		cfg.Users = base.HostUsers(nil)
		cfg.System = base.HostSystem()
		cfg.IPC = base.HostIPC()
		cfg.Services = base.HostServices(nil)
		cfg.Syscalls = base.HostSyscalls()
		cfg.Signals = hostSignalsConfig()
		cfg.TakeSignal = hostSignalTaker()
		cfg.Children = hostChildrenConfig()
		cfg.SignalIgnored = hostSignalIgnoredFunc()
		cfg.Numbering = base.HostNumbering()
		cfg.MonotonicClock = hostMonotonicClock()
		cfg.Listen = base.AllowAllListen
		// A program that is the host process is where the host is: when it
		// names no time zone (TZ), its local time is the host's, its working
		// directory and environment are the host process's. An isolated
		// program is in UTC.
		cfg.LocalZone = time.Local
		if opts.OwnState {
			cfg.Process = newOwnPIDProcess(proc)
		} else {
			cfg.Chdir = hostChdir
			cfg.CommandDir = hostCommandDir
			cfg.Setenv = hostSetenv
		}
	}
	if opts.Syslog {
		cfg.Syslog = NewSyslogger()
	}
	if opts.Fork {
		cfg.Fork = forker{family: family}
	}
	// Last: the instance takes the files over from here on, and closes them.
	if opts.InheritDescriptors {
		files, err := inheritedFiles(opts)
		if err != nil {
			return nil, fmt.Errorf("perl: the descriptors the process was started with: %w", err)
		}
		cfg.Files = files
	}
	wasi := base.NewWASI(cfg)
	// A program with OwnState receives no signal of the host process: the
	// host's are the embedder's.
	if opts.HostProcess && !opts.OwnState {
		if err := bindHostSignals(wasi); err != nil {
			return nil, errors.Join(fmt.Errorf("perl: signals of the host process: %w", err), unbindHostSignals(wasi), wasi.CloseAll())
		}
	}
	return wasi, nil
}

// sendsFromAPickedPort is the Listen policy of a program that is not the
// host process: of the local addresses a socket may take, only the one the
// host binds a datagram socket at when it first sends to an address
// without having been bound, the unspecified address at a port the host
// picks, which the send needs. An address the runtime gives that is not an
// IP and a port is denied.
func sendsFromAPickedPort(network, address string) bool {
	if network != "udp" {
		return false
	}
	addr, err := netip.ParseAddrPort(address)
	return err == nil && addr.Addr().IsUnspecified() && addr.Port() == 0
}

// boot brings the instance's wasm module up and obtains the interpreter
// handle — via the copy-on-write snapshot when possible, privately otherwise.
// The first instance of a process starts on memory of its own (see
// sharedSnapshot). GO_PERL_NO_SHARED_IMAGE asks for memory of the Go heap
// for every instance.
func (p *Perl) boot(opts InstanceOptions, wasi *base.WasiStubs) error {
	if os.Getenv("GO_PERL_NO_SHARED_IMAGE") != "" {
		return p.bootPrivate(opts, wasi)
	}
	snap := sharedSnapshot(opts.StdlibDir, opts.SiteDir, opts.Env)
	if snap == nil {
		return p.bootOwn(opts, wasi)
	}
	if snap.err == nil {
		mem, err := snap.img.Memory(memoryCeiling(opts, snap.img.Size()))
		if err == nil {
			p.m.g = wasm2go.NewFromSnapshot(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m}, mem, snap.img.Size(), snap.img.Globals())
			p.h = snap.handle
			p.mapped = mem
			return nil
		}
		p.onGoHeap(err)
	} else {
		p.onGoHeap(snap.err)
	}
	return p.bootPrivate(opts, wasi)
}

// onGoHeap keeps why the memory of an instance that was to be a mapping of
// the runtime's is of the Go heap instead (bootPrivate). It moves when it
// grows past its room, which matters to native code that holds addresses
// of it only: the reason is a failure once there is such code
// (nativeHoldsAddresses). A system without such mappings
// (errors.ErrUnsupported) is no failure.
func (p *Perl) onGoHeap(err error) {
	if !errors.Is(err, errors.ErrUnsupported) {
		p.onHeap = err
	}
}

// nativeHoldsAddresses is told that native code holds addresses of the
// instance's memory from now on. Memory of the Go heap that the instance
// has for a failure (onGoHeap) moves under them: the failure is the
// instance's Close's. The instance runs all the same.
func (p *Perl) nativeHoldsAddresses() {
	if p.onHeap != nil {
		p.family.fail(failGoHeap, p.onHeap)
		p.onHeap = nil
	}
}

// memoryCeiling is the limit of an instance's linear memory: the memory
// cap of opts rounded down to a wasm page, the host's memory for
// HostMemoryLimit (when it can be told), or defaultMemoryCeiling without
// one (or with one below least, the memory the instance starts with). It
// is the length of the mapping the instance's memory is, and its growth
// cap, however the instance gets its memory.
func memoryCeiling(opts InstanceOptions, least uint64) int {
	switch {
	case opts.MaxMemoryBytes == HostMemoryLimit:
		if size, ok := hostMemory(); ok && size/wasmPage*wasmPage >= least {
			return int(size / wasmPage * wasmPage)
		}
	case opts.MaxMemoryBytes > 0:
		if max := opts.MaxMemoryBytes / wasmPage * wasmPage; uint64(max) >= least {
			return max
		}
	}
	return defaultMemoryCeiling
}

// bootOwn starts the instance on memory of its own: a mapping of the
// runtime's that keeps one address for the instance's life (native code
// holds pointers into it), into which the data segments are installed
// before the module is constructed on it. Nothing is imaged or copied
// first; what a fork does with the memory is base.ForkMemory's (see
// fork.go: copy-on-write on Linux, a copy elsewhere). Without such a
// mapping the memory is of the Go heap (bootPrivate, onGoHeap).
func (p *Perl) bootOwn(opts InstanceOptions, wasi *base.WasiStubs) error {
	mem, err := base.MapMemory(memoryCeiling(opts, wasm2go.InitialMemoryBytes))
	if err != nil {
		p.onGoHeap(err)
		return p.bootPrivate(opts, wasi)
	}
	wasm2go.InstallDataSegments(mem)
	p.m.g = wasm2go.NewWithMemory(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m}, mem, wasm2go.InitialMemoryBytes)
	p.mapped = mem
	return p.start(opts)
}

// bootPrivate is the no-snapshot path: run the reactor initializer and
// perl_new on a private linear memory of the Go heap, which gets the room
// MemoryReserveBytes asks for, and moves when it grows past it.
func (p *Perl) bootPrivate(opts InstanceOptions, wasi *base.WasiStubs) (err error) {
	ceiling := memoryCeiling(opts, wasm2go.InitialMemoryBytes)
	if reserve := opts.MemoryReserveBytes; reserve > 0 {
		p.m.g = wasm2go.NewWithWASIReserve(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m}, reserve)
	} else {
		p.m.g = wasm2go.NewWithWASI(wasi, envStubs{m: p.m}, wasmifyStubs{m: p.m})
	}
	// Cap linear-memory growth so a runaway allocation fails in the guest
	// instead of growing the host process unbounded: the same limit as an
	// instance on a mapping has.
	wasm2go.SetMaxMemory(p.m.g, uint64(ceiling))
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
	out, invokeErr := p.invoke(0, mid, buf, inv)
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
	resp, err := p.invoke(0, midClose, buf, wasm2go.Inv_0_6)
	if err == nil {
		err = pbExtractError(resp)
	}
	// the native modules' state goes on the thread they ran on
	var relErr error
	p.OnThread(func() { relErr = p.releaseHostState() })
	p.endThread()
	return errors.Join(err, relErr)
}

// releaseHostState gives back what the host holds for an instance that
// will never be entered again: its open descriptors and its memory mapping.
// The end of the family's first process is the end of the pseudo-processes
// forked from it that are still running.
func (p *Perl) releaseHostState() error {
	if xsDynamicLoader != nil && xsDynamicLoader.Release != nil {
		xsDynamicLoader.Release(p)
	}
	var err error
	first := !p.ForkChild()
	if !first && p.family.isOver() {
		// A pseudo-process that goes because the family is over takes its
		// commands with it. It does that itself, here, before its process
		// table goes: whoever ended it may find the table gone already.
		err = p.wasi.KillCommands()
	}
	err = errors.Join(err, unbindHostSignals(p.wasi), p.wasi.CloseAll())
	if first && p.family != nil {
		err = errors.Join(err, p.family.end())
	}
	return errors.Join(err, p.releaseMapping())
}

// stderr is where the host reports failures of a pseudo-process that has no
// caller to return them to.
func (p *Perl) stderr() io.Writer {
	if p.opts.Stderr != nil {
		return p.opts.Stderr
	}
	return io.Discard
}

// releaseMapping returns the instance's mapping to the OS. Must only run
// when the module will never be entered again.
func (p *Perl) releaseMapping() error {
	if p.mapped == nil {
		return nil
	}
	// Under the lock of base.AccessMemory: another goroutine that is
	// setting a flag in this memory (a signal for a fork child) is done
	// first, and one that comes later finds the instance closed.
	var err error
	base.AccessMemory(p.m.g, func([]byte) { err = base.UnmapMemory(p.mapped) })
	p.mapped = nil
	if err != nil {
		return fmt.Errorf("unmap the instance's memory: %w", err)
	}
	return nil
}

// ArchTag names the kind of machine this process runs on, the way the
// directories of architecture-specific files are named: the native modules
// a process loads are built for it.
func ArchTag() string { return runtime.GOOS + "_" + runtime.GOARCH }

// perlNew starts the interpreter with its library in stdlibDir and, when
// siteDir is not empty, its site directories under siteDir, telling it the
// name of the architecture (ArchName): the host's, for which the native
// modules a program loads into this process are built; and the name of the
// system (OSName), the host's too.
func (p *Perl) perlNew(stdlibDir, siteDir string) (uint64, error) {
	var buf []byte
	buf = pbAppendString(buf, 1, stdlibDir)
	buf = pbAppendString(buf, 2, siteDir)
	buf = pbAppendString(buf, 3, ArchName())
	buf = pbAppendString(buf, 4, OSName())
	resp, err := p.invoke(0, midNew, buf, wasm2go.Inv_0_16)
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
	resp, err := p.invoke(0, midInterruptAddr, buf, wasm2go.Inv_0_13)
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
