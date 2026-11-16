package internal

import "runtime"

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
// goroutine waits; an entry made on the thread itself (by native code, or
// by a Go function the program called) runs where it is. Instances without
// native code run on their callers' goroutines, and so does a fork child,
// which runs no native code (see the xs package's fork child).
type instanceThread struct {
	// tid is the thread's id (threadID): an entry made with it runs in
	// place. Which thread a goroutine runs on can be learnt nowhere else.
	tid uint64
	// calls carries entries from other goroutines to the thread.
	calls chan func()
	// quit ends the thread's goroutine (Close); entries made after it run
	// on their callers' goroutines, on an instance that is closed.
	quit chan struct{}
}

// KeepOnOneThread makes every later entry into the guest run on one thread
// of the process (see instanceThread); calling it again does nothing. Where
// the thread a goroutine runs on cannot be told (threadIDs is false) it
// does nothing either.
func (p *Perl) KeepOnOneThread() {
	if !threadIDs || p.thread.Load() != nil {
		return
	}
	t := &instanceThread{calls: make(chan func()), quit: make(chan struct{})}
	ready := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		t.tid = threadID()
		close(ready)
		for {
			select {
			case fn := <-t.calls:
				fn()
			case <-t.quit:
				// The thread goes back to the runtime rather than ending
				// with the goroutine: a thread made through the C library
				// (purego's runtime support on Linux) does not survive the
				// runtime ending it (the process died of SIGSEGV).
				runtime.UnlockOSThread()
				return
			}
		}
	}()
	<-ready
	if !p.thread.CompareAndSwap(nil, t) {
		close(t.quit) // another caller's thread won
	}
}

// OnThread runs fn on the instance's thread, if it has one, and waits for
// it; a panic of fn is the caller's.
func (p *Perl) OnThread(fn func()) {
	t := p.thread.Load()
	if t == nil || threadID() == t.tid {
		fn()
		return
	}
	var recovered any
	done := make(chan struct{})
	call := func() {
		defer func() {
			recovered = recover()
			close(done)
		}()
		fn()
	}
	select {
	case t.calls <- call:
		<-done
	case <-t.quit:
		fn()
		return
	}
	if recovered != nil {
		panic(recovered)
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
