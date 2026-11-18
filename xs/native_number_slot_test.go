//go:build darwin || linux

package xs_test

import "testing"

// The number fields of a scalar (SvIVX, SvNVX) are the scalar's slots, as
// on perl: an address native code keeps of one designates the field in
// later calls, however many other fields it designates meanwhile, and a
// store through it is the scalar's at once (Async::Interrupt keeps
// &SvIVX(value) and stores into it from later calls and from a signal
// handler), until the scalar takes another body, as on perl (a scalar
// read as a string does, and a fraction added to). Before, the address was a mirror slot that went
// to another field once enough others were designated, and the store
// landed there.
const numberSlotSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static IV *kept_iv;
static NV *kept_nv;
XS(XS_MyNs_keep) {
    dXSARGS;
    SV *const iv = SvRV(ST(0));
    SV *const nv = SvRV(ST(1));
    PERL_UNUSED_VAR(items);
    sv_setiv(iv, 0);
    SvIOK_only(iv);
    sv_setnv(nv, 0.0);
    SvNOK_only(nv);
    kept_iv = &SvIVX(iv);
    kept_nv = &SvNVX(nv);
    XSRETURN_EMPTY;
}
/* churn(\@refs, \@nums): the referents of the references, and every number
 * one more, through the fields */
XS(XS_MyNs_churn) {
    dXSARGS;
    AV *const refs = (AV *)SvRV(ST(0));
    AV *const nums = (AV *)SvRV(ST(1));
    SSize_t i;
    IV seen = 0;
    PERL_UNUSED_VAR(items);
    for (i = 0; i <= av_len(refs); i++)
        if (SvRV(*av_fetch(refs, i, 0))) seen++;
    for (i = 0; i <= av_len(nums); i++) {
        SV *const e = *av_fetch(nums, i, 0);
        SvIVX(e) = SvIVX(e) + 1;
    }
    XSRETURN_IV(seen);
}
XS(XS_MyNs_bump) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    *kept_iv += SvIV(ST(0));
    *kept_nv += SvNV(ST(1));
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MyNs) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyNs::keep", XS_MyNs_keep);
    newXS_deffile("MyNs::churn", XS_MyNs_churn);
    newXS_deffile("MyNs::bump", XS_MyNs_bump);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeKeepsTheAddressOfANumberField(t *testing.T) {
	so := buildModuleSource(t, "MyNs", numberSlotSource)
	got := requireModule(t, "MyNs", so, `my ($i, $n);
MyNs::keep(\$i, \$n);
my @refs = map { \my $x } 1 .. 40;
my @nums = map { $_ + 0 } 1 .. 40;
my $seen = MyNs::churn(\@refs, \@nums);
MyNs::bump(5, 0.5);
my $mid = $i + 0; # not $n: perl's addition gives a fraction a body for an IV too
$seen += MyNs::churn(\@refs, \@nums);
MyNs::bump(2, 0.25);
my $sum = 0;
$sum += $_ for @nums;
"$mid; $i $n; seen $seen sum $sum"`)
	if want := "5; 7 0.75; seen 80 sum 900"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
