//go:build darwin || linux

package xs_test

import "testing"

// A croak in a function of the module's own that has dXSARGS and returns a
// value (XS::JIT compiles such functions and mortalizes what they return)
// goes past its caller to the interpreter, as perl's croak does. Before,
// the function returned to its caller with whatever its return register
// held, which the caller then used (XS-JIT: "sv_2mortal: ... not one of
// this interpreter's" instead of the croak's message).
func TestNativeCroakInANestedFunctionWithDXSARGS(t *testing.T) {
	so := buildModuleSource(t, "NestedCroak", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static SV *must_have_arg(pTHX_ CV *cv) {
    dXSARGS;
    PERL_UNUSED_VAR(cv);
    if (items < 1) croak("Argument required");
    return newSVsv(ST(0));
}
XS(XS_NestedCroak_call) {
    dXSARGS;
    SV *r;
    I32 i;
    PUSHMARK(SP);
    for (i = 0; i < items; i++) XPUSHs(ST(i));
    PUTBACK;
    r = must_have_arg(aTHX_ cv);
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
XS_EXTERNAL(boot_NestedCroak) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("NestedCroak::call", XS_NestedCroak_call);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	if got, want := requireModule(t, "NestedCroak", so, `my $r = eval { NestedCroak::call(); 1 } ? "lived" : $@;
($r =~ /^Argument required at / ? "died" : $r) . " " . NestedCroak::call(42)`), "died 42"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
