//go:build darwin || linux

package xs_test

import "testing"

// The class of a method op ($obj->Class::meth, op_rclass_sv of a METHOP as
// perl without threads keeps it) is what native code reads in it, and a
// class native code stores there is the op's, as next::XS rewrites the
// method ops of super::NAME calls at run time (cMETHOPx_rclass, with the
// reference counts kept by the module). Before, the op's shadow had no
// class and a stored one stayed in the shadow. The expected string is what
// perl prints with the same module code; a perl with threads keeps the
// class in the sub's pad, where the module reads and stores it then.
const methodClassSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

static OP *MyMcls_find(pTHX_ OP *o, int type) {
    OP *k;
    if (!o) return NULL;
    if (o->op_type == type) return o;
    if (o->op_flags & OPf_KIDS)
        for (k = cUNOPo->op_first; k; k = OpSIBLING(k)) {
            OP *r = MyMcls_find(aTHX_ k, type);
            if (r) return r;
        }
    return NULL;
}
/* the slot of the class of the sub's method_redir op */
static SV **MyMcls_slot(pTHX_ CV *c) {
    OP *const o = MyMcls_find(aTHX_ CvROOT(c), OP_METHOD_REDIR);
    if (!o) croak("no method_redir op");
#ifdef USE_ITHREADS
    return &AvARRAY(PadlistARRAY(CvPADLIST(c))[1])[cMETHOPx(o)->op_rclass_targ];
#else
    return &cMETHOPx(o)->op_rclass_sv;
#endif
}
XS(XS_MyMcls_class) {
    dXSARGS;
    SV **slot;
    if (items != 1) croak_xs_usage(cv, "code");
    slot = MyMcls_slot(aTHX_ (CV *)SvRV(ST(0)));
    ST(0) = sv_2mortal(*slot ? newSVsv(*slot) : newSVpvs("none"));
    XSRETURN(1);
}
XS(XS_MyMcls_reclass) {
    dXSARGS;
    SV **slot, *old;
    if (items != 2) croak_xs_usage(cv, "code, class");
    slot = MyMcls_slot(aTHX_ (CV *)SvRV(ST(0)));
    old = *slot;
    *slot = newSVsv(ST(1));
    SvREFCNT_dec(old);
    XSRETURN_EMPTY;
}

XS_EXTERNAL(boot_MyMcls) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyMcls::class", XS_MyMcls_class);
    newXS_deffile("MyMcls::reclass", XS_MyMcls_reclass);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeClassOfAMethodOp(t *testing.T) {
	so := buildModuleSource(t, "MyMcls", methodClassSource)
	got := requireModule(t, "MyMcls", so, `package First { sub who { "first" } }
package Second { sub who { "second" } }
my $obj = bless {}, "Main";
my $call = sub { $obj->First::who() };
my @r = (MyMcls::class($call), $call->());
MyMcls::reclass($call, "Second");
push @r, MyMcls::class($call), $call->(), $call->();
join "|", @r`)
	want := "First|first|Second|second|second"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
