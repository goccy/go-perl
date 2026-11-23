// Package fs provides the filesystem backends a Perl instance runs on.
//
// Every file the interpreter opens goes through one FS value (perl.Config.FS):
//
//   - NewMemFS gives a private in-memory filesystem — nothing touches the
//     host disk, and separate MemFS values are fully isolated (this backs
//     the library default, pre-loaded with the standard library);
//   - NewHostFS passes straight through to the operating system's
//     filesystem — how the perl command behaves, and what gperl uses;
//   - DirFS is the host filesystem scoped to one directory, which becomes
//     the guest's root;
//   - any other implementation of FS plugs in the same way.
package fs

import "github.com/goccy/perlwasm2go/base"

// FS is the read/write filesystem backend a Perl instance is given via
// perl.Config.FS. It is a write-capable superset of io/fs.FS. Names are
// guest paths relative to the backend's root: slash-separated with no
// leading slash ("" is the root). Methods should return the standard
// io/fs errors so the runtime maps them to the right guest errno.
type FS = base.FS

// File is an open file returned by FS.OpenFile.
type File = base.File

// MemFS is an in-memory read/write FS. Separate MemFS values are fully
// isolated from one another. Safe for concurrent use.
type MemFS = base.MemFS

// NewMemFS returns an empty in-memory filesystem.
func NewMemFS() *MemFS { return base.NewMemFS() }

// NewHostFS returns the operating system's whole filesystem: the guest sees
// the host's real root, so absolute paths mean the same thing inside and
// outside the instance. This is how the perl command behaves (and how gperl
// runs scripts).
//
// It is the runtime's own backend for the operating system's filesystem, so
// everything the host provides — file modes and owners, link counts, the
// three file times, flock — reaches the guest.
func NewHostFS() FS {
	fsys, err := base.HostFS("/")
	if err != nil {
		// HostFS opens nothing for the whole filesystem, so it has no error
		// to give; one would be a broken runtime.
		panic(err)
	}
	return fsys
}

// DirFS returns the host filesystem scoped to dir: dir becomes the guest's
// root, and nothing outside it is reachable, through ".." or through a
// symbolic link (it is built on os.OpenRoot, whose error it returns). The
// FS holds dir open until its Close method (it is an io.Closer) is called,
// which the caller does once no instance uses it any more.
func DirFS(dir string) (FS, error) { return base.HostFS(dir) }
