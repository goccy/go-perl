// Package gperl is the library behind the gperl command: the perl-meets-go
// toolchain for go-perl. It resolves CPAN dependencies (cpanm/carton
// conventions: cpanfile + ./local), runs Perl programs on the embedded
// interpreter, and builds self-contained Go binaries that embed a script,
// its vendored modules, and the interpreter.
//
// Everything the CLI does is callable programmatically: EnsureDeps for
// dependency vendoring, Run for execution, Build for binary production, and
// XSBuild (with the xs package's LoadDir) for native XS modules — an XS
// distribution is compiled
// once against the bundled XS SDK (its own Makefile.PL/Build.PL drives the
// build) and the resulting shared library loads at runtime with no
// compiler present.
package gperl

import (
	"errors"
	"fmt"
	"os"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// HostConfig is the perl-like base configuration every gperl entry point
// (and every binary `gperl build` produces) starts from: the operating
// system's filesystem with the embedded stdlib extracted onto it, and
// every capability hook allowing — perl(1) does not sandbox, so neither
// does gperl (the library's zero Config denies these instead).
func HostConfig() (perl.Config, error) {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		return perl.Config{}, fmt.Errorf("extract embedded stdlib: %w", err)
	}
	return perl.Config{
		FS:            goperlfs.NewHostFS(),
		StdlibDir:     stdlib,
		Dial:          func(string, string, string, int) bool { return true },
		Resolve:       func(string) bool { return true },
		Exec:          func(string, []string) bool { return true },
		Fork:          true,
		NativeModules: true,
	}, nil
}

// Run executes the Perl script on the embedded interpreter with the host's
// stdio, os.Environ(), and the project's vendored modules (./local next to
// the script) on @INC, resolving cpanfile dependencies first when needed.
//
// The returned status is the process exit status the run maps to: 0 on
// success, the code of a Perl-level exit(), and perl's own status for an
// uncaught die (255 unless errno says otherwise) — whose message perl has
// already written to stderr. err reports a failure to get the program
// running at all; it is nil whenever the program ran.
func Run(script string, args []string) (status int, err error) {
	cmd := &perlCommand{prog: perl.Program{File: script, Args: args}}
	return runPerlCommand(cmd, perlRunOpts{project: true})
}

// RunCLI executes a `gperl run` command line: the command line of the perl
// command (perlrun's switches, the program file or - for standard input,
// the program's arguments). Two modes share the entry point:
//
//   - `gperl run script.pl [args...]` (no perl switches) is the project
//     runner: cpanfile dependencies resolve into ./local, the vendored
//     tree and the script's directory join @INC, and native modules
//     from local/xs load lazily.
//   - Any perl(1) switch (-e, -I, -M, -n, ...), a program on standard
//     input — or running as a child of the XS build pipeline
//     (GPERL_PERL_PRELOAD set: the shim that make rules and $^X
//     re-invocations exec) — selects plain perl semantics: the invocation
//     behaves exactly like the perl command, with no project-level
//     dependency resolution. A scratch dist's own cpanfile must not start
//     vendoring mid-build.
func RunCLI(argv []string) (status int, err error) {
	env := os.Environ()
	cmd, err := parsePerlCommand(argv, env)
	if err != nil {
		return switchFailure(err, perlRunOpts{})
	}
	project := !cmd.switches && cmd.action == actionRun && cmd.prog.File != "" &&
		envValue(env, "GPERL_PERL_PRELOAD") == ""
	return runPerlCommand(cmd, perlRunOpts{project: project})
}

// PrintRunError writes the error of a run that could not be carried out. A
// *perl.PerlError prints its message verbatim, the way the perl command
// reports a die; anything else is a gperl failure. Shared by the CLI and
// the build pipeline's re-entry hook.
func PrintRunError(err error) {
	var pe *perl.PerlError
	if errors.As(err, &pe) {
		msg := pe.Message
		if len(msg) == 0 || msg[len(msg)-1] != '\n' {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "gperl: %v\n", err)
	}
}
