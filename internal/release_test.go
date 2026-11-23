package internal

import (
	"testing"
	"time"

	goperlfs "github.com/goccy/go-perl/fs"
	"github.com/goccy/perlwasm2go/base"
)

// An instance gives its memory back only when no other goroutine is inside
// it. The host sets flags in an instance's memory from other goroutines —
// the kill of a fork child sets the child's interrupt word — under the lock
// of base.AccessMemory, and a child that is killed while it is ending
// releases its memory at that very time: without the lock around the
// release the flag was written into memory that was gone, or that the next
// instance had been given.
func TestMemoryIsReleasedUnderTheMemoryLock(t *testing.T) {
	// The host's filesystem has the /dev/null perl needs.
	p, err := New(InstanceOptions{FS: goperlfs.NewHostFS()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.mapped == nil {
		p.Close()
		t.Skip("the instance does not live in a mapping of its own")
	}
	inside, leave, left := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(left)
		base.AccessMemory(p.m.g, func(mem []byte) {
			close(inside)
			<-leave
			// what signalForkChild does: the instance is not closed, so
			// its memory is there
			if !p.closed.Load() {
				_ = mem[p.intrAddr]
			}
		})
	}()
	<-inside
	closed := make(chan struct{})
	go func() {
		defer close(closed)
		p.Close()
	}()
	select {
	case <-closed:
		t.Error("the instance was closed, and its memory released, while another goroutine was inside the memory")
	case <-time.After(500 * time.Millisecond):
	}
	close(leave)
	<-left
	<-closed
}
