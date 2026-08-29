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
	p, err := perl.New(perl.Config{})
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
		[]byte("package Demo::XS;\nour $VERSION = '0.01';\n1;\n"), 0o644); err != nil {
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
