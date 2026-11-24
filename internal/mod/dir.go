// Package mod manages a project's CPAN dependencies the way the go command
// manages modules: the project keeps a cpanfile and a cpanfile.snapshot,
// and the distributions themselves live in a module directory shared by
// every project on the machine.
//
// Layout of the module directory:
//
//	cache/download/authors/id/A/AU/AUTHOR/Dist-1.23.tar.gz   the tarballs
//	Dist@1.23/                                               one extracted tree
package mod

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/goccy/go-perl/internal"
)

// Target names the platform this process runs on, as the directory its
// native artifacts live under ("darwin_arm64").
func Target() string { return runtime.GOOS + "_" + runtime.GOARCH }

// BuildDir is where a distribution tree keeps what gperl derives from it.
func BuildDir(tree string) string { return filepath.Join(tree, "gperl") }

// ABIDir is the tree's Go package of native modules for the current SDK.
func ABIDir(tree string) string {
	return filepath.Join(BuildDir(tree), fmt.Sprintf("abi%d", internal.XSABI))
}

// TargetDir holds the tree's native artifacts for target, laid out like a
// perl architecture directory (auto/<Module/Path>/<Name>.a and .so).
func TargetDir(tree, target string) string { return filepath.Join(ABIDir(tree), target) }

// builtMarker names the file that records a finished build; it holds
// GperlVersion of the gperl that built.
const builtMarker = "built"

// GperlVersion identifies this gperl in what it builds and generates: the
// version of its main module from the build information, and the VCS
// revision when the build recorded one, so that another gperl builds and
// generates them again. A development or test build without either is
// "gperl (devel)". It is never empty, as the mark of an earlier gperl is.
var GperlVersion = sync.OnceValue(func() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "gperl (no build information)"
	}
	version := "gperl " + info.Main.Version
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			version += " " + s.Value
		}
	}
	return version
})

// Built reports whether the tree is ready to use on target: it needs no
// build, or this gperl has built it.
func Built(tree, target string) bool {
	for _, dir := range []string{BuildDir(tree), TargetDir(tree, target)} {
		if data, err := os.ReadFile(filepath.Join(dir, builtMarker)); err == nil && string(data) == GperlVersion() {
			return true
		}
	}
	return false
}

// MarkBuilt records that the tree is ready: for target when it carries
// native modules, for every target when target is empty.
func MarkBuilt(tree, target string) error {
	dir := BuildDir(tree)
	if target != "" {
		dir = TargetDir(tree, target)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, builtMarker), []byte(GperlVersion()), 0o644)
}

// LibDir returns the tree's module directory for @INC: the built one when
// the distribution has a build step, its own lib/ otherwise.
func LibDir(tree string) string {
	built := filepath.Join(BuildDir(tree), "lib")
	if _, err := os.Stat(built); err == nil {
		return built
	}
	return filepath.Join(tree, "lib")
}

// DefaultMirror is the CPAN mirror distributions are fetched from unless
// GPERLMIRROR names another.
const DefaultMirror = "https://cpan.metacpan.org/"

// Dir returns the module directory: GPERLMODCACHE, or ~/gperl/pkg/mod.
func Dir() (string, error) {
	if dir := os.Getenv("GPERLMODCACHE"); dir != "" {
		return filepath.Abs(dir)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate the module directory (set GPERLMODCACHE): %w", err)
	}
	return filepath.Join(home, "gperl", "pkg", "mod"), nil
}

// Mirror returns the CPAN mirror in use, always with a trailing slash.
func Mirror() string {
	mirror := os.Getenv("GPERLMIRROR")
	if mirror == "" {
		mirror = DefaultMirror
	}
	return strings.TrimRight(mirror, "/") + "/"
}

// MirrorOnly reports whether resolution must use only the mirror's own
// index (GPERLMIRRORONLY).
func MirrorOnly() bool {
	switch os.Getenv("GPERLMIRRORONLY") {
	case "", "0", "false":
		return false
	}
	return true
}

// distFile matches a distribution archive's base name: name, version,
// archive suffix.
var distFile = regexp.MustCompile(`^(.+)-(v?[0-9][^-]*?)(\.tar\.gz|\.tgz|\.tar\.bz2|\.zip)$`)

// SplitPathname splits a distribution's path below authors/id
// ("R/RU/RURBAN/Cpanel-JSON-XS-4.38.tar.gz") into its name and version.
func SplitPathname(pathname string) (name, version string, err error) {
	m := distFile.FindStringSubmatch(pathname[strings.LastIndex(pathname, "/")+1:])
	if m == nil {
		return "", "", fmt.Errorf("%s: not a distribution archive name (want Name-Version.tar.gz)", pathname)
	}
	return m[1], m[2], nil
}

// TarballPath returns where the module directory keeps the archive of the
// distribution at pathname.
func TarballPath(dir, pathname string) string {
	return filepath.Join(dir, "cache", "download", "authors", "id", filepath.FromSlash(pathname))
}

// TreePath returns where the module directory keeps the extracted tree of
// the distribution at pathname.
func TreePath(dir, pathname string) (string, error) {
	name, version, err := SplitPathname(pathname)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+"@"+version), nil
}
