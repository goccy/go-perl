//go:build darwin || linux

package gperl_test

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// A write in blocking mode to a terminal nobody reads takes what there is
// room for and waits for room for the rest. A signal ends the wait as it
// ends the system call in perl: syswrite returns what it wrote, and the
// handler runs (IO::Pty::Easy's write with a time limit is an alarm around
// such a write).
func TestBlockingWriteToATerminalIsEndedByASignal(t *testing.T) {
	cmd := perlProcess("-e", ptyPrelude+`
my ($master, $path) = open_pty();
if (!$master) { print "no pseudo-terminal\n"; exit 0 }
open(my $tty, '+<', $path) or die "open $path: $!";
# The terminal keeps what it is sent until it is read, rather than dropping
# what does not fit in a line.
my $t = POSIX::Termios->new;
$t->getattr(fileno $tty) or die "getattr: $!";
$t->setlflag($t->getlflag & ~(POSIX::ICANON() | POSIX::ECHO()));
$t->setattr(fileno $tty, POSIX::TCSANOW()) or die "setattr: $!";
my $fired = 0;
$SIG{ALRM} = sub { $fired++ };
alarm 1;
my $n = syswrite($master, "x" x (1 << 20));
my $err = $!;
alarm 0;
print defined $n ? ($n > 0 && $n < (1 << 20) ? "wrote a part" : "wrote $n bytes") : "failed: $err",
    ", the handler ran $fired time(s)\n";
`)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("perl: %v\nstdout: %s\nstderr: %s", err, out.String(), errOut.String())
		}
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		<-done
		t.Fatalf("the write was not ended by the signal in 30 s (stdout %q, stderr %q)", out.String(), errOut.String())
	}
	got := out.String()
	if strings.TrimSpace(got) == "no pseudo-terminal" {
		t.Skip("the system gives no pseudo-terminal")
	}
	if want := "wrote a part, the handler ran 1 time(s)\n"; got != want {
		t.Fatalf("got %q, want %q (stderr %q)", got, want, errOut.String())
	}
}
