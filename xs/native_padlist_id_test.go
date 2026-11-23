//go:build darwin || linux

package xs_test

import "testing"

// Native code gives a sub it makes a padlist of its own with pad_new and
// copies the id of another sub's padlist into it, as Future::AsyncAwait's
// copy of a suspended sub does, so that a closure made inside the copy
// finds its outer sub by the id (cv_clone's find_runcv): the interpreter
// has the id (B reads it). Before, the shadow CvPADLIST gave took the store
// but was not written back, and the copy's padlist kept the id 0.
const padlistIDSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MyPi_copyid) {
    dXSARGS;
    CV *const orig = (CV *)SvRV(ST(0));
    CV *const copy = (CV *)newSV_type(SVt_PVCV);
    PERL_UNUSED_VAR(items);
    ENTER;
    SAVESPTR(PL_compcv);
    PL_compcv = copy;
    SAVESPTR(PL_comppad_name);
    PL_comppad_name = PadlistNAMES(CvPADLIST(orig));
    CvPADLIST_set(copy, pad_new(padnew_CLONE | padnew_SAVE));
    CvPADLIST(copy)->xpadl_id = CvPADLIST(orig)->xpadl_id;
    av_fill(PL_comppad, 0);
    LEAVE;
    ST(0) = sv_2mortal(newRV_noinc((SV *)copy));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyPi) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyPi::copyid", XS_MyPi_copyid);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeCopiesAPadlistID(t *testing.T) {
	so := buildModuleSource(t, "MyPi", padlistIDSource)
	got := requireModule(t, "MyPi", so, `use B; my $x = 1; my $f = sub { $x };
my $c = MyPi::copyid($f);
my ($i, $j) = map { B::svref_2object($_)->PADLIST->id } $c, $f;
$i == $j && $j ? "same" : "different $i $j"`)
	if want := "same"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
