//go:build darwin || linux

package gperl_test

import "testing"

// Datagram sockets (UDP over IPv4): a bound socket receives what is sent to
// it and learns who sent it, a connected one sends and receives without
// addresses, and IO::Socket::INET drives both.
func TestUDPSockets(t *testing.T) {
	got := runForkScript(t, `
use Socket;
use IO::Socket::INET;
use IO::Select;

socket(my $server, AF_INET, SOCK_DGRAM, getprotobyname('udp')) or die "socket: $!";
bind($server, pack_sockaddr_in(0, INADDR_LOOPBACK)) or die "bind: $!";
my ($port, $addr) = unpack_sockaddr_in(getsockname($server));
report("bound to ", inet_ntoa($addr), ", port chosen=", ($port > 0 ? 1 : 0),
       ", type=", (unpack("i", getsockopt($server, SOL_SOCKET, SO_TYPE)) == SOCK_DGRAM ? "datagram" : "other"));
report("nothing to receive yet: ", (IO::Select->new($server)->can_read(0.05) ? "readable" : "not readable"));

socket(my $client, AF_INET, SOCK_DGRAM, 0) or die "socket: $!";
my $to = pack_sockaddr_in($port, INADDR_LOOPBACK);
report("sent ", send($client, "first", 0, $to), " and ", send($client, "second datagram", 0, $to), " bytes");
report("now ", (IO::Select->new($server)->can_read(5) ? "readable" : "not readable"));
my $from = recv($server, my $buf, 1024, 0) // die "recv: $!";
my ($from_port, $from_addr) = unpack_sockaddr_in($from);
report("received '$buf' from ", inet_ntoa($from_addr), ", the sender's port=", ($from_port == (unpack_sockaddr_in(getsockname($client)))[0] ? 1 : 0));
recv($server, $buf, 6, 0) // die "recv: $!";
report("a short buffer takes the start of the datagram: '$buf'");
send($server, "reply", 0, $from) // die "send: $!";
recv($client, $buf, 1024, 0) // die "recv: $!";
report("the sender got '$buf' back");

# a connected datagram socket, through IO::Socket::INET, talking to a child
my $echo = IO::Socket::INET->new(Proto => 'udp', LocalAddr => '127.0.0.1', LocalPort => 0) or die "udp server: $@";
my $pid = fork // die "fork: $!";
if (!$pid) {
    for (1 .. 2) {
        my $peer = $echo->recv(my $msg, 1024) // exit 1;
        $echo->send(uc $msg, 0, $peer) // exit 2;
    }
    exit 0;
}
my $talk = IO::Socket::INET->new(Proto => 'udp', PeerAddr => '127.0.0.1', PeerPort => $echo->sockport) or die "udp client: $@";
report("connected to port=", ($talk->peerport == $echo->sockport ? "the server's" : "another"));
for my $msg ("hello", "again") {
    $talk->send($msg) // die "send: $!";
    $talk->recv(my $answer, 1024) // die "recv: $!";
    report("echo of '$msg': '$answer'");
}
waitpid($pid, 0);
report("the echo server exited with ", $? >> 8);

report("listen on a datagram socket: ", (listen($server, 5) ? "ok" : ($!{EOPNOTSUPP} || $!{ENOTSUP} ? "not supported" : "error $!")));
`)
	wantLines(t, got,
		"bound to 127.0.0.1, port chosen=1, type=datagram",
		"nothing to receive yet: not readable",
		"sent 5 and 15 bytes",
		"now readable",
		"received 'first' from 127.0.0.1, the sender's port=1",
		"a short buffer takes the start of the datagram: 'second'",
		"the sender got 'reply' back",
		"connected to port=the server's",
		"echo of 'hello': 'HELLO'",
		"echo of 'again': 'AGAIN'",
		"the echo server exited with 0",
		"listen on a datagram socket: not supported",
	)
}
