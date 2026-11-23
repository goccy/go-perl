package gperl_test

import (
	"net"
	"runtime"
	"testing"
)

// socketDir is a prelude for programs that bind Unix-domain sockets: $DIR is
// a fresh directory with a short path (the name of such a socket has room
// for about a hundred bytes), removed when the program ends.
const socketDir = `
use File::Temp ();
my $DIR = File::Temp::tempdir("gpXXXXXX", DIR => "/tmp", CLEANUP => 1);
`

func skipWithoutUnixSockets(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("the tests of Unix-domain sockets run on linux and darwin")
	}
}

func skipWithoutIPv6(t *testing.T) {
	t.Helper()
	l, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skipf("this machine has no IPv6 loopback: %v", err)
	}
	l.Close()
}

// A Unix-domain stream socket bound to a name: a forked client connects to
// it, both ends know the names, the socket file is a socket, stays when the
// listener is closed and keeps the name taken until it is unlinked.
func TestUnixStreamSocketByName(t *testing.T) {
	skipWithoutUnixSockets(t)
	got := runForkScript(t, socketDir+`
use IO::Socket::UNIX;
use Socket qw(SOCK_STREAM);
use Errno qw(EADDRINUSE ECONNREFUSED ENOENT);
my $path = "$DIR/srv.sock";
my $server = IO::Socket::UNIX->new(Local => $path, Listen => 5, Type => SOCK_STREAM) or die "listen: $!";
report("socket file: ", (-S $path ? "socket" : "something else"), "; listener at ", ($server->hostpath eq $path ? "its path" : $server->hostpath));

my $client = fork // die "fork: $!";
if ($client == 0) {
    close $server;
    my $sock = IO::Socket::UNIX->new(Peer => $path, Type => SOCK_STREAM) or die "connect: $!";
    print {$sock} "from the client\n";
    my $reply = <$sock>;
    exit($reply eq "ack from the client\n" && $sock->peerpath eq $path ? 0 : 1);
}
my $conn = $server->accept or die "accept: $!";
my $line = <$conn>;
print {$conn} "ack $line";
report("accepted at ", ($conn->hostpath eq $path ? "its path" : $conn->hostpath), "; peer name: '", ($conn->peerpath // ''), "'");
waitpid($client, 0);   # the client asks for its peer's name while the connection stands
close $conn;
chomp $line;
report("server got: $line; client exit=", $? >> 8);

close $server;
report("after close the file is ", (-S $path ? "still there" : "gone"));
my $late = IO::Socket::UNIX->new(Peer => $path, Type => SOCK_STREAM);
report("connect to a closed listener: ", ($late ? "connected" : $! == ECONNREFUSED ? "ECONNREFUSED" : "$!"));
my $again = IO::Socket::UNIX->new(Local => $path, Listen => 5, Type => SOCK_STREAM);
report("bind of the name that is taken: ", ($again ? "bound" : $! == EADDRINUSE ? "EADDRINUSE" : "$!"));
unlink $path or die "unlink: $!";
my $missing = IO::Socket::UNIX->new(Peer => $path, Type => SOCK_STREAM);
report("connect to a name that is gone: ", ($missing ? "connected" : $! == ENOENT ? "ENOENT" : "$!"));
$again = IO::Socket::UNIX->new(Local => $path, Listen => 5, Type => SOCK_STREAM);
report("bind after unlink: ", ($again ? "bound" : "$!"));

# a name relative to the working directory
chdir $DIR or die "chdir: $!";
my $rel = IO::Socket::UNIX->new(Local => "rel.sock", Listen => 1, Type => SOCK_STREAM) or die "listen: $!";
report("relative name: file ", (-S "$DIR/rel.sock" ? "in the working directory" : "elsewhere"), ", known as ", $rel->hostpath);
`)
	wantLines(t, got,
		"socket file: socket; listener at its path",
		"accepted at its path; peer name: ''",
		"server got: from the client; client exit=0",
		"after close the file is still there",
		"connect to a closed listener: ECONNREFUSED",
		"bind of the name that is taken: EADDRINUSE",
		"connect to a name that is gone: ENOENT",
		"bind after unlink: bound",
		"relative name: file in the working directory, known as rel.sock",
	)
}

// Unix-domain datagram sockets: a bound socket receives from a sender with
// a name and answers to it, and a datagram from a sender without a name has
// no return address.
func TestUnixDatagramSocketByName(t *testing.T) {
	skipWithoutUnixSockets(t)
	got := runForkScript(t, socketDir+`
use Socket qw(PF_UNIX SOCK_DGRAM pack_sockaddr_un unpack_sockaddr_un);
socket(my $server, PF_UNIX, SOCK_DGRAM, 0) or die "socket: $!";
bind($server, pack_sockaddr_un("$DIR/dg.sock")) or die "bind: $!";
report("bound to ", ((unpack_sockaddr_un(getsockname($server)))[0] eq "$DIR/dg.sock" ? "its path" : "another name"));

socket(my $anonymous, PF_UNIX, SOCK_DGRAM, 0) or die "socket: $!";
defined send($anonymous, "no return address", 0, pack_sockaddr_un("$DIR/dg.sock")) or die "send: $!";
my $from = recv($server, my $data, 100, 0);
report("received '$data' from ", (defined $from && length $from ? "a name" : "nobody"));

socket(my $client, PF_UNIX, SOCK_DGRAM, 0) or die "socket: $!";
bind($client, pack_sockaddr_un("$DIR/client.sock")) or die "bind: $!";
connect($client, pack_sockaddr_un("$DIR/dg.sock")) or die "connect: $!";
defined send($client, "with a return address", 0) or die "send: $!";
$from = recv($server, $data, 100, 0);
report("received '$data' from ", ((unpack_sockaddr_un($from))[0] eq "$DIR/client.sock" ? "the client's name" : "another name"));
defined send($server, "answer", 0, $from) or die "answer: $!";
recv($client, $data, 100, 0);
report("the client received '$data'");
report("datagrams keep their boundaries: ", do {
    send($client, "one", 0); send($client, "two", 0);
    recv($server, my $first, 100, 0); recv($server, my $second, 100, 0);
    "$first|$second";
});
`)
	wantLines(t, got,
		"bound to its path",
		"received 'no return address' from nobody",
		"received 'with a return address' from the client's name",
		"the client received 'answer'",
		"datagrams keep their boundaries: one|two",
	)
}

// socketpair makes a pair of every kind the Unix domain has, and reports
// the families that have no pairs as socketpair(2) does.
func TestSocketPairOfEveryKind(t *testing.T) {
	skipWithoutUnixSockets(t)
	got := runForkScript(t, `
use Socket qw(AF_UNIX AF_INET SOCK_STREAM SOCK_DGRAM SOCK_RAW PF_UNSPEC SOL_SOCKET SO_TYPE);
use Errno qw(EOPNOTSUPP ENOTSUP EPROTONOSUPPORT);
for my $kind ([stream => SOCK_STREAM], [datagram => SOCK_DGRAM]) {
    my ($name, $type) = @$kind;
    socketpair(my $a, my $b, AF_UNIX, $type, PF_UNSPEC) or die "socketpair($name): $!";
    my $is = unpack("i", getsockopt($a, SOL_SOCKET, SO_TYPE)) == $type ? "of that type" : "of another type";
    syswrite($a, "first"); syswrite($a, "second");
    sysread($b, my $data, 100);
    # across a fork, the other way
    my $pid = fork // die "fork: $!";
    if ($pid == 0) { close $a; syswrite($b, "from the child"); exit 0 }
    close $b;
    sysread($a, my $reply, 100);
    waitpid($pid, 0);
    report("$name pair $is: read '$data', then '$reply'");
}
report("pair of IPv4 sockets: ", (socketpair(my $a, my $b, AF_INET, SOCK_STREAM, 0) ? "made" : ($! == EOPNOTSUPP || $! == ENOTSUP) ? "not supported" : "$!"));
report("pair of raw sockets: ", (socketpair(my $c, my $d, AF_UNIX, SOCK_RAW, 0) ? "made" : $! == EPROTONOSUPPORT ? "EPROTONOSUPPORT" : "$!"));
`)
	wantLines(t, got,
		"stream pair of that type: read 'firstsecond', then 'from the child'",
		"datagram pair of that type: read 'first', then 'from the child'",
		"pair of IPv4 sockets: not supported",
		"pair of raw sockets: EPROTONOSUPPORT",
	)
}

// IPv6 sockets: a stream server and client with their names, datagrams,
// and IPV6_V6ONLY deciding whether a socket bound to every address takes
// IPv4 connections as well. An IPv4 address given to bind is the host's
// error: EINVAL on Linux, whose bind takes it as an address too short for
// the socket's family, EAFNOSUPPORT elsewhere.
func TestIPv6Sockets(t *testing.T) {
	skipWithoutIPv6(t)
	got := runForkScript(t, `
use Socket qw(AF_INET AF_INET6 SOCK_STREAM SOCK_DGRAM IPPROTO_IPV6 IPV6_V6ONLY
              pack_sockaddr_in6 unpack_sockaddr_in6 pack_sockaddr_in inet_pton inet_ntop inet_aton);
use Errno qw(ECONNREFUSED EAFNOSUPPORT EINVAL);
my $loop = inet_pton(AF_INET6, "::1");
socket(my $server, AF_INET6, SOCK_STREAM, 0) or die "socket: $!";
report("bind to an IPv4 address: ", (bind($server, pack_sockaddr_in(0, inet_aton("127.0.0.1"))) ? "bound" : $! == EAFNOSUPPORT ? "EAFNOSUPPORT" : $! == EINVAL ? "EINVAL" : "$!"));
bind($server, pack_sockaddr_in6(0, $loop)) or die "bind: $!";
listen($server, 5) or die "listen: $!";
my ($port, $addr) = unpack_sockaddr_in6(getsockname($server));
report("listening at ", inet_ntop(AF_INET6, $addr), " on ", ($port ? "a port" : "no port"));

my $pid = fork // die "fork: $!";
if ($pid == 0) {
    socket(my $client, AF_INET6, SOCK_STREAM, 0) or die "socket: $!";
    connect($client, pack_sockaddr_in6($port, $loop)) or die "connect: $!";
    my ($pport, $paddr) = unpack_sockaddr_in6(getpeername($client));
    syswrite($client, "over IPv6 to " . inet_ntop(AF_INET6, $paddr) . ($pport == $port ? "" : " on another port"));
    exit 0;
}
my $peer = accept(my $conn, $server) or die "accept: $!";
sysread($conn, my $data, 100);
waitpid($pid, 0);
report("server got '$data' from ", inet_ntop(AF_INET6, (unpack_sockaddr_in6($peer))[1]));

socket(my $receiver, AF_INET6, SOCK_DGRAM, 0) or die "socket: $!";
bind($receiver, pack_sockaddr_in6(0, $loop)) or die "bind: $!";
my ($dport) = unpack_sockaddr_in6(getsockname($receiver));
socket(my $sender, AF_INET6, SOCK_DGRAM, 0) or die "socket: $!";
defined send($sender, "datagram", 0, pack_sockaddr_in6($dport, $loop)) or die "send: $!";
my $from = recv($receiver, $data, 100, 0);
my ($sport, $saddr) = unpack_sockaddr_in6($from);
report("received '$data' from ", inet_ntop(AF_INET6, $saddr), ($sport == (unpack_sockaddr_in6(getsockname($sender)))[0] ? " at the sender's port" : " at another port"));

for my $only (1, 0) {
    socket(my $l, AF_INET6, SOCK_STREAM, 0) or die "socket: $!";
    setsockopt($l, IPPROTO_IPV6, IPV6_V6ONLY, pack("i", $only)) or die "setsockopt: $!";
    bind($l, pack_sockaddr_in6(0, inet_pton(AF_INET6, "::"))) or die "bind: $!";
    listen($l, 5) or die "listen: $!";
    my ($p) = unpack_sockaddr_in6(getsockname($l));
    socket(my $v4, AF_INET, SOCK_STREAM, 0) or die "socket: $!";
    my $ok = connect($v4, pack_sockaddr_in($p, inet_aton("127.0.0.1")));
    report("IPV6_V6ONLY ", unpack("i", getsockopt($l, IPPROTO_IPV6, IPV6_V6ONLY)), ": an IPv4 connection is ",
        ($ok ? "taken" : $! == ECONNREFUSED ? "refused" : "$!"));
}
`)
	bindError := "EAFNOSUPPORT"
	if runtime.GOOS == "linux" {
		bindError = "EINVAL"
	}
	wantLines(t, got,
		"bind to an IPv4 address: "+bindError,
		"listening at ::1 on a port",
		"server got 'over IPv6 to ::1' from ::1",
		"received 'datagram' from ::1 at the sender's port",
		"IPV6_V6ONLY 1: an IPv4 connection is refused",
		"IPV6_V6ONLY 0: an IPv4 connection is taken",
	)
}

// A connect on a non-blocking socket does not wait: the socket becomes
// writable when the attempt has ended and SO_ERROR tells how. A blocking
// connect that waits is ended by a signal, and IO::Socket's timeout bounds
// one.
func TestConnectDoesNotBlockTheProgram(t *testing.T) {
	got := runForkScript(t, `
use IO::Socket::INET;
use IO::Select;
use Socket qw(SOL_SOCKET SO_ERROR pack_sockaddr_in inet_aton);
use Errno qw(EINPROGRESS ECONNREFUSED EISCONN EINTR ETIMEDOUT);
use Time::HiRes qw(time alarm);

my $server = IO::Socket::INET->new(Listen => 5, LocalAddr => '127.0.0.1', LocalPort => 0) or die "listen: $@";
my $port = $server->sockport;

my $sock = IO::Socket::INET->new(Proto => 'tcp', Blocking => 0) or die "socket: $@";
my $rc = connect($sock, pack_sockaddr_in($port, inet_aton('127.0.0.1')));
my $first = $rc ? "done" : $! == EINPROGRESS ? "EINPROGRESS" : "$!";
my $writable = IO::Select->new($sock)->can_write(5) ? 1 : 0;
my $err = unpack("i", getsockopt($sock, SOL_SOCKET, SO_ERROR));
my $again = connect($sock, pack_sockaddr_in($port, inet_aton('127.0.0.1'))) ? "done" : $! == EISCONN ? "EISCONN" : "$!";
report("non-blocking connect: ", ($first eq "done" || $first eq "EINPROGRESS" ? "started" : $first), ", writable=$writable, SO_ERROR=$err, again=$again, peer port ", ($sock->peerport == $port ? "right" : "wrong"));

# nobody listens there any more
my $gone = IO::Socket::INET->new(Listen => 1, LocalAddr => '127.0.0.1', LocalPort => 0) or die "listen: $@";
my $refused = $gone->sockport;
close $gone;
$sock = IO::Socket::INET->new(Proto => 'tcp', Blocking => 0) or die "socket: $@";
$rc = connect($sock, pack_sockaddr_in($refused, inet_aton('127.0.0.1')));
my $how = "refused at once";
if (!$rc && $! == EINPROGRESS) {
    $writable = IO::Select->new($sock)->can_write(5) ? 1 : 0;
    $! = unpack("i", getsockopt($sock, SOL_SOCKET, SO_ERROR));
    $how = "writable=$writable then SO_ERROR";
}
report("refused connect: ", ($rc ? "connected" : $! == ECONNREFUSED ? "ECONNREFUSED" : "$!"));

# an address that does not answer (reserved for documentation)
my $start = time;
my $timed = IO::Socket::INET->new(PeerAddr => '192.0.2.1', PeerPort => 9, Timeout => 0.5);
report("connect with a timeout: ", ($timed ? "connected" : "failed"), " in ", (time - $start < 5 ? "time" : "too long"));

$start = time;
my $ok = eval {
    local $SIG{ALRM} = sub { die "alarm\n" };
    alarm 0.3;
    my $s = IO::Socket::INET->new(Proto => 'tcp') or die "socket: $@";
    connect($s, pack_sockaddr_in(9, inet_aton('192.0.2.1')));
    alarm 0;
    1;
};
report("blocking connect under an alarm: ", ($ok ? "returned" : "died: $@" =~ s/\n//r), " in ", (time - $start < 5 ? "time" : "too long"));
`)
	// What a second connect on the socket answers once the first has
	// ended is the system's: Linux reports the end of the attempt it was
	// told is in progress with success, once; macOS says the socket is
	// connected.
	again := "EISCONN"
	if runtime.GOOS == "linux" {
		again = "done"
	}
	wantLines(t, got,
		"non-blocking connect: started, writable=1, SO_ERROR=0, again="+again+", peer port right",
		"refused connect: ECONNREFUSED",
		"connect with a timeout: failed in time",
		"blocking connect under an alarm: died: alarm in time",
	)
}

// getaddrinfo without a host: the addresses a server binds to (AI_PASSIVE),
// which "*" asks for as well, and the loopback addresses otherwise; with
// the service by number or by name, once per socket type.
func TestGetaddrinfoWithoutAHost(t *testing.T) {
	got := runForkScript(t, `
use Socket qw(getaddrinfo getnameinfo AI_PASSIVE AI_NUMERICHOST NI_NUMERICHOST NI_NUMERICSERV AF_INET AF_INET6 AF_UNSPEC SOCK_STREAM SOCK_DGRAM IPPROTO_TCP);
sub show {
    my ($err, @res) = @_;
    return "error: $err" if $err;
    return join ", ", map {
        my ($e, $host, $serv) = getnameinfo($_->{addr}, NI_NUMERICHOST | NI_NUMERICSERV);
        ($_->{family} == AF_INET ? "inet" : $_->{family} == AF_INET6 ? "inet6" : "?") . " "
          . ($_->{socktype} == SOCK_STREAM ? "stream" : $_->{socktype} == SOCK_DGRAM ? "dgram" : "?") . " $host:$serv";
    } @res;
}
report("passive: ", show(getaddrinfo(undef, "8080", { flags => AI_PASSIVE, socktype => SOCK_STREAM })));
report("star: ", show(getaddrinfo("*", "8080", { flags => AI_PASSIVE, socktype => SOCK_STREAM, family => AF_INET })));
report("loopback: ", show(getaddrinfo(undef, "8080", { socktype => SOCK_STREAM })));
report("inet6 only: ", show(getaddrinfo(undef, "8080", { flags => AI_PASSIVE, socktype => SOCK_DGRAM, family => AF_INET6 })));
report("every type: ", show(getaddrinfo("127.0.0.1", "53", { flags => AI_NUMERICHOST })));
report("numeric IPv6: ", show(getaddrinfo("::1", "80", { flags => AI_NUMERICHOST, socktype => SOCK_STREAM })));
report("neither: ", show(getaddrinfo(undef, undef)) =~ /^error/ ? "error" : "addresses");
report("not a number: ", show(getaddrinfo("localhost", "80", { flags => AI_NUMERICHOST })) =~ /^error/ ? "error" : "addresses");
`)
	wantLines(t, got,
		"passive: inet stream 0.0.0.0:8080, inet6 stream :::8080",
		"star: inet stream 0.0.0.0:8080",
		"loopback: inet stream 127.0.0.1:8080, inet6 stream ::1:8080",
		"inet6 only: inet6 dgram :::8080",
		"every type: inet stream 127.0.0.1:53, inet dgram 127.0.0.1:53",
		"numeric IPv6: inet6 stream ::1:80",
		"neither: error",
		"not a number: error",
	)
}

// The name of a Unix-domain socket is reported whole whatever its length:
// by getsockname, by getpeername and by accept. (The interpreter took such
// names into a buffer of 48 bytes, cut them after 46 bytes of path and set
// the end of the string past the buffer.)
func TestLongUnixSocketNames(t *testing.T) {
	skipWithoutUnixSockets(t)
	got := runForkScript(t, socketDir+`
use Socket qw(AF_UNIX SOCK_STREAM pack_sockaddr_un unpack_sockaddr_un);
my $canary = "x" x 64;
for my $length (30, 46, 47, 70, 100) {
    my $path = "$DIR/" . ("s" x ($length - length("$DIR/")));
    socket(my $listener, AF_UNIX, SOCK_STREAM, 0) or die "socket: $!";
    bind($listener, pack_sockaddr_un($path)) or die "bind: $!";
    listen($listener, 5) or die "listen: $!";
    socket(my $client, AF_UNIX, SOCK_STREAM, 0) or die "socket: $!";
    connect($client, pack_sockaddr_un($path)) or die "connect: $!";
    my $from = accept(my $conn, $listener) or die "accept: $!";
    my %name = (
        getsockname => unpack_sockaddr_un(getsockname($listener)),
        getpeername => unpack_sockaddr_un(getpeername($client)),
        "the accepted socket" => unpack_sockaddr_un(getsockname($conn)),
    );
    report("a path of ", length($path), " bytes: ", join(", ", map { "$_ " . ($name{$_} eq $path ? "right" : "wrong (" . length($name{$_}) . " bytes)") } sort keys %name));
}
report("what lay beside the names: ", ($canary eq "x" x 64 ? "untouched" : "overwritten"));
`)
	wantLines(t, got,
		"a path of 30 bytes: getpeername right, getsockname right, the accepted socket right",
		"a path of 46 bytes: getpeername right, getsockname right, the accepted socket right",
		"a path of 47 bytes: getpeername right, getsockname right, the accepted socket right",
		"a path of 70 bytes: getpeername right, getsockname right, the accepted socket right",
		"a path of 100 bytes: getpeername right, getsockname right, the accepted socket right",
		"what lay beside the names: untouched",
	)
}

// An abstract Unix-domain name (Linux: one that starts with a NUL) is the
// bytes the address has, NULs included, and comes back as long as it is. A
// socket is bound to the name pack_sockaddr_un makes of "\0" (the whole
// field, NULs alone) and found under it; a name with text in it, or with a
// NUL inside, is reported by getsockname and getpeername without padding;
// a socket without a name reports none.
func TestAbstractUnixSocketNames(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("abstract Unix-domain names are Linux's")
	}
	got := runForkScript(t, `
use Socket qw(AF_UNIX SOCK_STREAM pack_sockaddr_un unpack_sockaddr_un);
sub shown { my $name = shift; length($name) . " bytes, " . join("", map { $_ eq "\0" ? "\\0" : $_ } split //, $name) }
sub listener {
    my $address = shift;
    socket(my $listener, AF_UNIX, SOCK_STREAM, 0) or die "socket: $!";
    bind($listener, $address) or return;
    listen($listener, 5) or die "listen: $!";
    $listener;
}
sub exchange {
    my ($listener, $address) = @_;
    socket(my $client, AF_UNIX, SOCK_STREAM, 0) or die "socket: $!";
    connect($client, $address) or return "connect: $!";
    accept(my $conn, $listener) or return "accept: $!";
    syswrite($client, "over\n");
    my $line = <$conn>;
    chomp $line;
    my $peer = unpack_sockaddr_un(getpeername($client));
    my $unnamed = unpack_sockaddr_un(getsockname($client));
    "received: $line; the peer is " . shown($peer =~ s/-$$\z//r) . "; the client has " . (length($unnamed) ? "the name " . shown($unnamed) : "no name");
}

# the name of NULs alone
my $nuls = pack_sockaddr_un("\0");
my $listener = listener($nuls);
report("bound to the name of \"\\0\": ", $listener ? "yes" : "no: $!");
report("a second socket with that name: ", listener($nuls) ? "bound" : $!{EADDRINUSE} ? "EADDRINUSE" : "error $!");
socket(my $client, AF_UNIX, SOCK_STREAM, 0) or die "socket: $!";
report("connected by that name: ", connect($client, $nuls) ? "yes" : "no: $!");
my $reported = unpack_sockaddr_un(getsockname($listener));
socket(my $second, AF_UNIX, SOCK_STREAM, 0) or die "socket: $!";
report("the name reported has ", length($reported), " bytes, ", ($reported =~ /\A\0+\z/ ? "NULs alone" : "not NULs alone"), "; connected by it: ", connect($second, pack_sockaddr_un($reported)) ? "yes" : "no: $!");

for my $name ("\0gperl-test-$$", "\0with\0inside-$$") {
    my $listener = listener(pack_sockaddr_un($name)) or die "bind: $!";
    my $bound = unpack_sockaddr_un(getsockname($listener));
    report("bound to ", shown($bound =~ s/-$$\z//r), ($bound eq $name ? "" : " (not the name given)"), "; ", exchange($listener, pack_sockaddr_un($bound)));
}
`)
	wantLines(t, got,
		`bound to the name of "\0": yes`,
		"a second socket with that name: EADDRINUSE",
		"connected by that name: yes",
		"the name reported has 108 bytes, NULs alone; connected by it: yes",
		`bound to 11 bytes, \0gperl-test; received: over; the peer is 11 bytes, \0gperl-test; the client has no name`,
		`bound to 12 bytes, \0with\0inside; received: over; the peer is 12 bytes, \0with\0inside; the client has no name`,
	)
}
