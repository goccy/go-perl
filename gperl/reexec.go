package gperl

// Executable re-entry for the build pipeline's child processes.
//
// The pipeline itself never spawns a perl process — Makefile.PL,
// Build.PL, and cpanm run on in-process interpreters. But the BUILD
// spawns children of its own: make rules exec $(PERL), $^X
// re-invocations exec the interpreter path, cpanm's workers exec perl.
// Those land on the perl shim, which execs the CURRENT executable as
// `<exe> run <perl args>` — the public `gperl run` surface, whose plain
// mode is perl(1)-compatible.
//
// The current executable is not always the gperl CLI: a test binary or
// an application embedding this package can host an XS build too. This
// init serves the re-entry before the host program's own main (or the
// test runner) sees the process. The trigger is deliberately narrow:
// BOTH the environment marker (set only by the shim) AND the `run` argv
// must be present, so a child that merely inherits the environment is
// left alone. The marker is dropped before the interpreter starts, so
// processes IT spawns see a clean environment (their shims re-add it
// for their own exec).

import (
	"os"
)

// perlCLIEnv marks a process as a shim re-entry. Exported nowhere; the
// shim scripts set it.
const perlCLIEnv = "GPERL_INTERNAL_PERL_CLI"

func init() {
	if os.Getenv(perlCLIEnv) == "" || len(os.Args) < 2 || os.Args[1] != "run" {
		return
	}
	os.Unsetenv(perlCLIEnv)
	status, err := RunCLI(os.Args[2:])
	if err != nil {
		PrintRunError(err)
	}
	os.Exit(status)
}
