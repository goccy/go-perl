//go:build linux

package gperl_test

import (
	"os"
	"testing"
)

// The address functions and databases of Socket answer as glibc's do:
// inet_aton takes every notation glibc's does ("127.1", "2130706433",
// "0x7f000001", "127.0.0.01"), inet_ntop writes an IPv4-compatible IPv6
// address with its IPv4 part in dots, getaddrinfo answers for a raw
// socket, and getprotobyname finds the protocols of /etc/protocols.
func TestSocketAddressesAndDatabases(t *testing.T) {
	if _, err := os.Stat("/etc/protocols"); err != nil {
		t.Skip("no /etc/protocols")
	}
	out, status := hostPerl(t, "-e", `
use Socket qw(:DEFAULT inet_ntop getaddrinfo SOCK_RAW);
for my $addr ('127.1', '2130706433', '0x7f000001', '127.0.0.01', '1.2.3.256', '1.2.3.4.5') {
    my $packed = inet_aton($addr);
    print "inet_aton $addr: ", (defined $packed ? inet_ntoa($packed) : "undef"), "\n";
}
for my $hex ('00000000000000000000000001020304', '00000000000000000000ffff01020304',
             '00000000000000000000000000000001', '00000000000000000000000000000000',
             '20010db8000000000000000000000001', '20010db8000100000001000000000001') {
    print "inet_ntop $hex: ", inet_ntop(AF_INET6, pack('H*', $hex)), "\n";
}
my ($err, @res) = getaddrinfo("127.0.0.1", "", { socktype => SOCK_RAW, protocol => 1 });
print "getaddrinfo SOCK_RAW: ", ($err ? "error $err" : scalar(@res) . " result, type " . ($res[0]{socktype} == SOCK_RAW ? "raw" : $res[0]{socktype}) . ", protocol $res[0]{protocol}"), "\n";
for my $name (qw(gre sctp tcp icmp IPv6-ICMP)) {
    my @p = getprotobyname($name);
    print "getprotobyname $name: ", (@p ? "$p[0] $p[2]" : "none"), "\n";
}
my @n = getprotobynumber(47);
print "getprotobynumber 47: ", (@n ? $n[0] : "none"), "\n";
`)
	want := "inet_aton 127.1: 127.0.0.1\n" +
		"inet_aton 2130706433: 127.0.0.1\n" +
		"inet_aton 0x7f000001: 127.0.0.1\n" +
		"inet_aton 127.0.0.01: 127.0.0.1\n" +
		"inet_aton 1.2.3.256: undef\n" +
		"inet_aton 1.2.3.4.5: undef\n" +
		"inet_ntop 00000000000000000000000001020304: ::1.2.3.4\n" +
		"inet_ntop 00000000000000000000ffff01020304: ::ffff:1.2.3.4\n" +
		"inet_ntop 00000000000000000000000000000001: ::1\n" +
		"inet_ntop 00000000000000000000000000000000: ::\n" +
		"inet_ntop 20010db8000000000000000000000001: 2001:db8::1\n" +
		"inet_ntop 20010db8000100000001000000000001: 2001:db8:1:0:1::1\n" +
		"getaddrinfo SOCK_RAW: 1 result, type raw, protocol 1\n" +
		"getprotobyname gre: gre 47\n" +
		"getprotobyname sctp: sctp 132\n" +
		"getprotobyname tcp: tcp 6\n" +
		"getprotobyname icmp: icmp 1\n" +
		"getprotobyname IPv6-ICMP: ipv6-icmp 58\n" +
		"getprotobynumber 47: gre\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}

// The names programs import from Socket and POSIX on Linux are there, with
// Linux's numbers, and the options the host carries can be set: a UDP
// socket that broadcasts (SO_BROADCAST), a listener that defers accepting
// (TCP_DEFER_ACCEPT); a Unix-domain socket pair takes PF_UNIX as its
// protocol.
func TestSocketNamesAndOptionsOfLinux(t *testing.T) {
	out, status := hostPerl(t, "-e", `
use Socket qw(:all);
use POSIX ();
for my $c (qw(SO_BINDTODEVICE SO_ATTACH_FILTER SCM_RIGHTS)) {
    my $v = eval { Socket->can($c)->() };
    print "$c: ", (defined $v ? $v : "missing"), "\n";
}
print "SIGRTMIN: ", POSIX::SIGRTMIN(), " SIGRTMAX: ", POSIX::SIGRTMAX(), "\n";
socket(my $u, AF_INET, SOCK_DGRAM, 0) or die "socket: $!";
print "SO_BROADCAST: ", (setsockopt($u, SOL_SOCKET, SO_BROADCAST, 1) ? "set" : "failed: $!"), ", reads ", unpack('i', getsockopt($u, SOL_SOCKET, SO_BROADCAST)), "\n";
socket(my $l, AF_INET, SOCK_STREAM, 0) or die "socket: $!";
print "TCP_DEFER_ACCEPT: ", (setsockopt($l, IPPROTO_TCP, TCP_DEFER_ACCEPT, 5) ? "set" : "failed: $!"), "\n";
print "socketpair with PF_UNIX: ", (socketpair(my $a, my $b, AF_UNIX, SOCK_STREAM, PF_UNIX) ? "made" : "failed: $!"), "\n";
`)
	want := "SO_BINDTODEVICE: 25\nSO_ATTACH_FILTER: 26\nSCM_RIGHTS: 1\n" +
		"SIGRTMIN: 34 SIGRTMAX: 64\n" +
		"SO_BROADCAST: set, reads 1\n" +
		"TCP_DEFER_ACCEPT: set\n" +
		"socketpair with PF_UNIX: made\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
