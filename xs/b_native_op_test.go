//go:build darwin || linux

package xs_test

import "testing"

// A B::OP object that native code makes of an op of its own: B::CallChecker
// wraps a struct on its C stack whose op_next is the op it wants the object
// of and asks B for the next op; B::OPCheck sets op_first of a static UNOP
// and calls the XSUB of B::UNOP::first. B is handed an address of the host
// process and gets an op that stands for it, whose ops are the interpreter's
// (B-CallChecker, Err, B-OPCheck, Hash-Lazy and self died with "go-perl
// gives native code no access to this interpreter internal: a B::OP object
// of an op of native code"). What is expected is what perl 5.36 on Linux
// gives with the same module.
func TestNativeCodeMakesBObjectsOfItsOwnOps(t *testing.T) {
	so := buildModuleSource(t, "BNative", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* B::OP objects of the op that calls an XSUB, made the two ways modules
 * make them without B's own make_op_object: an op of the module's own whose
 * op_next is the op, wrapped in a B::OP object whose next op B is asked for
 * (B::CallChecker), and a static UNOP whose op_first is the op, whose first
 * op B::UNOP::first is asked for by calling its XSUB (B::OPCheck). */
static UNOP fake_unop;
static SV *bop_by_next(pTHX_ OP *op) {
    OP stalk;
    SV *stalkbop, *bop;
    dSP;
    /* the rest of the op holds what the stack held, as B::CallChecker's
     * does: here a byte pattern of addresses nothing is mapped at */
    memset(&stalk, 0x08, sizeof stalk);
    stalk.op_next = op;
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
static SV *bop_by_first(pTHX_ OP *op) {
    CV *first = get_cv("B::UNOP::first", 0);
    SV *fake = sv_2mortal(sv_bless(newRV_noinc(newSVuv((UV)&fake_unop)), gv_stashpv("B::UNOP", 0)));
    SV *bop;
    dSP;
    fake_unop.op_first = op;
    PUSHMARK(SP);
    XPUSHs(fake);
    PUTBACK;
    (void)CvXSUB(first)(aTHX_ first);
    SPAGAIN;
    bop = POPs;
    PUTBACK;
    return SvREFCNT_inc(bop);
}
/* ops(): the B::OP objects of the op calling this, both ways */
XS(XS_BNative_ops) {
    dXSARGS;
    SV *a, *b;
    PERL_UNUSED_VAR(items);
    a = bop_by_next(aTHX_ PL_op);
    b = bop_by_first(aTHX_ PL_op);
    SP -= items;
    XPUSHs(sv_2mortal(a));
    XPUSHs(sv_2mortal(b));
    PUTBACK;
}
XS_EXTERNAL(boot_BNative) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    load_module(PERL_LOADMOD_NOIMPORT, newSVpvs("B"), NULL);
    newXS("BNative::ops", XS_BNative_ops, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my ($a, $b) = BNative::ops(1, 2);
my $arg = $a->first;
$arg = $arg->first if $arg->sibling->isa("B::NULL");
my @names;
for (my $o = $arg->sibling; !$o->isa("B::NULL"); $o = $o->sibling) { push @names, $o->name }
join ",", ref($a), $a->name, ref($b), $b->name, ($$a == $$b ? "same" : "differ"), "@names"`
	if got, want := requireModule(t, "BNative", so, program), "B::UNOP,entersub,B::UNOP,entersub,same,const const null"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
