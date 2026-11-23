//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// Two instances of the interpreter run native code at once on two threads:
// an interpreter whose native call uses the interpreter while it waits,
// and a copy of it that calls XSUBs meanwhile. Each runs its calls on its
// own frames. Before, a call into the copy's XSUBs took the other
// instance's waiting call for one it was made in, as a thread of native
// code's own does, and ran in it: the waiting call then found the copy's
// frame current and the two mixed their operations. The answers are the
// arithmetic of the calls, which perl gives for each program alone.
const instancesAtOnceModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <unistd.h>
static int holding;
static int calls;
/* hold(n): waits, using the interpreter all along, until count has been
 * called n times (ten seconds at most); returns how many times it was */
XS(XS_AtOnce_hold) {
    dXSARGS;
    int n, i;
    IV made = 0;
    if (items != 1) croak_xs_usage(cv, "n");
    n = (int)SvIV(ST(0));
    __atomic_store_n(&holding, 1, __ATOMIC_SEQ_CST);
    for (i = 0; i < 100000 && __atomic_load_n(&calls, __ATOMIC_SEQ_CST) < n; i++) {
        SV *sv = newSViv(i);
        made += SvIV(sv) == i;
        SvREFCNT_dec(sv);
        usleep(100);
    }
    __atomic_store_n(&holding, 0, __ATOMIC_SEQ_CST);
    if (made != i) croak("hold: values read back wrong");
    XSRETURN_IV(__atomic_load_n(&calls, __ATOMIC_SEQ_CST));
}
XS(XS_AtOnce_holding) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_IV(__atomic_load_n(&holding, __ATOMIC_SEQ_CST));
}
/* count(x): x + 1 */
XS(XS_AtOnce_count) {
    dXSARGS;
    IV x;
    if (items != 1) croak_xs_usage(cv, "x");
    x = SvIV(ST(0));
    __atomic_add_fetch(&calls, 1, __ATOMIC_SEQ_CST);
    XSRETURN_IV(x + 1);
}
XS_EXTERNAL(boot_AtOnce) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("AtOnce::hold", XS_AtOnce_hold, __FILE__);
    newXS("AtOnce::holding", XS_AtOnce_holding, __FILE__);
    newXS("AtOnce::count", XS_AtOnce_count, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCallsOfTwoInstancesAtOnce(t *testing.T) {
	so := buildModuleSource(t, "AtOnce", instancesAtOnceModule)
	p := newHostPerl(t)
	loadXSModule(t, p, "AtOnce", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "AtOnce.pm"),
		[]byte("package AtOnce;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if r, err := p.Eval(ctx, `unshift @INC, '`+libdir+`'; require AtOnce; 1`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	clone, err := p.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	defer clone.Close()
	held := make(chan string, 1)
	go func() {
		r, err := p.Eval(ctx, `AtOnce::hold(2000)`)
		switch {
		case err != nil:
			held <- "Eval: " + err.Error()
		case r.Error != nil:
			held <- "died: " + r.Error.Error()
		default:
			held <- resultStr(r)
		}
	}()
	r, err := clone.Eval(ctx, `select undef, undef, undef, 0.001 until AtOnce::holding();
my $sum = 0;
$sum += AtOnce::count($_) for 1 .. 2000;
$sum`)
	if err != nil || r.Error != nil {
		t.Fatalf("copy: err=%v error=%v", err, r.Error)
	}
	if got, want := resultStr(r), "2003000"; got != want {
		t.Errorf("copy: got %q, want %q", got, want)
	}
	if got, want := <-held, "2000"; got != want {
		t.Errorf("waiting call: got %q, want %q", got, want)
	}
}
