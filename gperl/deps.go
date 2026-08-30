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
	"sort"
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
// step. A present local/ is trusted (delete local/ to force a re-run after
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
	argv, err := cpanmArgv(work)
	if err != nil {
		return err
	}
	cmd := exec.Command(argv[0], append(argv[1:], args...)...)
	cmd.Dir = work
	cmd.Stdout = os.Stderr // progress is progress, not program output
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cpanm: %w", err)
	}
	warnHostXS(local)
	return nil
}

// cpanmArgv resolves how to invoke cpanm: the installed command when
// present, else a copy bootstrapped from cpanmin.us into work.
func cpanmArgv(work string) ([]string, error) {
	if path, err := exec.LookPath("cpanm"); err == nil {
		return []string{path}, nil
	}
	script := filepath.Join(work, "cpanm")
	curl := exec.Command("curl", "-fsSL", "-o", script, "https://cpanmin.us")
	curl.Stderr = os.Stderr
	if err := curl.Run(); err != nil {
		return nil, fmt.Errorf("bootstrap cpanm: %w", err)
	}
	return []string{"perl", script}, nil
}

// warnHostXS points out vendored host-perl XS objects: cpanm builds XS
// against the HOST perl, and those shared objects never load into the
// embedded interpreter. The .pm halves it installed are fine; the native
// halves come from `gperl xs build <dist>`, which compiles the same
// distribution against the go-perl XS SDK.
func warnHostXS(local string) {
	seen := map[string]bool{}
	root := filepath.Join(local, "lib", "perl5")
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".so", ".bundle", ".dylib", ".dll":
			// .../auto/Devel/NYTProf/NYTProf.bundle -> Devel::NYTProf
			if i := strings.Index(path, string(filepath.Separator)+"auto"+string(filepath.Separator)); i >= 0 {
				mod := filepath.Dir(path[i+6:])
				seen[strings.ReplaceAll(mod, string(filepath.Separator), "::")] = true
			}
		}
		return nil
	})
	if len(seen) == 0 {
		return
	}
	var mods []string
	for m := range seen {
		mods = append(mods, m)
	}
	sort.Strings(mods)
	fmt.Fprintf(os.Stderr,
		"gperl: these vendored modules contain XS built for the HOST perl, which the embedded interpreter cannot use:\n  %s\nbuild their native halves with `gperl xs build <dist-source>` (they load automatically afterwards)\n",
		strings.Join(mods, "\n  "))
}
