//go:build darwin || linux

package gperl_test

import (
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"unsafe"
)

// kernelSockName is the address the kernel names a socket of the test's by:
// the bytes of the host's struct sockaddr.
func kernelSockName(t *testing.T, conn syscall.Conn) []byte {
	t.Helper()
	raw, err := conn.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 256)
	n := uint32(len(buf))
	var errno syscall.Errno
	if err := raw.Control(func(fd uintptr) {
		_, _, errno = syscall.Syscall(syscall.SYS_GETSOCKNAME, fd, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&n)))
	}); err != nil {
		t.Fatal(err)
	}
	if errno != 0 {
		t.Fatal(errno)
	}
	return buf[:n]
}

// The packed socket addresses a program holds are the host's structures:
// what Socket packs, what bind, connect and send take, what accept,
// getpeername, getsockname and recv return, what getaddrinfo returns and
// getnameinfo takes are the bytes the kernel names a socket by — the bytes
// a native library makes and takes. An address the program packs by hand
// the way it would for the system (pack 'S n a4 x8' on Linux) connects.
func TestPackedSocketAddressesAreTheSystems(t *testing.T) {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	inet := kernelSockName(t, l.(*net.TCPListener))
	port := l.Addr().(*net.TCPAddr).Port

	out, status := hostPerl(t, "-e", `
use strict; use warnings;
use Socket qw(:DEFAULT :addrinfo inet_pton);
my ($hex, $port) = @ARGV;
my $kernel = pack("H*", $hex);
print "pack_sockaddr_in: ", (pack_sockaddr_in($port, inet_aton("127.0.0.1")) eq $kernel ? "the system's" : unpack("H*", pack_sockaddr_in($port, inet_aton("127.0.0.1")))), "\n";
my ($p, $ip) = unpack_sockaddr_in($kernel);
print "unpack_sockaddr_in: $p ", inet_ntoa($ip), "\n";
print "sockaddr_family: ", (sockaddr_family($kernel) == AF_INET ? "AF_INET" : sockaddr_family($kernel)), "\n";
socket(my $c, AF_INET, SOCK_STREAM, 0) or die "socket: $!";
connect($c, $kernel) or die "connect: $!";
print "getpeername: ", (getpeername($c) eq $kernel ? "the system's" : unpack("H*", getpeername($c))), "\n";
my ($err, @res) = getaddrinfo("127.0.0.1", $port, { flags => AI_NUMERICHOST | AI_NUMERICSERV, socktype => SOCK_STREAM });
print "getaddrinfo: ", ($err ? "error $err" : $res[0]{addr} eq $kernel ? "the system's" : unpack("H*", $res[0]{addr})), "\n";
($err, my $host, my $service) = getnameinfo($kernel, NI_NUMERICHOST | NI_NUMERICSERV);
print "getnameinfo: $host $service\n";

socket(my $l, AF_INET, SOCK_STREAM, 0) or die; bind($l, pack_sockaddr_in(0, INADDR_LOOPBACK)) or die "bind: $!"; listen($l, 5) or die;
my $name = getsockname($l);
print "getsockname: ", (length($name) == length($kernel) && sockaddr_family($name) == AF_INET ? "an address of the system's" : unpack("H*", $name)), "\n";
socket(my $d, AF_INET, SOCK_STREAM, 0) or die; connect($d, $name) or die "connect: $!";
my $peer = accept(my $s, $l) or die "accept: $!";
print "accept: ", ($peer eq getsockname($d) ? "the client's address" : unpack("H*", $peer)), "\n";

socket(my $u1, AF_INET, SOCK_DGRAM, 0) or die; bind($u1, pack_sockaddr_in(0, INADDR_LOOPBACK)) or die;
socket(my $u2, AF_INET, SOCK_DGRAM, 0) or die; bind($u2, pack_sockaddr_in(0, INADDR_LOOPBACK)) or die;
send($u1, "datagram", 0, getsockname($u2)) or die "send: $!";
my $from = recv($u2, my $buf, 64, 0);
print "recv: $buf from ", ($from eq getsockname($u1) ? "the sender's address" : unpack("H*", $from // "")), "\n";

if ($^O eq "linux") {
    socket(my $h, AF_INET, SOCK_STREAM, 0) or die;
    connect($h, pack("S n a4 x8", AF_INET, $port, inet_aton("127.0.0.1"))) or die "connect by hand: $!";
}
print "packed by hand: connected\n";
`, hex.EncodeToString(inet), strconv.Itoa(port))
	want := "pack_sockaddr_in: the system's\n" +
		"unpack_sockaddr_in: " + strconv.Itoa(port) + " 127.0.0.1\n" +
		"sockaddr_family: AF_INET\n" +
		"getpeername: the system's\n" +
		"getaddrinfo: the system's\n" +
		"getnameinfo: 127.0.0.1 " + strconv.Itoa(port) + "\n" +
		"getsockname: an address of the system's\n" +
		"accept: the client's address\n" +
		"recv: datagram from the sender's address\n" +
		"packed by hand: connected\n"
	if status != 0 || out != want {
		t.Errorf("status %d\n got: %q\nwant: %q", status, out, want)
	}
}

// IPv6 and Unix-domain addresses are the host's structures too.
func TestPackedSocketAddressesOfOtherFamiliesAreTheSystems(t *testing.T) {
	skipWithoutUnixSockets(t)
	dir, err := os.MkdirTemp("/tmp", "gpsa")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "s")
	ul, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer ul.Close()
	unix := kernelSockName(t, ul.(*net.UnixListener))

	// Where the family is in the structure: two bytes in front on Linux,
	// the byte after the length byte on darwin.
	famOff, famLen := 0, 2
	if runtime.GOOS == "darwin" {
		famOff, famLen = 1, 1
	}
	inet6, port6 := "", "0"
	if l6, err := net.Listen("tcp6", "[::1]:0"); err == nil {
		defer l6.Close()
		inet6 = hex.EncodeToString(kernelSockName(t, l6.(*net.TCPListener)))
		port6 = strconv.Itoa(l6.Addr().(*net.TCPAddr).Port)
	}

	out, status := hostPerl(t, "-e", `
use strict; use warnings;
use Socket qw(:DEFAULT inet_pton pack_sockaddr_in6 unpack_sockaddr_in6 AF_INET6);
my ($unix_hex, $path, $inet6_hex, $port6, $fam_off, $fam_len) = @ARGV;
my $kernel = pack("H*", $unix_hex);
my $packed = pack_sockaddr_un($path);
print "pack_sockaddr_un: ", (substr($packed, $fam_off, $fam_len) eq substr($kernel, $fam_off, $fam_len) && substr($packed, 2, length $path) eq $path ? "the system's family and path" : unpack("H*", $packed)), "\n";
print "unpack_sockaddr_un: ", (unpack_sockaddr_un($packed) eq $path ? "the path" : unpack_sockaddr_un($packed)), "\n";
print "sockaddr_family: ", (sockaddr_family($packed) == AF_UNIX ? "AF_UNIX" : sockaddr_family($packed)), "\n";
socket(my $c, AF_UNIX, SOCK_STREAM, 0) or die "socket: $!";
connect($c, $packed) or die "connect: $!";
print "getpeername: ", (unpack_sockaddr_un(getpeername($c)) eq $path ? "the path" : unpack("H*", getpeername($c))), "\n";
socketpair(my $a, my $b, AF_UNIX, SOCK_STREAM, 0) or die;
my $unnamed = getsockname($a);
print "a socket without a name: ", (sockaddr_family($unnamed) == AF_UNIX ? "AF_UNIX" : unpack("H*", $unnamed)), "\n";
if ($inet6_hex ne "") {
    my $k6 = pack("H*", $inet6_hex);
    print "pack_sockaddr_in6: ", (pack_sockaddr_in6($port6, inet_pton(AF_INET6, "::1")) eq $k6 ? "the system's" : unpack("H*", pack_sockaddr_in6($port6, inet_pton(AF_INET6, "::1")))), "\n";
    my ($p) = unpack_sockaddr_in6($k6);
    print "unpack_sockaddr_in6: ", ($p == $port6 ? "the port" : $p), "\n";
    socket(my $s6, AF_INET6, SOCK_STREAM, 0) or die; connect($s6, $k6) or die "connect: $!";
    print "getpeername: ", (getpeername($s6) eq $k6 ? "the system's" : unpack("H*", getpeername($s6))), "\n";
} else {
    print "pack_sockaddr_in6: the system's\nunpack_sockaddr_in6: the port\ngetpeername: the system's\n";
}
`, hex.EncodeToString(unix), path, inet6, port6, strconv.Itoa(famOff), strconv.Itoa(famLen))
	want := "pack_sockaddr_un: the system's family and path\n" +
		"unpack_sockaddr_un: the path\n" +
		"sockaddr_family: AF_UNIX\n" +
		"getpeername: the path\n" +
		"a socket without a name: AF_UNIX\n" +
		"pack_sockaddr_in6: the system's\n" +
		"unpack_sockaddr_in6: the port\n" +
		"getpeername: the system's\n"
	if status != 0 || out != want {
		t.Errorf("status %d\n got: %q\nwant: %q", status, out, want)
	}
}
