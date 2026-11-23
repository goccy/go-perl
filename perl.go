// Package perl embeds the Perl interpreter — compiled to wasm and
// transpiled to pure Go — behind a sandboxed, multi-instance API. Each Perl
// value is an isolated interpreter: its own memory, filesystem view, and
// environment. No cgo, no system perl.
package perl

import (
	"context"
	"errors"
	"fmt"
	"sync"
	_ "unsafe" // go:linkname

	// The link set of the project's CPAN dependencies.
	_ "github.com/goccy/go-perl/cpan"
	"github.com/goccy/go-perl/internal"
)

// Perl is one isolated Perl interpreter instance.
type Perl struct {
	raw *internal.Perl

	// funcs maps the ids baked into bound Perl subs to their Go functions
	// (see bridge.go).
	funcsMu    sync.RWMutex
	funcs      map[int32]GoFunc
	nextFuncID int32
	// dynaLoaderIDs are the function ids of DynaLoader's host functions
	// (dynaloader.go), zero without Config.NativeModules. They are kept
	// because those functions act on the instance that calls them: a clone
	// or fork child shares the ids but must get functions of its own.
	dynaLoaderIDs [3]int32
}

// Result is the outcome of one Eval: the expression's value and the
// Perl-level error state. A die/croak/compile error is a fact ABOUT the
// evaluation, so it lives here as Error (*PerlError, $@'s text), not in
// Eval's own error return — that one reports host and transport failures.
type Result struct {
	// Value is the evaluated expression's value (the last statement, in
	// scalar context), valid when Error is nil. See Value for the Perl
	// coercions its accessors apply.
	Value Value
	// Error is non-nil when the eval died: a Perl-level die/croak or a
	// compile error, carrying $@ as a *PerlError.
	Error error
	// Stdout / Stderr capture what the eval printed to STDOUT / STDERR
	// (the bridge redirects both onto in-memory scalars for the duration).
	Stdout string
	Stderr string
}

// New builds a fresh Perl instance, applies the sandbox config, and returns
// it ready to Eval. When the config permits, the instance is mapped
// copy-on-write from a process-wide snapshot of an already-initialized
// interpreter, so construction is cheap and the instances share the
// read-only bulk of their memory.
func New(cfg Config) (*Perl, error) { return newPerl(cfg, false) }

// newOwnState is New for an instance that is the host process
// (Config.HostProcess) and a process of its own inside it: the host process
// does not follow its working directory or environment, and its pid is its
// own (internal.InstanceOptions.OwnState). Not part of the API: the gperl
// command reaches it through go:linkname.
//
//go:linkname newOwnState
func newOwnState(cfg Config) (*Perl, error) { return newPerl(cfg, true) }

// onThread runs fn on the thread p's guest code runs on
// (internal.Perl.OnThread), where the entries into the guest fn makes run in
// place; on the calling goroutine for an instance that has no such thread.
// Not part of the API: the PSGI server reaches it through go:linkname.
//
//go:linkname onThread
func onThread(p *Perl, fn func()) { p.raw.OnThread(fn) }

func newPerl(cfg Config, ownState bool) (*Perl, error) {
	// Resolve the filesystem and standard library location: every file the
	// instance opens goes through Config.FS, and StdlibDir is a path inside
	// it. With a supplied backend the stdlib must live in the FS (host
	// backends pair with ExtractStdlib; NewStdlibMemFS pre-loads a MemFS).
	// The nil default is a PRIVATE in-memory filesystem pre-loaded with the
	// stdlib - sandboxed, nothing touches the host disk.
	var memfs *MemFS
	if cfg.FS == nil {
		var err error
		if memfs, err = NewStdlibMemFS(); err != nil {
			return nil, fmt.Errorf("build in-memory stdlib: %w", err)
		}
		cfg.FS = memfs
	}
	if cfg.StdlibDir == "" {
		cfg.StdlibDir = "/"
	}
	cpanInc, err := cpanLibInc(cfg.FS, memfs)
	if err != nil {
		return nil, fmt.Errorf("the CPAN module tree: %w", err)
	}

	// Perl cannot boot without a /dev/null (the -e bootstrap opens the bit
	// bucket as its script filehandle); synthesize one over a backend that
	// has none.
	fsys, err := withDevNull(cfg.FS)
	if err != nil {
		return nil, err
	}

	p := &Perl{}
	raw, err := internal.New(internal.InstanceOptions{
		StdlibDir:          cfg.StdlibDir,
		SiteDir:            cfg.SiteDir,
		Env:                cfg.Env,
		FS:                 fsys,
		Dial:               cfg.Dial,
		Resolve:            cfg.Resolve,
		Exec:               cfg.Exec,
		Fork:               cfg.Fork,
		HostProcess:        cfg.HostProcess,
		OwnState:           ownState,
		InheritDescriptors: cfg.InheritDescriptors,
		Syslog:             cfg.Syslog,
		Stdin:              cfg.Stdin,
		Stdout:             cfg.Stdout,
		Stderr:             cfg.Stderr,
		MaxMemoryBytes:     cfg.MaxMemoryBytes,
		MemoryReserveBytes: cfg.MemoryReserveBytes,
	})
	if err != nil {
		return nil, err
	}
	raw.UserHandler = p.handleUserCallback
	raw.OnFork = p.adoptForkChild
	p.raw = raw
	if cfg.NativeModules {
		// native code runs here: the interpreter runs on one thread for
		// its life, whatever goroutines call it (native modules keep state
		// per thread, see internal.Perl.KeepOnOneThread), and pointers pack
		// 'p' makes are addresses of the process from the start (see
		// internal.Perl.MarkMemoryBase)
		raw.KeepOnOneThread()
		raw.MarkMemoryBase()
		if err := p.installDynaLoader(); err != nil {
			p.Close()
			return nil, err
		}
	}
	if err := internal.XSLoadLinked(raw); err != nil {
		return nil, errors.Join(err, p.Close())
	}
	if cpanInc != "" {
		if err := p.AddInc(context.Background(), cpanInc); err != nil {
			return nil, errors.Join(err, p.Close())
		}
	}
	return p, nil
}

// adoptForkChild gives a pseudo-process forked from p its own host side of
// the Go bridge: the child's memory holds the function ids of every sub p
// bound, so the child resolves them through a copy of p's table.
func (p *Perl) adoptForkChild(raw *internal.Perl) {
	c := &Perl{raw: raw}
	p.funcsMu.RLock()
	c.funcs = make(map[int32]GoFunc, len(p.funcs))
	for id, fn := range p.funcs {
		c.funcs[id] = fn
	}
	c.nextFuncID = p.nextFuncID
	c.dynaLoaderIDs = p.dynaLoaderIDs
	p.funcsMu.RUnlock()
	c.bindDynaLoader()
	raw.UserHandler = c.handleUserCallback
	raw.OnFork = c.adoptForkChild
}

// Eval compiles and runs src in this instance's persistent package (its
// main::) and returns the structured result. A Perl-level die is reported via
// Result.Error (*PerlError), not as Eval's own error; a Go error indicates a
// host/transport failure (a wasm trap, encoding problem, ...) or a context
// cancellation.
//
// Cancelling ctx while the eval runs stops it at the next Perl opcode (the
// interpreter runs a pluggable run loop that tests a host-writable flag; see
// perl-wasm's bridge) and Eval returns ctx.Err(). A single long-running
// opcode — a pathological regex, a big sort — is not preempted until it
// returns to the run loop.
func (p *Perl) Eval(ctx context.Context, src string) (Result, error) {
	resp, interrupted, err := p.raw.EvalOp(ctx, src)
	if err != nil {
		return Result{}, err
	}
	return p.decodeEvalResult(ctx, resp, interrupted)
}

// Clone returns a new instance mapped copy-on-write from this instance's
// CURRENT state: every module compiled and every value that exists in p
// exists in the clone, and from here the two diverge privately, sharing the
// read-only bulk of their memory.
//
// Take clones at rest — between requests, never while an Eval/Call is
// running in p. The first Clone builds the image (a full memory copy);
// subsequent clones of the same instance reuse it, so p must not run
// anything between clones of one batch either. The clone inherits p's
// configuration (filesystem, environment, hooks) and its Go bindings.
func (p *Perl) Clone() (*Perl, error) {
	c := &Perl{}

	// The guest memory carries p's compiled state, including the Go-bridge
	// function ids baked into bound subs; the host tables those ids resolve
	// through come along. Copy them BEFORE the raw clone so callbacks fired
	// by clone hooks already resolve on the new owner.
	p.funcsMu.RLock()
	c.funcs = make(map[int32]GoFunc, len(p.funcs))
	for id, fn := range p.funcs {
		c.funcs[id] = fn
	}
	c.nextFuncID = p.nextFuncID
	c.dynaLoaderIDs = p.dynaLoaderIDs
	p.funcsMu.RUnlock()

	raw, err := p.raw.Clone(c.handleUserCallback)
	if err != nil {
		return nil, err
	}
	raw.OnFork = c.adoptForkChild
	c.raw = raw
	c.bindDynaLoader()
	return c, nil
}

// Close finalizes the instance. The Perl must not be used afterward: every
// later Eval/Call errors, and outstanding *Ref handles become inert (their
// finalizers only touch host-side state). Idempotent.
//
// Pseudo-processes forked from the instance (Config.Fork) that are still
// running are ended with it.
func (p *Perl) Close() error {
	return p.raw.Close()
}
