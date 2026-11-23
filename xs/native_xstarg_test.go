//go:build darwin || linux

package xs_test

import "testing"

// An XSUB's TARG (dXSTARG) is the pad target of the entersub that calls it,
// as perl gives it: the same SV at one call site, call after call, and no
// SV made for the value a call returns. A new mortal for every call took
// the head of an SV freed just before, which perl gives to the next SV the
// XSUB makes: B::Hooks::OP::Check's hook ids are such addresses, and
// Unicode::Casing, removing a hook by the id an earlier one had, left the
// override of uc in place after `no Unicode::Casing` (t/01basic.t test 19).
// What is expected is what perl 5.36 on Linux gives with the same module.
func TestNativeXSTargIsThePadTarget(t *testing.T) {
	so := buildModuleSource(t, "SvTarg", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* An XSUB's TARG (dXSTARG) is the pad target of the entersub that calls
 * it, as perl gives it, not a new SV: make() makes an SV, keeps it and
 * returns its address through TARG; drop() frees the SV kept last. The
 * address of the SV made after one is freed is the freed one's, as perl
 * reuses the head (B::Hooks::OP::Check's hook ids are such addresses, and
 * Unicode::Casing removes a hook by an id an earlier hook had). */
static AV *keep;
XS(XS_SvTarg_make) {
    dXSARGS;
    dXSTARG;
    PERL_UNUSED_VAR(items);
    if (!keep) keep = newAV();
    SV *made = newSVuv(1);
    av_push(keep, made);
    XSprePUSH;
    PUSHu(PTR2UV(made));
    XSRETURN(1);
}
XS(XS_SvTarg_drop) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    av_delete(keep, av_len(keep), G_DISCARD);
    XSRETURN_EMPTY;
}
/* targ_is_pad(): whether TARG is the caller's pad target (a call that
 * gives one) */
XS(XS_SvTarg_twice) {
    dXSARGS;
    dXSTARG;
    PERL_UNUSED_VAR(items);
    XSprePUSH;
    PUSHu(PTR2UV(TARG));
    XSRETURN(1);
}
XS_EXTERNAL(boot_SvTarg) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("SvTarg::make", XS_SvTarg_make, __FILE__);
    newXS("SvTarg::drop", XS_SvTarg_drop, __FILE__);
    newXS("SvTarg::targ", XS_SvTarg_twice, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $a = SvTarg::make();
SvTarg::drop();
my $b = SvTarg::make();
my @t = map { SvTarg::targ() } 1 .. 2;
"after a free " . ($a == $b ? "same" : "other") . ", targ of one call site " . ($t[0] == $t[1] ? "same" : "other")`
	if got, want := requireModule(t, "SvTarg", so, program), "after a free same, targ of one call site same"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
