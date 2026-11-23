//go:build darwin || linux

package xs_test

import "testing"

// CALLRUNOPS runs from an op of native code's own, on a stack native code
// pushed with PUSHSTACKi, as perl's amagic_call runs an overloaded
// operator's sub (Syntax::Infix::Smartmatch runs an overloaded ~~ so): an
// entersub op on the C stack is PL_op under SAVEOP, the run loop calls the
// sub, its value is the top of the pushed stack, and after LEAVE and
// POPSTACK the pp function goes on with its own op and stack. Before,
// CALLRUNOPS refused an op that was not the interpreter's, and PUSHSTACKi
// left native code's sp on the stack below.
const runopsOwnOpSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static OP *MyRo_pp_pcall(pTHX) {
    dSP;
    SV *const cb = get_sv("MyRo::cb", 0);
    SV *const d = get_sv("MyRo::val", 0);
    SV *sv;
    UNOP myop;
    Zero(&myop, 1, UNOP);
    myop.op_flags = OPf_STACKED | OPf_WANT_SCALAR;
    myop.op_ppaddr = PL_ppaddr[OP_ENTERSUB];
    myop.op_type = OP_ENTERSUB;
    PUSHSTACKi(PERLSI_OVERLOAD);
    ENTER;
    SAVEOP();
    PL_op = (OP *)&myop;
    PUSHMARK(SP);
    EXTEND(SP, 2);
    PUSHs(d);
    PUSHs(SvRV(cb));
    PUTBACK;
    CALLRUNOPS(aTHX);
    SPAGAIN;
    LEAVE;
    sv = POPs;
    PUTBACK;
    POPSTACK;
    SPAGAIN;
    XPUSHs(sv);
    RETURN;
}
static OP *MyRo_ck(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    OP *o;
    PERL_UNUSED_ARG(namegv);
    PERL_UNUSED_ARG(ckobj);
    op_free(entersubop);
    o = newOP(OP_CUSTOM, 0);
    o->op_ppaddr = MyRo_pp_pcall;
    return o;
}
XS(XS_MyRo_stub) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    croak("the sub was not compiled to its op");
}
XS(XS_MyRo_install) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    cv_set_call_checker(get_cv("MyRo::pcall", 0), MyRo_ck, &PL_sv_undef);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MyRo) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyRo::pcall", XS_MyRo_stub);
    newXS_deffile("MyRo::install", XS_MyRo_install);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeRunsOpsFromAnOpOfItsOwn(t *testing.T) {
	so := buildModuleSource(t, "MyRo", runopsOwnOpSource)
	got := requireModule(t, "MyRo", so, `MyRo::install();
my $c = eval q{ sub { local ($MyRo::cb, $MyRo::val) = @_; my @l = (1, MyRo::pcall(), 3); "@l" } }
    or die $@;
join "|", $c->(sub { "<$_[0]>" }, "v"), $c->(sub { scalar reverse $_[0] }, "abc")`)
	if want := "1 <v> 3|1 cba 3"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
