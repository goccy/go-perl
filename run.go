package perl

// Whole-script execution, the `perl script.pl` shape. Used by cmd/gperl and
// by the binaries `gperl build` generates.

import (
	"context"
	"fmt"
)

// runFileGlue is the in-guest launcher. It is defined via Eval (whose stdio
// capture sees nothing — the definition prints nothing) and then driven via
// Call, which does NOT redirect stdio, so the script's output streams to the
// instance's configured stdout/stderr as it happens. @INC additions and
// @ARGV cross over the bridge as values, so no Perl-source quoting is
// involved.
const runFileGlue = `
sub main::__goperl_run_file {
    my ($path, $inc, @args) = @_;
    unshift @INC, @$inc;
    local @ARGV = @args;
    local $0 = $path;
    my $ok = do $path;
    if (!defined $ok) {
        die $@ if $@;
        die "$path: $!\n";
    }
    return 1;
}
1;
`

// RunFile executes the Perl program at path — a guest path against
// Config.FS when one is set, a host path otherwise — with incDirs prepended
// to @INC and args as @ARGV. The program's output goes to the instance's
// configured stdio (not captured), and its package state persists on the
// instance afterwards.
//
// An uncaught die comes back as *PerlError; a Perl-level exit() comes back
// as the error ExitCode recognizes, after which the instance must only be
// Closed. Both conventions let a command-line front end translate the
// outcome into a process exit status.
func (p *Perl) RunFile(ctx context.Context, path string, incDirs []string, args []string) error {
	r, err := p.Eval(ctx, runFileGlue)
	if err != nil {
		return err
	}
	if !r.Ok {
		return fmt.Errorf("install script runner: %s", r.Error)
	}
	callArgs := make([]any, 0, len(args)+2)
	inc := make([]any, len(incDirs))
	for i, d := range incDirs {
		inc[i] = d
	}
	callArgs = append(callArgs, path, inc)
	for _, a := range args {
		callArgs = append(callArgs, a)
	}
	_, err = p.Call(ctx, "__goperl_run_file", callArgs...)
	return err
}
