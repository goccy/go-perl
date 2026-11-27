// The tests run programs in the test process, and in processes started from
// the test binary as the perl command: like the gperl command, it leaves
// the lowest descriptor numbers to the program (see cmd/gperl).
//
//go:debug containermaxprocs=0
package gperl_test

import (
	"debug/buildinfo"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/goccy/go-perl/internal/gperl"
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
func devReplaces(t testing.TB) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
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
	// Without native modules a build for this machine has cgo when it has
	// a C compiler.
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no cc available")
	}
	t.Setenv("CC", cc)
	setGoEnvOtherThan(t, "1")
	if err := gperl.Build(script, out); err != nil {
		t.Fatalf("Build: %v", err)
	}
	wantBuildSettings(t, out, runtime.GOOS, runtime.GOARCH, "1")

	got, err := exec.Command(out, "gopher").Output()
	if err != nil {
		t.Fatalf("run built binary: %v", err)
	}
	if want := "greetings, gopher\n"; string(got) != want {
		t.Fatalf("binary output = %q, want %q", got, want)
	}
}

// GPERLOS and GPERLARCH name the platform gperl build makes a binary for.
func TestBuildForGPERLOSAndGPERLARCH(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a full binary; skipped in -short")
	}
	goarch := "arm64"
	if runtime.GOOS == "linux" && runtime.GOARCH == goarch {
		goarch = "amd64"
	}
	script := filepath.Join(t.TempDir(), "hello.pl")
	if err := os.WriteFile(script, []byte("print \"hello\\n\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GPERL_DEV_REPLACE", devReplaces(t))
	t.Setenv("GPERLOS", "linux")
	t.Setenv("GPERLARCH", goarch)
	setGoEnvOtherThan(t, "0")
	out := filepath.Join(t.TempDir(), "hello")
	if err := gperl.Build(script, out); err != nil {
		t.Fatalf("Build: %v", err)
	}
	wantBuildSettings(t, out, "linux", goarch, "0")
}

// A platform Go does not build for, and a GPERLNATIVE gperl does not know,
// are errors that name them.
func TestBuildConfigErrors(t *testing.T) {
	script := filepath.Join(t.TempDir(), "hello.pl")
	if err := os.WriteFile(script, []byte("print \"hello\\n\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GPERLMODCACHE", filepath.Join(t.TempDir(), "mod"))
	for _, c := range []struct {
		goos, goarch, native string
		want                 []string
	}{
		{native: "static", want: []string{"GPERLNATIVE=static", "link", "load"}},
		{goos: "windows", goarch: "s390x", want: []string{"GPERLOS=windows GPERLARCH=s390x"}},
		{goos: "nosuchos", want: []string{"GPERLOS=nosuchos GPERLARCH=" + runtime.GOARCH}},
	} {
		t.Setenv("GPERLOS", c.goos)
		t.Setenv("GPERLARCH", c.goarch)
		t.Setenv("GPERLNATIVE", c.native)
		for name, err := range map[string]error{
			"Build":       gperl.Build(script, filepath.Join(t.TempDir(), "hello")),
			"ModDownload": gperl.ModDownload(t.TempDir()),
		} {
			if err == nil {
				t.Errorf("%s with %+v succeeded", name, c)
				continue
			}
			for _, want := range c.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("%s with %+v: error %q does not name %q", name, c, err, want)
				}
			}
		}
	}
}

// setGoEnvOtherThan sets Go's own variables to what no build gperl runs
// may use: a platform Go has no port for, and cgo other than cgo ("0" or
// "1").
func setGoEnvOtherThan(t *testing.T, cgo string) {
	t.Helper()
	t.Setenv("GOOS", "plan9")
	t.Setenv("GOARCH", "mips")
	t.Setenv("CGO_ENABLED", map[string]string{"0": "1", "1": "0"}[cgo])
}

// wantBuildSettings verifies the platform and cgo setting the binary at
// path was built with.
func wantBuildSettings(t *testing.T, path, goos, goarch, cgo string) {
	t.Helper()
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": cgo}
	for _, s := range info.Settings {
		if w, ok := want[s.Key]; ok {
			if s.Value != w {
				t.Errorf("built with %s=%s, want %s", s.Key, s.Value, w)
			}
			delete(want, s.Key)
		}
	}
	for key := range want {
		t.Errorf("the binary records no %s", key)
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
