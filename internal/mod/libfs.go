package mod

import (
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

// LibFS presents the module directories of several distributions as one
// tree, so a project's whole dependency set is a single @INC entry and a
// module lookup is one map access however many distributions there are.
// When two distributions carry the same file, the earlier directory wins.
type LibFS struct {
	files map[string]string   // slash path -> host file
	dirs  map[string][]string // slash path -> sorted child names
}

// NewLibFS indexes dirs; directories that do not exist contribute nothing.
func NewLibFS(dirs []string) (*LibFS, error) {
	l := &LibFS{files: map[string]string{}, dirs: map[string][]string{".": nil}}
	children := map[string]map[string]bool{".": {}}
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if p == dir && os.IsNotExist(err) {
					return fs.SkipAll
				}
				return err
			}
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			name := filepath.ToSlash(rel)
			if name == "." {
				return nil
			}
			if d.IsDir() {
				if children[name] == nil {
					children[name] = map[string]bool{}
				}
			} else if _, ok := l.files[name]; !ok {
				l.files[name] = p
			}
			children[path.Dir(name)][path.Base(name)] = true
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for dir, set := range children {
		names := make([]string, 0, len(set))
		for name := range set {
			names = append(names, name)
		}
		sort.Strings(names)
		l.dirs[dir] = names
	}
	return l, nil
}

// Files returns the indexed files: slash path -> host file.
func (l *LibFS) Files() map[string]string { return l.files }

func (l *LibFS) notExist(op, name string) error {
	return &fs.PathError{Op: op, Path: name, Err: fs.ErrNotExist}
}

// Stat implements fs.StatFS.
func (l *LibFS) Stat(name string) (fs.FileInfo, error) {
	if host, ok := l.files[name]; ok {
		return os.Stat(host)
	}
	if _, ok := l.dirs[name]; ok {
		return libDirInfo(path.Base(name)), nil
	}
	return nil, l.notExist("stat", name)
}

// ReadFile implements fs.ReadFileFS.
func (l *LibFS) ReadFile(name string) ([]byte, error) {
	host, ok := l.files[name]
	if !ok {
		return nil, l.notExist("open", name)
	}
	return os.ReadFile(host)
}

// ReadDir implements fs.ReadDirFS.
func (l *LibFS) ReadDir(name string) ([]fs.DirEntry, error) {
	names, ok := l.dirs[name]
	if !ok {
		return nil, l.notExist("readdir", name)
	}
	entries := make([]fs.DirEntry, len(names))
	for i, child := range names {
		full := path.Join(name, child)
		if _, isDir := l.dirs[full]; isDir {
			entries[i] = fs.FileInfoToDirEntry(libDirInfo(child))
			continue
		}
		info, err := os.Stat(l.files[full])
		if err != nil {
			return nil, err
		}
		entries[i] = fs.FileInfoToDirEntry(info)
	}
	return entries, nil
}

// Open implements fs.FS.
func (l *LibFS) Open(name string) (fs.File, error) {
	if host, ok := l.files[name]; ok {
		return os.Open(host)
	}
	entries, err := l.ReadDir(name)
	if err != nil {
		return nil, l.notExist("open", name)
	}
	return &libDir{info: libDirInfo(path.Base(name)), entries: entries}, nil
}

// libDir is an open directory of a LibFS.
type libDir struct {
	info    fs.FileInfo
	entries []fs.DirEntry
}

func (d *libDir) Stat() (fs.FileInfo, error) { return d.info, nil }
func (d *libDir) Close() error               { return nil }
func (d *libDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.info.Name(), Err: fs.ErrInvalid}
}

// ReadDir follows fs.ReadDirFile: n <= 0 returns the rest, otherwise at
// most n entries and io.EOF once none are left.
func (d *libDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if n > 0 && len(d.entries) == 0 {
		return nil, io.EOF
	}
	if n <= 0 || n > len(d.entries) {
		n = len(d.entries)
	}
	out := d.entries[:n]
	d.entries = d.entries[n:]
	return out, nil
}

// libDirInfo describes a directory of a LibFS, which exists only as the
// union of the indexed trees.
type libDirInfo string

func (i libDirInfo) Name() string       { return string(i) }
func (i libDirInfo) Size() int64        { return 0 }
func (i libDirInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o555 }
func (i libDirInfo) ModTime() time.Time { return time.Time{} }
func (i libDirInfo) IsDir() bool        { return true }
func (i libDirInfo) Sys() any           { return nil }
