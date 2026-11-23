package perl

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// The cached standard library is unpacked once: every caller, including
// ones racing on an empty cache, gets the same complete directory, and no
// partial copy is left behind.
func TestStdlibInCacheIsUnpackedOnce(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache")
	const callers = 4
	dirs := make([]string, callers)
	errs := make([]error, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dirs[i], errs[i] = stdlibInCache(root)
		}()
	}
	wg.Wait()
	for i := range callers {
		if errs[i] != nil {
			t.Fatalf("caller %d: %v", i, errs[i])
		}
		if dirs[i] != dirs[0] {
			t.Fatalf("caller %d got %q, caller 0 got %q", i, dirs[i], dirs[0])
		}
	}
	if _, err := os.Stat(filepath.Join(dirs[0], "strict.pm")); err != nil {
		t.Fatalf("the cached library has no strict.pm: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Fatalf("cache holds %v, want only %q", names, filepath.Base(dirs[0]))
	}

	// A later call finds the directory and unpacks nothing.
	marker := filepath.Join(dirs[0], "unpacked-once")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	again, err := stdlibInCache(root)
	if err != nil {
		t.Fatal(err)
	}
	if again != dirs[0] {
		t.Fatalf("second call returned %q, want %q", again, dirs[0])
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("the directory was replaced: %v", err)
	}
}
