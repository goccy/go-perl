package perl

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	iofs "io/fs"
	"os"
	"runtime"
	"sort"
	"syscall"
	"testing"
	"testing/fstest"
	"time"

	goperlfs "github.com/goccy/go-perl/fs"
	"github.com/goccy/go-perl/internal"
	"github.com/goccy/perlwasm2go/base"
)

func newStdlibFS(t *testing.T) (FS, *MemFS) {
	t.Helper()
	mem := goperlfs.NewMemFS()
	fsys, err := WithStdlib(mem)
	if err != nil {
		t.Fatal(err)
	}
	return fsys, mem
}

func zipContent(t *testing.T, name string) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(stdlibZip), int64(len(stdlibZip)))
	if err != nil {
		t.Fatal(err)
	}
	f, err := zr.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func readFile(t *testing.T, fsys FS, name string) []byte {
	t.Helper()
	f, err := fsys.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return b
}

func dirNames(t *testing.T, fsys FS, name string) []string {
	t.Helper()
	f, err := fsys.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	defer f.Close()
	ents, err := f.ReadDir(-1)
	if err != nil {
		t.Fatalf("readdir %s: %v", name, err)
	}
	names := make([]string, 0, len(ents))
	for _, e := range ents {
		names = append(names, e.Name())
	}
	return names
}

func TestWithStdlibServesTheLibrary(t *testing.T) {
	fsys, _ := newStdlibFS(t)
	want := zipContent(t, "strict.pm")
	for _, name := range []string{
		stdlibMount + "/strict.pm",
		"/" + stdlibMount + "/strict.pm",
		stdlibMount + "/File/../strict.pm",
	} {
		if got := readFile(t, fsys, name); !bytes.Equal(got, want) {
			t.Errorf("%s: content differs from the embedded strict.pm", name)
		}
	}
	fi, err := fsys.Stat(stdlibMount + "/File")
	if err != nil || !fi.IsDir() {
		t.Fatalf("stat File = %v, %v; want a directory", fi, err)
	}
	names := dirNames(t, fsys, stdlibMount)
	if !sort.StringsAreSorted(names) {
		t.Errorf("library listing is not sorted: %v", names)
	}
	for _, want := range []string{"strict.pm", "File"} {
		if i := sort.SearchStrings(names, want); i == len(names) || names[i] != want {
			t.Errorf("library listing lacks %s", want)
		}
	}
	// The mount point is reachable by listing the root, and lists the
	// library.
	for dir, child := range map[string]string{"": goperlMount, goperlMount: "stdlib"} {
		if got := dirNames(t, fsys, dir); len(got) != 1 || got[0] != child {
			t.Errorf("listing %q = %v, want [%s]", dir, got, child)
		}
	}
}

func TestWithStdlibIsReadOnly(t *testing.T) {
	fsys, mem := newStdlibFS(t)
	lib := stdlibMount + "/"
	isPerm := func(what string, err error) {
		t.Helper()
		if !errors.Is(err, iofs.ErrPermission) {
			t.Errorf("%s: err = %v, want a permission error", what, err)
		}
	}
	for _, flag := range []int{os.O_WRONLY, os.O_RDWR, os.O_RDONLY | os.O_TRUNC, os.O_WRONLY | os.O_APPEND} {
		_, err := fsys.OpenFile(lib+"strict.pm", flag, 0)
		isPerm("open strict.pm for writing", err)
	}
	_, err := fsys.OpenFile(lib+"New.pm", os.O_WRONLY|os.O_CREATE, 0o644)
	isPerm("create New.pm", err)
	isPerm("mkdir newdir", fsys.Mkdir(lib+"newdir", 0o755))
	if err := fsys.Mkdir(lib+"File", 0o755); !errors.Is(err, iofs.ErrExist) {
		t.Errorf("mkdir File: err = %v, want exist", err)
	}
	isPerm("unlink strict.pm", fsys.Unlink(lib+"strict.pm"))
	isPerm("rmdir the mount point", fsys.Rmdir(stdlibMount))
	isPerm("rename strict.pm out", fsys.Rename(lib+"strict.pm", "strict.pm"))

	f, err := fsys.OpenFile(lib+"strict.pm", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write([]byte("x"))
	isPerm("write through a read-only handle", err)
	isPerm("truncate through a read-only handle", f.Truncate(0))
	isPerm("chmod through a read-only handle", f.Chmod(0o644))
	isPerm("chown through a read-only handle", f.Chown(0, 0))

	// Metadata changes are refused; on a missing name, the name's own error.
	now := time.Now()
	for what, err := range map[string]error{
		"chmod":    fsys.Chmod(lib+"strict.pm", 0o644),
		"chown":    fsys.Chown(lib+"strict.pm", 0, 0),
		"lchown":   fsys.Lchown(lib+"strict.pm", 0, 0),
		"chtimes":  fsys.Chtimes(lib+"strict.pm", now, now),
		"lchtimes": fsys.Lchtimes(lib+"strict.pm", now, now),
		"symlink":  fsys.Symlink("strict.pm", lib+"link.pm"),
		"link":     fsys.Link(lib+"strict.pm", "strict.pm"),
	} {
		isPerm(what, err)
	}
	if err := fsys.Chmod(lib+"No/Such.pm", 0o644); !errors.Is(err, iofs.ErrNotExist) {
		t.Errorf("chmod of a missing file: %v", err)
	}
	if _, err := fsys.OpenFile(lib+"strict.pm", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644); !errors.Is(err, iofs.ErrExist) {
		t.Errorf("exclusive create of strict.pm: %v", err)
	}

	// access(2) answers by the modes, and never for writing.
	access := fsys.(base.AccessFS)
	if err := access.Access(lib+"strict.pm", 4, true); err != nil {
		t.Errorf("R_OK on strict.pm: %v", err)
	}
	if err := access.Access(lib+"File", 1, true); err != nil {
		t.Errorf("X_OK on File: %v", err)
	}
	isPerm("W_OK on strict.pm", access.Access(lib+"strict.pm", 2, true))
	isPerm("X_OK on strict.pm", access.Access(lib+"strict.pm", 1, true))
	// A lock on a library file is granted; nothing can change it.
	if err := fsys.(base.LockFS).Flock(f, 2); err != nil {
		t.Errorf("flock strict.pm: %v", err)
	}

	// The mount point cannot be moved, replaced or removed.
	if err := mem.WriteFile("x", nil, 0o644); err != nil {
		t.Fatal(err)
	}
	isPerm("rename the mount point", fsys.Rename(goperlMount, "elsewhere"))
	isPerm("rename over the mount point", fsys.Rename("x", goperlMount))
	isPerm("rmdir the mount point", fsys.Rmdir(goperlMount))
}

func TestWithStdlibWritesGoToTheMemFS(t *testing.T) {
	fsys, mem := newStdlibFS(t)
	f, err := fsys.OpenFile("app.pl", os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("print 1;\n")); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, mem, "app.pl"); string(got) != "print 1;\n" {
		t.Fatalf("app.pl in the MemFS = %q", got)
	}
	if err := fsys.Mkdir("usr", 0o755); err != nil {
		t.Fatalf("mkdir beside the mount point: %v", err)
	}
	// Another instance's FS sees none of it.
	other, _ := newStdlibFS(t)
	if _, err := other.Stat("app.pl"); !errors.Is(err, iofs.ErrNotExist) {
		t.Fatalf("app.pl leaked into another FS: %v", err)
	}
}

// open(2), unlink(2) and a trailing slash behave as on a read-only mount,
// and a trailing slash still reaches the MemFS beside it.
func TestWithStdlibErrnos(t *testing.T) {
	fsys, mem := newStdlibFS(t)
	lib := stdlibMount + "/"
	f, err := fsys.OpenFile(lib+"strict.pm", os.O_RDONLY|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatalf("O_CREAT on an existing library file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	for what, c := range map[string]struct {
		err  error
		want error
	}{
		"open File for writing":   {openErr(fsys, lib+"File", os.O_WRONLY), syscall.EISDIR},
		"unlink File":             {fsys.Unlink(lib + "File"), syscall.EISDIR},
		"stat strict.pm/":         {statErr(fsys, lib+"strict.pm/"), syscall.ENOTDIR},
		"open strict.pm/":         {openErr(fsys, lib+"strict.pm/", os.O_RDONLY), syscall.ENOTDIR},
		"create in a missing dir": {openErr(fsys, lib+"No/New.pm", os.O_WRONLY|os.O_CREATE), iofs.ErrPermission},
	} {
		if !errors.Is(c.err, c.want) {
			t.Errorf("%s: err = %v, want %v", what, c.err, c.want)
		}
	}
	if err := mem.WriteFile("app.pl", nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := statErr(fsys, "app.pl/"); !errors.Is(err, syscall.ENOTDIR) {
		t.Errorf("stat app.pl/: err = %v, want ENOTDIR", err)
	}
	if err := statErr(fsys, lib+"File/"); err != nil {
		t.Errorf("stat File/: %v", err)
	}
}

func openErr(fsys FS, name string, flag int) error {
	f, err := fsys.OpenFile(name, flag, 0o644)
	if err == nil {
		err = f.Close()
		if err == nil {
			return errors.New("opened")
		}
	}
	return err
}

func statErr(fsys FS, name string) error {
	_, err := fsys.Stat(name)
	return err
}

func TestWithStdlibFileOffsets(t *testing.T) {
	fsys, _ := newStdlibFS(t)
	want := zipContent(t, "strict.pm")
	f, err := fsys.OpenFile(stdlibMount+"/strict.pm", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Seek(-1, io.SeekStart); err == nil {
		t.Error("seek before the start succeeded")
	}
	if _, err := f.ReadAt(make([]byte, 1), -1); err == nil {
		t.Error("ReadAt a negative offset succeeded")
	}
	if off, err := f.Seek(-3, io.SeekEnd); err != nil || off != int64(len(want)-3) {
		t.Fatalf("seek from the end = %d, %v", off, err)
	}
	got, err := io.ReadAll(f)
	if err != nil || !bytes.Equal(got, want[len(want)-3:]) {
		t.Fatalf("read after seeking from the end = %q, %v", got, err)
	}
	if n, err := f.Read(make([]byte, 1)); n != 0 || err != io.EOF {
		t.Fatalf("read at the end = %d, %v", n, err)
	}
}

// The module tree the go-perl/cpan package carries is mounted the same way,
// beside the standard library.
func TestWithStdlibMountsTheCPANTree(t *testing.T) {
	internal.SetCPANLib(fstest.MapFS{
		"Greeting.pm":  {Data: []byte("package Greeting;\n1;\n"), Mode: 0o644},
		"script/greet": {Data: []byte("#!perl\n"), Mode: 0o755},
	})
	t.Cleanup(func() { internal.SetCPANLib(nil) })
	fsys, _ := newStdlibFS(t)
	if got := readFile(t, fsys, cpanLibDir+"/Greeting.pm"); string(got) != "package Greeting;\n1;\n" {
		t.Fatalf("Greeting.pm = %q", got)
	}
	fi, err := fsys.Stat(cpanLibDir + "/script/greet")
	if err != nil || fi.Mode() != 0o555 {
		t.Fatalf("stat script/greet = %v, %v; want mode 0555", fi, err)
	}
	if err := fsys.Unlink(cpanLibDir + "/Greeting.pm"); !errors.Is(err, iofs.ErrPermission) {
		t.Errorf("unlink Greeting.pm: %v", err)
	}
	if got := dirNames(t, fsys, goperlMount); len(got) != 2 || got[0] != "cpan" || got[1] != "stdlib" {
		t.Errorf("listing %s = %v, want [cpan stdlib]", goperlMount, got)
	}
	if inc, err := cpanLibInc(fsys); err != nil || inc != "/"+cpanLibDir {
		t.Errorf("cpanLibInc = %q, %v", inc, err)
	}
}

func TestWithStdlibRefusesAnOccupiedMountPoint(t *testing.T) {
	mem := goperlfs.NewMemFS()
	if err := mem.WriteFile(stdlibMount+"/strict.pm", []byte("shadowed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := WithStdlib(mem); err == nil {
		t.Fatal("WithStdlib mounted over files the MemFS holds")
	}
}

// Mounting the library must not copy it: per FS it should cost a handful of
// directories, not the ~55 MB of file contents.
func TestWithStdlibSharesTheLibrary(t *testing.T) {
	newStdlibFS(t) // decompress the library
	const n = 10
	keep := make([]FS, 0, n)
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	for range n {
		fsys, _ := newStdlibFS(t)
		keep = append(keep, fsys)
	}
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(keep)
	if perFS := (after.TotalAlloc - before.TotalAlloc) / n; perFS > 64<<10 {
		t.Fatalf("WithStdlib allocated %d bytes per call", perFS)
	}
}
