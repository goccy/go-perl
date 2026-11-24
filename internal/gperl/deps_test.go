package gperl_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/goccy/go-perl/internal/gperl"
)

// mirrorDist is one distribution on the test mirror.
type mirrorDist struct {
	pathname string // below authors/id
	module   string // the module the index lists it under
	version  string
	files    map[string]string // archive members, below the top directory
}

// startMirror serves a CPAN mirror holding dists — tarballs, CHECKSUMS,
// and the 02packages index — and points the environment at it, with an
// empty module directory. Resolution uses the mirror's index only, so the
// tests reach no other host.
func startMirror(t *testing.T, dists ...mirrorDist) {
	t.Helper()
	tarballs := map[string][]byte{}
	for _, d := range dists {
		top := strings.TrimSuffix(path.Base(d.pathname), ".tar.gz")
		names := make([]string, 0, len(d.files))
		for name := range d.files {
			names = append(names, name)
		}
		sort.Strings(names)
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		tw := tar.NewWriter(zw)
		for _, name := range names {
			body := d.files[name]
			if err := tw.WriteHeader(&tar.Header{Name: top + "/" + name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write([]byte(body)); err != nil {
				t.Fatal(err)
			}
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		tarballs[d.pathname] = buf.Bytes()
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		switch p := r.URL.Path; {
		case p == "/modules/02packages.details.txt.gz":
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			fmt.Fprint(zw, "File: 02packages.details.txt\n\n")
			for _, d := range dists {
				fmt.Fprintf(zw, "%s %s %s\n", d.module, d.version, d.pathname)
			}
			if err := zw.Close(); err != nil {
				t.Error(err)
			}
			body = buf.Bytes()
		case strings.HasSuffix(p, "/CHECKSUMS"):
			author := strings.TrimSuffix(strings.TrimPrefix(p, "/authors/id/"), "CHECKSUMS")
			var b strings.Builder
			b.WriteString("$cksum = {\n")
			for _, d := range dists {
				if strings.HasPrefix(d.pathname, author) {
					sum := sha256.Sum256(tarballs[d.pathname])
					fmt.Fprintf(&b, "  '%s' => {\n    'sha256' => '%s'\n  },\n", path.Base(d.pathname), hex.EncodeToString(sum[:]))
				}
			}
			b.WriteString("};\n")
			body = []byte(b.String())
		default:
			body = tarballs[strings.TrimPrefix(p, "/authors/id/")]
		}
		if body == nil {
			http.NotFound(w, r)
			return
		}
		if _, err := w.Write(body); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GPERLMIRROR", srv.URL)
	t.Setenv("GPERLMIRRORONLY", "1")
	t.Setenv("GPERLMODCACHE", filepath.Join(t.TempDir(), "mod"))
}

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "Demo-XS-dist", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// testDists is a small dependency graph covering the three kinds of
// distribution: one used as extracted (static install), one with a build
// step, and one with a native (XS) module.
func testDists(t *testing.T) []mirrorDist {
	return []mirrorDist{
		{pathname: "S/ST/STATIC/Pure-Static-1.00.tar.gz", module: "Pure::Static", version: "1.00", files: map[string]string{
			"META.json": `{"abstract":"x","author":["a"],"dynamic_config":0,"generated_by":"t","license":["perl_5"],
				"meta-spec":{"version":2},"name":"Pure-Static","version":"1.00","x_static_install":1}`,
			"lib/Pure/Static.pm": "package Pure::Static;\nour $VERSION = '1.00';\nsub name { 'static' }\n1;\n",
		}},
		{pathname: "B/BU/BUILT/Pure-Built-0.50.tar.gz", module: "Pure::Built", version: "0.50", files: map[string]string{
			"Makefile.PL": "use ExtUtils::MakeMaker;\nWriteMakefile(NAME => 'Pure::Built', VERSION_FROM => 'lib/Pure/Built.pm', PREREQ_PM => { 'Pure::Static' => '1.00' });\n",
			"META.json": `{"abstract":"x","author":["a"],"dynamic_config":1,"generated_by":"t","license":["perl_5"],
				"meta-spec":{"version":2},"name":"Pure-Built","version":"0.50",
				"prereqs":{"runtime":{"requires":{"Pure::Static":"1.00"}}}}`,
			"lib/Pure/Built.pm": "package Pure::Built;\nour $VERSION = '0.50';\nuse Pure::Static;\nsub name { 'built on ' . Pure::Static::name() }\n1;\n",
		}},
		{pathname: "D/DE/DEMO/Demo-XS-0.01.tar.gz", module: "Demo::XS", version: "0.01", files: map[string]string{
			"Makefile.PL":    readTestdata(t, "Makefile.PL"),
			"XS.xs":          readTestdata(t, "XS.xs"),
			"lib/Demo/XS.pm": readTestdata(t, "lib/Demo/XS.pm"),
		}},
		// Two modules that define the same global names: each must keep
		// its own when both are in one binary.
		twinDist("A", 1),
		twinDist("B", 2),
	}
}

// twinDist is an XS distribution in two C files whose shared helper and
// state are ordinary global symbols — the same names in every twin.
func twinDist(letter string, value int) mirrorDist {
	return mirrorDist{
		pathname: "T/TW/TWIN/Twin-" + letter + "-0.01.tar.gz", module: "Twin::" + letter, version: "0.01",
		files: map[string]string{
			"Makefile.PL": "use ExtUtils::MakeMaker;\nWriteMakefile(NAME => 'Twin::" + letter + "', VERSION => '0.01', OBJECT => '" + letter + ".o helper.o');\n",
			"helper.c":    fmt.Sprintf("int twin_state = %d;\nint twin_helper(void) { return twin_state; }\n", value),
			letter + ".xs": "#define PERL_NO_GET_CONTEXT\n#include \"EXTERN.h\"\n#include \"perl.h\"\n#include \"XSUB.h\"\n\nint twin_helper(void);\n\n" +
				"MODULE = Twin::" + letter + "  PACKAGE = Twin::" + letter + "\n\nint\nvalue()\n  CODE:\n    RETVAL = twin_helper();\n  OUTPUT:\n    RETVAL\n",
			"lib/Twin/" + letter + ".pm": "package Twin::" + letter + ";\nour $VERSION = '0.01';\nuse XSLoader;\nXSLoader::load(__PACKAGE__, $VERSION);\n1;\n",
		},
	}
}

const testCPANfile = "requires 'Pure::Built';\nrequires 'Demo::XS';\nrequires 'Twin::A';\nrequires 'Twin::B';\n"

const testScript = `
use strict;
use Pure::Built;
use Demo::XS;
use Twin::A;
use Twin::B;
print Pure::Built::name(), " / ", Demo::XS::add(40, 2), " / ", Demo::XS::greet("gopher"), " / ", Twin::A::value(), Twin::B::value(), "\n";
`

const testOutput = "built on static / 42 / hello, gopher / 12\n"

func requireToolchain(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("builds distributions and binaries; skipped in -short")
	}
	for _, tool := range []string{"cc", "make", "go"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("no %s available", tool)
		}
	}
}

// runScript runs the test binary as the gperl command would run a script:
// in a child process, so its output can be captured.
func runScript(t *testing.T, script string) string {
	t.Helper()
	out, err := perlProcess(script).CombinedOutput()
	if err != nil {
		t.Fatalf("run %s: %v\n%s", script, err, out)
	}
	return string(out)
}

// The dependency commands resolve a cpanfile against the mirror, write the
// snapshot, and build every kind of distribution; a script then runs on
// them, and a built binary carries them.
func TestDependenciesEndToEnd(t *testing.T) {
	requireToolchain(t)
	startMirror(t, testDists(t)...)
	dir := t.TempDir()
	for name, content := range map[string]string{"cpanfile": testCPANfile, "app.pl": testScript} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := gperl.ModTidy(dir); err != nil {
		t.Fatalf("ModTidy: %v", err)
	}
	snapshot, err := os.ReadFile(filepath.Join(dir, "cpanfile.snapshot"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"  Demo-XS-0.01\n", "  Pure-Built-0.50\n", "  Pure-Static-1.00\n", "  Twin-A-0.01\n", "  Twin-B-0.01\n", "      Pure::Static 1.00\n"} {
		if !strings.Contains(string(snapshot), want) {
			t.Errorf("snapshot lacks %q:\n%s", want, snapshot)
		}
	}
	modDir := os.Getenv("GPERLMODCACHE")
	for _, artifact := range []string{
		"Pure-Static@1.00/gperl/built",
		"Pure-Built@0.50/gperl/lib/Pure/Built.pm",
		"Demo-XS@0.01/gperl/lib/Demo/XS.pm",
		"Demo-XS@0.01/gperl/abi42/xs.go",
		"Demo-XS@0.01/gperl/abi42/go.mod",
	} {
		if _, err := os.Stat(filepath.Join(modDir, filepath.FromSlash(artifact))); err != nil {
			t.Errorf("module directory: %v", err)
		}
	}
	matches, err := filepath.Glob(filepath.Join(modDir, "Demo-XS@0.01", "gperl", "abi42", "*", "auto", "Demo", "XS", "XS.*"))
	if err != nil || len(matches) != 3 {
		t.Fatalf("native artifacts = %v (%v), want XS.a, XS.ldflags, XS.so", matches, err)
	}

	// A second sync has nothing to do, and the mirror is not needed for it.
	t.Setenv("GPERLMIRROR", "http://127.0.0.1:1")
	if err := gperl.ModDownload(dir); err != nil {
		t.Fatalf("ModDownload with everything present: %v", err)
	}

	if got := runScript(t, filepath.Join(dir, "app.pl")); got != testOutput {
		t.Fatalf("gperl run printed %q, want %q", got, testOutput)
	}
	// The run had the module trees as one directory.
	merged, err := filepath.Glob(filepath.Join(modDir, "cache", "lib", "*", "Demo", "XS.pm"))
	if err != nil || len(merged) != 1 {
		t.Fatalf("merged module trees = %v (%v), want one with Demo/XS.pm", merged, err)
	}
	for _, pm := range []string{"Pure/Built.pm", "Twin/A.pm", "Twin/B.pm"} {
		if _, err := os.Stat(filepath.Join(filepath.Dir(filepath.Dir(merged[0])), filepath.FromSlash(pm))); err != nil {
			t.Errorf("merged module trees: %v", err)
		}
	}

	t.Setenv("GPERL_DEV_REPLACE", devReplaces(t))
	for name, cgo := range map[string]string{"static": "1", "nocgo": "0"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("CGO_ENABLED", cgo)
			out := filepath.Join(t.TempDir(), "app")
			if err := gperl.Build(filepath.Join(dir, "app.pl"), out); err != nil {
				t.Fatalf("Build: %v", err)
			}
			// The binary runs with neither the project nor the module
			// directory.
			cmd := exec.Command(out)
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), "GPERLMODCACHE="+filepath.Join(t.TempDir(), "absent"))
			got, err := cmd.CombinedOutput()
			if err != nil || string(got) != testOutput {
				t.Fatalf("binary printed %q (%v), want %q", got, err, testOutput)
			}
		})
	}
}

func TestGetAndClean(t *testing.T) {
	requireToolchain(t)
	startMirror(t, testDists(t)...)
	dir := t.TempDir()
	if err := gperl.Get(dir, "Pure::Built"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	cpanfile, err := os.ReadFile(filepath.Join(dir, "cpanfile"))
	if err != nil || string(cpanfile) != "requires 'Pure::Built';\n" {
		t.Fatalf("cpanfile = %q, %v", cpanfile, err)
	}
	if _, err := os.Stat(filepath.Join(os.Getenv("GPERLMODCACHE"), "Pure-Built@0.50", "gperl", "built")); err != nil {
		t.Fatal(err)
	}
	if err := gperl.Get(dir, "No::Such::Module"); err == nil {
		t.Fatal("Get of a module the mirror lacks succeeded")
	}
	if err := gperl.CleanModCache(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(os.Getenv("GPERLMODCACHE")); !os.IsNotExist(err) {
		t.Fatalf("module directory after clean: %v", err)
	}
	// A project without a cpanfile has nothing to sync.
	if err := gperl.ModDownload(t.TempDir()); err != nil {
		t.Fatalf("ModDownload without a cpanfile: %v", err)
	}
}

// A distribution whose native code does not compile fails the sync, and
// the error says where the build log is.
func TestBrokenXSFailsTheSync(t *testing.T) {
	requireToolchain(t)
	broken := twinDist("A", 1)
	broken.files["helper.c"] = "#error this distribution does not compile\n"
	startMirror(t, broken)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cpanfile"), []byte("requires 'Twin::A';\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := gperl.ModTidy(dir)
	if err == nil {
		t.Fatal("ModTidy succeeded with a distribution that does not compile")
	}
	for _, want := range []string{"build Twin-A-0.01", "this distribution does not compile", "build-error.log"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q:\n%v", want, err)
		}
	}
	logs, globErr := filepath.Glob(filepath.Join(os.Getenv("GPERLMODCACHE"), "Twin-A@0.01", "gperl", "abi42", "*", "build-error.log"))
	if globErr != nil || len(logs) != 1 {
		t.Fatalf("build log: %v, %v", logs, globErr)
	}
	// Nothing claims the distribution is built: the next sync tries again.
	if err := gperl.ModDownload(dir); err == nil {
		t.Fatal("a second sync succeeded")
	}
}

// tidyProject writes a project whose cpanfile is cpanfile and runs
// ModTidy on it.
func tidyProject(t *testing.T, cpanfile string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cpanfile"), []byte(cpanfile), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, gperl.ModTidy(dir)
}

// A distribution's build finds the modules it carries in its own directory
// (inc/), unless PERL_USE_UNSAFE_INC says otherwise.
func TestDistBuildFindsIncModules(t *testing.T) {
	requireToolchain(t)
	dist := mirrorDist{pathname: "I/IN/INC/Inc-Probe-0.01.tar.gz", module: "Inc::Probe", version: "0.01", files: map[string]string{
		"Makefile.PL":        "use inc::Gperl::Probe;\nuse ExtUtils::MakeMaker;\nWriteMakefile(NAME => 'Inc::Probe', VERSION => '0.01');\n",
		"inc/Gperl/Probe.pm": "package inc::Gperl::Probe;\n1;\n",
		"lib/Inc/Probe.pm":   "package Inc::Probe;\nour $VERSION = '0.01';\n1;\n",
	}}
	t.Run("default", func(t *testing.T) {
		t.Setenv("PERL_USE_UNSAFE_INC", "")
		if err := os.Unsetenv("PERL_USE_UNSAFE_INC"); err != nil {
			t.Fatal(err)
		}
		startMirror(t, dist)
		if _, err := tidyProject(t, "requires 'Inc::Probe';\n"); err != nil {
			t.Fatalf("ModTidy: %v", err)
		}
		if _, err := os.Stat(filepath.Join(os.Getenv("GPERLMODCACHE"), "Inc-Probe@0.01", "gperl", "lib", "Inc", "Probe.pm")); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("user setting", func(t *testing.T) {
		t.Setenv("PERL_USE_UNSAFE_INC", "0")
		startMirror(t, dist)
		_, err := tidyProject(t, "requires 'Inc::Probe';\n")
		if err == nil || !strings.Contains(err.Error(), "Can't locate inc/Gperl/Probe.pm") {
			t.Fatalf("ModTidy with PERL_USE_UNSAFE_INC=0: %v, want the build not to find inc/Gperl/Probe.pm", err)
		}
	})
}

// A distribution whose Build.PL does not configure it is configured by its
// Makefile.PL, as under cpanm; when both fail, the error carries both.
func TestBuildPLFallsBackToMakefilePL(t *testing.T) {
	requireToolchain(t)
	const meta = `{"abstract":"x","author":["a"],"dynamic_config":1,"generated_by":"t","license":["perl_5"],"meta-spec":{"version":2},
		"name":"Two-Ways","version":"0.01","prereqs":{"configure":{"requires":{"ExtUtils::MakeMaker":"0"}}}}`
	dist := mirrorDist{pathname: "T/TW/TWO/Two-Ways-0.01.tar.gz", module: "Two::Ways", version: "0.01", files: map[string]string{
		"META.json":       meta,
		"Build.PL":        "die qq{Build.PL cannot configure\\n};\n",
		"Makefile.PL":     "use ExtUtils::MakeMaker;\nWriteMakefile(NAME => 'Two::Ways', VERSION => '0.01');\n",
		"lib/Two/Ways.pm": "package Two::Ways;\nour $VERSION = '0.01';\n1;\n",
	}}
	t.Run("Makefile.PL", func(t *testing.T) {
		startMirror(t, dist)
		if _, err := tidyProject(t, "requires 'Two::Ways';\n"); err != nil {
			t.Fatalf("ModTidy: %v", err)
		}
		if _, err := os.Stat(filepath.Join(os.Getenv("GPERLMODCACHE"), "Two-Ways@0.01", "gperl", "lib", "Two", "Ways.pm")); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("both fail", func(t *testing.T) {
		dist.files["Makefile.PL"] = "die qq{Makefile.PL cannot configure\\n};\n"
		startMirror(t, dist)
		_, err := tidyProject(t, "requires 'Two::Ways';\n")
		if err == nil {
			t.Fatal("ModTidy succeeded with neither Build.PL nor Makefile.PL configuring")
		}
		for _, want := range []string{"perl Build.PL", "perl Makefile.PL"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error lacks %q:\n%v", want, err)
			}
		}
	})
}

// A Build.PL distribution that requires nothing to configure gets
// Module::Build, as under cpanm.
func TestBuildPLGetsModuleBuild(t *testing.T) {
	requireToolchain(t)
	// A stand-in Module::Build whose Build copies lib/ to blib/lib.
	moduleBuild := mirrorDist{pathname: "M/MB/MB/Module-Build-0.4234.tar.gz", module: "Module::Build", version: "0.4234", files: map[string]string{
		"META.json": `{"abstract":"x","author":["a"],"dynamic_config":0,"generated_by":"t","license":["perl_5"],
			"meta-spec":{"version":2},"name":"Module-Build","version":"0.4234","x_static_install":1}`,
		"lib/Module/Build.pm": `package Module::Build;
our $VERSION = '0.4234';
sub new { my ($class, %args) = @_; bless {%args}, $class }
sub create_build_script {
    open my $fh, '>', 'Build' or die "Build: $!";
    print $fh <<'BUILD';
use strict;
use File::Basename qw(dirname);
use File::Copy qw(copy);
use File::Find qw(find);
use File::Path qw(make_path);
find({ no_chdir => 1, wanted => sub {
    return unless -f $_;
    (my $to = $_) =~ s{^lib/}{blib/lib/};
    make_path(dirname($to));
    copy($_, $to) or die "copy $_: $!";
} }, 'lib');
BUILD
    close $fh or die "Build: $!";
}
1;
`,
	}}
	user := mirrorDist{pathname: "U/US/USER/MB-User-0.01.tar.gz", module: "MB::User", version: "0.01", files: map[string]string{
		"META.json": `{"abstract":"x","author":["a"],"dynamic_config":1,"generated_by":"t","license":["perl_5"],
			"meta-spec":{"version":2},"name":"MB-User","version":"0.01"}`,
		"Build.PL":       "use Module::Build;\nModule::Build->new(module_name => 'MB::User')->create_build_script;\n",
		"lib/MB/User.pm": "package MB::User;\nour $VERSION = '0.01';\n1;\n",
	}}
	startMirror(t, moduleBuild, user)
	dir, err := tidyProject(t, "requires 'MB::User';\n")
	if err != nil {
		t.Fatalf("ModTidy: %v", err)
	}
	snapshot, err := os.ReadFile(filepath.Join(dir, "cpanfile.snapshot"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"  Module-Build-0.4234\n", "      Module::Build 0.38\n"} {
		if !strings.Contains(string(snapshot), want) {
			t.Errorf("snapshot lacks %q:\n%s", want, snapshot)
		}
	}
	if _, err := os.Stat(filepath.Join(os.Getenv("GPERLMODCACHE"), "MB-User@0.01", "gperl", "lib", "MB", "User.pm")); err != nil {
		t.Fatal(err)
	}
}

// A PSGI application builds into a server binary that carries its
// directory and dependencies, takes its settings from the build, the
// environment, and the command line, and stops cleanly on SIGTERM.
func TestPSGIBuild(t *testing.T) {
	requireToolchain(t)
	startMirror(t, testDists(t)...)
	dir := t.TempDir()
	for name, content := range map[string]string{
		"cpanfile":        testCPANfile,
		"static/note.txt": "from the project directory",
		"app.psgi": `
use strict;
use Pure::Built;
use Demo::XS;
# The application's signal handling is its own: the server still stops on
# SIGINT.
$SIG{INT} = 'IGNORE';
sub {
    my ($env) = @_;
    open(my $fh, '<', 'static/note.txt') or die "note: $!";
    my $note = <$fh>;
    [200, ['Content-Type' => 'text/plain'],
     [join('|', Pure::Built::name(), Demo::XS::add(40, 2), $note, $ENV{PLACK_ENV}, $env->{'psgi.multiprocess'} ? 'multi' : 'single')]];
};
`,
	} {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GPERL_DEV_REPLACE", devReplaces(t))
	server := filepath.Join(t.TempDir(), "server")
	if err := gperl.PSGIBuild([]string{"-o", server, "--workers", "2", "-E", "staging", filepath.Join(dir, "app.psgi")}); err != nil {
		t.Fatalf("PSGIBuild: %v", err)
	}

	// serveUntil runs the server on a socket the test listens on (handed
	// over as Server::Starter does), sends it concurrent requests, sends
	// it stop, and returns the body every request got and how the server
	// ended.
	serveUntil := func(stop os.Signal, env []string, args ...string) (string, error) {
		t.Helper()
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		f, err := ln.(*net.TCPListener).File()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(server, args...)
		cmd.Dir = t.TempDir()
		cmd.Env = append(append(os.Environ(), "GPERLMODCACHE="+filepath.Join(t.TempDir(), "absent"),
			"SERVER_STARTER_PORT="+ln.Addr().String()+"=3"), env...)
		cmd.ExtraFiles = []*os.File{f}
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		if err := errors.Join(f.Close(), ln.Close()); err != nil {
			t.Fatal(err)
		}
		url := "http://" + ln.Addr().String() + "/"
		const clients, each = 16, 25
		bodies := make([]string, clients*each)
		var wg sync.WaitGroup
		for c := range clients {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range each {
					res, err := http.Get(url)
					if err != nil {
						t.Errorf("request: %v", err)
						return
					}
					data, rerr := io.ReadAll(res.Body)
					if err := errors.Join(rerr, res.Body.Close()); err != nil {
						t.Errorf("read response: %v", err)
						return
					}
					bodies[c*each+i] = string(data)
				}
			}()
		}
		wg.Wait()
		if err := cmd.Process.Signal(stop); err != nil {
			t.Fatal(err)
		}
		waitErr := cmd.Wait()
		if waitErr != nil {
			waitErr = fmt.Errorf("%w\n%s", waitErr, stderr.String())
		}
		for _, body := range bodies[1:] {
			if body != bodies[0] {
				t.Errorf("concurrent requests got %q and %q", bodies[0], body)
				break
			}
		}
		return bodies[0], waitErr
	}
	serve := func(env []string, args ...string) string {
		t.Helper()
		body, err := serveUntil(os.Interrupt, env, args...)
		if err != nil {
			t.Fatalf("server exit after SIGINT: %v", err)
		}
		return body
	}
	if got, want := serve(nil), "built on static|42|from the project directory|staging|multi"; got != want {
		t.Errorf("with the build-time defaults: %q, want %q", got, want)
	}
	if got, want := serve([]string{"GPERL_PSGI_ENV=production", "GPERL_PSGI_WORKERS=1"}), "built on static|42|from the project directory|production|single"; got != want {
		t.Errorf("with the environment: %q, want %q", got, want)
	}
	if got, want := serve([]string{"GPERL_PSGI_ENV=production"}, "-E", "development", "--workers", "3"), "built on static|42|from the project directory|development|multi"; got != want {
		t.Errorf("with the command line: %q, want %q", got, want)
	}
	// A signal the server leaves to the Go runtime does what it does to any
	// Go program — SIGQUIT ends it — rather than going to a worker.
	var exit *exec.ExitError
	if _, err := serveUntil(syscall.SIGQUIT, nil); !errors.As(err, &exit) {
		t.Errorf("server after SIGQUIT: %v, want it ended", err)
	}
}

// gperl psgi run loads the dependencies' native modules from @INC when
// they have some, and otherwise makes workers that load none.
func TestPSGIRunNativeModules(t *testing.T) {
	requireToolchain(t)
	startMirror(t, testDists(t)...)
	for _, c := range []struct {
		name     string
		cpanfile string
		app      string
		native   bool
		body     string
	}{
		{name: "native", cpanfile: "requires 'Demo::XS';\n", native: true, body: "42",
			app: "use Demo::XS;\nsub { [200, [], [Demo::XS::add(40, 2)]] };\n"},
		{name: "pure", native: false, body: "pure",
			app: "sub { [200, [], ['pure']] };\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if c.cpanfile != "" {
				if err := os.WriteFile(filepath.Join(dir, "cpanfile"), []byte(c.cpanfile), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			appPath := filepath.Join(dir, "app.psgi")
			if err := os.WriteFile(appPath, []byte(c.app), 0o644); err != nil {
				t.Fatal(err)
			}
			server, native, closeServer, err := gperl.PSGIServer(appPath)
			if err != nil {
				t.Fatal(err)
			}
			defer closeServer()
			if native != c.native {
				t.Errorf("workers load native modules = %v, want %v", native, c.native)
			}
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if got := rec.Body.String(); rec.Code != http.StatusOK || got != c.body {
				t.Errorf("response %d %q, want 200 %q", rec.Code, got, c.body)
			}
		})
	}
}
