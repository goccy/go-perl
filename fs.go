package perl

// Filesystem plumbing.
//
// An instance opens every file through an FS backend (Config.FS); the
// backends themselves live in the go-perl/fs package (fs.NewMemFS,
// fs.NewHostFS, fs.DirFS). This file re-exports the types Config needs and
// wraps every backend with a working /dev/null (perl requires one to boot);
// stdlibfs.go mounts the standard library for the library default.

import (
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"syscall"
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

// Chmod forwards to the wrapped backend when it supports changing modes
// (the runtime discovers the capability by interface assertion, which an
// embedded interface field would otherwise hide).
func (d devNullFS) Chmod(name string, mode iofs.FileMode) error {
	if ch, ok := d.FS.(interface {
		Chmod(string, iofs.FileMode) error
	}); ok {
		return ch.Chmod(name, mode)
	}
	return &iofs.PathError{Op: "chmod", Path: name, Err: syscall.ENOSYS}
}

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
