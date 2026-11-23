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
// Memory that is a mapping of the runtime's is duplicated by
// base.ForkMemory. It is copy-on-write where the runtime can tell which
// pages a process wrote (Linux): the first fork of an instance writes the
// pages it has written to a memfd, which parent and child then map
// privately, and a later fork copies the pages written since. Everywhere
// else it is a copy: on darwin and mips, under vm.overcommit_memory=2,
// with an RLIMIT_FSIZE below the memory, and with a System V segment
// attached over the memory. Memory of the Go heap is copied (heapFork).
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

// forker is the Forker of a fork family (WASIConfig.Fork; without one the
// import reports ENOSYS and fork returns undef). Every instance of the
// family has it, the children too (a forked instance has its parent's
// configuration): a fork is made by the instance that forks (forkGuest).
type forker struct{ family *forkFamily }

func (f forker) Fork(parent *base.Module, child *base.WasiStubs, pid int32) (*base.Module, func(status int32), error) {
	p := f.family.member(parent)
	if p == nil {
		return nil, nil, errors.New("perl: fork of an instance that is not of the fork family")
	}
	return p.forkGuest(parent, child, pid)
}

// forkGuest builds the child instance from p's current state, for the
// runtime to run. It runs on the guest's goroutine, inside the proc_fork
// import. ended is called when the child's program has ended, before its
// status is there for a wait to take.
func (p *Perl) forkGuest(_ *base.Module, childWASI *base.WasiStubs, pid int32) (child *base.Module, ended func(status int32), err error) {
	if p.closed.Load() {
		return nil, nil, ErrClosed
	}
	m := &Module{}
	c := &Perl{m: m, wasi: childWASI, opts: p.opts, h: p.h, intrAddr: p.intrAddr, family: p.family}
	if p.mapped != nil {
		mem, size, globals, err := base.ForkMemory(p.m.g)
		if err != nil {
			return nil, nil, fmt.Errorf("perl: duplicate the forking instance's memory: %w", err)
		}
		c.mapped = mem
		m.g = wasm2go.NewFromSnapshot(childWASI, envStubs{m: m}, wasmifyStubs{m: m}, mem, size, globals)
	} else {
		m.g = heapFork(p.m.g, childWASI, m, memoryCeiling(p.opts, p.m.g.MemSize.Load()))
		c.onHeap = p.onHeap
	}
	c.writeMemoryBase(true)
	if p.OnFork != nil {
		p.OnFork(c)
	}
	if err := p.handOver(c); err != nil {
		return nil, nil, errors.Join(err, c.discard())
	}
	if !p.family.start(c, pid) {
		return nil, nil, errors.Join(errFamilyOver, c.discard())
	}
	// The runtime's run of the child is the one entry into its module.
	m.mu.Lock()
	return m.g, func(int32) {
		m.mu.Unlock()
		// Whichever way the program ended, no interpreter is left to
		// finalize: only the host's side of the instance remains.
		if err := c.discard(); err != nil {
			p.family.fail(failChildEnd, fmt.Errorf("pid %d: %w", pid, err))
		}
		p.family.ended(c)
	}, nil
}

// heapFork returns the child's module of a fork of parent, whose memory is
// of the Go heap (bootPrivate): a copy of the memory, as far as its size,
// and of the globals, made while parent is stopped in its fork; the child's
// memory is of the Go heap too, with the parent's room (it grows in place
// as far as the parent's does) and the parent's ceiling.
func heapFork(parent *base.Module, childWASI *base.WasiStubs, m *Module, ceiling int) *base.Module {
	var memory []byte
	base.AccessMemory(parent, func(mem []byte) {
		memory = make([]byte, parent.MemSize.Load(), cap(mem))
		copy(memory, mem)
	})
	g := wasm2go.NewFromSnapshot(childWASI, envStubs{m: m}, wasmifyStubs{m: m}, memory, uint64(len(memory)), base.SaveGlobals(parent))
	wasm2go.SetMaxMemory(g, uint64(ceiling))
	return g
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
func (c *Perl) discard() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	return c.releaseHostState()
}
