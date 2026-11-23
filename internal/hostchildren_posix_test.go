//go:build darwin || linux

package internal

import (
	"errors"
	"syscall"
	"testing"

	"github.com/goccy/perlwasm2go/base"
)

// startHostChild starts /bin/sh -c script as a child of the test process
// that nothing waits for, and returns its pid.
func startHostChild(t *testing.T, script string) int32 {
	t.Helper()
	pid, err := syscall.ForkExec("/bin/sh", []string{"sh", "-c", script}, &syscall.ProcAttr{Files: []uintptr{0, 1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	return int32(pid)
}

// A child of the host process is adopted, its end is told in the guest's
// numbers without reaping it, its kill reaches it, and release reaps it.
// A pid that is no child of the process is not adopted.
func TestHostChildrenAdoptAChildOfTheProcess(t *testing.T) {
	children := hostChildren{}
	pid := startHostChild(t, "exec sleep 1000")
	ended := make(chan int32, 1)
	kill, release, ok, err := children.Adopt(nil, pid, func(status int32) { ended <- status })
	if err != nil || !ok {
		t.Fatalf("Adopt = %v, %v, want the child adopted", ok, err)
	}
	if err := kill(15); err != nil {
		t.Fatalf("kill: %v", err)
	}
	if got := <-ended; got != 15 {
		t.Errorf("status %#x, want killed by SIGTERM", got)
	}
	if _, kept, err := awaitChild(pid, true); err != nil || !kept {
		t.Fatalf("the ended child was reaped before its release: %v", err)
	}
	if err := release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	if _, _, err := awaitChild(pid, true); !errors.Is(err, syscall.ECHILD) {
		t.Errorf("the released child is still there: %v", err)
	}

	exitedPID := startHostChild(t, "exit 4")
	if _, _, err := awaitChild(exitedPID, false); err != nil {
		t.Fatal(err)
	}
	var status int32 = -1
	_, release, ok, err = children.Adopt(nil, exitedPID, func(s int32) { status = s })
	if err != nil || !ok || status != 4<<8 {
		t.Fatalf("Adopt of an ended child = %v, %v, status %#x, want adopted with an exit of 4", ok, err, status)
	}
	if err := release(); err != nil {
		t.Fatalf("release: %v", err)
	}

	if _, _, ok, err := children.Adopt(nil, exitedPID, func(int32) { t.Error("exited called for no child") }); ok || err != nil {
		t.Errorf("Adopt of a reaped pid = %v, %v, want no child", ok, err)
	}
	if !children.CommandPID(base.ExecRequest{}, 1) {
		t.Error("CommandPID = false, want every command known by its host pid")
	}
}
