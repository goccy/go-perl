package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// A program may assign to $0. perl rewrites the argument area the
// interpreter was started with when it does, so that area has to be the
// interpreter's for good: the program goes on and ends normally, whatever
// its switches were.
func TestProgramAssignsItsName(t *testing.T) {
	for _, inc := range [][]string{nil, {"blib/lib", "blib/arch"}, {"a", "b", "c", "d"}} {
		r := runProgram(t, perl.Config{}, perl.Program{
			Inc: inc,
			Code: code(
				`$0 = "$0";`,
				`$0 = "a much longer name than the interpreter was started with";`,
				`print "name: $0\n";`,
				`$0 = "x";`,
				`my @fill = map { "v$_" x 20 } 1 .. 200;`,
				`print "end ", scalar(@fill), "\n";`,
			),
		}, "")
		r.want(t, 0, "name: a much longer name than the interpreter was started with\nend 200\n")
	}
}
