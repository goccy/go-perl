//go:build darwin

package xs

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// nativeInterruptSignal is the host signal that interrupts native code:
// SIGINFO, which the Go runtime does not take for itself and which no
// signal of the program is relayed as (the program's signals are Linux's,
// which has no SIGINFO).
const nativeInterruptSignal = 29

// darwinSigaction is struct sigaction as the C library's sigaction takes
// it.
type darwinSigaction struct {
	handler uintptr
	mask    uint32
	flags   int32
}

const saRestart = 0x0002

// The C library's functions this file calls.
var nativeInterruptLibc struct {
	pthreadSelf, pthreadKill uintptr
}

// nativeInterruptReady reports whether nativeInterruptSignal can interrupt
// a system call; the first call makes it so.
//
// The Go runtime has its handler installed for the signal, with
// SA_RESTART: a system call it interrupts would be restarted, which is
// what the runtime wants of the signals it handles itself. For this signal
// it handles nothing (nobody asks for it, and the handler returns at once),
// so the flag is taken off its installation: the handler stays, and a
// system call the signal interrupts fails with EINTR. A signal without the
// runtime's handler is not used: its default action is not to interrupt.
var nativeInterruptReady = sync.OnceValue(func() bool {
	libc := &nativeInterruptLibc
	var sigaction uintptr
	for _, sym := range []struct {
		name string
		dst  *uintptr
	}{{"pthread_self", &libc.pthreadSelf}, {"pthread_kill", &libc.pthreadKill}, {"sigaction", &sigaction}} {
		fn, err := purego.Dlsym(purego.RTLD_DEFAULT, sym.name)
		if err != nil {
			return false
		}
		*sym.dst = fn
	}
	var act darwinSigaction
	if rc, _, _ := purego.SyscallN(sigaction, nativeInterruptSignal, 0, uintptr(unsafe.Pointer(&act))); int32(rc) != 0 {
		return false
	}
	const sigDFL, sigIGN = 0, 1
	if act.handler == sigDFL || act.handler == sigIGN {
		return false
	}
	if act.flags&saRestart == 0 {
		return true
	}
	act.flags &^= saRestart
	rc, _, _ := purego.SyscallN(sigaction, nativeInterruptSignal, uintptr(unsafe.Pointer(&act)), 0)
	return int32(rc) == 0
})

// nativeThreadSelf returns the thread the caller runs on (its OS thread
// locked), for interruptNativeThread; ok is false where native code cannot
// be interrupted.
func nativeThreadSelf() (thread uint64, ok bool) {
	if !nativeInterruptReady() {
		return 0, false
	}
	self, _, _ := purego.SyscallN(nativeInterruptLibc.pthreadSelf)
	return uint64(self), true
}

// interruptNativeThread sends nativeInterruptSignal to the thread.
func interruptNativeThread(thread uint64) {
	purego.SyscallN(nativeInterruptLibc.pthreadKill, uintptr(thread), nativeInterruptSignal)
}
