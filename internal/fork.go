package internal

// fork, as a pseudo-process.
//
// WASI has no fork and the guest cannot duplicate itself, so the host does
// it — the way perl emulates fork on Windows, with the whole instance
// standing in for perl_clone. When the guest forks (the proc_fork host
// import), forkGuest duplicates the instance's linear memory and globals
// into a NEW instance that also gets a copy of the fd table (its slots
// share the parent's open descriptors, as a fork child's do), and runs that
// instance on its own goroutine as the child. The guest is single-threaded
// and suspended inside the import while the duplicate is made, so it is
// consistent.
//
// The duplicate is copy-on-write where the host can tell which pages an
// instance changed (see base.ForkMemory): parent and child are private
// mappings of one image, and a fork copies what the parent changed since
// that image was made, not everything it holds.
//
// The child has the parent's memory but none of its native call stack, so
// the runtime enters it through the function the guest gave its fork
// (perl-wasm's perl_fork_child_entry), which resumes the program right after
// the forking op and returns when the child's program ends. The runtime
// publishes its exit status to the parent's process table, where waitpid
// finds it like any spawned process's.

import (
	"encoding/binary"
	"errors"
	"fmt"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// interruptSignal marks an interrupt-flag value as a terminating signal
// (its low bits are the signal number); see perl-wasm's pl.h.
const interruptSignal = 0x100

// sigKILL is the one signal a process cannot catch, block or ignore.
const sigKILL = 9

// errFamilyOver is the failure of a fork made by a pseudo-process that is
// being ended because the first process of its family has ended.
var errFamilyOver = errors.New("perl: fork: the program has ended")

// installForkHook lets the guest fork when the configuration allows it.
// Without the hook the import reports ENOSYS and fork returns undef.
func (p *Perl) installForkHook() {
	if p.opts.Fork {
		p.wasi.SetForkHook(p.forkGuest)
	}
}

// forkGuest is the fork hook: it builds the child instance from p's current
// state, for the runtime to run. It runs on the guest's goroutine, inside
// the proc_fork import. ended is called when the child's program has ended,
// before its status is there for a wait to take.
func (p *Perl) forkGuest(_ *base.Module, childWASI *base.WasiStubs, pid int32) (child *base.Module, ended func(status int32), err error) {
	if p.closed.Load() {
		return nil, nil, ErrClosed
	}
	var from *base.SharedImage
	if p.image != nil {
		from = p.image.img
	}
	dup, err := base.ForkMemory(p.m.g, from, memoryCeiling(p.opts, p.m.g.MemSize.Load()))
	if err != nil {
		return nil, nil, fmt.Errorf("perl: duplicate the forking instance's memory: %w", err)
	}
	image := p.image
	if dup.Image != from {
		image = &memoryImage{img: dup.Image}
		if dup.Image == nil {
			image = nil
		}
		if dup.Rebased {
			p.image.release()
			p.image = image.retain()
		}
	}

	m := &Module{}
	c := &Perl{m: m, wasi: childWASI, opts: p.opts, mapped: dup.Memory, image: image.retain(), h: p.h, intrAddr: p.intrAddr, family: p.family}
	m.g = wasm2go.NewFromSnapshot(childWASI, envStubs{m: m}, wasmifyStubs{m: m}, dup.Memory, dup.Size, dup.Globals)
	c.writeMemoryBase(true)
	// The copied fd table came with the parent's hook; the child's forks
	// must snapshot the child.
	c.installForkHook()
	if p.OnFork != nil {
		p.OnFork(c)
	}
	if err := p.handOver(c); err != nil {
		c.discard()
		return nil, nil, err
	}
	if !p.family.start(c, pid) {
		c.discard()
		return nil, nil, errFamilyOver
	}
	// The runtime's run of the child is the one entry into its module.
	m.mu.Lock()
	return m.g, func(int32) {
		m.mu.Unlock()
		// Whichever way the program ended, no interpreter is left to
		// finalize: only the host's side of the instance remains.
		c.discard()
		p.family.ended(c)
	}, nil
}

// ForkChild reports whether the instance is a pseudo-process: one that a
// fork of another instance made (forkGuest), whose family's first process
// is that other instance or one it was forked from.
func (p *Perl) ForkChild() bool {
	return p.family != nil && p.family.origin != p
}

// signalForkChild delivers sig to a running fork child, for the end of the
// family. A signal is posted to the child's guest, which
// handles it like any process does: its %SIG handler runs, or the signal is
// ignored, or its default action ends the child. SIGKILL is not the guest's
// to act on: it stops the child at its next op, even one in a busy loop
// that never reaches a point where signals are noticed.
func (c *Perl) signalForkChild(sig int32) error {
	if c.closed.Load() {
		return nil // already finished; like signalling a zombie
	}
	if sig != sigKILL {
		c.wasi.PostSignal(sig)
		return nil
	}
	if !c.killed.CompareAndSwap(0, sig) {
		return nil // already dying of an earlier signal
	}
	// Posted as well, so that a child waiting in a blocking call wakes up
	// and meets its end.
	c.wasi.PostSignal(sig)
	base.AccessMemory(c.m.g, func(mem []byte) {
		// The child may have ended since it was found running above — of
		// the signal just posted, even — and given its memory back. It is
		// marked closed before that, and the memory goes under the lock
		// held here (releaseMapping): a child that is not closed still
		// has it.
		if c.closed.Load() {
			return
		}
		binary.LittleEndian.PutUint32(mem[c.intrAddr:], interruptSignal|uint32(sig))
	})
	return nil
}

// discard drops an instance without running the interpreter's teardown:
// the end of a child that was killed, left through _exit, or never started.
func (c *Perl) discard() {
	if !c.closed.CompareAndSwap(false, true) {
		return
	}
	c.releaseHostState()
}
