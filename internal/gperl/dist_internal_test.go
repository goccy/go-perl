package gperl

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-perl/internal/mod"
)

// The archlib files of a distribution with native modules (inside auto/
// too) reach the module tree from the build for each target, though the
// build left them read-only (as ExtUtils::MakeMaker does); the native
// modules reach the target's directory.
func TestHarvestArchlibFilesForTwoTargets(t *testing.T) {
	src := t.TempDir()
	tree := filepath.Join(t.TempDir(), "X@0.01")
	for name, mode := range map[string]os.FileMode{
		"blib/arch/X/Util.pm": 0o444, "blib/arch/auto/X/x.h": 0o444,
		"blib/arch/auto/X/X.a": 0o444, "blib/arch/auto/X/X.so": 0o555, "blib/arch/auto/X/X.ldflags": 0o444,
	} {
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
		for _, name := range []string{"X.a", "X.so", "X.ldflags"} {
			if _, err := os.Stat(filepath.Join(mod.TargetDir(tree, target), "auto", "X", name)); err != nil {
				t.Errorf("native module for %s: %v", target, err)
			}
			if _, err := os.Stat(filepath.Join(mod.BuildDir(tree), "lib", "auto", "X", name)); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("auto/X/%s in the module tree: %v, want none", name, err)
			}
		}
	}
	for _, name := range []string{"X/Util.pm", "auto/X/x.h"} {
		got, err := os.ReadFile(filepath.Join(mod.BuildDir(tree), "lib", filepath.FromSlash(name)))
		if err != nil || string(got) != "blib/arch/"+name {
			t.Fatalf("%s in the module tree = %q, %v", name, got, err)
		}
	}
}

// The copy a distribution builds in keeps the modification time of each
// file (a symbolic link's: its target's), as make compares them.
func TestCopyDistSourceKeepsModTimes(t *testing.T) {
	tree := t.TempDir()
	want := map[string]time.Time{}
	for i, name := range []string{"README.md", "lib/X.pm", "lib/X/Util.pm"} {
		path := filepath.Join(tree, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(name), 0o444); err != nil {
			t.Fatal(err)
		}
		modTime := time.Date(2001, 2, 3, 4, 5, i, 0, time.UTC)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatal(err)
		}
		want[name] = modTime
	}
	if err := os.Symlink("X.pm", filepath.Join(tree, "lib", "Y.pm")); err != nil {
		t.Fatal(err)
	}
	want["lib/Y.pm"] = want["lib/X.pm"]
	dst := filepath.Join(t.TempDir(), "src")
	if err := copyDistSource(tree, dst); err != nil {
		t.Fatal(err)
	}
	for name, modTime := range want {
		info, err := os.Stat(filepath.Join(dst, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().Equal(modTime) {
			t.Errorf("%s: modification time %v, want %v", name, info.ModTime(), modTime)
		}
	}
}

// The copy a distribution builds in follows a link to a directory, and
// fails for a link to a directory the copy is already inside: one holding
// the link, or one a loop of links leads back to.
func TestCopyDistSourceLinkedDirectories(t *testing.T) {
	build := func(links map[string]string) string {
		tree := t.TempDir()
		for _, dir := range []string{"lib/a", "lib/b", "lib/c"} {
			if err := os.MkdirAll(filepath.Join(tree, filepath.FromSlash(dir)), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(tree, "lib", "c", "X.pm"), []byte("1;"), 0o644); err != nil {
			t.Fatal(err)
		}
		for name, link := range links {
			if err := os.Symlink(link, filepath.Join(tree, filepath.FromSlash(name))); err != nil {
				t.Fatal(err)
			}
		}
		return tree
	}
	dst := filepath.Join(t.TempDir(), "src")
	if err := copyDistSource(build(map[string]string{"lib/a/c": "../c"}), dst); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(dst, "lib", "a", "c", "X.pm")); err != nil || string(got) != "1;" {
		t.Fatalf("lib/a/c/X.pm = %q, %v", got, err)
	}
	for name, links := range map[string]map[string]string{
		"link to the top directory": {"lib/a/top": "../.."},
		"link to its own directory": {"lib/a/self": "."},
		"link to the tree's parent": {"lib/a/up": "../../.."},
		"loop of links":             {"lib/a/b": "../b", "lib/b/a": "../a"},
	} {
		err := copyDistSource(build(links), filepath.Join(t.TempDir(), "src"))
		if !errors.Is(err, errLinkInsideCopy) {
			t.Errorf("%s: %v", name, err)
		}
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
