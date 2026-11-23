package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// $! reads as perl on Linux words each error (glibc's texts), whatever the
// interpreter's C library calls it: a program, or its test, compares the
// text with perl's ("Input/output error", not "I/O error"). A number that is
// no error is an unknown one, by its number.
func TestErrorTextsAreThoseOfPerlOnLinux(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Errno;`,
		`for my $name (qw(EPERM ENOENT EIO EBUSY ECHILD EMSGSIZE ENOTSUP EOPNOTSUPP EADDRINUSE EADDRNOTAVAIL ECONNABORTED ENOTCONN ENOTSOCK ENOTTY ESPIPE ETIMEDOUT EXDEV EILSEQ)) {`,
		`    $! = eval "Errno::$name()"; print "$name: $!\n";`,
		`}`,
		`$! = 9999; print "9999: $!\n";`,
	)}, "")
	r.want(t, 0, "EPERM: Operation not permitted\n"+
		"ENOENT: No such file or directory\n"+
		"EIO: Input/output error\n"+
		"EBUSY: Device or resource busy\n"+
		"ECHILD: No child processes\n"+
		"EMSGSIZE: Message too long\n"+
		"ENOTSUP: Operation not supported\n"+
		"EOPNOTSUPP: Operation not supported\n"+
		"EADDRINUSE: Address already in use\n"+
		"EADDRNOTAVAIL: Cannot assign requested address\n"+
		"ECONNABORTED: Software caused connection abort\n"+
		"ENOTCONN: Transport endpoint is not connected\n"+
		"ENOTSOCK: Socket operation on non-socket\n"+
		"ENOTTY: Inappropriate ioctl for device\n"+
		"ESPIPE: Illegal seek\n"+
		"ETIMEDOUT: Connection timed out\n"+
		"EXDEV: Invalid cross-device link\n"+
		"EILSEQ: Invalid or incomplete multibyte or wide character\n"+
		"9999: Unknown error 9999\n")
}
