//go:build darwin || linux

package xs_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

// buildPlainLibrary compiles a shared library that is not a native XS
// module: plain C, with one function.
func buildPlainLibrary(t *testing.T) string {
	t.Helper()
	return buildSharedLibrary(t, t.TempDir(), "plain", "int plain_answer(void) { return 42; }\n")
}

// buildSharedLibrary compiles the C source code into lib<name>.so in dir.
// Symbols the code uses and does not define are left to be found when the
// library is loaded.
func buildSharedLibrary(t *testing.T, dir, name, code string) string {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	src := filepath.Join(dir, name+".c")
	if err := os.WriteFile(src, []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}
	so := filepath.Join(dir, "lib"+name+".so")
	args := []string{"-shared", "-fPIC", "-o", so, src}
	if runtime.GOOS == "darwin" {
		args = append(args, "-undefined", "dynamic_lookup")
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile lib%s.so: %v\n%s", name, err, out)
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

// dl_load_file's flags decide whether a library's symbols are global, as
// they do on perl (dl_dlopen.xs: bit 0x01 is RTLD_GLOBAL): a library
// loaded with 0x01 provides its symbols to the libraries loaded after it,
// and one loaded with 0 does not. XS::Loader loads every module with
// 0x01, and a module built on another module's library (Test::Catch on
// XS::libcatch) is found only then. A global symbol stays for the process,
// so each case's libraries have a symbol of their own.
func TestDynaLoaderFlagsMakeALibrarysSymbolsGlobal(t *testing.T) {
	load := func(flags int) string {
		symbol := "provided_value_" + strconv.Itoa(flags)
		dir := t.TempDir()
		lib := buildSharedLibrary(t, dir, "provider", "int "+symbol+" = 42;\n")
		// a data reference, bound when the library is loaded
		user := buildSharedLibrary(t, dir, "consumer", "extern int "+symbol+";\nint consumed_value(void) { return "+symbol+"; }\n")
		return evalWithInc(t, true, true, nil, `require DynaLoader;
DynaLoader::dl_load_file('`+lib+`', `+strconv.Itoa(flags)+`) or die "dl_load_file provider: " . DynaLoader::dl_error();
my $user = DynaLoader::dl_load_file('`+user+`', 0);
$user ? (DynaLoader::dl_find_symbol($user, "consumed_value") ? "linked" : "no symbol: " . DynaLoader::dl_error())
      : "unresolved"`)
	}
	if got := load(0x01); got != "linked" {
		t.Errorf("provider loaded with flags 0x01: got %q, want %q", got, "linked")
	}
	if got := load(0); got != "unresolved" {
		t.Errorf("provider loaded with flags 0: got %q, want %q", got, "unresolved")
	}
}
