package gperl_test

import (
	"debug/elf"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/goccy/go-perl/internal/gperl"
)

// crossEnv makes TestCrossBuild required: with it set, a missing Zig fails
// the test instead of skipping it (make test-cross).
const crossEnv = "GPERL_TEST_CROSS"

// crossCase is one build of a test project for another platform.
type crossCase struct {
	name     string
	goarch   string
	cgo      string
	triple   string // Zig's target for CC; none builds without a C compiler
	cpanfile string
	script   string
	output   string
	static   bool // the binary needs no dynamic loader
	linked   bool // the binary has the native modules linked in
}

// gperl build makes binaries for linux/amd64 and linux/arm64 from any host:
// with cgo the native modules are compiled by Zig for the target and
// linked in (into a fully static binary on musl), without cgo they are
// carried as files, and a project without native modules needs no C
// compiler at all. A binary for the platform the test runs on is run.
func TestCrossBuild(t *testing.T) {
	requireToolchain(t)
	zig := gperl.ManagedBin("zig")
	if zig == "" {
		if path, err := exec.LookPath("zig"); err == nil {
			zig = path
		}
	}
	if zig == "" {
		if os.Getenv(crossEnv) != "" {
			t.Fatalf("%s is set and no Zig was found: run `gperl install toolchain` or put zig on PATH", crossEnv)
		}
		t.Skipf("no Zig to cross compile with (set %s to require it)", crossEnv)
	}
	startMirror(t, testDists(t)...)
	t.Setenv("GPERL_DEV_REPLACE", devReplaces(t))

	const pureCPANfile = "requires 'Pure::Built';\n"
	const pureScript = "use strict;\nuse Pure::Built;\nprint Pure::Built::name(), \"\\n\";\n"
	var cases []crossCase
	for _, arch := range []struct{ goarch, cpu string }{{"amd64", "x86_64"}, {"arm64", "aarch64"}} {
		cases = append(cases,
			crossCase{name: "linux_" + arch.goarch + "_musl_static", goarch: arch.goarch, cgo: "1",
				triple: arch.cpu + "-linux-musl", cpanfile: testCPANfile, script: testScript, output: testOutput,
				static: true, linked: true},
			crossCase{name: "linux_" + arch.goarch + "_gnu", goarch: arch.goarch, cgo: "1",
				triple: arch.cpu + "-linux-gnu", cpanfile: testCPANfile, script: testScript, output: testOutput,
				linked: true},
			// Without cgo the binary loads the C library through the
			// dynamic loader (purego), so it has a program interpreter.
			crossCase{name: "linux_" + arch.goarch + "_nocgo", goarch: arch.goarch, cgo: "0",
				triple: arch.cpu + "-linux-gnu", cpanfile: testCPANfile, script: testScript, output: testOutput},
			crossCase{name: "linux_" + arch.goarch + "_pure_nocgo", goarch: arch.goarch, cgo: "0",
				cpanfile: pureCPANfile, script: pureScript, output: "built on static\n"},
		)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range map[string]string{"cpanfile": c.cpanfile, "app.pl": c.script} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("GOOS", "linux")
			t.Setenv("GOARCH", c.goarch)
			t.Setenv("CGO_ENABLED", c.cgo)
			if c.triple != "" {
				t.Setenv("CC", zig+" cc -target "+c.triple)
				t.Setenv("AR", zig+" ar")
			} else {
				// No C compiler of any kind: not the environment's, not an
				// installed one.
				t.Setenv("CC", "")
				t.Setenv("AR", "")
				t.Setenv("GPERLTOOLCHAIN", t.TempDir())
			}
			out := filepath.Join(t.TempDir(), "app")
			if err := gperl.Build(filepath.Join(dir, "app.pl"), out); err != nil {
				t.Fatalf("Build: %v", err)
			}
			inspectCrossBinary(t, out, c)
			if runtime.GOOS != "linux" || runtime.GOARCH != c.goarch {
				return
			}
			cmd := exec.Command(out)
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), "GPERLMODCACHE="+filepath.Join(t.TempDir(), "absent"))
			got, err := cmd.CombinedOutput()
			if err != nil || string(got) != c.output {
				t.Fatalf("binary printed %q (%v), want %q", got, err, c.output)
			}
		})
	}
}

// inspectCrossBinary verifies the binary at path is an ELF executable for
// the case's architecture, static where the case says so, and — for a
// build that links the native modules — carries every module's entry table
// with the modules' own symbols kept apart.
func inspectCrossBinary(t *testing.T, path string, c crossCase) {
	t.Helper()
	f, err := elf.Open(path)
	if err != nil {
		t.Fatalf("not an ELF executable: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()
	want := map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64}[c.goarch]
	if f.Machine != want {
		t.Errorf("machine = %v, want %v", f.Machine, want)
	}
	interp := slices.ContainsFunc(f.Progs, func(p *elf.Prog) bool { return p.Type == elf.PT_INTERP })
	if interp == c.static {
		t.Errorf("program interpreter present = %v, want %v", interp, !c.static)
	}
	if !c.linked {
		return
	}
	syms, err := f.Symbols()
	if err != nil {
		t.Fatalf("symbols: %v", err)
	}
	names := make([]string, 0, len(syms))
	for _, s := range syms {
		names = append(names, s.Name)
	}
	for _, sym := range []string{
		"goperl_entry_Demo__XS", "goperl_entry_Twin__A", "goperl_entry_Twin__B",
		// The two modules' helpers of the same name stay apart.
		"twin_helper..gperl.Twin__A", "twin_helper..gperl.Twin__B",
	} {
		if !slices.Contains(names, sym) {
			t.Errorf("no symbol %s", sym)
		}
	}
	if slices.Contains(names, "twin_helper") {
		t.Errorf("a module's helper kept its own name")
	}
}
