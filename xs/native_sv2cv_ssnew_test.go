//go:build darwin || linux

package xs_test

import "testing"

// sv_2cv gives what perl's gives in its out-parameters: the glob and the
// stash, also for the glob of a sub that is not defined (no cv), which
// Devel::NYTProf names the sub of a call that will die by (t/test13.t: the
// call of main::x was missing). And SSNEW gives a block made inside the
// scope of another a higher index, as perl's savestack offsets are, however
// many blocks were made before: Devel::NYTProf gave up profiling ("stack is
// confused") once 4096 had been made.
const sv2cvModuleSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MySv2cv_resolve) {
    dXSARGS;
    HV *stash = NULL;
    GV *gv = NULL;
    CV *code;
    if (items != 1) croak_xs_usage(cv, "sv");
    code = sv_2cv(ST(0), &stash, &gv, 0);
    ST(0) = sv_2mortal(newSVpvf("%s %s %s", code ? "cv" : "no cv",
                                gv ? GvNAME(gv) : "no glob", stash ? HvNAME(stash) : "no stash"));
    XSRETURN(1);
}
/* n blocks one inside the scope of the other, each from a call of the sub
 * given, after as many blocks made and released one after the other: 1
 * when every block has a higher index than the one it is inside */
static int nested_ok;
static void nest(pTHX_ SV *code, int n, SSize_t outer) {
    SSize_t ix;
    ENTER;
    ix = SSNEW(16);
    if (ix <= outer) nested_ok = 0;
    if (n > 1) nest(aTHX_ code, n - 1, ix);
    LEAVE;
}
XS(XS_MySv2cv_ssnew) {
    dXSARGS;
    int i, made = (int)SvIV(ST(0)), depth = (int)SvIV(ST(1));
    PERL_UNUSED_VAR(items);
    for (i = 0; i < made; i++) {
        ENTER;
        (void)SSNEW(8);
        LEAVE;
    }
    nested_ok = 1;
    nest(aTHX_ NULL, depth, -1);
    ST(0) = sv_2mortal(newSViv(nested_ok));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MySv2cv) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySv2cv::resolve", XS_MySv2cv_resolve);
    newXS_deffile("MySv2cv::ssnew", XS_MySv2cv_ssnew);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeSv2cvAndSsnew(t *testing.T) {
	so := buildModuleSource(t, "MySv2cv", sv2cvModuleSource)
	for _, tc := range []struct{ name, code, want string }{
		{"sv_2cv of the glob of a defined sub", `sub Some::defined { 1 } MySv2cv::resolve(\*Some::defined)`, "cv defined Some"},
		{"sv_2cv of the glob of an undefined sub", `MySv2cv::resolve(\*main::not_defined_here)`, "no cv not_defined_here main"},
		{"sv_2cv of a code reference", `MySv2cv::resolve(sub { 1 })`, "cv no glob main"},
		{"SSNEW nested across 4096 blocks made", `MySv2cv::ssnew(4080, 50)`, "1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := requireModule(t, "MySv2cv", so, tc.code); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
