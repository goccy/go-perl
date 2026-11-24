package mod

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	perl "github.com/goccy/go-perl"
)

// sharedHelper is one interpreter for the whole test binary.
var sharedHelper *Helper

func TestMain(m *testing.M) {
	sharedHelper = &Helper{}
	code := m.Run()
	if err := sharedHelper.Close(); err != nil {
		println("close helper:", err.Error())
		code = 1
	}
	os.Exit(code)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHelperCPANfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cpanfile")
	writeFile(t, path, `
requires 'Plack', '1.0050';
requires 'Mojolicious';
requires 'Old::Thing', '>= 1.0, < 2.0';
requires 'Pinned', version => '== 1.23', dist => 'X/XX/XXX/Pinned-1.23.tar.gz';
recommends 'Cpanel::JSON::XS', '4.0';
on 'test' => sub {
    requires 'Test::More', '0.98';
};
on develop => sub { requires 'Perl::Tidy' };
configure_requires 'Module::Build::Tiny';
feature 'sqlite', 'SQLite support' => sub {
    requires 'DBD::SQLite';
};
`)
	got, err := sharedHelper.CPANfile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []Prereq{
		{"Plack", "1.0050", "runtime", "requires"},
		{"Mojolicious", "0", "runtime", "requires"},
		{"Old::Thing", ">= 1.0, < 2.0", "runtime", "requires"},
		{"Pinned", "== 1.23", "runtime", "requires"},
		{"Cpanel::JSON::XS", "4.0", "runtime", "recommends"},
		{"Test::More", "0.98", "test", "requires"},
		{"Perl::Tidy", "0", "develop", "requires"},
		{"Module::Build::Tiny", "0", "configure", "requires"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CPANfile =\n%v\nwant\n%v", got, want)
	}

	writeFile(t, path, "requires 'Plack'\nthis is not perl (\n")
	if _, err := sharedHelper.CPANfile(path); err == nil {
		t.Fatal("a cpanfile with a syntax error evaluated")
	}
	if _, err := sharedHelper.CPANfile(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("a missing cpanfile evaluated")
	}
}

func TestHelperVersions(t *testing.T) {
	for _, c := range []struct {
		rng      string
		versions []string
		want     int
	}{
		{"0", []string{"1.0", "2.0", "1.5"}, 1},
		{"< 2.0", []string{"1.0", "2.0", "1.5"}, 2},
		{"== 1.0", []string{"1.0", "2.0", "1.5"}, 0},
		{">= 3", []string{"1.0", "2.0"}, -1},
		{">= 1.2.3", []string{"v1.2.2", "v1.10.0", "v1.9.0"}, 1},
		{"0", []string{"undef"}, 0},
		{"1.0", []string{"undef"}, -1},
	} {
		got, err := sharedHelper.Best(c.rng, c.versions...)
		if err != nil || got != c.want {
			t.Errorf("Best(%q, %v) = %d, %v; want %d", c.rng, c.versions, got, err, c.want)
		}
	}
	if ok, err := sharedHelper.Accepts(">= 1.0, != 1.5", "1.5"); err != nil || ok {
		t.Errorf("Accepts(!= 1.5, 1.5) = %v, %v", ok, err)
	}
	if ok, err := sharedHelper.Accepts(">= 1.0, != 1.5", "1.6"); err != nil || !ok {
		t.Errorf("Accepts(!= 1.5, 1.6) = %v, %v", ok, err)
	}
	if _, err := sharedHelper.Best("not a range ~~", "1.0"); err == nil {
		t.Error("a malformed range was accepted")
	}
}

func TestHelperBundled(t *testing.T) {
	if v, ok, err := sharedHelper.Bundled("perl"); err != nil || !ok || v < "5.04" {
		t.Errorf("Bundled(perl) = %q, %v, %v", v, ok, err)
	}
	if v, ok, err := sharedHelper.Bundled("File::Spec"); err != nil || !ok || v == "" {
		t.Errorf("Bundled(File::Spec) = %q, %v, %v", v, ok, err)
	}
	if v, ok, err := sharedHelper.Bundled("No::Such::Module::Anywhere"); err != nil || ok || v != "" {
		t.Errorf("Bundled(missing) = %q, %v, %v", v, ok, err)
	}
}

func TestHelperMeta(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "META.json"), `{
  "abstract": "x", "author": ["a"], "dynamic_config": 0, "generated_by": "t",
  "license": ["perl_5"], "meta-spec": {"version": 2}, "name": "Pure-Dist", "version": "1.02",
  "x_static_install": 1,
  "prereqs": {
    "configure": {"requires": {"Module::Build::Tiny": "0.035"}},
    "runtime":   {"requires": {"Try::Tiny": "0.30", "perl": "5.008001"}, "recommends": {"Fast::XS": "0"}},
    "test":      {"requires": {"Test::More": "0.98"}}
  },
  "provides": {"Pure::Dist": {"file": "lib/Pure/Dist.pm", "version": "1.02"}, "Pure::Dist::Util": {"file": "lib/Pure/Dist/Util.pm"}}
}`)
	m, err := sharedHelper.Meta(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.DynamicConfig || !m.StaticInstall || m.HasXS {
		t.Errorf("flags = %+v", m)
	}
	wantPrereqs := []Prereq{
		{"Module::Build::Tiny", "0.035", "configure", "requires"},
		{"Test::More", "0.98", "test", "requires"},
		{"Try::Tiny", "0.30", "runtime", "requires"},
		{"perl", "5.008001", "runtime", "requires"},
		{"Fast::XS", "0", "runtime", "recommends"},
	}
	if !reflect.DeepEqual(m.Prereqs, wantPrereqs) {
		t.Errorf("Prereqs =\n%v\nwant\n%v", m.Prereqs, wantPrereqs)
	}
	if want := []Package{{"Pure::Dist", "1.02"}, {"Pure::Dist::Util", "undef"}}; !reflect.DeepEqual(m.Provides, want) {
		t.Errorf("Provides = %v", m.Provides)
	}

	// Without META: provides come from scanning lib/, and the prereqs are
	// only known after configure.
	bare := t.TempDir()
	writeFile(t, filepath.Join(bare, "lib", "Bare", "XS.pm"), "package Bare::XS;\nour $VERSION = '0.03';\n1;\n")
	writeFile(t, filepath.Join(bare, "XS.xs"), "MODULE = Bare::XS PACKAGE = Bare::XS\n")
	writeFile(t, filepath.Join(bare, "t", "helper.c"), "int x;\n")
	m, err = sharedHelper.Meta(bare)
	if err != nil {
		t.Fatal(err)
	}
	if !m.DynamicConfig || m.StaticInstall || !m.HasXS || len(m.Prereqs) != 0 {
		t.Errorf("bare tree flags = %+v", m)
	}
	if want := []Package{{"Bare::XS", "0.03"}}; !reflect.DeepEqual(m.Provides, want) {
		t.Errorf("bare tree Provides = %v", m.Provides)
	}

	// A Build.PL distribution that requires nothing to configure requires
	// Module::Build, as under cpanm.
	moduleBuild := []Prereq{{"Module::Build", "0.38", "configure", "requires"}}
	for _, c := range []struct {
		name  string
		dir   string // the tree's directory name
		files map[string]string
		want  []Prereq
	}{
		{name: "Build.PL, no configure requires", dir: "MB-Dist@1.0", want: moduleBuild, files: map[string]string{
			"Build.PL":  "",
			"META.json": `{"abstract":"x","author":["a"],"dynamic_config":1,"generated_by":"t","license":["perl_5"],"meta-spec":{"version":2},"name":"MB-Dist","version":"1.0"}`,
		}},
		{name: "Build.PL, no META", dir: "MB-Dist@1.0", want: moduleBuild, files: map[string]string{"Build.PL": ""}},
		{name: "Build.PL, configure requires", dir: "MB-Dist@1.0", want: []Prereq{{"ExtUtils::MakeMaker", "0", "configure", "requires"}}, files: map[string]string{
			"Build.PL":  "",
			"META.json": `{"abstract":"x","author":["a"],"dynamic_config":1,"generated_by":"t","license":["perl_5"],"meta-spec":{"version":2},"name":"MB-Dist","version":"1.0","prereqs":{"configure":{"requires":{"ExtUtils::MakeMaker":"0"}}}}`,
		}},
		{name: "no Build.PL", dir: "MM-Dist@1.0", files: map[string]string{"Makefile.PL": ""}},
		{name: "a distribution Module::Build needs", dir: "ExtUtils-ParseXS@3.51", files: map[string]string{"Build.PL": ""}},
	} {
		tree := filepath.Join(t.TempDir(), c.dir)
		for name, content := range c.files {
			writeFile(t, filepath.Join(tree, name), content)
		}
		m, err := sharedHelper.Meta(tree)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !reflect.DeepEqual(m.Prereqs, c.want) {
			t.Errorf("%s: Prereqs = %v, want %v", c.name, m.Prereqs, c.want)
		}
	}

	// Scanning lib/: a version line that cannot be evaluated gives an
	// undefined version (this relies on Module::Metadata's private
	// _evaluate_version_line), and one that loads a module finds it in the
	// distribution's lib/.
	versions := t.TempDir()
	writeFile(t, filepath.Join(versions, "lib", "Good.pm"), "package Good;\nour $VERSION = '1.0';\n1;\n")
	writeFile(t, filepath.Join(versions, "lib", "Bad.pm"), "package Bad;\n$Bad::VERSION = '2.0'; use No::Such::Module;\n1;\n")
	writeFile(t, filepath.Join(versions, "lib", "Dep.pm"), "package Dep;\nour $VERSION = '3.0';\n1;\n")
	writeFile(t, filepath.Join(versions, "lib", "User.pm"), "package User;\nuse Dep; $User::VERSION = $Dep::VERSION;\n1;\n")
	cwd := helperCwd(t)
	m, err = sharedHelper.Meta(versions)
	if err != nil {
		t.Fatal(err)
	}
	if want := []Package{{"Bad", "undef"}, {"Dep", "3.0"}, {"Good", "1.0"}, {"User", "3.0"}}; !reflect.DeepEqual(m.Provides, want) {
		t.Errorf("Provides with version lines that fail = %v, want %v", m.Provides, want)
	}
	if got := helperCwd(t); got != cwd {
		t.Errorf("helper working directory after Meta = %q, want %q", got, cwd)
	}
}

// helperCwd returns the helper interpreter's working directory.
func helperCwd(t *testing.T) string {
	t.Helper()
	sharedHelper.start.Do(sharedHelper.boot)
	if sharedHelper.err != nil {
		t.Fatal(sharedHelper.err)
	}
	sharedHelper.mu.Lock()
	defer sharedHelper.mu.Unlock()
	r, err := sharedHelper.p.Eval(t.Context(), "require Cwd; Cwd::getcwd()")
	if err == nil && r.Error != nil {
		err = r.Error
	}
	if err != nil {
		t.Fatal(err)
	}
	s, err := perl.As[perl.ScalarValue](r.Value)
	if err != nil {
		t.Fatal(err)
	}
	return s.String()
}
