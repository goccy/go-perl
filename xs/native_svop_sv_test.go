//go:build darwin || linux

package xs_test

import "testing"

// The SV native code stores into an SVOP is the op's: B::Generate's
// SVOP::sv replaces a constant with cSVOPx(o)->op_sv = newSVsv(...). Before,
// the store stayed in the op's shadow and the sub went on returning the old
// constant. The expected string is what perl prints with the same module
// code; a perl with threads keeps a sub's constant in the pad, where the
// module stores it then.
const svopSvSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

static OP *MySvop_find(pTHX_ OP *o, int type) {
    OP *k;
    if (!o) return NULL;
    if (o->op_type == type) return o;
    if (o->op_flags & OPf_KIDS)
        for (k = cUNOPo->op_first; k; k = OpSIBLING(k)) {
            OP *r = MySvop_find(aTHX_ k, type);
            if (r) return r;
        }
    return NULL;
}

/* the constant of a sub replaced with another value, where the op keeps it */
XS(XS_MySvop_reconst) {
    dXSARGS;
    CV *c;
    OP *o;
    if (items != 2) croak_xs_usage(cv, "code, value");
    c = (CV *)SvRV(ST(0));
    o = MySvop_find(aTHX_ CvROOT(c), OP_CONST);
    if (!o) croak("no const op");
    if (cSVOPx(o)->op_sv) {
        cSVOPx(o)->op_sv = newSVsv(ST(1));
    } else {
        PAD *const pad = PadlistARRAY(CvPADLIST(c))[1];
        AvARRAY(pad)[o->op_targ] = newSVsv(ST(1));
    }
    XSRETURN_EMPTY;
}

XS_EXTERNAL(boot_MySvop) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySvop::reconst", XS_MySvop_reconst);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeStoreOfTheSVOfAnSVOP(t *testing.T) {
	so := buildModuleSource(t, "MySvop", svopSvSource)
	got := requireModule(t, "MySvop", so, `my $g = sub { 30 + $_[0] };
my $before = $g->(1);
MySvop::reconst($g, 13);
my $after = $g->(1);
MySvop::reconst($g, "7");
join "|", $before, $after, $g->(2)`)
	want := "31|14|9"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
