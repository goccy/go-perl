//go:build linux && (amd64 || arm64 || riscv64 || loong64 || ppc64 || ppc64le || s390x)

package xs

import (
	"sync"
	"syscall"
	"unsafe"
)

// nativeInterruptSignal is the host signal that interrupts native code:
// the last real-time signal, which neither the Go runtime nor the C
// library takes for itself, and which no signal of the program is relayed
// as.
const nativeInterruptSignal = 64

// kernelSigaction is struct sigaction as the kernel's rt_sigaction takes it
// on these architectures.
type kernelSigaction struct {
	handler  uintptr
	flags    uint64
	restorer uintptr
	mask     uint64
}

const saRestart = 0x10000000

// nativeInterruptReady reports whether nativeInterruptSignal can interrupt
// a system call; the first call makes it so.
//
// The Go runtime has its handler installed for the signal, with
// SA_RESTART: a system call it interrupts would be restarted, which is
// what the runtime wants of the signals it handles itself. For this signal
// it handles nothing (nobody asks for it, and the handler returns at once),
// so the flag is taken off its installation: the handler stays, and a
// system call the signal interrupts fails with EINTR. A signal without the
// runtime's handler is not used: its default action ends the process.
var nativeInterruptReady = sync.OnceValue(func() bool {
	var act kernelSigaction
	if _, _, errno := syscall.RawSyscall6(syscall.SYS_RT_SIGACTION, nativeInterruptSignal,
		0, uintptr(unsafe.Pointer(&act)), unsafe.Sizeof(act.mask), 0, 0); errno != 0 {
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
	_, _, errno := syscall.RawSyscall6(syscall.SYS_RT_SIGACTION, nativeInterruptSignal,
		uintptr(unsafe.Pointer(&act)), 0, unsafe.Sizeof(act.mask), 0, 0)
	return errno == 0
})

// nativeThreadSelf returns the thread the caller runs on (its OS thread
// locked), for interruptNativeThread; ok is false where native code cannot
// be interrupted.
func nativeThreadSelf() (thread uint64, ok bool) {
	if !nativeInterruptReady() {
		return 0, false
	}
	return uint64(syscall.Gettid()), true
}

// interruptNativeThread sends nativeInterruptSignal to the thread.
func interruptNativeThread(thread uint64) {
	_ = syscall.Tgkill(syscall.Getpid(), int(thread), syscall.Signal(nativeInterruptSignal))
}
