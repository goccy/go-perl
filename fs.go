package perl

// Filesystem plumbing.
//
// An instance opens every file through an FS backend (Config.FS); the
// backends themselves live in the go-perl/fs package (fs.NewMemFS,
// fs.NewHostFS, fs.DirFS). This file re-exports the types Config needs and
// wraps every backend that has no /dev/null with a working one (perl
// requires one to boot); stdlibfs.go mounts the standard library for the
// library default.

import (
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"syscall"
	"time"

	goperlfs "github.com/goccy/go-perl/fs"
	"github.com/goccy/perlwasm2go/base"
)

// FS is the read/write filesystem backend a Perl instance is given via
// Config.FS. It is a write-capable superset of io/fs.FS; implementations
// live in the go-perl/fs package.
type FS = goperlfs.FS

// File is an open file returned by FS.OpenFile.
type File = goperlfs.File

// MemFS is an in-memory read/write FS (see the go-perl/fs package).
type MemFS = goperlfs.MemFS

// withDevNull returns fsys as it is when it has a /dev/null of its own (the
// host's filesystem has), and otherwise fsys wrapped so the guest sees a
// working one. Perl itself requires one: a `-e` bootstrap (which perl_new
// uses) opens the bit bucket as its script filehandle, so booting on an FS
// without it fails outright. The wrapper's reads hit EOF, its writes are
// discarded, and everything else passes through to the wrapped backend.
// The host's filesystem is not wrapped, so that the runtime, which tells it
// by its type, treats it as the host's (in how the instance's umask is
// applied, access checks, locks and Unix-domain sockets).
func withDevNull(fsys FS) (FS, error) {
	fi, err := fsys.Stat(devNullName)
	switch {
	case err == nil && fi.Mode()&iofs.ModeCharDevice != 0:
		return fsys, nil
	case err == nil || errors.Is(err, iofs.ErrNotExist):
		return devNullFS{FS: fsys}, nil
	}
	return nil, fmt.Errorf("look up /dev/null: %w", err)
}

type devNullFS struct{ FS }

// Flock grants every lock on the bit bucket — it has no content for a lock
// to protect — and leaves the files of the wrapped backend to that backend.
// A backend that takes no locks (not a LockFS) fails the lock with ENOSYS,
// as the runtime does for it unwrapped.
func (d devNullFS) Flock(f File, how int) error {
	if _, ok := f.(nullFile); ok {
		return nil
	}
	locks, can := d.FS.(base.LockFS)
	if !can {
		return syscall.ENOSYS
	}
	return locks.Flock(f, how)
}

// Access lets anyone read and write the bit bucket, and leaves the files of
// the wrapped backend to that backend when it decides access itself
// (base.AccessFS); otherwise the runtime answers by the file's mode.
func (d devNullFS) Access(name string, mode uint32, effective bool) error {
	if name == devNullName || name == "dev" {
		return nil
	}
	decides, can := d.FS.(base.AccessFS)
	if !can {
		return errors.ErrUnsupported
	}
	return decides.Access(name, mode, effective)
}

// Mkfifo and Chown pass what the host filesystem can do beyond FS — make a
// FIFO, change a file's owner — on to the wrapped backend. A backend
// without them (an in-memory one) reports errors.ErrUnsupported, which the
// guest sees as a filesystem that has no FIFOs and keeps no owners.
func (d devNullFS) Mkfifo(name string, perm iofs.FileMode) error {
	maker, can := d.FS.(interface {
		Mkfifo(name string, perm iofs.FileMode) error
	})
	if !can {
		return errors.ErrUnsupported
	}
	return maker.Mkfifo(name, perm)
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
		return nullFileInfo{name: "null", mode: devNullMode}, nil
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

// devNullMode is the mode of the guest's /dev/null: a character device
// anyone may read and write, as on a real system.
const devNullMode = iofs.ModeDevice | iofs.ModeCharDevice | 0o666

// nullFile is the guest's /dev/null: an always-empty sink.
type nullFile struct{}

func (nullFile) Read(p []byte) (int, error)               { return 0, io.EOF }
func (nullFile) ReadAt(p []byte, off int64) (int, error)  { return 0, io.EOF }
func (nullFile) Write(p []byte) (int, error)              { return len(p), nil }
func (nullFile) WriteAt(p []byte, off int64) (int, error) { return len(p), nil }
func (nullFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}
func (nullFile) Close() error             { return nil }
func (nullFile) Chown(uid, gid int) error { return nil }
func (nullFile) Stat() (os.FileInfo, error) {
	return nullFileInfo{name: "null", mode: devNullMode}, nil
}
func (nullFile) ReadDir(n int) ([]os.DirEntry, error) {
	return nil, fmt.Errorf("dev/null is not a directory")
}
func (nullFile) Sync() error                    { return nil }
func (nullFile) Truncate(size int64) error      { return nil }
func (nullFile) Chmod(mode iofs.FileMode) error { return nil }
func (nullFile) Name() string                   { return "null" }

type nullFileInfo struct {
	name string
	mode iofs.FileMode
}

func (fi nullFileInfo) Name() string        { return fi.name }
func (fi nullFileInfo) Size() int64         { return 0 }
func (fi nullFileInfo) Mode() iofs.FileMode { return fi.mode }
func (fi nullFileInfo) ModTime() time.Time  { return time.Time{} }
func (fi nullFileInfo) IsDir() bool         { return fi.mode.IsDir() }

// Sys describes the null device as the host does where the host has one:
// its device numbers and owner are what a program reads to tell a special
// file from an ordinary one (the st_rdev of stat). The synthetic directory
// above it has nothing of the host's.
func (fi nullFileInfo) Sys() any {
	if fi.mode&iofs.ModeCharDevice == 0 {
		return nil
	}
	host, err := os.Stat(os.DevNull)
	if err != nil || host.Mode()&iofs.ModeCharDevice == 0 {
		return nil
	}
	return host.Sys()
}
