//go:build darwin || linux

package xs_test

import "testing"

// sv_grow, perl's function, gives a new scalar a string body and a buffer
// whatever the length, and SvGROW, the macro, calls it only for a buffer
// too short: Text::Xslate makes "" by sv_grow(sv, 0) and SvPOK_on on a new
// scalar, which was undef (its uri_escape("")). What is expected is what
// perl 5.36 on Linux gives with the same module.
func TestNativeSvGrowOfANewScalar(t *testing.T) {
	so := buildModuleSource(t, "Grow", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* empty(): "" as Text::Xslate makes it, by sv_grow(sv, 0) and SvPOK_on on
 * a new scalar */
XS(XS_Grow_empty) {
    dXSARGS;
    SV *dest = sv_newmortal();
    PERL_UNUSED_VAR(items);
    sv_grow(dest, 0);
    SvPOK_on(dest);
    ST(0) = dest;
    XSRETURN(1);
}
/* grown(n): whether sv_grow gives a new scalar a string body and a buffer
 * for n bytes, and whether SvGROW of a string asks for no more than it has */
XS(XS_Grow_grown) {
    dXSARGS;
    SV *a = sv_newmortal(), *b = sv_2mortal(newSVpvs("abc"));
    STRLEN n = (STRLEN)SvUV(ST(0));
    char *before;
    PERL_UNUSED_VAR(items);
    sv_grow(a, n);
    before = SvPVX(b);
    ST(0) = sv_2mortal(newSVpvf("%s/%s", SvTYPE(a) == SVt_PV && SvPVX(a) ? "pv" : "other",
                                SvGROW(b, 1) == before ? "same" : "moved"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_Grow) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("Grow::empty", XS_Grow_empty, __FILE__);
    newXS("Grow::grown", XS_Grow_grown, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $e = Grow::empty();
join ",", (defined $e ? "defined" : "undef"), length($e), Grow::grown(0), Grow::grown(10)`
	if got, want := requireModule(t, "Grow", so, program), "defined,0,pv/same,pv/same"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
