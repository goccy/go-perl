//go:build darwin || linux

package gperl

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal"
)

// A run of the perl command that nothing else arranged (no build pipeline
// around it) knows a path it can be run by, and installs with it.

// plainEnv is the environment of the process without what the build
// pipeline puts there.
func plainEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		switch name {
		case "GPERL_PERL_EXE", "GPERL_PERL_PRELOAD", "PERL", "PERL5LIB":
			continue
		}
		env = append(env, kv)
	}
	return env
}

// ownPerl returns $^X of a plain run.
func ownPerl(t *testing.T) string {
	t.Helper()
	var out, errOut strings.Builder
	status, err := runPerlCLI([]string{"-e", `print $^X`}, perlRunOpts{
		env: plainEnv(), stdin: strings.NewReader(""), stdout: &out, stderr: &errOut,
	})
	if err != nil || status != 0 {
		t.Fatalf("status=%d err=%v stderr=%q", status, err, errOut.String())
	}
	return out.String()
}

// $^X is a path that runs this interpreter as perl, and %Config's perlpath
// and startperl name the same path, however %Config is asked.
func TestThePerlCommandNamesItself(t *testing.T) {
	self := ownPerl(t)
	if !filepath.IsAbs(self) {
		t.Fatalf("$^X = %q, want an absolute path", self)
	}
	cmd := exec.Command(self, "-MConfig", "-le",
		`Config::config_vars("perlpath"); print for $^X, $Config{perlpath}, $Config{startperl}, sprintf("%vd", $^V), "@ARGV"`,
		"a b", "c")
	cmd.Env = plainEnv()
	var errOut strings.Builder
	cmd.Stderr = &errOut
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s: %v\n%s", self, err, errOut.String())
	}
	version := perlVersion(t)
	// config_vars prints its line itself, with the line end of -l after it.
	want := "perlpath='" + self + "';\n\n" + self + "\n" + self + "\n#!" + self + "\n" + version + "\na b c\n"
	if string(out) != want {
		t.Errorf("got:\n%s\nwant:\n%s", out, want)
	}
}

// A program file run by that path is run as perl runs it: the switches of
// its #! line apply, and a file beside it that a project would resolve
// dependencies from is not looked at.
func TestTheNamedPerlRunsProgramFiles(t *testing.T) {
	self := ownPerl(t)
	dir := t.TempDir()
	script := filepath.Join(dir, "lines.pl")
	if err := os.WriteFile(script, []byte("#!"+self+" -l\nprint 'x'; print \"@ARGV\";\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cpanfile"), []byte("requires 'No::Such::Module::Anywhere';\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(script, "1", "2")
	cmd.Env, cmd.Dir = plainEnv(), t.TempDir()
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != "x\n1 2\n" {
		t.Errorf("%s: %v, output %q", script, err, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "local")); !os.IsNotExist(err) {
		t.Errorf("a project directory was made beside the program (stat: %v)", err)
	}
}

// perl Makefile.PL && make && make install by a user who set nothing up:
// the installed program names this perl on its #! line and runs by itself,
// and the record of the installation (perllocal.pod) is with the site
// directories, not in the directory the standard library was unpacked to.
func TestMakeInstallWithNothingArranged(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("no make on PATH")
	}
	root := userWithEmptyData(t)
	lib := filepath.Join(root, "site_perl", perlVersion(t))
	arch := filepath.Join(lib, internal.ArchName())
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatal(err)
	}
	self := ownPerl(t)
	env := append(plainEnv(), "PERL_MM_USE_DEFAULT=1")

	dist := t.TempDir()
	files := map[string]string{
		"Makefile.PL":                  "use ExtUtils::MakeMaker;\nWriteMakefile(NAME => 'Site::Unarranged::Probe', VERSION_FROM => 'lib/Site/Unarranged/Probe.pm', EXE_FILES => ['bin/unarranged-probe']);\n",
		"lib/Site/Unarranged/Probe.pm": "package Site::Unarranged::Probe;\nour $VERSION = '0.02';\n1;\n",
		"bin/unarranged-probe":         "#!/usr/bin/perl -l\nuse Site::Unarranged::Probe;\nprint \"probe \", Site::Unarranged::Probe->VERSION, \" by $^X\";\n",
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
	runStep(t, dist, env, self, "Makefile.PL")
	runStep(t, dist, env, "make")
	runStep(t, dist, env, "make", "install")

	program := filepath.Join(root, "bin", "unarranged-probe")
	src, err := os.ReadFile(program)
	if err != nil {
		t.Fatal(err)
	}
	first, _, _ := strings.Cut(string(src), "\n")
	if want := "#!" + self + " -l"; first != want {
		t.Errorf("the installed program starts with %q, want %q", first, want)
	}
	if got := runStep(t, t.TempDir(), env, program); got != "probe 0.02 by "+self+"\n" {
		t.Errorf("the installed program printed %q", got)
	}

	record, err := os.ReadFile(filepath.Join(arch, "perllocal.pod"))
	if err != nil {
		t.Errorf("no record of the installation with the site directories: %v", err)
	} else if !strings.Contains(string(record), "Site::Unarranged::Probe") {
		t.Errorf("perllocal.pod does not name the module:\n%s", record)
	}
	if other, err := os.ReadFile(filepath.Join(stdlib, "perllocal.pod")); err == nil && strings.Contains(string(other), "Site::Unarranged::Probe") {
		t.Errorf("the installation was recorded in the standard library's directory %s", stdlib)
	}
}
