//go:build darwin || linux

package xs_test

import "testing"

// A sub's leavesub of native code's own, as Data::Alias has its own
// pp_return: the value goes to the caller's stack, the scopes the sub
// opened are left (CX_LEAVE_SCOPE: a local in it is put back), and the
// sub's context is popped (cx_popsub, cx_popblock, CX_POP), and the op
// after the call is where the program goes on. Before, CX_LEAVE_SCOPE and
// cx_popsub were not in the SDK.
const leavesubSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int calls;
/* leavesub in scalar and void context */
static OP *my_leavesub(pTHX) {
    PERL_CONTEXT *const cx = CX_CUR();
    const U8 gimme = cx->blk_gimme;
    SV **const oldsp = PL_stack_base + cx->blk_oldsp;
    OP *retop;
    if (CxTYPE(cx) != CXt_SUB) croak("not a sub's context");
    if (gimme == G_SCALAR) {
        SV *const ret = PL_stack_sp > oldsp ? *PL_stack_sp : &PL_sv_undef;
        oldsp[1] = sv_mortalcopy(ret);
        PL_stack_sp = oldsp + 1;
    } else {
        PL_stack_sp = oldsp;
    }
    CX_LEAVE_SCOPE(cx);
    cx_popsub(cx);
    cx_popblock(cx);
    retop = cx->blk_sub.retop;
    CX_POP(cx);
    calls++;
    return retop;
}
/* install(\\&sub): the sub's leavesub op runs my_leavesub */
XS(XS_MyLs_install) {
    dXSARGS;
    CV *const sub = (CV *)SvRV(ST(0));
    OP *const root = CvROOT(sub);
    PERL_UNUSED_VAR(items);
    if (!root || root->op_type != OP_LEAVESUB) croak("no leavesub");
    root->op_ppaddr = my_leavesub;
    XSRETURN_EMPTY;
}
XS(XS_MyLs_calls) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(calls));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyLs) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyLs::install", XS_MyLs_install);
    newXS_deffile("MyLs::calls", XS_MyLs_calls);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeLeavesASub(t *testing.T) {
	so := buildModuleSource(t, "MyLs", leavesubSource)
	got := requireModule(t, "MyLs", so, `our $Y = 1;
sub t { my $x = shift; local $Y = 5; $x * 2 + $Y }
MyLs::install(\&t);
my @r = (scalar(t(20)), $Y, scalar(t(1)));
t(3);
push @r, $Y, MyLs::calls();
join "|", @r`)
	if want := "45|1|7|1|3"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
