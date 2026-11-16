//go:build darwin || linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// runForkingProgram runs code as a program that may fork, on an instance
// with the PerlAPI fixture available to `use`.
func runForkingProgram(t *testing.T, code ...string) (status int, stdout, stderr string) {
	t.Helper()
	so := buildPerlAPI(t)
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

// A fork child has its memory in a place of its own, and pack 'P' there
// makes the address of the child's string: native code writes into the
// child's copy, and the parent's string stays as it was.
func TestForkPackedAddressesAreTheChilds(t *testing.T) {
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		`my $buf = "\0" x 4;`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) { PerlAPI::write_at(unpack("J", pack("P", $buf)), "kid"); exit(substr($buf, 0, 3) eq "kid" ? 7 : 1) }`,
		`waitpid $pid, 0;`,
		`print "child ", $? >> 8, " parent ", ($buf eq "\0" x 4 ? "untouched" : "changed"), "\n";`,
	)
	if status != 0 || stdout != "child 7 parent untouched\n" {
		t.Fatalf("status %d, stdout %q, stderr %q; want child 7 parent untouched", status, stdout, stderr)
	}
}

// A program forks while a variable carries magic of a native module and an
// XSUB a value in its own slot. The fork succeeds and the child has both.
// The slot is the child's own from then on. The C memory behind the magic
// is the module's and cannot be forked: the child leaving, and with it its
// copy of the variable, does not release what the parent still uses, while
// the magic the child attached itself is released when the child ends.
func TestForkWithNativeState(t *testing.T) {
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		`my @box = ("v"); PerlAPI::tally_attach($box[0], 10, "kept"); PerlAPI::sub_slot(7);`,
		`pipe(my $r, my $w) or die "pipe: $!";`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) {`,
		`    close $r;`,
		`    my $seen = PerlAPI::sub_slot(); PerlAPI::sub_slot(8);`,
		`    my $y = "w"; PerlAPI::tally_attach($y, 1, "made in the child");`,
		`    syswrite($w, join("|", PerlAPI::tally_add($box[0], 0), $seen, PerlAPI::sub_slot(), PerlAPI::tally_add($y, 1)) . "\n");`,
		`    exit 3;`,
		`}`,
		`close $w; my $line = <$r>; waitpid($pid, 0);`,
		`print "child: $line", "status ", $? >> 8, "\n";`,
		`print "parent: ", join("|", PerlAPI::tally_freed(), PerlAPI::tally_add($box[0], 5), PerlAPI::sub_slot()), "\n";`,
		`@box = (); print "released: ", PerlAPI::tally_freed(), "\n";`,
	)
	want := "child: kept=10|7|8|made in the child=2\nstatus 3\nparent: 1|kept=15|7\nreleased: 2\n"
	if status != 0 || stdout != want {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q", status, stdout, stderr, want)
	}
}

// The parent of a fork lets go of the native state it shares with its
// child: the child goes on using the native state it was forked with. The
// parent looks at what was freed while the child still waits for it: a
// child that has ended has released its copy too, and that release frees
// the state.
func TestForkChildKeepsNativeStateTheParentReleased(t *testing.T) {
	dir := t.TempDir()
	report := filepath.Join(dir, "report")
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI;`,
		`our $report = '`+report+`';`,
		`my @box = ("v"); PerlAPI::tally_attach($box[0], 10, "kept");`,
		`pipe(my $r, my $w) or die "pipe: $!";`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) {`,
		`    close $w; my $gone = <$r>;`,
		`    open my $fh, ">", $report or die $!;`,
		`    print $fh PerlAPI::tally_add($box[0], 1), "\n"; close $fh;`,
		`    exit 0;`,
		`}`,
		`@box = ();`,
		`print "parent released: ", PerlAPI::tally_freed(), "\n";`,
		`close $w; close $r;`,
		// The end of the program would be the end of the child.
		`waitpid($pid, 0);`,
	)
	if status != 0 || stdout != "parent released: 0\n" {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	got, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("the child left no report: %v (stderr %q)", err, stderr)
	}
	if string(got) != "kept=11\n" {
		t.Fatalf("child saw %q, want %q", got, "kept=11\n")
	}
}

// The program ends while a fork child is inside native code that does not
// return: it makes a read that the end's signal interrupts again, as C
// libraries do, so the host cannot end the call. The end of the program
// does not wait for it: the child is ended all the same, and runs no Perl
// code once its native call does return, and its goroutine goes then.
func TestProgramEndsWhileAChildIsInNativeCode(t *testing.T) {
	so := buildPerlAPI(t)
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	// What the native code reads: nothing is written until the program has
	// ended.
	held, release, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	defer release.Close()
	dir := t.TempDir()
	report := filepath.Join(dir, "report")
	if err := os.WriteFile(filepath.Join(dir, "PerlAPI.pm"),
		[]byte("package PerlAPI;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, Fork: true, Stdout: &out, Stderr: &errOut})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	loadXSModule(t, p, "PerlAPI", so)

	type outcome struct {
		status int
		err    error
	}
	ended := make(chan outcome, 1)
	go func() {
		status, err := p.RunProgram(context.Background(), perl.Program{
			Code: []string{
				`use PerlAPI;`,
				`pipe(my $r, my $w) or die "pipe: $!";`,
				`my $pid = fork // die "fork: $!";`,
				`if (!$pid) {`,
				`    close $r; syswrite($w, "in\n"); close $w;`,
				`    PerlAPI::read_host_descriptor($ARGV[0]);`,
				`    open my $fh, ">", $ARGV[1] or die $!; print $fh "the child went on\n"; close $fh;`,
				`    exit 0;`,
				`}`,
				`close $w; my $in = <$r>; select(undef, undef, undef, 0.3);`,
				`exit 5;`,
			},
			Args: []string{strconv.Itoa(int(held.Fd())), report},
			Inc:  []string{dir},
		})
		if err == nil {
			err = p.Close()
		}
		ended <- outcome{status, err}
	}()
	select {
	case got := <-ended:
		if got.err != nil || got.status != 5 {
			t.Fatalf("the program ended with (%d, %v), want (5, nil); stderr %q", got.status, got.err, errOut.String())
		}
	case <-time.After(30 * time.Second):
		release.Close()
		t.Fatal("the program's end waits for the child's native call")
	}
	// The native call returns now; the child must not get any further.
	before := runtime.NumGoroutine()
	if _, err := release.Write([]byte{1}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for runtime.NumGoroutine() >= before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if runtime.NumGoroutine() >= before {
		t.Errorf("the child's goroutine is still there after its native call returned")
	}
	if got, err := os.ReadFile(report); err == nil {
		t.Fatalf("the child ran on after the program had ended: %q", got)
	}
}

// Magic can point into the interpreter's own memory (DBI keeps the address
// of a handle's structure, the string buffer of an SV, that way). After a
// fork the child's magic points at the child's copy of that memory: what
// the child writes there it reads back, and the parent's is untouched.
func TestForkMagicPointsIntoTheChildsMemory(t *testing.T) {
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		// The first substr gives the string a buffer of its own, which the
		// later one then changes in place.
		`my $buffer = "Parent text"; substr($buffer, 0, 1, "p");`,
		`my $holder = "h"; PerlAPI::view_attach($holder, $buffer);`,
		`pipe(my $r, my $w) or die "pipe: $!";`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) {`,
		`    close $r; substr($buffer, 0, 6, "child ");`,
		`    syswrite($w, PerlAPI::view_read($holder, 11) . "\n");`,
		`    exit 0;`,
		`}`,
		`close $w; my $line = <$r>; waitpid($pid, 0);`,
		`print "child: $line", "parent: ", PerlAPI::view_read($holder, 11), "\n";`,
	)
	want := "child: child  text\nparent: parent text\n"
	if status != 0 || stdout != want {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q", status, stdout, stderr, want)
	}
}

// A fork child has what native modules registered with the interpreter
// before the fork: it runs an op checker and a keyword plugin the parent
// installed when it compiles code, an op of a module's own struct the
// parent compiled, and the destructor a module left in a scope of the
// parent's when it leaves that scope; and a function it gives to
// call_atexit runs when its interpreter is destroyed. The interpreter
// calls each of them through the host.
func TestForkChildRunsNativeHooksOfTheInterpreter(t *testing.T) {
	report := filepath.Join(t.TempDir(), "report")
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		`PerlAPI::hooks("add", 1); PerlAPI::lexical_keywords();`,
		`my $tagged = eval q{ sub { my $v = tag_op; $v } } or die $@;`,
		`pipe(my $r, my $w) or die "pipe: $!";`,
		`my $pid;`,
		`{`,
		`    my $who = "parent";`,
		`    PerlAPI::scope_guard(sub { syswrite($w, "guard of the $who\n") if $who eq "child"; $main::guarded = $who });`,
		`    $pid = fork // die "fork: $!";`,
		`    $who = "child" if !$pid;`,
		`}`,
		`if (!$pid) {`,
		`    close $r;`,
		`    syswrite($w, join("|", eval(q{my ($x, $y) = (1, 2); $x + $y}) // "error: $@", eval(q{one_into_x; $x}) // "error: $@", $tagged->()) . "\n");`,
		`    PerlAPI::at_exit_write('`+report+`');`,
		`    exit 4;`,
		`}`,
		`close $w; my @lines = <$r>; waitpid($pid, 0);`,
		`print "child: ", @lines, "status ", $? >> 8, "\n", "guard in the parent: $main::guarded\n";`,
	)
	want := "child: guard of the child\n1003|1|4242\nstatus 4\nguard in the parent: parent\n"
	if status != 0 || stdout != want {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q", status, stdout, stderr, want)
	}
	got, err := os.ReadFile(report)
	if err != nil || string(got) != "the interpreter is gone\n" {
		t.Fatalf("report %q (%v), want the line of the child's call_atexit function; stderr %q", got, err, stderr)
	}
}

// A function a module gave to call_atexit before a fork releases what the
// module keeps, which the parent and the child share: it runs once, when
// the last of them ends. A child that ends first leaves it to the parent.
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

// The process id native code asks for is the id of the process the program
// is: in a fork child, the child's, as $$ says.
func TestForkChildProcessIDInNativeCode(t *testing.T) {
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		`sub ids { join ",", map { $_ == $$ ? "self" : "other" } PerlAPI::process_id() }`,
		`pipe(my $r, my $w) or die "pipe: $!";`,
		`my $before = ids();`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) { close $r; syswrite($w, ids() . "\n"); exit 0 }`,
		`close $w; my $line = <$r>; waitpid($pid, 0);`,
		`print "before: $before\nchild: $line", "parent: ", ids(), "\n";`,
	)
	want := "before: self,self\nchild: self,self\nparent: self,self\n"
	if status != 0 || stdout != want {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q", status, stdout, stderr, want)
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

// A forked process calls native code through the host-call channel as its
// parent does: a function a module gives to call_atexit in the child runs
// when the child's interpreter is destroyed, and the child ends with its
// own status.
func TestForkChildCallsNativeCodeThroughTheHostChannel(t *testing.T) {
	file := filepath.Join(t.TempDir(), "at-exit.txt")
	status, stdout, stderr := runForkingProgram(t,
		`use PerlAPI; $| = 1;`,
		`my $pid = fork // die "fork: $!";`,
		`if (!$pid) { PerlAPI::at_exit_write('`+file+`'); exit 3 }`,
		`waitpid($pid, 0);`,
		`print "status ", $? >> 8, "\n";`,
	)
	if status != 0 || stdout != "status 3\n" || stderr != "" {
		t.Fatalf("status=%d stdout=%q stderr=%q, want status=0 stdout=%q", status, stdout, stderr, "status 3\n")
	}
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("the child's function did not run: %v", err)
	}
	if string(got) != "the interpreter is gone\n" {
		t.Fatalf("got %q", got)
	}
}
