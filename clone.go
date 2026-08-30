package perl

// Cloning a prepared instance.
//
// Clone snapshots a LIVE instance — everything it has compiled and loaded so
// far — and maps new instances copy-on-write from that image. This is what
// lets a server prepare one interpreter (modules vendored, application
// compiled) and then scale it to N workers without re-running any of that
// work per worker: the psgi package's Workers is built on it.

import (
	"fmt"
	"sync"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// cloneImage is the lazily built copy-on-write image of one instance's
// current state, shared by every clone taken from it.
type cloneImage struct {
	once sync.Once
	img  *base.SharedImage
}

// AddCloneHook registers fn to run on every future Clone of this instance,
// after the clone's core state is wired. Subsystems holding host-side state
// for the instance (the xs package's native-module loader) use it to attach
// their own per-clone state.
func (p *Perl) AddCloneHook(fn func(clone *Perl) error) {
	p.funcsMu.Lock()
	p.cloneHooks = append(p.cloneHooks, fn)
	p.funcsMu.Unlock()
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
	if p.closed.Load() {
		return nil, fmt.Errorf("perl: Clone on a closed instance")
	}
	p.cloneImg.once.Do(func() {
		proto := p.m.g
		p.cloneImg.img = base.NewSharedSnapshot(func() (*base.Module, error) {
			return proto, nil
		})
	})
	img := p.cloneImg.img
	if err := img.Err(); err != nil {
		return nil, fmt.Errorf("perl: snapshot instance: %w", err)
	}
	mem, err := img.Memory(snapshotCeiling)
	if err != nil {
		return nil, fmt.Errorf("perl: map clone memory: %w", err)
	}

	m := &Module{}
	wasi := buildWASI(p.cfg)
	c := &Perl{m: m, wasi: wasi, cfg: p.cfg}
	m.g = wasm2go.NewFromSnapshot(wasi, envStubs{m: m}, wasmifyStubs{m: m}, mem, img.Size(), img.Globals())
	c.mapped = mem
	c.h = p.h
	c.intrAddr = p.intrAddr

	// The guest memory carries p's compiled state, including the Go-bridge
	// function ids baked into bound subs and the registered dispatcher; the
	// host tables those ids resolve through come along.
	p.funcsMu.RLock()
	c.funcs = make(map[int32]GoFunc, len(p.funcs))
	for id, fn := range p.funcs {
		c.funcs[id] = fn
	}
	c.nextFuncID = p.nextFuncID
	hadDispatcher := p.dispatcherSet
	hooks := append([]func(*Perl) error(nil), p.cloneHooks...)
	p.funcsMu.RUnlock()

	// The guest memory holds a callback handle from p's registration, but
	// callback registries are per-module host state: register the clone's
	// own dispatcher so the handle the guest uses resolves here.
	if hadDispatcher {
		if err := c.ensureDispatcher(); err != nil {
			c.Close()
			return nil, fmt.Errorf("perl: clone dispatcher: %w", err)
		}
	}

	for _, hook := range hooks {
		if err := hook(c); err != nil {
			c.Close()
			return nil, fmt.Errorf("perl: clone hook: %w", err)
		}
	}
	return c, nil
}
