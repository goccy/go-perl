package perl_test

import (
	"bytes"
	"context"
	iofs "io/fs"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// faultingFS fails the way a crash does — the call into it never returns —
// when the file named fault is opened.
type faultingFS struct {
	perl.FS
	fault string
}

func (f faultingFS) OpenFile(name string, flag int, perm iofs.FileMode) (perl.File, error) {
	if name == f.fault {
		panic("the filesystem gave way under " + name)
	}
	return f.FS.OpenFile(name, flag, perm)
}

// A program whose interpreter crashes has not succeeded, and nothing of it
// runs afterwards: a crashed process runs no END blocks and no destructors.
func TestCrashEndsTheRunAbnormally(t *testing.T) {
	cfg := hostConfig(t)
	cfg.FS = faultingFS{FS: cfg.FS, fault: "no/way/back"}
	var stdout, stderr bytes.Buffer
	cfg.Stdin, cfg.Stdout, cfg.Stderr = strings.NewReader(""), &stdout, &stderr
	p, err := perl.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	status, err := p.RunProgram(context.Background(), perl.Program{Code: code(
		`$| = 1;`,
		`package Guard { sub new { bless {}, shift } sub DESTROY { print "destructor ran\n" } }`,
		`our $guard = Guard->new;`,
		`END { print "END block ran\n" }`,
		`print "before the crash\n";`,
		`open(my $fh, "<", "/no/way/back");`,
		`print "after the crash\n";`,
	)})
	if err == nil || status == 0 {
		t.Fatalf("RunProgram = (%d, %v), want a failure status and an error", status, err)
	}
	if got, want := stdout.String(), "before the crash\n"; got != want {
		t.Fatalf("the program printed %q, want %q", got, want)
	}
}
