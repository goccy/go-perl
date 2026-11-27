package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// Under -T everything that comes from outside the program is tainted:
// its arguments, its name, the interpreter's name and the environment.
func TestRunProgramTaintsOutsideData(t *testing.T) {
	cfg := perl.Config{Env: []string{"PATH=/bin"}}
	r := runProgram(t, cfg, perl.Program{
		Taint: perl.TaintOn,
		Args:  []string{"arg"},
		Code: code(
			`use Scalar::Util "tainted";`,
			`print join(" ", map { tainted($_) ? 1 : 0 } $ARGV[0], $0, $^X, $ENV{PATH}), "\n";`,
			`print tainted("literal") ? 1 : 0, "\n";`,
		),
	}, "")
	r.want(t, 0, "1 1 1 1\n0\n")
}
