//go:build darwin || linux

package xs_test

import "testing"

// The links of a module's op on its C stack may hold stack bytes, here the
// address of memory mapped without access (as the interpreter's reserved
// range past its memory is): B follows a link of a module's op only when
// it is asked for it, as perl does, so the program goes on. Before, the
// links of the op were followed when B was handed the op, and reading what
// that one pointed at died of SEGV_ACCERR (Err t/01basic.t). B's oplist of
// a sort op of the module's steps through its block's ops the same way,
// past a link it does not take. What is expected is what perl 5.36 on
// Linux gives with the same module.
func TestNativeOpLinksIntoMemoryWithoutAccess(t *testing.T) {
	so := buildModuleSource(t, "BNativeNoAccess", `#include <sys/mman.h>
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* the B::OP object of the op that calls an XSUB, made as B::CallChecker
 * makes it: an op of the module's own whose op_next is the op, wrapped in
 * a B::OP object whose next op B is asked for; the op's other link is an
 * address of a page mapped without access */
static void *noaccess;
static SV *bop_by_next(pTHX_ OP *op) {
    OP stalk;
    SV *stalkbop, *bop;
    dSP;
    memset(&stalk, 0, sizeof stalk);
    stalk.op_next = op;
    stalk.op_sibparent = (OP *)noaccess;
    stalkbop = sv_2mortal(newRV_noinc(newSViv(INT2PTR(IV, &stalk))));
    sv_bless(stalkbop, gv_stashpv("B::OP", GV_ADD));
    PUSHMARK(SP);
    XPUSHs(stalkbop);
    PUTBACK;
    call_method("next", G_SCALAR);
    SPAGAIN;
    bop = POPs;
    PUTBACK;
    return SvREFCNT_inc(bop);
}
/* a sort op of the module's own, with the ops B's oplist steps through to
 * the op after its block (pushmark's sibling, rv2gv's first, leave's
 * first, enter's next), and an op after it; enter's other link is an
 * address of the page without access */
static LISTOP sort_op;
static OP pushmark_op, enter_op, after_op;
static UNOP rv2gv_op, leave_op;
static void make_sort_ops(void) {
    sort_op.op_type = OP_SORT;
    sort_op.op_flags = OPf_STACKED | OPf_SPECIAL | OPf_KIDS;
    sort_op.op_first = &pushmark_op;
    sort_op.op_opt = 1;
    pushmark_op.op_type = OP_PUSHMARK;
    OpMORESIB_set(&pushmark_op, (OP *)&rv2gv_op);
    rv2gv_op.op_type = OP_RV2GV;
    rv2gv_op.op_flags = OPf_KIDS;
    rv2gv_op.op_first = (OP *)&leave_op;
    leave_op.op_type = OP_LEAVE;
    leave_op.op_flags = OPf_KIDS;
    leave_op.op_first = &enter_op;
    enter_op.op_type = OP_ENTER;
    enter_op.op_next = &after_op;
    enter_op.op_sibparent = (OP *)noaccess;
    after_op.op_type = OP_PUSHMARK;
    after_op.op_opt = 1;
}
/* sort_op(): the B::OP object of the module's sort op */
XS(XS_BNativeNoAccess_sort_op) {
    dXSARGS;
    SV *a;
    PERL_UNUSED_VAR(items);
    a = bop_by_next(aTHX_ (OP *)&sort_op);
    SP -= items;
    XPUSHs(sv_2mortal(a));
    PUTBACK;
}
/* op(): the B::OP object of the op calling this */
XS(XS_BNativeNoAccess_op) {
    dXSARGS;
    SV *a;
    PERL_UNUSED_VAR(items);
    a = bop_by_next(aTHX_ PL_op);
    SP -= items;
    XPUSHs(sv_2mortal(a));
    PUTBACK;
}
XS_EXTERNAL(boot_BNativeNoAccess) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    noaccess = mmap(NULL, 65536, PROT_NONE, MAP_PRIVATE | MAP_ANON, -1, 0);
    if (noaccess == MAP_FAILED) croak("mmap failed");
    make_sort_ops();
    load_module(PERL_LOADMOD_NOIMPORT, newSVpvs("B"), NULL);
    newXS("BNativeNoAccess::op", XS_BNativeNoAccess_op, __FILE__);
    newXS("BNativeNoAccess::sort_op", XS_BNativeNoAccess_sort_op, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my ($a) = BNativeNoAccess::op(1, 2);
my ($s) = BNativeNoAccess::sort_op();
join ",", ref($a), $a->name, map { ref($_) . ":" . $_->name } $s->oplist`
	if got, want := requireModule(t, "BNativeNoAccess", so, program), "B::UNOP,entersub,B::LISTOP:sort,B::OP:pushmark"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
