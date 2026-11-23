//go:build darwin || linux

package xs_test

import "testing"

// Native code pushes contexts of its own and pops them, as
// Syntax::Keyword::Defer runs a defer block: cx_pushblock, cx_pusheval with
// PL_in_eval = EVAL_INEVAL|EVAL_KEEPERR, the block's code under JMPENV, and
// dounwind back; a die inside comes back to the JMPENV. An operation that
// dies inside such an eval context (EVAL_KEEPERR) warns "\t(in cleanup)
// ...", leaves $@ alone and comes back to the module's JMPENV, which caught
// it: the program goes on after the XSUB, as on perl. A context pointer
// native code holds stays that context across operations, as in perl's
// cxstack. Before, cx_pushblock and the others were not there and
// PL_in_eval could not be stored into: the module did not build.
const contextPushSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static SV *report;
/* eval_block(code): code inside an eval context the module pushes itself */
XS(XS_MyCx_eval_block) {
    dXSARGS;
    SV *code = ST(0);
    I32 was = cxstack_ix;
    int ret;
    dJMPENV;
    PERL_UNUSED_VAR(items);
    report = sv_2mortal(newSVpvs(""));
    JMPENV_PUSH(ret);
    if (ret == 0) {
        PERL_CONTEXT *cx = cx_pushblock(CXt_EVAL|CXp_EVALBLOCK, G_VOID, PL_stack_sp, PL_savestack_ix);
        cx_pusheval(cx, NULL, NULL);
        PL_in_eval = EVAL_INEVAL|EVAL_KEEPERR;
        sv_catpvf(report, "pushed %d eval %d in_eval %d", (int)(cxstack_ix - was),
                  CxTYPE(CX_CUR()) == CXt_EVAL ? 1 : 0, (int)PL_in_eval);
        PUSHMARK(SP);
        call_sv(code, G_VOID | G_DISCARD);
        sv_catpvs(report, " returned");
    } else {
        sv_catpvs(report, " caught");
    }
    JMPENV_POP;
    dounwind(was);
    sv_catpvf(report, " left %d", (int)(cxstack_ix - was));
    ST(0) = report;
    XSRETURN(1);
}
/* store_in_eval(sv): an operation that dies (a store into a read-only
 * value) inside an eval context the module pushes, under its JMPENV */
XS(XS_MyCx_store_in_eval) {
    dXSARGS;
    SV *target = ST(0);
    I32 was = cxstack_ix;
    int ret;
    dJMPENV;
    PERL_UNUSED_VAR(items);
    report = sv_2mortal(newSVpvs(""));
    JMPENV_PUSH(ret);
    if (ret == 0) {
        PERL_CONTEXT *cx = cx_pushblock(CXt_EVAL|CXp_EVALBLOCK, G_VOID, PL_stack_sp, PL_savestack_ix);
        cx_pusheval(cx, NULL, NULL);
        PL_in_eval = EVAL_INEVAL|EVAL_KEEPERR;
        sv_setiv(target, 1);
        sv_catpvs(report, "stored");
    } else {
        sv_catpvs(report, "caught");
    }
    JMPENV_POP;
    dounwind(was);
    sv_catpvf(report, " left %d", (int)(cxstack_ix - was));
    ST(0) = report;
    XSRETURN(1);
}
/* block(): a block context pushed and popped with cx_popblock and CX_POP */
XS(XS_MyCx_block) {
    dXSARGS;
    I32 was = cxstack_ix;
    PERL_CONTEXT *cx;
    SV *out;
    PERL_UNUSED_VAR(items);
    cx = cx_pushblock(CXt_BLOCK, G_VOID, PL_stack_sp, PL_savestack_ix);
    out = sv_2mortal(newSVpvf("pushed %d block %d", (int)(cxstack_ix - was),
                              CxTYPE(cx) == CXt_BLOCK ? 1 : 0));
    cx_popblock(cx);
    CX_POP(cx);
    sv_catpvf(out, " left %d", (int)(cxstack_ix - was));
    ST(0) = out;
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyCx) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyCx::eval_block", XS_MyCx_eval_block);
    newXS_deffile("MyCx::block", XS_MyCx_block);
    newXS_deffile("MyCx::store_in_eval", XS_MyCx_store_in_eval);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodePushesContexts(t *testing.T) {
	so := buildModuleSource(t, "MyCx", contextPushSource)
	got := requireModule(t, "MyCx", so, `my @r;
push @r, MyCx::block();
push @r, MyCx::eval_block(sub { 1 });
push @r, MyCx::eval_block(sub { die "boom\n" });
{
    use warnings;
    my @w;
    local $SIG{__WARN__} = sub { push @w, $_[0] };
    local $@ = "outer\n";
    push @r, MyCx::store_in_eval(1);
    push @r, "\$@ [$@] warned [" . join("", map { /^\t\(in cleanup\) Modification of a read-only value attempted/ ? "in cleanup" : $_ } @w) . "]";
}
push @r, "after " . (eval { 1 } ? "eval works" : "eval broken");
join "; ", @r`)
	// PL_in_eval: EVAL_INEVAL (1) | EVAL_KEEPERR (4)
	want := "pushed 1 block 1 left 0; pushed 1 eval 1 in_eval 5 returned left 0; " +
		"pushed 1 eval 1 in_eval 5 caught left 0; caught left 0; $@ [outer\n] warned [in cleanup]; after eval works"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
