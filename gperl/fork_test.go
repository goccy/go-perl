package gperl_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// runForkScript runs a Perl program that reports what it observed, one
// line per fact, into the file named by its first argument, and returns
// those lines. The programs fork, so their facts come from several
// pseudo-processes; each appends whole lines with syswrite.
func runForkScript(t *testing.T, src string) []string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fork tests drive /bin/sh and /bin/echo")
	}
	const prelude = `
use strict; use warnings;
my $REPORT = shift @ARGV;
sub report { open(my $fh, '>>', $REPORT) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
`
	script := writeScript(t, "fork.pl", prelude+src)
	report := filepath.Join(t.TempDir(), "report.txt")
	status, err := gperl.Run(script, []string{report})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	raw, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("no report written: %v", err)
	}
	return strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
}

func wantLines(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("report:\n  %s\nwant:\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
}

// The child gets a copy of the program's state: what it changes stays its
// own, it has its own pid, and its exit status reaches the parent.
func TestForkChildIsASeparateProcess(t *testing.T) {
	got := runForkScript(t, `
my $state = 'before';
my $parent = $$;
my $pid = fork;
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    $state = 'changed by child';
    report("child: pid differs=", ($$ != $parent ? 1 : 0), " ppid matches=", (getppid() == $parent ? 1 : 0));
    exit 7;
}
my $reaped = waitpid($pid, 0);
report("parent: reaped=", ($reaped == $pid ? 1 : 0), " exit=", $? >> 8, " state=$state pid kept=", ($$ == $parent ? 1 : 0));
`)
	wantLines(t, got,
		"child: pid differs=1 ppid matches=1",
		"parent: reaped=1 exit=7 state=before pid kept=1",
	)
}

// A descriptor the child closes stays open in the parent, and a child that
// never execs runs concurrently with its parent: here it writes more than a
// pipe holds while the parent reads.
func TestForkChildWritesPipeWithoutExec(t *testing.T) {
	got := runForkScript(t, `
pipe(my $r, my $w) or die "pipe: $!";
my $pid = fork;
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    close $r;
    print {$w} ("x" x 1023, "\n") for 1 .. 300;
    close $w;
    exit 0;
}
close $w;
my ($lines, $bytes) = (0, 0);
while (my $line = <$r>) { $lines++; $bytes += length $line }
close $r;
waitpid($pid, 0);
report("lines=$lines bytes=$bytes exit=", $? >> 8);
`)
	wantLines(t, got, "lines=300 bytes=307200 exit=0")
}

// The fork-then-exec idiom with redirected stdio, the way IPC::Open3 does
// it: the child points its stdout at the pipe by descriptor number and
// closes the ends it does not use.
func TestForkExecWithRedirectedStdio(t *testing.T) {
	got := runForkScript(t, `
pipe(my $r, my $w) or die "pipe: $!";
my $pid = fork;
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    close $r;
    open(STDOUT, ">&=" . fileno($w)) or die "redirect: $!";
    exec('/bin/sh', '-c', 'echo from-exec; exit 3') or die "exec: $!";
}
close $w;
my @out = <$r>;
close $r;
waitpid($pid, 0);
chomp @out;
report("out=@out exit=", $? >> 8);
`)
	wantLines(t, got, "out=from-exec exit=3")
}

func TestIPCOpen3(t *testing.T) {
	got := runForkScript(t, `
use IPC::Open3;
use Symbol 'gensym';
my $err = gensym;
my $pid = open3(my $in, my $out, $err, '/bin/sh', '-c', 'read line; echo "got $line"; echo oops >&2; exit 2');
print {$in} "hello\n";
close $in;
my @out = <$out>;
my @err = <$err>;
waitpid($pid, 0);
chomp(@out, @err);
report("out=@out err=@err exit=", $? >> 8);
`)
	wantLines(t, got, "out=got hello err=oops exit=2")
}

// open(FH, "-|") forks with the child's stdout on the pipe.
func TestForkingPipeOpen(t *testing.T) {
	got := runForkScript(t, `
my $pid = open(my $kid, '-|');
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    print "line $_\n" for 1 .. 3;
    exit 0;
}
my @lines = <$kid>;
close $kid;
chomp @lines;
report("read=", join('|', @lines), " exit=", $? >> 8);

$pid = open(my $sink, '|-');
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    my @in = <STDIN>;
    chomp @in;
    report("sink child read=", join('|', @in));
    exit 0;
}
print {$sink} "a\nb\n";
close $sink;
report("sink exit=", $? >> 8);
`)
	wantLines(t, got,
		"read=line 1|line 2|line 3 exit=0",
		"sink child read=a|b",
		"sink exit=0",
	)
}

// wait reaps whichever child finishes, and children can fork in turn.
func TestWaitForAnyChildAndGrandchild(t *testing.T) {
	got := runForkScript(t, `
my %want;
for my $code (1 .. 3) {
    my $pid = fork;
    die "fork: $!" unless defined $pid;
    if ($pid == 0) {
        if ($code == 3) {
            my $grand = fork;
            die "fork: $!" unless defined $grand;
            exit 40 if $grand == 0;
            waitpid($grand, 0);
            exit(($? >> 8) + 2);
        }
        exit $code;
    }
    $want{$pid} = $code == 3 ? 42 : $code;
}
my @seen;
while ((my $pid = wait) > 0) {
    push @seen, ($want{$pid} == ($? >> 8) ? "ok" : "bad") . ":" . ($? >> 8);
}
report(join(' ', sort @seen));
report("no more children: ", (wait == -1 ? 1 : 0));
`)
	wantLines(t, got, "ok:1 ok:2 ok:42", "no more children: 1")
}

// A terminating signal ends a child that would never finish by itself, and
// the parent sees a signal death.
func TestKillForkChild(t *testing.T) {
	got := runForkScript(t, `
pipe(my $r, my $w) or die "pipe: $!";
my $pid = fork;
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    close $r;
    syswrite($w, "ready\n");
    my $n = 0;
    $n++ while 1;
}
close $w;
my $ready = <$r>;
my $sent = kill 'TERM', $pid;
waitpid($pid, 0);
report("sent=$sent signal=", $? & 127, " exit=", $? >> 8);
`)
	wantLines(t, got, "sent=1 signal=15 exit=0")
}

// A die nothing catches ends the child like any perl process: status 255.
func TestForkChildUncaughtDie(t *testing.T) {
	got := runForkScript(t, `
my $pid = fork;
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    open(STDERR, '>', '/dev/null') or exit 1;
    $! = 0; # an uncaught die exits with errno when one is pending
    die "child gives up\n";
}
waitpid($pid, 0);
report("exit=", $? >> 8);

# a die inside an eval the child inherited is caught there
my $caught = eval {
    my $kid = fork;
    die "fork: $!" unless defined $kid;
    die "inside eval\n" if $kid == 0;
    waitpid($kid, 0);
    report("eval parent sees exit=", $? >> 8);
    "parent";
};
if (!defined $caught) {
    chomp(my $why = $@);
    report("child caught: $why");
    exit 9;
}
`)
	wantLines(t, got,
		"exit=255",
		"child caught: inside eval",
		"eval parent sees exit=9",
	)
}

// A child's exit status is the one it ends with after its END blocks ran:
// an END block may assign $?, as test libraries do.
func TestForkChildEndBlockSetsStatus(t *testing.T) {
	got := runForkScript(t, `
my $parent = $$;
END { $? = 5 if $$ != $parent }
my $pid = fork // die "fork: $!";
exit 0 if !$pid;
waitpid($pid, 0);
report("child status ", $? >> 8);
`)
	wantLines(t, got, "child status 5")
}

// Global destruction belongs to the child's run too: a destructor there
// decides the status its parent sees.
func TestForkChildStatusFromGlobalDestruction(t *testing.T) {
	got := runForkScript(t, `
use POSIX ();
my $parent = $$;
our $how = '';
sub Guard::DESTROY {
    return if $$ == $parent;
    $? = 5 if $how eq 'assign';
    exit 7 if $how eq 'exit';
    POSIX::_exit(219) if $how eq '_exit';
    kill TERM => $$ if $how eq 'signal';
}
our $guard = bless [], 'Guard';
for my $way (qw(assign exit _exit signal)) {
    my $pid = fork // die "fork: $!";
    if (!$pid) { $how = $way; exit 0 }
    waitpid($pid, 0);
    report("$way: signal=", $? & 127, " exit=", $? >> 8);
}
`)
	wantLines(t, got,
		"assign: signal=0 exit=5",
		"exit: signal=0 exit=7",
		"_exit: signal=0 exit=219",
		"signal: signal=15 exit=0",
	)
}

// system and exec look a bare command name up on the PATH the program has
// at that moment, as execvp does — not on the PATH gperl was started with.
func TestSystemSearchesTheProgramsPath(t *testing.T) {
	dir := t.TempDir()
	tool := filepath.Join(dir, "gperl-test-tool")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\nexit 7\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := runForkScript(t, `
my $missing = system "gperl-test-tool", "arg";
report("before: ", $missing);
{
    local $ENV{PATH} = "`+dir+`:$ENV{PATH}";
    report("found: ", system("gperl-test-tool", "arg") >> 8);
}
report("after: ", system("gperl-test-tool", "arg"));
`)
	wantLines(t, got, "before: -1", "found: 7", "after: -1")
}

// A new descriptor takes the lowest free number, so a program can re-point
// a standard stream by closing it and opening something else.
func TestDescriptorNumbersAreReused(t *testing.T) {
	got := runForkScript(t, `
use POSIX ();
my $first = POSIX::dup(0); POSIX::close($first);
for (1 .. 20) { open(my $fh, '<', $0) or die "open: $!"; pipe(my $r, my $w) or die "pipe: $!"; }
my $again = POSIX::dup(0); POSIX::close($again);
report($first == $again ? "same number" : "leaked: $first then $again");
close(STDOUT) or die "close: $!";
open(my $fh, '>', '/dev/null') or die "open: $!";
report("stdout's number is ", fileno($fh));
`)
	wantLines(t, got, "same number", "stdout's number is 1")
}
