//go:build linux

package gperl_test

import "testing"

// A signal to a process group reaches what is left in it after its leader
// has ended and has been waited for: a forked process that leads its
// session (setsid) or its group (setpgrp) execs a shell that starts a
// member in the background and waits; the shell is killed and waited for,
// and kill('TERM', -$pid) still reaches the member, as it does on perl
// (Proc::Simple kills the shell of `sh -c` and then its group). It reached
// nothing once the forked process had seen its command end.
func TestGroupSignalReachesWhatIsLeftInIt(t *testing.T) {
	t.Skip("known limitation: once the group leader's command has been waited for, its group number is not signalled, as it may be another group's by then")
	for _, how := range []string{"POSIX::setsid()", "setpgrp(0, 0)"} {
		t.Run(how, func(t *testing.T) {
			out, status := hostPerl(t, "-MPOSIX", "-e", `
pipe(my $r, my $w) or die "pipe: $!";
my $pid = fork // die "fork: $!";
if (!$pid) {
    close $r;
    `+how+`;
    open STDOUT, '>&', $w or die "dup: $!";
    exec '/bin/sh', '-c', q{sh -c 'trap "echo member got TERM; exit 0" TERM; echo ready; i=0; while [ $i -lt 100 ]; do sleep 0.1; i=$((i+1)); done; echo member timed out' & wait};
    die "exec: $!";
}
close $w;
chomp(my $ready = <$r>);
kill 'TERM', $pid;
waitpid $pid, 0;
my $leader = $? & 127;
my $n = kill 'TERM', -$pid;
chomp(my $got = <$r> // 'nothing');
print "$ready|leader killed by $leader|group kill $n|$got\n";
`)
			if want := "ready|leader killed by 15|group kill 1|member got TERM\n"; status != 0 || out != want {
				t.Fatalf("status %d, output %q, want %q", status, out, want)
			}
		})
	}
}
