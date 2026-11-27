//go:build darwin || linux

package xs_test

import "testing"

// The context above the top of the context stack is the one perl popped
// last, and perl leaves it as it was: native code that runs just after a
// scope was left reads its type there, as Variable::Magic's free does to
// learn whether an eval was being left (vmg_svt_free: cxstack_ix <
// cxstack_max and CxTYPE(cxstack + cxstack_ix + 1) == CXt_EVAL, then
// PL_scopestack[PL_scopestack_ix] as the scope to leave). Before, that
// entry was the SDK's saved copy of the first context, so the free took an
// eval's way, with a scope of its own to leave, wherever the first context
// of the stack was an eval (any stack a call_sv runs on).
func TestNativeContextAboveTheTop(t *testing.T) {
	so := buildModuleSource(t, "Ghost", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_Ghost_type) {
    dXSARGS;
    const char *r = "none";
    PERL_UNUSED_VAR(items);
    if (cxstack_ix < cxstack_max) {
        switch (CxTYPE(cxstack + cxstack_ix + 1)) {
        case CXt_EVAL: r = "eval"; break;
        case CXt_SUB: r = "sub"; break;
        case CXt_BLOCK: r = "block"; break;
        case CXt_NULL: r = "null"; break;
        default: r = "other"; break;
        }
    }
    ST(0) = sv_2mortal(newSVpv(r, 0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_Ghost) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("Ghost::type", XS_Ghost_type);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	for _, tc := range []struct{ name, code, want string }{
		{"after an eval", `eval { 1 }; Ghost::type()`, "eval"},
		{"after a sub", `sub f { 1 } f(); Ghost::type()`, "sub"},
		{"after an eval in a sub", `sub g { eval { 1 }; Ghost::type() } g()`, "eval"},
	} {
		if got := requireModule(t, "Ghost", so, tc.code); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
