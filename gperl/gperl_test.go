package gperl_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/gperl"
)

func writeScript(t *testing.T, name, src string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunSuccess(t *testing.T) {
	script := writeScript(t, "ok.pl", `my $x = 40 + 2; exit 0 if $x != 42; 1;`)
	status, err := gperl.Run(script, nil)
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
}

func TestRunExitStatus(t *testing.T) {
	script := writeScript(t, "exit.pl", `exit 7;`)
	status, err := gperl.Run(script, nil)
	if err != nil || status != 7 {
		t.Fatalf("Run = (%d, %v), want (7, nil)", status, err)
	}
}

func TestRunUncaughtDie(t *testing.T) {
	script := writeScript(t, "die.pl", `die "kaboom\n";`)
	status, err := gperl.Run(script, nil)
	if status != 255 {
		t.Fatalf("status = %d, want 255", status)
	}
	var pe *perl.PerlError
	if !errors.As(err, &pe) || !strings.Contains(pe.Message, "kaboom") {
		t.Fatalf("err = %v, want *perl.PerlError containing kaboom", err)
	}
}

// devReplaces reproduces this checkout's module wiring for a generated
// program: go-perl itself plus whatever replace directives go-perl's go.mod
// carries (the unreleased perlwasm2go bundle during development).
func devReplaces(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	pairs := []string{"github.com/goccy/go-perl=" + root}
	gomod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(?m)^replace\s+(\S+)\s*=>\s*(\S+)`)
	for _, m := range re.FindAllStringSubmatch(string(gomod), -1) {
		pairs = append(pairs, m[1]+"="+m[2])
	}
	return strings.Join(pairs, ",")
}

func TestBuildProducesStandaloneBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a full binary; skipped in -short")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "greet.pl")
	if err := os.WriteFile(script, []byte(`
		use strict;
		use myapp::greeting;
		print myapp::greeting::text($ARGV[0]), "\n";
		exit 0;
	`), 0o644); err != nil {
		t.Fatal(err)
	}
	// A module in the project's lib/ proves the embedded tree lands on @INC.
	if err := os.MkdirAll(filepath.Join(dir, "lib", "myapp"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "lib", "myapp", "greeting.pm"), []byte(`
		package myapp::greeting;
		sub text { "greetings, " . ($_[0] // "world") }
		1;
	`), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "greet-bin")
	t.Setenv("GPERL_DEV_REPLACE", devReplaces(t))
	if err := gperl.Build(script, out); err != nil {
		t.Fatalf("Build: %v", err)
	}

	got, err := exec.Command(out, "gopher").Output()
	if err != nil {
		t.Fatalf("run built binary: %v", err)
	}
	if want := "greetings, gopher\n"; string(got) != want {
		t.Fatalf("binary output = %q, want %q", got, want)
	}
}
