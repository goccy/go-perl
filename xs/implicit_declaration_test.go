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

// buildModuleSource compiles a native module from the C source code against
// the SDK, with the warnings perl's build gives and nothing made an error
// that perl's does not make one: a module whose code calls a function it
// does not declare compiles, as it does on perl.
func buildModuleSource(t *testing.T, name, code string) string {
	t.Helper()
	so, out, err := compileModuleSource(t, name, code)
	if err != nil {
		t.Fatalf("compile %s.so: %v\n%s", name, err, out)
	}
	return so
}

// compileModuleSource is buildModuleSource with more flags of the compiler,
// which gives what the compiler said when the build fails.
func compileModuleSource(t *testing.T, name, code string, flags ...string) (string, []byte, error) {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, name+".c")
	if err := os.WriteFile(src, []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}
	sdk, err := filepath.Abs("sdk/include")
	if err != nil {
		t.Fatal(err)
	}
	so := filepath.Join(dir, name+".so")
	args := append([]string{"-shared", "-fPIC", "-std=gnu99", `-DXS_VERSION="1.00"`, "-I", sdk, "-o", so}, flags...)
	args = append(args, src)
	if runtime.GOOS == "darwin" {
		args = append(args, "-undefined", "dynamic_lookup", "-Wno-error=implicit-function-declaration")
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	} else {
		args = append([]string{"-D_GNU_SOURCE"}, args...)
	}
	out, err := exec.Command(cc, args...).CombinedOutput()
	return so, out, err
}

// requireModule puts the module at so where DynaLoader finds it
// (auto/<name>/<name>.so beside <name>.pm in a directory of @INC) and
// requires it, returning what code then returns, or why the require died.
func requireModule(t *testing.T, name, so, code string) string {
	t.Helper()
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, NativeModules: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	libdir := t.TempDir()
	auto := filepath.Join(libdir, "auto", name)
	if err := os.MkdirAll(auto, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(so, filepath.Join(auto, name+".so")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, name+".pm"),
		[]byte("package "+name+";\nour $VERSION = '1.00';\nuse XSLoader;\nXSLoader::load(__PACKAGE__, $VERSION);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := p.Eval(context.Background(), `unshift @INC, '`+libdir+`'; my $r = eval { require `+name+`; `+code+` }; defined $r ? $r : "died: $@"`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("Eval died: %v", r.Error)
	}
	return resultStr(r)
}

// A module whose C calls a function of its own that it does not declare
// (perl's build warns, and the module works) builds and works under the
// SDK. Before, the SDK made the implicit declaration an error.
func TestNativeModuleCallingItsOwnUndeclaredFunction(t *testing.T) {
	so := buildModuleSource(t, "OwnUndeclared", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_OwnUndeclared_sum) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "a, b");
    ST(0) = sv_2mortal(newSViv(own_undeclared_sum((int)SvIV(ST(0)), (int)SvIV(ST(1)))));
    XSRETURN(1);
}
int own_undeclared_sum(int a, int b) { return a + b; }
XS_EXTERNAL(boot_OwnUndeclared) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("OwnUndeclared::sum", XS_OwnUndeclared_sum);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	if got := requireModule(t, "OwnUndeclared", so, `OwnUndeclared::sum(2, 3)`); got != "5" {
		t.Fatalf("got %q, want 5", got)
	}
}

// A module that calls a function of the perl API the SDK does not provide
// fails to load, and the error names the function: the loader resolves
// every symbol of a module when it loads it, so nothing of such a module
// runs, and nothing fails at a first call later.
func TestNativeModuleCallingAnAPIFunctionTheSDKLacks(t *testing.T) {
	so := buildModuleSource(t, "LacksAPI", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_LacksAPI_call) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    Perl_goperl_sdk_has_no_such_function(aTHX_ ST(0));
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_LacksAPI) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("LacksAPI::call", XS_LacksAPI_call);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	got := requireModule(t, "LacksAPI", so, `"loaded"`)
	if !strings.HasPrefix(got, "died: ") || !strings.Contains(got, "Perl_goperl_sdk_has_no_such_function") {
		t.Fatalf("got %q, want the require to die naming the missing function", got)
	}
}
