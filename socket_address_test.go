package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// Socket converts addresses with inet_pton and inet_ntop and packs the IPv6
// socket address and the multicast requests: what comes out of a packer
// goes back in through its unpacker unchanged.
func TestSocketAddressConversionAndPacking(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Socket qw(:DEFAULT inet_pton inet_ntop pack_sockaddr_in6 unpack_sockaddr_in6 sockaddr_family`,
		`              pack_ip_mreq unpack_ip_mreq pack_ip_mreq_source unpack_ip_mreq_source`,
		`              pack_ipv6_mreq unpack_ipv6_mreq);`,
		`print inet_ntop(AF_INET, inet_pton(AF_INET, "192.0.2.7")), "\n";`,
		`my $v6 = inet_pton(AF_INET6, "2001:db8::8a2e:370:7334");`,
		`print length($v6), " ", inet_ntop(AF_INET6, $v6), "\n";`,
		`print defined(inet_pton(AF_INET6, "not an address")) ? "accepted" : "refused", "\n";`,
		`my $sa = pack_sockaddr_in6(8080, $v6, 3, 5);`,
		`my ($port, $addr, $scope, $flow) = unpack_sockaddr_in6($sa);`,
		`print join(" ", (sockaddr_family($sa) == AF_INET6 ? "inet6" : "other"), $port, inet_ntop(AF_INET6, $addr), $scope, $flow), "\n";`,
		`print join(" ", map { inet_ntoa($_) } unpack_ip_mreq(pack_ip_mreq(inet_aton("224.0.0.251"), inet_aton("10.0.0.1")))), "\n";`,
		`print join(" ", map { inet_ntoa($_) } unpack_ip_mreq_source(pack_ip_mreq_source(inet_aton("232.1.1.1"), inet_aton("192.0.2.9"), inet_aton("10.0.0.1")))), "\n";`,
		`my ($group, $ifindex) = unpack_ipv6_mreq(pack_ipv6_mreq(inet_pton(AF_INET6, "ff02::fb"), 2));`,
		`print inet_ntop(AF_INET6, $group), " $ifindex\n";`,
		`print eval { require IO::Socket::IP; 1 } ? "IO::Socket::IP loads\n" : "IO::Socket::IP: $@";`,
	)}, "")
	r.want(t, 0, "192.0.2.7\n"+
		"16 2001:db8::8a2e:370:7334\n"+
		"refused\n"+
		"inet6 8080 2001:db8::8a2e:370:7334 3 5\n"+
		"224.0.0.251 10.0.0.1\n"+
		"232.1.1.1 192.0.2.9 10.0.0.1\n"+
		"ff02::fb 2\n"+
		"IO::Socket::IP loads\n")
}
