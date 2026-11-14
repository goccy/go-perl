//go:build darwin || linux

package xs_test

import (
	"context"
	"fmt"
	"syscall"
	"testing"
	"time"
)

// A signal for the program interrupts a system call native code is waiting
// in, as a signal does in perl: alarm while a module's XSUB reads a pipe
// nobody writes to ends the read with EINTR, the XSUB returns, and the
// handler runs. A read that a signal does not interrupt is not disturbed by
// the alarm set before it.
func TestSignalInterruptsNativeBlockingCall(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	var fds [2]int
	if err := syscall.Pipe(fds[:]); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		syscall.Close(fds[0])
		syscall.Close(fds[1])
	})
	evalWithin := func(code string, release func()) string {
		t.Helper()
		done := make(chan string, 1)
		go func() {
			r, err := p.Eval(context.Background(), code)
			switch {
			case err != nil:
				done <- "Eval: " + err.Error()
			case r.Error != nil:
				done <- "died: " + r.Error.Error()
			default:
				done <- resultStr(r)
			}
		}()
		select {
		case got := <-done:
			return got
		case <-time.After(20 * time.Second):
			release()
			got := <-done
			t.Fatalf("the signal did not interrupt the native call (released after 20 s: %s)", got)
			return ""
		}
	}
	release := func() { syscall.Write(fds[1], []byte("released")) }

	got := evalWithin(fmt.Sprintf(`
		my $hit = 0;
		local $SIG{ALRM} = sub { $hit++ };
		alarm 1;
		my $got = PerlAPI::fd_read(%d, 16);
		my $eintr = $!{EINTR} ? "EINTR" : "errno " . ($! + 0);
		alarm 0;
		(defined $got ? "read [$got]" : "undef") . ":$eintr:$hit"`, fds[0]), release)
	if want := "undef:EINTR:1"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	// Data that is there is read; a signal that came and went before the
	// call does not end it.
	if _, err := syscall.Write(fds[1], []byte("data")); err != nil {
		t.Fatal(err)
	}
	got = evalWithin(fmt.Sprintf(`
		my $hit = 0;
		local $SIG{USR1} = sub { $hit++ };
		kill 'USR1', $$;
		my $got = PerlAPI::fd_read(%d, 16);
		(defined $got ? "read [$got]" : "undef") . ":$hit"`, fds[0]), release)
	if want := "read [data]:1"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
