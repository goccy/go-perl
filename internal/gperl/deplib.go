package gperl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-perl/internal/mod"
)

// projectDeps is a project's dependency set, ready to use on one target.
type projectDeps struct {
	// lib is the module trees of every distribution as one tree (slash
	// path -> host file, mod.LibFiles); empty when the project has no
	// dependencies.
	lib map[string]string
	// libDir is lib written out under the module directory (mergedLib).
	libDir string
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
	if deps.lib, err = mod.LibFiles(libDirs); err != nil {
		return nil, err
	}
	deps.libDir, err = mergedLib(m, deps.trees, deps.lib)
	return deps, err
}

// inc lists the directories a program on the dependencies has on @INC: the
// module trees as one, then the architecture directories, where the
// interpreter's loaders find the native modules.
func (d *projectDeps) inc() []string {
	if d.libDir == "" {
		return nil
	}
	return append([]string{d.libDir}, d.nativeDirs...)
}

// mergedLib writes lib under the module directory, keyed by the trees it
// is made of, and returns the directory; one already there is used as it
// is. The directory is only ever there complete: it is written beside its
// place and renamed into it.
func mergedLib(m *mod.Manager, trees []string, lib map[string]string) (string, error) {
	key := sha256.New()
	for _, tree := range trees {
		fmt.Fprintf(key, "%s\n", tree)
	}
	dir := filepath.Join(m.Dir, "cache", "lib", hex.EncodeToString(key.Sum(nil)[:8]))
	if dirExists(dir) {
		return dir, nil
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dir), "tmp-")
	if err != nil {
		return "", err
	}
	if err := writeLib(lib, tmp); err != nil {
		return "", errors.Join(err, os.RemoveAll(tmp))
	}
	if err := os.Rename(tmp, dir); err != nil {
		if !dirExists(dir) {
			return "", errors.Join(err, os.RemoveAll(tmp))
		}
		// Another process put the same tree in place first.
		return dir, os.RemoveAll(tmp)
	}
	return dir, nil
}

// writeLib writes the files of lib into dir, as hard links where possible.
func writeLib(lib map[string]string, dir string) error {
	for name, host := range lib {
		if err := linkOrCopy(host, filepath.Join(dir, filepath.FromSlash(name))); err != nil {
			return err
		}
	}
	return nil
}

// declaresDeps reports whether the project in dir declares dependencies: a
// cpanfile or cpanfile.snapshot.
func declaresDeps(dir string) bool {
	return fileExists(filepath.Join(dir, mod.CPANfile)) || fileExists(filepath.Join(dir, mod.SnapshotFile))
}

// hostDeps loads the dependencies of the project in dir for this platform.
// A project that declares none has none, and needs neither the module
// directory nor a C compiler.
func hostDeps(dir string) (*projectDeps, error) {
	if !declaresDeps(dir) {
		return &projectDeps{}, nil
	}
	var deps *projectDeps
	err := withDeps(func(m *mod.Manager) error {
		var err error
		deps, err = loadDeps(context.Background(), m, dir, nativeTarget(mod.Target()))
		return err
	})
	return deps, err
}
