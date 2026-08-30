package perl

// Filesystem plumbing.
//
// An instance opens every file through an FS backend (Config.FS); the
// backends themselves live in the go-perl/fs package (fs.NewMemFS,
// fs.NewHostFS, fs.DirFS). This file re-exports the types Config needs,
// provides NewStdlibMemFS (the library default's filesystem), and wraps
// every backend with a working /dev/null (perl requires one to boot).

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"sync"
	"time"

	goperlfs "github.com/goccy/go-perl/fs"
)

// FS is the read/write filesystem backend a Perl instance is given via
// Config.FS. It is a write-capable superset of io/fs.FS; implementations
// live in the go-perl/fs package.
type FS = goperlfs.FS

// File is an open file returned by FS.OpenFile.
type File = goperlfs.File

// MemFS is an in-memory read/write FS (see the go-perl/fs package).
type MemFS = goperlfs.MemFS

// withDevNull wraps a custom FS backend so the guest always sees a working
// /dev/null. Perl itself requires one: a `-e` bootstrap (which perl_new uses)
// opens the bit bucket as its script filehandle, so booting on an FS without
// it fails outright. Reads hit EOF, writes are discarded, and everything
// else passes through to the wrapped backend (whose own dev/null, if any, is
// shadowed for consistency).
func withDevNull(fsys FS) FS { return devNullFS{FS: fsys} }

type devNullFS struct{ FS }

const devNullName = "dev/null"

func (d devNullFS) OpenFile(name string, flag int, perm iofs.FileMode) (File, error) {
	if name == devNullName {
		return nullFile{}, nil
	}
	return d.FS.OpenFile(name, flag, perm)
}

func (d devNullFS) Stat(name string) (os.FileInfo, error) {
	switch name {
	case devNullName:
		return nullFileInfo{name: "null", mode: 0o666}, nil
	case "dev":
		// The synthetic parent directory, so path resolution can traverse it
		// even when the wrapped FS has no dev/ entry.
		if fi, err := d.FS.Stat(name); err == nil {
			return fi, nil
		}
		return nullFileInfo{name: "dev", mode: iofs.ModeDir | 0o755}, nil
	}
	return d.FS.Stat(name)
}

func (d devNullFS) Lstat(name string) (os.FileInfo, error) {
	if name == devNullName || name == "dev" {
		return d.Stat(name)
	}
	return d.FS.Lstat(name)
}

// nullFile is the guest's /dev/null: an always-empty sink.
type nullFile struct{}

func (nullFile) Read(p []byte) (int, error)               { return 0, io.EOF }
func (nullFile) ReadAt(p []byte, off int64) (int, error)  { return 0, io.EOF }
func (nullFile) Write(p []byte) (int, error)              { return len(p), nil }
func (nullFile) WriteAt(p []byte, off int64) (int, error) { return len(p), nil }
func (nullFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}
func (nullFile) Close() error               { return nil }
func (nullFile) Stat() (os.FileInfo, error) { return nullFileInfo{name: "null", mode: 0o666}, nil }
func (nullFile) ReadDir(n int) ([]os.DirEntry, error) {
	return nil, fmt.Errorf("dev/null is not a directory")
}
func (nullFile) Sync() error               { return nil }
func (nullFile) Truncate(size int64) error { return nil }
func (nullFile) Name() string              { return "null" }

type nullFileInfo struct {
	name string
	mode iofs.FileMode
}

func (fi nullFileInfo) Name() string        { return fi.name }
func (fi nullFileInfo) Size() int64         { return 0 }
func (fi nullFileInfo) Mode() iofs.FileMode { return fi.mode }
func (fi nullFileInfo) ModTime() time.Time  { return time.Time{} }
func (fi nullFileInfo) IsDir() bool         { return fi.mode.IsDir() }
func (fi nullFileInfo) Sys() any            { return nil }

// stdlibEntries caches the DECOMPRESSED embedded stdlib once per process:
// building a MemFS per instance then only pays for copying bytes in
// (MemFS.WriteFile copies, so instances stay fully isolated).
var (
	stdlibEntriesOnce sync.Once
	stdlibEntries     []stdlibEntry
	stdlibEntriesErr  error
)

type stdlibEntry struct {
	name string
	dir  bool
	data []byte
}

func loadStdlibEntries() ([]stdlibEntry, error) {
	stdlibEntriesOnce.Do(func() {
		zr, err := zip.NewReader(bytes.NewReader(stdlibZip), int64(len(stdlibZip)))
		if err != nil {
			stdlibEntriesErr = fmt.Errorf("open embedded stdlib: %w", err)
			return
		}
		for _, f := range zr.File {
			if f.FileInfo().IsDir() {
				stdlibEntries = append(stdlibEntries, stdlibEntry{name: f.Name, dir: true})
				continue
			}
			rc, err := f.Open()
			if err != nil {
				stdlibEntriesErr = err
				return
			}
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(rc); err != nil {
				rc.Close()
				stdlibEntriesErr = err
				return
			}
			rc.Close()
			stdlibEntries = append(stdlibEntries, stdlibEntry{name: f.Name, data: buf.Bytes()})
		}
	})
	return stdlibEntries, stdlibEntriesErr
}

// NewStdlibMemFS returns an in-memory filesystem pre-loaded with the embedded
// Perl standard library at the root, ready to back an Interpreter:
//
//	fs := perl.NewStdlibMemFS()
//	interp, _ := perl.NewInterpreter(perl.Config{FS: fs}) // StdlibDir = "/"
//
// Each call returns an independent FS, so interpreters built from separate
// NewStdlibMemFS() values share no filesystem state. This is also the
// filesystem an instance gets by default (Config with a nil FS).
func NewStdlibMemFS() (*MemFS, error) {
	entries, err := loadStdlibEntries()
	if err != nil {
		return nil, err
	}
	fsys := goperlfs.NewMemFS()
	for _, e := range entries {
		if e.dir {
			if err := fsys.MkdirAll(e.name, 0o755); err != nil {
				return nil, err
			}
			continue
		}
		if err := fsys.WriteFile(e.name, e.data, 0o644); err != nil {
			return nil, err
		}
	}
	return fsys, nil
}
