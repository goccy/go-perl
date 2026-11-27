package gperl_test

import (
	"debug/elf"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
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
	native   string // GPERLNATIVE
	triple   string // Zig's target for CC; none builds without a C compiler
	cpanfile string
	script   string
	output   string
	static   bool // the binary needs no dynamic loader
	linked   bool // the binary has the native modules linked in
}

// idsScript also changes the effective uid of the process, which a binary
// with native modules linked in does through the C library linked into it
// (and not one the dynamic loader would load beside it): root takes
// another uid and goes back, anyone else is refused uid 0 (EPERM).
const (
	idsScript = testScript + `my $other = $< == 0 ? 65534 : 0;
$! = 0;
$> = $other;
my $answer = $< == 0 ? ($> == $other ? 'answered' : "not changed: $!") : ($!{EPERM} ? 'answered' : "not refused: $!");
$> = $<;
print "ids $answer\n";
`
	idsOutput = testOutput + "ids answered\n"
)

// gperl build makes binaries for linux/amd64 and linux/arm64 from any host:
// the native modules are compiled by Zig for the target and linked in
// (into a fully static binary on musl), with GPERLNATIVE=load they are
// carried as files, and a project without native modules needs no C
// compiler for another platform. Go's own variables in the environment
// play no part. A binary for the platform the test runs on is run.
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
			crossCase{name: "linux_" + arch.goarch + "_musl_static", goarch: arch.goarch,
				triple: arch.cpu + "-linux-musl", cpanfile: testCPANfile, script: idsScript, output: idsOutput,
				static: true, linked: true},
			crossCase{name: "linux_" + arch.goarch + "_gnu", goarch: arch.goarch, native: "link",
				triple: arch.cpu + "-linux-gnu", cpanfile: testCPANfile, script: idsScript, output: idsOutput,
				linked: true},
			// Without cgo the binary loads the C library through the
			// dynamic loader (purego), so it has a program interpreter.
			crossCase{name: "linux_" + arch.goarch + "_load", goarch: arch.goarch, native: "load",
				triple: arch.cpu + "-linux-gnu", cpanfile: testCPANfile, script: testScript, output: testOutput},
			crossCase{name: "linux_" + arch.goarch + "_pure", goarch: arch.goarch,
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
			t.Setenv("GPERLOS", "linux")
			t.Setenv("GPERLARCH", c.goarch)
			t.Setenv("GPERLNATIVE", c.native)
			// Without native modules a build for this machine has cgo
			// with the C compiler on PATH (requireToolchain).
			cgo := "0"
			if c.linked || (c.native != "load" && runtime.GOOS == "linux" && runtime.GOARCH == c.goarch) {
				cgo = "1"
			}
			setGoEnvOtherThan(t, cgo)
			if c.triple != "" {
				t.Setenv("CC", zig+" cc -target "+c.triple)
				t.Setenv("AR", zig+" ar")
			} else {
				// No C compiler but the one on PATH: not the environment's,
				// not an installed one.
				t.Setenv("CC", "")
				t.Setenv("AR", "")
				t.Setenv("GPERLTOOLCHAIN", t.TempDir())
			}
			out := filepath.Join(t.TempDir(), "app")
			if err := gperl.Build(filepath.Join(dir, "app.pl"), out); err != nil {
				t.Fatalf("Build: %v", err)
			}
			inspectCrossBinary(t, out, c)
			wantBuildSettings(t, out, "linux", c.goarch, cgo)
			// A C library's native modules have a directory of their own.
			if c.triple != "" {
				native := "linux_" + c.goarch
				if strings.HasSuffix(c.triple, "-musl") {
					native += "_musl"
				}
				archives, err := filepath.Glob(filepath.Join(os.Getenv("GPERLMODCACHE"), "Demo-XS@0.01", "gperl", "abi*", native, "auto", "Demo", "XS", "XS.a"))
				if err != nil || len(archives) != 1 {
					t.Errorf("Demo::XS built for %s: %v (%v)", native, archives, err)
				}
			}
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

	// Native modules built for another platform before, and no C compiler
	// for it now: linking them in fails with what is missing.
	t.Run("link_without_cross_compiler", func(t *testing.T) {
		goarch, cpu := "arm64", "aarch64"
		if runtime.GOOS == "linux" && runtime.GOARCH == goarch {
			goarch, cpu = "amd64", "x86_64"
		}
		dir := t.TempDir()
		for name, content := range map[string]string{"cpanfile": testCPANfile, "app.pl": testScript} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		t.Setenv("GPERLOS", "linux")
		t.Setenv("GPERLARCH", goarch)
		t.Setenv("GPERLNATIVE", "link")
		t.Setenv("CC", zig+" cc -target "+cpu+"-linux-gnu")
		t.Setenv("AR", zig+" ar")
		if err := gperl.ModDownload(dir); err != nil {
			t.Fatalf("ModDownload: %v", err)
		}
		t.Setenv("CC", "")
		t.Setenv("AR", "")
		t.Setenv("GPERLTOOLCHAIN", t.TempDir())
		err := gperl.Build(filepath.Join(dir, "app.pl"), filepath.Join(t.TempDir(), "app"))
		if err == nil || !strings.Contains(err.Error(), "link the native modules in: ") || !strings.Contains(err.Error(), "needs a C cross compiler") {
			t.Fatalf("Build = %v, want the missing C cross compiler", err)
		}
	})
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
