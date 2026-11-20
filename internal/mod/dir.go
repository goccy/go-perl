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
	"strings"
)

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
