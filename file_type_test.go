package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// In the sandbox too, /dev/null is a character device, and the standard
// streams the embedder supplies are pipes to the program.
func TestSandboxFileTypes(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`sub types { stat($_[0]) or return "missing"; join("+", grep { defined } (-f _ ? "file" : undef), (-d _ ? "directory" : undef), (-c _ ? "character device" : undef), (-b _ ? "block device" : undef), (-p _ ? "pipe" : undef), (-S _ ? "socket" : undef)) || "no type" }`,
		`print "/dev/null: ", types("/dev/null"), "\n";`,
		`open(my $null, "<", "/dev/null") or die "open: $!";`,
		`print "a handle on it: ", types($null), "\n";`,
		`print "/dev: ", types("/dev"), "\n";`,
		`print "stdin, stdout: ", types(\*STDIN), ", ", types(\*STDOUT), "\n";`,
	)}, "")
	r.want(t, 0, "/dev/null: character device\na handle on it: character device\n/dev: directory\nstdin, stdout: pipe, pipe\n")
}
