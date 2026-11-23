//go:build darwin || linux

package xs_test

import "testing"

// PL_op that the interpreter changed under an operation and native code put
// back is the interpreter's again, as on perl: B::Generate's constructors
// save PL_op, build an op the compiler folds (fold_constants runs the
// folded ops and leaves PL_op NULL), and restore it themselves; the program
// goes on from the call. Before, the restore looked like no change to the
// SDK, nothing was written back, and the program ended after the call.
const plopRestoreSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* fold(): 1 + 2 built at run time, folded; -> what the op became */
XS(XS_MyFold_fold) {
    dXSARGS;
    OP *const saved_op = PL_op;
    COP *const saved_cop = PL_curcop;
    OP *o;
    const char *what;
    PERL_UNUSED_VAR(items);
    PL_curcop = &PL_compiling;
    o = newBINOP(OP_ADD, 0, newSVOP(OP_CONST, 0, newSViv(1)), newSVOP(OP_CONST, 0, newSViv(2)));
    PL_curcop = saved_cop;
    PL_op = saved_op;
    what = o->op_type == OP_CONST ? "folded" : "not folded";
    op_free(o);
    ST(0) = sv_2mortal(newSVpv(what, 0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyFold) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyFold::fold", XS_MyFold_fold);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodePutsPLopBack(t *testing.T) {
	so := buildModuleSource(t, "MyFold", plopRestoreSource)
	got := requireModule(t, "MyFold", so, `my @r;
for my $i (1 .. 2) { push @r, MyFold::fold() . " $i" }
push @r, "after";
join ", ", @r`)
	if want := "folded 1, folded 2, after"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
