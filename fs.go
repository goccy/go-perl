package perl

// Filesystem backend API.
//
// An Interpreter opens every file through an FS backend (Config.FS). By
// default that is the host filesystem; supplying a custom FS — e.g. an
// in-memory MemFS — gives the interpreter a private, arbitrary filesystem so
// its reads/writes never touch disk and are invisible to other interpreters.
// These are re-exports of the generic backend defined in the wasm2go runtime
// so callers only need this package.

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"time"

	"github.com/goccy/perlwasm2go/base"
)

// FS is the read/write filesystem backend a Perl instance is given via
// Config.FS. It is a write-capable superset of io/fs.FS.
type FS = base.FS

// File is an open file returned by FS.OpenFile.
type File = base.File

// MemFS is an in-memory read/write FS. Separate MemFS values are fully
// isolated from one another.
type MemFS = base.MemFS

// NewMemFS returns an empty in-memory filesystem.
func NewMemFS() *MemFS { return base.NewMemFS() }

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

// NewStdlibMemFS returns an in-memory filesystem pre-loaded with the embedded
// Perl standard library at the root, ready to back an Interpreter:
//
//	fs := perl.NewStdlibMemFS()
//	interp, _ := perl.NewInterpreter(perl.Config{FS: fs}) // StdlibDir = "/"
//
// Each call returns an independent FS, so interpreters built from separate
// NewStdlibMemFS() values share no filesystem state.
func NewStdlibMemFS() (*MemFS, error) {
	zr, err := zip.NewReader(bytes.NewReader(stdlibZip), int64(len(stdlibZip)))
	if err != nil {
		return nil, fmt.Errorf("open embedded stdlib: %w", err)
	}
	fsys := NewMemFS()
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			if err := fsys.MkdirAll(f.Name, 0o755); err != nil {
				return nil, err
			}
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(rc); err != nil {
			rc.Close()
			return nil, err
		}
		rc.Close()
		if err := fsys.WriteFile(f.Name, buf.Bytes(), 0o644); err != nil {
			return nil, err
		}
	}
	return fsys, nil
}
