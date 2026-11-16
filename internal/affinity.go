package internal

import (
	"runtime"
	"sync/atomic"
	"time"
)

// The thread of an instance that runs native code.
//
// Native modules keep state per thread of the process: thread-local
// variables of their own and of the libraries they use (OpenSSL's error
// queue, glibc's malloc arena of the thread). On perl every call of a
// module runs on the program's one thread, so state one call leaves is
// there for the next. A goroutine runs on whichever thread the Go scheduler
// gives it, from one call to the next. An instance that runs native code
// therefore runs all of its guest code on one thread for its life
// (KeepOnOneThread): a goroutine of its own, locked to its thread, runs
// every entry into the guest that another goroutine makes, while that
// goroutine waits; an entry made on the thread itself (by native code) runs
// where it is. A Go function the program calls runs on a goroutine of its
// own while the thread serves the entries other goroutines make (callGo):
// the function may wait for a goroutine that calls the instance. Instances
// without native code run on their callers' goroutines, and so does a fork
// child, which runs no native code (see the xs package's fork child).
type instanceThread struct {
	// tid is the thread's id (threadID): an entry made with it runs in
	// place. Which thread a goroutine runs on can be learnt nowhere else.
	tid uint64
	// calls carries entries from other goroutines to the thread. It holds
	// a few, so that a goroutine hands one over without waiting for the
	// thread to take it.
	calls chan *threadCall
	// quit ends the thread's goroutine (Close); entries made after it run
	// on their callers' goroutines, on an instance that is closed.
	quit chan struct{}
}

// threadCall is one entry another goroutine hands to the instance's thread.
type threadCall struct {
	fn func()
	// state says whether the thread took the entry (callTaken) or its
	// goroutine withdrew it to run it itself (callWithdrawn), which it does
	// when the thread has ended without taking it: which of the two came
	// first is known only here.
	state atomic.Int32
	// done is closed when fn has returned on the thread, recovered then
	// holding what it panicked with.
	done      chan struct{}
	recovered any
}

const (
	callQueued int32 = iota
	callTaken
	callWithdrawn
)

// handOffSpin is how long each side of a hand-off looks for the other
// before it waits. A goroutine locked to its thread that waits puts the
// thread to sleep, and the goroutine that hands it an entry then has to wake
// the thread through the system, and the thread the caller when the entry
// is done, which takes longer than an entry that evaluates a constant (tens
// of microseconds on darwin). So the thread looks for the next entry this
// long after each one, and the caller for the end of its entry while the
// thread runs it.
//
// Spinning pays only when the two sides run at once. Where the processors
// are busy (other programs, other goroutines), the side that spins holds
// the processor the other side needs, and every hand-off then costs the
// whole spin: the caller spins only while the thread is running its entry
// (it took it), and the thread only while its spins keep finding entries
// (see serve). With one processor for Go code neither side spins.
const handOffSpin = 50 * time.Microsecond

// spinning reports whether the two sides of a hand-off spin (handOffSpin).
func spinning() bool { return runtime.GOMAXPROCS(0) > 1 }

// KeepOnOneThread makes every later entry into the guest run on one thread
// of the process (see instanceThread); calling it again does nothing. Where
// the thread a goroutine runs on cannot be told (threadIDs is false) it
// does nothing either.
func (p *Perl) KeepOnOneThread() {
	if !threadIDs || p.thread.Load() != nil {
		return
	}
	t := &instanceThread{calls: make(chan *threadCall, 16), quit: make(chan struct{})}
	ready := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		t.tid = threadID()
		close(ready)
		t.serve()
		// The thread goes back to the runtime rather than ending with the
		// goroutine: a thread made through the C library (purego's runtime
		// support on Linux) does not survive the runtime ending it (the
		// process died of SIGSEGV).
		runtime.UnlockOSThread()
	}()
	<-ready
	if !p.thread.CompareAndSwap(nil, t) {
		close(t.quit) // another caller's thread won
	}
}

// serve runs the entries handed to the thread until quit; those handed over
// before quit was seen run first.
//
// The thread spins for the next entry while spinning finds one: credit
// counts the recent spins that did (a few at most), and a spin that finds
// none takes credit away; with none left the thread waits for its entries
// and tries a spin again after a while (probeAfter waits).
func (t *instanceThread) serve() {
	const maxCredit, probeAfter = 4, 64
	credit, waits := maxCredit, 0
	for {
		select {
		case c := <-t.calls:
			c.run()
			continue
		default:
		}
		if credit > 0 || waits >= probeAfter {
			waits = 0
			if c := t.spinForCall(); c != nil {
				credit = min(credit+1, maxCredit)
				c.run()
				continue
			}
			credit = max(credit-2, 0)
		} else {
			waits++
		}
		select {
		case c := <-t.calls:
			c.run()
		case <-t.quit:
			for {
				select {
				case c := <-t.calls:
					c.run()
				default:
					return
				}
			}
		}
	}
}

// spinForCall looks for an entry for handOffSpin without waiting, and
// returns it, or nil.
func (t *instanceThread) spinForCall() *threadCall {
	if !spinning() {
		return nil
	}
	start := time.Now()
	for i := 1; ; i++ {
		select {
		case c := <-t.calls:
			return c
		default:
		}
		if i%64 == 0 && time.Since(start) > handOffSpin {
			return nil
		}
	}
}

// run runs the entry on the thread, unless its goroutine withdrew it.
func (c *threadCall) run() {
	if !c.state.CompareAndSwap(callQueued, callTaken) {
		return
	}
	defer close(c.done)
	defer func() { c.recovered = recover() }()
	c.fn()
}

// OnThread runs fn on the instance's thread, if it has one, and waits for
// it; a panic of fn is the caller's.
func (p *Perl) OnThread(fn func()) {
	t := p.thread.Load()
	if t == nil || threadID() == t.tid {
		fn()
		return
	}
	c := &threadCall{fn: fn, done: make(chan struct{})}
	select {
	case t.calls <- c:
	case <-t.quit:
		fn()
		return
	}
	if !c.spinForDone() {
		select {
		case <-c.done:
		case <-t.quit:
			if c.state.CompareAndSwap(callQueued, callWithdrawn) {
				fn() // the thread ended without taking it
				return
			}
			<-c.done
		}
	}
	if c.recovered != nil {
		panic(c.recovered)
	}
}

// spinForDone looks for the end of the entry without waiting, while the
// thread runs it, for handOffSpin at most, and reports whether it saw it.
func (c *threadCall) spinForDone() bool {
	if !spinning() {
		return false
	}
	start := time.Now()
	for i := 1; ; i++ {
		select {
		case <-c.done:
			return true
		default:
		}
		if i%64 == 0 && (c.state.Load() != callTaken || time.Since(start) > handOffSpin) {
			return false
		}
	}
}

// callGo runs a call the program makes of a Go function (the Perl->Go
// bridge, UserHandler). On the instance's thread the function runs on a
// goroutine of its own, and the thread serves the entries other goroutines
// hand it until the function returns: the function may wait for another
// goroutine that calls the instance, as it may on an instance without a
// thread, where that call runs nested in the program's call of the
// function. The function's own entries into the instance are handed to the
// thread like any goroutine's, so native code stays on the thread. A panic
// of the function is the thread's, as it would be the caller's.
func (p *Perl) callGo(methodID int32, req []byte) ([]byte, error) {
	t := p.thread.Load()
	if t == nil || threadID() != t.tid {
		return p.UserHandler(methodID, req)
	}
	type outcome struct {
		resp      []byte
		err       error
		recovered any
	}
	done := make(chan outcome, 1)
	go func() {
		var o outcome
		defer func() {
			o.recovered = recover()
			done <- o
		}()
		o.resp, o.err = p.UserHandler(methodID, req)
	}()
	for {
		select {
		case c := <-t.calls:
			c.run()
		case o := <-done:
			if o.recovered != nil {
				panic(o.recovered)
			}
			return o.resp, o.err
		}
	}
}

// invoke is m.invoke on the instance's thread (see instanceThread).
func (p *Perl) invoke(svc, mid int32, req []byte, call invoker) (resp []byte, err error) {
	if t := p.thread.Load(); t == nil || threadID() == t.tid {
		return p.m.invoke(svc, mid, req, call)
	}
	p.OnThread(func() { resp, err = p.m.invoke(svc, mid, req, call) })
	return resp, err
}

// endThread ends the instance's thread, after its last entry (Close).
func (p *Perl) endThread() {
	if t := p.thread.Load(); t != nil {
		close(t.quit)
	}
}

// HasThread reports whether the instance runs its guest code on a thread
// of its own (KeepOnOneThread).
func (p *Perl) HasThread() bool { return p.thread.Load() != nil }
