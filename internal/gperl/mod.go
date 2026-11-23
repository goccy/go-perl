package gperl

// The dependency commands: a project declares its CPAN dependencies in a
// cpanfile, cpanfile.snapshot pins the resolved releases, and the
// distributions themselves live — fetched and built once — in the module
// directory shared by every project (GPERLMODCACHE, default
// ~/gperl/pkg/mod).

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-perl/internal/mod"
)

// withDeps runs fn with a dependency manager for the environment.
func withDeps(fn func(*mod.Manager) error) error {
	m, err := mod.NewManager()
	if err != nil {
		return err
	}
	return errors.Join(fn(m), m.Close())
}

// finishSync reports a finished sync and, when the project directory is a
// Go module, wires the dependencies into its builds.
func finishSync(m *mod.Manager, projectDir, target string) error {
	deps, err := loadDeps(context.Background(), m, projectDir, target)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "gperl: %d distributions ready in %s\n", len(deps.trees), m.Dir)
	return writeProjectGoWork(m, projectDir, target, deps)
}

// ModTidy resolves the project's cpanfile, writes cpanfile.snapshot, and
// fetches and builds every distribution the snapshot names. In a Go module
// it also writes go.work, so that the module's builds carry the
// dependencies.
func ModTidy(projectDir string) error {
	return withDeps(func(m *mod.Manager) error {
		target := buildTarget()
		if _, err := syncDeps(context.Background(), m, projectDir, target, true); err != nil {
			return err
		}
		return finishSync(m, projectDir, target)
	})
}

// ModDownload fetches and builds the distributions cpanfile.snapshot
// names; a project without a snapshot is resolved first. In a Go module it
// also writes go.work.
func ModDownload(projectDir string) error {
	return withDeps(func(m *mod.Manager) error { return finishSync(m, projectDir, buildTarget()) })
}

// Get adds modules to the project's cpanfile — "Module::Name" for the
// latest release, "Module::Name@1.23" for exactly that version — and
// resolves, fetches, and builds again. In a Go module it also writes
// go.work.
func Get(projectDir string, modules ...string) error {
	return withDeps(func(m *mod.Manager) error {
		if _, err := m.Get(context.Background(), projectDir, modules...); err != nil {
			return err
		}
		target := buildTarget()
		if _, err := syncDeps(context.Background(), m, projectDir, target, true); err != nil {
			return err
		}
		return finishSync(m, projectDir, target)
	})
}

// CleanModCache removes the module directory.
func CleanModCache() error {
	return withDeps(func(m *mod.Manager) error { return m.Clean() })
}

// buildTarget is the platform a build is for (mod.BuildTarget), as its
// native modules are built (nativeTarget).
func buildTarget() string { return nativeTarget(mod.BuildTarget()) }
