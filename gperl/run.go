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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
	"github.com/goccy/go-perl/xs"
)

// Run executes the Perl script on the embedded interpreter with the host's
// stdio, os.Environ(), and the project's vendored modules (./local next to
// the script) on @INC, resolving cpanfile dependencies first when needed.
//
// The returned status is the process exit status the run maps to: 0 on
// success, the code of a Perl-level exit(), 255 for an uncaught die (perl's
// own convention). err carries the die (*perl.PerlError) or a host failure;
// it is nil for success and for a plain exit().
// HostConfig is the perl-like base configuration every gperl entry point
// (and every binary `gperl build` produces) starts from: the operating
// system's filesystem, with the embedded stdlib extracted onto it.
func HostConfig() (perl.Config, error) {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		return perl.Config{}, fmt.Errorf("extract embedded stdlib: %w", err)
	}
	return perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib}, nil
}

func Run(script string, args []string) (status int, err error) {
	script, err = filepath.Abs(script)
	if err != nil {
		return 1, err
	}
	if _, err := os.Stat(script); err != nil {
		return 1, err
	}
	projectDir := filepath.Dir(script)
	if err := EnsureDeps(projectDir); err != nil {
		return 1, err
	}
	var inc []string
	if lib := localLib(projectDir); lib != "" {
		inc = append(inc, lib)
	}
	inc = append(inc, projectDir)

	cfg, err := HostConfig()
	if err != nil {
		return 1, err
	}
	// gperl behaves like the perl command: the script sees the host
	// filesystem, stdio, and environment.
	cfg.Stdin = os.Stdin
	cfg.Stdout = os.Stdout
	cfg.Stderr = os.Stderr
	cfg.Env = os.Environ()
	p, err := perl.New(cfg)
	if err != nil {
		return 1, err
	}
	// Native XS modules built by `gperl xs build` register lazily; a
	// stock `use Module;` boots them through the XSLoader contract.
	if err := xs.LoadDir(p, xsDir(projectDir)); err != nil {
		p.Close()
		return 1, err
	}
	runErr := p.RunFile(context.Background(), script, inc, args)
	// Close before reporting: a guest exit() unwound cleanly inside the
	// bridge, so destruction flushes PerlIO (the script's buffered output)
	// and runs END blocks, exactly like the tail of a perl process.
	p.Close()
	return statusFromRun(runErr)
}

// RunCLI executes a `gperl run` command line. Two modes share the
// entry point:
//
//   - `gperl run script.pl [args...]` (no perl switches) is the project
//     runner: cpanfile dependencies resolve into ./local, the vendored
//     tree and the script's directory join @INC, and native modules
//     from local/xs load lazily.
//   - Any perl(1) switch (-e, -I, -M, -l, ...) — or running as a child
//     of the XS build pipeline (GPERL_PERL_PRELOAD set: the shim that
//     make rules and $^X re-invocations exec) — selects plain perl
//     semantics: the invocation behaves exactly like the perl command,
//     with no project-level dependency resolution. A scratch dist's own
//     cpanfile must not start vendoring mid-build.
func RunCLI(argv []string) (status int, err error) {
	pa, perr := parsePerlArgv(argv)
	if perr != nil {
		return 2, perr
	}
	if pa.sawFlag || envValue(os.Environ(), "GPERL_PERL_PRELOAD") != "" {
		return runPerlCLI(argv, perlRunOpts{})
	}
	if pa.script == "" {
		return 2, fmt.Errorf("no script given")
	}
	return Run(pa.script, pa.args)
}

// PrintRunError writes err the way the perl command would report it: an
// uncaught die prints its message verbatim; anything else is a gperl
// failure. Shared by the CLI and the build pipeline's re-entry hook.
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

// statusFromRun maps a RunFile outcome onto the perl process conventions.
func statusFromRun(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	if code, ok := perl.ExitCode(err); ok {
		return code, nil
	}
	var pe *perl.PerlError
	if errors.As(err, &pe) {
		return 255, err
	}
	return 1, err
}
