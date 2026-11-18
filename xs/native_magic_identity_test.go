//go:build darwin || linux

package xs_test

import "testing"

// Two lookups of the interpreter's own magic with mg_find give the same
// MAGIC, also across a call of Perl code in between (a tied FETCH), as on
// perl, where the pointer is the magic's own: Sereal::Encoder compares the
// tie magic it found before a callback with the one after to tell a
// container was retied. The buffers of the boolean scalars are PL_Yes and
// PL_No, which Sereal::Encoder compares a boolean's string with to write
// true or false. Before, every mg_find gave a new copy, and PL_Yes and
// PL_No were strings of the module's.
const magicIdentitySource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* tied(\%h): the tie magic before and after a fetch through it */
XS(XS_MyMi_tied) {
    dXSARGS;
    HV *const hv = (HV *)SvRV(ST(0));
    MAGIC *const before = mg_find((SV *)hv, PERL_MAGIC_tied);
    SV **const v = hv_fetchs(hv, "k", 0);
    MAGIC *const after = mg_find((SV *)hv, PERL_MAGIC_tied);
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%s %s %s", before == after ? "same" : "other",
                                before && after && before->mg_obj == after->mg_obj ? "obj" : "noobj",
                                v ? SvPV_nolen(*v) : "none"));
    XSRETURN(1);
}
/* truth(v): which of PL_Yes and PL_No a value's string is */
XS(XS_MyMi_truth) {
    dXSARGS;
    STRLEN len;
    const char *const pv = SvPV(ST(0), len);
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpv(pv == PL_Yes ? "yes" : pv == PL_No ? "no" : "other", 0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyMi) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyMi::tied", XS_MyMi_tied);
    newXS_deffile("MyMi::truth", XS_MyMi_truth);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeFindsTheSameMagicAndTheBooleanStrings(t *testing.T) {
	so := buildModuleSource(t, "MyMi", magicIdentitySource)
	got := requireModule(t, "MyMi", so, `package Tie::Count { sub TIEHASH { bless {}, shift } sub FETCH { "v$_[1]" } }
package main;
tie my %h, "Tie::Count";
join "; ", MyMi::tied(\%h), MyMi::truth(!0), MyMi::truth(!1), MyMi::truth("x")`)
	if want := "same obj vk; yes; no; other"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
