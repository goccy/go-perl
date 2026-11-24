package gperl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-perl/internal/mod"
)

// The archlib files of a distribution with native modules reach the module
// tree from the build for each target, though the build left them
// read-only (as ExtUtils::MakeMaker does).
func TestHarvestArchlibFilesForTwoTargets(t *testing.T) {
	src := t.TempDir()
	tree := filepath.Join(t.TempDir(), "X@0.01")
	for name, mode := range map[string]os.FileMode{"blib/arch/X/Util.pm": 0o444, "blib/arch/auto/X/X.a": 0o444, "blib/arch/auto/X/X.so": 0o555} {
		path := filepath.Join(src, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(name), mode); err != nil {
			t.Fatal(err)
		}
	}
	for _, target := range []string{"linux_amd64", "linux_arm64"} {
		if err := harvest(src, tree, target); err != nil {
			t.Fatalf("harvest for %s: %v", target, err)
		}
		if !mod.Built(tree, target) {
			t.Errorf("not built for %s", target)
		}
	}
	got, err := os.ReadFile(filepath.Join(mod.BuildDir(tree), "lib", "X", "Util.pm"))
	if err != nil || string(got) != "blib/arch/X/Util.pm" {
		t.Fatalf("X/Util.pm in the module tree = %q, %v", got, err)
	}
}

// Each module's link line ends with the libraries perl itself is linked
// with (%Config{perllibs}), after the module's own.
func TestBridgeLinksPerlLibsAfterTheModule(t *testing.T) {
	tree := filepath.Join(t.TempDir(), "X@0.01")
	auto := filepath.Join(mod.TargetDir(tree, "linux_amd64"), "auto", "X")
	if err := os.MkdirAll(auto, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"X.a": "", "X.ldflags": "-lfoo"} {
		if err := os.WriteFile(filepath.Join(auto, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := writeBridge(tree); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(filepath.Join(mod.ABIDir(tree), "xs.go"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "#cgo linux,amd64,!musl LDFLAGS: ${SRCDIR}/linux_amd64/auto/X/X.a -lfoo -lm\n"; !strings.Contains(string(src), want) {
		t.Fatalf("xs.go lacks %q:\n%s", want, src)
	}
}
