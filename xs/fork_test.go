//go:build darwin || linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// runForkingProgram runs code as a program that may fork, on an instance
// with the PerlAPI fixture available to `use`.
func runForkingProgram(t *testing.T, code ...string) (status int, stdout, stderr string) {
	t.Helper()
	return runForkingProgramWith(t, buildPerlAPI(t), code...)
}

// runForkingProgramWith is runForkingProgram with the PerlAPI fixture
// already built at so.
func runForkingProgramWith(t *testing.T, so string, code ...string) (status int, stdout, stderr string) {
	t.Helper()
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var out, errOut bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, Fork: true, NativeModules: true, Stdout: &out, Stderr: &errOut})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	loadXSModule(t, p, "PerlAPI", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "PerlAPI.pm"),
		[]byte("package PerlAPI;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err = p.RunProgram(context.Background(), perl.Program{Code: code, Inc: []string{libdir}})
	if err != nil {
		p.Close()
		t.Fatalf("RunProgram: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return status, out.String(), errOut.String()
}

// A fork child has its memory in a place of its own: pack 'P' there makes
// the address of the child's string, which unpack reads back, and what the
// child changes the parent does not see.
func TestForkPackedAddressesAreTheChilds(t *testing.T) {
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		`my $buf = "parent";`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) { substr($buf, 0, 6, "child!"); exit(unpack("P6", pack("P", $buf)) eq "child!" ? 7 : 1) }`,
		`waitpid $pid, 0;`,
		`print "child ", $? >> 8, " parent ", unpack("P6", pack("P", $buf)), "\n";`,
	)
	if status != 0 || stdout != "child 7 parent parent\n" {
		t.Fatalf("status %d, stdout %q, stderr %q; want child 7 parent parent", status, stdout, stderr)
	}
}

// forkRoute is one way a fork child's interpreter could enter native code:
// what the parent sets up before it forks, what the child then evaluates,
// and what the parent evaluates when the child has ended.
type forkRoute struct {
	name, setup, child, after string
	// want is the child's line (what its code returned, or why it died)
	// and the parent's value.
	want string
}

// forkRouteProgram is the program that forks for r: the child reports
// what its code returned or why it died (the message without its place,
// and the directory of a library left out), and ends with status 0.
func forkRouteProgram(r forkRoute) []string {
	return []string{
		`use strict; use warnings; no warnings 'once'; $| = 1;`,
		`sub why { my ($e) = split /\n/, shift; $e =~ s{[^\s()]*/(\w+\.so)}{$1}g; $e =~ s/0x[0-9a-f]+/0x.../g; $e =~ s/ at (?:\S+|\(eval \d+\)) line \d+\.\z//; $e }`,
		r.setup,
		`pipe(my $r, my $w) or die "pipe: $!";`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) { close $r; my $out = eval { ` + r.child + ` }; syswrite($w, (defined $out ? "returned $out" : "died: " . why($@)) . "\n"); exit 0 }`,
		`close $w; my $line = <$r>; waitpid($pid, 0);`,
		`print "child: ", $line // "nothing\n", "status ", $? >> 8, "\n", "parent: ", do { ` + r.after + ` }, "\n";`,
	}
}

// Native code is not available in a fork child: its memory is the
// process's, which the parent and the child would share (see the xs
// package's fork child). Each way the child's interpreter enters native
// code ends in a Perl error that names the code and the module, which
// eval catches, or, for a destructor, in nothing; the parent goes on using
// the module as before. Before, the child ran the native code, on the
// parent's native state.
func TestForkChildCannotUseNativeCode(t *testing.T) {
	so := buildPerlAPI(t)
	const refused = ": native modules cannot be used in a forked child"
	for _, r := range []forkRoute{
		{
			name:  "an XSUB",
			setup: `use PerlAPI; PerlAPI::sub_slot(7);`,
			child: `PerlAPI::sub_slot()`,
			after: `PerlAPI::sub_slot()`,
			want:  "died: PerlAPI::sub_slot" + refused + " (PerlAPI.so)\nstatus 0\nparent: 7",
		},
		{
			name:  "a module the child loads",
			setup: ``,
			child: `require PerlAPI; 1`,
			after: `require PerlAPI; PerlAPI::sub_slot(3)`,
			want:  "died: DynaLoader::native::PerlAPI::bootstrap" + refused + " (PerlAPI.so)\nstatus 0\nparent: 3",
		},
		{
			name:  "a library the child loads",
			setup: `require DynaLoader; our $so = '` + so + `';`,
			child: `DynaLoader::dl_load_file($main::so, 0) // "nothing: " . why(DynaLoader::dl_error())`,
			after: `DynaLoader::dl_load_file($main::so, 0) ? "loaded" : DynaLoader::dl_error()`,
			want:  "returned nothing: PerlAPI.so" + refused + "\nstatus 0\nparent: loaded",
		},
		{
			name:  "the destructor of a native object",
			setup: `use PerlAPI; { package Native; *DESTROY = \&PerlAPI::sub_slot } PerlAPI::sub_slot(7); my $obj = bless [], 'Native';`,
			child: `undef $obj; "after the destructor"`,
			after: `PerlAPI::sub_slot()`,
			want:  "returned after the destructor\nstatus 0\nparent: 7",
		},
		{
			name:  "native code a destructor calls",
			setup: `use PerlAPI; { package Holder; sub DESTROY { PerlAPI::sub_slot(9) } } PerlAPI::sub_slot(7); my $obj = bless [], 'Holder';`,
			child: `undef $obj; "after the destructor"`,
			after: `PerlAPI::sub_slot()`,
			want:  "returned after the destructor\nstatus 0\nparent: 7",
		},
		{
			name:  "magic of a native module freed",
			setup: `use PerlAPI; my @box = ("v"); PerlAPI::tally_attach($box[0], 10, "kept");`,
			child: `@box = (); "freed"`,
			after: `join "|", PerlAPI::tally_freed(), PerlAPI::tally_add($box[0], 5)`,
			want:  "returned freed\nstatus 0\nparent: 0|kept=15",
		},
		{
			name:  "set magic of a native module",
			setup: `use PerlAPI; our $v = "a"; PerlAPI::watch($v);`,
			child: `$v = "b"; "assigned"`,
			after: `$v = "c"; PerlAPI::watched()`,
			want:  "died: magic of a native module" + refused + "\nstatus 0\nparent: |c",
		},
		{
			name:  "local on a variable with magic of a native module",
			setup: `use PerlAPI; our $t = "v"; PerlAPI::tally_attach($t, 10, "kept");`,
			child: `local $t; "localized"`,
			after: `PerlAPI::tally_add($t, 1)`,
			want:  "died: local on a variable with magic of a native module" + refused + "\nstatus 0\nparent: kept=11",
		},
		{
			name:  "uvar magic of a native module",
			setup: `use PerlAPI; my %h; PerlAPI::uvar_hash(\%h); $h{real} = "value";`,
			child: `$h{alias} // "none"`,
			after: `$h{alias}`,
			want:  "died: magic of a native module" + refused + "\nstatus 0\nparent: value",
		},
		{
			name:  "an op of a native module",
			setup: `use PerlAPI; PerlAPI::lexical_keywords(); my $tagged = eval q{ sub { my $v = tag_op; $v } } or die $@;`,
			child: `$tagged->()`,
			after: `$tagged->()`,
			want:  "died: an op (custom) of a native module" + refused + " (PerlAPI.so)\nstatus 0\nparent: 4242",
		},
		{
			name:  "an op type a native module claimed",
			setup: `use PerlAPI; PerlAPI::mderef_watch(); my $f = eval q{ sub { our %h = (k => 7); $main::h{k} } } or die $@;`,
			child: `$f->()`,
			after: `$f->()`,
			want:  "died: an op (multideref) of a native module" + refused + " (PerlAPI.so)\nstatus 0\nparent: 7",
		},
		{
			name:  "a keyword plugin",
			setup: `use PerlAPI; PerlAPI::install_kw42();`,
			child: `eval q{ my $x = 2; $x } // die $@`,
			after: `eval q{ goperl_kw42 + 1 }`,
			want:  `died: the keyword plugin of a native module (compiling "my")` + refused + "\nstatus 0\nparent: 43",
		},
		{
			name:  "an op checker",
			setup: `use PerlAPI; PerlAPI::hooks("add", 1);`,
			child: `eval q{ my ($x, $y) = (1, 2); $x + $y } // die $@`,
			after: `eval q{ my ($x, $y) = (1, 2); $x + $y }`,
			want:  "died: the op checker of a native module" + refused + "\nstatus 0\nparent: 1003",
		},
		{
			name:  "a peephole optimizer",
			setup: `use PerlAPI; PerlAPI::hooks("peep", 1);`,
			child: `eval q{ my $x = 2; $x } // die $@`,
			after: `eval q{ my $x = 2; $x }`,
			want:  "died: the peephole optimizer of a native module" + refused + "\nstatus 0\nparent: 2",
		},
		{
			name:  "a call checker",
			setup: `use PerlAPI; sub answer { "computed" } PerlAPI::constant_calls(\&answer, "constant");`,
			child: `eval q{ answer() } // die $@`,
			after: `eval q{ answer() }`,
			want:  "died: the call checker of a native module" + refused + "\nstatus 0\nparent: constant",
		},
		{
			name:  "a block hook",
			setup: `use PerlAPI; PerlAPI::block_starts();`,
			child: `eval q{ { 1 } } // die $@`,
			after: `eval q{ { 5 } }`,
			want:  "died: the block hook of a native module" + refused + "\nstatus 0\nparent: 5",
		},
		{
			name:  "an infix operator",
			setup: `use PerlAPI; PerlAPI::enable_infix();`,
			child: `eval q{ ("a" equ "a") ? 1 : 0 } // die $@`,
			after: `eval q{ ("a" equ "a") ? 1 : 0 }`,
			want:  `died: the infix operator plugin of a native module` + refused + "\nstatus 0\nparent: 1",
		},
		{
			name:  "memory of native code",
			setup: `use PerlAPI; my $at = PerlAPI::native_string();`,
			child: `unpack("p", pack("J", $at))`,
			after: `unpack("p", pack("J", $at))`,
			want:  "died: memory of native code at 0x..." + refused + "\nstatus 0\nparent: native bytes",
		},
	} {
		t.Run(strings.ReplaceAll(r.name, " ", "_"), func(t *testing.T) {
			status, stdout, stderr := runForkingProgramWith(t, so, forkRouteProgram(r)...)
			want := "child: " + r.want + "\n"
			if status != 0 || stdout != want || stderr != "" {
				t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q and nothing on stderr", status, stdout, stderr, want)
			}
		})
	}
}

// A destructor a module left in a scope of the parent's (a save-stack
// destructor) does not run when the child leaves that scope: it would
// release what the parent's still uses. The parent's runs as before.
func TestForkChildLeavesTheScopeOfANativeDestructor(t *testing.T) {
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		`pipe(my $r, my $w) or die "pipe: $!";`,
		`my $pid;`,
		`{`,
		`    my $who = "parent";`,
		`    PerlAPI::scope_guard(sub { syswrite($w, "guard of the $who\n") if $who eq "child"; $main::guarded = $who });`,
		`    $pid = fork // die "fork: $!";`,
		`    $who = "child" if !$pid;`,
		`}`,
		`if (!$pid) { close $r; syswrite($w, "left the scope\n"); exit 4 }`,
		`close $w; my @lines = <$r>; waitpid($pid, 0);`,
		`print "child: ", @lines, "status ", $? >> 8, "\n", "guard in the parent: $main::guarded\n";`,
	)
	want := "child: left the scope\nstatus 4\nguard in the parent: parent\n"
	if status != 0 || stdout != want || stderr != "" {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q", status, stdout, stderr, want)
	}
}

// In the child's teardown native code is skipped, not refused: an END
// block the parent registered and a destructor that is native code run at
// the child's end without native code and without an error, and the child
// ends with the status it exits with. The parent's run as before.
func TestForkChildEndsWithoutNativeCode(t *testing.T) {
	status, stdout, stderr := runForkingProgram(t,
		`use strict; use warnings; no warnings 'once'; use PerlAPI; $| = 1;`,
		`{ package Native; *DESTROY = \&PerlAPI::sub_slot }`,
		`our $obj = bless [], 'Native';`,
		`PerlAPI::sub_slot(7);`,
		`our $parent = $$;`,
		`END { print "END in the parent: ", PerlAPI::sub_slot(), "\n" if $$ == $parent }`,
		`END { PerlAPI::sub_slot(8) if $$ != $parent }`,
		`my $pid = fork // die "fork: $!";`,
		`exit 3 if !$pid;`,
		`waitpid($pid, 0);`,
		`print "child status ", $? >> 8, "\n";`,
	)
	want := "child status 3\nEND in the parent: 7\n"
	if status != 0 || stdout != want || stderr != "" {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q", status, stdout, stderr, want)
	}
}

// A program that runs under a run loop of native code (PL_runops, as a
// profiler sets it) cannot run a fork child: the child dies at once with
// the error, the parent goes on.
func TestForkChildOfANativeRunLoop(t *testing.T) {
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		`PerlAPI::runloop(1);`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) { print "the child ran\n"; exit 0 }`,
		`waitpid($pid, 0);`,
		`print "child status ", $? >> 8, ", parent ", PerlAPI::sub_slot(5), "\n";`,
	)
	want := "child status 255, parent 5\n"
	if status != 0 || stdout != want || !strings.Contains(stderr, "the run loop of a native module: native modules cannot be used in a forked child") {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q and the error on stderr", status, stdout, stderr, want)
	}
}

// A module may fork the process itself with fork(2). The child of such a
// fork cannot run the interpreter (the Go runtime does not survive fork):
// a child that calls it ends at that call with a message and status 255; a
// child that execs runs as on perl. Before, the child entered the runtime
// in a copy of the process without its threads.
func TestNativeCodeForksTheProcess(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a child that calls the interpreter", `my ($status, $err) = PerlAPI::fork_and_call("call"); join ",", $status >> 8, $status & 127, $err =~ /^native code forked this process with fork\(2\), and its child called the interpreter: .* the child exits\n\z/s ? "message" : "[$err]"`, "255,0,message"},
		{"a child that execs", `my ($status, $err) = PerlAPI::fork_and_call("exec"); join ",", $status >> 8, $status & 127, "[$err]"`, "7,0,[]"},
		{"the parent goes on", `PerlAPI::sub_slot(4)`, "4"},
	})
}

// childWatchProgram reports, line by line into the file report, what a
// module's child watcher (PerlAPI::reap_children, run from the program's
// SIGCHLD handler, and PerlAPI::count_sigchld, a handler of the signal in
// native code) sees of the program's children: a fork child, a fork child
// that execs, a command the program started, and a child the module
// started itself.
func childWatchProgram(report string) []string {
	return []string{
		`use strict; use warnings; use PerlAPI; use POSIX (); $| = 1;`,
		`open my $log, '>', '` + report + `' or die "open: $!";`,
		`sub note { print $log @_, "\n" }`,
		`sub settle { my $done = shift; for (1 .. 1000) { return if $done->(); select undef, undef, undef, 0.01 } }`,
		`PerlAPI::count_sigchld();`,
		`my (@reaped, $handled);`,
		`$SIG{CHLD} = sub { $handled++; push @reaped, PerlAPI::reap_children() };`,
		`sub ended { my $pid = shift; join ",", map { /^$pid:(\d+)\z/ ? $1 : () } @reaped }`,
		`my $pid = fork // die "fork: $!";`,
		`POSIX::_exit(3) if !$pid;`,
		`settle(sub { ended($pid) ne "" });`,
		`select undef, undef, undef, 0.3;`,
		`note "fork child: ", ended($pid), " then ", $reaped[-1], ", handled ", $handled, ", native handler ", PerlAPI::count_sigchld();`,
		`$pid = fork // die "fork: $!";`,
		`if (!$pid) { exec('/bin/sh', '-c', 'kill -USR1 $$') or POSIX::_exit(127) }`,
		`settle(sub { ended($pid) ne "" });`,
		`note "fork child that execs: ", ended($pid);`,
		`$pid = open(my $from, '-|', '/bin/sh', '-c', 'exit 5') // die "open: $!";`,
		`settle(sub { ended($pid) ne "" });`,
		`note "command: ", ended($pid);`,
		`$pid = PerlAPI::spawn_exit(4);`,
		`my ($got, $status, $err) = split / /, PerlAPI::wait_child($pid, 0);`,
		`note "module's own child: ", ($got == $pid ? "found" : "got $got"), " $status $err";`,
		`note "no children: ", join(",", PerlAPI::reap_children()), " ", (split / /, PerlAPI::wait_child(-1, 1))[2];`,
		`PerlAPI::count_sigchld(0);`,
		`close $log or die "close: $!";`,
	}
}

// Native code's waitpid is the program's (the module's own waitpid, wait3
// and wait4 ask the interpreter): an event loop's child watcher, which
// reaps with waitpid(-1, WNOHANG) when SIGCHLD comes, finds the program's
// fork children, which run inside the process, and the commands it started,
// with their status words in the host's numbers, and then ECHILD. The end
// of a fork child reaches the program's own handler of SIGCHLD once; the
// module's handler, which the program's replaced (set after it), not at all
// on Linux, as in perl, where the last sigaction wins. Elsewhere the runtime
// does not know the native handler and leaves it the process's action, and
// the end reaches it once, as a real SIGCHLD. A child the module
// started itself is still its own to wait for. Before, waitpid was the C
// library's: it knew no fork child (ECHILD at once) and took the commands
// from under the runtime, and no SIGCHLD came for a fork child.
func TestNativeWaitpidIsTheProgramsOwn(t *testing.T) {
	so := buildPerlAPI(t)
	report := filepath.Join(t.TempDir(), "report")
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var out, errOut bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, Fork: true, HostProcess: true, NativeModules: true, Stdout: &out, Stderr: &errOut,
		Exec: func(string, []string) bool { return true }})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	loadXSModule(t, p, "PerlAPI", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "PerlAPI.pm"),
		[]byte("package PerlAPI;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := p.RunProgram(context.Background(), perl.Program{Code: childWatchProgram(report), Inc: []string{libdir}})
	if err != nil {
		p.Close()
		t.Fatalf("RunProgram: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	got, _ := os.ReadFile(report)
	nativeHandler := "1"
	if runtime.GOOS == "linux" {
		nativeHandler = "0"
	}
	want := strings.Join([]string{
		"fork child: 768 then ECHILD, handled 1, native handler " + nativeHandler,
		"fork child that execs: " + strconv.Itoa(int(syscall.SIGUSR1)),
		"command: 1280",
		"module's own child: found 1024 -",
		"no children: ECHILD ECHILD",
	}, "\n") + "\n"
	if status != 0 || string(got) != want {
		t.Fatalf("status %d, report:\n%s\nwant:\n%s\nstdout %q stderr %q", status, got, want, out.String(), errOut.String())
	}
}

// The program that is not the host process: its fork child is found by
// native code's waitpid(-1), and a child native code started itself (whose
// pid is the host's, which such a program does not share) is the C
// library's.
func TestNativeWaitpidOfAnIsolatedProgram(t *testing.T) {
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; use POSIX (); $| = 1;`,
		`my $pid = fork // die "fork: $!";`,
		`POSIX::_exit(6) if !$pid;`,
		`my @r; for (1 .. 1000) { @r = PerlAPI::reap_children(); last if grep /^$pid:/, @r; select undef, undef, undef, 0.01 }`,
		`my $n = PerlAPI::spawn_exit(2);`,
		`my ($got, $st) = split / /, PerlAPI::wait_child($n, 0);`,
		`print join(",", map { s/^$pid:/child:/r } @r), " ", ($got == $n ? "own" : "got $got"), " $st\n";`,
	)
	if want := "child:1536,ECHILD own 512\n"; status != 0 || stdout != want {
		t.Fatalf("status=%d stdout=%q stderr=%q, want %q", status, stdout, stderr, want)
	}
}

// my_exit ends the program from inside a native call with the status it is
// given, as `exit` does (END blocks run); my_failure_exit ends it with 255
// when neither $! nor $? says otherwise. Before, neither name existed in
// the SDK.
func TestNativeCodeEndsTheProgram(t *testing.T) {
	for _, tc := range []struct {
		name, call, out string
		status          int
	}{
		{"my_exit", `PerlAPI::exit_with(3)`, "end ran\n", 3},
		{"my_failure_exit", `$! = 0; $? = 0; PerlAPI::exit_failing()`, "end ran\n", 255},
	} {
		status, stdout, stderr := runForkingProgram(t,
			`use PerlAPI; $| = 1; END { print "end ran\n" }`,
			tc.call+`; print "not here\n";`,
		)
		if status != tc.status || stdout != tc.out {
			t.Errorf("%s: status=%d stdout=%q stderr=%q, want status=%d stdout=%q", tc.name, status, stdout, stderr, tc.status, tc.out)
		}
	}
}

// A function a module gave to call_atexit before a fork releases what the
// module keeps: it runs once, when the parent ends, and not when the child
// does.
func TestForkSharesAtexitFunctionsWithTheParent(t *testing.T) {
	report := filepath.Join(t.TempDir(), "report")
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		`PerlAPI::at_exit_write('`+report+`');`,
		`my $pid = fork // die "fork: $!";`,
		`exit 0 if !$pid;`,
		`waitpid($pid, 0);`,
		`print "child status ", $? >> 8, ", ", (-e '`+report+`' ? "run by the child" : "not run yet"), "\n";`,
	)
	want := "child status 0, not run yet\n"
	if status != 0 || stdout != want {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q", status, stdout, stderr, want)
	}
	got, err := os.ReadFile(report)
	if err != nil || string(got) != "the interpreter is gone\n" {
		t.Fatalf("report %q (%v), want the line of the parent's call_atexit function; stderr %q", got, err, stderr)
	}
}

// A Clone has the op checkers and keyword plugins native modules installed
// in the original and the ops of a module's own struct compiled there, and
// installing a checker in the clone leaves the original without it.
func TestCloneWithNativeHooksOfTheInterpreter(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `PerlAPI::hooks("add", 1); PerlAPI::lexical_keywords(); our $tagged = eval q{ sub { my $v = tag_op; $v } } or die $@; 1`); err != nil || r.Error != nil {
		t.Fatalf("prepare: err=%v error=%v", err, r.Error)
	}
	c, err := p.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	defer c.Close()
	runEvalCases(t, c, []evalCase{
		{"the op checker", `my ($x, $y) = (1, 2); $x + $y`, "1003"},
		{"the keyword plugin", `one_into_x; $x`, "1"},
		{"the op of a module's struct", `$tagged->()`, "4242"},
		{"one of the clone's own", `PerlAPI::hooks("join", 1); eval q{join "-", 1, 2}`, "1+2"},
	})
	runEvalCases(t, p, []evalCase{
		{"the original keeps its own", `my ($x, $y) = (1, 2); $x + $y`, "1003"},
		{"and its op of a module's struct", `$tagged->()`, "4242"},
		{"and not the clone's", `join "-", 1, 2`, "1-2"},
	})
}

// A Clone of an instance whose variables carry native magic has the magic
// too, and closing the clone leaves the original's intact.
func TestCloneWithNativeState(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `our @box = ("v"); PerlAPI::tally_attach($box[0], 10, "kept"); PerlAPI::sub_slot(7); 1`); err != nil || r.Error != nil {
		t.Fatalf("prepare: err=%v error=%v", err, r.Error)
	}
	c, err := p.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	runEvalCases(t, c, []evalCase{
		{"the clone has the magic", `PerlAPI::tally_add($box[0], 1)`, "kept=11"},
		{"and the slot", `PerlAPI::sub_slot()`, "7"},
		{"which is its own", `PerlAPI::sub_slot(9)`, "9"},
	})
	if err := c.Close(); err != nil {
		t.Fatalf("Close clone: %v", err)
	}
	runEvalCases(t, p, []evalCase{
		{"the original after the clone is gone", `join "|", PerlAPI::tally_freed(), PerlAPI::tally_add($box[0], 1), PerlAPI::sub_slot()`, "0|kept=12|7"},
		{"released by its last holder", `@box = (); PerlAPI::tally_freed()`, "1"},
	})
}
