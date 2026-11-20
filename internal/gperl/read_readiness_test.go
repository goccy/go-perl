//go:build darwin || linux

package gperl_test

import (
	"runtime"
	"testing"
)

// What was written to a pipe or sent to a socket of this machine is there
// when the write has returned: a read, a receive or an accept that does not
// wait finds it, and select with no time to wait reports it — every time,
// however busy the machine is. (A datagram is there at once on Linux; the
// loopback of other systems delivers it after the send has returned, and a
// receive without waiting may be too early there for perl as well.)
func TestInputThatIsThereIsFoundWithoutWaiting(t *testing.T) {
	got := runForkScriptArgs(t, `
my $datagrams_at_once = shift @ARGV;
use IO::Socket::INET;
use Socket;
my $rounds = 3000;
sub readable_now { my $fh = shift; my $rin = ''; vec($rin, fileno($fh), 1) = 1; scalar select($rin, undef, undef, 0) }
sub settle { my $fh = shift; my $rin = ''; vec($rin, fileno($fh), 1) = 1; select($rin, undef, undef, 5) }
sub tally { my $seen = shift; join(", ", map { "$_ x $seen->{$_}" } sort keys %$seen) }

my $to = IO::Socket::INET->new(Proto => 'udp', LocalAddr => '127.0.0.1', LocalPort => 0) or die "socket: $!";
my $from = IO::Socket::INET->new(Proto => 'udp', PeerAddr => '127.0.0.1', PeerPort => $to->sockport) or die "socket: $!";
$to->blocking(0);
my %seen;
for my $round (1 .. $rounds) {
    defined $from->send("datagram $round") or die "send: $!";
    settle($to) unless $datagrams_at_once;
    my $sender = $to->recv(my $data, 100);
    $seen{defined $sender ? ($data eq "datagram $round" ? "received" : "received '$data'") : $!{EAGAIN} ? "EAGAIN" : "error $!"}++;
    if (!defined $sender) { settle($to); $to->recv($data, 100) }
}
report("datagrams: ", tally(\%seen));

socketpair(my $s1, my $s2, AF_UNIX, SOCK_STREAM, 0) or die "socketpair: $!";
$s2->blocking(0);
%seen = ();
for my $round (1 .. $rounds) {
    defined syswrite($s1, "x") or die "write: $!";
    my $n = sysread($s2, my $byte, 1);
    $seen{defined $n ? "read" : $!{EAGAIN} ? "EAGAIN" : "error $!"}++;
    if (!defined $n) { settle($s2); sysread($s2, $byte, 1) }
}
report("a stream socket: ", tally(\%seen));

pipe(my $r, my $w) or die "pipe: $!";
$r->blocking(0);
%seen = ();
for my $round (1 .. $rounds) {
    defined syswrite($w, "x") or die "write: $!";
    my $n = sysread($r, my $byte, 1);
    $seen{defined $n ? "read" : $!{EAGAIN} ? "EAGAIN" : "error $!"}++;
    if (!defined $n) { settle($r); sysread($r, $byte, 1) }
}
report("a pipe: ", tally(\%seen));

%seen = ();
for my $round (1 .. $rounds) {
    defined syswrite($w, "x") or die "write: $!";
    $seen{readable_now($r) ? "readable" : "not readable"}++;
    settle($r);
    sysread($r, my $byte, 1);
}
report("select on the pipe without waiting: ", tally(\%seen));

my $listener = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => 0) or die "listen: $!";
$listener->blocking(0);
%seen = ();
for my $round (1 .. 300) {
    my $client = IO::Socket::INET->new(PeerAddr => '127.0.0.1', PeerPort => $listener->sockport) or die "connect: $!";
    my $conn = $listener->accept;
    $seen{$conn ? "accepted" : $!{EAGAIN} ? "EAGAIN" : "error $!"}++;
    if (!$conn) { settle($listener); $listener->accept }
}
report("connections: ", tally(\%seen));
`, map[bool]string{true: "1", false: "0"}[runtime.GOOS == "linux"])
	wantLines(t, got,
		"datagrams: received x 3000",
		"a stream socket: read x 3000",
		"a pipe: read x 3000",
		"select on the pipe without waiting: readable x 3000",
		"connections: accepted x 300",
	)
}
