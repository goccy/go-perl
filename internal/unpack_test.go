package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// A tree is written once per content: the same tree again is the same
// directory, another tree another one, and both have the files as written.
// Without a user cache directory (no HOME) the tree goes into this user's
// directory in the temporary directory.
func TestUnpackTree(t *testing.T) {
	for _, place := range []string{"cache", "temp"} {
		t.Run(place, func(t *testing.T) {
			if place == "cache" {
				t.Setenv("HOME", t.TempDir())
				t.Setenv("XDG_CACHE_HOME", "")
			} else {
				t.Setenv("HOME", "")
				t.Setenv("XDG_CACHE_HOME", "")
				t.Setenv("TMPDIR", t.TempDir())
			}
			a := fstest.MapFS{"A.pm": {Data: []byte("a")}, "B/C.pm": {Data: []byte("c")}}
			dir, err := UnpackTree(a, "tree")
			if err != nil {
				t.Fatal(err)
			}
			if place == "temp" && !strings.HasPrefix(dir, os.TempDir()) {
				t.Errorf("tree in %s, want it in the temporary directory %s", dir, os.TempDir())
			}
			if got, err := os.ReadFile(filepath.Join(dir, "B", "C.pm")); err != nil || string(got) != "c" {
				t.Errorf("B/C.pm = %q (%v)", got, err)
			}
			again, err := UnpackTree(a, "tree")
			if err != nil || again != dir {
				t.Errorf("the same tree again went to %s (%v), want %s", again, err, dir)
			}
			// The same bytes split differently between a content and the
			// next name are another tree.
			other, err := UnpackTree(fstest.MapFS{"A.pm": {Data: []byte("aB/")}, "C.pm": {Data: []byte("c")}}, "tree")
			if err != nil || other == dir {
				t.Errorf("another tree went to %s (%v), the first one's %s", other, err, dir)
			}
		})
	}
}
