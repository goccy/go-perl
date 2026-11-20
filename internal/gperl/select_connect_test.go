//go:build darwin || linux

package gperl_test

import "testing"

// select reports a socket whose non-blocking connect failed writable, and
// SO_ERROR then gives the error, as on the host: event loops wait for a
// connect to end that way (POE's SocketFactory, Future-IO). The socket
// connects to a port of the loopback address that was just let go, so
// nothing listens there. select used to return 0 at once without the
// socket in any set, and such a loop spun.
func TestSelectReportsAFailedConnectWritable(t *testing.T) {
	out, status := hostPerl(t, "-e", `
use Socket qw(PF_INET SOCK_STREAM SOL_SOCKET SO_ERROR SO_REUSEADDR INADDR_LOOPBACK pack_sockaddr_in unpack_sockaddr_in);
use Fcntl qw(F_GETFL F_SETFL O_NONBLOCK);
use Errno qw(EINPROGRESS ECONNREFUSED);
socket(my $l, PF_INET, SOCK_STREAM, 0) or die "socket: $!";
bind($l, pack_sockaddr_in(0, INADDR_LOOPBACK)) or die "bind: $!";
my ($port) = unpack_sockaddr_in(getsockname($l));
close $l;
socket(my $s, PF_INET, SOCK_STREAM, 0) or die "socket: $!";
fcntl($s, F_SETFL, fcntl($s, F_GETFL, 0) | O_NONBLOCK) or die "fcntl: $!";
if (connect($s, pack_sockaddr_in($port, INADDR_LOOPBACK))) { print "connected\n"; exit }
my $err = $! + 0;
if ($err == EINPROGRESS) {
    my $w = '';
    vec($w, fileno($s), 1) = 1;
    my $n = select(undef, my $wout = $w, undef, 5);
    print "select: n=$n writable=", vec($wout, fileno($s), 1), "\n";
    $err = unpack('i', getsockopt($s, SOL_SOCKET, SO_ERROR));
} else {
    print "select: n=1 writable=1\n";
}
print "error: ", ($err == ECONNREFUSED ? "refused" : "$err"), "\n";
`)
	if want := "select: n=1 writable=1\nerror: refused\n"; status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}
