package main

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

func TestExtractRecordsMakesOnlyFilesAndDirectories(t *testing.T) {
	// The tar stream docker cp writes of a container's /out, as a
	// malicious distribution can make it: links, a fifo, a hard link, and
	// names that lead out of the directory.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	add := func(h *tar.Header, body string) {
		t.Helper()
		h.Size = int64(len(body))
		if h.Mode == 0 {
			h.Mode = 0o4755
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	add(&tar.Header{Name: "out/", Typeflag: tar.TypeDir}, "")
	add(&tar.Header{Name: "out/phases.jsonl", Typeflag: tar.TypeReg}, "{}\n")
	add(&tar.Header{Name: "out/sub/", Typeflag: tar.TypeDir}, "")
	add(&tar.Header{Name: "out/sub/test.log", Typeflag: tar.TypeReg}, "ok\n")
	add(&tar.Header{Name: "out/unmet.json", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"}, "")
	add(&tar.Header{Name: "out/escape", Typeflag: tar.TypeSymlink, Linkname: "../../../../etc/hosts"}, "")
	add(&tar.Header{Name: "out/fifo", Typeflag: tar.TypeFifo}, "")
	add(&tar.Header{Name: "out/hard", Typeflag: tar.TypeLink, Linkname: "out/phases.jsonl"}, "")
	add(&tar.Header{Name: "out/../evil", Typeflag: tar.TypeReg}, "evil\n")
	add(&tar.Header{Name: "/abs", Typeflag: tar.TypeReg}, "evil\n")
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	parent := t.TempDir()
	dir := filepath.Join(parent, "Dist")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := extractRecords(&buf, dir); err != nil {
		t.Fatal(err)
	}
	var left []string
	err := filepath.WalkDir(parent, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(parent, path)
		left = append(left, rel+" "+e.Type().String())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{". d---------", "Dist d---------", "Dist/phases.jsonl ----------", "Dist/sub d---------", "Dist/sub/test.log ----------"}
	if !reflect.DeepEqual(left, want) {
		t.Errorf("made %q; want %q", left, want)
	}
	fi, err := os.Stat(filepath.Join(dir, "phases.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode() != 0o644 {
		t.Errorf("phases.jsonl mode %v; want -rw-r--r--", fi.Mode())
	}
	if raw, _ := os.ReadFile(filepath.Join(dir, "sub", "test.log")); string(raw) != "ok\n" {
		t.Errorf("sub/test.log holds %q", raw)
	}
}

func TestRemoveSpecialFiles(t *testing.T) {
	// Files and directories stay; a link (to something outside, to a file
	// inside) and a fifo go, also below a subdirectory.
	dir := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.WriteFile(filepath.Join(dir, "phases.jsonl"), []byte("{}\n"), 0o644))
	must(os.Mkdir(filepath.Join(dir, "sub"), 0o755))
	must(os.WriteFile(filepath.Join(dir, "sub", "test.log"), []byte("ok\n"), 0o644))
	must(os.Symlink("/etc/passwd", filepath.Join(dir, "tests.jsonl")))
	must(os.Symlink("phases.jsonl", filepath.Join(dir, "unmet.json")))
	must(os.Symlink("..", filepath.Join(dir, "sub", "up")))
	must(syscall.Mkfifo(filepath.Join(dir, "build.log"), 0o644))
	must(syscall.Mkfifo(filepath.Join(dir, "sub", "fifo"), 0o644))
	must(removeSpecialFiles(dir))
	var left []string
	must(filepath.WalkDir(dir, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		left = append(left, rel)
		return nil
	}))
	if want := []string{".", "phases.jsonl", "sub", "sub/test.log"}; !reflect.DeepEqual(left, want) {
		t.Errorf("left %v; want %v", left, want)
	}
}
