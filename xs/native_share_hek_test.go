//go:build darwin || linux

package xs_test

import "testing"

// share_hek_hek of a hash's key shares it once more: a COW scalar of LEN 0
// with the key as its string keeps the key alive after the hash lets it
// go, and freeing the scalar unshares it without a warning, as Sereal::
// Encoder keeps a deleted key whose address names a COPY target. Before,
// share_hek_hek counted nothing: the key was freed with the hash, and a
// new key could take its memory.
const shareHekSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MyHk_keep) {
    dXSARGS;
    HV *const hv = (HV *)SvRV(ST(0));
    HE *he;
    HEK *hek;
    SV *keeper;
    PERL_UNUSED_VAR(items);
    hv_iterinit(hv);
    he = hv_iternext(hv);
    hek = HeKEY_hek(he);
    keeper = newSV_type(SVt_PV);
    SvPV_set(keeper, (char *)HEK_KEY(share_hek_hek(hek)));
    SvCUR_set(keeper, HEK_LEN(hek));
    SvLEN_set(keeper, 0);
    SvIsCOW_on(keeper);
    SvPOK_on(keeper);
    ST(0) = sv_2mortal(newRV_noinc(keeper));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyHk) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyHk::keep", XS_MyHk_keep);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeSharesAKeyOnceMore(t *testing.T) {
	so := buildModuleSource(t, "MyHk", shareHekSource)
	got := requireModule(t, "MyHk", so, `my @w; local $SIG{__WARN__} = sub { push @w, @_ };
my %h = (pack("a32", "old-key") => 1);
my $k = MyHk::keep(\%h);
%h = ();
my %h2 = map { (pack("a32", "new-$_") => $_) } 1 .. 20;
my $kept = $$k =~ s/\0+$//r;
undef $k;
"$kept warnings " . scalar(@w)`)
	if want := "old-key warnings 0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
