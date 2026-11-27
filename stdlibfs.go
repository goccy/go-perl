package perl

// The standard library as a shared, read-only mount.
//
// Every instance on the default filesystem needs the whole standard library
// (and the module tree the go-perl/cpan package carries, when it carries
// one), and no instance needs to change them. So each tree is decompressed
// once per process into an immutable tree, and every instance mounts those
// same trees read-only, side by side in /.goperl, on top of its own MemFS
// (as on the host filesystem they sit side by side in go-perl's directory
// of the user's cache): they cost memory once per process, not once per
// instance, and an instance's writes all land in its private MemFS.

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/goccy/go-perl/internal"
	"github.com/goccy/perlwasm2go/base"
)

// StdlibDir is where WithStdlib mounts the standard library: the default
// Config.StdlibDir.
const StdlibDir = "/" + stdlibMount

// goperlMount is the directory, as an FS name, WithStdlib mounts read-only,
// and stdlibMount the standard library's directory in it (StdlibDir as an
// FS name); the CPAN module tree is cpanLibDir.
const (
	goperlMount = ".goperl"
	stdlibMount = goperlMount + "/stdlib"
)

// WithStdlib returns an FS that serves fsys with the embedded standard
// library mounted read-only at StdlibDir (and the module tree the
// go-perl/cpan package carries, when it carries one, beside it). It is what
// a nil Config.FS defaults to (on an empty MemFS); build one explicitly to
// add files of your own:
//
//	mem := fs.NewMemFS()
//	mem.WriteFile("app.pl", src, 0o644)
//	fsys, _ := perl.WithStdlib(mem)
//	p, _ := perl.New(perl.Config{FS: fsys})
//
// The mounted directory, /.goperl, is shared by every FS WithStdlib returns
// and cannot be modified: writing, creating, removing, renaming or changing
// the metadata of anything in it fails with a permission error. Everything
// else is fsys, so instances given separate MemFS values share no writable
// state. WithStdlib refuses a fsys that already holds files in /.goperl.
func WithStdlib(fsys *MemFS) (FS, error) {
	lib, err := loadStdlibTree()
	if err != nil {
		return nil, err
	}
	root := &roNode{info: newInfo(goperlMount, os.ModeDir|0o555, 0, time.Time{}), children: map[string]*roNode{lib.root.info.name: lib.root}}
	if src, set := internal.CPANLib(); src != nil {
		cpan, err := loadCPANTree(src, set)
		if err != nil {
			return nil, err
		}
		root.children[cpan.root.info.name] = cpan.root
	}
	for name := range root.children {
		root.names = append(root.names, name)
	}
	slices.Sort(root.names)
	// The mount point exists in fsys, so it shows up when the guest lists
	// the root.
	if err := fsys.MkdirAll(goperlMount, 0o755); err != nil {
		return nil, fmt.Errorf("mount /%s: %w", goperlMount, err)
	}
	if err := ensureEmptyDir(fsys, goperlMount); err != nil {
		return nil, err
	}
	return &mountFS{upper: fsys, tree: &roTree{root: root}}, nil
}

// ensureEmptyDir fails when the directory name of fsys holds anything: the
// mount would hide it.
func ensureEmptyDir(fsys *MemFS, name string) (err error) {
	f, err := fsys.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	ents, err := f.ReadDir(1)
	if err != nil && err != io.EOF {
		return err
	}
	if len(ents) > 0 {
		return fmt.Errorf("mount /%s: the directory already holds files", name)
	}
	return nil
}

// mountFS is upper with tree mounted read-only at goperlMount.
type mountFS struct {
	upper *MemFS
	tree  *roTree
}

// route cleans name and reports whether the tree serves it (nil: upper),
// with the name that backend should see.
func (m *mountFS) route(name string) (string, *roTree) {
	c := strings.TrimPrefix(path.Clean("/"+name), "/")
	// Cleaning drops a trailing slash, which asks for a directory; it goes
	// back on.
	slash := ""
	if strings.HasSuffix(name, "/") && c != "" {
		slash = "/"
	}
	if c == goperlMount {
		return slash, m.tree
	}
	if rest, ok := strings.CutPrefix(c, goperlMount+"/"); ok {
		return rest + slash, m.tree
	}
	return c + slash, nil
}

func readOnly(op, name string) error {
	return &iofs.PathError{Op: op, Path: name, Err: iofs.ErrPermission}
}

// inTree is the error of a metadata change name asks for in a mounted tree:
// the file's own error when it does not exist, a permission error when it
// does.
func inTree(t *roTree, rel, op, name string) error {
	if _, err := t.lookup(rel); err != nil {
		return &iofs.PathError{Op: op, Path: name, Err: err}
	}
	return readOnly(op, name)
}

func (m *mountFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	rel, t := m.route(name)
	if t != nil {
		return t.open(rel, name, flag)
	}
	return m.upper.OpenFile(rel, flag, perm)
}

func (m *mountFS) Mkdir(name string, perm os.FileMode) error {
	rel, t := m.route(name)
	if t != nil {
		if _, err := t.lookup(rel); err == nil {
			return &iofs.PathError{Op: "mkdir", Path: name, Err: iofs.ErrExist}
		}
		return readOnly("mkdir", name)
	}
	return m.upper.Mkdir(rel, perm)
}

func (m *mountFS) Chmod(name string, mode os.FileMode) error {
	rel, t := m.route(name)
	if t != nil {
		return inTree(t, rel, "chmod", name)
	}
	return m.upper.Chmod(rel, mode)
}

func (m *mountFS) Chown(name string, uid, gid int) error {
	rel, t := m.route(name)
	if t != nil {
		return inTree(t, rel, "chown", name)
	}
	return m.upper.Chown(rel, uid, gid)
}

func (m *mountFS) Lchown(name string, uid, gid int) error {
	rel, t := m.route(name)
	if t != nil {
		return inTree(t, rel, "lchown", name)
	}
	return m.upper.Lchown(rel, uid, gid)
}

func (m *mountFS) Chtimes(name string, atime, mtime time.Time) error {
	rel, t := m.route(name)
	if t != nil {
		return inTree(t, rel, "chtimes", name)
	}
	return m.upper.Chtimes(rel, atime, mtime)
}

func (m *mountFS) Lchtimes(name string, atime, mtime time.Time) error {
	rel, t := m.route(name)
	if t != nil {
		return inTree(t, rel, "lchtimes", name)
	}
	return m.upper.Lchtimes(rel, atime, mtime)
}

func (m *mountFS) Unlink(name string) error {
	rel, t := m.route(name)
	if t != nil {
		if n, err := t.lookup(rel); err == nil && n.info.IsDir() {
			return &iofs.PathError{Op: "unlink", Path: name, Err: syscall.EISDIR}
		}
		return inTree(t, rel, "unlink", name)
	}
	return m.upper.Unlink(rel)
}

func (m *mountFS) Rmdir(name string) error {
	rel, t := m.route(name)
	if t != nil {
		return inTree(t, rel, "rmdir", name)
	}
	return m.upper.Rmdir(rel)
}

func (m *mountFS) Rename(oldName, newName string) error {
	oldRel, oldTree := m.route(oldName)
	newRel, newTree := m.route(newName)
	if oldTree != nil || newTree != nil {
		return readOnly("rename", oldName)
	}
	return m.upper.Rename(oldRel, newRel)
}

func (m *mountFS) Stat(name string) (os.FileInfo, error) {
	rel, t := m.route(name)
	if t != nil {
		n, err := t.lookup(rel)
		if err != nil {
			return nil, &iofs.PathError{Op: "stat", Path: name, Err: err}
		}
		return n.info, nil
	}
	return m.upper.Stat(rel)
}

// A mounted tree has no links, so Lstat is Stat there.
func (m *mountFS) Lstat(name string) (os.FileInfo, error) {
	if rel, t := m.route(name); t == nil {
		return m.upper.Lstat(rel)
	}
	return m.Stat(name)
}

func (m *mountFS) Symlink(oldName, newName string) error {
	rel, t := m.route(newName)
	if t != nil {
		return readOnly("symlink", newName)
	}
	return m.upper.Symlink(oldName, rel)
}

func (m *mountFS) Link(oldName, newName string) error {
	oldRel, oldTree := m.route(oldName)
	newRel, newTree := m.route(newName)
	if oldTree != nil || newTree != nil {
		return readOnly("link", newName)
	}
	return m.upper.Link(oldRel, newRel)
}

func (m *mountFS) Readlink(name string) (string, error) {
	rel, t := m.route(name)
	if t != nil {
		return "", &iofs.PathError{Op: "readlink", Path: name, Err: iofs.ErrInvalid}
	}
	return m.upper.Readlink(rel)
}

func (m *mountFS) Close() error { return m.upper.Close() }

// Access answers for the mounted trees by their modes, which no write is
// allowed through, and leaves the rest to upper (base.AccessFS).
func (m *mountFS) Access(name string, mode uint32, effective bool) error {
	rel, t := m.route(name)
	if t == nil {
		return m.upper.Access(rel, mode, effective)
	}
	n, err := t.lookup(rel)
	if err != nil {
		return &iofs.PathError{Op: "access", Path: name, Err: err}
	}
	perm := n.info.mode.Perm()
	if mode&2 != 0 || (mode&4 != 0 && perm&0o444 == 0) || (mode&1 != 0 && perm&0o111 == 0) {
		return readOnly("access", name)
	}
	return nil
}

// Flock grants every lock on a file of a mounted tree — nothing can change
// what a lock would protect — and leaves the files of upper to upper
// (base.LockFS).
func (m *mountFS) Flock(f File, how int) error {
	if _, ok := f.(*roFile); ok {
		return nil
	}
	return m.upper.Flock(f, how)
}

var _ interface {
	base.AccessFS
	base.LockFS
} = (*mountFS)(nil)

// roTree is a decompressed read-only tree. It is built once per process
// and never modified afterwards, so every instance reads it without
// locking.
type roTree struct {
	root *roNode
}

type roNode struct {
	info     roInfo
	data     []byte
	children map[string]*roNode
	names    []string // children's names, sorted, for ReadDir
}

// lookup finds name, which must be a directory when it ends in a slash.
func (t *roTree) lookup(name string) (*roNode, error) {
	dir := strings.HasSuffix(name, "/")
	name = strings.TrimSuffix(name, "/")
	n := t.root
	if name != "" {
		for _, part := range strings.Split(name, "/") {
			if !n.info.IsDir() {
				return nil, syscall.ENOTDIR
			}
			c, ok := n.children[part]
			if !ok {
				return nil, iofs.ErrNotExist
			}
			n = c
		}
	}
	if dir && !n.info.IsDir() {
		return nil, syscall.ENOTDIR
	}
	return n, nil
}

// open opens rel as open(2) does on a read-only mount: an existing file
// opens for reading (O_CREAT without O_EXCL included), anything that would
// write or create fails.
func (t *roTree) open(rel, name string, flag int) (File, error) {
	n, err := t.lookup(rel)
	write := flag&(os.O_WRONLY|os.O_RDWR|os.O_TRUNC|os.O_APPEND) != 0
	switch {
	case err == nil && flag&(os.O_CREATE|os.O_EXCL) == os.O_CREATE|os.O_EXCL:
		return nil, &iofs.PathError{Op: "open", Path: name, Err: iofs.ErrExist}
	case err == nil && write && n.info.IsDir():
		return nil, &iofs.PathError{Op: "open", Path: name, Err: syscall.EISDIR}
	case err == nil && write, errors.Is(err, iofs.ErrNotExist) && flag&os.O_CREATE != 0:
		return nil, readOnly("open", name)
	case err != nil:
		return nil, &iofs.PathError{Op: "open", Path: name, Err: err}
	}
	return &roFile{node: n}, nil
}

var (
	stdlibTreeOnce sync.Once
	stdlibTreeVal  *roTree
	stdlibTreeErr  error

	cpanTreeMu  sync.Mutex
	cpanTreeSet uint64 // the internal.CPANLib number cpanTreeVal was built for; 0: none
	cpanTreeVal *roTree
)

// loadStdlibTree decompresses stdlib.zip once per process.
func loadStdlibTree() (*roTree, error) {
	stdlibTreeOnce.Do(func() {
		zr, err := zip.NewReader(bytes.NewReader(stdlibZip), int64(len(stdlibZip)))
		if err != nil {
			stdlibTreeErr = fmt.Errorf("open embedded stdlib: %w", err)
			return
		}
		stdlibTreeVal, stdlibTreeErr = buildTree(zr, path.Base(stdlibMount))
		if stdlibTreeErr != nil {
			stdlibTreeErr = fmt.Errorf("embedded stdlib: %w", stdlibTreeErr)
		}
	})
	return stdlibTreeVal, stdlibTreeErr
}

// loadCPANTree reads src, the module tree the go-perl/cpan package carries
// (internal.CPANLib), once for every time one is installed (set).
func loadCPANTree(src iofs.FS, set uint64) (*roTree, error) {
	cpanTreeMu.Lock()
	defer cpanTreeMu.Unlock()
	if cpanTreeSet == set {
		return cpanTreeVal, nil
	}
	t, err := buildTree(src, path.Base(cpanLibDir))
	if err != nil {
		return nil, fmt.Errorf("the CPAN module tree: %w", err)
	}
	cpanTreeSet, cpanTreeVal = set, t
	return t, nil
}

// buildTree reads src whole into a tree whose root is named rootName: its
// files read-only with the rest of their permission bits, its directories
// 0o555.
func buildTree(src iofs.FS, rootName string) (*roTree, error) {
	newDir := func(name string, modTime time.Time) *roNode {
		return &roNode{info: newInfo(name, os.ModeDir|0o555, 0, modTime), children: map[string]*roNode{}}
	}
	root := newDir(rootName, time.Time{})
	nodes := map[string]*roNode{".": root}
	err := iofs.WalkDir(src, ".", func(name string, d iofs.DirEntry, err error) error {
		if err != nil || name == "." {
			return err
		}
		parent := nodes[path.Dir(name)]
		info, err := d.Info()
		if err != nil {
			return err
		}
		var n *roNode
		switch {
		case d.IsDir():
			n = newDir(d.Name(), info.ModTime())
			nodes[name] = n
		case d.Type().IsRegular():
			data, err := iofs.ReadFile(src, name)
			if err != nil {
				return err
			}
			n = &roNode{info: newInfo(d.Name(), info.Mode().Perm()&^0o222, int64(len(data)), info.ModTime()), data: data}
		default:
			return fmt.Errorf("%s: %v is neither a file nor a directory", name, d.Type())
		}
		parent.children[d.Name()] = n
		parent.names = append(parent.names, d.Name()) // WalkDir visits in lexical order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &roTree{root: root}, nil
}

// treeInos numbers the nodes of every tree, so that no two files of the
// mounts share an inode number (Sys).
var treeInos atomic.Uint64

// treeDev is the device the files of the mounts are on (Sys): not the
// device of a MemFS file, which is 1, so that a file of the mounts and one
// of the instance's MemFS never share a device and inode.
const treeDev = 2

// treeBlockSize is the block size a file of the mounts reports for I/O, as
// a MemFS file does.
const treeBlockSize = 4096

// newInfo describes a new node; a zero modTime (a directory an archive does
// not list) is the epoch.
func newInfo(name string, mode os.FileMode, size int64, modTime time.Time) roInfo {
	if modTime.IsZero() {
		modTime = time.Unix(0, 0)
	}
	return roInfo{name: name, ino: treeInos.Add(1), size: size, mode: mode, modTime: modTime}
}

type roInfo struct {
	name    string
	ino     uint64
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi roInfo) Name() string       { return fi.name }
func (fi roInfo) Size() int64        { return fi.size }
func (fi roInfo) Mode() os.FileMode  { return fi.mode }
func (fi roInfo) ModTime() time.Time { return fi.modTime }
func (fi roInfo) IsDir() bool        { return fi.mode.IsDir() }

type roDirEntry struct{ info roInfo }

func (e roDirEntry) Name() string               { return e.info.name }
func (e roDirEntry) IsDir() bool                { return e.info.IsDir() }
func (e roDirEntry) Type() os.FileMode          { return e.info.mode.Type() }
func (e roDirEntry) Info() (os.FileInfo, error) { return e.info, nil }

// roFile is an open, read-only handle on a tree node. The node is shared
// and immutable; only the handle's own offsets change.
type roFile struct {
	node   *roNode
	mu     sync.Mutex
	off    int64
	dirOff int
}

func (f *roFile) Name() string               { return f.node.info.name }
func (f *roFile) Close() error               { return nil }
func (f *roFile) Sync() error                { return nil }
func (f *roFile) Stat() (os.FileInfo, error) { return f.node.info, nil }

func (f *roFile) Write([]byte) (int, error)          { return 0, readOnly("write", f.Name()) }
func (f *roFile) WriteAt([]byte, int64) (int, error) { return 0, readOnly("write", f.Name()) }
func (f *roFile) Truncate(int64) error               { return readOnly("truncate", f.Name()) }
func (f *roFile) Chmod(os.FileMode) error            { return readOnly("chmod", f.Name()) }
func (f *roFile) Chown(int, int) error               { return readOnly("chown", f.Name()) }

func (f *roFile) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, err := f.readAt(p, f.off)
	f.off += int64(n)
	if err == io.EOF && n > 0 {
		err = nil
	}
	return n, err
}

func (f *roFile) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, &iofs.PathError{Op: "readat", Path: f.Name(), Err: iofs.ErrInvalid}
	}
	return f.readAt(p, off)
}

func (f *roFile) readAt(p []byte, off int64) (int, error) {
	if f.node.info.IsDir() {
		return 0, &iofs.PathError{Op: "read", Path: f.Name(), Err: syscall.EISDIR}
	}
	if off >= int64(len(f.node.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.node.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (f *roFile) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var base int64
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		base = f.off
	case io.SeekEnd:
		base = int64(len(f.node.data))
	default:
		return 0, &iofs.PathError{Op: "seek", Path: f.Name(), Err: iofs.ErrInvalid}
	}
	if base+offset < 0 {
		return 0, &iofs.PathError{Op: "seek", Path: f.Name(), Err: iofs.ErrInvalid}
	}
	f.off = base + offset
	return f.off, nil
}

func (f *roFile) ReadDir(n int) ([]os.DirEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.node.info.IsDir() {
		return nil, &iofs.PathError{Op: "readdir", Path: f.Name(), Err: syscall.ENOTDIR}
	}
	names := f.node.names
	if f.dirOff >= len(names) {
		if n <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}
	end := len(names)
	if n > 0 && f.dirOff+n < end {
		end = f.dirOff + n
	}
	out := make([]os.DirEntry, 0, end-f.dirOff)
	for _, name := range names[f.dirOff:end] {
		out = append(out, roDirEntry{f.node.children[name].info})
	}
	f.dirOff = end
	return out, nil
}
