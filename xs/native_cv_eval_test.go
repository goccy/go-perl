//go:build darwin || linux

package xs_test

import "testing"

// CvEVAL says the sub being compiled is a string eval's (unique and not
// fake, as cv.h has it), which Future::AsyncAwait asks to word its error
// for an await there. Before, the SDK had CvSPECIAL's test under CvEVAL's
// name: a string eval's sub was not CvEVAL.
const cvEvalSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MyCe_kind) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%s%s", CvEVAL(PL_compcv) ? "e" : "-",
                                PL_compcv == PL_main_cv ? "m" : "-"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyCe) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyCe::kind", XS_MyCe_kind);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeTellsAStringEvalsSub(t *testing.T) {
	so := buildModuleSource(t, "MyCe", cvEvalSource)
	got := requireModule(t, "MyCe", so, `our @r;
eval q{ BEGIN { push @r, MyCe::kind() } 1 } or die $@;
eval q{ sub named { BEGIN { push @r, MyCe::kind() } } 1 } or die $@;
join ",", @r`)
	if want := "e-,--"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
