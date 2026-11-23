//go:build darwin || linux

package xs_test

import "testing"

// The names Scope::Upper uses besides the contexts: save_alloc makes room
// on the save stack in whole entries under a SAVEt_ALLOC entry and says
// where it begins, leave_scope undoes the saves above an index, save_aelem
// localizes an element of an array, a given block's context has its leave
// op (blk_givwhen), and PL_DBgv is the interpreter's. Before, save_alloc
// and leave_scope were refused as internals, and save_aelem, blk_givwhen
// and PL_DBgv were not there.
const scopeUpperNamesSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MySu_alloc) {
    dXSARGS;
    I32 before, after, back;
    SSize_t start;
    PERL_UNUSED_VAR(items);
    ENTER;
    before = PL_savestack_ix;
    start = save_alloc(3 * sizeof(ANY), 0);
    after = PL_savestack_ix;
    leave_scope(before);
    back = PL_savestack_ix;
    LEAVE;
    ST(0) = sv_2mortal(newSVpvf("%d %d %d", (int)(after - before),
                                start == (SSize_t)before * (SSize_t)sizeof(ANY) ? 1 : 0,
                                back == before ? 1 : 0));
    XSRETURN(1);
}
/* aelem(\@a, i): the element localized to 9 inside a scope, and after it */
XS(XS_MySu_aelem) {
    dXSARGS;
    AV *const av = (AV *)SvRV(ST(0));
    const I32 idx = (I32)SvIV(ST(1));
    SV **svp;
    IV inside;
    PERL_UNUSED_VAR(items);
    ENTER;
    svp = av_fetch(av, idx, 1);
    save_aelem(av, idx, svp);
    sv_setiv(*av_fetch(av, idx, 0), 9);
    inside = SvIV(*av_fetch(av, idx, 0));
    LEAVE;
    ST(0) = sv_2mortal(newSVpvf("%d %d", (int)inside, (int)SvIV(*av_fetch(av, idx, 0))));
    XSRETURN(1);
}
XS(XS_MySu_given) {
    dXSARGS;
    I32 ix;
    PERL_UNUSED_VAR(items);
    for (ix = cxstack_ix; ix >= 0; ix--) {
        const PERL_CONTEXT *cx = &cxstack[ix];
        if (CxTYPE(cx) == CXt_GIVEN) {
            ST(0) = sv_2mortal(newSVpvf("given %s", cx->blk_givwhen.leave_op ? "op" : "noop"));
            XSRETURN(1);
        }
    }
    ST(0) = sv_2mortal(newSVpvs("no given"));
    XSRETURN(1);
}
XS(XS_MySu_dbgv) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpv(PL_DBgv ? GvNAME(PL_DBgv) : "none", 0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MySu) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySu::alloc", XS_MySu_alloc);
    newXS_deffile("MySu::aelem", XS_MySu_aelem);
    newXS_deffile("MySu::given", XS_MySu_given);
    newXS_deffile("MySu::dbgv", XS_MySu_dbgv);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeUsesScopeUppersNames(t *testing.T) {
	so := buildModuleSource(t, "MySu", scopeUpperNamesSource)
	got := requireModule(t, "MySu", so, `use feature "switch"; no warnings;
my $g; given ("x") { $g = MySu::given() }
my @a = (1, 2, 3);
join "; ", MySu::alloc(), $g, MySu::dbgv(), MySu::aelem(\@a, 1), "@a"`)
	if want := "4 1 1; given op; none; 9 2; 1 2 3"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
