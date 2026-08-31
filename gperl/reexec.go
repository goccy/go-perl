package gperl

// Self re-execution for the build pipeline.
//
// The XS build drives Makefile.PL/Build.PL/cpanm on the embedded
// interpreter by re-entering the CURRENT executable (the perl shim scripts
// exec it with the internal "__perl" argument). That executable is not
// always the gperl CLI: any binary that links this package — a test
// binary, an application embedding the library — can end up as the shim
// target. This init intercepts those re-entries before the host program's
// own main (or the test runner) sees the process.
//
// The trigger is deliberately narrow: BOTH the environment marker (set
// only by the shim and the cpanm launcher) AND the "__perl" argv must be
// present, so a child that merely inherits the environment and runs the
// real CLI normally is left alone. The marker is dropped from the
// environment before the interpreter starts, so processes IT spawns see a
// clean slate (their shims re-add it for their own exec).

import (
	"fmt"
	"os"
)

// perlCLIEnv marks a process as an internal perl-CLI re-entry. Exported
// nowhere; the shim scripts and cpanmCmd set it.
const perlCLIEnv = "GPERL_INTERNAL_PERL_CLI"

func init() {
	if os.Getenv(perlCLIEnv) == "" || len(os.Args) < 2 || os.Args[1] != "__perl" {
		return
	}
	os.Unsetenv(perlCLIEnv)
	status, err := RunPerlCLI(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "gperl: %v\n", err)
	}
	os.Exit(status)
}
