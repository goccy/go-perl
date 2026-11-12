package gperl

// The perl command over the embedded interpreter.
//
// The XS build pipeline is driven entirely by Perl programs — Makefile.PL,
// Build.PL, xsubpp, ExtUtils::Command one-liners from the generated
// Makefile, cpanm — and all of them are pure Perl. The pipeline runs them
// on IN-PROCESS interpreters, so building native modules needs no perl
// installed on the system and spawns no perl processes of its own: gperl
// IS the perl that drives the build. The only external it execs is make
// (cc runs under it); the perl children that make itself (or guest code)
// spawns come back as `gperl run` through the shim.
//
// perlswitch.go reads the command line; this file runs what it describes.

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	perl "github.com/goccy/go-perl"
)

// perlRunOpts parameterizes one perl(1)-style run. Zero values mean the
// current process: its working directory, environment, and stdio.
type perlRunOpts struct {
	dir    string    // the program's working directory
	env    []string  // the instance environment
	stdin  io.Reader // program stdin
	stdout io.Writer // program stdout
	stderr io.Writer // program stderr
	// project runs the program file as a project's entry point: its
	// cpanfile dependencies resolve into ./local first, the vendored tree
	// and the program's directory join @INC, and native modules from
	// local/xs load lazily.
	project bool
}

// startupGlue prepares the interpreter for a run the way the process
// around a perl command would have: its working directory, the path it
// knows itself by ($^X), and the build pipeline's %Config overlay.
const startupGlue = `sub {
    my ($dir, $exe, $preload) = @_;
    chdir $dir or die "chdir $dir: $!\n";
    $^X = $exe if defined $exe;
    if (defined $preload) {
        # do FILE returns undef for an unreadable file and for a file whose
        # last statement is undef alike; establish readability first.
        open(my $fh, '<', $preload) or die "preload $preload: $!\n";
        close $fh;
        do $preload;
        die $@ if $@;
        # The preload belongs to how the interpreter was started, not to
        # the program: a program that lists what it loaded (a dependency
        # scanner reading %INC) must not find it.
        delete $INC{$preload};
    }
    return;
}`

// runPerlCLI executes a perl(1)-style command line on an in-process
// interpreter and returns the would-be process exit status. The build
// pipeline uses it two ways: directly for its own Makefile.PL/Build.PL/
// cpanm runs (no subprocess involved), and through the executable
// re-entry hook (reexec.go) for the children the BUILD spawns — make
// rules and $^X re-invocations exec the shim, which is a new OS process
// by nature. Recognized environment (from opts.env):
//
//   - GPERL_PERL_EXE: the value for $^X (the perl shim path, so build
//     systems that re-invoke $^X — Module::Build's ./Build — come back
//     into the embedded interpreter);
//   - GPERL_PERL_PRELOAD: a Perl file `do`ne before anything else (the XS
//     build's %Config toolchain overlay).
func runPerlCLI(argv []string, opts perlRunOpts) (status int, err error) {
	env := opts.env
	if env == nil {
		env = os.Environ()
	}
	cmd, err := parsePerlCommand(argv, env)
	if err != nil {
		return switchFailure(err, opts)
	}
	return runPerlCommand(cmd, opts)
}

// switchFailure reports a command line perl would have refused, the way
// perl does: the message on stderr, a failure status.
func switchFailure(err error, opts perlRunOpts) (int, error) {
	stderr := opts.stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	msg := err.Error()
	if !strings.HasSuffix(msg, ".") {
		msg += "."
	}
	fmt.Fprintln(stderr, msg)
	return 255, nil
}

// runPerlCommand runs a parsed command line.
func runPerlCommand(cmd *perlCommand, opts perlRunOpts) (status int, err error) {
	env := opts.env
	if env == nil {
		env = os.Environ()
	}
	stdout, stderr := opts.stdout, opts.stderr
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	prog := cmd.prog
	switch cmd.action {
	case actionUsage:
		name := envValue(env, "GPERL_PERL_EXE")
		if name == "" {
			name = "gperl run"
		}
		fmt.Fprintf(stdout, perlUsage, name)
		return 0, nil
	case actionVersion:
		prog = perl.Program{Code: perlVersionProgram}
	}
	for _, note := range cmd.notes {
		fmt.Fprintln(stderr, note)
	}
	env = append(env[:len(env):len(env)], cmd.env...)

	nativeDir := ""
	if opts.project {
		script, err := filepath.Abs(prog.File)
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
		if lib := localLib(projectDir); lib != "" {
			prog.Inc = append(prog.Inc[:len(prog.Inc):len(prog.Inc)], lib)
		}
		prog.Inc = append(prog.Inc[:len(prog.Inc):len(prog.Inc)], projectDir)
		prog.File = script
		nativeDir = xsDir(projectDir)
	}

	// gperl behaves like the perl command: the program sees the host
	// filesystem, stdio, and environment.
	cfg, err := HostConfig()
	if err != nil {
		return 1, err
	}
	cfg.Stdin = opts.stdin
	if cfg.Stdin == nil {
		cfg.Stdin = os.Stdin
	}
	cfg.Stdout = stdout
	cfg.Stderr = stderr
	cfg.Env = env
	p, err := perl.New(cfg)
	if err != nil {
		return 1, err
	}
	// The program's run includes the tail of a perl process — the flush of
	// its buffered output and global destruction — so its status is final;
	// Close gives back what the host holds for the instance.
	defer p.Close()

	ctx := context.Background()
	if nativeDir != "" {
		// Native XS modules built by `gperl xs build` register lazily; a
		// stock `use Module;` boots them through the XSLoader contract.
		if err := p.AddXSDir(nativeDir); err != nil {
			return 1, err
		}
	}
	// A new instance starts in the root directory; a process starts where
	// it was launched.
	dir := opts.dir
	if dir == "" {
		if dir, err = os.Getwd(); err != nil {
			return 1, err
		}
	}
	if err := startInterpreter(ctx, p, dir, envValue(env, "GPERL_PERL_EXE"), envValue(env, "GPERL_PERL_PRELOAD")); err != nil {
		return 1, err
	}
	return p.RunProgram(ctx, prog)
}

// startInterpreter applies startupGlue; an empty exe or preload leaves that
// part of the interpreter as it is.
func startInterpreter(ctx context.Context, p *perl.Perl, dir, exe, preload string) error {
	r, err := p.Eval(ctx, startupGlue)
	if err != nil {
		return err
	}
	if r.Error != nil {
		return fmt.Errorf("install the startup routine: %w", r.Error)
	}
	ref, err := perl.As[perl.RefValue](r.Value)
	if err != nil {
		return fmt.Errorf("install the startup routine: %w", err)
	}
	routine, err := ref.Deref(ctx)
	if err != nil {
		return fmt.Errorf("install the startup routine: %w", err)
	}
	start, err := perl.As[perl.CodeValue](routine)
	if err != nil {
		return fmt.Errorf("install the startup routine: %w", err)
	}
	optional := func(s string) perl.Value {
		if s == "" {
			return perl.Undef()
		}
		return perl.ValueOf(s)
	}
	_, err = start.Call(ctx, perl.ValueOf(dir), optional(exe), optional(preload))
	return err
}

// envValue reads key from an environ-shaped slice (last assignment wins,
// like the environment itself).
func envValue(env []string, key string) string {
	v := ""
	for _, kv := range env {
		if strings.HasPrefix(kv, key+"=") {
			v = kv[len(key)+1:]
		}
	}
	return v
}

// runPerlInProcess runs one perl(1)-style invocation of the build
// pipeline on an in-process interpreter, treating a nonzero exit as an
// error the way a subprocess wrapper would.
func runPerlInProcess(dir string, env []string, stdout io.Writer, args ...string) error {
	status, err := runPerlCLI(args, perlRunOpts{dir: dir, env: env, stdout: stdout})
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("perl %s: exit status %d", strings.Join(args, " "), status)
	}
	return nil
}
