//go:build darwin || linux

package xs_test

import "testing"

// An op of native code's own run through the PL_ppaddr slot of its type is
// run as an op of that type, also when the type shares its pp function
// with another: Tie::Hash::Indexed's dor_assign makes a BINOP of type
// OP_DOR, points PL_op at it and calls PL_ppaddr[OP_DOR], which is
// pp_defined, the function of OP_DEFINED too, telling the two apart by the
// op's type. Before, the slot ran it as OP_DEFINED: the call gave whether
// the value is defined instead of the value or the default. The expected
// string is what perl prints with the same module code.
const sharedPpTypeSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

/* the logical operator of the type on value and default, as
 * Tie::Hash::Indexed's dor_assign and or_assign run them */
static SV *MyDor_run(pTHX_ int type, SV *value, SV *dflt) {
    dSP;
    OP *oldop;
    BINOP myop;
    SV *r;
    Zero(&myop, 1, struct op);
    myop.op_flags = OPf_STACKED;
    myop.op_type = type;
    ENTER;
    SAVETMPS;
    PUSHMARK(SP);
    XPUSHs(dflt);
    XPUSHs(value);
    PUTBACK;
    oldop = PL_op;
    PL_op = (OP *)&myop;
    PL_ppaddr[PL_op->op_type](aTHX);
    PL_op = oldop;
    SPAGAIN;
    r = newSVsv(TOPs);
    POPMARK;
    FREETMPS;
    LEAVE;
    return r;
}
XS(XS_MyDor_dor) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "value, default");
    ST(0) = sv_2mortal(MyDor_run(aTHX_ OP_DOR, ST(0), ST(1)));
    XSRETURN(1);
}
XS(XS_MyDor_or) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "value, default");
    ST(0) = sv_2mortal(MyDor_run(aTHX_ OP_OR, ST(0), ST(1)));
    XSRETURN(1);
}

XS_EXTERNAL(boot_MyDor) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyDor::dor", XS_MyDor_dor);
    newXS_deffile("MyDor::or", XS_MyDor_or);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeOpOfATypeThatSharesItsPpFunction(t *testing.T) {
	so := buildModuleSource(t, "MyDor", sharedPpTypeSource)
	got := requireModule(t, "MyDor", so, `join "|", map { defined $_ ? "[$_]" : "undef" }
    MyDor::dor(0, 1), MyDor::dor(undef, 1), MyDor::dor("x", 1),
    MyDor::or(0, 1), MyDor::or(undef, 1), MyDor::or("x", 1)`)
	want := "[0]|[1]|[x]|[1]|[1]|[x]"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
