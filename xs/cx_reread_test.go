//go:build darwin || linux

package xs_test

import "testing"

// The contexts native code reads while it calls back into Perl, where
// they are read again deeper in, and after: the stackinfo's shadow is made
// again in place, and an operation asked for while it was being made wrote
// back the half-made contexts as if native code had changed them, putting
// back what was there before (DBI's caller lookup under trace read a sub's
// context as a loop, had its sub described as an op, and the interpreter
// trapped). What is expected is what perl 5.36 on Linux gives with the same
// module.
func TestNativeCodeReadsContextsAcrossACallback(t *testing.T) {
	so := buildModuleSource(t, "CxWalk", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* The contexts native code reads, from the top down to the eval the
 * program runs in (DBI's caller lookup under trace walks them so): their
 * types, the sub of a sub's and the package of the statement each was
 * entered from. through() reads them, calls back into Perl, where they are
 * read deeper in, and reads them again. */
static SV *walk_sv(pTHX) {
    SV *out = newSVpvs("");
    I32 i;
    for (i = cxstack_ix; i >= 0 && CxTYPE(&cxstack[i]) != CXt_EVAL; i--) {
        const PERL_CONTEXT *cx = &cxstack[i];
        if (i != cxstack_ix) sv_catpvs(out, ",");
        sv_catpvf(out, "%d", (int)CxTYPE(cx));
        if (CxTYPE(cx) == CXt_SUB) {
            GV *gv = CvGV(cx->blk_sub.cv);
            sv_catpvf(out, ":%s", gv ? GvNAME(gv) : "?");
        }
        sv_catpvf(out, "@%s", CopSTASHPV(cx->blk_oldcop));
    }
    return out;
}
XS(XS_CxWalk_walk) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(walk_sv(aTHX));
    XSRETURN(1);
}
XS(XS_CxWalk_through) {
    dXSARGS;
    SV *code, *before, *inner, *after;
    int n;
    if (items != 1) croak_xs_usage(cv, "code");
    code = ST(0);
    before = sv_2mortal(walk_sv(aTHX));
    PUSHMARK(SP);
    PUTBACK;
    n = call_sv(code, G_SCALAR);
    SPAGAIN;
    inner = sv_2mortal(n ? newSVsv(POPs) : newSVpvs(""));
    PUTBACK;
    after = sv_2mortal(walk_sv(aTHX));
    ST(0) = sv_2mortal(newSVpvf("%s | %s | %s", SvPV_nolen(before), SvPV_nolen(inner),
                                SvPV_nolen(after)));
    XSRETURN(1);
}
XS_EXTERNAL(boot_CxWalk) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("CxWalk::walk", XS_CxWalk_walk, __FILE__);
    newXS("CxWalk::through", XS_CxWalk_through, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `sub inner {
    for my $x (1) {
        for my $y (1) {
            return CxWalk::walk();
        }
    }
}
sub outer {
    CxWalk::through(sub { inner() });
}
sub deeper {
    for my $z (1) {
        return outer();
    }
}
join " / ", outer(), deeper()`
	if got, want := requireModule(t, "CxWalk", so, program), "9:outer@main | 7@main,7@main,9:inner@main,9:__ANON__@main,9:outer@main | 9:outer@main / 9:outer@main,7@main,9:deeper@main | 7@main,7@main,9:inner@main,9:__ANON__@main,9:outer@main,7@main,9:deeper@main | 9:outer@main,7@main,9:deeper@main"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
