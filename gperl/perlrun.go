package gperl

// A perl-compatible command line over the embedded interpreter.
//
// The XS build pipeline is driven entirely by Perl programs — Makefile.PL,
// Build.PL, xsubpp, ExtUtils::Command one-liners from the generated
// Makefile, cpanm — and all of them are pure Perl. The pipeline runs them
// on IN-PROCESS interpreters, so building native modules needs no perl
// installed on the system and spawns no perl processes of its own: gperl
// IS the perl that drives the build. The only externals it execs are make
// and curl (cc runs under make); the perl children that make itself (or
// guest code) spawns come back as `gperl run` through the shim.
//
// The flag surface is the subset those build tools use: -e, -I, -M/-m, -w,
// `--`, and script execution. Anything else is rejected loudly rather than
// half-interpreted.

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
	stdout io.Writer // program stdout
	stderr io.Writer // program stderr
}

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
	pa, perr := parsePerlArgv(argv)
	if perr != nil {
		return 2, perr
	}
	inc, preUse, eChunks := pa.inc, pa.preUse, pa.eChunks
	script, args, autoLine := pa.script, pa.args, pa.autoLine
	if script == "" && len(eChunks) == 0 {
		return 2, fmt.Errorf("no -e code and no script given")
	}

	// The launcher runs as a synthesized program file so the ordinary
	// run-file conventions (die, exit, @ARGV, $0) apply unchanged.
	dir := opts.dir
	if dir == "" {
		if dir, err = os.Getwd(); err != nil {
			return 1, err
		}
	}
	env := opts.env
	if env == nil {
		env = os.Environ()
	}
	var prog strings.Builder
	// perl resolves '.'-relative requires and files against ITS cwd; make
	// the guest's match the run's working directory.
	prog.WriteString("BEGIN { chdir '" + perlQuote(dir) + "' or die \"chdir: $!\\n\"; }\n")
	if pre := envValue(env, "GPERL_PERL_PRELOAD"); pre != "" {
		fmt.Fprintf(&prog, "BEGIN { my $p = $ENV{GPERL_PERL_PRELOAD}; my $r = do $p; die $@ if $@; die \"preload $p: $!\\n\" if !defined $r && $!; }\n")
	}
	if exe := envValue(env, "GPERL_PERL_EXE"); exe != "" {
		prog.WriteString("BEGIN { $^X = $ENV{GPERL_PERL_EXE}; }\n")
	}
	if autoLine {
		prog.WriteString("$\\ = \"\\n\";\n")
	}
	for _, u := range preUse {
		prog.WriteString(u + "\n")
	}
	switch {
	case len(eChunks) > 0:
		prog.WriteString("#line 1 \"-e\"\n")
		prog.WriteString(strings.Join(eChunks, "\n"))
		prog.WriteString("\n;1;\n")
	default:
		// Concatenate the script body after the preamble so it runs as
		// the MAIN program: caller() is empty at its top level (modulino
		// programs like cpanm guard their entry point with
		// `unless (caller)`), and __END__/__DATA__ stay attached.
		abs := script
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(dir, script)
		}
		body, rerr := os.ReadFile(abs)
		if rerr != nil {
			return 1, rerr
		}
		// BEGIN: the script body's own compile-time code (a
		// `use inc::Module::Install` reads $0 through FindBin) must
		// already see the script's identity.
		prog.WriteString("BEGIN { $0 = '" + perlQuote(abs) + "'; }\n")
		prog.WriteString("#line 1 \"" + strings.ReplaceAll(abs, `"`, ``) + "\"\n")
		prog.Write(body)
		prog.WriteString("\n")
	}
	tmp, err := os.CreateTemp("", "gperl-perl-*.pl")
	if err != nil {
		return 1, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.WriteString(prog.String()); err != nil {
		tmp.Close()
		return 1, err
	}
	tmp.Close()

	cfg, err := HostConfig()
	if err != nil {
		return 1, err
	}
	cfg.Stdin = os.Stdin
	cfg.Stdout = opts.stdout
	if cfg.Stdout == nil {
		cfg.Stdout = os.Stdout
	}
	cfg.Stderr = opts.stderr
	if cfg.Stderr == nil {
		cfg.Stderr = os.Stderr
	}
	cfg.Env = env
	p, err := perl.New(cfg)
	if err != nil {
		return 1, err
	}
	runErr := p.RunFile(context.Background(), tmpPath, inc, args)
	p.Close()
	return statusFromRun(runErr)
}

// perlArgs is one parsed perl(1)-style command line.
type perlArgs struct {
	inc      []string // -I
	preUse   []string // -M/-m, as ready `use` statements
	eChunks  []string // -e, in order
	script   string   // the script path ("" with -e)
	args     []string // @ARGV
	autoLine bool     // -l (output half)
	sawFlag  bool     // any perl switch appeared (mode selection)
}

// parsePerlArgv parses argv the way perl(1) reads its command line for
// the flag subset the CPAN build toolchain uses: -e, -I, -M/-m, -l, the
// warning/taint toggles, clustered single-letter groups (-wle), `--`,
// and script execution. Anything else is rejected loudly rather than
// half-interpreted.
func parsePerlArgv(argv []string) (perlArgs, error) {
	var pa perlArgs
	i := 0
	for ; i < len(argv); i++ {
		a := argv[i]
		if a == "--" {
			pa.sawFlag = true
			i++
			goto rest
		}
		if len(a) < 2 || a[0] != '-' {
			// first non-option: the script
			if len(pa.eChunks) == 0 {
				pa.script = a
				i++
			}
			goto rest
		}
		pa.sawFlag = true
		// A perl option group: single-letter switches cluster (-wle);
		// -e/-I/-M/-m end the cluster and take the rest (or the next
		// argument) as their value.
		for j := 1; j < len(a); j++ {
			switch a[j] {
			case 'e':
				v := a[j+1:]
				if v == "" {
					i++
					if i >= len(argv) {
						return pa, fmt.Errorf("-e requires an argument")
					}
					v = argv[i]
				}
				pa.eChunks = append(pa.eChunks, v)
				j = len(a)
			case 'I':
				v := a[j+1:]
				if v == "" {
					i++
					if i >= len(argv) {
						return pa, fmt.Errorf("-I requires an argument")
					}
					v = argv[i]
				}
				pa.inc = append(pa.inc, v)
				j = len(a)
			case 'M', 'm':
				v := a[j+1:]
				if v == "" {
					i++
					if i >= len(argv) {
						return pa, fmt.Errorf("-%c requires an argument", a[j])
					}
					v = argv[i]
				}
				pa.preUse = append(pa.preUse, useStatement("-"+string(a[j])+v))
				j = len(a)
			case 'l':
				// output half of -l (input auto-chomp belongs to the
				// unsupported -n/-p loops)
				pa.autoLine = true
			case 'w', 'W', 'X', 'T', 't':
				// warning/taint toggles: accepted for compatibility; the
				// embedded build has no taint support and warnings stay
				// as the program sets them
			default:
				return pa, fmt.Errorf("unsupported perl option %q", a)
			}
		}
	}
rest:
	if i < len(argv) {
		pa.args = append(pa.args, argv[i:]...)
	}
	return pa, nil
}

// useStatement turns a -M/-m option into the `use` line perl would run:
// -MFoo → use Foo; -MFoo=a,b → use Foo split-list; -mFoo → use Foo ().
func useStatement(opt string) string {
	kind, rest := opt[1], opt[2:]
	name, arglist, hasArgs := strings.Cut(rest, "=")
	switch {
	case hasArgs:
		parts := strings.Split(arglist, ",")
		q := make([]string, len(parts))
		for i, s := range parts {
			q[i] = "'" + strings.ReplaceAll(s, "'", "\\'") + "'"
		}
		return "use " + name + " (" + strings.Join(q, ",") + ");"
	case kind == 'm':
		return "use " + name + " ();"
	default:
		return "use " + name + ";"
	}
}

// perlQuote escapes s for a single-quoted Perl string literal.
func perlQuote(s string) string {
	return strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s)
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
