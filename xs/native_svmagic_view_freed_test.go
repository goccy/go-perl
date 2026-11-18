//go:build darwin || linux

package xs_test

import (
	"bytes"
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

// What SvMAGIC showed native code of a value is done with once it has been
// written back: the interpreter may free the value and its magic from then
// on. A view kept past that was written back again at the next operation of
// native code, inside a native call that was still running (Object::Pad's
// parse of a class, where the namespace::autoclean of a module the class
// uses cleans a package): it walked the freed MAGIC of native code's chain,
// and the process died of SIGSEGV in the next keyword hook (Sub-HandlesVia
// t/50objectpad.t). The child process runs with the allocator filling freed
// memory, so that a walk of a freed MAGIC fails every time.
const svmagicViewModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static MGVTBL my_vtbl;
XS(XS_MyMv_attach) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    sv_magicext(SvRV(ST(0)), NULL, PERL_MAGIC_ext, &my_vtbl, NULL, 0);
    XSRETURN_EMPTY;
}
/* the number of MAGICs SvMAGIC shows of the referent */
XS(XS_MyMv_count) {
    dXSARGS;
    IV n = 0;
    MAGIC *mg;
    PERL_UNUSED_VAR(items);
    for (mg = SvMAGIC(SvRV(ST(0))); mg; mg = mg->mg_moremagic) n++;
    ST(0) = sv_2mortal(newSViv(n));
    XSRETURN(1);
}
/* an operation of native code's */
XS(XS_MyMv_touch) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(1));
    XSRETURN(1);
}
/* runs the code inside a call of native code */
XS(XS_MyMv_within) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    call_sv(ST(0), G_VOID | G_DISCARD | G_NOARGS);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MyMv) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyMv::attach", XS_MyMv_attach);
    newXS_deffile("MyMv::count", XS_MyMv_count);
    newXS_deffile("MyMv::touch", XS_MyMv_touch);
    newXS_deffile("MyMv::within", XS_MyMv_within);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

const svmagicViewCode = `use Scalar::Util ();
my @r;
MyMv::within(sub {
    {
        my $x = 1;
        my $w = \$x;
        Scalar::Util::weaken($w);   # the interpreter's magic on $x
        MyMv::attach(\$x);          # and native code's
        push @r, MyMv::count(\$x);
    }                               # $x goes, with its magic
    MyMv::touch() for 1 .. 3;
});
"count=$r[0] ok"`

func TestNativeSvMagicViewOfAFreedValue(t *testing.T) {
	if so := os.Getenv("GOPERL_SVMAGIC_CHILD"); so != "" {
		svmagicViewChild(t, so)
		return
	}
	so := buildModuleSource(t, "MyMv", svmagicViewModule)
	cmd := exec.Command(os.Args[0], "-test.run=^TestNativeSvMagicViewOfAFreedValue$")
	scribble := "MALLOC_PERTURB_=85"
	if runtime.GOOS == "darwin" {
		scribble = "MallocScribble=1"
	}
	cmd.Env = append(os.Environ(), "GOPERL_SVMAGIC_CHILD="+so, scribble)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	got := out.String()
	if i := strings.Index(got, "result: "); i >= 0 {
		got = got[i+len("result: "):]
		if j := strings.Index(got, "\n"); j >= 0 {
			got = got[:j]
		}
	}
	// perl: the '~' of native code and the back references of the weak
	// reference
	if want := "count=2 ok"; err != nil || got != want {
		errText := errOut.String()
		if len(errText) > 2000 {
			errText = errText[:2000]
		}
		t.Fatalf("err=%v got %q, want %q\nstderr %s", err, got, want, errText)
	}
}

func svmagicViewChild(t *testing.T, so string) {
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
	auto := filepath.Join(libdir, "auto", "MyMv")
	if err := os.MkdirAll(auto, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(so, filepath.Join(auto, "MyMv.so")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, "MyMv.pm"),
		[]byte("package MyMv;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := p.Eval(context.Background(), `unshift @INC, '`+libdir+`'; require MyMv; `+svmagicViewCode)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("Eval died: %v", r.Error)
	}
	os.Stdout.WriteString("result: " + resultStr(r) + "\n")
}
