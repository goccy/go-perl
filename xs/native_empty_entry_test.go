//go:build darwin || linux

package xs_test

import "testing"

// A fetch with HV_FETCH_LVALUE | HV_FETCH_EMPTY_HE makes an entry with no
// value, which is how List::Util's uniqnum tells a key it has not seen
// from one it has: the value is NULL until the module gives it one.
// Before, the entry got an undef value, every key looked seen, and uniqnum
// returned an empty list (List-Uniqnum).
func TestNativeEmptyEntryOfAFetch(t *testing.T) {
	so := buildModuleSource(t, "EmptyEntry", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_EmptyEntry_uniq) {
    dXSARGS;
    {
        HV *seen = (HV *)sv_2mortal((SV *)newHV());
        I32 i, n = 0;
        for (i = 0; i < items; i++) {
            STRLEN len;
            const char *key = SvPV(ST(i), len);
            HE *he = (HE *)hv_common(seen, NULL, key, len, 0,
                                     HV_FETCH_LVALUE | HV_FETCH_EMPTY_HE, NULL, 0);
            if (HeVAL(he)) continue;
            HeVAL(he) = &PL_sv_undef;
            ST(n++) = ST(i);
        }
        XSRETURN(n);
    }
}
XS_EXTERNAL(boot_EmptyEntry) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("EmptyEntry::uniq", XS_EmptyEntry_uniq);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	if got, want := requireModule(t, "EmptyEntry", so, `join ",", EmptyEntry::uniq(qw(a b a c b d))`), "a,b,c,d"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
