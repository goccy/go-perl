//go:build darwin || linux

package xs_test

import "testing"

// Native code pushes a try context of perl's feature 'try' and pops it, as
// Future::AsyncAwait rebuilds one when it resumes an async sub that awaited
// inside try: cx_pushblock with CXp_TRY, then cx_pushtry with the op to go
// on at. The context is a try (CxTRY) and an eval (CXt_EVAL) with that
// op, and its fields of an eval are what cx_pushtry set in between; before
// cx_pushtry they are not what an earlier context there left. Before,
// cx_pushtry was not there, and the shadow of the context stack read the
// poison of a new context's fields as ops.
const tryContextSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MyTc_try) {
    dXSARGS;
    const I32 was = cxstack_ix;
    PERL_CONTEXT *cx;
    SV *const report = sv_2mortal(newSVpvs(""));
    PERL_UNUSED_VAR(items);
    cx = cx_pushblock(CXt_EVAL | CXp_EVALBLOCK | CXp_TRY, G_VOID, PL_stack_sp, PL_savestack_ix);
    cx_pushtry(cx, PL_op->op_next);
    cx = CX_CUR();
    sv_catpvf(report, "pushed %d try %d eval %d retop %s", (int)(cxstack_ix - was),
              CxTRY(cx) ? 1 : 0, CxTYPE(cx) == CXt_EVAL ? 1 : 0,
              cx->blk_eval.retop == PL_op->op_next ? "next" : "other");
    cx_popeval(cx);
    cx_popblock(cx);
    CX_POP(cx);
    sv_catpvf(report, " left %d", (int)(cxstack_ix - was));
    ST(0) = report;
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyTc) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyTc::try", XS_MyTc_try);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodePushesATryContext(t *testing.T) {
	so := buildModuleSource(t, "MyTc", tryContextSource)
	got := requireModule(t, "MyTc", so, `my $r = MyTc::try(); $r`)
	if want := "pushed 1 try 1 eval 1 retop next left 0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
