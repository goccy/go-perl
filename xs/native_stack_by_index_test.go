//go:build darwin || linux

package xs_test

import "testing"

// The perl stack of an XSUB is the interpreter's by index, as on perl: an
// index native code has from PL_markstack or a context reads the value
// there, below the XSUB's own arguments too. Want finds the value an
// assignment assigns to an lvalue sub that way: the mark of the sub's
// call, PL_markstack[cx->blk_oldmarksp + 1], is the index of the value
// below the sub's arguments (want('ASSIGN') read the XSUB's own argument
// instead, and `backstr(my $foo) = 'robin'` left 1 in $foo). What is
// expected is what perl 5.36 on Linux gives for the same module.
func TestNativeStackIsTheInterpretersByIndex(t *testing.T) {
	so := buildModuleSource(t, "StackIndex", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* the value assigned to the lvalue sub this XSUB is called in */
XS(XS_StackIndex_assigned) {
    dXSARGS;
    I32 i;
    PERL_UNUSED_VAR(items);
    for (i = cxstack_ix; i >= 0; i--) {
        PERL_CONTEXT *cx = &cxstack[i];
        if (CxTYPE(cx) == CXt_SUB) {
            SV *v = PL_stack_base[PL_markstack[cx->blk_oldmarksp + 1]];
            ST(0) = sv_2mortal(newSVsv(v));
            XSRETURN(1);
        }
    }
    XSRETURN_UNDEF;
}
/* its arguments where the interpreter has them: ST(n) and
 * PL_stack_base[ax + n] are one slot */
XS(XS_StackIndex_args) {
    dXSARGS;
    SV *joined = sv_2mortal(newSVpvs(""));
    I32 i;
    for (i = 0; i < items; i++) {
        if (i) sv_catpvs(joined, ",");
        sv_catsv(joined, PL_stack_base[ax + i]);
    }
    ST(0) = joined;
    XSRETURN(1);
}
XS_EXTERNAL(boot_StackIndex) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("StackIndex::assigned", XS_StackIndex_assigned);
    newXS_deffile("StackIndex::args", XS_StackIndex_args);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `
my @got;
sub target :lvalue { push @got, StackIndex::assigned(); my $x; $x }
target() = "robin";
target("an", "argument") = "nibor";
for my $n (1 .. 3) { target() = 23 }
my @list = (1, 2, 3);
join("|", @got) . " " . StackIndex::args(@list, "x")`
	if got, want := requireModule(t, "StackIndex", so, program), "robin|nibor|23|23|23 1,2,3,x"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
