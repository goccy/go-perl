package gperl_test

import "testing"

// shutdown takes 0, 1 or 2 — stop reading, stop writing, both — as programs
// pass it, and the names Socket exports have those values.
func TestShutdownTakesTheUsualNumbers(t *testing.T) {
	got := runForkScript(t, `
use Socket;
use IO::Socket::INET;
report("SHUT_RD=", SHUT_RD, " SHUT_WR=", SHUT_WR, " SHUT_RDWR=", SHUT_RDWR);

my $listener = IO::Socket::INET->new(Listen => 5, LocalAddr => "127.0.0.1", LocalPort => 0) or die "listen: $!";
sub connection {
    my $client = IO::Socket::INET->new(PeerAddr => "127.0.0.1", PeerPort => $listener->sockport) or die "connect: $!";
    my $server = $listener->accept or die "accept: $!";
    $_->autoflush(1) for $client, $server;
    return ($client, $server);
}

# The client has said all it has to say: 1 stops its writing, and only that.
my ($client, $server) = connection();
print {$client} "request\n";
report("shutdown(1): ", (shutdown($client, 1) ? "ok" : "failed: $!"));
my @request = <$server>;
report("server read to end of input: ", scalar(@request), " line(s)");
print {$server} "reply\n";
my $reply = <$client>;
chomp $reply;
report("client still reads: $reply");

# 0 stops the reading side; the writing side goes on.
($client, $server) = connection();
report("shutdown(0): ", (shutdown($server, 0) ? "ok" : "failed: $!"));
print {$server} "more\n";
my $more = <$client>;
chomp $more;
report("the peer reads what is written after it: $more");

# 2 stops both.
($client, $server) = connection();
report("shutdown(2): ", (shutdown($server, 2) ? "ok" : "failed: $!"));
report("the peer is at end of input: ", (defined(scalar <$client>) ? "more" : "eof"));
report("shutdown(3): ", (shutdown($client, 3) ? "ok" : ($!{EINVAL} ? "EINVAL" : "error $!")));
`)
	wantLines(t, got,
		"SHUT_RD=0 SHUT_WR=1 SHUT_RDWR=2",
		"shutdown(1): ok",
		"server read to end of input: 1 line(s)",
		"client still reads: reply",
		"shutdown(0): ok",
		"the peer reads what is written after it: more",
		"shutdown(2): ok",
		"the peer is at end of input: eof",
		"shutdown(3): EINVAL",
	)
}

// The poll and socket-option names programs build their masks from exist.
func TestPollAndSocketOptionNames(t *testing.T) {
	got := runForkScript(t, `
use IO::Poll qw(POLLIN POLLOUT POLLRDNORM POLLWRNORM POLLRDBAND POLLWRBAND POLLNORM POLLPRI POLLERR POLLHUP POLLNVAL);
use Socket qw(SO_ACCEPTCONN SO_RCVTIMEO SO_SNDTIMEO SOL_SOCKET);
my %bit = (POLLIN => POLLIN, POLLOUT => POLLOUT, POLLRDBAND => POLLRDBAND, POLLWRBAND => POLLWRBAND,
           POLLPRI => POLLPRI, POLLERR => POLLERR, POLLHUP => POLLHUP, POLLNVAL => POLLNVAL);
my %seen; $seen{$_}++ for values %bit;
report("distinct event bits: ", (scalar(keys %seen) == scalar(keys %bit) ? "yes" : "no"));
report("POLLNORM is POLLRDNORM: ", (POLLNORM == POLLRDNORM ? "yes" : "no"));
report("socket options: ", join(" ", map { defined($_) ? "defined" : "undef" } SO_ACCEPTCONN, SO_RCVTIMEO, SO_SNDTIMEO));

# A mask with the band bits is accepted, and reports what is there.
pipe(my $r, my $w) or die "pipe: $!";
syswrite($w, "x");
my $poll = IO::Poll->new;
$poll->mask($r => POLLIN | POLLRDBAND);
$poll->mask($w => POLLOUT | POLLWRBAND);
my $n = $poll->poll(1);
report("poll: $n ready; read end readable=", ($poll->events($r) & POLLIN ? 1 : 0), " write end writable=", ($poll->events($w) & POLLOUT ? 1 : 0));
`)
	wantLines(t, got,
		"distinct event bits: yes",
		"POLLNORM is POLLRDNORM: yes",
		"socket options: defined defined defined",
		"poll: 2 ready; read end readable=1 write end writable=1",
	)
}
