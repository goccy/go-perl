//go:build darwin || linux

package gperl_test

import "testing"

// $! and POSIX::strerror read as perl on Linux words each error (glibc's
// texts), the errors the interpreter's C library has a number for and those
// only the host has (EHOSTDOWN, ESHUTDOWN) alike, under the host's numbers.
func TestErrorTextsOfTheHostProcess(t *testing.T) {
	out, status := hostPerl(t, "-MErrno", "-MPOSIX", "-e", `
for my $name (qw(EIO EMSGSIZE ENOTSUP EADDRINUSE ECONNRESET EHOSTDOWN ESHUTDOWN ETOOMANYREFS)) {
    my $number = eval "Errno::$name()";
    $! = $number;
    print "$name: $! / ", POSIX::strerror($number), "\n";
}
`)
	want := "EIO: Input/output error / Input/output error\n" +
		"EMSGSIZE: Message too long / Message too long\n" +
		"ENOTSUP: Operation not supported / Operation not supported\n" +
		"EADDRINUSE: Address already in use / Address already in use\n" +
		"ECONNRESET: Connection reset by peer / Connection reset by peer\n" +
		"EHOSTDOWN: Host is down / Host is down\n" +
		"ESHUTDOWN: Cannot send after transport endpoint shutdown / Cannot send after transport endpoint shutdown\n" +
		"ETOOMANYREFS: Too many references: cannot splice / Too many references: cannot splice\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
