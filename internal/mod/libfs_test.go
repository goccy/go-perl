package mod

import (
	"io"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestLibFSMergesTrees(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(a, "Plack.pm"), "package Plack;")
	writeFile(t, filepath.Join(a, "Plack", "Util.pm"), "from a")
	writeFile(t, filepath.Join(b, "Plack", "Util.pm"), "from b")
	writeFile(t, filepath.Join(b, "Plack", "Request.pm"), "package Plack::Request;")
	writeFile(t, filepath.Join(b, "auto", "share", "dist", "X", "data.txt"), "shared")
	l, err := NewLibFS([]string{a, filepath.Join(t.TempDir(), "missing"), b})
	if err != nil {
		t.Fatal(err)
	}
	// The union behaves as one well-formed file system.
	if err := fstest.TestFS(l, "Plack.pm", "Plack/Util.pm", "Plack/Request.pm", "auto/share/dist/X/data.txt"); err != nil {
		t.Fatal(err)
	}
	if data, err := fs.ReadFile(l, "Plack/Util.pm"); err != nil || string(data) != "from a" {
		t.Fatalf("the earlier tree does not win: %q, %v", data, err)
	}
	if got := len(l.Files()); got != 4 {
		t.Fatalf("Files has %d entries, want 4", got)
	}
	entries, err := fs.ReadDir(l, "Plack")
	if err != nil || len(entries) != 2 || entries[0].Name() != "Request.pm" || entries[1].Name() != "Util.pm" {
		t.Fatalf("ReadDir(Plack) = %v, %v", entries, err)
	}

	d, err := l.Open("Plack")
	if err != nil {
		t.Fatal(err)
	}
	dir := d.(fs.ReadDirFile)
	if one, err := dir.ReadDir(1); err != nil || len(one) != 1 {
		t.Fatalf("ReadDir(1) = %v, %v", one, err)
	}
	if rest, err := dir.ReadDir(5); err != nil || len(rest) != 1 {
		t.Fatalf("ReadDir(5) = %v, %v", rest, err)
	}
	if _, err := dir.ReadDir(1); err != io.EOF {
		t.Fatalf("ReadDir at the end: %v", err)
	}
	if _, err := d.Read(make([]byte, 1)); err == nil {
		t.Fatal("read a directory")
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"Missing.pm", "Plack/Missing.pm"} {
		if _, err := l.Open(name); err == nil {
			t.Errorf("Open(%s) succeeded", name)
		}
		if _, err := l.ReadFile(name); err == nil {
			t.Errorf("ReadFile(%s) succeeded", name)
		}
		if _, err := l.Stat(name); err == nil {
			t.Errorf("Stat(%s) succeeded", name)
		}
	}
	if _, err := l.ReadDir("Missing"); err == nil {
		t.Error("ReadDir(Missing) succeeded")
	}
}
