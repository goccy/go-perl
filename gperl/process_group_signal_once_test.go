//go:build linux

package gperl_test

import "testing"

// A signal to a process group reaches a member that has become a command
// (exec) once: a forked process leads its group (setpgrp) and execs a
// program that counts the INTs its handler gets; kill('INT', -$pid) is
// one, as on perl. The command got it twice, once through its group and
// once more from the process it replaced, which passed on what was posted
// to it (App::Access2CSV's handler of INT then ran again after the program
// had dealt with the first, and its exit status was not its own).
func TestGroupSignalReachesAnExecdMemberOnce(t *testing.T) {
	out, status := hostPerl(t, "-e", `
pipe(my $r, my $w) or die "pipe: $!";
my $pid = fork // die "fork: $!";
if (!$pid) {
    close $r;
    setpgrp(0, 0);
    open STDOUT, '>&', $w or die "dup: $!";
    exec $^X, '-e', q{my $n = 0; $SIG{INT} = sub { $n++ }; $| = 1; print "ready\n"; my $t = time; select undef, undef, undef, 0.1 while time - $t < 2; print "INT $n\n"};
    die "exec: $!";
}
close $w;
chomp(my $ready = <$r>);
my $k = kill 'INT', -$pid;
waitpid $pid, 0;
my $status = $?;
chomp(my $got = <$r> // 'nothing');
print "$ready|group kill $k|$got|$status\n";
`)
	if want := "ready|group kill 1|INT 1|0\n"; status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}
