package mod

import (
	"os"
	"path/filepath"
	"testing"
)

// A tree is built for this gperl only: the mark of another gperl, or the
// empty one of an earlier gperl, asks for a build again.
func TestBuiltByThisGperl(t *testing.T) {
	tree := filepath.Join(t.TempDir(), "X@0.01")
	if Built(tree, "linux_amd64") {
		t.Fatal("a tree never built is built")
	}
	if err := MarkBuilt(tree, "linux_amd64"); err != nil {
		t.Fatal(err)
	}
	if !Built(tree, "linux_amd64") || Built(tree, "linux_arm64") {
		t.Fatalf("built for linux_amd64 = %v, for linux_arm64 = %v; want true, false", Built(tree, "linux_amd64"), Built(tree, "linux_arm64"))
	}
	for _, mark := range []string{"", "gperl v0.0.1"} {
		if err := os.WriteFile(filepath.Join(TargetDir(tree, "linux_amd64"), builtMarker), []byte(mark), 0o644); err != nil {
			t.Fatal(err)
		}
		if Built(tree, "linux_amd64") {
			t.Errorf("built with the mark %q of another gperl", mark)
		}
	}
	if err := MarkBuilt(tree, ""); err != nil {
		t.Fatal(err)
	}
	if !Built(tree, "linux_arm64") {
		t.Error("a tree built for every target is not built for linux_arm64")
	}
}
