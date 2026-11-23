package perl_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

type programRun struct {
	status int
	stdout string
	stderr string
}

// runProgram runs prog on a fresh instance with stdin as its standard
// input and returns what the "process" left behind. A cfg with no FS runs
// in the default in-memory sandbox.
func runProgram(t *testing.T, cfg perl.Config, prog perl.Program, stdin string) programRun {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cfg.Stdin = strings.NewReader(stdin)
	cfg.Stdout = &stdout
	cfg.Stderr = &stderr
	p, err := perl.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	status, err := p.RunProgram(context.Background(), prog)
	if err != nil {
		p.Close()
		t.Fatalf("RunProgram: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return programRun{status: status, stdout: stdout.String(), stderr: stderr.String()}
}

func hostConfig(t *testing.T) perl.Config {
	t.Helper()
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("ExtractStdlib: %v", err)
	}
	return perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib}
}

func (r programRun) want(t *testing.T, status int, stdout string) {
	t.Helper()
	if r.status != status || r.stdout != stdout {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=%d stdout=%q", r.status, r.stdout, r.stderr, status, stdout)
	}
}

func code(lines ...string) []string { return lines }

func ptr[T any](v T) *T { return &v }

// The program text is the main program: nothing called it.
func TestRunProgramHasNoCaller(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`sub f { my @c = caller(1); print defined $c[3] ? "outer frame: $c[3]\n" : "no outer frame\n" }`,
		`f();`,
		`print defined(caller()) ? "has a caller\n" : "no caller\n";`,
		`print "$0\n";`,
	)}, "")
	r.want(t, 0, "no outer frame\nno caller\n-e\n")
}

func TestRunProgramPhases(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`print "run ${^GLOBAL_PHASE}\n";`,
		`END { print "end ${^GLOBAL_PHASE}\n" }`,
		`INIT { print "init ${^GLOBAL_PHASE}\n" }`,
		`CHECK { print "check ${^GLOBAL_PHASE}\n" }`,
		`BEGIN { print "begin ${^GLOBAL_PHASE}\n" }`,
	)}, "")
	r.want(t, 0, "begin START\ncheck CHECK\ninit INIT\nrun RUN\nend END\n")
}

// The exit status is the one the process would end with: exit's argument,
// as an END block may have changed it.
func TestRunProgramExitStatus(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(`print "a\n"; exit 7; print "b\n";`)}, "")
	r.want(t, 7, "a\n")

	r = runProgram(t, perl.Config{}, perl.Program{Code: code(`END { $? = 3 } print "ran\n";`)}, "")
	r.want(t, 3, "ran\n")

	r = runProgram(t, perl.Config{}, perl.Program{Code: code(`END { print "end\n" } BEGIN { exit 4 } print "never\n";`)}, "")
	r.want(t, 4, "end\n")
}

// An uncaught die is reported by perl itself, with perl's status.
func TestRunProgramUncaughtDie(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(`$! = 0; die "kaboom\n";`)}, "")
	r.want(t, 255, "")
	if r.stderr != "kaboom\n" {
		t.Fatalf("stderr = %q, want the die message", r.stderr)
	}

	r = runProgram(t, perl.Config{}, perl.Program{Code: code(`$! = 5; die "with errno\n";`)}, "")
	r.want(t, 5, "")

	r = runProgram(t, perl.Config{}, perl.Program{Code: code(`$SIG{__DIE__} = sub { print "handler saw: $_[0]" }; $! = 0; die "x\n";`)}, "")
	r.want(t, 255, "handler saw: x\n")
}

// A die the program catches resumes the program.
func TestRunProgramCaughtDie(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`for my $i (1, 2) { eval { die "oops $i\n" }; print "caught $@"; }`,
		`my $v = eval "1 +"; print defined $v ? "compiled\n" : "syntax error caught\n";`,
		`print "done\n";`,
	)}, "")
	r.want(t, 0, "caught oops 1\ncaught oops 2\nsyntax error caught\ndone\n")
}

func TestRunProgramCompileError(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(`print "x"; my $v = ;`)}, "")
	if r.status != 255 || r.stdout != "" || !strings.Contains(r.stderr, "Execution of -e aborted due to compilation errors.") {
		t.Fatalf("status=%d stdout=%q stderr=%q, want perl's compilation failure", r.status, r.stdout, r.stderr)
	}
}

func TestRunProgramArgsIncPreambleDefines(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Probe.pm"), []byte("package Probe; sub answer { 42 } 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runProgram(t, hostConfig(t), perl.Program{
		Code:     code(`print join(",", @ARGV), " ", Probe::answer(), " ", sum(1, 2), " $flag $name $INC[0]\n";`),
		Args:     []string{"a", "b c"},
		Inc:      []string{dir},
		Preamble: []string{"use Probe;", "use List::Util qw(sum);"},
		Defines:  []perl.Define{{Name: "flag"}, {Name: "name", Value: ptr("v=1")}},
	}, "")
	r.want(t, 0, "a,b c 42 3 1 v=1 "+dir+"\n")
}

// A directory given to @INC brings its version- and architecture-specific
// subdirectories along, ahead of itself, as perl -I does.
func TestRunProgramIncSubdirectories(t *testing.T) {
	names := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Config; print "$Config{version} $Config{archname}";`,
	)}, "")
	version, arch, ok := strings.Cut(names.stdout, " ")
	if names.status != 0 || !ok {
		t.Fatalf("version and archname: %+v", names)
	}
	dir := t.TempDir()
	for _, sub := range []string{filepath.Join(version, arch), arch} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	plain := t.TempDir()
	r := runProgram(t, hostConfig(t), perl.Program{
		Code: code(`print join("\n", @INC[0 .. 4]), "\n";`),
		Inc:  []string{dir, plain},
	}, "")
	r.want(t, 0, strings.Join([]string{
		filepath.Join(dir, version, arch),
		filepath.Join(dir, version),
		filepath.Join(dir, arch),
		dir,
		plain,
	}, "\n")+"\n")
}

func TestRunProgramFromStdin(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Stdin: true, Args: []string{"x"}},
		"print \"$0 @ARGV\\n\";\nprint defined(caller()) ? \"caller\\n\" : \"main\\n\";\n")
	r.want(t, 0, "- x\nmain\n")
}

// What follows __END__ in a program read from standard input is the
// program's input.
func TestRunProgramFromStdinKeepsRestAsInput(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Stdin: true},
		"print uc while <STDIN>;\n__END__\none\ntwo\n")
	r.want(t, 0, "ONE\nTWO\n")
}

func TestRunProgramFile(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "prog.pl")
	src := "#!/usr/bin/perl -l\nprint $0 eq $ARGV[0] ? 'zero is the path' : \"zero is $0\";\nprint __FILE__ eq $0 ? 'file' : 'not file', ' line ', __LINE__;\nprint scalar <DATA>;\n__DATA__\npayload\n"
	if err := os.WriteFile(script, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runProgram(t, hostConfig(t), perl.Program{File: script, Args: []string{script}}, "")
	// -l on the #! line: every print ends in a newline (the DATA line
	// keeps its own, too).
	r.want(t, 0, "zero is the path\nfile line 3\npayload\n\n")
}

func TestRunProgramMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent.pl")
	r := runProgram(t, hostConfig(t), perl.Program{File: missing}, "")
	want := "Can't open perl script \"" + missing + "\": No such file or directory\n"
	// The status is the guest's errno for a missing file.
	if r.status == 0 || r.stderr != want {
		t.Fatalf("status=%d stderr=%q, want a failure with stderr=%q", r.status, r.stderr, want)
	}
}

func TestRunProgramLoops(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(`print uc`), Loop: perl.LoopRead}, "a\nb\n")
	r.want(t, 0, "A\nB\n")

	r = runProgram(t, perl.Config{}, perl.Program{Code: code(`s/a/x/`), Loop: perl.LoopPrint}, "a\nba\n")
	r.want(t, 0, "x\nbx\n")

	r = runProgram(t, perl.Config{}, perl.Program{
		Code: code(`print "$F[1]-$F[0]"`), Loop: perl.LoopRead, AutoSplit: true, LineEnd: ptr("\n"),
	}, "a b\nc d\n")
	r.want(t, 0, "b-a\nd-c\n")

	r = runProgram(t, perl.Config{}, perl.Program{
		Code: code(`print $F[1]`), Loop: perl.LoopRead, AutoSplit: true, SplitPattern: ptr(":"), LineEnd: ptr("|"),
	}, "a:b\nc:d\n")
	r.want(t, 0, "b|d|")

	r = runProgram(t, perl.Config{}, perl.Program{
		Code: code(`print "[$_]"`), Loop: perl.LoopRead, RecordSeparator: ptr(";"),
	}, "a;b;c")
	r.want(t, 0, "[a;][b;][c]")

	r = runProgram(t, perl.Config{}, perl.Program{
		Code: code(`print length`), Loop: perl.LoopRead, Slurp: true,
	}, "a\nb\nc\n")
	r.want(t, 0, "6")

	r = runProgram(t, perl.Config{}, perl.Program{
		Code: code(`print "[$_]"`), Loop: perl.LoopRead, RecordSeparator: ptr(""),
	}, "a\nb\n\n\nc\n")
	r.want(t, 0, "[a\nb\n\n][c\n]")
}

func TestRunProgramInPlaceEdit(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(file, []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runProgram(t, hostConfig(t), perl.Program{
		Code: code(`s/o/0/g`), Loop: perl.LoopPrint, InPlace: ptr(".bak"), Args: []string{file},
	}, "")
	r.want(t, 0, "")
	got, err := os.ReadFile(file)
	if err != nil || string(got) != "0ne\ntw0\n" {
		t.Fatalf("edited file = %q, %v", got, err)
	}
	backup, err := os.ReadFile(file + ".bak")
	if err != nil || string(backup) != "one\ntwo\n" {
		t.Fatalf("backup = %q, %v", backup, err)
	}
}

func TestRunProgramWarnings(t *testing.T) {
	const src = `my $u; my $s = "a" . $u; print "ok\n";`
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(src)}, "")
	r.want(t, 0, "ok\n")
	if r.stderr != "" {
		t.Fatalf("without -w: stderr = %q", r.stderr)
	}
	r = runProgram(t, perl.Config{}, perl.Program{Code: code(src), Warnings: perl.WarningsOn}, "")
	r.want(t, 0, "ok\n")
	if r.stderr != "Use of uninitialized value $u in concatenation (.) or string at -e line 1.\n" {
		t.Fatalf("with -w: stderr = %q", r.stderr)
	}
	r = runProgram(t, perl.Config{}, perl.Program{Code: code(`use warnings; ` + src), Warnings: perl.WarningsNone}, "")
	r.want(t, 0, "ok\n")
	if r.stderr != "" {
		t.Fatalf("with -X: stderr = %q", r.stderr)
	}
	r = runProgram(t, perl.Config{}, perl.Program{Code: code(`no warnings; ` + src), Warnings: perl.WarningsAll}, "")
	r.want(t, 0, "ok\n")
	if !strings.Contains(r.stderr, "uninitialized value $u") {
		t.Fatalf("with -W: stderr = %q", r.stderr)
	}
}

func TestRunProgramSyntaxOnly(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{
		Code: code(`BEGIN { print "begin\n" } END { print "end\n" } print "run\n";`), SyntaxOnly: true,
	}, "")
	r.want(t, 0, "begin\n")
	if r.stderr != "-e syntax OK\n" {
		t.Fatalf("stderr = %q", r.stderr)
	}
}

func TestRunProgramFeatures(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(`say "hello"`), Features: true}, "")
	r.want(t, 0, "hello\n")

	r = runProgram(t, perl.Config{}, perl.Program{Code: code(`say "hello"`)}, "")
	if r.status == 0 {
		t.Fatalf("say compiled without the feature bundle: stdout=%q", r.stdout)
	}
}

func TestRunProgramUnicode(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{
		Code:    code(`print chr(0x263a), " ", length($ARGV[0]), " ${^UNICODE}\n";`),
		Unicode: ptr(perl.UnicodeStdout | perl.UnicodeArgv),
		Args:    []string{"☺"},
	}, "")
	r.want(t, 0, "☺ 1 34\n")
	if r.stderr != "" {
		t.Fatalf("stderr = %q", r.stderr)
	}
}

func TestRunProgramTaint(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{
		Code:  code(`print "${^TAINT}\n"; my $ok = eval { open(my $fh, '>', $ARGV[0]); 1 }; print $ok ? "allowed\n" : "refused: $@";`),
		Taint: perl.TaintOn,
		Args:  []string{"/tmp/tainted-name"},
	}, "")
	if r.status != 0 || !strings.HasPrefix(r.stdout, "1\nrefused: Insecure dependency in open while running with -T switch") {
		t.Fatalf("status=%d stdout=%q stderr=%q", r.status, r.stdout, r.stderr)
	}
}

// A taint failure that an eval caught leaves no taint behind: the code
// compiled next loads its modules as usual.
func TestRunProgramTaintFailureInsideEvalDoesNotSpread(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{
		Code: code(
			// the refused eval is the last thing its BEGIN block does
			`BEGIN { eval { eval("1" . substr($^X, 0, 0)); print "not refused\n" } }`,
			`use Text::Wrap; print "loaded\n";`,
		),
		Taint: perl.TaintOn,
	}, "")
	r.want(t, 0, "loaded\n")
}

func TestRunProgramExtract(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Stdin: true, Extract: true},
		"To: someone\nSubject: a program in a message\n\n#!perl -l\nprint 'extracted';\nprint __LINE__;\n__END__\nsignature\n")
	r.want(t, 0, "extracted\n3\n")
}

func TestRunProgramCancel(t *testing.T) {
	var stderr bytes.Buffer
	p, err := perl.New(perl.Config{Stderr: &stderr})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = p.RunProgram(ctx, perl.Program{Code: code(`my $n = 0; while (1) { $n++ }`)})
	if err == nil || err != ctx.Err() {
		t.Fatalf("err = %v, want ctx.Err() (%v)", err, ctx.Err())
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("cancellation took %v", elapsed)
	}
}

// $^T is the time the program started, and the ages of files (-M, -A, -C)
// count from it — also in an interpreter made long after the first of the
// process, which is a copy of that one.
func TestBaseTimeIsTheStartOfTheProgram(t *testing.T) {
	first, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	first.Close()
	time.Sleep(2100 * time.Millisecond)
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`my $late = time - $^T;`,
		`print $late >= 0 && $late <= 1 ? "started now\n" : "started $late seconds ago\n";`,
	)}, "")
	r.want(t, 0, "started now\n")
}

// times reports seconds: what the C library returns is scaled by the same
// clock tick perl divides by. The time of a program that is the host
// process is the process's (Config.HostProcess), so it lies between what
// the process had used before the program ran and what it has used after
// (where the test can ask; elsewhere it is at least not in ticks).
func TestTimesReportsSeconds(t *testing.T) {
	before, known := processUserSeconds()
	r := runProgram(t, perl.Config{HostProcess: true}, perl.Program{Code: code(
		`my $n = 0; $n++ for 1 .. 200000;`,
		`my ($user) = times;`,
		`print "$user\n";`,
	)}, "")
	after, _ := processUserSeconds()
	user, err := strconv.ParseFloat(strings.TrimSpace(r.stdout), 64)
	if r.status != 0 || err != nil {
		t.Fatalf("status=%d stdout=%q stderr=%q", r.status, r.stdout, r.stderr)
	}
	if !known {
		before, after = 0, 60
	}
	// times counts in clock ticks of a hundredth of a second
	if user < before-0.02 || user > after+0.02 {
		t.Fatalf("times reports %v seconds of user time; the process had used %v before the program and %v after", user, before, after)
	}
}

// fork is a capability: denied unless Config.Fork grants it, and granted
// on its own — a forked child needs no permission to spawn host processes.
func TestForkIsDeniedByDefault(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`my $pid = fork; print defined $pid ? "forked\n" : "denied\n";`,
	)}, "")
	r.want(t, 0, "denied\n")
}

// Loading native code is a capability too: without Config.NativeModules
// DynaLoader has nothing to load a shared object with.
func TestNativeModulesAreDeniedByDefault(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`require DynaLoader;`,
		`print defined(&DynaLoader::dl_load_file) ? "can load\n" : "cannot load\n";`,
	)}, "")
	r.want(t, 0, "cannot load\n")
}

func TestForkWithoutExec(t *testing.T) {
	r := runProgram(t, perl.Config{Fork: true}, perl.Program{Code: code(
		`pipe(my $r, my $w) or die "pipe: $!";`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) { close $r; syswrite($w, "hello from the child\n"); exit 3 }`,
		`close $w; my $line = <$r>; waitpid($pid, 0);`,
		`print $line, "status ", $? >> 8, "\n";`,
	)}, "")
	r.want(t, 0, "hello from the child\nstatus 3\n")
}

func TestRunProgramRejectsAmbiguousSource(t *testing.T) {
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	for _, prog := range []perl.Program{
		{},
		{File: "a.pl", Stdin: true},
		{File: "a.pl", Code: code("1")},
	} {
		if _, err := p.RunProgram(context.Background(), prog); err == nil {
			t.Fatalf("RunProgram(%+v) succeeded", prog)
		}
	}
}
