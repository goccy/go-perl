//go:build darwin || linux

package xs

// Signals interrupt native code.
//
// In perl a signal interrupts the process wherever it is: a module's C code
// waiting in a system call (a read, mq_receive, a lock) sees it fail with
// EINTR, returns, and perl runs the %SIG handler at its next safe point.
// Here a signal for the program is the runtime's to deliver: it ends the
// program's own blocking calls, but a system call that native code makes on
// the instance's thread is the kernel's and would wait on. So for as long
// as native code runs, the runtime is told how to interrupt it: send the
// thread running it a signal of the host's whose handler is installed
// without SA_RESTART, so that the system call in progress fails with EINTR
// (nativeInterruptSignal, made ready once by prepareNativeInterrupt).
// The signal's own handler does nothing for it; the program's signal is
// delivered by the runtime as always, when the native code has returned.

// enterNative is called on the instance's goroutine, with its OS thread
// locked, before native code runs for the guest; leaveNative is called
// with what it returned once that code has returned. Calls nest (native
// code can call back into the program, which can call native code again).
func (s *state) enterNative() (prevThread uint64, prevInterrupt func(), entered bool) {
	thread, ok := nativeThreadSelf()
	if !ok {
		return 0, nil, false
	}
	prevThread = s.nativeThread.Swap(thread)
	return prevThread, s.p.SetOutsideInterrupt(s.interrupt), true
}

func (s *state) leaveNative(prevThread uint64, prevInterrupt func(), entered bool) {
	if !entered {
		return
	}
	// Taken back before the thread is forgotten: once SetOutsideInterrupt
	// has returned, no signal can be sent to the thread for this call.
	s.p.SetOutsideInterrupt(prevInterrupt)
	s.nativeThread.Store(prevThread)
}

// interruptNative is what a signal for the program does while native code
// runs: the thread that runs it is sent nativeInterruptSignal.
func (s *state) interruptNative() {
	if thread := s.nativeThread.Load(); thread != 0 {
		interruptNativeThread(thread)
	}
}
