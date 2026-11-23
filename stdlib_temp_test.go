//go:build unix

package perl

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// withoutUserCache makes the user's cache directory unusable, as a test of
// a CPAN distribution makes it by running with HOME unset, and gives the
// process a temporary directory of the test's.
func withoutUserCache(t *testing.T) (tmp string) {
	t.Helper()
	tmp = t.TempDir()
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("TMPDIR", tmp)
	if _, err := os.UserCacheDir(); err == nil {
		t.Skip("the user cache directory is usable without HOME here")
	}
	return tmp
}

func userTempRoot(tmp string) string {
	return filepath.Join(tmp, "go-perl-"+strconv.Itoa(os.Getuid()))
}

// Without a user cache directory, the standard library goes to one fixed
// place of the user's in the temporary directory, once: every later process
// finds it there. Before, each process unpacked it into a new temporary
// directory and left it behind (796 of them, 24 GB, in one survey's
// containers).
func TestStdlibWithoutUserCacheIsUnpackedOnce(t *testing.T) {
	tmp := withoutUserCache(t)
	first, err := cachedStdlib()
	if err != nil {
		t.Fatal(err)
	}
	root := userTempRoot(tmp)
	if filepath.Dir(first) != root {
		t.Fatalf("the library is at %s, want it in %s", first, root)
	}
	if _, err := os.Stat(filepath.Join(first, "strict.pm")); err != nil {
		t.Fatalf("the library has no strict.pm: %v", err)
	}
	fi, err := os.Lstat(root)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o700 {
		t.Fatalf("%s has mode %#o, want 0700", root, perm)
	}
	marker := filepath.Join(first, "unpacked-once")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	// each call stands for a process of its own: nothing is remembered
	// between them but what is on disk
	for i := 0; i < 3; i++ {
		again, err := cachedStdlib()
		if err != nil {
			t.Fatal(err)
		}
		if again != first {
			t.Fatalf("call %d got %s, the first got %s", i, again, first)
		}
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("the library was unpacked again: %v", err)
	}
	for _, d := range []string{tmp, root} {
		entries, err := os.ReadDir(d)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			names := make([]string, len(entries))
			for i, e := range entries {
				names[i] = e.Name()
			}
			t.Fatalf("%s holds %v, want one entry", d, names)
		}
	}
}

// The temporary directory is shared with other users: what stands at the
// user's place there is used only when it is a directory of the user's own,
// with mode 0700. Anything else is refused with an error that says why, and
// nothing is written through it.
func TestStdlibTempRootRefusesWhatIsNotTheUsers(t *testing.T) {
	t.Run("a symbolic link", func(t *testing.T) {
		tmp := withoutUserCache(t)
		elsewhere := t.TempDir()
		if err := os.Symlink(elsewhere, userTempRoot(tmp)); err != nil {
			t.Fatal(err)
		}
		_, err := cachedStdlib()
		if err == nil || !strings.Contains(err.Error(), "symbolic link") {
			t.Fatalf("got %v, want a refusal of the symbolic link", err)
		}
		if entries, _ := os.ReadDir(elsewhere); len(entries) != 0 {
			t.Fatalf("the link was followed: its target holds %d entries", len(entries))
		}
	})
	t.Run("a file", func(t *testing.T) {
		tmp := withoutUserCache(t)
		if err := os.WriteFile(userTempRoot(tmp), nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := cachedStdlib(); err == nil || !strings.Contains(err.Error(), "not a directory") {
			t.Fatalf("got %v, want a refusal of the file", err)
		}
	})
	t.Run("a directory others can write to", func(t *testing.T) {
		tmp := withoutUserCache(t)
		root := userTempRoot(tmp)
		if err := os.Mkdir(root, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(root, 0o777); err != nil {
			t.Fatal(err)
		}
		if _, err := cachedStdlib(); err == nil || !strings.Contains(err.Error(), "not 0700") {
			t.Fatalf("got %v, want a refusal of the mode", err)
		}
		if entries, _ := os.ReadDir(root); len(entries) != 0 {
			t.Fatalf("the directory was used: it holds %d entries", len(entries))
		}
	})
	t.Run("a directory of another user's", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "planted")
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		// the directory is this user's; to the check it belongs to
		// someone else when another user asks
		other := os.Getuid() + 1
		err := privateDir(dir, other, false)
		if err == nil || !strings.Contains(err.Error(), "belongs to user") {
			t.Fatalf("got %v, want a refusal of the owner", err)
		}
		if err := privateDir(dir, os.Getuid(), false); err != nil {
			t.Fatalf("the user's own directory was refused: %v", err)
		}
	})
	t.Run("a mode the umask narrowed", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "made")
		if err := os.Mkdir(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		if err := privateDir(dir, os.Getuid(), true); err != nil {
			t.Fatal(err)
		}
		if fi, _ := os.Lstat(dir); fi.Mode().Perm() != 0o700 {
			t.Fatalf("mode %#o, want 0700", fi.Mode().Perm())
		}
	})
}

// Processes that start at once all get the one library, and none ever sees
// a part of it: the library's directory exists only complete. A watcher
// that looks while they unpack finds no directory or every file.
func TestStdlibWithoutUserCacheAppearsWhole(t *testing.T) {
	tmp := withoutUserCache(t)
	zr, err := zip.NewReader(bytes.NewReader(stdlibZip), int64(len(stdlibZip)))
	if err != nil {
		t.Fatal(err)
	}
	files := 0
	for _, f := range zr.File {
		if !f.FileInfo().IsDir() {
			files++
		}
	}
	root := userTempRoot(tmp)
	var done atomic.Bool
	var seen, partial atomic.Int32
	watcher := make(chan struct{})
	go func() {
		defer close(watcher)
		for !done.Load() {
			entries, err := os.ReadDir(root)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if !strings.HasPrefix(e.Name(), "stdlib-") {
					continue
				}
				n := 0
				filepath.WalkDir(filepath.Join(root, e.Name()), func(_ string, d fs.DirEntry, err error) error {
					if err == nil && !d.IsDir() {
						n++
					}
					return nil
				})
				seen.Add(1)
				if n != files {
					partial.Add(1)
				}
			}
		}
	}()
	const processes = 4
	dirs := make([]string, processes)
	errs := make([]error, processes)
	var wg sync.WaitGroup
	for i := range processes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dirs[i], errs[i] = cachedStdlib()
		}()
	}
	wg.Wait()
	done.Store(true)
	<-watcher
	for i := range processes {
		if errs[i] != nil {
			t.Fatalf("process %d: %v", i, errs[i])
		}
		if dirs[i] != dirs[0] {
			t.Fatalf("process %d got %s, process 0 got %s", i, dirs[i], dirs[0])
		}
	}
	if partial.Load() != 0 {
		t.Fatalf("the library was seen in part %d of %d times", partial.Load(), seen.Load())
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("%s holds %d entries, want the library alone", root, len(entries))
	}
}
