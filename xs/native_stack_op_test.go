//go:build darwin || linux

package xs_test

import "testing"

// A pp function that runs one of perl's own on an op of its making on the C
// stack, with only the fields perl reads set (Syntax::Operator::In's
// comparison op: `OP cmpop; cmpop.op_type = OP_EQ; cmpop.op_flags = 0;
// cmpop.op_ppaddr = PL_ppaddr[OP_EQ]; PL_op = &cmpop;`), gets perl's result.
// The rest of the struct, the SDK's own fields among them, is whatever was
// on the stack: before, a token there was taken for the interpreter's op
// ("native code handed the interpreter 0x... as a OP in the operation
// RUN_ORIGINAL").
func TestNativeOpOnTheCStack(t *testing.T) {
	so := buildModuleSource(t, "StackOp", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <string.h>
static int eq_on_stack(pTHX_ SV *a, SV *b) {
    dSP;
    OP cmpop;
    SV *ret;
    int r;
    /* what was on the stack before: not zero, as it seldom is */
    memset(&cmpop, 0xA5, sizeof cmpop);
    cmpop.op_type = OP_SEQ;
    cmpop.op_flags = 0;
    cmpop.op_ppaddr = PL_ppaddr[OP_SEQ];
    ENTER;
    SAVEVPTR(PL_op);
    PL_op = &cmpop;
    EXTEND(SP, 2);
    PUSHs(a);
    PUSHs(b);
    PUTBACK;
    (*cmpop.op_ppaddr)(aTHX);
    SPAGAIN;
    ret = POPs;
    r = SvTRUE(ret);
    PUTBACK;
    LEAVE;
    return r;
}
XS(XS_StackOp_seq) {
    dXSARGS;
    int r;
    PERL_UNUSED_VAR(items);
    r = eq_on_stack(aTHX_ ST(0), ST(1));
    ST(0) = r ? &PL_sv_yes : &PL_sv_no;
    XSRETURN(1);
}
XS_EXTERNAL(boot_StackOp) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("StackOp::seq", XS_StackOp_seq);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	got := requireModule(t, "StackOp", so, `join ",", map { StackOp::seq(@$_) ? "eq" : "ne" } ["a", "a"], ["a", "b"], [10, "10"]`)
	if want := "eq,ne,eq"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
