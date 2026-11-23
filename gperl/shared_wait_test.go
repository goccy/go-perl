//go:build darwin || linux

package gperl_test

import "testing"

// Workers blocked on a descriptor they share — in accept on one listener,
// reading one pipe — each take the signal sent to them.
func TestEveryWorkerOnASharedDescriptorTakesItsSignal(t *testing.T) {
	got := runForkScript(t, `
use IO::Socket::INET;
use POSIX ();
sub round {
    my ($name, $wait) = @_;
    my @pids;
    for (1 .. 5) {
        my $pid = fork // die "fork: $!";
        if (!$pid) {
            $SIG{TERM} = sub { POSIX::_exit(7) };
            $wait->() while 1;
        }
        push @pids, $pid;
    }
    select(undef, undef, undef, 0.3);
    kill("TERM", $_) for @pids;
    my ($left, $by_handler, $deadline) = (scalar(@pids), 0, time + 10);
    while ($left && time < $deadline) {
        my $pid = waitpid(-1, POSIX::WNOHANG());
        if ($pid > 0) { $left--; $by_handler++ if $? >> 8 == 7 } else { select(undef, undef, undef, 0.05) }
    }
    report("$name: ", ($left ? "$left of 5 still blocked" : "all 5 ended"), ", $by_handler through their handler");
    kill("KILL", $_) for @pids;
    1 while waitpid(-1, 0) > 0;
}
my $listener = IO::Socket::INET->new(Listen => 5, LocalAddr => "127.0.0.1") or die "listen: $!";
round("accept", sub { $listener->accept }) for 1 .. 2;
pipe(my $r, my $w) or die "pipe: $!";
round("read", sub { sysread($r, my $buf, 1) }) for 1 .. 2;
round("select", sub { my $rin = ""; vec($rin, fileno($listener), 1) = 1; select($rin, undef, undef, undef) });

# writers of one full pipe
pipe(my $full_r, my $full_w) or die "pipe: $!";
{
    use Fcntl;
    my $flags = fcntl($full_w, F_GETFL, 0);
    fcntl($full_w, F_SETFL, $flags | O_NONBLOCK) or die "fcntl: $!";
    1 while defined syswrite($full_w, "x" x 4096);
    fcntl($full_w, F_SETFL, $flags) or die "fcntl: $!";
}
round("write", sub { syswrite($full_w, "more") }) for 1 .. 2;
round("select for writing", sub { my $win = ""; vec($win, fileno($full_w), 1) = 1; select(undef, $win, undef, undef) });
`)
	wantLines(t, got,
		"accept: all 5 ended, 5 through their handler",
		"accept: all 5 ended, 5 through their handler",
		"read: all 5 ended, 5 through their handler",
		"read: all 5 ended, 5 through their handler",
		"select: all 5 ended, 5 through their handler",
		"write: all 5 ended, 5 through their handler",
		"write: all 5 ended, 5 through their handler",
		"select for writing: all 5 ended, 5 through their handler",
	)
}
