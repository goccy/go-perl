package gperl

import (
	"context"
	"path/filepath"

	"github.com/goccy/go-perl/internal/mod"
)

// projectDeps is a project's dependency set, ready to use on one target.
type projectDeps struct {
	// lib is the module trees of every distribution as one tree, which a
	// built binary embeds; nil when the project has no dependencies.
	lib *mod.LibFS
	// libDirs are the module trees of the distributions, in snapshot order.
	libDirs []string
	// nativeDirs are the architecture directories (auto/<Module>/<Name>.so)
	// of the distributions with native modules.
	nativeDirs []string
	// trees are the distribution trees, in snapshot order.
	trees []string
}

// loadDeps syncs the project's dependencies (fetching and building what is
// missing; resolving only when the project has no snapshot) for target
// and indexes them for use.
func loadDeps(ctx context.Context, m *mod.Manager, projectDir, target string) (*projectDeps, error) {
	snap, err := syncDeps(ctx, m, projectDir, target, false)
	if err != nil {
		return nil, err
	}
	deps := &projectDeps{}
	if len(snap.Distributions) == 0 {
		return deps, nil
	}
	libDirs := make([]string, 0, len(snap.Distributions))
	for _, d := range snap.Distributions {
		tree, err := mod.TreePath(m.Dir, d.Pathname)
		if err != nil {
			return nil, err
		}
		deps.trees = append(deps.trees, tree)
		libDirs = append(libDirs, mod.LibDir(tree))
		if native := mod.TargetDir(tree, target); dirExists(filepath.Join(native, "auto")) {
			deps.nativeDirs = append(deps.nativeDirs, native)
		}
	}
	deps.libDirs = libDirs
	deps.lib, err = mod.NewLibFS(libDirs)
	return deps, err
}

// inc lists the directories a program on the dependencies has on @INC: the
// module trees, then the architecture directories, where the interpreter's
// loaders find the native modules.
func (d *projectDeps) inc() []string {
	return append(d.libDirs[:len(d.libDirs):len(d.libDirs)], d.nativeDirs...)
}
