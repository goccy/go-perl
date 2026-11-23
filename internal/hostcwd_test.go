//go:build darwin || linux

package internal

import (
	"os"
	"path/filepath"
	"testing"
)

// The directory a command inherits is the one the host process is in now:
// a chdir of one program, or of the embedder, is seen by every other.
func TestCommandDirIsTheHostProcesssDirectoryNow(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	t.Chdir(first)
	// One program goes into first, another then into second.
	if err := hostChdir(first); err != nil {
		t.Fatal(err)
	}
	if err := hostChdir(second); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ dir, want string }{{first, first}, {second, ""}} {
		got, err := hostCommandDir(c.dir)
		if err != nil || got != c.want {
			t.Errorf("hostCommandDir(%q) = %q, %v; want %q", c.dir, got, err, c.want)
		}
	}
	// The embedder goes back into first itself.
	if err := os.Chdir(first); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ dir, want string }{{first, ""}, {second, second}} {
		got, err := hostCommandDir(c.dir)
		if err != nil || got != c.want {
			t.Errorf("after the embedder's chdir, hostCommandDir(%q) = %q, %v; want %q", c.dir, got, err, c.want)
		}
	}
}

// A command started while the host process is in a directory that is gone
// starts in its own directory, as without the hook.
func TestCommandDirWhenTheHostProcesssDirectoryIsGone(t *testing.T) {
	gone := filepath.Join(t.TempDir(), "gone")
	if err := os.Mkdir(gone, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(gone)
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	got, err := hostCommandDir("/")
	if err != nil || got != "/" {
		t.Errorf(`hostCommandDir("/") = %q, %v; want "/"`, got, err)
	}
}
