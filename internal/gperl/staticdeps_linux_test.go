package gperl_test

import (
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/goccy/go-perl/internal/gperl"
)

// libraryChain writes, in a new directory, libgpa that uses libgpb that
// uses libgpc, each as a shared library (lib<name>.so.1, with the
// lib<name>.so a -l finds) recording the next as its dependency, and as an
// archive; libgpd as an archive only and libgpe as a shared library only.
// It returns the directory.
func libraryChain(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("no cc available")
	}
	dir := t.TempDir()
	run := func(name string, args ...string) {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
	}
	for _, lib := range []struct {
		name, src       string
		shared, archive bool
		needs           string
	}{
		{name: "gpc", src: "int gpc(void) { return 3; }\n", shared: true, archive: true},
		{name: "gpb", src: "int gpc(void);\nint gpb(void) { return gpc() + 2; }\n", shared: true, archive: true, needs: "-lgpc"},
		{name: "gpa", src: "int gpb(void);\nint gpa(void) { return gpb() + 1; }\n", shared: true, archive: true, needs: "-lgpb"},
		{name: "gpd", src: "int gpd(void) { return 4; }\n", archive: true},
		{name: "gpe", src: "int gpe(void) { return 5; }\n", shared: true},
	} {
		if err := os.WriteFile(filepath.Join(dir, lib.name+".c"), []byte(lib.src), 0o644); err != nil {
			t.Fatal(err)
		}
		if lib.shared {
			args := []string{"-shared", "-fPIC", "-o", "lib" + lib.name + ".so.1", "-Wl,-soname,lib" + lib.name + ".so.1", lib.name + ".c", "-L."}
			if lib.needs != "" {
				args = append(args, lib.needs)
			}
			run("cc", args...)
			if err := os.Symlink("lib"+lib.name+".so.1", filepath.Join(dir, "lib"+lib.name+".so")); err != nil {
				t.Fatal(err)
			}
		}
		if lib.archive {
			run("cc", "-c", "-fPIC", "-o", lib.name+".o", lib.name+".c")
			run("ar", "rcs", "lib"+lib.name+".a", lib.name+".o")
		}
	}
	return dir
}

// The libraries a module's shared libraries depend on are named through
// every level, each before the ones it depends on; a library without an
// archive, or one linked from its archive alone, adds nothing.
func TestStaticDepsFollowTheRecordedDependencies(t *testing.T) {
	dir := libraryChain(t)
	src := filepath.Join(dir, "module.c")
	if err := os.WriteFile(src, []byte("int gpa(void);\nint gpd(void);\nint gpe(void);\nint module(void) { return gpa() + gpd() + gpe(); }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	so := filepath.Join(dir, "module.so")
	libs := []string{"-L" + dir, "-lgpa", "-lgpd", "-lgpe"}
	if out, err := exec.Command("cc", append([]string{"-shared", "-fPIC", "-o", so, src}, libs...)...).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s", err, out)
	}
	got, err := gperl.StaticDeps([]string{"cc"}, "linux_amd64", so, libs)
	if err != nil {
		t.Fatal(err)
	}
	a, b, c := slices.Index(got, "-lgpa"), slices.Index(got, "-lgpb"), slices.Index(got, "-lgpc")
	if a < 0 || b < 0 || c < 0 || a > b || b > c {
		t.Errorf("StaticDeps = %v, want -lgpa, -lgpb, -lgpc in that order", got)
	}
	for _, absent := range []string{"-lgpd", "-lgpe"} {
		if slices.Contains(got, absent) {
			t.Errorf("StaticDeps = %v, want no %s", got, absent)
		}
	}
	// A darwin module is never linked statically.
	if got, err := gperl.StaticDeps([]string{"cc"}, "darwin_arm64", so, libs); err != nil || got != nil {
		t.Errorf("StaticDeps for darwin = %v, %v; want none", got, err)
	}

	// Files of a library's name that the linker would not take for the
	// module (a linker script, an object that is no shared library) are
	// passed over for the next directory's; a library with only such files
	// is not named, though its archive is there.
	decoys := t.TempDir()
	for name, from := range map[string]string{"libgpb.so.1": "", "libgpc.so.1": "gpc.o"} {
		content := []byte("INPUT(" + name + ")\n")
		if from != "" {
			var err error
			if content, err = os.ReadFile(filepath.Join(dir, from)); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(decoys, name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err = gperl.StaticDeps([]string{"cc"}, "linux_amd64", so, []string{"-L" + decoys, "-L" + dir})
	if err != nil {
		t.Fatal(err)
	}
	if a, b, c := slices.Index(got, "-lgpa"), slices.Index(got, "-lgpb"), slices.Index(got, "-lgpc"); a < 0 || b < 0 || c < 0 || a > b || b > c {
		t.Errorf("StaticDeps past files the linker skips = %v, want -lgpa, -lgpb, -lgpc in that order", got)
	}
	for _, name := range []string{"libgpa.so.1", "libgpa.a", "libgpb.a", "libgpc.a"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(decoys, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err = gperl.StaticDeps([]string{"cc"}, "linux_amd64", so, []string{"-L" + decoys})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got, "-lgpa") {
		t.Errorf("StaticDeps with only files the linker skips = %v, want -lgpa", got)
	}
	for _, absent := range []string{"-lgpb", "-lgpc"} {
		if slices.Contains(got, absent) {
			t.Errorf("StaticDeps with only files the linker skips = %v, want no %s", got, absent)
		}
	}
}

// A native module whose library depends on others links into a binary, a
// static one too.
func TestNativeModuleLinksLibraryDependencies(t *testing.T) {
	requireToolchain(t)
	dir := libraryChain(t)
	startMirror(t, mirrorDist{pathname: "G/GP/GP/Gp-XS-0.01.tar.gz", module: "Gp::XS", version: "0.01", files: map[string]string{
		"Makefile.PL": "use ExtUtils::MakeMaker;\nWriteMakefile(NAME => 'Gp::XS', VERSION => '0.01', LIBS => ['-L" + dir + " -lgpa']);\n",
		"XS.xs": "#define PERL_NO_GET_CONTEXT\n#include \"EXTERN.h\"\n#include \"perl.h\"\n#include \"XSUB.h\"\n\nint gpa(void);\n\n" +
			"MODULE = Gp::XS  PACKAGE = Gp::XS\n\nint\nvalue()\n  CODE:\n    RETVAL = gpa();\n  OUTPUT:\n    RETVAL\n",
		"lib/Gp/XS.pm": "package Gp::XS;\nour $VERSION = '0.01';\nuse XSLoader;\nXSLoader::load(__PACKAGE__, $VERSION);\n1;\n",
	}})
	project, err := tidyProject(t, "requires 'Gp::XS';\n")
	if err != nil {
		t.Fatalf("ModTidy: %v", err)
	}
	script := filepath.Join(project, "app.pl")
	if err := os.WriteFile(script, []byte("use Gp::XS;\nprint Gp::XS::value(), qq{\\n};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GPERL_DEV_REPLACE", devReplaces(t))
	t.Setenv("CGO_ENABLED", "1")
	build := func(t *testing.T, env ...string) {
		t.Helper()
		out := filepath.Join(t.TempDir(), "app")
		if err := gperl.Build(script, out); err != nil {
			t.Fatalf("Build: %v", err)
		}
		cmd := exec.Command(out)
		cmd.Dir = t.TempDir()
		cmd.Env = append(append(os.Environ(), "GPERLMODCACHE="+filepath.Join(t.TempDir(), "absent")), env...)
		if got, err := cmd.CombinedOutput(); err != nil || string(got) != "6\n" {
			t.Fatalf("binary printed %q (%v), want %q", got, err, "6\n")
		}
	}
	t.Run("dynamic", func(t *testing.T) {
		build(t, "LD_LIBRARY_PATH="+dir)
	})
	t.Run("static", func(t *testing.T) {
		libc, err := exec.Command("cc", "-print-file-name=libc.a").Output()
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsAbs(strings.TrimSpace(string(libc))) {
			t.Skip("the C compiler has no static C library")
		}
		t.Setenv("GOFLAGS", strings.TrimSpace(os.Getenv("GOFLAGS")+" -ldflags=-extldflags=-static"))
		build(t)
	})
}

// A module's references to what a native module it depends on exports go
// to that module's renamed definitions, unless a library the module links
// with defines them; two such modules exporting one referenced name are an
// error, a name nobody references is not.
func TestModuleImportsBindToProviders(t *testing.T) {
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("no cc available")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("cc", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cc %v: %v\n%s", args, err, out)
		}
	}
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("gpq.c", "int q_fn(void) { return 2; }\n")
	run("-shared", "-fPIC", "-o", "libgpq.so", "gpq.c")
	write("user.c", "int p_fn(void);\nint q_fn(void);\nint user(void) { return p_fn() + q_fn(); }\n")
	so := filepath.Join(dir, "user.so")
	libs := []string{"-L" + dir, "-lgpq"}
	run(append([]string{"-shared", "-fPIC", "-o", so, "user.c"}, libs...)...)
	provider := filepath.Join(dir, "prov")
	write("prov/auto/P/Q/Q.exports", "p_fn\nq_fn\n__goperl_xs_init\n")
	other := filepath.Join(dir, "other")
	write("other/auto/O/O.exports", "__goperl_xs_init\n")

	got, err := gperl.ModuleImports([]string{"cc"}, "linux_amd64", so, libs, []string{provider, other})
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]string{"p_fn": "p_fn..gperl.P__Q"}; !maps.Equal(got, want) {
		t.Errorf("ModuleImports = %v, want %v", got, want)
	}
	write("other/auto/O/O.exports", "p_fn\n")
	if _, err := gperl.ModuleImports([]string{"cc"}, "linux_amd64", so, libs, []string{provider, other}); err == nil {
		t.Error("ModuleImports with p_fn exported by two modules succeeded")
	}
}

// A native module that calls what the native module of a distribution it
// depends on exports (lent through dl_load_flags) runs in a script and links
// into a binary, a static one too.
func TestNativeModuleUsesAnothersSymbols(t *testing.T) {
	requireToolchain(t)
	const header = "#define PERL_NO_GET_CONTEXT\n#include \"EXTERN.h\"\n#include \"perl.h\"\n#include \"XSUB.h\"\n\n"
	prov := mirrorDist{pathname: "P/PR/PROV/Prov-XS-0.01.tar.gz", module: "Prov::XS", version: "0.01", files: map[string]string{
		"Makefile.PL": "use ExtUtils::MakeMaker;\nWriteMakefile(NAME => 'Prov::XS', VERSION => '0.01');\n",
		"XS.xs":       header + "int prov_value(void) { return 7; }\n\nMODULE = Prov::XS  PACKAGE = Prov::XS\n\nPROTOTYPES: DISABLE\n",
		"lib/Prov/XS.pm": "package Prov::XS;\nour $VERSION = '0.01';\nrequire DynaLoader;\nour @ISA = ('DynaLoader');\n" +
			"sub dl_load_flags { 0x01 }\n__PACKAGE__->bootstrap($VERSION);\n1;\n",
	}}
	user := mirrorDist{pathname: "U/US/USER/User-XS-0.01.tar.gz", module: "User::XS", version: "0.01", files: map[string]string{
		"META.json": `{"abstract":"x","author":["a"],"dynamic_config":1,"generated_by":"t","license":["perl_5"],"meta-spec":{"version":2},
			"name":"User-XS","version":"0.01","prereqs":{"runtime":{"requires":{"Prov::XS":"0"}}}}`,
		"Makefile.PL":    "use ExtUtils::MakeMaker;\nWriteMakefile(NAME => 'User::XS', VERSION => '0.01', PREREQ_PM => { 'Prov::XS' => 0 });\n",
		"XS.xs":          header + "int prov_value(void);\n\nMODULE = User::XS  PACKAGE = User::XS\n\nint\nvalue()\n  CODE:\n    RETVAL = prov_value();\n  OUTPUT:\n    RETVAL\n",
		"lib/User/XS.pm": "package User::XS;\nour $VERSION = '0.01';\nuse Prov::XS;\nuse XSLoader;\nXSLoader::load(__PACKAGE__, $VERSION);\n1;\n",
	}}
	startMirror(t, prov, user)
	project, err := tidyProject(t, "requires 'User::XS';\n")
	if err != nil {
		t.Fatalf("ModTidy: %v", err)
	}
	script := filepath.Join(project, "app.pl")
	if err := os.WriteFile(script, []byte("use User::XS;\nprint User::XS::value(), qq{\\n};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := runScript(t, script); got != "7\n" {
		t.Fatalf("gperl run printed %q, want %q", got, "7\n")
	}
	t.Setenv("GPERL_DEV_REPLACE", devReplaces(t))
	t.Setenv("CGO_ENABLED", "1")
	build := func(t *testing.T) {
		t.Helper()
		out := filepath.Join(t.TempDir(), "app")
		if err := gperl.Build(script, out); err != nil {
			t.Fatalf("Build: %v", err)
		}
		cmd := exec.Command(out)
		cmd.Dir = t.TempDir()
		cmd.Env = append(os.Environ(), "GPERLMODCACHE="+filepath.Join(t.TempDir(), "absent"))
		if got, err := cmd.CombinedOutput(); err != nil || string(got) != "7\n" {
			t.Fatalf("binary printed %q (%v), want %q", got, err, "7\n")
		}
	}
	t.Run("dynamic", build)
	t.Run("static", func(t *testing.T) {
		libc, err := exec.Command("cc", "-print-file-name=libc.a").Output()
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsAbs(strings.TrimSpace(string(libc))) {
			t.Skip("the C compiler has no static C library")
		}
		t.Setenv("GOFLAGS", strings.TrimSpace(os.Getenv("GOFLAGS")+" -ldflags=-extldflags=-static"))
		build(t)
	})
}
