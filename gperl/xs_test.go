package gperl

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hostXSToolchain reports whether the machine can drive an ordinary XS
// build (host perl + make + cc), skipping otherwise.
func hostXSToolchain(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"perl", "make", "cc"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("no %s on PATH; XS build unavailable", tool)
		}
	}
}

// TestXSBuildAndRun drives the whole pipeline on the bundled minimal EUMM
// distribution: gperl xs build (Makefile.PL + make against the embedded
// SDK), artifact layout under local/, and gperl run loading the module for
// a stock `use Demo::XS;`.
func TestXSBuildAndRun(t *testing.T) {
	hostXSToolchain(t)
	project := t.TempDir()

	dist, err := filepath.Abs(filepath.Join("testdata", "Demo-XS-dist"))
	if err != nil {
		t.Fatal(err)
	}
	modules, err := XSBuild(project, []string{dist})
	if err != nil {
		t.Fatalf("XSBuild: %v", err)
	}
	if len(modules) != 1 || modules[0] != "Demo::XS" {
		t.Fatalf("built modules = %v, want [Demo::XS]", modules)
	}
	so := filepath.Join(xsDir(project), "Demo-XS.so")
	if _, err := os.Stat(so); err != nil {
		t.Fatalf("native library missing: %v", err)
	}
	pm := filepath.Join(project, "local", "lib", "perl5", "Demo", "XS.pm")
	if _, err := os.Stat(pm); err != nil {
		t.Fatalf("pure-Perl half missing: %v", err)
	}

	script := filepath.Join(project, "app.pl")
	if err := os.WriteFile(script, []byte(`use strict;
use warnings;
use Demo::XS;
print "sum=", Demo::XS::add(40, 2), " ", Demo::XS::greet("gopher"), "\n";
`), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout := captureStdout(t, func() {
		status, err := Run(script, nil)
		if err != nil || status != 0 {
			t.Fatalf("Run: status=%d err=%v", status, err)
		}
	})
	if want := "sum=42 hello, gopher\n"; stdout != want {
		t.Fatalf("script output = %q, want %q", stdout, want)
	}
}

// TestXSBuildFromTarball covers the .tar.gz input path.
func TestXSBuildFromTarball(t *testing.T) {
	hostXSToolchain(t)
	project := t.TempDir()

	dist, err := filepath.Abs(filepath.Join("testdata", "Demo-XS-dist"))
	if err != nil {
		t.Fatal(err)
	}
	tarball := filepath.Join(t.TempDir(), "Demo-XS-0.01.tar.gz")
	cmd := exec.Command("tar", "-czf", tarball, "-C", filepath.Dir(dist), filepath.Base(dist))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("tar: %v\n%s", err, out)
	}

	modules, err := XSBuild(project, []string{tarball})
	if err != nil {
		t.Fatalf("XSBuild(tarball): %v", err)
	}
	if len(modules) != 1 || modules[0] != "Demo::XS" {
		t.Fatalf("built modules = %v, want [Demo::XS]", modules)
	}
}

// captureStdout redirects the process stdout around fn (Run wires the
// interpreter to os.Stdout).
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			sb.Write(buf[:n])
			if err != nil {
				break
			}
		}
		done <- sb.String()
	}()
	fn()
	w.Close()
	os.Stdout = old
	return <-done
}

// A native module laid out the way perl installs one — Module.pm next to
// auto/Module/Module.so on @INC, as in a dist's blib or an install tree —
// loads through the stock XSLoader, with nothing registered up front.
func TestNativeModuleLoadsFromInc(t *testing.T) {
	hostXSToolchain(t)
	project := t.TempDir()
	dist, err := filepath.Abs(filepath.Join("testdata", "Demo-XS-dist"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := XSBuild(project, []string{dist}); err != nil {
		t.Fatalf("XSBuild: %v", err)
	}

	lib := t.TempDir()
	place := func(src string, dst ...string) {
		t.Helper()
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(append([]string{lib}, dst...)...)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, data, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	place(filepath.Join(project, "local", "lib", "perl5", "Demo", "XS.pm"), "Demo", "XS.pm")
	place(filepath.Join(xsDir(project), "Demo-XS.so"), "auto", "Demo", "XS", "XS.so")

	status, stdout, stderr := runCLI(t, t.TempDir(), "", "-I"+lib, "-MDemo::XS", "-e",
		`print "sum=", Demo::XS::add(40, 2), "\n";`,
		`-e`, `my $pid = fork // die "fork: $!"; if (!$pid) { print "child=", Demo::XS::add(1, 2), "\n"; exit 0 } waitpid($pid, 0);`)
	if status != 0 || stdout != "sum=42\nchild=3\n" {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}

	// A module whose shared object is not there fails the way it does on
	// any perl.
	if err := os.Remove(filepath.Join(lib, "auto", "Demo", "XS", "XS.so")); err != nil {
		t.Fatal(err)
	}
	status, _, stderr = runCLI(t, t.TempDir(), "", "-I"+lib, "-MDemo::XS", "-e", "1")
	if status == 0 || !strings.Contains(stderr, "Can't locate loadable object for module Demo::XS in @INC") {
		t.Fatalf("status=%d stderr=%q, want DynaLoader's own failure", status, stderr)
	}
}
