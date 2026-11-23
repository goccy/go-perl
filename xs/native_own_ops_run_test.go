//go:build darwin || linux

package xs_test

import "testing"

// A match op of the module's own run from the pp function of a custom op,
// as Syntax::Infix::Smartmatch matches a value against a qr//: the op made
// by newPMOP and given its regexp with PM_SETRE, PL_op pointed at it under
// SAVEOP and its pp function called. What the module wrote into the op is
// the op's when the pp function reads it, and PL_op is the custom op again
// after LEAVE. Before, the match ran with no regexp, and the custom op's pp
// function went on from the match op, which ended the program. The expected
// string is what perl prints with the same module code.
const ownOpsRunSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

/* the pp function a call of MyOwn::pmatch() is compiled to: a match of
 * $MyOwn::re (a qr//) against $MyOwn::val by a match op of the module's
 * own */
static OP *MyOwn_pp_pmatch(pTHX) {
    dSP;
    SV *const re = get_sv("MyOwn::re", 0);
    SV *const d = get_sv("MyOwn::val", 0);
    PMOP *const matcher = cPMOPx(newPMOP(OP_MATCH, OPf_WANT_SCALAR | OPf_STACKED));
    bool result;
    PM_SETRE(matcher, ReREFCNT_inc((REGEXP *)SvRV(re)));
    ENTER;
    SAVEFREEOP((OP *)matcher);
    SAVEOP();
    PL_op = (OP *)matcher;
    XPUSHs(d);
    PUTBACK;
    (void)PL_ppaddr[OP_MATCH](aTHX);
    SPAGAIN;
    result = SvTRUEx(POPs);
    PUTBACK;
    LEAVE;
    XPUSHs(result ? &PL_sv_yes : &PL_sv_no);
    RETURN;
}

static OP *MyOwn_ck(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    OP *o;
    PERL_UNUSED_ARG(namegv);
    PERL_UNUSED_ARG(ckobj);
    op_free(entersubop);
    o = newOP(OP_CUSTOM, 0);
    o->op_ppaddr = MyOwn_pp_pmatch;
    return o;
}
XS(XS_MyOwn_pmatch) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    croak("MyOwn::pmatch was not compiled to its op");
}
XS(XS_MyOwn_install) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    cv_set_call_checker(get_cv("MyOwn::pmatch", 0), MyOwn_ck, &PL_sv_undef);
    XSRETURN_EMPTY;
}

XS_EXTERNAL(boot_MyOwn) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyOwn::pmatch", XS_MyOwn_pmatch);
    newXS_deffile("MyOwn::install", XS_MyOwn_install);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeMatchOpOfTheModulesOwn(t *testing.T) {
	so := buildModuleSource(t, "MyOwn", ownOpsRunSource)
	got := requireModule(t, "MyOwn", so, `MyOwn::install();
my $m = eval q{ sub { local ($MyOwn::re, $MyOwn::val) = @_; my @l = ("a", MyOwn::pmatch(), "z"); "@l" } }
    or die $@;
join "|", map { $m->(@$_) } [qr/b+/, "abbc"], [qr/^b/, "abbc"], [qr/C$/i, "abbc"]`)
	want := "a 1 z|a  z|a 1 z"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
