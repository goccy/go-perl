package gperl

// A perl-compatible command line over the embedded interpreter.
//
// The XS build pipeline is driven entirely by Perl programs — Makefile.PL,
// Build.PL, xsubpp, ExtUtils::Command one-liners from the generated
// Makefile, cpanm — and all of them are pure Perl. RunPerlCLI runs them on
// the embedded interpreter, so building native modules needs no perl
// installed on the system: gperl IS the perl that drives the build (the
// remaining external tools are the C toolchain: cc and make).
//
// The flag surface is the subset those build tools use: -e, -I, -M/-m, -w,
// `--`, and script execution. Anything else is rejected loudly rather than
// half-interpreted.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	perl "github.com/goccy/go-perl"
)

// RunPerlCLI executes a perl(1)-style command line on the embedded
// interpreter and returns the process exit status. Recognized environment:
//
//   - GPERL_PERL_EXE: the value for $^X (the perl shim path, so build
//     systems that re-invoke $^X — Module::Build's ./Build — come back
//     into the embedded interpreter);
//   - GPERL_PERL_PRELOAD: a Perl file `do`ne before anything else (the XS
//     build's %Config toolchain overlay).
func RunPerlCLI(argv []string) (status int, err error) {
	var (
		inc      []string
		preUse   []string
		eChunks  []string
		script   string
		args     []string
		autoLine bool
	)
	i := 0
	for ; i < len(argv); i++ {
		a := argv[i]
		if a == "--" {
			i++
			goto rest
		}
		if len(a) < 2 || a[0] != '-' {
			// first non-option: the script
			if len(eChunks) == 0 {
				script = a
				i++
			}
			goto rest
		}
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
						return 2, fmt.Errorf("-e requires an argument")
					}
					v = argv[i]
				}
				eChunks = append(eChunks, v)
				j = len(a)
			case 'I':
				v := a[j+1:]
				if v == "" {
					i++
					if i >= len(argv) {
						return 2, fmt.Errorf("-I requires an argument")
					}
					v = argv[i]
				}
				inc = append(inc, v)
				j = len(a)
			case 'M', 'm':
				v := a[j+1:]
				if v == "" {
					i++
					if i >= len(argv) {
						return 2, fmt.Errorf("-%c requires an argument", a[j])
					}
					v = argv[i]
				}
				preUse = append(preUse, useStatement("-"+string(a[j])+v))
				j = len(a)
			case 'l':
				// output half of -l (input auto-chomp belongs to the
				// unsupported -n/-p loops)
				autoLine = true
			case 'w', 'W', 'X', 'T', 't':
				// warning/taint toggles: accepted for compatibility; the
				// embedded build has no taint support and warnings stay
				// as the program sets them
			default:
				return 2, fmt.Errorf("unsupported perl option %q", a)
			}
		}
	}
rest:
	if i < len(argv) {
		args = append(args, argv[i:]...)
	}
	if script == "" && len(eChunks) == 0 {
		return 2, fmt.Errorf("no -e code and no script given")
	}

	// The launcher runs as a synthesized program file so the ordinary
	// run-file conventions (die, exit, @ARGV, $0) apply unchanged.
	cwd, err := os.Getwd()
	if err != nil {
		return 1, err
	}
	os.Setenv("GPERL_PERL_CWD", cwd)
	var prog strings.Builder
	// perl resolves '.'-relative requires and files against ITS cwd; make
	// the guest's match the process's.
	prog.WriteString("BEGIN { chdir $ENV{GPERL_PERL_CWD} or die \"chdir $ENV{GPERL_PERL_CWD}: $!\\n\"; }\n")
	if pre := os.Getenv("GPERL_PERL_PRELOAD"); pre != "" {
		fmt.Fprintf(&prog, "BEGIN { my $p = $ENV{GPERL_PERL_PRELOAD}; my $r = do $p; die $@ if $@; die \"preload $p: $!\\n\" if !defined $r && $!; }\n")
	}
	if exe := os.Getenv("GPERL_PERL_EXE"); exe != "" {
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
		abs, aerr := filepath.Abs(script)
		if aerr != nil {
			return 1, aerr
		}
		body, rerr := os.ReadFile(abs)
		if rerr != nil {
			return 1, rerr
		}
		// BEGIN: the script body's own compile-time code (a
		// `use inc::Module::Install` reads $0 through FindBin) must
		// already see the script's identity.
		q := strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(abs)
		prog.WriteString("BEGIN { $0 = '" + q + "'; }\n")
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
	cfg.Stdout = os.Stdout
	cfg.Stderr = os.Stderr
	cfg.Env = os.Environ()
	p, err := perl.New(cfg)
	if err != nil {
		return 1, err
	}
	runErr := p.RunFile(context.Background(), tmpPath, inc, args)
	p.Close()
	return statusFromRun(runErr)
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
