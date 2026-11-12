//go:build darwin || linux

package gperl_test

import (
	"path/filepath"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// A TCP server and its client in one process: the listening socket reports
// the port the host picked, select sees the listener readable only once a
// client has connected, and each end knows the other's address.
func TestTCPServerAndClientInOneProcess(t *testing.T) {
	got := runForkScript(t, `
use IO::Socket::INET;
use IO::Select;
my $server = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => 0, ReuseAddr => 1)
    or die "listen: $@";
my $port = $server->sockport;
report("listening: host=", $server->sockhost, " port assigned=", ($port > 0 ? 1 : 0));
report("reuse address reads back: ", $server->sockopt(SO_REUSEADDR) ? 1 : 0);

my $select = IO::Select->new($server);
report("idle listener readable: ", scalar(() = $select->can_read(0.2)));

my $client = IO::Socket::INET->new(PeerAddr => '127.0.0.1', PeerPort => $port, Proto => 'tcp')
    or die "connect: $@";
report("listener readable after connect: ", scalar(() = $select->can_read(30)));
my $conn = $server->accept or die "accept: $!";
report("accepted from ", $conn->peerhost, " client port matches=", ($conn->peerport == $client->sockport ? 1 : 0),
       " server port matches=", ($client->peerport == $port ? 1 : 0));
report("listener idle again: ", scalar(() = $select->can_read(0.2)));

print {$client} "hello server\n"; $client->flush;
my $line = <$conn>;
print {$conn} "hello client\n"; $conn->flush;
my $reply = <$client>;
chomp($line, $reply);
report("server read: $line; client read: $reply");

close $client;
report("end of input after the client closed: ", (defined(my $more = <$conn>) ? "more" : "eof"));

my $second = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => $port);
report("the port is taken: ", ($second ? "bound again" : ($!{EADDRINUSE} ? "EADDRINUSE" : "error $!")));
report("unknown option: ", (defined setsockopt($server, SOL_SOCKET, 9999, 1) ? "set" : ($!{ENOPROTOOPT} ? "ENOPROTOOPT" : "error $!")));
report("raw socket: ", (socket(my $raw, PF_INET, SOCK_RAW, 0) ? "created" : ($!{EPROTONOSUPPORT} ? "EPROTONOSUPPORT" : "error $!")));
`)
	wantLines(t, got,
		"listening: host=127.0.0.1 port assigned=1",
		"reuse address reads back: 1",
		"idle listener readable: 0",
		"listener readable after connect: 1",
		"accepted from 127.0.0.1 client port matches=1 server port matches=1",
		"listener idle again: 0",
		"server read: hello server; client read: hello client",
		"end of input after the client closed: eof",
		"the port is taken: EADDRINUSE",
		"unknown option: ENOPROTOOPT",
		"raw socket: EPROTONOSUPPORT",
	)
}

// A server that forks a child per connection, a forked client, and an
// accept that an alarm interrupts.
func TestTCPServerAcrossFork(t *testing.T) {
	got := runForkScript(t, `
use IO::Socket::INET;
my $server = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => 0) or die "listen: $@";
my $port = $server->sockport;

# the client is a forked child
my $client = fork // die "fork: $!";
if ($client == 0) {
    close $server;
    my $sock = IO::Socket::INET->new(PeerAddr => '127.0.0.1', PeerPort => $port) or die "connect: $@";
    print {$sock} "from the child\n";
    my $reply = <$sock>;
    exit($reply eq "ack from the child\n" ? 0 : 1);
}
my $conn = $server->accept or die "accept: $!";
my $line = <$conn>;
print {$conn} "ack $line";
close $conn;
waitpid($client, 0);
chomp $line;
report("server got: $line; client exit=", $? >> 8);

# the connection is served by a forked child, the parent is the client
my $worker = fork // die "fork: $!";
if ($worker == 0) {
    my $conn = $server->accept or die "accept: $!";
    my $line = <$conn>;
    print {$conn} "worker $$ saw: $line";
    close $conn;
    exit 0;
}
my $sock = IO::Socket::INET->new(PeerAddr => '127.0.0.1', PeerPort => $port) or die "connect: $@";
print {$sock} "from the parent\n";
my $reply = <$sock>;
close $sock;
waitpid($worker, 0);
report("reply came from the worker: ", ($reply eq "worker $worker saw: from the parent\n" ? 1 : 0), " exit=", $? >> 8);

# nobody connects: the alarm ends the accept
my $start = time;
my $ok = eval {
    local $SIG{ALRM} = sub { die "alarm\n" };
    alarm 1;
    $server->accept;
    alarm 0;
    1;
};
report("idle accept: ", ($ok ? "returned" : "died: $@" =~ s/\n//r), " quickly=", (time - $start < 10 ? 1 : 0));
`)
	wantLines(t, got,
		"server got: from the child; client exit=0",
		"reply came from the worker: 1 exit=0",
		"idle accept: died: alarm quickly=1",
	)
}

// The client is another process: a gperl started with the port to connect
// to.
func TestTCPServerWithASpawnedClient(t *testing.T) {
	shim := perlCommandShim(t)
	script := writeScript(t, "server.pl", `
use strict; use warnings;
use IO::Socket::INET;
my ($report, $perl) = @ARGV;
sub report { open(my $fh, '>>', $report) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
my $server = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => 0) or die "listen: $@";
my $port = $server->sockport;
my $pid = open(my $out, '-|', $perl, '-MIO::Socket::INET', '-e',
    'my $s = IO::Socket::INET->new(PeerAddr => "127.0.0.1", PeerPort => shift) or die "connect: $@";'
  . 'print {$s} "from another process\n"; my $r = <$s>; print "client read: $r";', $port) // die "spawn: $!";
my $conn = $server->accept or die "accept: $!";
my $line = <$conn>;
print {$conn} "served $line";
close $conn;
my $said = <$out>;
close $out;
chomp($line, $said);
report("server read: $line");
report("$said; exit=", $? >> 8);
`)
	report := filepath.Join(t.TempDir(), "report.txt")
	status, err := gperl.Run(script, []string{report, shim})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	wantLines(t, readLines(t, report),
		"server read: from another process",
		"client read: served from another process; exit=0",
	)
}

// socketpair gives two connected sockets, usable across fork.
func TestSocketpair(t *testing.T) {
	got := runForkScript(t, `
use Socket;
socketpair(my $parent, my $child, AF_UNIX, SOCK_STREAM, PF_UNSPEC) or die "socketpair: $!";
$_->autoflush(1) for $parent, $child;
my $pid = fork // die "fork: $!";
if ($pid == 0) {
    close $parent;
    my $line = <$child>;
    print {$child} "child got: $line";
    exit 0;
}
close $child;
print {$parent} "ping\n";
my $reply = <$parent>;
waitpid($pid, 0);
chomp $reply;
report("$reply; exit=", $? >> 8);
report("end of input once the child is gone: ", (defined(my $more = <$parent>) ? "more" : "eof"));
report("inet socketpair: ", (socketpair(my $a, my $b, AF_INET, SOCK_STREAM, 0) ? "made" : ($!{EOPNOTSUPP} ? "EOPNOTSUPP" : "error $!")));
`)
	wantLines(t, got,
		"child got: ping; exit=0",
		"end of input once the child is gone: eof",
		"inet socketpair: EOPNOTSUPP",
	)
}
