//go:build darwin || linux

package xs_test

import "testing"

// A pp function registered as a custom op twice, under two names, has the
// name and the description registered last, for every op that runs it, as
// perl's custom_op_register replaces what a pp function was registered
// with. XS::Parse::Infix registers both spellings of an operator (≡ and
// =:=), and B::Deparse writes the op by its name: the SDK kept the first
// registration, and Syntax::Operator::Identical's =:= deparsed as ≡
// (t/12deparse.t). What is expected is what perl 5.36 on Linux gives with
// the same module.
func TestNativeCustomOpRegisteredAgainHasTheLastName(t *testing.T) {
	so := buildModuleSource(t, "XopTwice", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* One pp function registered as a custom op twice, under two names, as
 * XS::Parse::Infix registers both spellings of an operator (≡ and =:=):
 * the ops that run it have the name registered last. pair($a, $b) is the
 * custom op, giving "$a+$b". */
static XOP first_xop, second_xop;
static OP *pp_pair(pTHX) {
    dSP;
    SV *b = TOPs;
    SV *a = TOPm1s;
    SV *r = sv_2mortal(newSVpvf("%" SVf "+%" SVf, SVfARG(a), SVfARG(b)));
    SP--;
    SETs(r);
    RETURN;
}
static OP *ck_pair(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    OP *pushop, *aop, *bop, *cvop, *newop;
    PERL_UNUSED_ARG(namegv);
    PERL_UNUSED_ARG(ckobj);
    pushop = cUNOPx(entersubop)->op_first;
    if (!OpHAS_SIBLING(pushop))
        pushop = cUNOPx(pushop)->op_first;
    if (!(aop = OpSIBLING(pushop)) || !(bop = OpSIBLING(aop)) || !(cvop = OpSIBLING(bop)))
        return entersubop;
    OpMORESIB_set(pushop, cvop);
    OpLASTSIB_set(bop, NULL);
    OpLASTSIB_set(aop, NULL);
    newop = newBINOP(OP_NULL, 0, aop, bop);
    newop->op_type = OP_CUSTOM;
    newop->op_ppaddr = pp_pair;
    op_free(entersubop);
    return newop;
}
XS(XS_XopTwice_pair) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    croak("pair is not called as a sub");
}
XS_EXTERNAL(boot_XopTwice) {
    dXSBOOTARGSXSAPIVERCHK;
    CV *code;
    PERL_UNUSED_VAR(items);
    XopENTRY_set(&first_xop, xop_name, "pair_first");
    XopENTRY_set(&first_xop, xop_desc, "pair, first name");
    XopENTRY_set(&first_xop, xop_class, OA_BINOP);
    Perl_custom_op_register(aTHX_ pp_pair, &first_xop);
    XopENTRY_set(&second_xop, xop_name, "pair_second");
    XopENTRY_set(&second_xop, xop_desc, "pair, second name");
    XopENTRY_set(&second_xop, xop_class, OA_BINOP);
    Perl_custom_op_register(aTHX_ pp_pair, &second_xop);
    code = newXS("XopTwice::pair", XS_XopTwice_pair, __FILE__);
    cv_set_call_checker(code, ck_pair, (SV *)code);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `use B ();
my $code = eval q{ sub { XopTwice::pair($_[0], $_[1]) } };
sub names {
    my $o = shift;
    my @n;
    for (; $o && !$o->isa("B::NULL"); $o = $o->sibling) {
        push @n, $o->name . ":" . $o->desc if $o->name =~ /^pair/;
        push @n, names($o->first) if $o->flags & B::OPf_KIDS;
    }
    @n
}
join ",", $code->("a", "b"), names(B::svref_2object($code)->ROOT)`
	if got, want := requireModule(t, "XopTwice", so, program), "a+b,pair_second:pair, second name"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
