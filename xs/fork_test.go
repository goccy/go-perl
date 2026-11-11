//go:build darwin || linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

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
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, Fork: true, Stdout: &out, Stderr: &errOut})
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

// The parent of a fork ends first, as a program that puts itself in the
// background does: the child goes on using the native state it was forked
// with.
func TestForkParentEndsFirstWithNativeState(t *testing.T) {
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
		`@box = (); close $w; close $r;`,
		`print "parent released: ", PerlAPI::tally_freed(), "\n";`,
	)
	if status != 0 || stdout != "parent released: 0\n" {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	// RunProgram and Close return once the instance's children have ended.
	got, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("the child left no report: %v (stderr %q)", err, stderr)
	}
	if string(got) != "kept=11\n" {
		t.Fatalf("child saw %q, want %q", got, "kept=11\n")
	}
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
