package gperl

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/goccy/go-perl/internal/mod"
)

// sharedLibrary compiles src into a shared library at path.
func sharedLibrary(t *testing.T, path, src string) {
	t.Helper()
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("no cc available")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	c := path + ".c"
	if err := os.WriteFile(c, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	args := []string{"-shared", "-fPIC", "-o", path, c}
	if runtime.GOOS == "darwin" {
		// The darwin linker refuses an undefined symbol in a shared library
		// unless told to leave it to the dynamic loader, as ELF linkers do.
		args = append(args, "-undefined", "dynamic_lookup")
	}
	if out, err := exec.Command("cc", args...).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s", err, out)
	}
	if err := os.Remove(c); err != nil {
		t.Fatal(err)
	}
}

// A shared library exports a name to the dynamic loader when it defines it
// with default visibility.
func TestExportsSymbol(t *testing.T) {
	dir := t.TempDir()
	target := runtime.GOOS + "_" + runtime.GOARCH
	for name, c := range map[string]struct {
		src  string
		want bool
	}{
		"defines": {"void boot_X(void) {}\n", true},
		"hidden":  {"__attribute__((visibility(\"hidden\"))) void boot_X(void) {}\nvoid other(void) { boot_X(); }\n", false},
		"uses":    {"void boot_X(void);\nvoid other(void) { boot_X(); }\n", false},
		"none":    {"void other(void) {}\n", false},
	} {
		so := filepath.Join(dir, name+".so")
		sharedLibrary(t, so, c.src)
		got, err := exportsSymbol(target, so, "boot_X")
		if err != nil || got != c.want {
			t.Errorf("%s: exportsSymbol = %v, %v; want %v", name, got, err, c.want)
		}
	}
}

// Of the shared libraries without gperl's archive, the one perl boots as a
// module (it exports boot_<Module>) was linked by something else; the
// others, and the distribution's own archives, are kept in the target
// directory where the build put them.
func TestHarvestDistributionLibraries(t *testing.T) {
	target := runtime.GOOS + "_" + runtime.GOARCH
	src := t.TempDir()
	auto := filepath.Join(src, "blib", "arch", "auto", "X")
	sharedLibrary(t, filepath.Join(auto, "X.so"), "void boot_X(void) {}\n")
	if err := harvest(src, filepath.Join(t.TempDir(), "X@0.01"), target); err == nil || !strings.Contains(err.Error(), "was not linked by gperl") {
		t.Fatalf("harvest of a module without its archive: %v", err)
	}

	for _, name := range []string{"X.a", "X.ldflags", "libh.a"} {
		if err := os.WriteFile(filepath.Join(auto, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sharedLibrary(t, filepath.Join(auto, "plug-in", "plug-in.so"), "int plug(void) { return 1; }\n")
	tree := filepath.Join(t.TempDir(), "X@0.01")
	if err := harvest(src, tree, target); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"X.so", "X.a", "X.ldflags", "libh.a", "plug-in/plug-in.so"} {
		if _, err := os.Stat(filepath.Join(mod.TargetDir(tree, target), "auto", "X", filepath.FromSlash(name))); err != nil {
			t.Errorf("auto/X/%s: %v", name, err)
		}
	}
}
