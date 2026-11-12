//go:build darwin || linux

package gperl_test

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// A process whose parent has ended is adopted by process 1.
func TestOrphanIsAdoptedByProcessOne(t *testing.T) {
	got := runForkScript(t, `
use POSIX ();
pipe(my $r, my $w) or die "pipe: $!";
my $middle = fork // die "fork: $!";
if (!$middle) {
    close $r;
    pipe(my $seen_r, my $seen_w) or die "pipe: $!";
    my $child = fork // die "fork: $!";
    if (!$child) {
        close $seen_r;
        my $parent = getppid();
        close $seen_w; # the parent may go now
        select(undef, undef, undef, 0.05) while kill(0, $parent);
        syswrite($w, ($parent == getppid() ? "unchanged" : getppid()) . "\n");
        POSIX::_exit(0);
    }
    close $seen_w;
    my $eof = <$seen_r>;
    POSIX::_exit(0);
}
close $w;
waitpid($middle, 0);
my $said = <$r>;
chomp $said;
report("parent of the child with a parent=", (getppid() == 1 ? "1" : "not 1"));
report("parent of the orphan=$said");
`)
	wantLines(t, got,
		"parent of the child with a parent=not 1",
		"parent of the orphan=1",
	)
}

// Processes form groups: a forked process starts in its parent's, may start
// its own, and a signal to a group reaches its members and nobody else.
func TestProcessGroupsOfForkedProcesses(t *testing.T) {
	got := runForkScript(t, `
use POSIX ();
sub child(&) {
    my ($body) = @_;
    my $pid = fork // die "fork: $!";
    return $pid if $pid;
    $body->();
    POSIX::_exit(0);
}
sub await_signal { my $end = time + 10; select(undef, undef, undef, 0.05) while time < $end; }

my $group = getpgrp();
report("root: in a group=", ($group > 0 ? 1 : 0));
pipe(my $ready_r, my $ready_w) or die "pipe: $!";
my $outsider = child {
    close $ready_r;
    $SIG{USR1} = sub { report("outsider: signalled") };
    report("outsider: group is the parent's=", (getpgrp() == $group ? 1 : 0));
    close $ready_w;
    await_signal();
};
my $leader = child {
    close $ready_r;
    report("leader: setpgrp=", (setpgrp(0, 0) ? 1 : 0), " leads its group=", (getpgrp() == $$ ? 1 : 0));
    $SIG{USR1} = sub { report("leader: USR1"); };
    my $member = child {
        $SIG{USR1} = sub { report("member: USR1"); POSIX::_exit(0) };
        report("member: group is the leader's=", (getpgrp() == getppid() ? 1 : 0));
        close $ready_w;
        await_signal();
    };
    close $ready_w;
    waitpid($member, 0);
    report("leader: setsid as a group leader=", (POSIX::setsid() != -1 ? "allowed" : ($!{EPERM} ? "EPERM" : "error $!")));
};
close $ready_w;
my $eof = <$ready_r>;
report("root: group of the leader, asked from outside=", (getpgrp($leader) == $leader ? "its own" : getpgrp($leader)));
report("root: signal to the group=", kill("USR1", -$leader));
waitpid($leader, 0);
report("root: group without a process=", kill("USR1", -$leader), " ", ($!{ESRCH} ? "ESRCH" : "error $!"));
kill("KILL", $outsider);
waitpid($outsider, 0);

my $session = child {
    my $sid = POSIX::setsid();
    report("session: setsid=", ($sid == $$ ? "its pid" : "$sid $!"), " group=", (getpgrp() == $$ ? "its own" : "not its own"));
};
waitpid($session, 0);
`)
	sort.Strings(got)
	wantLines(t, got,
		"leader: USR1",
		"leader: setpgrp=1 leads its group=1",
		"leader: setsid as a group leader=EPERM",
		"member: USR1",
		"member: group is the leader's=1",
		"outsider: group is the parent's=1",
		"root: group of the leader, asked from outside=its own",
		"root: group without a process=0 ESRCH",
		"root: in a group=1",
		"root: signal to the group=1",
		"session: setsid=its pid group=its own",
	)
}

// A signal to the process's own group reaches the process and the
// processes it forked. The program is a process of its own here (a signal
// to the group of the test would reach the test).
func TestSignalToTheProcessGroupOfAGperlProcess(t *testing.T) {
	shim := perlCommandShim(t)
	report := filepath.Join(t.TempDir(), "report.txt")
	script := writeScript(t, "group.pl", `
use strict; use warnings;
my ($report, $perl) = @ARGV;
system($perl, '-e', <<'PROGRAM', $report);
use POSIX ();
my $report = shift;
sub report { open(my $fh, '>>', $report) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
setpgrp(0, 0) or die "setpgrp: $!";
report("leads a group of the host=", (getpgrp() == $$ ? 1 : 0));
my @children;
for my $n (1, 2) {
    my $pid = fork // die "fork: $!";
    if (!$pid) {
        $SIG{TERM} = sub { report("child $n: TERM"); POSIX::_exit(0) };
        my $end = time + 10; select(undef, undef, undef, 0.05) while time < $end;
        POSIX::_exit(1);
    }
    push @children, $pid;
}
select(undef, undef, undef, 0.3);
my $got = 0;
$SIG{TERM} = sub { $got++ };
report("kill(TERM, 0)=", kill("TERM", 0));
waitpid($_, 0) for @children;
select(undef, undef, undef, 0.2);
report("the process itself got it=", $got);
$SIG{TERM} = 'DEFAULT';
kill("TERM", -getpgrp());
select(undef, undef, undef, 5);
report("survived its own group's TERM");
PROGRAM
open(my $fh, '>>', $report) or die; syswrite($fh, "signal=" . ($? & 127) . "\n"); close $fh;
`)
	status, err := gperl.Run(script, []string{report, shim})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	got := readLines(t, report)
	sort.Strings(got)
	wantLines(t, got,
		"child 1: TERM",
		"child 2: TERM",
		"kill(TERM, 0)=1",
		"leads a group of the host=1",
		"signal=15",
		"the process itself got it=1",
	)
}
