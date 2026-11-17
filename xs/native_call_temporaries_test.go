//go:build darwin || linux

package xs_test

import "testing"

// call_sv with G_DISCARD leaves no temporary behind, as perl's does: a walk
// that calls back for each element inside one scope of temporaries (HAMT's
// each, with G_DISCARD|G_EVAL) sees the same count of temporaries at every
// call. The results came over in a mortal array of the interpreter's, made
// for none too, one more for each call until the walk's FREETMPS (HAMT's
// count ran from 2 to 2001 over 2000 calls). What is expected is what perl
// 5.36 on Linux gives with the same module.
func TestNativeCallsLeaveNoTemporaries(t *testing.T) {
	so := buildModuleSource(t, "Tmps", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* walk(code, n, flags): calls code n times in one scope of temporaries, as
 * HAMT's each walks a map: with G_DISCARD|G_EVAL (flags 0) or for a scalar
 * it drops (flags 1) */
XS(XS_Tmps_walk) {
    dXSARGS;
    SV *code = ST(0);
    IV n = SvIV(ST(1)), i;
    int scalar = items > 2 && SvIV(ST(2));
    SV *arg;
    ENTER;
    SAVETMPS;
    arg = sv_newmortal(); /* one for all the calls, as HAMT's key */
    for (i = 0; i < n; i++) {
        dSP;
        I32 got;
        sv_setiv(arg, i);
        PUSHMARK(SP);
        XPUSHs(arg);
        PUTBACK;
        got = call_sv(code, scalar ? G_SCALAR | G_EVAL : G_DISCARD | G_VOID | G_EVAL);
        SPAGAIN;
        SP -= got;
        PUTBACK;
    }
    FREETMPS;
    LEAVE;
    XSRETURN_EMPTY;
}
/* tmps_ix(): how many temporaries perl holds */
XS(XS_Tmps_tmps_ix) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv((IV)PL_tmps_ix));
    XSRETURN(1);
}
XS_EXTERNAL(boot_Tmps) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("Tmps::walk", XS_Tmps_walk, __FILE__);
    newXS("Tmps::tmps_ix", XS_Tmps_tmps_ix, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my ($lo, $hi);
Tmps::walk(sub {
    my $t = Tmps::tmps_ix();
    $lo = $t if !defined $lo || $t < $lo;
    $hi = $t if !defined $hi || $t > $hi;
    return;
}, 500, 0);
"spread " . ($hi - $lo)`
	if got, want := requireModule(t, "Tmps", so, program), "spread 0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
