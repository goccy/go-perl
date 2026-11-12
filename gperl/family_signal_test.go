//go:build darwin || linux

package gperl_test

import (
	"sort"
	"testing"
)

// A process signals any process it knows the pid of, not only its own
// children: a sibling, the fork child that is its parent, a grandchild.
// Waiting stays the parent's.
func TestSignalsAcrossTheProcessFamily(t *testing.T) {
	got := runForkScript(t, `
use POSIX ();
sub child(&) {
    my ($body) = @_;
    my $pid = fork // die "fork: $!";
    return $pid if $pid;
    $body->();
    POSIX::_exit(0);
}
sub await_signal { my ($seconds) = @_; my $end = time + $seconds; select(undef, undef, undef, 0.05) while time < $end; }

# a sibling
my $first = child { $SIG{USR1} = sub { report("first child: USR1 from its sibling"); POSIX::_exit(0) }; await_signal(10); report("first child: no signal") };
select(undef, undef, undef, 0.2);
my $second = child { report("second child: kill sibling=", kill("USR1", $first)) };
waitpid($_, 0) for $first, $second;

# the parent of a fork child's child, and a grandchild
pipe(my $up_r, my $up_w) or die "pipe: $!";
my $middle = child {
    close $up_r;
    $SIG{TERM} = sub { report("middle: TERM from its child") };
    my $grandchild = child {
        $SIG{USR2} = sub { report("grandchild: USR2 from its grandparent"); POSIX::_exit(0) };
        select(undef, undef, undef, 0.2);
        report("grandchild: kill parent=", kill("TERM", getppid()));
        await_signal(10);
        report("grandchild: no signal");
    };
    syswrite($up_w, "$grandchild\n");
    close $up_w;
    my $reaped = waitpid($grandchild, 0);
    report("middle: reaped its child=", ($reaped == $grandchild ? 1 : 0));
};
close $up_w;
my $grandchild = <$up_r>;
chomp $grandchild;
select(undef, undef, undef, 0.6);
report("root: probe grandchild=", kill(0, $grandchild));
report("root: wait for a grandchild=", waitpid($grandchild, POSIX::WNOHANG()), " ", ($!{ECHILD} ? "ECHILD" : "error $!"));
report("root: kill grandchild=", kill("USR2", $grandchild));
waitpid($middle, 0);
report("root: probe after its end=", kill(0, $grandchild), " ", ($!{ESRCH} ? "ESRCH" : "error $!"));
`)
	sort.Strings(got)
	wantLines(t, got,
		"first child: USR1 from its sibling",
		"grandchild: USR2 from its grandparent",
		"grandchild: kill parent=1",
		"middle: TERM from its child",
		"middle: reaped its child=1",
		"root: kill grandchild=1",
		"root: probe after its end=0 ESRCH",
		"root: probe grandchild=1",
		"root: wait for a grandchild=-1 ECHILD",
		"second child: kill sibling=1",
	)
}

// A signal that a fork child sends to the program is pending for the program
// when the child's kill returns: whatever the child does next — here, a
// byte through a pipe — the program sees after its handler has run.
func TestSignalFromAForkChildIsPendingWhenItsKillReturns(t *testing.T) {
	got := runForkScript(t, `
my $parent = $$;
my $rounds = 400;
my $handled = 0;
$SIG{USR1} = sub { $handled++ };
pipe(my $note_r, my $note_w) or die "pipe: $!";
pipe(my $ack_r, my $ack_w) or die "pipe: $!";
my $pid = fork // die "fork: $!";
if ($pid == 0) {
    close $note_r; close $ack_w;
    for my $round (1 .. $rounds) {
        kill(USR1 => $parent) == 1 or die "kill: $!";
        syswrite($note_w, "n") == 1 or die "note: $!";
        sysread($ack_r, my $ack, 1) == 1 or die "ack: $!";
    }
    exit 0;
}
close $note_w; close $ack_r;
my $late = 0;
for my $round (1 .. $rounds) {
    my $read;
    do { $read = sysread($note_r, my $note, 1) } until defined $read;   # the signal interrupts the read
    $late++ if $handled < $round;
    syswrite($ack_w, "a") // die "ack: $!";
}
waitpid($pid, 0);
report("child exit=", $? >> 8, "; signals handled: $handled of $rounds; handled after the child had gone on: $late");
`)
	wantLines(t, got, "child exit=0; signals handled: 400 of 400; handled after the child had gone on: 0")
}

// A signal that arrives while the program is between two calls interrupts
// the call that has to wait, not one that is carried out at once: writing
// to a pipe with room and reading what is in it never fail with EINTR,
// however many signals come in.
func TestSignalsDoNotInterruptCallsThatNeedNotWait(t *testing.T) {
	got := runForkScript(t, `
my $parent = $$;
my $handled = 0;
$SIG{USR1} = sub { $handled++ };
pipe(my $started_r, my $started_w) or die "pipe: $!";
my $sender = fork // die "fork: $!";
if ($sender == 0) {
    close $started_r;
    kill(USR1 => $parent);
    syswrite($started_w, "s");
    kill(USR1 => $parent) while 1;
}
close $started_w;
my $started;
do { $started = sysread($started_r, my $byte, 1) } until defined $started;
pipe(my $r, my $w) or die "pipe: $!";
my ($writes, $reads) = (0, 0);
my %failed;
for my $round (1 .. 20000) {
    if (defined syswrite($w, "x")) { $writes++ } else { $failed{"write: $!"}++; next }
    if (defined sysread($r, my $byte, 1)) { $reads++ } else { $failed{"read: $!"}++ }
}
kill(KILL => $sender);
waitpid($sender, 0);
report("signals came in: ", ($handled > 0 ? "yes" : "no"), "; writes: $writes; reads: $reads; failed: ", join(", ", map { "$_ x $failed{$_}" } sort keys %failed));
`)
	wantLines(t, got, "signals came in: yes; writes: 20000; reads: 20000; failed: ")
}
