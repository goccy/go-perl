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
// it is entered through one bridge call (perl_fork_child) that resumes the
// program right after the forking op and returns when the child's program
// ends. Its exit status is published to the parent's process table, where
// waitpid finds it like any spawned process's.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

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
// state and starts it. It runs on the guest's goroutine, inside the
// proc_fork import.
func (p *Perl) forkGuest(_ *base.Module, childWASI *base.WasiStubs, pid int32, exited func(status int32)) (kill func(sig int32) error, err error) {
	if p.closed.Load() {
		return nil, ErrClosed
	}
	var from *base.SharedImage
	if p.image != nil {
		from = p.image.img
	}
	dup, err := base.ForkMemory(p.m.g, from, memoryCeiling(p.opts, p.m.g.MemSize.Load()))
	if err != nil {
		return nil, fmt.Errorf("perl: duplicate the forking instance's memory: %w", err)
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
		return nil, err
	}

	if !p.family.start(c, pid, exited) {
		c.discard()
		return nil, errFamilyOver
	}
	return c.signalForkChild, nil
}

// runForkChild runs the child's program to its end, tears the instance
// down, and returns the POSIX wait status to report to the parent.
//
// The bridge call covers the child's whole life as a process: its END
// blocks, the flush of PerlIO and global destruction run inside it, and its
// result is the status all of that leaves. A destructor that exits, calls
// _exit or is killed during global destruction therefore ends the call the
// way the same thing does anywhere else in the program.
func (c *Perl) runForkChild(pid int32) (status int32) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, c.h)
	buf = pbAppendInt32(buf, 2, pid)
	resp, err := c.m.invoke(0, midForkChild, buf, wasm2go.Inv_0_9)
	if err == nil {
		err = pbExtractError(resp)
	}

	exitCode := 255
	var raw *base.WasiExitError
	var killed *base.WasiSignalError
	switch {
	case errors.As(err, &killed):
		// The default action of a signal the child neither handles nor
		// ignores.
		c.killed.CompareAndSwap(0, killed.Signal)
	case errors.As(err, &raw):
		exitCode = int(raw.Code)
	case err != nil:
		fmt.Fprintf(c.stderr(), "perl: fork child %d: %v\n", pid, err)
	default:
		if code := readScalarAtField(resp, 1, (*pbReader).readUint32); code != math.MaxUint32 {
			exitCode = int(code)
		} else {
			fmt.Fprintf(c.stderr(), "perl: fork child %d: the instance holds no pending fork\n", pid)
		}
	}
	// Whichever way the call ended, no interpreter is left to finalize:
	// only the host's side of the instance remains.
	c.discard()
	if sig := c.killed.Load(); sig != 0 {
		return sig & 0x7f
	}
	return int32(exitCode&0xff) << 8
}

// signalForkChild delivers sig to a running fork child: the proc_kill path
// for a pseudo-process. A signal is posted to the child's guest, which
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
