//go:build darwin || linux

package xs_test

import "testing"

// The copies mg_find and HvARRAY make in one call of native code are found
// again by the address of what they copy, among many: a call that goes
// through the magic of, or the buckets of, 600 values twice reads in the
// second pass what it read in the first, with the copies' table grown on
// the way (Sereal::Encoder goes through HvARRAY for every hash of the
// structure it encodes, in one call).
const calltabSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* tag($x, $n): ext magic on $x with mg_len $n */
XS(XS_MyCt_tag) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    sv_magicext(ST(0), NULL, PERL_MAGIC_ext, NULL, NULL, (I32)SvIV(ST(1)));
    XSRETURN_EMPTY;
}
/* tags(\@refs): the sums of the mg_len of every referent's ext magic, in
 * two passes of one call */
XS(XS_MyCt_tags) {
    dXSARGS;
    AV *const av = (AV *)SvRV(ST(0));
    IV sum[2] = {0, 0};
    PERL_UNUSED_VAR(items);
    for (int pass = 0; pass < 2; pass++) {
        for (SSize_t i = 0; i <= av_len(av); i++) {
            MAGIC *const mg = mg_find(SvRV(*av_fetch(av, i, 0)), PERL_MAGIC_ext);
            if (mg) sum[pass] += mg->mg_len;
        }
    }
    ST(0) = sv_2mortal(newSVpvf("%" IVdf " %" IVdf, sum[0], sum[1]));
    XSRETURN(1);
}
/* walk(\@hashes): the sums of the values of every hash, read through
 * HvARRAY, in two passes of one call */
XS(XS_MyCt_walk) {
    dXSARGS;
    AV *const av = (AV *)SvRV(ST(0));
    IV sum[2] = {0, 0};
    PERL_UNUSED_VAR(items);
    for (int pass = 0; pass < 2; pass++) {
        for (SSize_t i = 0; i <= av_len(av); i++) {
            HV *const hv = (HV *)SvRV(*av_fetch(av, i, 0));
            HE **const arr = HvARRAY(hv);
            for (STRLEN b = 0; arr && b <= HvMAX(hv); b++) {
                for (HE *he = arr[b]; he; he = HeNEXT(he)) sum[pass] += SvIV(HeVAL(he));
            }
        }
    }
    ST(0) = sv_2mortal(newSVpvf("%" IVdf " %" IVdf, sum[0], sum[1]));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyCt) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyCt::tag", XS_MyCt_tag);
    newXS_deffile("MyCt::tags", XS_MyCt_tags);
    newXS_deffile("MyCt::walk", XS_MyCt_walk);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeFindsCopiesInATable(t *testing.T) {
	so := buildModuleSource(t, "MyCt", calltabSource)
	got := requireModule(t, "MyCt", so, `my @refs = map { my $x = $_; MyCt::tag($x, $_); \$x } 1 .. 600;
my @hashes = map { +{ a => $_, b => 1, c => 2 } } 1 .. 600;
join " | ", MyCt::tags(\@refs), MyCt::walk(\@hashes)`)
	if want := "180300 180300 | 182100 182100"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
