//go:build darwin || linux

package xs_test

import "testing"

// AvARRAY of a reference reads the body slot as perl's macro does, which of
// a reference is the referent: Array::Heap takes AvARRAY of the reference
// for that of the array and tests its first word before it reads the
// array's first element, and perl finds the array's body there. The SDK
// gave NULL for anything but an array, and the module died on a null
// pointer as soon as a heap held array references. What is expected is
// what perl 5.36 on Linux gives with the same module.
func TestNativeAvARRAYOfAReference(t *testing.T) {
	so := buildModuleSource(t, "AvRef", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* first(x): the first element of an array behind a reference, taken as
 * Array::Heap's sv_first takes it: AvARRAY of the reference (not of the
 * array) tested for its first word before the array's first element is
 * read; anything else is itself */
static SV *first_of(SV *sv) {
    if (SvROK(sv) && SvTYPE(SvRV(sv)) == SVt_PVAV) {
        AV *av = (AV *)SvRV(sv);
        sv = AvFILLp(av) < 0 || !AvARRAY(sv)[0] ? &PL_sv_undef : AvARRAY(av)[0];
    }
    return sv;
}
XS(XS_AvRef_first) {
    dXSARGS;
    SV *const sv = first_of(ST(0));
    PERL_UNUSED_VAR(items);
    ST(0) = SvOK(sv) ? sv_2mortal(newSVsv(sv)) : sv_2mortal(newSVpvs("undef"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_AvRef) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("AvRef::first", XS_AvRef_first, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `join ",", map { AvRef::first($_) } [4, "hi"], 3, [7], [], [undef, 2], "x"`
	if got, want := requireModule(t, "AvRef", so, program), "4,3,7,undef,undef,x"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
