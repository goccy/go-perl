// The tests run programs in the test process, and in processes started from
// the test binary as the perl command: like the gperl command, it leaves
// the lowest descriptor numbers to the program (see cmd/gperl).
//
//go:debug containermaxprocs=0
package gperl_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// writeScript writes a program for a test to run. A program that runs in
// the test process is that process while it runs (the process follows its
// working directory and its environment), so the test gives the process
// back as it found it.
func writeScript(t *testing.T, name, src string) string {
	t.Helper()
	keepProcessState(t)
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunSuccess(t *testing.T) {
	script := writeScript(t, "ok.pl", `my $x = 40 + 2; exit 0 if $x != 42; 1;`)
	status, err := runInThisProcess(t, script, nil)
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
}

// A program succeeds when it runs to its end, whatever its last statement
// evaluates to and whatever errno an earlier, handled failure left behind.
func TestRunLastStatementUndefAfterFailedSyscall(t *testing.T) {
	script := writeScript(t, "undef.pl", `open(my $fh, '<', '/nonexistent/file') and die "opened"; undef;`)
	status, err := runInThisProcess(t, script, nil)
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
}

func TestRunMissingScript(t *testing.T) {
	status, err := runInThisProcess(t, filepath.Join(t.TempDir(), "absent.pl"), nil)
	if err == nil || status == 0 {
		t.Fatalf("Run = (%d, %v), want a failure", status, err)
	}
}

func TestRunExitStatus(t *testing.T) {
	script := writeScript(t, "exit.pl", `exit 7;`)
	status, err := runInThisProcess(t, script, nil)
	if err != nil || status != 7 {
		t.Fatalf("Run = (%d, %v), want (7, nil)", status, err)
	}
}

func TestRunUncaughtDie(t *testing.T) {
	script := writeScript(t, "die.pl", `$! = 0; die "kaboom\n";`)
	status, err := runInThisProcess(t, script, nil)
	// perl reports the die on stderr itself; the run only carries its
	// status.
	if status != 255 || err != nil {
		t.Fatalf("Run = (%d, %v), want (255, nil)", status, err)
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
		target := m[2]
		// A relative replace target is relative to go-perl's go.mod; the
		// generated program lives elsewhere, so absolutize it.
		if strings.HasPrefix(target, ".") {
			target = filepath.Join(root, target)
		}
		pairs = append(pairs, m[1]+"="+target)
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

// keepProcessState restores the working directory and the environment of
// the test process when the test is over.
func keepProcessState(t *testing.T) {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	env := os.Environ()
	t.Cleanup(func() {
		if err := os.Chdir(dir); err != nil {
			t.Errorf("back to the working directory of the test: %v", err)
		}
		os.Clearenv()
		for _, kv := range env {
			if name, value, ok := strings.Cut(kv, "="); ok {
				os.Setenv(name, value)
			}
		}
	})
}
