package gperl

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

func ptr[T any](v T) *T { return &v }

// Each case is a perl command line and the program it describes.
func TestParsePerlCommand(t *testing.T) {
	cases := []struct {
		name string
		argv []string
		env  []string
		want perl.Program
	}{
		{"script and arguments", []string{"a.pl", "-x", "y"},
			nil, perl.Program{File: "a.pl", Args: []string{"-x", "y"}}},
		{"no arguments reads standard input", nil,
			nil, perl.Program{Stdin: true}},
		{"dash reads standard input", []string{"-", "a"},
			nil, perl.Program{Stdin: true, Args: []string{"a"}}},
		{"switches end at --", []string{"-w", "--", "-n.pl", "a"},
			nil, perl.Program{File: "-n.pl", Args: []string{"a"}, Warnings: perl.WarningsOn}},
		{"-e joined and separate, repeated", []string{"-eprint 1", "-e", "print 2", "a"},
			nil, perl.Program{Code: []string{"print 1", "print 2"}, Args: []string{"a"}}},
		{"-e with an empty program", []string{"-e", ""},
			nil, perl.Program{Code: []string{""}}},
		{"-E enables features", []string{"-E", "say 1"},
			nil, perl.Program{Code: []string{"say 1"}, Features: true}},
		{"clustered switches", []string{"-wnle", "print"},
			nil, perl.Program{Code: []string{"print"}, Warnings: perl.WarningsOn, Loop: perl.LoopRead, LineEnd: ptr("\n")}},
		{"-p wins over -n", []string{"-p", "-n", "-e1"},
			nil, perl.Program{Code: []string{"1"}, Loop: perl.LoopPrint}},
		{"-a implies -n", []string{"-ae", "1"},
			nil, perl.Program{Code: []string{"1"}, Loop: perl.LoopRead, AutoSplit: true}},
		{"-F takes the rest and implies -a -n", []string{"-F:", "-e1"},
			nil, perl.Program{Code: []string{"1"}, Loop: perl.LoopRead, AutoSplit: true, SplitPattern: ptr(":")}},
		{"-I joined and separate", []string{"-Ilib", "-I", "t/lib", "a.pl"},
			nil, perl.Program{File: "a.pl", Inc: []string{"lib", "t/lib"}}},
		{"-M forms", []string{"-MFoo", "-MBar=a,b", "-M-strict", "-mBaz", "-MQux qw(x)", "-M", "Sep", "-e1"},
			nil, perl.Program{Code: []string{"1"}, Preamble: []string{
				"use Foo", "use Bar split(/,/,q\x00a,b\x00)", " no strict", "use Baz ()", "use Qux qw(x)", "use Sep",
			}}},
		{"-0 alone is NUL", []string{"-0", "-e1"},
			nil, perl.Program{Code: []string{"1"}, RecordSeparator: ptr("\x00")}},
		{"-00 is paragraph mode", []string{"-00", "-e1"},
			nil, perl.Program{Code: []string{"1"}, RecordSeparator: ptr("")}},
		{"-0777 slurps", []string{"-0777", "-e1"},
			nil, perl.Program{Code: []string{"1"}, Slurp: true}},
		{"-0 octal byte", []string{"-0351", "-e1"},
			nil, perl.Program{Code: []string{"1"}, RecordSeparator: ptr("\xe9")}},
		{"-0 hexadecimal character", []string{"-0x263a", "-e1"},
			nil, perl.Program{Code: []string{"1"}, RecordSeparator: ptr("☺")}},
		{"-0 then clustered switch", []string{"-072ne", "1"},
			nil, perl.Program{Code: []string{"1"}, RecordSeparator: ptr(":"), Loop: perl.LoopRead}},
		{"-0xDIR is -0 -xDIR", []string{"-0xdir/sub", "a.pl"},
			nil, perl.Program{File: "a.pl", RecordSeparator: ptr("\x00"), Extract: true, ExtractDir: "dir/sub"}},
		{"-g slurps", []string{"-ge", "1"},
			nil, perl.Program{Code: []string{"1"}, Slurp: true}},
		{"-l copies the separator in effect", []string{"-00", "-l", "-e1"},
			nil, perl.Program{Code: []string{"1"}, RecordSeparator: ptr(""), LineEnd: ptr("\n\n")}},
		{"-l before -0 keeps newline", []string{"-l", "-0", "-e1"},
			nil, perl.Program{Code: []string{"1"}, RecordSeparator: ptr("\x00"), LineEnd: ptr("\n")}},
		{"-l with octal", []string{"-l054e", "1"},
			nil, perl.Program{Code: []string{"1"}, LineEnd: ptr(",")}},
		{"-i with and without suffix", []string{"-pi.bak", "-e1", "f"},
			nil, perl.Program{Code: []string{"1"}, Loop: perl.LoopPrint, InPlace: ptr(".bak"), Args: []string{"f"}}},
		{"-i bare", []string{"-i", "-pe1"},
			nil, perl.Program{Code: []string{"1"}, Loop: perl.LoopPrint, InPlace: ptr("")}},
		{"-C forms", []string{"-C0", "-e1"},
			nil, perl.Program{Code: []string{"1"}, Unicode: ptr(uint32(0))}},
		{"-C letters", []string{"-CSA", "-e1"},
			nil, perl.Program{Code: []string{"1"}, Unicode: ptr(uint32(0x27))}},
		{"-C bare", []string{"-C", "-e1"},
			nil, perl.Program{Code: []string{"1"}, Unicode: ptr(uint32(0x5f))}},
		{"warning switches", []string{"-W", "-w", "-e1"},
			nil, perl.Program{Code: []string{"1"}, Warnings: perl.WarningsAll}},
		{"-X", []string{"-X", "-e1"},
			nil, perl.Program{Code: []string{"1"}, Warnings: perl.WarningsNone}},
		{"taint", []string{"-T", "a.pl"},
			nil, perl.Program{File: "a.pl", Taint: perl.TaintOn}},
		{"taint warnings", []string{"-t", "a.pl"},
			nil, perl.Program{File: "a.pl", Taint: perl.TaintWarn}},
		{"-c -U -f", []string{"-cUf", "a.pl"},
			nil, perl.Program{File: "a.pl", SyntaxOnly: true, Unsafe: true}},
		{"-x with directory", []string{"-x/tmp", "a.pl"},
			nil, perl.Program{File: "a.pl", Extract: true, ExtractDir: "/tmp"}},
		{"-s takes leading arguments", []string{"-s", "a.pl", "-v", "-name=x=y", "--", "-kept", "b"},
			nil, perl.Program{File: "a.pl", Defines: []perl.Define{{Name: "v"}, {Name: "name", Value: ptr("x=y")}}, Args: []string{"-kept", "b"}}},
		{"-V reads no program", []string{"-V:osname"},
			nil, perl.Program{File: "/dev/null", Preamble: []string{"use Config; Config::config_vars(qw\x00osname\x00)"}}},
		{"-d", []string{"-d", "a.pl"},
			nil, perl.Program{File: "a.pl", Debug: true}},
		{"PERL5OPT", []string{"-Ilib", "a.pl"}, []string{"PERL5OPT=-Mstrict -w -Iopt"},
			perl.Program{File: "a.pl", Inc: []string{"opt", "lib"}, Preamble: []string{"use strict"}, Warnings: perl.WarningsOn}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := parsePerlCommand(tc.argv, tc.env)
			if err != nil {
				t.Fatalf("parsePerlCommand(%q): %v", tc.argv, err)
			}
			if cmd.action != actionRun {
				t.Fatalf("action = %d, want a run", cmd.action)
			}
			got := cmd.prog
			if len(got.Args) == 0 {
				got.Args = nil
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("parsePerlCommand(%q)\n got %+v\nwant %+v", tc.argv, describe(got), describe(tc.want))
			}
		})
	}
}

// describe renders a Program with its pointer fields dereferenced.
func describe(p perl.Program) map[string]any {
	m := map[string]any{}
	v := reflect.ValueOf(p)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.IsZero() {
			continue
		}
		if f.Kind() == reflect.Pointer {
			f = f.Elem()
		}
		m[v.Type().Field(i).Name] = f.Interface()
	}
	return m
}

func TestParsePerlCommandDebuggerModule(t *testing.T) {
	cmd, err := parsePerlCommand([]string{"-d:NYTProf=start=no", "a.pl"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !cmd.prog.Debug || !reflect.DeepEqual(cmd.env, []string{"PERL5DB=use Devel::NYTProf split(/,/,q{start=no});"}) {
		t.Fatalf("Debug=%v env=%q", cmd.prog.Debug, cmd.env)
	}
}

func TestParsePerlCommandActions(t *testing.T) {
	for argv, want := range map[string]cliAction{
		"-v": actionVersion, "--version": actionVersion,
		"-h": actionUsage, "--help": actionUsage, "-w -?": actionUsage,
	} {
		cmd, err := parsePerlCommand(strings.Fields(argv), nil)
		if err != nil || cmd.action != want {
			t.Fatalf("%s: action=%v err=%v, want %v", argv, cmd, err, want)
		}
	}
}

func TestParsePerlCommandErrors(t *testing.T) {
	for _, tc := range []struct {
		argv []string
		env  []string
		want string
	}{
		{[]string{"-e"}, nil, "No code specified for -e"},
		{[]string{"-I"}, nil, "No directory specified for -I"},
		{[]string{"-M"}, nil, "Missing argument to -M"},
		{[]string{"-M=x", "-e1"}, nil, "Module name required with -M option"},
		{[]string{"-mFoo=", "-e1"}, nil, ""},
		{[]string{"-mFoo bar", "-e1"}, nil, "Can't use ' ' after -mname"},
		{[]string{"-z"}, nil, "Unrecognized switch: -z  (-h will show valid options)"},
		{[]string{"--frob"}, nil, "Unrecognized switch: --frob  (-h will show valid options)"},
		{[]string{"-CZ", "-e1"}, nil, "Unknown Unicode option letter 'Z'"},
		{[]string{"-C512", "-e1"}, nil, "Unknown Unicode option value 512"},
		{[]string{"-e1"}, []string{"PERL5OPT=-e1"}, "Illegal switch in PERL5OPT: -e"},
	} {
		_, err := parsePerlCommand(tc.argv, tc.env)
		got := ""
		if err != nil {
			got = err.Error()
		}
		if got != tc.want {
			t.Errorf("parsePerlCommand(%q): error %q, want %q", tc.argv, got, tc.want)
		}
	}
}

// runCLI runs a perl command line in dir and returns the process's
// status and output.
func runCLI(t *testing.T, dir, stdin string, argv ...string) (status int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	status, err := runPerlCLI(argv, perlRunOpts{
		dir: dir, stdin: strings.NewReader(stdin), stdout: &out, stderr: &errOut,
	})
	if err != nil {
		t.Fatalf("perl %q: %v", argv, err)
	}
	return status, out.String(), errOut.String()
}

// The command lines the CPAN toolchain and its test suites use.
func TestPerlCommandLines(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("caller.pl", "print defined(caller()) ? \"called\\n\" : \"main\\n\"; print \"$0 @ARGV\\n\";\n")
	write("shebang.pl", "#!perl -l\nprint 'x'; print 'y';\n")
	write("taint.t", "#!/usr/bin/perl -T\nprint \"taint ${^TAINT}\\n\";\n")
	write("lines.txt", "b 2\na 1\n")
	write("Mod.pm", "package Mod; sub import { shift; print \"import(@_)\\n\" } 1;\n")

	for _, tc := range []struct {
		name   string
		stdin  string
		argv   []string
		status int
		stdout string
		stderr string
	}{
		{name: "script is the main program", argv: []string{"caller.pl", "a", "b"},
			stdout: "main\ncaller.pl a b\n"},
		{name: "program on standard input", stdin: "print qq(from stdin\\n);",
			stdout: "from stdin\n"},
		{name: "program on standard input, dash", stdin: "print qq(@ARGV\\n);", argv: []string{"-", "x"},
			stdout: "x\n"},
		{name: "-C0", argv: []string{"-C0", "-e", `print "ok ${^UNICODE}\n"`},
			stdout: "ok 0\n"},
		{name: "-n loop", stdin: "a\nb\n", argv: []string{"-ne", "print uc"},
			stdout: "A\nB\n"},
		{name: "-lane over a file", argv: []string{"-lane", "print $F[1]", "lines.txt"},
			stdout: "2\n1\n"},
		{name: "switches on the #! line", argv: []string{"shebang.pl"},
			stdout: "x\ny\n"},
		{name: "-T on the command line and the #! line", argv: []string{"-T", "taint.t"},
			stdout: "taint 1\n"},
		{name: "-T only on the #! line", argv: []string{"taint.t"},
			status: 255, stderr: "\"-T\" is on the #! line, it must also be used on the command line at taint.t line 1.\n"},
		{name: "-M with an import list", argv: []string{"-I.", "-MMod=a,b", "-e", "1"},
			stdout: "import(a b)\n"},
		{name: "-m imports nothing", argv: []string{"-I.", "-mMod", "-e", "print qq(loaded\\n) if $INC{'Mod.pm'}"},
			stdout: "loaded\n"},
		{name: "-s", argv: []string{"-s", "-e", `print "$x $y @ARGV\n"`, "--", "-x", "-y=5", "rest"},
			stdout: "1 5 rest\n"},
		{name: "-c", argv: []string{"-c", "caller.pl"},
			stderr: "caller.pl syntax OK\n"},
		{name: "-V:name", argv: []string{"-V:version"},
			stdout: "version='5.44.0';\n"},
		{name: "exit status", argv: []string{"-e", "exit 3"},
			status: 3},
		{name: "END sets the status", argv: []string{"-e", "END { $? = 9 }"},
			status: 9},
		{name: "unknown switch", argv: []string{"-z"},
			status: 255, stderr: "Unrecognized switch: -z  (-h will show valid options).\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, stdout, stderr := runCLI(t, dir, tc.stdin, tc.argv...)
			if status != tc.status || stdout != tc.stdout || stderr != tc.stderr {
				t.Fatalf("perl %q\n got status=%d stdout=%q stderr=%q\nwant status=%d stdout=%q stderr=%q",
					tc.argv, status, stdout, stderr, tc.status, tc.stdout, tc.stderr)
			}
		})
	}
}

func TestPerlVersionAndUsage(t *testing.T) {
	status, stdout, _ := runCLI(t, t.TempDir(), "", "-v")
	if status != 0 || !strings.HasPrefix(stdout, "\nThis is perl 5, version 44, subversion 0 (v5.44.0) built for ") {
		t.Fatalf("-v: status=%d stdout=%q", status, stdout)
	}
	status, stdout, _ = runCLI(t, t.TempDir(), "", "-h")
	if status != 0 || !strings.HasPrefix(stdout, "\nUsage: gperl run [switches] [--] [programfile] [arguments]\n") {
		t.Fatalf("-h: status=%d stdout=%q", status, stdout)
	}
}
