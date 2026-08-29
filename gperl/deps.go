package gperl

// CPAN dependency handling, following the cpanm/carton conventions the Perl
// world already uses: `cpanfile` declares, cpanm installs into ./local, and
// @INC gains local/lib/perl5.

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// localLib returns the vendored module tree for the project dir (carton's
// layout), or "" when none exists.
func localLib(projectDir string) string {
	dir := filepath.Join(projectDir, "local", "lib", "perl5")
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		return dir
	}
	return ""
}

// EnsureDeps makes sure the project's cpanfile dependencies are vendored:
// when a cpanfile exists but local/ does not, it runs the cpanm resolution
// step. A present local/ is trusted (re-run Get with --installdeps after
// editing the cpanfile).
func EnsureDeps(projectDir string) error {
	if _, err := os.Stat(filepath.Join(projectDir, "cpanfile")); err != nil {
		return nil // no declared dependencies
	}
	if localLib(projectDir) != "" {
		return nil
	}
	fmt.Fprintln(os.Stderr, "gperl: resolving cpanfile dependencies into local/")
	return runCpanm(projectDir, []string{"--installdeps", projectDir})
}

// Get vendors the named modules (cpanm arguments, e.g. module names or
// --installdeps .) into projectDir/local.
func Get(projectDir string, args []string) error {
	return runCpanm(projectDir, args)
}

// runCpanm invokes cpanm -L <project>/local with the standard sandboxing
// conventions. cpanm itself is bootstrapped from cpanmin.us when it is not
// installed. It always runs from a scratch directory: on a case-insensitive
// filesystem, resolving a module name like "Plack" from inside a directory
// named "plack" would otherwise install the directory.
func runCpanm(projectDir string, extra []string) error {
	local, err := filepath.Abs(filepath.Join(projectDir, "local"))
	if err != nil {
		return err
	}
	work, err := os.MkdirTemp("", "gperl-cpanm-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	args := append([]string{"-L", local, "--notest"}, extra...)
	var cmd *exec.Cmd
	if path, err := exec.LookPath("cpanm"); err == nil {
		cmd = exec.Command(path, args...)
	} else {
		script := filepath.Join(work, "cpanm")
		curl := exec.Command("curl", "-fsSL", "-o", script, "https://cpanmin.us")
		curl.Stderr = os.Stderr
		if err := curl.Run(); err != nil {
			return fmt.Errorf("bootstrap cpanm: %w", err)
		}
		cmd = exec.Command("perl", append([]string{script}, args...)...)
	}
	cmd.Dir = work
	cmd.Stdout = os.Stderr // progress is progress, not program output
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cpanm: %w", err)
	}
	return rejectNativeXS(local)
}

// rejectNativeXS refuses a vendored tree containing host-native XS objects.
// cpanm builds XS against the HOST perl; those shared objects can never load
// into the embedded interpreter, so failing at vendor time with the module
// names beats a cryptic runtime failure. The planned XS pipeline replaces
// this: each XS dist becomes a shared-memory wasm side module transpiled to
// a Go package (see perl-wasm's scripts/xs-wasm-build.sh) that go build
// links in.
func rejectNativeXS(local string) error {
	var offenders []string
	root := filepath.Join(local, "lib", "perl5")
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".so", ".bundle", ".dylib", ".dll":
			rel, _ := filepath.Rel(root, path)
			offenders = append(offenders, rel)
		}
		return nil
	})
	if len(offenders) == 0 {
		return nil
	}
	return fmt.Errorf("vendored tree contains host-native XS objects gperl cannot load yet:\n  %s\nXS support (XS -> wasm side module -> Go) is planned; until then depend on pure-Perl alternatives",
		strings.Join(offenders, "\n  "))
}
