//go:build darwin || linux

package gperl

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/goccy/go-perl/internal/mod"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal"
	"github.com/goccy/go-perl/xs"
)

// The site directories of the perl command are a directory of the user's:
// %Config names them, @INC has them, and `make install` without arguments
// installs there.

// userWithEmptyData gives the test a user whose data directory is empty,
// and returns the directory gperl's site directories are under for that
// user. The standard library the user has unpacked stays in use.
func userWithEmptyData(t *testing.T) (siteRoot string) {
	t.Helper()
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := perl.ExtractStdlib(); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	if runtime.GOOS == "darwin" {
		// The user's caches and data are both found from the home
		// directory: the new home gets the caches of the old one.
		if err := os.MkdirAll(filepath.Join(home, "Library"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(cache, filepath.Join(home, "Library", "Caches")); err != nil {
			t.Fatal(err)
		}
		t.Setenv("HOME", home)
		return filepath.Join(home, "Library", "Application Support", "go-perl")
	}
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, "data"))
	return filepath.Join(home, "data", "go-perl")
}

// perlVersion is the version of the interpreter, as it says it.
func perlVersion(t *testing.T) string {
	t.Helper()
	status, stdout, stderr := runCLI(t, t.TempDir(), "", "-e", `printf "%vd", $^V`)
	if status != 0 || stdout == "" {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	return stdout
}

func TestSiteDirectoriesAreTheUsers(t *testing.T) {
	root := userWithEmptyData(t)
	lib := filepath.Join(root, "site_perl", perlVersion(t))
	arch := filepath.Join(lib, internal.ArchName())
	bin := filepath.Join(root, "bin")
	man1, man3 := filepath.Join(root, "man", "man1"), filepath.Join(root, "man", "man3")

	want := map[string]string{
		"siteprefix": root, "siteprefixexp": root,
		"sitelib_stem": filepath.Join(root, "site_perl"),
		"sitelib":      lib, "sitelibexp": lib, "installsitelib": lib,
		"sitearch": arch, "sitearchexp": arch, "installsitearch": arch,
		"sitebin": bin, "sitebinexp": bin, "installsitebin": bin,
		"sitescript": bin, "sitescriptexp": bin, "installsitescript": bin,
		"siteman1dir": man1, "siteman1direxp": man1, "installsiteman1dir": man1,
		"siteman3dir": man3, "siteman3direxp": man3, "installsiteman3dir": man3,
	}
	// Every key of %Config that names a site directory, as perl -V:name
	// prints them (which reads the whole configuration, not only the
	// values Config.pm holds itself).
	status, stdout, stderr := runCLI(t, t.TempDir(), "", "-MConfig", "-e",
		`for my $key (sort grep { /site/ && !/^(?:d_|use)/ && !/html/ } keys %Config) { print "$key=$Config{$key}\n" }`,
		"-e", `print "tied=", ref(tied %Config), "\n";`,
		"-e", `print "config_sh=", (Config::config_sh() =~ /^sitelib='([^']*)'$/m)[0], "\n";`)
	if status != 0 {
		t.Fatalf("status=%d stderr=%q", status, stderr)
	}
	got := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		key, value, _ := strings.Cut(line, "=")
		got[key] = value
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("$Config{%s} = %q, want %q", key, got[key], value)
		}
		delete(got, key)
	}
	if got["tied"] != "Config" {
		t.Errorf("%%Config is tied to %q, want Config", got["tied"])
	}
	if got["config_sh"] != lib {
		t.Errorf("config_sh names sitelib %q, want %q", got["config_sh"], lib)
	}
	delete(got, "tied")
	delete(got, "config_sh")
	for key, value := range got {
		t.Errorf("$Config{%s} = %q is a site directory this test does not know", key, value)
	}

	// Nothing was created by looking.
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Errorf("the site directory exists after a run that installed nothing (stat: %v)", err)
	}
}

// @INC has the site directories where perl has them: after the -I
// directories and PERL5LIB, before the standard library, the
// architecture-specific one first.
func TestSiteDirectoriesInInc(t *testing.T) {
	root := userWithEmptyData(t)
	lib := filepath.Join(root, "site_perl", perlVersion(t))
	arch := filepath.Join(lib, internal.ArchName())
	dir := t.TempDir()
	env := append(os.Environ(), "PERL5LIB="+filepath.Join(dir, "five"))
	const program = `my ($stdlib) = grep { -e "$_/strict.pm" } @INC; print join("\n", map { $_ eq $stdlib ? "STDLIB" : $_ } @INC), "\n";`
	run := func(args ...string) string {
		t.Helper()
		var out, errOut strings.Builder
		status, err := runPerlCLI(args, perlRunOpts{dir: dir, env: env, stdout: &out, stderr: &errOut})
		if err != nil || status != 0 {
			t.Fatalf("perl %q: status=%d err=%v stderr=%q", args, status, err, errOut.String())
		}
		return out.String()
	}
	inc := func(dirs ...string) string { return strings.Join(dirs, "\n") + "\n" }

	if got, want := run("-Iswitch", "-e", program), inc("switch", filepath.Join(dir, "five"), arch, lib, "STDLIB"); got != want {
		t.Errorf("@INC:\n%swant:\n%s", got, want)
	}
	// Under taint checks the directories of the environment go; the site
	// directories are the interpreter's own and stay.
	if got, want := run("-T", "-Iswitch", "-e", program), inc("switch", arch, lib, "STDLIB"); got != want {
		t.Errorf("@INC under -T:\n%swant:\n%s", got, want)
	}
	// A fork child has the same.
	child := `my $pid = fork // die "fork: $!"; if (!$pid) { ` + program + ` exit 0 } waitpid($pid, 0);`
	if got, want := run("-e", child), inc(filepath.Join(dir, "five"), arch, lib, "STDLIB"); got != want {
		t.Errorf("@INC of a fork child:\n%swant:\n%s", got, want)
	}
	// A module in a site directory shadows the standard library's, and
	// one in PERL5LIB shadows the site directory's.
	write := func(base, from string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(base, "Text"), 0o755); err != nil {
			t.Fatal(err)
		}
		src := "package Text::Tabs;\nour $FROM = '" + from + "';\n1;\n"
		if err := os.WriteFile(filepath.Join(base, "Text", "Tabs.pm"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	const from = `require Text::Tabs; print $Text::Tabs::FROM // "the standard library", "\n";`
	if got := run("-e", from); got != "the standard library\n" {
		t.Errorf("Text::Tabs from %q with an empty site directory", got)
	}
	write(lib, "sitelib")
	if got := run("-e", from); got != "sitelib\n" {
		t.Errorf("Text::Tabs from %q, want the site directory's", got)
	}
	write(arch, "sitearch")
	if got := run("-e", from); got != "sitearch\n" {
		t.Errorf("Text::Tabs from %q, want the architecture-specific site directory's", got)
	}
	write(filepath.Join(dir, "five"), "PERL5LIB")
	if got := run("-e", from); got != "PERL5LIB\n" {
		t.Errorf("Text::Tabs from %q, want the PERL5LIB directory's", got)
	}
}

// siteInstallEnv is the environment of a user who installs a distribution
// with the perl command: perl is the shim (also for the Makefile's own
// rules), with, for a distribution with native code, the description of
// this machine's C toolchain.
func siteInstallEnv(t *testing.T, native bool) (env []string, shim string) {
	t.Helper()
	work := t.TempDir()
	shim, err := xsWritePerlShim(work)
	if err != nil {
		t.Fatal(err)
	}
	env = append(os.Environ(),
		"GPERL_PERL_EXE="+shim,
		"PERL="+shim,
		"PATH="+filepath.Dir(shim)+string(os.PathListSeparator)+os.Getenv("PATH"),
		"PERL_MM_USE_DEFAULT=1",
	)
	if native {
		sdkDir := filepath.Join(work, "sdk")
		if err := xs.WriteSDK(sdkDir); err != nil {
			t.Fatal(err)
		}
		stdlib, err := perl.ExtractStdlib()
		if err != nil {
			t.Fatal(err)
		}
		fakeArch, err := xsStandInArchlib(work, sdkDir, stdlib)
		if err != nil {
			t.Fatal(err)
		}
		overlay, err := xsWriteConfigOverlay(work, targetToolchain(runtime.GOOS, "cc", "ar"), sdkDir, stdlib, fakeArch, shim)
		if err != nil {
			t.Fatal(err)
		}
		env = append(env, "GPERL_PERL_PRELOAD="+overlay)
		env = append(env, xsArchEnv(mod.Target())...)
	}
	return env, shim
}

// runStep runs one command of an installation in dir and fails the test
// with its output when it fails.
func runStep(t *testing.T, dir string, env []string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir, cmd.Env = dir, env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return string(out)
}

// perl Makefile.PL && make && make install, with no argument that says
// where: a pure-Perl module, its program and its manual page land in the
// site directories, which the installation creates, and the module then
// loads from any directory.
func TestMakeInstallGoesToTheSiteDirectories(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("no make on PATH")
	}
	root := userWithEmptyData(t)
	lib := filepath.Join(root, "site_perl", perlVersion(t))
	env, shim := siteInstallEnv(t, false)

	dist := t.TempDir()
	files := map[string]string{
		"Makefile.PL": "use ExtUtils::MakeMaker;\nWriteMakefile(NAME => 'Site::Probe', VERSION_FROM => 'lib/Site/Probe.pm', EXE_FILES => ['bin/site-probe']);\n",
		"lib/Site/Probe.pm": "package Site::Probe;\nour $VERSION = '0.01';\nsub where { __FILE__ }\n1;\n" +
			"__END__\n\n=head1 NAME\n\nSite::Probe - found in the site directory\n\n=cut\n",
		"bin/site-probe": "#!perl\nuse Site::Probe;\nprint \"probe \", Site::Probe->VERSION, \"\\n\";\n",
	}
	for name, src := range files {
		path := filepath.Join(dist, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runStep(t, dist, env, shim, "Makefile.PL")
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Errorf("the site directory exists before anything was installed (stat: %v)", err)
	}
	runStep(t, dist, env, "make")
	runStep(t, dist, env, "make", "install")

	module := filepath.Join(lib, "Site", "Probe.pm")
	for _, path := range []string{module, filepath.Join(root, "bin", "site-probe"), filepath.Join(root, "man", "man3", "Site::Probe.3")} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("not installed: %v", err)
		}
	}
	elsewhere := t.TempDir()
	if got := runStep(t, elsewhere, env, shim, "-MSite::Probe", "-e", "print Site::Probe::where()"); got != module {
		t.Errorf("Site::Probe loaded from %q, want %q", got, module)
	}
	if got := runStep(t, elsewhere, env, shim, filepath.Join(root, "bin", "site-probe")); got != "probe 0.01\n" {
		t.Errorf("the installed program printed %q", got)
	}
}

// The same for a distribution with native code: its shared object lands in
// the architecture-specific site directory and loads from there.
func TestMakeInstallOfANativeModuleGoesToTheSiteDirectories(t *testing.T) {
	hostXSToolchain(t)
	root := userWithEmptyData(t)
	lib := filepath.Join(root, "site_perl", perlVersion(t))
	arch := filepath.Join(lib, internal.ArchName())
	env, shim := siteInstallEnv(t, true)

	source, err := filepath.Abs(filepath.Join("testdata", "Demo-XS-dist"))
	if err != nil {
		t.Fatal(err)
	}
	dist := filepath.Join(t.TempDir(), "Demo-XS")
	if err := copyTree(source, dist); err != nil {
		t.Fatal(err)
	}
	runStep(t, dist, env, shim, "Makefile.PL")
	runStep(t, dist, env, "make")
	runStep(t, dist, env, "make", "install")

	for _, path := range []string{filepath.Join(arch, "Demo", "XS.pm"), filepath.Join(arch, "auto", "Demo", "XS", "XS.so")} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("not installed: %v", err)
		}
	}
	// The record of the installation is with the site directories too.
	if record, err := os.ReadFile(filepath.Join(arch, "perllocal.pod")); err != nil || !strings.Contains(string(record), "Demo::XS") {
		t.Errorf("perllocal.pod in %s: %v\n%s", arch, err, record)
	}
	got := runStep(t, t.TempDir(), env, shim, "-MDemo::XS", "-e", `print "sum=", Demo::XS::add(40, 2), " from ", $INC{"Demo/XS.pm"}`)
	if want := "sum=42 from " + filepath.Join(arch, "Demo", "XS.pm"); got != want {
		t.Errorf("the installed module printed %q, want %q", got, want)
	}
}
