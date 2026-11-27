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
	"sort"
	"strings"
	"syscall"

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
	// cpanfile dependencies are synced first, and they and the program's
	// directory join @INC.
	project bool
	// isolated runs the program as a process of its own inside this one
	// (perl.Config.HostProcess unset): its working directory and
	// environment are the instance's, so several such programs run at once.
	isolated bool
}

// startupGlue prepares the interpreter for a run the way the process
// around a perl command would have: its working directory, the path it
// knows itself by ($^X), the host's C toolchain in %Config (runToolchain)
// and the build pipeline's %Config overlay.
const startupGlue = `sub {
    my ($dir, $exe, $preload, %toolchain) = @_;
    # the directory the process was started in, which it may not search
    Internals::start_directory($dir) or die "chdir $dir: $!\n";
    $^X = $exe if defined $exe;
    # Config.pm reports %Config::Override in place of its own values once it
    # is loaded; nothing is loaded here
    @Config::Override{keys %toolchain} = values %toolchain;
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
//     into the embedded interpreter); without it $^X is the executable's
//     own name as the perl command (selfperl.go);
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
//
// The interpreter is the host process while it runs: the process follows
// its working directory and its environment, so that native code it loads
// finds them. A run that is given a directory or an environment of its own
// is not the process, which is only lent to it: the process gets its own
// working directory and environment back when the run is over.
func runPerlCommand(cmd *perlCommand, opts perlRunOpts) (status int, err error) {
	if !opts.isolated && (opts.dir != "" || opts.env != nil) {
		giveBack, err := lendProcess(opts.env)
		if err != nil {
			return 1, err
		}
		defer giveBack()
	}
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

	if opts.project {
		script, err := filepath.Abs(prog.File)
		if err != nil {
			return 1, err
		}
		if _, err := os.Stat(script); err != nil {
			return 1, err
		}
		projectDir := filepath.Dir(script)
		deps, err := hostDeps(projectDir)
		if err != nil {
			return 1, err
		}
		prog.Inc = prog.Inc[:len(prog.Inc):len(prog.Inc)]
		if lib := filepath.Join(projectDir, "lib"); dirExists(lib) {
			prog.Inc = append(prog.Inc, lib)
		}
		prog.Inc = append(append(prog.Inc, projectDir), deps.inc()...)
		prog.File = script
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
	if opts.isolated {
		cfg.HostProcess, cfg.InheritDescriptors = false, false
	}
	p, err := perl.New(cfg)
	if err != nil {
		return 1, err
	}
	// The program's run includes the tail of a perl process — the flush of
	// its buffered output and global destruction — so its status is final;
	// Close gives back what the host holds for the instance.
	defer p.Close()

	ctx := context.Background()
	// A new instance starts in the root directory; a process starts where
	// it was launched.
	dir := opts.dir
	if dir == "" {
		if dir, err = os.Getwd(); err != nil {
			// os.Getwd looks at "." when PWD is set, which a directory
			// the user may not search refuses; getcwd(2) names it (on
			// Linux), and perl runs there.
			wd, werr := syscall.Getwd()
			if werr != nil {
				return 1, fmt.Errorf("the working directory: %w", err)
			}
			dir = wd
		}
	}
	// The path the program can run its own perl by: the one the
	// environment names (the build pipeline's), or this executable's name
	// as the perl command.
	exe := envValue(env, "GPERL_PERL_EXE")
	if exe == "" {
		exe = selfPerl()
	}
	if err := startInterpreter(ctx, p, dir, exe, envValue(env, "GPERL_PERL_PRELOAD")); err != nil {
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
	args := []perl.Value{perl.ValueOf(dir), optional(exe), optional(preload)}
	tc := runToolchain()
	keys := make([]string, 0, len(tc))
	for k := range tc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		args = append(args, perl.ValueOf(k), perl.ValueOf(tc[k]))
	}
	_, err = start.CallScalar(ctx, args...)
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

// lendProcess gives the process the environment env (when it is not nil) and
// returns the function that restores the working directory and the
// environment the process has now.
func lendProcess(env []string) (giveBack func(), err error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	own := os.Environ()
	// An environment is a list of name=value entries: the form the
	// operating system defines for it.
	setEnviron := func(vars []string) {
		os.Clearenv()
		for _, kv := range vars {
			if name, value, ok := strings.Cut(kv, "="); ok {
				os.Setenv(name, value)
			}
		}
	}
	if env != nil {
		setEnviron(env)
	}
	return func() {
		setEnviron(own)
		os.Chdir(dir)
	}, nil
}
