//go:build darwin || linux

package xs_test

import "testing"

// A fetch, store, exists or delete of native code's on a hash it freed
// already answers nothing, as perl's hv_common answers for a freed hash
// (SvIS_FREED), and the program goes on: Prima tears its font cache down
// after the hash of the faces the cache looks up, at its END. Before,
// gperl refused the freed hash ("that scalar was freed already") and
// Prima's test files ended with exit 255 after every test had passed.
const freedHashModuleSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_FreedHash_use) {
    dXSARGS;
    HV *hv = newHV();
    SV *key = sv_2mortal(newSVpvs("k"));
    HE *he;
    SV **svp;
    int exists;
    SV *deleted;
    SV **stored;
    SV *value = newSViv(2);
    PERL_UNUSED_VAR(items);
    (void)hv_store(hv, "k", 1, newSViv(1), 0);
    SvREFCNT_dec((SV *)hv);
    he = hv_fetch_ent(hv, key, 0, 0);
    svp = hv_fetch(hv, "k", 1, 0);
    exists = hv_exists_ent(hv, key, 0);
    deleted = hv_delete_ent(hv, key, G_DISCARD, 0);
    stored = hv_store(hv, "k", 1, value, 0);
    if (!stored) SvREFCNT_dec(value);
    ST(0) = sv_2mortal(newSVpvf("fetch_ent %s, fetch %s, exists %d, delete %s, store %s",
                                he ? "found" : "none", svp ? "found" : "none", exists,
                                deleted ? "found" : "none", stored ? "stored" : "none"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_FreedHash) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("FreedHash::use", XS_FreedHash_use);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeUsesAFreedHash(t *testing.T) {
	so := buildModuleSource(t, "FreedHash", freedHashModuleSource)
	got := requireModule(t, "FreedHash", so, `FreedHash::use() . "; after"`)
	if want := "fetch_ent none, fetch none, exists 0, delete none, store none; after"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
