//go:build darwin || linux

package xs_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// buildPlainLibrary compiles a shared library that is not a native XS
// module: plain C, with one function.
func buildPlainLibrary(t *testing.T) string {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "plain.c")
	if err := os.WriteFile(src, []byte("int plain_answer(void) { return 42; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	so := filepath.Join(dir, "libplain.so")
	args := []string{"-shared", "-fPIC", "-o", so, src}
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile libplain.so: %v\n%s", err, out)
	}
	return so
}

// DynaLoader's dl_load_file loads any shared library, as dlopen does, and
// dl_find_symbol finds the library's functions: FFI::CheckLib looks for a C
// library with a given function that way. A symbol the library does not
// have is not found, and dl_error says why.
func TestDynaLoaderLoadsAPlainSharedLibrary(t *testing.T) {
	so := buildPlainLibrary(t)
	got := evalWithInc(t, true, true, nil, `require DynaLoader;
my $lib = DynaLoader::dl_load_file('`+so+`', 0) or die "dl_load_file: " . DynaLoader::dl_error();
my $found = DynaLoader::dl_find_symbol($lib, "plain_answer") ? "found" : "missing: " . DynaLoader::dl_error();
my $other = DynaLoader::dl_find_symbol($lib, "no_such_function") ? "found" : (DynaLoader::dl_error() ne "" ? "missing, with an error" : "missing, silently");
"$found|$other"`)
	if want := "found|missing, with an error"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
