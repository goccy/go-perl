//go:build darwin || linux

package internal

import (
	"errors"
	"os"
	"slices"
	"syscall"
	"testing"
)

// openNow returns the descriptors of the process that are open.
func openNow(t *testing.T) []int {
	t.Helper()
	fds, err := openDescriptors()
	if err != nil {
		t.Fatal(err)
	}
	var open []int
	for _, fd := range fds {
		_, err := descriptorFlags(fd)
		if errors.Is(err, syscall.EBADF) {
			continue // the listing's own, closed by now
		}
		if err != nil {
			t.Fatal(err)
		}
		open = append(open, fd)
	}
	return open
}

// TestFailedNewGivesBackWhatItMade pins that a New that fails after its
// host state is made gives it back: here the interpreter cannot start (no
// library), and no descriptor the instance made is left open.
func TestFailedNewGivesBackWhatItMade(t *testing.T) {
	// The Go runtime's own descriptors are made before the count.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := errors.Join(r.Close(), w.Close()); err != nil {
		t.Fatal(err)
	}
	want := openNow(t)

	p, err := New(InstanceOptions{StdlibDir: t.TempDir(), Env: []string{"GO_PERL_TEST_KEY=" + t.Name()}})
	if err == nil {
		if cerr := p.Close(); cerr != nil {
			t.Fatalf("New with an empty library directory succeeded; Close: %v", cerr)
		}
		t.Fatal("New with an empty library directory succeeded")
	}
	if p != nil {
		t.Fatalf("New returned an instance with its error %v", err)
	}
	if after := openNow(t); !slices.Equal(after, want) {
		t.Fatalf("open descriptors after the failed New: %v, want %v", after, want)
	}
}
