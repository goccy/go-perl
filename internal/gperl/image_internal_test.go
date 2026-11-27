package gperl

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A static binary needs no loader, a position-independent one does, one
// that needs a library beyond the C library's is refused, and what is not
// an ELF binary is an error.
func TestNeedsLoader(t *testing.T) {
	if testing.Short() {
		t.Skip("builds binaries; skipped in -short")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go available")
	}
	// build builds a linux/amd64 program without cgo from src with the go
	// build flags.
	build := func(name, src string, flags ...string) string {
		t.Helper()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+name+"\n\ngo 1.25\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(dir, name)
		cmd := exec.Command(goBin, append(append([]string{"build", "-o", out}, flags...), ".")...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0", "GOFLAGS=")
		if msg, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go build %s: %v: %s", name, err, msg)
		}
		return out
	}
	const empty = "package main\n\nfunc main() {}\n"
	if dynamic, err := needsLoader(build("static", empty)); dynamic || err != nil {
		t.Errorf("needsLoader(static binary) = %v, %v", dynamic, err)
	}
	if dynamic, err := needsLoader(build("pie", empty, "-buildmode=pie")); !dynamic || err != nil {
		t.Errorf("needsLoader(position-independent binary) = %v, %v", dynamic, err)
	}
	// The directive purego uses to bind a library without cgo.
	const foreign = "package main\n\nimport _ \"unsafe\"\n\n//go:cgo_import_dynamic libfoo_bar bar \"libfoo.so.1\"\n//go:cgo_import_dynamic _ _ \"libfoo.so.1\"\n\nfunc main() {}\n"
	if _, err := needsLoader(build("foreign", foreign, "-buildmode=pie")); err == nil || !strings.Contains(err.Error(), "the shared library libfoo.so.1") {
		t.Errorf("needsLoader(binary needing libfoo.so.1) = %v", err)
	}
	src := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(src, []byte(empty), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := needsLoader(src); err == nil || !strings.Contains(err.Error(), "main.go: ") {
		t.Errorf("needsLoader(source file) = %v", err)
	}
}
