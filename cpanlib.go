package perl

// The module tree the go-perl/cpan package carries (internal.SetCPANLib)
// joins @INC the way the standard library reaches an instance: written onto
// the host filesystem once per user for an instance on it, loaded into the
// in-memory filesystem an instance gets by default. An instance on a
// filesystem of the embedder's has the tree where the embedder puts it.

import (
	iofs "io/fs"
	"path"
	"path/filepath"
	"sync"

	goperlfs "github.com/goccy/go-perl/fs"
	"github.com/goccy/go-perl/internal"
	"github.com/goccy/perlwasm2go/base"
)

// cpanLibDir is where the tree is in the default in-memory filesystem, as
// an FS name; the guest sees it at "/" + cpanLibDir.
const cpanLibDir = ".goperl/cpan"

var (
	cpanLibOnce sync.Once
	cpanLibPath string
	cpanLibErr  error
)

// cpanLibInc returns the guest directory of the carried tree for an
// instance on fsys ("" when there is no tree, or fsys is the embedder's
// own). memfs is the in-memory filesystem the instance got by default, nil
// when the embedder gave one; the tree is written into it.
func cpanLibInc(fsys FS, memfs *MemFS) (string, error) {
	lib := internal.CPANLib()
	switch {
	case lib == nil:
		return "", nil
	case memfs != nil:
		return "/" + cpanLibDir, copyTree(memfs, cpanLibDir, lib)
	}
	host, ok := fsys.(base.HostRootFS)
	if !ok {
		return "", nil
	}
	cpanLibOnce.Do(func() { cpanLibPath, cpanLibErr = internal.UnpackTree(lib, "cpan") })
	if cpanLibErr != nil {
		return "", cpanLibErr
	}
	rel, err := filepath.Rel(host.HostRoot(), cpanLibPath)
	if err != nil {
		return "", err
	}
	return "/" + filepath.ToSlash(rel), nil
}

// copyTree writes tree into memfs under dir.
func copyTree(memfs *goperlfs.MemFS, dir string, tree iofs.FS) error {
	return iofs.WalkDir(tree, ".", func(name string, d iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst := path.Join(dir, name)
		if d.IsDir() {
			return memfs.MkdirAll(dst, 0o755)
		}
		data, err := iofs.ReadFile(tree, name)
		if err != nil {
			return err
		}
		return memfs.WriteFile(dst, data, 0o644)
	})
}
