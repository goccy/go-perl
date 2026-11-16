//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
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

// newHostPerl builds an instance on the host filesystem (acceptance dists,
// their vendored trees, and test files live on disk).
func newHostPerl(t testing.TB) *perl.Perl {
	t.Helper()
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { p.Close() })
	return p
}

func TestNativeXSModule(t *testing.T) {
	so := buildDemo(t)
	p := newHostPerl(t)
	ctx := context.Background()

	loadXSModule(t, p, "Demo::XS", so)

	// `use Demo::XS` resolves the .pm half from @INC like any module.
	libdir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(libdir, "Demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, "Demo", "XS.pm"),
		[]byte("package Demo::XS;\nour $VERSION = '0.01';\nuse XSLoader;\nXSLoader::load(__PACKAGE__, $VERSION);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r, err := p.Eval(ctx, `sub __t_add_inc { unshift @INC, $_[0]; 1 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define inc helper: err=%v error=%v", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_add_inc", perl.ValueOf(libdir)); err != nil {
		t.Fatalf("add inc: %v", err)
	}

	r, err := p.Eval(ctx, `use Demo::XS; Demo::XS::add(40, 2)`)
	if err != nil {
		t.Fatalf("Eval use+add: %v", err)
	}
	if r.Error != nil || resultStr(r) != "42" {
		t.Fatalf("add = ok=%v result=%q error=%v", (r.Error == nil), resultStr(r), r.Error)
	}

	r, err = p.Eval(ctx, `Demo::XS::greet("gopher")`)
	if err != nil {
		t.Fatalf("Eval greet: %v", err)
	}
	if r.Error != nil || resultStr(r) != "hello, gopher" {
		t.Fatalf("greet = ok=%v result=%q error=%v", (r.Error == nil), resultStr(r), r.Error)
	}

	// Wrong arity: the SDK's croak crosses back as an ordinary Perl die.
	r, err = p.Eval(ctx, `eval { Demo::XS::add(1) }; $@`)
	if err != nil {
		t.Fatalf("Eval usage croak: %v", err)
	}
	if r.Error != nil || !strings.Contains(resultStr(r), "Usage: Demo::XS::add(a, b)") {
		t.Fatalf("usage croak = %q (ok=%v error=%v)", resultStr(r), (r.Error == nil), r.Error)
	}

	// Native XSUBs interleave with everything else on the instance.
	got, err := p.Call(ctx, "Demo::XS::add", perl.ValueOf(20), perl.ValueOf(22))
	if err != nil {
		t.Fatalf("Call add: %v", err)
	}
	if sv, err := perl.As[perl.ScalarValue](got[0]); err != nil || sv.Int() != 42 {
		t.Fatalf("Call add = %#v (err=%v), want 42", got[0], err)
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

	p := newHostPerl(t)
	ctx := context.Background()
	loadXSModule(t, p, "Obj::Demo", so)
	libdir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(libdir, "Obj"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, "Obj", "Demo.pm"), []byte(
		"package Obj::Demo;\nour $VERSION = '0.01';\nuse XSLoader;\nXSLoader::load(__PACKAGE__, $VERSION);\n"+
			"sub new { my ($class, $label) = @_; $class->_new($label) }\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r, err := p.Eval(ctx, `sub __t_inc2 { unshift @INC, $_[0]; 1 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("inc helper: err=%v error=%v", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_inc2", perl.ValueOf(libdir)); err != nil {
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
	if want := "42|Obj::Demo::Stats|hits|42"; r.Error != nil || resultStr(r) != want {
		t.Fatalf("ObjDemo = ok=%v result=%q error=%v, want %q", (r.Error == nil), resultStr(r), r.Error, want)
	}

	// dXSARGS in a function that was given no interpreter context.
	r, err = p.Eval(ctx, `join ",", Obj::Demo::count_args(7, 8, 9), Obj::Demo::count_args()`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if want := "3,0"; r.Error != nil || resultStr(r) != want {
		t.Fatalf("count_args = result=%q error=%v, want %q", resultStr(r), r.Error, want)
	}
}

// buildCxxPlain compiles testdata/cxxplain into a module: its C++ unit
// alone, or with mixed also its C unit (helper.c).
func buildCxxPlain(t *testing.T, mixed bool) string {
	t.Helper()
	cxx := findCXX(t)
	dir := t.TempDir()
	var arch []string
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			arch = []string{"-arch", "arm64"}
		case "amd64":
			arch = []string{"-arch", "x86_64"}
		}
	}
	run := func(name string, args ...string) {
		t.Helper()
		if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
	}
	objs := []string{filepath.Join(dir, "CxxPlain.o")}
	args := append(append(append([]string{}, arch...), cxx[1:]...),
		"-c", "-fPIC", "-x", "c++", "-std=c++11", "-I", "sdk/include", "-o", objs[0], "testdata/cxxplain/CxxPlain.cc")
	if mixed {
		args = append(args, "-DCXXPLAIN_MIXED")
	}
	run(cxx[0], args...)
	if mixed {
		cc, err := exec.LookPath("cc")
		if err != nil {
			t.Skip("no host C compiler available")
		}
		objs = append(objs, filepath.Join(dir, "helper.o"))
		run(cc, append(append([]string{}, arch...),
			"-c", "-fPIC", "-I", "sdk/include", "-o", objs[1], "testdata/cxxplain/helper.c")...)
	}
	so := filepath.Join(dir, "CxxPlain.so")
	run(cxx[0], append(append(append(append([]string{}, arch...), cxx[1:]...), "-shared", "-o", so), objs...)...)
	return so
}

// A C++ module that includes the perl headers outside extern "C", after
// C++ headers of its own (Algorithm::LibLinear), loads and runs: the
// loader's entry points it defines have their C names. They had mangled
// ones then, the loader did not initialize the module, and its boot
// function faulted on the missing vtable. A module of a C++ unit and a C
// unit shares one vtable between them: with mangled names each language's
// units had their own, and the C++ unit's was never set.
func TestNativeCxxModuleWithoutExternC(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		so := buildCxxPlain(t, mixed)
		p := newHostPerl(t)
		loadXSModule(t, p, "CxxPlain", so)
		libdir := t.TempDir()
		if err := os.WriteFile(filepath.Join(libdir, "CxxPlain.pm"),
			[]byte("package CxxPlain;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		code := `unshift @INC, '` + libdir + `'; require CxxPlain; CxxPlain::greet("perl")`
		want := "hello, perl"
		if mixed {
			code += ` . "|" . CxxPlain::twice(21)`
			want += "|42"
		}
		r, err := p.Eval(context.Background(), code)
		if err != nil {
			t.Fatalf("mixed=%v: Eval: %v", mixed, err)
		}
		if r.Error != nil || resultStr(r) != want {
			t.Errorf("mixed=%v: got %q (error %v), want %q", mixed, resultStr(r), r.Error, want)
		}
	}
}

// The SDK headers compile as C++ with every C++ compiler of the host: a
// module includes perl.h and XSUB.h inside extern "C" and says dXSARGS in
// functions with and without an interpreter context. (g++ refuses what
// clang++ lets pass, overloads with C linkage among them.)
func TestSDKCompilesAsCxx(t *testing.T) {
	tried := 0
	for _, cxx := range []string{"c++", "g++", "clang++"} {
		path, err := exec.LookPath(cxx)
		if err != nil {
			continue
		}
		args := []string{"-fsyntax-only", "-x", "c++", "-std=c++11", "-I", "sdk/include", "testdata/objdemo/ObjDemo.c"}
		if runtime.GOOS == "darwin" {
			if out, err := exec.Command("xcrun", "--show-sdk-path").Output(); err == nil {
				args = append([]string{"-isysroot", strings.TrimSpace(string(out))}, args...)
			}
		}
		// Some hosts have a C++ compiler whose standard library headers
		// are not installed: one that cannot compile a program of its own
		// says nothing about the SDK's.
		probe := filepath.Join(t.TempDir(), "probe.cpp")
		if err := os.WriteFile(probe, []byte("#include <string>\nint main(){return 0;}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		probeArgs := append(append([]string{}, args[:len(args)-3]...), probe)
		if exec.Command(path, probeArgs...).Run() != nil {
			continue
		}
		tried++
		if out, err := exec.Command(path, args...).CombinedOutput(); err != nil {
			t.Errorf("%s: %v\n%s", cxx, err, out)
		}
	}
	if tried == 0 {
		t.Skip("no working C++ compiler found")
	}
}
