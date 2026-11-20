//go:build darwin || linux

package gperl_test

import "testing"

// recv and send take their flags: looking at input without taking it, not
// waiting on a socket that otherwise waits, waiting for all of it, and a
// send that fails with EPIPE without the signal.
func TestSendAndRecvFlags(t *testing.T) {
	got := runForkScript(t, `
use Socket qw(MSG_PEEK MSG_DONTWAIT MSG_WAITALL MSG_NOSIGNAL);
use IO::Socket::INET;
use POSIX ();
my $listener = IO::Socket::INET->new(Listen => 5, LocalAddr => "127.0.0.1", LocalPort => 0) or die "listen: $!";
my $client = IO::Socket::INET->new(PeerAddr => "127.0.0.1", PeerPort => $listener->sockport) or die "connect: $!";
my $server = $listener->accept or die "accept: $!";
my $buf;

report("nothing there, not waiting: ", (defined recv($server, $buf, 16, MSG_DONTWAIT) ? "received" : ($!{EAGAIN} ? "EAGAIN" : "error $!")));
syswrite($client, "hello");
for my $look (1, 2) {
    defined recv($server, $buf, 16, MSG_PEEK) or die "peek: $!";
    report("peek $look: $buf");
}
defined recv($server, $buf, 2, 0) or die "recv: $!";
report("recv 2: $buf");
defined recv($server, $buf, 16, 0) or die "recv: $!";
report("recv the rest: $buf");

my $pid = fork // die "fork: $!";
if (!$pid) { syswrite($client, "ab"); select(undef, undef, undef, 0.2); syswrite($client, "cdef"); POSIX::_exit(0) }
defined recv($server, $buf, 4, MSG_WAITALL) or die "recv: $!";
report("recv all 4: $buf");
waitpid($pid, 0);
defined recv($server, $buf, 16, 0) or die "recv: $!";
report("what was left: $buf");

report("send with flags: ", send($client, "more", MSG_NOSIGNAL | MSG_DONTWAIT));
defined recv($server, $buf, 16, 0) or die "recv: $!";
report("received: $buf");

my $pipe_signals = 0;
local $SIG{PIPE} = sub { $pipe_signals++ };
close $server;
my $sent = 1;
for (1 .. 200) {
    $sent = send($client, "into the void", MSG_NOSIGNAL);
    last unless defined $sent;
    select(undef, undef, undef, 0.01);
}
report("send to a closed peer: ", (defined $sent ? "sent" : (($!{EPIPE} || $!{ECONNRESET}) ? "refused" : "error $!")), ", SIGPIPE ${pipe_signals}x");
`)
	wantLines(t, got,
		"nothing there, not waiting: EAGAIN",
		"peek 1: hello",
		"peek 2: hello",
		"recv 2: he",
		"recv the rest: llo",
		"recv all 4: abcd",
		"what was left: ef",
		"send with flags: 4",
		"received: more",
		"send to a closed peer: refused, SIGPIPE 0x",
	)
}
