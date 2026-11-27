//go:build darwin || linux

package xs_test

import "testing"

// An op native code saw while it was being built (a module's check
// function: autobox rewrites method calls in one) has its links as the
// compiler left them, also in a later call of native code that reaches it
// across OpSIBLING: Want walks the statement of its caller for the entersub
// whose op_next is the context's retop, and found an op_next of NULL (the
// one it had when the check function saw it), so want() died with "Want
// panicked: null list in lastop" (autobox-Core's s). What is expected is
// what perl 5.36 on Linux gives for the same module.
func TestNativeOpSeenWhileBuiltHasItsLinksLater(t *testing.T) {
	so := buildModuleSource(t, "OpSeen", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static Perl_check_t prev_entersub;
static OP *seen_entersub(pTHX_ OP *o) {
    /* looks at the op while it is built, as a module's check does */
    (void)o->op_next;
    return prev_entersub(aTHX_ o);
}
static int walk(pTHX_ OP *start, OP *next) {
    OP *o;
    for (o = start; o; o = OpSIBLING(o)) {
        if (o->op_type == OP_ENTERSUB && o->op_next == next) return 1;
        if ((o->op_flags & OPf_KIDS) && walk(aTHX_ cUNOPo->op_first, next)) return 1;
    }
    return 0;
}
/* whether the statement of the caller of the sub this XSUB is called in
 * has the call of that sub, by the op_next of its entersub */
XS(XS_OpSeen_found) {
    dXSARGS;
    I32 i;
    PERL_UNUSED_VAR(items);
    for (i = cxstack_ix; i >= 0; i--) {
        PERL_CONTEXT *cx = &cxstack[i];
        if (CxTYPE(cx) == CXt_SUB) {
            ST(0) = walk(aTHX_ (OP *)cx->blk_oldcop, cx->blk_sub.retop) ? &PL_sv_yes : &PL_sv_no;
            XSRETURN(1);
        }
    }
    XSRETURN_UNDEF;
}
XS_EXTERNAL(boot_OpSeen) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("OpSeen::found", XS_OpSeen_found);
    wrap_op_checker(OP_ENTERSUB, seen_entersub, &prev_entersub);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `
my @found;
eval q{
    sub target { push @found, OpSeen::found() ? "found" : "not found"; 1 }
    target();
    my $r = target();
    target() for 1 .. 2;
    1;
} or die $@;
join ",", @found`
	if got, want := requireModule(t, "OpSeen", so, program), "found,found,found,found"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
