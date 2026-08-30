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
