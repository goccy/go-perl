package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// Global destruction is part of the program's run: what a destructor does
// to the status there decides it, as it does for the perl command.
func TestRunProgramStatusFromGlobalDestruction(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`sub Guard::DESTROY { print "destroyed in ${^GLOBAL_PHASE}\n"; $? = 5 }`,
		`our $guard = bless [], 'Guard'; print "ran\n";`,
	)}, "")
	r.want(t, 5, "ran\ndestroyed in DESTRUCT\n")

	r = runProgram(t, perl.Config{}, perl.Program{Code: code(
		`sub Guard::DESTROY { $? = 5 }`,
		`our $guard = bless [], 'Guard'; exit 3;`,
	)}, "")
	r.want(t, 5, "")

	r = runProgram(t, perl.Config{}, perl.Program{Code: code(
		`sub Guard::DESTROY { exit 7 }`,
		`our $guard = bless [], 'Guard'; print "ran\n";`,
	)}, "")
	r.want(t, 7, "ran\n")

	r = runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use POSIX ();`,
		`sub Guard::DESTROY { POSIX::_exit(219) }`,
		`our $guard = bless [], 'Guard'; print "ran\n";`,
	)}, "")
	r.want(t, 219, "ran\n")

	// a destructor that dies there changes nothing
	r = runProgram(t, perl.Config{}, perl.Program{Code: code(
		`sub Guard::DESTROY { die "ignored\n" }`,
		`our $guard = bless [], 'Guard'; exit 3;`,
	)}, "")
	r.want(t, 3, "")
}
