// Package perl embeds the Perl interpreter — compiled to wasm and
// transpiled to pure Go — behind a sandboxed, multi-instance API. Each Perl
// value is an isolated interpreter: its own memory, filesystem view, and
// environment. No cgo, no system perl.
package perl

import (
	"context"
	"fmt"
	"sync"

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
func New(cfg Config) (*Perl, error) {
	// Resolve the filesystem and standard library location: every file the
	// instance opens goes through Config.FS, and StdlibDir is a path inside
	// it. With a supplied backend the stdlib must live in the FS (host
	// backends pair with ExtractStdlib; NewStdlibMemFS pre-loads a MemFS).
	// The nil default is a PRIVATE in-memory filesystem pre-loaded with the
	// stdlib - sandboxed, nothing touches the host disk.
	if cfg.FS == nil {
		fsys, err := NewStdlibMemFS()
		if err != nil {
			return nil, fmt.Errorf("build in-memory stdlib: %w", err)
		}
		cfg.FS = fsys
	}
	if cfg.StdlibDir == "" {
		cfg.StdlibDir = "/"
	}

	p := &Perl{}
	raw, err := internal.New(internal.InstanceOptions{
		StdlibDir: cfg.StdlibDir,
		Env:       cfg.Env,
		// Perl cannot boot without a /dev/null (the -e bootstrap opens the
		// bit bucket as its script filehandle); synthesize one over any
		// backend.
		FS:                 withDevNull(cfg.FS),
		FSAccess:           cfg.FSAccess,
		NetAccess:          cfg.NetAccess,
		Dial:               cfg.Dial,
		Resolve:            cfg.Resolve,
		Exec:               cfg.Exec,
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
	p.raw = raw
	return p, nil
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
	p.funcsMu.RUnlock()

	raw, err := p.raw.Clone(c.handleUserCallback)
	if err != nil {
		return nil, err
	}
	c.raw = raw
	return c, nil
}

// Close finalizes the instance. The Perl must not be used afterward: every
// later Eval/Call errors, and outstanding *Ref handles become inert (their
// finalizers only touch host-side state). Idempotent.
func (p *Perl) Close() error {
	return p.raw.Close()
}
