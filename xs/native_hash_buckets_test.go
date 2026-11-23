//go:build darwin || linux

package xs_test

import "testing"

// HvARRAY gives the same buckets twice in one call while the hash is not
// split, also across a call of Perl code in between, and the bucket of the
// iterator's position is (HvRITER ^ HvAUX(hv)->xhv_rand) & HvMAX, as perl
// randomizes it (PERL_HASH_RANDOMIZE_KEYS): there the entry HvEITER holds
// is found. Sereal::Encoder does both to keep a caller's each() going
// across an encode. Before, every HvARRAY was a new array, and the SDK had
// neither the randomization nor xhv_rand, so the bucket was the position.
const hashBucketsSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MyHb_probe) {
    dXSARGS;
    HV *const hv = (HV *)SvRV(ST(0));
    HE **const before = HvARRAY(hv);
    HE **after;
    const I32 riter = HvRITER_get(hv);
    HE *const eiter = HvEITER_get(hv);
    HE *he = NULL;
    PERL_UNUSED_VAR(items);
    PUSHMARK(SP);
    call_sv(ST(1), G_DISCARD | G_NOARGS);
    after = HvARRAY(hv);
#ifdef PERL_HASH_RANDOMIZE_KEYS
    if (riter >= 0 && after) {
        const STRLEN bucket = ((STRLEN)riter ^ (STRLEN)HvAUX(hv)->xhv_rand) & HvMAX(hv);
        for (he = after[bucket]; he; he = HeNEXT(he))
            if (he == eiter) break;
    }
#endif
    ST(0) = sv_2mortal(newSVpvf("%s %s", before == after ? "same" : "other",
                                he && he == eiter ? "found" : "lost"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyHb) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyHb::probe", XS_MyHb_probe);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeFindsTheIteratorsBucket(t *testing.T) {
	so := buildModuleSource(t, "MyHb", hashBucketsSource)
	got := requireModule(t, "MyHb", so, `my %h = map { ("k$_" => $_) } 1 .. 50;
my $calls = 0;
my $r;
while (my ($k) = each %h) { $r = MyHb::probe(\%h, sub { $calls++ }); last }
keys %h;
"$r $calls"`)
	if want := "same found 1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
