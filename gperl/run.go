// Package gperl is the library behind the gperl command: the perl-meets-go
// toolchain for go-perl. It resolves CPAN dependencies (cpanm/carton
// conventions: cpanfile + ./local), runs Perl programs on the embedded
// interpreter, and builds self-contained Go binaries that embed a script,
// its vendored modules, and the interpreter.
//
// Everything the CLI does is callable programmatically: Get/EnsureDeps for
// dependency vendoring, Run for execution, Build for binary production, and
// XSBuild/LoadXS for native XS modules — an XS distribution is compiled
// once against the bundled XS SDK (its own Makefile.PL/Build.PL drives the
// build) and the resulting shared library loads at runtime with no
// compiler present.
package gperl

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	perl "github.com/goccy/go-perl"
)

// Run executes the Perl script on the embedded interpreter with the host's
// stdio, os.Environ(), and the project's vendored modules (./local next to
// the script) on @INC, resolving cpanfile dependencies first when needed.
//
// The returned status is the process exit status the run maps to: 0 on
// success, the code of a Perl-level exit(), 255 for an uncaught die (perl's
// own convention). err carries the die (*perl.PerlError) or a host failure;
// it is nil for success and for a plain exit().
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

	p, err := perl.New(perl.Config{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env:    os.Environ(),
	})
	if err != nil {
		return 1, err
	}
	// Native XS modules built by `gperl xs build` register lazily; a
	// stock `use Module;` boots them through the XSLoader contract.
	if err := LoadXS(p, xsDir(projectDir)); err != nil {
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
