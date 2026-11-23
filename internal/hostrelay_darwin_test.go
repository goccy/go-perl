package internal

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// TestSignalRelaySendsTheHostTheRootsChildEnd pins the host process's
// SIGCHLD for a child of the root that ended, which the guest has as well:
// native code's handler of it is not read on this system.
func TestSignalRelaySendsTheHostTheRootsChildEnd(t *testing.T) {
	chld := make(chan os.Signal, 1)
	signal.Notify(chld, syscall.SIGCHLD)
	defer signal.Stop(chld)
	root := bindFake(t)
	if taken, err := hostRelay.TakeSignal(root, guestSIGCHLD); taken || err != nil {
		t.Fatalf("the root's child end: taken %v (%v), want not", taken, err)
	}
	select {
	case <-chld:
	case <-time.After(10 * time.Second):
		t.Fatal("the host process was sent no SIGCHLD")
	}
}
