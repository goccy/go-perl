//go:build darwin || linux

package gperl_test

import "testing"

// A listening socket says that it is one, and a socket's receive timeout
// ends a receive or an accept that waited that long.
func TestListeningStateAndSocketTimeouts(t *testing.T) {
	got := runForkScript(t, `
use Socket qw(SOL_SOCKET SO_ACCEPTCONN SO_RCVTIMEO SO_SNDTIMEO);
use IO::Socket::INET;
use Time::HiRes qw(time);
my $listener = IO::Socket::INET->new(Listen => 5, LocalAddr => "127.0.0.1", LocalPort => 0) or die "listen: $!";
my $client = IO::Socket::INET->new(PeerAddr => "127.0.0.1", PeerPort => $listener->sockport) or die "connect: $!";
my $server = $listener->accept or die "accept: $!";
report("accepting: listener=", unpack("i", getsockopt($listener, SOL_SOCKET, SO_ACCEPTCONN)),
       " connection=", unpack("i", getsockopt($client, SOL_SOCKET, SO_ACCEPTCONN)));

# the way programs build a timeout: two native longs
setsockopt($server, SOL_SOCKET, SO_RCVTIMEO, pack("l!l!", 0, 300_000)) or die "SO_RCVTIMEO: $!";
my $start = time;
my $n = sysread($server, my $buf, 16);
my $waited = time - $start;
report("read on an idle socket: ", (defined $n ? "returned $n" : ($!{EAGAIN} ? "EAGAIN" : "error $!")),
       " after about ", ($waited > 0.2 && $waited < 5 ? "0.3s" : sprintf("%.2fs", $waited)));
syswrite($client, "in time");
report("input that arrives is read: ", (sysread($server, $buf, 16) ? $buf : "error $!"));
setsockopt($server, SOL_SOCKET, SO_SNDTIMEO, pack("l!l!", 1, 0)) or die "SO_SNDTIMEO: $!";
report("a send within its timeout: ", syswrite($server, "ok"));

setsockopt($listener, SOL_SOCKET, SO_RCVTIMEO, pack("l!l!", 0, 300_000)) or die "SO_RCVTIMEO: $!";
$start = time;
my $accepted = accept(my $conn, $listener);
$waited = time - $start;
report("accept with nobody connecting: ", ($accepted ? "accepted" : ($!{EAGAIN} ? "EAGAIN" : "error $!")),
       " after about ", ($waited > 0.2 && $waited < 5 ? "0.3s" : sprintf("%.2fs", $waited)));
report("a timeout of the wrong size: ", (setsockopt($server, SOL_SOCKET, SO_RCVTIMEO, pack("C", 1)) ? "set" : ($!{EINVAL} ? "EINVAL" : "error $!")));
`)
	wantLines(t, got,
		"accepting: listener=1 connection=0",
		"read on an idle socket: EAGAIN after about 0.3s",
		"input that arrives is read: in time",
		"a send within its timeout: 2",
		"accept with nobody connecting: EAGAIN after about 0.3s",
		"a timeout of the wrong size: EINVAL",
	)
}
