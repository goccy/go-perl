package perl_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	perl "github.com/goccy/go-perl"
)

// Without Config.HostProcess the program is a process of its own: it has
// nothing of the host process's identity.
func TestIsolatedProcessIdentity(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`print "pid=$$ ppid=", getppid(), " uid=$< euid=$> gid=", $( + 0, " egid=", $) + 0, "\n";`,
		`print "user=", (defined getpwuid($<) ? "found" : "none"), " group=", (defined getgrgid(0) ? "found" : "none"), "\n";`,
		`print "groups=", scalar(() = split ' ', $(), "\n";`,
		`print "kill of an outside pid: ", (kill(0, 1) ? "reached" : ($!{ESRCH} ? "ESRCH" : "error $!")), "\n";`,
	)}, "")
	r.want(t, 0, "pid=42 ppid=1 uid=0 euid=0 gid=0 egid=0\n"+
		"user=none group=none\n"+
		"groups=1\n"+
		"kill of an outside pid: ESRCH\n")
}

// signalRun runs prog and returns RunProgram's results and what the program
// wrote before it ended.
func signalRun(t *testing.T, cfg perl.Config, prog perl.Program) (status int, err error, stdout string) {
	t.Helper()
	var out bytes.Buffer
	cfg.Stdout = &out
	p, newErr := perl.New(cfg)
	if newErr != nil {
		t.Fatalf("New: %v", newErr)
	}
	status, err = p.RunProgram(context.Background(), prog)
	// Close is the end of the process: it flushes what the program left
	// buffered, unless a signal ended it.
	if closeErr := p.Close(); closeErr != nil {
		t.Fatalf("Close: %v", closeErr)
	}
	return status, err, out.String()
}

// A program that a signal ends is reported as such: the run returns a
// *SignalError carrying the signal, and nothing of the program runs after
// the signal — no END block, no flush of buffered output.
func TestRunProgramEndedBySignal(t *testing.T) {
	_, err, stdout := signalRun(t, perl.Config{}, perl.Program{Code: code(
		`END { print "END block ran\n" }`,
		`$| = 1; print "before\n"; $| = 0;`,
		`print "buffered, never flushed\n";`,
		`kill TERM => $$;`,
		`print "after\n";`,
	)})
	var killed *perl.SignalError
	if !errors.As(err, &killed) {
		t.Fatalf("RunProgram error = %v, want a *perl.SignalError", err)
	}
	if killed.Signal != 15 {
		t.Fatalf("signal = %d, want 15 (SIGTERM)", killed.Signal)
	}
	if stdout != "before\n" {
		t.Fatalf("stdout = %q, want only what was flushed before the signal", stdout)
	}
}

// A handled signal does not end the program, and an ignored one does
// nothing.
func TestRunProgramHandlesItsSignals(t *testing.T) {
	status, err, stdout := signalRun(t, perl.Config{}, perl.Program{Code: code(
		`$SIG{TERM} = sub { print "caught $_[0]\n" };`,
		`$SIG{HUP} = 'IGNORE';`,
		`kill TERM => $$; kill HUP => $$; kill CHLD => $$;`,
		`print "alive\n"; exit 3;`,
	)})
	if err != nil || status != 3 || stdout != "caught TERM\nalive\n" {
		t.Fatalf("status=%d err=%v stdout=%q", status, err, stdout)
	}
}

// alarm works in the sandbox too: the timer is the instance's own.
func TestAlarmInTheSandbox(t *testing.T) {
	start := time.Now()
	status, err, stdout := signalRun(t, perl.Config{}, perl.Program{Code: code(
		`$SIG{ALRM} = sub { die "timed out\n" };`,
		`alarm 1;`,
		`my $ok = eval { sleep 600; 1 };`,
		`print $ok ? "slept" : "interrupted: $@";`,
	)})
	if err != nil || status != 0 || stdout != "interrupted: timed out\n" {
		t.Fatalf("status=%d err=%v stdout=%q", status, err, stdout)
	}
	if d := time.Since(start); d > time.Minute {
		t.Fatalf("the alarm took %v to interrupt the sleep", d)
	}
	// An alarm nobody handles ends the program by SIGALRM.
	_, err, _ = signalRun(t, perl.Config{}, perl.Program{Code: code(`alarm 1; sleep 600;`)})
	var killed *perl.SignalError
	if !errors.As(err, &killed) || killed.Signal != 14 {
		t.Fatalf("RunProgram error = %v, want death by signal 14 (SIGALRM)", err)
	}
	if !strings.Contains(err.Error(), "signal 14") {
		t.Fatalf("error text %q does not name the signal", err.Error())
	}
}

// A sandboxed program cannot become a server on the host's network: binding
// a socket to a local address is the host process's privilege.
func TestSandboxCannotListen(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Socket;`,
		`socket(my $s, PF_INET, SOCK_STREAM, 0) or die "socket: $!";`,
		`print "bind: ", (bind($s, pack_sockaddr_in(0, pack("C4", 127, 0, 0, 1))) ? "bound" : ($!{EACCES} ? "EACCES" : "error $!")), "\n";`,
		`print "listen: ", (listen($s, 5) ? "listening" : ($!{EACCES} ? "EACCES" : "error $!")), "\n";`,
	)}, "")
	r.want(t, 0, "bind: EACCES\nlisten: EACCES\n")
}
