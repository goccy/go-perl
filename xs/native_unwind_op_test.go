//go:build darwin || linux

package xs_test

import "testing"

// What native code assigns to PL_op is the interpreter's PL_op, as on
// perl. Scope::Upper's unwind() returns values from a sub further up so: a
// destructor of the XSUB's call puts the stack pointer back where the XSUB
// found it, marks its arguments, runs pp_return through an op of its own
// (PL_op = &return_op; PL_op = PL_op->op_ppaddr(aTHX)), and points PL_op
// at another op of its own whose op_next is the op pp_return returned;
// pp_entersub then goes on from PL_op->op_next. Before, the interpreter
// never saw the assignment and went on after the call as if nothing had
// been unwound.
const unwindOpSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static LISTOP return_op;
static OP proxy_op;
static I32 kept_cxix, kept_items;
static SV **kept_sp;
static void unwind_dtor(pTHX_ void *p) {
    I32 mark;
    PERL_UNUSED_VAR(p);
    PL_stack_sp = kept_sp;
    if (cxstack_ix > kept_cxix) dounwind(kept_cxix);
    mark = PL_markstack[cxstack[kept_cxix].blk_oldmarksp];
    PUSHMARK(PL_stack_sp - kept_items);
    PL_op = (OP *)&return_op;
    PL_op = PL_op->op_ppaddr(aTHX);
    *PL_markstack_ptr = mark;
    proxy_op.op_next = PL_op;
    PL_op = &proxy_op;
}
/* leave(@values): the sub that called leave returns @values */
XS(XS_MyUw_leave) {
    dXSARGS;
    I32 i = cxstack_ix;
    while (i >= 0 && CxTYPE(&cxstack[i]) != CXt_SUB) i--;
    if (i < 0) croak("leave outside a sub");
    kept_cxix = i;
    kept_items = items;
    kept_sp = PL_stack_sp;
    if (GIMME_V == G_SCALAR) PL_stack_sp = PL_stack_base + PL_markstack_ptr[1] + 1;
    SAVEDESTRUCTOR_X(unwind_dtor, NULL);
    return;
}
XS_EXTERNAL(boot_MyUw) {
    dXSBOOTARGSXSAPIVERCHK;
    Zero(&return_op, 1, LISTOP);
    return_op.op_type = OP_RETURN;
    return_op.op_ppaddr = PL_ppaddr[OP_RETURN];
    Zero(&proxy_op, 1, OP);
    proxy_op.op_type = OP_STUB;
    proxy_op.op_ppaddr = NULL;
    newXS_deffile("MyUw::leave", XS_MyUw_leave);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeAssignsPLop(t *testing.T) {
	so := buildModuleSource(t, "MyUw", unwindOpSource)
	got := requireModule(t, "MyUw", so, `sub inner { MyUw::leave(@_); "not reached" }
sub outer { my @r = inner(1, 2, 3); "got @r" }
my $s = sub { my $x = inner("a", "b"); "s=$x" }->();
my @l = map { inner($_ * 2) } 1 .. 3;
join "; ", outer(), $s, "@l", "after"`)
	if want := "got 1 2 3; s=b; 2 4 6; after"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
