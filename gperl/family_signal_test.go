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
