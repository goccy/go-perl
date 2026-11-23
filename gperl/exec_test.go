//go:build darwin || linux

package gperl_test

import "testing"

// A process that exec'd a command is that command to everyone else: a
// signal sent to it reaches the command, SIGKILL included, and whoever
// waits for it sees how the command ended.
func TestExecHandsTheProcessOverToTheCommand(t *testing.T) {
	got := runForkScript(t, `
use POSIX ();
sub run_child {
    my ($signal, @command) = @_;
    my $pid = fork // die "fork: $!";
    if (!$pid) { exec(@command) or POSIX::_exit(99) }
    if (defined $signal) {
        select(undef, undef, undef, 0.5);
        my $sent = kill($signal, $pid);
        my $start = time;
        waitpid($pid, 0);
        return "kill=$sent signal=" . ($? & 127) . " exit=" . ($? >> 8) . " promptly=" . (time - $start < 10 ? 1 : 0);
    }
    waitpid($pid, 0);
    return "signal=" . ($? & 127) . " exit=" . ($? >> 8);
}
report("exit status: ", run_child(undef, "/bin/sh", "-c", "exit 3"));
report("TERM: ", run_child("TERM", "/bin/sleep", "30"));
report("KILL: ", run_child("KILL", "/bin/sleep", "30"));
report("a command that kills itself: ", run_child(undef, "/bin/sh", "-c", 'kill -HUP $$'));
report("a signal the command handles: ", run_child("USR1", "/bin/sh", "-c", 'trap "exit 5" USR1; while :; do sleep 0.1; done'));
`)
	wantLines(t, got,
		"exit status: signal=0 exit=3",
		"TERM: kill=1 signal=15 exit=0 promptly=1",
		"KILL: kill=1 signal=9 exit=0 promptly=1",
		"a command that kills itself: signal=1 exit=0",
		"a signal the command handles: kill=1 signal=0 exit=5 promptly=1",
	)
}
