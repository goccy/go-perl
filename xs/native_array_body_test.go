//go:build darwin || linux

package xs_test

import "testing"

// AvARRAY is the array's body where the interpreter has it, as on perl: a
// body taken before a store into the array sees the store, and a store
// through the body is the array's at once. A foreach over a lexical keeps
// the address of the variable's pad slot in its context, and that address
// is &PAD_SVl of the loop's target, as Future::AsyncAwait checks before it
// suspends a loop. Before, AvARRAY was a copy written back at the next
// operation: an earlier body did not see a store, and the slot addresses
// of PL_curpad were the copy's, not the pad's.
const arrayBodySource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* live(): a body taken before stores into the array, and a store through it */
XS(XS_MyAb_live) {
    dXSARGS;
    AV *const av = (AV *)sv_2mortal((SV *)newAV());
    SV **body;
    SV **got;
    PERL_UNUSED_VAR(items);
    av_extend(av, 3);
    av_store(av, 0, newSViv(1));
    body = AvARRAY(av);
    av_store(av, 1, newSViv(2));
    body[2] = newSViv(3);
    AvFILLp(av) = 2;
    got = av_fetch(av, 2, 0);
    ST(0) = sv_2mortal(newSVpvf("%d %d %d", (int)SvIV(body[0]), body[1] ? (int)SvIV(body[1]) : -1,
                                got ? (int)SvIV(*got) : -1));
    XSRETURN(1);
}
/* loopvar(): the innermost loop's iteration variable, and whether its
 * address is the pad slot of the loop's target */
XS(XS_MyAb_loopvar) {
    dXSARGS;
    I32 ix;
    PERL_UNUSED_VAR(items);
    for (ix = cxstack_ix; ix >= 0; ix--) {
        const PERL_CONTEXT *cx = &cxstack[ix];
        if (CxTYPE_is_LOOP(cx) && CxPADLOOP(cx)) {
            SV **const var = CxITERVAR(cx);
            ST(0) = sv_2mortal(newSVpvf(
                "%s %d", var == &PAD_SVl(cx->blk_loop.my_op->op_targ) ? "slot" : "elsewhere",
                (int)SvIV(*var)));
            XSRETURN(1);
        }
    }
    ST(0) = sv_2mortal(newSVpvs("no loop"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyAb) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyAb::live", XS_MyAb_live);
    newXS_deffile("MyAb::loopvar", XS_MyAb_loopvar);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeWorksOnTheBodyOfAnArray(t *testing.T) {
	so := buildModuleSource(t, "MyAb", arrayBodySource)
	got := requireModule(t, "MyAb", so, `my @l = (MyAb::live());
for my $i (4, 5) { push @l, MyAb::loopvar() }
sub f { my @r; for my $j (7 .. 8) { push @r, MyAb::loopvar() } @r }
join "; ", @l, f()`)
	if want := "1 2 3; slot 4; slot 5; slot 7; slot 8"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
