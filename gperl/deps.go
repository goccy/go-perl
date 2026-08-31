package gperl

// CPAN dependency handling, following the cpanm/carton conventions the Perl
// world already uses: `cpanfile` declares, cpanm installs into ./local, and
// @INC gains local/lib/perl5.

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/xs"
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
	if err := runCpanmArgs(work, work, args); err != nil {
		return fmt.Errorf("cpanm: %w", err)
	}
	warnHostXS(local)
	return nil
}

// cpanmDriver is the program that runs cpanm on the embedded interpreter.
// cpanm is a modulino: run as a plain file it only defines
// App::cpanminus::script and enters its main under `unless (caller)` —
// and inside an embedded interpreter the script runner is always a
// caller. So the driver loads the file for its definitions and then
// enters cpanm the way cpanm's own main does.
const cpanmDriver = `
my $r = do $ENV{GPERL_CPANM};
die $@ if $@;
App::cpanminus::script->can("new")
    or die "$ENV{GPERL_CPANM}: does not define App::cpanminus::script\n";
my $app = App::cpanminus::script->new;
$app->parse_options(@ARGV);
exit $app->doit;
`

// runCpanmArgs runs cpanm (the installed script when present, else a copy
// bootstrapped from cpanmin.us into work) with the given cpanm arguments
// on an IN-PROCESS interpreter — cpanm is pure Perl, so no perl needs to
// be installed and no perl process is spawned for it. cpanm's own child
// builds still re-invoke $^X (Makefile.PL, Build.PL) and some dists
// shell out to a bare `perl`: both resolve into the shim, which
// re-enters this executable as `gperl run` with the %Config overlay
// preloaded. cpanm progress streams to stderr (progress is progress,
// not program output).
func runCpanmArgs(work, dir string, cpanmArgs []string) error {
	script := filepath.Join(work, "cpanm")
	if path, lerr := exec.LookPath("cpanm"); lerr == nil {
		script = path
	} else if err := fetchCpanm(script); err != nil {
		return fmt.Errorf("bootstrap cpanm: %w", err)
	}
	shim, err := xsWritePerlShim(work)
	if err != nil {
		return err
	}
	stdlibDir, err := perl.ExtractStdlib()
	if err != nil {
		return err
	}
	// XS distributions cpanm builds along the way (build-time deps like
	// Devel::PPPort) compile against the SDK, exactly like the dists
	// `gperl xs build` drives directly.
	sdkDir := filepath.Join(work, "sdk")
	if err := xs.WriteSDK(sdkDir); err != nil {
		return fmt.Errorf("materialize SDK headers: %w", err)
	}
	fakeArch := filepath.Join(work, "fakearchlib")
	if err := os.MkdirAll(fakeArch, 0o755); err != nil {
		return err
	}
	if err := os.Symlink(sdkDir, filepath.Join(fakeArch, "CORE")); err != nil {
		return err
	}
	overlay, err := xsWriteConfigOverlay(work, sdkDir, stdlibDir, fakeArch)
	if err != nil {
		return err
	}
	env := append(os.Environ(),
		"GPERL_CPANM="+script,
		"GPERL_PERL_EXE="+shim,
		"GPERL_PERL_PRELOAD="+overlay,
		"PATH="+filepath.Dir(shim)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	argv := append([]string{"-e", cpanmDriver, "--"}, cpanmArgs...)
	return runPerlInProcess(dir, env, os.Stderr, argv...)
}

// fetchCpanm downloads the fatpacked cpanm script to path. cpanm is the
// one thing the pipeline cannot run before it has it; everything after
// this flows through the embedded interpreter's own networking.
func fetchCpanm(path string) error {
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Get("https://cpanmin.us")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET https://cpanmin.us: %s", resp.Status)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// warnHostXS points out vendored host-perl XS objects — a local/ tree
// populated by a host perl toolchain (carton install, a system cpanm)
// contains shared objects built against that perl, and those never load
// into the embedded interpreter. The .pm halves are fine; the native
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
