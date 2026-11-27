package perl

// The module tree the go-perl/cpan package carries (internal.SetCPANLib)
// joins @INC the way the standard library reaches an instance: written onto
// the host filesystem once per user for an instance on it, mounted
// read-only on the in-memory filesystem WithStdlib builds (stdlibfs.go). An
// instance on a filesystem of the embedder's has the tree where the
// embedder puts it.

import (
	"path"
	"path/filepath"
	"sync"

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
// own).
func cpanLibInc(fsys FS) (string, error) {
	lib, _ := internal.CPANLib()
	if lib == nil {
		return "", nil
	}
	if m, ok := fsys.(*mountFS); ok {
		// WithStdlib mounted the tree, unless it was built before the tree
		// was installed.
		if _, mounted := m.tree.root.children[path.Base(cpanLibDir)]; !mounted {
			return "", nil
		}
		return "/" + cpanLibDir, nil
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
	if !filepath.IsLocal(rel) {
		// The filesystem does not reach the cache: it is the embedder's.
		return "", nil
	}
	return "/" + filepath.ToSlash(rel), nil
}
