//go:build darwin || linux

package gperl_test

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// readLines returns the lines of a report file.
func readLines(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no report written: %v", err)
	}
	return strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
}

// perlCommandShim writes an executable that runs this test binary as the
// perl command (the re-entry the XS build pipeline uses), and returns its
// path: a program under test starts real gperl processes with it.
func perlCommandShim(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(t.TempDir(), "perl")
	script := "#!/bin/sh\nGPERL_INTERNAL_PERL_CLI=1\nexport GPERL_INTERNAL_PERL_CLI\nexec \"" + exe + "\" run \"$@\"\n"
	if err := os.WriteFile(shim, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return shim
}

// The program is the process: its pid, parent, user and group are the
// process's own, and the user database is the host's.
func TestProcessIdentityIsTheHostProcess(t *testing.T) {
	got := runForkScript(t, `
report("pid=$$ ppid=", getppid(), " uid=$< euid=$> gid=", $( + 0, " egid=", $) + 0);
my @pw = getpwuid($<);
report("name=$pw[0] uid=$pw[2] gid=$pw[3] dir=$pw[7]");
my @byname = getpwnam($pw[0]);
report("byname uid=$byname[2]");
report("missing user=", defined(getpwnam("no-such-user-in-any-database")) ? "found" : "undef");
my @gr = getgrgid($( + 0);
report("group=$gr[0] gid=$gr[2] byname=", scalar(getgrnam($gr[0])));
report("login=", scalar(getlogin()));
report("owns its report file=", (-o $REPORT ? 1 : 0), " owns /=", (-o "/" ? 1 : 0));
`)
	me, err := user.Current()
	if err != nil {
		t.Skipf("the host has no entry for the current user: %v", err)
	}
	group, err := user.LookupGroupId(strconv.Itoa(os.Getgid()))
	if err != nil {
		t.Skipf("the host has no entry for the current group: %v", err)
	}
	ownsRoot := 0
	if os.Getuid() == 0 {
		ownsRoot = 1
	}
	wantLines(t, got,
		fmt.Sprintf("pid=%d ppid=%d uid=%d euid=%d gid=%d egid=%d",
			os.Getpid(), os.Getppid(), os.Getuid(), os.Geteuid(), os.Getgid(), os.Getegid()),
		fmt.Sprintf("name=%s uid=%s gid=%s dir=%s", me.Username, me.Uid, me.Gid, me.HomeDir),
		"byname uid="+me.Uid,
		"missing user=undef",
		fmt.Sprintf("group=%s gid=%s byname=%s", group.Name, group.Gid, group.Gid),
		"login="+me.Username,
		fmt.Sprintf("owns its report file=1 owns /=%d", ownsRoot),
	)
}

// A fork child has a pid of its own and the same user.
func TestForkChildKeepsTheUser(t *testing.T) {
	got := runForkScript(t, `
my $parent = $$;
my $pid = fork // die "fork: $!";
if ($pid == 0) {
    report("child: own pid=", ($$ != $parent ? 1 : 0), " ppid=parent=", (getppid() == $parent ? 1 : 0),
           " uid=$< user=", scalar(getpwuid($<)));
    exit 0;
}
waitpid($pid, 0);
report("parent: uid=$< user=", scalar(getpwuid($<)));
`)
	me, err := user.Current()
	if err != nil {
		t.Skipf("the host has no entry for the current user: %v", err)
	}
	wantLines(t, got,
		fmt.Sprintf("child: own pid=1 ppid=parent=1 uid=%s user=%s", me.Uid, me.Username),
		fmt.Sprintf("parent: uid=%s user=%s", me.Uid, me.Username),
	)
}

// A process may set its ids to what they are, and nothing else.
func TestProcessCannotChangeItsUser(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("the superuser may change its ids")
	}
	got := runForkScript(t, `
use POSIX ();
report("same uid: ", (POSIX::setuid($<) ? "ok" : "refused"));
report("other uid: ", (POSIX::setuid($< + 1) ? "ok" : ($!{EPERM} ? "EPERM" : "error $!")));
report("uid still mine=", ($< == (stat $REPORT)[4] ? 1 : 0));
`)
	wantLines(t, got, "same uid: ok", "other uid: EPERM", "uid still mine=1")
}

// A signal the program sends itself runs its handler at once, and the
// handler set with local is gone when the block is.
func TestSignalHandlerRunsForSelfKill(t *testing.T) {
	got := runForkScript(t, `
my @seen;
$SIG{USR1} = sub { push @seen, "USR1 handler got $_[0]" };
kill USR1 => $$;
report($_) for @seen;
{
    local $SIG{USR2} = sub { report("local USR2 handler") };
    kill USR2 => $$;
}
report("USR2 after the block: ", (defined $SIG{USR2} && length $SIG{USR2} ? $SIG{USR2} : "default"));
$SIG{HUP} = 'IGNORE';
kill HUP => $$;
report("survived an ignored HUP");
report("kill 0 on itself: ", kill(0, $$));
`)
	wantLines(t, got,
		"USR1 handler got USR1",
		"local USR2 handler",
		"USR2 after the block: default",
		"survived an ignored HUP",
		"kill 0 on itself: 1",
	)
}

// alarm delivers SIGALRM, whose handler ends whatever the program is
// waiting in: a sleep, a wait for a child, a read from a pipe.
func TestAlarmInterruptsBlockingCalls(t *testing.T) {
	got := runForkScript(t, `
sub timed {
    my ($what, $code) = @_;
    my $start = time;
    local $SIG{ALRM} = sub { die "alarm\n" };
    alarm 1;
    my $ok = eval { $code->(); 1 };
    my $left = alarm 0;
    report("$what: ", ($ok ? "finished" : "died: $@" =~ s/\n//r), " quickly=", (time - $start < 10 ? 1 : 0));
}
timed(sleep => sub { sleep 60 });
my $pid = fork // die "fork: $!";
if ($pid == 0) { sleep 600; exit 0 }
timed(waitpid => sub { waitpid($pid, 0) });
kill KILL => $pid;
waitpid($pid, 0);
report("the child died of signal ", $? & 127);
pipe(my $r, my $w) or die "pipe: $!";
timed(read => sub { my $line = <$r> });
timed(sysread => sub { sysread($r, my $buf, 10) // die "sysread: $!\n" });
timed(select => sub { my $rin = ''; vec($rin, fileno($r), 1) = 1; select($rin, undef, undef, 60) });
report("alarm reports the seconds left: ", do { alarm 100; alarm 0 });
`)
	wantLines(t, got,
		"sleep: died: alarm quickly=1",
		"waitpid: died: alarm quickly=1",
		"the child died of signal 9",
		"read: died: alarm quickly=1",
		"sysread: died: alarm quickly=1",
		"select: died: alarm quickly=1",
		"alarm reports the seconds left: 100",
	)
}

// A signal that a fork child neither handles nor ignores ends the child,
// and the parent sees a death by that signal; a child that handles the
// signal runs its handler instead.
func TestSignalsToAForkChild(t *testing.T) {
	got := runForkScript(t, `
sub child(&) {
    my ($body) = @_;
    pipe(my $r, my $w) or die "pipe: $!";
    my $pid = fork // die "fork: $!";
    if ($pid == 0) { close $r; $body->($w); exit 0 }
    close $w;
    my $ready = <$r>;   # the child says when it is set up
    return $pid;
}
sub outcome { my ($pid) = @_; waitpid($pid, 0); return "signal=" . ($? & 127) . " exit=" . ($? >> 8) }

my $plain = child { my ($w) = @_; syswrite($w, "ready\n"); sleep 600 };
kill TERM => $plain;
report("no handler: ", outcome($plain));

my $handler = child {
    my ($w) = @_;
    $SIG{TERM} = sub { report("child: TERM handler"); exit 3 };
    syswrite($w, "ready\n");
    sleep 600;
};
kill TERM => $handler;
report("handler: ", outcome($handler));

my $ignoring = child { my ($w) = @_; $SIG{TERM} = 'IGNORE'; syswrite($w, "ready\n"); sleep 600 };
kill TERM => $ignoring;
report("ignored: still running=", kill(0, $ignoring));
kill KILL => $ignoring;
report("then killed: ", outcome($ignoring));

my $self = child { my ($w) = @_; syswrite($w, "ready\n"); kill USR1 => $$; sleep 600 };
report("killed itself: ", outcome($self));

my $busy = child { my ($w) = @_; syswrite($w, "ready\n"); 1 while 1 };
kill KILL => $busy;
report("busy loop killed: ", outcome($busy));
`)
	wantLines(t, got,
		"no handler: signal=15 exit=0",
		"child: TERM handler",
		"handler: signal=0 exit=3",
		"ignored: still running=1",
		"then killed: signal=9 exit=0",
		"killed itself: signal=" + strconv.Itoa(int(syscall.SIGUSR1)) + " exit=0",
		"busy loop killed: signal=9 exit=0",
	)
}

// A signal sent to a child the instant it was forked — before the child has
// run a single statement — is not lost.
func TestSignalRightAfterFork(t *testing.T) {
	got := runForkScript(t, `
for my $sig (qw(USR1 TERM KILL)) {
    my $pid = fork // die "fork: $!";
    if ($pid == 0) { sleep 600; exit 0 }
    kill $sig => $pid;
    waitpid($pid, 0);
    report("$sig at once: signal=", $? & 127, " exit=", $? >> 8);
}
`)
	wantLines(t, got,
		"USR1 at once: signal=" + strconv.Itoa(int(syscall.SIGUSR1)) + " exit=0",
		"TERM at once: signal=15 exit=0",
		"KILL at once: signal=9 exit=0",
	)
}

// The parent is told when a child ends: its SIGCHLD handler runs and
// reaps the child.
func TestSIGCHLDReaper(t *testing.T) {
	got := runForkScript(t, `
use POSIX ":sys_wait_h";
my %status;
$SIG{CHLD} = sub { while ((my $kid = waitpid(-1, WNOHANG)) > 0) { $status{$kid} = $? >> 8 } };
my @kids = map { my $code = $_; my $pid = fork // die "fork: $!"; exit $code unless $pid; $pid } 3, 4;
my $deadline = time + 60;
sleep 1 while keys %status < 2 && time < $deadline;
report("reaped by the handler: ", join(" ", map { $status{$_} // "missing" } @kids));
report("nothing left to wait for: ", waitpid(-1, WNOHANG));
`)
	wantLines(t, got,
		"reaped by the handler: 3 4",
		"nothing left to wait for: -1",
	)
}

// A blocked signal is held until it is unblocked, and shows as pending
// meanwhile.
func TestSigprocmask(t *testing.T) {
	got := runForkScript(t, `
use POSIX qw(:signal_h);
my @order;
$SIG{USR1} = sub { push @order, "handler" };
my $set = POSIX::SigSet->new(SIGUSR1);
my $old = POSIX::SigSet->new;
sigprocmask(SIG_BLOCK, $set, $old) or die "sigprocmask: $!";
report("was blocked before: ", $old->ismember(SIGUSR1));
kill USR1 => $$;
push @order, "after kill";
my $pending = POSIX::SigSet->new;
sigpending($pending) or die "sigpending: $!";
report("pending while blocked: ", $pending->ismember(SIGUSR1), " other signal pending: ", $pending->ismember(SIGUSR2));
sigprocmask(SIG_UNBLOCK, $set) or die "sigprocmask: $!";
push @order, "after unblock";
report("order: @order");
sigpending($pending);
report("pending afterwards: ", $pending->ismember(SIGUSR1));

# a blocked signal from another process waits as well
my $parent = $$;
sigprocmask(SIG_BLOCK, $set);
my $pid = fork // die "fork: $!";
if ($pid == 0) { kill USR1 => $parent; exit 0 }
waitpid($pid, 0);
@order = ("child done");
sigprocmask(SIG_UNBLOCK, $set);
push @order, "unblocked";
report("from the child: @order");

# sigsuspend waits for a signal with a temporary mask
my $again = fork // die "fork: $!";
if ($again == 0) { sleep 1; kill USR1 => $parent; exit 0 }
@order = ();
sigprocmask(SIG_BLOCK, $set);
sigsuspend(POSIX::SigSet->new);
report("sigsuspend returned after: @order");
waitpid($again, 0);
`)
	wantLines(t, got,
		"was blocked before: 0",
		"pending while blocked: 1 other signal pending: 0",
		"order: after kill handler after unblock",
		"pending afterwards: 0",
		"from the child: child done handler unblocked",
		"sigsuspend returned after: handler",
	)
}

// A gperl process that a signal ends dies of that signal, as seen by the
// process that waits for it, and a gperl process is signalled like any
// other.
func TestSignalDeathOfAGperlProcess(t *testing.T) {
	shim := perlCommandShim(t)
	script := writeScript(t, "sig.pl", `
use strict; use warnings;
my ($report, $perl) = @ARGV;
sub report { open(my $fh, '>>', $report) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
sub status { "signal=" . ($? & 127) . " exit=" . ($? >> 8) }

system($perl, '-e', 'kill TERM => $$; print "not reached\n"');
report("killed itself with TERM: ", status());
system($perl, '-e', 'kill 9, $$');
report("killed itself with KILL: ", status());
system($perl, '-e', '$SIG{TERM} = sub { exit 4 }; kill TERM => $$; sleep 60');
report("handled TERM: ", status());

# signalled from outside, while it sleeps
my $pid = open(my $out, '-|', $perl, '-e', '$| = 1; print "ready\n"; sleep 600') // die "spawn: $!";
my $ready = <$out>;
kill USR1 => $pid;
close $out;
report("USR1 from outside: ", status());

# a handler in the child runs for a signal from outside
$pid = open($out, '-|', $perl, '-e', '$| = 1; $SIG{USR1} = sub { print "caught\n"; exit 5 }; print "ready\n"; sleep 600') // die "spawn: $!";
$ready = <$out>;
kill USR1 => $pid;
my $said = <$out> // "nothing\n";
close $out;
chomp $said;
report("USR1 handled in the child: said=$said ", status());

# the child signals its parent by pid
my $got = 0;
local $SIG{USR2} = sub { $got++ };
system($perl, '-e', 'kill USR2 => getppid()');
report("USR2 from the child: handler ran=$got ", status());
`)
	report := filepath.Join(t.TempDir(), "report.txt")
	status, err := gperl.Run(script, []string{report, shim})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	wantLines(t, readLines(t, report),
		"killed itself with TERM: signal=15 exit=0",
		"killed itself with KILL: signal=9 exit=0",
		"handled TERM: signal=0 exit=4",
		"USR1 from outside: signal=" + strconv.Itoa(int(syscall.SIGUSR1)) + " exit=0",
		"USR1 handled in the child: said=caught signal=0 exit=5",
		"USR2 from the child: handler ran=1 signal=0 exit=0",
	)
}

// A gperl process ends with the status its global destruction leaves: a
// destructor that assigns $?, exits, or is killed there decides it.
func TestStatusFromGlobalDestructionOfAGperlProcess(t *testing.T) {
	shim := perlCommandShim(t)
	script := writeScript(t, "destruct.pl", `
use strict; use warnings;
my ($report, $perl) = @ARGV;
sub report { open(my $fh, '>>', $report) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
sub status { "signal=" . ($? & 127) . " exit=" . ($? >> 8) }
my $guard = 'our $guard = bless [], "Guard";';

system($perl, '-e', "sub Guard::DESTROY { \$? = 5 } $guard");
report("assigns the status: ", status());
system($perl, '-e', "sub Guard::DESTROY { \$? = 5 } $guard exit 3");
report("assigns the status after exit: ", status());
system($perl, '-e', "sub Guard::DESTROY { exit 7 } $guard");
report("exits: ", status());
system($perl, '-MPOSIX', '-e', "sub Guard::DESTROY { POSIX::_exit(219) } $guard");
report("leaves through _exit: ", status());
system($perl, '-e', "sub Guard::DESTROY { kill TERM => \$\$ } $guard");
report("is killed: ", status());
`)
	report := filepath.Join(t.TempDir(), "report.txt")
	status, err := gperl.Run(script, []string{report, shim})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	wantLines(t, readLines(t, report),
		"assigns the status: signal=0 exit=5",
		"assigns the status after exit: signal=0 exit=5",
		"exits: signal=0 exit=7",
		"leaves through _exit: signal=0 exit=219",
		"is killed: signal=15 exit=0",
	)
}

// A signal the program ignores is ignored by the commands it starts, which
// inherit that, and no longer once the program has given the signal its
// default action back — nor by the commands of the next program of the
// process, when the first ended with the signal ignored.
func TestIgnoredSignalsGetTheirDefaultActionBack(t *testing.T) {
	const survives = `sub survives { my $sig = shift; system("/bin/sh", "-c", "kill -$sig \$\$; exit 0"); $? == 0 ? "survives" : ($? & 127) ? "dies" : "fails with $?" }`
	got := runForkScript(t, survives+`
for my $sig (qw(HUP INT TERM USR1)) {
    my @seen = (survives($sig));
    {
        local $SIG{$sig} = 'IGNORE';
        push @seen, survives($sig);
    }
    push @seen, survives($sig);
    $SIG{$sig} = 'IGNORE';
    $SIG{$sig} = 'DEFAULT';
    push @seen, survives($sig);
    report("a command that sends itself $sig: at first it $seen[0], while the program ignores the signal it $seen[1], after the scope of that it $seen[2], after DEFAULT it $seen[3]");
    $SIG{$sig} = 'IGNORE';   # and the program ends that way
}
`)
	var want []string
	for _, sig := range []string{"HUP", "INT", "TERM", "USR1"} {
		want = append(want, "a command that sends itself "+sig+": at first it dies, while the program ignores the signal it survives, after the scope of that it dies, after DEFAULT it dies")
	}
	wantLines(t, got, want...)

	got = runForkScript(t, survives+`
report("in the next program: ", join(", ", map { "$_ " . survives($_) } qw(HUP INT TERM USR1)));
`)
	wantLines(t, got, "in the next program: HUP dies, INT dies, TERM dies, USR1 dies")
}
