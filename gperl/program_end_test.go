//go:build darwin || linux

package gperl_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/goccy/go-perl/gperl"
)

// When the program ends, the pseudo-processes it forked that are still
// running end with it: the process does not wait for them, whatever they
// are doing, and nothing of them is left behind.

// endPrelude gives a program report (one line into the file named by its
// first argument) and child: a fork child that says when it is set up and
// then runs the body, in which it is expected to stay.
const endPrelude = `
use strict; use warnings;
my $REPORT = shift @ARGV;
sub report { open(my $fh, '>>', $REPORT) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
sub child(&) {
    my ($body) = @_;
    pipe(my $r, my $w) or die "pipe: $!";
    my $pid = fork // die "fork: $!";
    if ($pid == 0) { close $r; $body->(sub { syswrite($w, "ready\n"); close $w }); report("child: its body returned"); exit 0 }
    close $w;
    my $ready = <$r>;
    return $pid;
}
# long enough for a child that said it is ready to be inside its call
sub settle { select(undef, undef, undef, 0.3) }
`

// endOutcome is how a gperl process ended.
type endOutcome struct {
	status int // exit status, or -1 for a death by signal
	signal syscall.Signal
	took   time.Duration
	report []string
}

// runToItsEnd runs src (after endPrelude) as a gperl process of its own and
// returns how it ended. A process that is still there after the limit is
// killed, with its whole process group, and fails the test.
func runToItsEnd(t *testing.T, src string, args ...string) endOutcome {
	t.Helper()
	shim := perlCommandShim(t)
	script := writeScript(t, "end.pl", endPrelude+src)
	report := filepath.Join(t.TempDir(), "report.txt")
	cmd := exec.Command(shim, append([]string{script, report}, args...)...)
	var output bytes.Buffer
	// A file, not a pipe: a process left behind would hold a pipe open and
	// keep Wait from returning.
	outFile, err := os.Create(filepath.Join(t.TempDir(), "output.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()
	cmd.Stdout, cmd.Stderr = outFile, outFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	start := time.Now()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()
	const limit = 30 * time.Second
	select {
	case err = <-waited:
	case <-time.After(limit):
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-waited
		raw, _ := os.ReadFile(outFile.Name())
		t.Fatalf("the process was still running %v after it was started\noutput:\n%s", limit, raw)
	}
	out := endOutcome{took: time.Since(start)}
	var exit *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exit):
		ws := exit.Sys().(syscall.WaitStatus)
		if ws.Signaled() {
			out.status, out.signal = -1, ws.Signal()
		} else {
			out.status = ws.ExitStatus()
		}
	default:
		t.Fatal(err)
	}
	if raw, err := os.ReadFile(outFile.Name()); err == nil && len(raw) > 0 {
		output.Write(raw)
		t.Logf("output of the process:\n%s", raw)
	}
	if raw, err := os.ReadFile(report); err == nil {
		out.report = strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	}
	// Nothing of the process's group is left: the pseudo-processes are in
	// the process itself, and the commands they started are in its group.
	deadline := time.Now().Add(5 * time.Second)
	for syscall.Kill(-cmd.Process.Pid, 0) == nil {
		if time.Now().After(deadline) {
			left, _ := exec.Command("ps", "-o", "pid=,command=", "-g", strconv.Itoa(cmd.Process.Pid)).Output()
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			t.Fatalf("processes of the program's group are still running after it ended:\n%s", left)
		}
		time.Sleep(20 * time.Millisecond)
	}
	return out
}

// promptly fails the test when the process took about as long as one that
// waited for a child would have.
func promptly(t *testing.T, out endOutcome) {
	t.Helper()
	if out.took > 15*time.Second {
		t.Errorf("the process took %v to end", out.took)
	}
}

func TestProgramEndsWhileChildrenBlock(t *testing.T) {
	cases := []struct{ name, body string }{
		{"in accept", `
use IO::Socket::INET;
my $server = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => 0) or die "listen: $!";
child { my ($ready) = @_; $ready->(); my $client = $server->accept; };
`},
		{"in sleep", `
child { my ($ready) = @_; $ready->(); sleep 600; };
`},
		{"in select", `
child { my ($ready) = @_; $ready->(); select(undef, undef, undef, 600); };
`},
		{"reading a pipe", `
pipe(my $r, my $w) or die "pipe: $!";
child { my ($ready) = @_; $ready->(); my $line = <$r>; };
`},
		{"reading a socket", `
use Socket;
socketpair(my $a, my $b, AF_UNIX, SOCK_STREAM, 0) or die "socketpair: $!";
child { my ($ready) = @_; $ready->(); sysread($a, my $buf, 1); };
`},
		{"in waitpid for its own child", `
child {
    my ($ready) = @_;
    my $grandchild = fork // die "fork: $!";
    if ($grandchild == 0) { sleep 600; report("grandchild: its sleep returned"); exit 0 }
    $ready->();
    waitpid($grandchild, 0);
};
`},
		{"in a busy loop", `
child { my ($ready) = @_; $ready->(); 1 while 1; };
`},
		{"opening a FIFO", `
use POSIX ();
my $fifo = "$REPORT.fifo";
POSIX::mkfifo($fifo, 0600) or die "mkfifo: $!";
child { my ($ready) = @_; $ready->(); open(my $in, '<', $fifo); };
`},
		{"several of them", `
use IO::Socket::INET;
my $server = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => 0) or die "listen: $!";
for my $worker (1 .. 5) {
    child { my ($ready) = @_; $ready->(); my $client = $server->accept; };
}
`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := runToItsEnd(t, c.body+`
settle();
report("program: ending");
exit 7;
`)
			if out.status != 7 {
				t.Errorf("exit status = %d (signal %v), want 7", out.status, out.signal)
			}
			promptly(t, out)
			wantLines(t, out.report, "program: ending")
		})
	}
}

// The children are killed: none of their END blocks or destructors runs,
// while the program's own run before its status is fixed.
func TestProgramEndRunsNoCodeOfItsChildren(t *testing.T) {
	out := runToItsEnd(t, `
package Guard { sub new { bless {}, shift } sub DESTROY { main::report("destructor in process ", ($$ == $main::program ? "program" : "child")) } }
our $program = $$;
our $guard = Guard->new;
END { report("END block in process ", ($$ == $program ? "program" : "child")); $? = 5 if $$ == $program }
child { my ($ready) = @_; $ready->(); sleep 600; };
child { my ($ready) = @_; $ready->(); 1 while 1; };
settle();
exit 0;
`)
	if out.status != 5 {
		t.Errorf("exit status = %d (signal %v), want the 5 the program's END block left", out.status, out.signal)
	}
	promptly(t, out)
	wantLines(t, out.report, "END block in process program", "destructor in process program")
}

// The commands the children started and have not waited for are killed.
func TestProgramEndKillsTheCommandsOfItsChildren(t *testing.T) {
	sleep, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("no sleep command on this host")
	}
	out := runToItsEnd(t, `
my $sleep = shift @ARGV;
# waited for with system
child { my ($ready) = @_; $ready->(); system($sleep, "600"); };
# read from
child {
    my ($ready) = @_;
    my $pid = open(my $out, '-|', $sleep, "601") // die "spawn: $!";
    report("command $pid");
    $ready->();
    my $line = <$out>;
};
# the child became the command
child {
    my ($ready) = @_;
    $ready->();
    exec('/bin/sh', '-c', 'echo "command $$" >> "$0"; exec "$1" 602', $REPORT, $sleep) or die "exec: $!";
};
# a grandchild's
child {
    my ($ready) = @_;
    my $grandchild = fork // die "fork: $!";
    if ($grandchild == 0) {
        my $pid = open(my $out, '-|', $sleep, "603") // die "spawn: $!";
        report("command $pid");
        $ready->();
        my $line = <$out>;
        exit 0;
    }
    waitpid($grandchild, 0);
};
# until the three commands that say so have said that they run
for (1 .. 200) {
    open(my $fh, '<', $REPORT) or next;
    last if 3 == grep { /^command / } <$fh>;
} continue { select(undef, undef, undef, 0.05) }
settle();
report("program: ending");
exit 3;
`, sleep)
	if out.status != 3 {
		t.Errorf("exit status = %d (signal %v), want 3", out.status, out.signal)
	}
	promptly(t, out)
	commands := 0
	for _, line := range out.report {
		number, isCommand := strings.CutPrefix(line, "command ")
		if !isCommand {
			continue
		}
		commands++
		pid, err := strconv.Atoi(number)
		if err != nil {
			t.Fatalf("report line %q", line)
		}
		if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Errorf("command %d is still running after the program ended", pid)
		}
	}
	if commands != 3 {
		t.Errorf("%d commands reported their pid, want 3; report: %q", commands, out.report)
	}
	if last := out.report[len(out.report)-1]; last != "program: ending" {
		t.Errorf("something ran after the program ended: %q", out.report)
	}
}

// A program that a signal ends, or that leaves through _exit, takes its
// children with it as well.
func TestProgramKilledOrLeavingAtOnceEndsItsChildren(t *testing.T) {
	out := runToItsEnd(t, `
child { my ($ready) = @_; $ready->(); sleep 600; };
settle();
kill TERM => $$;
sleep 600;
`)
	if out.status != -1 || out.signal != syscall.SIGTERM {
		t.Errorf("the program ended with status %d, signal %v; want death by SIGTERM", out.status, out.signal)
	}
	promptly(t, out)

	out = runToItsEnd(t, `
use POSIX ();
child { my ($ready) = @_; $ready->(); sleep 600; };
settle();
POSIX::_exit(9);
`)
	if out.status != 9 {
		t.Errorf("exit status = %d (signal %v), want 9", out.status, out.signal)
	}
	promptly(t, out)
}

// A program run inside a process that goes on (the library, the build
// pipeline) leaves no goroutine of a pseudo-process behind.
func TestProgramEndLeavesNoGoroutineBehind(t *testing.T) {
	keepProcessState(t)
	// A first program leaves behind the goroutines the process keeps once
	// it has run one (the delivery of the host's signals).
	if status, err := runInThisProcess(t, writeScript(t, "first.pl", "exit 0;\n"), nil); err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	before := runtime.NumGoroutine()
	script := writeScript(t, "end.pl", endPrelude+`
use IO::Socket::INET;
my $server = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => 0) or die "listen: $!";
for my $worker (1 .. 4) {
    child { my ($ready) = @_; $ready->(); my $client = $server->accept; };
}
child { my ($ready) = @_; $ready->(); sleep 600; };
child { my ($ready) = @_; $ready->(); 1 while 1; };
settle();
exit 4;
`)
	report := filepath.Join(t.TempDir(), "report.txt")
	start := time.Now()
	status, err := runInThisProcess(t, script, []string{report})
	if err != nil || status != 4 {
		t.Fatalf("Run = (%d, %v), want (4, nil)", status, err)
	}
	if took := time.Since(start); took > 15*time.Second {
		t.Errorf("Run took %v", took)
	}
	deadline := time.Now().Add(10 * time.Second)
	for runtime.NumGoroutine() > before {
		if time.Now().After(deadline) {
			buf := make([]byte, 1<<20)
			buf = buf[:runtime.Stack(buf, true)]
			t.Fatalf("%d goroutines before the program, %d after it:\n%s", before, runtime.NumGoroutine(), buf)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// A binary built from a program ends like the program under the command.
func TestBuiltBinaryEndsWhileChildrenBlock(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a full binary; skipped in -short")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "server.pl")
	if err := os.WriteFile(script, []byte(`
use strict; use warnings;
use IO::Socket::INET;
my $server = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => 0) or die "listen: $!";
pipe(my $r, my $w) or die "pipe: $!";
for my $worker (1 .. 3) {
    my $pid = fork // die "fork: $!";
    next if $pid;
    my $grandchild = fork // die "fork: $!";
    if ($grandchild == 0) { syswrite($w, "ready\n"); sleep 600; exit 0 }
    syswrite($w, "ready\n");
    my $client = $server->accept;
    exit 0;
}
close $w;
for (1 .. 6) { my $ready = <$r> }
select(undef, undef, undef, 0.3);
print "ending\n";
exit 6;
`), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "server-bin")
	t.Setenv("GPERL_DEV_REPLACE", devReplaces(t))
	if err := gperl.Build(script, bin); err != nil {
		t.Fatalf("Build: %v", err)
	}
	cmd := exec.Command(bin)
	outFile, err := os.Create(filepath.Join(dir, "output.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()
	cmd.Stdout, cmd.Stderr = outFile, outFile
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()
	select {
	case err = <-waited:
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		<-waited
		t.Fatal("the built binary was still running 30s after it was started")
	}
	raw, _ := os.ReadFile(outFile.Name())
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 6 || string(raw) != "ending\n" {
		t.Fatalf("the built binary ended with %v and output %q, want exit status 6 and %q", err, raw, "ending\n")
	}
}
