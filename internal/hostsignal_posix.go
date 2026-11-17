//go:build darwin || linux

package internal

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

// hostSignals maps the guest's signal numbers — Linux's on every host, as
// wasi-libc defines them — to the host's signals.
var hostSignals = map[int]syscall.Signal{
	1:  syscall.SIGHUP,
	2:  syscall.SIGINT,
	3:  syscall.SIGQUIT,
	4:  syscall.SIGILL,
	5:  syscall.SIGTRAP,
	6:  syscall.SIGABRT,
	7:  syscall.SIGBUS,
	8:  syscall.SIGFPE,
	9:  syscall.SIGKILL,
	10: syscall.SIGUSR1,
	11: syscall.SIGSEGV,
	12: syscall.SIGUSR2,
	13: syscall.SIGPIPE,
	14: syscall.SIGALRM,
	15: syscall.SIGTERM,
	17: syscall.SIGCHLD,
	18: syscall.SIGCONT,
	19: syscall.SIGSTOP,
	20: syscall.SIGTSTP,
	21: syscall.SIGTTIN,
	22: syscall.SIGTTOU,
	23: syscall.SIGURG,
	24: syscall.SIGXCPU,
	25: syscall.SIGXFSZ,
	26: syscall.SIGVTALRM,
	27: syscall.SIGPROF,
	28: syscall.SIGWINCH,
	29: syscall.SIGIO,
	31: syscall.SIGSYS,
}

// DieOfSignal ends the host process the way the guest's signal sig ends a
// process: by that signal, so that whoever waits for the process sees a
// signal death. It returns only when that is not possible — the host has no
// such signal, the process was started with it ignored, or there is no
// shell — and the caller then falls back to an exit status.
//
// A Go program cannot simply send itself the signal: the Go runtime keeps
// its own handler installed for most signals and carries on after them,
// and offers no way back to the default action. An exec does restore the
// default action of every caught signal, and keeps the process — its pid,
// its parent, its open files. So the process becomes a shell whose only
// act is to send itself the signal (and, should it survive that, to exit
// with the status shells report for the signal).
func DieOfSignal(sig int) {
	hostSig, known := hostSignals[sig]
	if !known || signal.Ignored(hostSig) {
		return
	}
	const shell = "/bin/sh"
	number := strconv.Itoa(int(hostSig))
	script := "kill -" + number + " $$; exit " + strconv.Itoa(128+int(hostSig))
	_ = syscall.Exec(shell, []string{"sh", "-c", script}, os.Environ())
}

// hostChildEnded sends the host process the SIGCHLD of the end of a child
// that the kernel does not know of: a pseudo-process of the program that is
// the host process.
func hostChildEnded() {
	_ = syscall.Kill(os.Getpid(), syscall.SIGCHLD)
}
