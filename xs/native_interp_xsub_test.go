//go:build darwin || linux

package xs_test

import "testing"

// Native code may call an XSUB of the interpreter's through its C
// function, CvXSUB(cv)(aTHX_ cv), with the mark and the arguments it
// pushed, and read the results where the XSUB left them (B::OPCheck calls
// B::UNOP::first so). Before, CvXSUB of such a sub was NULL and the call
// crashed the process (B-OPCheck, self, Hash-Lazy).
func TestNativeCallsAnInterpreterXSUBByItsFunction(t *testing.T) {
	so := buildModuleSource(t, "Direct", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_Direct_call) {
    dXSARGS;
    CV *target = get_cv(SvPV_nolen(ST(0)), 0);
    SV *result;
    I32 i;
    {
        dSP;
        PUSHMARK(SP);
        for (i = 1; i < items; i++) XPUSHs(ST(i));
        PUTBACK;
        CvXSUB(target)(aTHX_ target);
        SPAGAIN;
        result = POPs;
        PUTBACK;
    }
    ST(0) = result;
    XSRETURN(1);
}
XS_EXTERNAL(boot_Direct) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("Direct::call", XS_Direct_call);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	for _, tc := range []struct {
		name, code, want string
	}{
		{"a string of characters", `Direct::call("utf8::is_utf8", "\x{100}") ? "yes" : "no"`, "yes"},
		{"a string of bytes", `Direct::call("utf8::is_utf8", "a") ? "yes" : "no"`, "no"},
		{"the result of a call in scalar context",
			`my $r = Direct::call("Internals::SvREADONLY", \1); $r ? "yes" : "no"`, "yes"},
		{"a die of the XSUB", `eval { Direct::call("utf8::is_utf8") ; 1 } ? "lived" : $@ =~ /^Usage: utf8::is_utf8/ ? "died" : $@`, "died"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := requireModule(t, "Direct", so, tc.code); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
