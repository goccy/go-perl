//go:build darwin || linux

package internal

// The children of the host process, for a program that is it
// (InstanceOptions.HostProcess): the guest knows the commands it starts by
// their host pids, and waits for the children native code started.

import (
	"errors"
	"os"
	"syscall"
	"unsafe"

	"github.com/goccy/perlwasm2go/base"
)

// hostChildren is the Children of a program that is the host process.
type hostChildren struct{}

// hostChildrenConfig returns the Children of a program that is the host
// process.
func hostChildrenConfig() base.Children { return hostChildren{} }

// CommandPID has the guest know every command it starts by its host pid,
// as a process knows its children.
func (hostChildren) CommandPID(base.ExecRequest, int) bool { return true }

// Adopt adopts pid when it is a child of the host process that has not
// been reaped: one native code started. Its end is learnt without reaping
// it, so that its pid names no other process until the guest has collected
// its status; release reaps it. Any process of the family may adopt it:
// native code of every pseudo-process runs in the host process, whose
// child it is, and the runtime gives it to the first that waits.
func (hostChildren) Adopt(_ base.SignalTarget, pid int32, exited func(status int32)) (func(sig int32) error, func() error, bool, error) {
	status, ended, err := awaitChild(pid, true)
	switch {
	case errors.Is(err, syscall.ECHILD):
		return nil, nil, false, nil
	case err != nil:
		return nil, nil, false, err
	}
	// failed is the error of the wait for the child's end, which release
	// returns: the child cannot be reaped then. Written before exited is
	// called, which comes before release.
	var failed error
	if ended {
		exited(status)
	} else {
		go func() {
			status, _, err := awaitChild(pid, false)
			if err != nil {
				// The child is gone without a status: all that can be
				// said is that it ended, as the runtime says of a
				// command whose wait failed.
				failed, status = err, 127<<8
			}
			exited(status)
		}()
	}
	kill := func(sig int32) error { return hostRelay.Kill(pid, sig) }
	release := func() error {
		if failed != nil {
			return failed
		}
		return reapChild(pid)
	}
	return kill, release, true, nil
}

// awaitChild waits for the host process's child pid to end, without
// reaping it (waitid(2) with WNOWAIT), and returns its wait status in the
// guest's numbers. With nohang it does not wait: ended reports whether the
// child has ended. ECHILD is no unreaped child of that pid.
func awaitChild(pid int32, nohang bool) (status int32, ended bool, err error) {
	const pPID = 1 // P_PID, of both Linux and darwin
	options := syscall.WEXITED | syscall.WNOWAIT
	if nohang {
		options |= syscall.WNOHANG
	}
	for {
		var info childSiginfo
		_, _, errno := syscall.Syscall6(syscall.SYS_WAITID, pPID, uintptr(pid), uintptr(unsafe.Pointer(&info)), uintptr(options), 0, 0)
		switch errno {
		case 0:
			// With WNOHANG, a child that has not ended leaves the
			// siginfo zeroed (si_pid 0).
			if info.pid == 0 {
				return 0, false, nil
			}
			return info.guestStatus(), true, nil
		case syscall.EINTR:
			continue
		}
		return 0, false, os.NewSyscallError("waitid", errno)
	}
}

// reapChild reaps the host process's child pid, which has ended.
func reapChild(pid int32) error {
	var status syscall.WaitStatus
	for {
		_, err := syscall.Wait4(int(pid), &status, 0, nil)
		if err != syscall.EINTR {
			return os.NewSyscallError("wait4", err)
		}
	}
}

// guestStatus returns the wait status, in the guest's numbers, of a child
// the siginfo tells the end of: how it ended (si_code: CLD_EXITED 1,
// CLD_KILLED 2, CLD_DUMPED 3, the same on Linux and darwin) and its exit
// status or signal (si_status). A signal the guest has no number for is
// given as it is.
func (info *childSiginfo) guestStatus() int32 {
	const cldExited = 1
	if info.code == cldExited {
		return (info.status & 0xff) << 8
	}
	if sig, known := guestSignals[syscall.Signal(info.status&0x7f)]; known {
		return int32(sig)
	}
	return info.status & 0x7f
}
