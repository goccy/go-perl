//go:build darwin || linux

package xs_test

import "testing"

// A run loop of native code (PL_runops, as Test::LeakTrace installs at its
// boot) runs custom ops whose pp functions are native code's, chained so
// that one's result is the other's operand, as JSON::YY's
// decode_json(encode_json(...)) keywords are: each pp function says by
// NORMAL which op runs next, and the loop runs that op by its shadow. What
// is expected is what perl 5.36 on Linux gives with the same module.
// Before, the next op was a shadow described when the checker made the op,
// before the interpreter linked it: the loop went round forever, or the
// code yielded nothing.
func TestNativeRunLoopRunsChainedCustomOps(t *testing.T) {
	so := buildModuleSource(t, "RunopsChain", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int my_runops(pTHX) {
    while (PL_op) {
        PL_op = PL_op->op_ppaddr(aTHX);
    }
    TAINT_NOT;
    return 0;
}
static XOP a_xop, b_xop;
/* wrapa($x) = "a($x)", wrapb($x) = "b($x)" */
static OP *pp_wrapa(pTHX) { dSP; SV *x = TOPs; SETs(sv_2mortal(newSVpvf("a(%" SVf ")", SVfARG(x)))); RETURN; }
static OP *pp_wrapb(pTHX) { dSP; SV *x = TOPs; SETs(sv_2mortal(newSVpvf("b(%" SVf ")", SVfARG(x)))); RETURN; }
static OP *ck_wrap(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    OP *pushop, *aop, *cvop, *newop;
    PERL_UNUSED_ARG(namegv);
    pushop = cUNOPx(entersubop)->op_first;
    if (!OpHAS_SIBLING(pushop)) pushop = cUNOPx(pushop)->op_first;
    if (!(aop = OpSIBLING(pushop)) || !(cvop = OpSIBLING(aop))) return entersubop;
    OpMORESIB_set(pushop, cvop);
    OpLASTSIB_set(aop, NULL);
    newop = newUNOP(OP_NULL, 0, aop);
    newop->op_type = OP_CUSTOM;
    newop->op_ppaddr = SvIV(ckobj) ? pp_wrapb : pp_wrapa;
    op_free(entersubop);
    return newop;
}
XS(XS_RunopsChain_install) { dXSARGS; PERL_UNUSED_VAR(items); PL_runops = my_runops; XSRETURN_EMPTY; }
XS(XS_RunopsChain_wrap) { dXSARGS; PERL_UNUSED_VAR(items); croak("not as a sub"); }
XS_EXTERNAL(boot_RunopsChain) {
    dXSBOOTARGSXSAPIVERCHK;
    CV *code;
    PERL_UNUSED_VAR(items);
    XopENTRY_set(&a_xop, xop_name, "wrapa");
    Perl_custom_op_register(aTHX_ pp_wrapa, &a_xop);
    XopENTRY_set(&b_xop, xop_name, "wrapb");
    Perl_custom_op_register(aTHX_ pp_wrapb, &b_xop);
    code = newXS("RunopsChain::wrapa", XS_RunopsChain_wrap, __FILE__);
    cv_set_call_checker(code, ck_wrap, newSViv(0));
    code = newXS("RunopsChain::wrapb", XS_RunopsChain_wrap, __FILE__);
    cv_set_call_checker(code, ck_wrap, newSViv(1));
    newXS("RunopsChain::install", XS_RunopsChain_install, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `RunopsChain::install();
# the interpreter goes on after this eval with its next call of the loop
eval { die "start\n" };
eval q{
    my @r;
    push @r, RunopsChain::wrapa("x");
    my $v = RunopsChain::wrapb(RunopsChain::wrapa("y"));
    push @r, $v, RunopsChain::wrapa(RunopsChain::wrapb(RunopsChain::wrapa("z")));
    push @r, "after";
    join ",", @r;
} // "died: $@"`
	const want = "a(x),b(a(y)),a(b(a(z))),after"
	if got := requireModule(t, "RunopsChain", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
