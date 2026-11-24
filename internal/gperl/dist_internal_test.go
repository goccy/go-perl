package gperl

import (
	"os"
	"path/filepath"
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
