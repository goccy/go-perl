//go:build darwin || linux

package xsnative_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/xsnative"
)

// buildDemo compiles the checked-in xsubpp output of Demo.xs against the SDK
// headers with the host C compiler. This mirrors the intended pipeline:
// building happens ahead of time with an ordinary cc; the runtime only loads.
func buildDemo(t *testing.T) string {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	so := filepath.Join(t.TempDir(), "Demo.so")
	args := []string{"-shared", "-fPIC", "-I", "sdk/include", "-o", so, "testdata/demo/Demo.c"}
	// The .so must match the test binary's architecture, not the C
	// compiler's default (which follows the process tree — an amd64 Go
	// toolchain under Rosetta spawns an x86_64 clang).
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	}
	cmd := exec.Command(cc, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile Demo.so: %v\n%s", err, out)
	}
	return so
}

func TestNativeXSModule(t *testing.T) {
	so := buildDemo(t)
	p, err := perl.New(perl.Config{HostFS: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	ctx := context.Background()

	if err := xsnative.Load(p, "Demo::XS", so); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// `use Demo::XS` resolves the .pm half from @INC like any module.
	libdir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(libdir, "Demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, "Demo", "XS.pm"),
		[]byte("package Demo::XS;\nour $VERSION = '0.01';\nuse XSLoader;\nXSLoader::load(__PACKAGE__, $VERSION);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r, err := p.Eval(ctx, `sub __t_add_inc { unshift @INC, $_[0]; 1 } 1;`); err != nil || !r.Ok {
		t.Fatalf("define inc helper: err=%v error=%q", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_add_inc", libdir); err != nil {
		t.Fatalf("add inc: %v", err)
	}

	r, err := p.Eval(ctx, `use Demo::XS; Demo::XS::add(40, 2)`)
	if err != nil {
		t.Fatalf("Eval use+add: %v", err)
	}
	if !r.Ok || r.Result != "42" {
		t.Fatalf("add = ok=%v result=%q error=%q", r.Ok, r.Result, r.Error)
	}

	r, err = p.Eval(ctx, `Demo::XS::greet("gopher")`)
	if err != nil {
		t.Fatalf("Eval greet: %v", err)
	}
	if !r.Ok || r.Result != "hello, gopher" {
		t.Fatalf("greet = ok=%v result=%q error=%q", r.Ok, r.Result, r.Error)
	}

	// Wrong arity: the SDK's croak crosses back as an ordinary Perl die.
	r, err = p.Eval(ctx, `eval { Demo::XS::add(1) }; $@`)
	if err != nil {
		t.Fatalf("Eval usage croak: %v", err)
	}
	if !r.Ok || !strings.Contains(r.Result, "Usage: Demo::XS::add(a, b)") {
		t.Fatalf("usage croak = %q (ok=%v error=%q)", r.Result, r.Ok, r.Error)
	}

	// Native XSUBs interleave with everything else on the instance.
	got, err := p.Call(ctx, "Demo::XS::add", 20, 22)
	if err != nil {
		t.Fatalf("Call add: %v", err)
	}
	if got[0] != float64(42) {
		t.Fatalf("Call add = %#v, want 42", got[0])
	}
}

// findCXX locates a working C++ compiler: some hosts have a `c++` whose
// standard-library headers are broken, so candidates are probe-compiled
// (with the platform sysroot when needed) before being trusted.
func findCXX(t *testing.T) []string {
	t.Helper()
	probe := filepath.Join(t.TempDir(), "probe.cpp")
	if err := os.WriteFile(probe, []byte("#include <cstdio>\nint main(){return 0;}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var sysroot []string
	if runtime.GOOS == "darwin" {
		if out, err := exec.Command("xcrun", "--show-sdk-path").Output(); err == nil {
			sysroot = []string{"-isysroot", strings.TrimSpace(string(out))}
		}
	}
	candidates := [][]string{{"c++"}, {"clang++"}, {"g++"},
		{"/opt/homebrew/opt/llvm/bin/clang++"}, {"/usr/local/opt/llvm/bin/clang++"}}
	if env := os.Getenv("CXX"); env != "" {
		candidates = append([][]string{{env}}, candidates...)
	}
	out := filepath.Join(t.TempDir(), "probe.o")
	for _, c := range candidates {
		if _, err := exec.LookPath(c[0]); err != nil {
			continue
		}
		args := append(append([]string{}, c[1:]...), sysroot...)
		args = append(args, "-c", probe, "-o", out)
		if exec.Command(c[0], args...).Run() == nil {
			return append(append([]string{c[0]}, c[1:]...), sysroot...)
		}
	}
	t.Skip("no working C++ compiler found")
	return nil
}

// TestNativeCppObjectModule pins the SDK v2 surface with a C++ module: a
// T_PTROBJ native object (host pointer as registry id in a blessed IV ref),
// method dispatch on it, and a blessed hash built with the AV/HV ops.
func TestNativeCppObjectModule(t *testing.T) {
	cxx := findCXX(t)
	so := filepath.Join(t.TempDir(), "ObjDemo.so")
	args := append(append([]string{}, cxx[1:]...),
		"-shared", "-fPIC", "-x", "c++", "-std=c++11",
		"-I", "sdk/include", "-o", so, "testdata/objdemo/ObjDemo.c")
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	}
	if out, err := exec.Command(cxx[0], args...).CombinedOutput(); err != nil {
		t.Fatalf("compile ObjDemo.so: %v\n%s", err, out)
	}

	p, err := perl.New(perl.Config{HostFS: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	ctx := context.Background()
	if err := xsnative.Load(p, "Obj::Demo", so); err != nil {
		t.Fatalf("Load: %v", err)
	}
	libdir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(libdir, "Obj"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, "Obj", "Demo.pm"), []byte(
		"package Obj::Demo;\nour $VERSION = '0.01';\nuse XSLoader;\nXSLoader::load(__PACKAGE__, $VERSION);\n"+
			"sub new { my ($class, $label) = @_; $class->_new($label) }\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r, err := p.Eval(ctx, `sub __t_inc2 { unshift @INC, $_[0]; 1 } 1;`); err != nil || !r.Ok {
		t.Fatalf("inc helper: err=%v error=%q", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_inc2", libdir); err != nil {
		t.Fatal(err)
	}

	r, err := p.Eval(ctx, `
		use Obj::Demo;
		my $o = Obj::Demo->new("hits");
		$o->incr(5);
		my $n = $o->incr(37);
		my $s = $o->stats;
		my $out = join("|", $n, ref($s), $s->{label}, $s->{count});
		undef $o;   # DESTROY tears the native object down
		$out;
	`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if want := "42|Obj::Demo::Stats|hits|42"; !r.Ok || r.Result != want {
		t.Fatalf("ObjDemo = ok=%v result=%q error=%q, want %q", r.Ok, r.Result, r.Error, want)
	}
}
