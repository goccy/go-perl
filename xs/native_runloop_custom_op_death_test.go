//go:build darwin || linux

package xs_test

import "testing"

// A run loop of native code (PL_runops, as Test::LeakTrace installs at its
// boot) runs a custom op whose pp function dies inside an eval, as JSON::YY's
// decode_json keyword does on input nested too deep: the eval catches the
// death, and the lexicals around it, a sub's arguments and the loop's later
// ops are as with perl's own loop. Before, the death left the scopes the
// loop's code had opened since the loop was entered (the eval's own, a
// lexical's clearing, the string eval's) before perl unwound to the eval, so
// the lexicals read undef and the next op the loop ran was a freed one. What
// is expected is what perl 5.36 on Linux gives with the same module.
func TestNativeRunLoopThroughACustomOpDeath(t *testing.T) {
	so := buildModuleSource(t, "RunopsCroakOp", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* A run loop of native code, as Test::LeakTrace installs at its boot, and
 * a custom op whose pp function dies, as JSON::YY's decode_json keyword
 * does on input nested too deep. */
static int my_runops(pTHX) {
    while ((PL_op = PL_op->op_ppaddr(aTHX))) {
        PERL_ASYNC_CHECK();
    }
    TAINT_NOT;
    return 0;
}
static XOP croaker_xop;
/* croaker($n) as a custom op: dies with $n when $n is true, else yes */
static OP *pp_croaker(pTHX) {
    dSP;
    SV *n = TOPs;
    if (SvTRUE(n)) croak("croaked with %d", (int)SvIV(n));
    SETs(&PL_sv_yes);
    RETURN;
}
static OP *ck_croaker(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    OP *pushop, *aop, *cvop, *newop;
    PERL_UNUSED_ARG(namegv);
    PERL_UNUSED_ARG(ckobj);
    pushop = cUNOPx(entersubop)->op_first;
    if (!OpHAS_SIBLING(pushop))
        pushop = cUNOPx(pushop)->op_first;
    if (!(aop = OpSIBLING(pushop)) || !(cvop = OpSIBLING(aop)))
        return entersubop;
    OpMORESIB_set(pushop, cvop);
    OpLASTSIB_set(aop, NULL);
    newop = newUNOP(OP_NULL, 0, aop);
    newop->op_type = OP_CUSTOM;
    newop->op_ppaddr = pp_croaker;
    op_free(entersubop);
    return newop;
}
XS(XS_RunopsCroakOp_install) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    PL_runops = my_runops;
    XSRETURN_EMPTY;
}
XS(XS_RunopsCroakOp_croaker) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    croak("croaker is not called as a sub");
}
XS_EXTERNAL(boot_RunopsCroakOp) {
    dXSBOOTARGSXSAPIVERCHK;
    CV *code;
    PERL_UNUSED_VAR(items);
    XopENTRY_set(&croaker_xop, xop_name, "croaker");
    XopENTRY_set(&croaker_xop, xop_desc, "die with a number");
    Perl_custom_op_register(aTHX_ pp_croaker, &croaker_xop);
    code = newXS("RunopsCroakOp::croaker", XS_RunopsCroakOp_croaker, __FILE__);
    cv_set_call_checker(code, ck_croaker, (SV *)code);
    newXS("RunopsCroakOp::install", XS_RunopsCroakOp_install, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `RunopsCroakOp::install();
# the interpreter goes on after this eval with its next call of the loop
eval { die "start\n" };
eval q{
    sub plus { my $n = shift; $n + 1 }
    my @r;
    my $LIMIT = 512;
    my $ok = eval { RunopsCroakOp::croaker(plus($LIMIT)); 1 };
    push @r, ($ok ? "lived" : $@ =~ /^croaked with 513 / ? "died" : "[$@]"), defined $LIMIT ? $LIMIT : "undef";
    $ok = eval { RunopsCroakOp::croaker(0) };
    push @r, defined $ok ? $ok : "[$@]", defined $LIMIT ? $LIMIT : "undef";
    sub body {
        my $n = shift;
        my $ok = eval { my $inner = "kept"; RunopsCroakOp::croaker($n); 1 };
        push @r, ($ok ? "lived" : $@ =~ /^croaked with $n / ? "died" : "[$@]"), defined $n ? $n : "undef";
    }
    body($_) for 7, 8;
    my $warned = 0;
    local $SIG{__WARN__} = sub { $warned++ };
    for my $i (1 .. 3) { eval { RunopsCroakOp::croaker($i) } }
    push @r, "warnings $warned", defined $LIMIT ? $LIMIT : "undef";
    join ",", @r;
} // "died: $@"`
	const want = "died,512,1,512,died,7,died,8,warnings 0,512"
	if got := requireModule(t, "RunopsCroakOp", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
