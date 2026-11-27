//go:build darwin || linux

package xs_test

import "testing"

// A run loop of native code that runs a custom op of a module calls the
// op's pp function itself, which works on the interpreter's stack as perl's
// pp functions do. LRU::Cache turns lru_set($cache, $key, $value) into a
// custom op whose pp function takes its three operands off the stack;
// under Test::LeakTrace's run loop it read what the run loop's frame held
// instead ("native code handed the interpreter 0x... as a scalar").
const runloopCustomOpSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int my_runops(pTHX) {
    while ((PL_op = PL_op->op_ppaddr(aTHX))) {
        PERL_ASYNC_CHECK();
    }
    TAINT_NOT;
    return 0;
}
static XOP join3_xop;
/* join3($a, $b, $c) as the custom op: "$a-$b-$c" in place of the three */
static OP *pp_join3(pTHX) {
    dSP;
    SV *c = TOPs;
    SV *b = TOPm1s;
    SV *a = *(SP - 2);
    SV *r = sv_2mortal(newSVpvf("%" SVf "-%" SVf "-%" SVf, SVfARG(a), SVfARG(b), SVfARG(c)));
    *(SP - 2) = r;
    SP -= 2;
    RETURN;
}
static OP *ck_join3(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    OP *pushop, *aop, *bop, *cop, *cvop, *inner, *newop;
    PERL_UNUSED_ARG(namegv);
    PERL_UNUSED_ARG(ckobj);
    pushop = cUNOPx(entersubop)->op_first;
    if (!OpHAS_SIBLING(pushop))
        pushop = cUNOPx(pushop)->op_first;
    if (!(aop = OpSIBLING(pushop)) || !(bop = OpSIBLING(aop)) ||
        !(cop = OpSIBLING(bop)) || !(cvop = OpSIBLING(cop)))
        return entersubop;
    OpMORESIB_set(pushop, cvop);
    OpLASTSIB_set(aop, NULL);
    OpLASTSIB_set(bop, NULL);
    OpLASTSIB_set(cop, NULL);
    inner = newBINOP(OP_NULL, 0, aop, bop);
    newop = newBINOP(OP_NULL, 0, inner, cop);
    newop->op_type = OP_CUSTOM;
    newop->op_ppaddr = pp_join3;
    op_free(entersubop);
    return newop;
}
static XOP join2_xop;
/* join2($a, $b) as the custom op: "$a+$b" in place of the two */
static OP *pp_join2(pTHX) {
    dSP;
    SV *b = TOPs;
    SV *a = TOPm1s;
    SV *r = sv_2mortal(newSVpvf("%" SVf "+%" SVf, SVfARG(a), SVfARG(b)));
    SP--;
    SETs(r);
    RETURN;
}
static OP *ck_join2(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    OP *pushop, *aop, *bop, *cvop, *newop;
    PERL_UNUSED_ARG(namegv);
    pushop = cUNOPx(entersubop)->op_first;
    if (!OpHAS_SIBLING(pushop))
        pushop = cUNOPx(pushop)->op_first;
    if (!(aop = OpSIBLING(pushop)) || !(bop = OpSIBLING(aop)) || !(cvop = OpSIBLING(bop)))
        return entersubop;
    OpMORESIB_set(pushop, cvop);
    OpLASTSIB_set(bop, NULL);
    OpLASTSIB_set(aop, NULL);
    /* made in the context it is called in already: nothing of the op the
     * interpreter compares changes after the checker (its links do) */
    newop = newBINOP(OP_NULL, ckobj == &PL_sv_yes ? OPf_WANT_SCALAR : 0, aop, bop);
    newop->op_type = OP_CUSTOM;
    newop->op_ppaddr = pp_join2;
    op_free(entersubop);
    return newop;
}
XS(XS_MyCo_join3) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    croak("join3 is not called as a sub");
}
XS_EXTERNAL(boot_MyCo) {
    dXSBOOTARGSXSAPIVERCHK;
    CV *code;
    XopENTRY_set(&join3_xop, xop_name, "join3");
    XopENTRY_set(&join3_xop, xop_desc, "join three values");
    Perl_custom_op_register(aTHX_ pp_join3, &join3_xop);
    code = newXS_deffile("MyCo::join3", XS_MyCo_join3);
    cv_set_call_checker(code, ck_join3, (SV *)code);
    XopENTRY_set(&join2_xop, xop_name, "join2");
    XopENTRY_set(&join2_xop, xop_desc, "join two values");
    Perl_custom_op_register(aTHX_ pp_join2, &join2_xop);
    code = newXS_deffile("MyCo::join2", XS_MyCo_join3);
    cv_set_call_checker(code, ck_join2, (SV *)code);
    code = newXS_deffile("MyCo::join2s", XS_MyCo_join3);
    cv_set_call_checker(code, ck_join2, &PL_sv_yes);
    PL_runops = my_runops;
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeRunLoopRunsACustomOp(t *testing.T) {
	so := buildModuleSource(t, "MyCo", runloopCustomOpSource)
	got := requireModule(t, "MyCo", so, `# the interpreter goes on after this eval with its next call of the loop
eval { die "start\n" };
eval q{
    my @r;
    my ($x, $y) = ("a", "b");
    push @r, MyCo::join3($x, $y, "c");
    for my $i (1 .. 3) { push @r, MyCo::join3("k", $i, $x . $i) }
    push @r, "<", MyCo::join3("p", "q", "r"), ">";
    push @r, join(",", "s", MyCo::join3("t", "u", "v"), "w");
    open my $fh, '>', \my $buf or die;
    print $fh "x:", MyCo::join3("y", "z", "0"), ".";
    close $fh;
    push @r, $buf;
    my $k = "m";
    my @l = ("y", MyCo::join2($k, "n"), "x");
    push @r, "@l";
    my $s = MyCo::join2s($k, "o");
    push @r, $s;
    "@r";
} // "died: $@"`)
	if want := "a-b-c k-1-a1 k-2-a2 k-3-a3 < p-q-r > s,t-u-v,w x:y-z-0. y m+n x m+o"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
