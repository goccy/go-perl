//go:build darwin || linux

package xs_test

import "testing"

// Native code reads the shared key behind a hash key's scalar, and the
// spelling of a v-string, as Sereal::Encoder does: SvSHARED_HEK_FROM_PV
// gives the HEK whose key the scalar's string is (its length, its hash, the
// same bytes), and SvVSTRING_mg the v-string magic with the literal's
// spelling. Before, neither name was there, and Sereal::Encoder did not
// load.
const sharedKeyVstringSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MySk_hek) {
    dXSARGS;
    SV *const sv = ST(0);
    PERL_UNUSED_VAR(items);
    if (SvIsCOW_shared_hash(sv)) {
        HEK *const hek = SvSHARED_HEK_FROM_PV(SvPVX(sv));
        ST(0) = sv_2mortal(newSVpvf("%d %s %s", (int)HEK_LEN(hek),
                                    HEK_KEY(hek) == SvPVX(sv) ? "same" : "other",
                                    HEK_HASH(hek) == SvSHARED_HASH(sv) ? "hash" : "nohash"));
    } else {
        ST(0) = sv_2mortal(newSVpvs("unshared"));
    }
    XSRETURN(1);
}
XS(XS_MySk_vs) {
    dXSARGS;
    MAGIC *const mg = SvVSTRING_mg(ST(0));
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(mg ? newSVpvn(mg->mg_ptr, mg->mg_len) : newSVpvs("none"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MySk) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySk::hek", XS_MySk_hek);
    newXS_deffile("MySk::vs", XS_MySk_vs);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeReadsSharedKeysAndVstrings(t *testing.T) {
	so := buildModuleSource(t, "MySk", sharedKeyVstringSource)
	got := requireModule(t, "MySk", so, `my %h = (alpha => 1); my ($k) = keys %h;
my $v = v1.22.333; my $s = "x";
join "; ", MySk::hek($k), MySk::hek("lit" . $s), MySk::vs($v), MySk::vs($s)`)
	if want := "5 same hash; unshared; v1.22.333; none"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
