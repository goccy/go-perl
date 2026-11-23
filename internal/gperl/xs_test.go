package gperl

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-perl/internal/mod"
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
	t.Setenv("GPERLMODCACHE", t.TempDir())
	tree := filepath.Join(t.TempDir(), "Demo-XS@0.01")
	if err := copyTree(filepath.Join("testdata", "Demo-XS-dist"), tree); err != nil {
		t.Fatal(err)
	}
	if err := withDeps(func(m *mod.Manager) error {
		return buildDist(context.Background(), m, tree, mod.Target(), nil)
	}); err != nil {
		t.Fatalf("buildDist: %v", err)
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
	place(filepath.Join(mod.LibDir(tree), "Demo", "XS.pm"), "Demo", "XS.pm")
	place(filepath.Join(mod.TargetDir(tree, mod.Target()), "auto", "Demo", "XS", "XS.so"), "auto", "Demo", "XS", "XS.so")

	// A fork child cannot use it (native code is not available there): the
	// call dies, naming the sub and the module's shared object.
	status, stdout, stderr := runCLI(t, t.TempDir(), "", "-I"+lib, "-MDemo::XS", "-e",
		`print "sum=", Demo::XS::add(40, 2), "\n";`,
		`-e`, `my $pid = fork // die "fork: $!"; if (!$pid) { my $r = eval { Demo::XS::add(1, 2) }; print "child=", $r // ($@ =~ m{^Demo::XS::add: native modules cannot be used in a forked child \(\S*/XS\.so\) at } ? "refused" : "other: $@"), "\n"; exit 0 } waitpid($pid, 0);`)
	if status != 0 || stdout != "sum=42\nchild=refused\n" {
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

// DynaLoader has its whole set of entry points. A loaded library stays
// loaded, so dl_unload_file reports that it unloaded nothing.
func TestDynaLoaderUnloadFileReportsFailure(t *testing.T) {
	status, stdout, stderr := runCLI(t, t.TempDir(), "", "-MDynaLoader", "-e",
		`print defined(&DynaLoader::dl_unload_file) ? "defined" : "undefined", "\n";`,
		"-e", `print DynaLoader::dl_unload_file(1) ? "unloaded" : "kept: " . DynaLoader::dl_error(), "\n";`)
	if status != 0 || stdout != "defined\nkept: dl_unload_file: a loaded library is never unloaded\n" {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
}
