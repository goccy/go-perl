package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// Datagram sockets are under the same network policy as stream sockets:
// with a zero Config a program can make one, but can neither take a local
// address nor send anywhere.
func TestUDPIsDeniedByDefault(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Socket;`,
		`sub outcome { $_[0] ? "allowed" : $!{EACCES} ? "EACCES" : "error $!" }`,
		`socket(my $s, AF_INET, SOCK_DGRAM, 0) or die "socket: $!";`,
		`print "bind: ", outcome(bind($s, pack_sockaddr_in(0, INADDR_LOOPBACK))), "\n";`,
		`print "send: ", outcome(defined send($s, "x", 0, pack_sockaddr_in(9, INADDR_LOOPBACK))), "\n";`,
		`print "connect: ", outcome(connect($s, pack_sockaddr_in(9, INADDR_LOOPBACK))), "\n";`,
	)}, "")
	r.want(t, 0, "bind: EACCES\nsend: EACCES\nconnect: EACCES\n")

	// The Dial policy is asked about a datagram's destination, as "udp".
	var asked []string
	cfg := perl.Config{Dial: func(network, host, ip string, port int) bool {
		asked = append(asked, network+" "+ip)
		return true
	}}
	r = runProgram(t, cfg, perl.Program{Code: code(
		`use Socket;`,
		`socket(my $s, AF_INET, SOCK_DGRAM, 0) or die "socket: $!";`,
		`print "send: ", (defined send($s, "x", 0, pack_sockaddr_in(9, INADDR_LOOPBACK)) ? "sent" : "failed: $!"), "\n";`,
	)}, "")
	r.want(t, 0, "send: sent\n")
	if len(asked) != 1 || asked[0] != "udp 127.0.0.1" {
		t.Fatalf("the Dial policy was asked %q, want one question about udp 127.0.0.1", asked)
	}
}
