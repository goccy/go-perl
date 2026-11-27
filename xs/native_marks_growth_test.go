//go:build darwin || linux

package xs_test

import "testing"

// The marks of a native call grow as perl's mark stack does. Tcl's command
// for perl code pushes a mark for eval_sv, which takes none, so each call
// leaves one behind on perl too; Tcl's event loop runs thousands of those
// in one call of vwait (Tcl's t/disposal-subs.t), and the SDK's 64 marks of
// a frame ran out ("mark stack overflow in native XS"). The module is the
// same code; what is expected is what perl 5.36 on Linux gives with it.
func TestNativeMarksGrowAsPerls(t *testing.T) {
	so := buildModuleSource(t, "LeakMark", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* Evals code n times as Tcl's command for perl code does: it pushes a mark
 * for eval_sv, which takes none, so every eval leaves one behind. Then
 * calls cb with the sum of what the evals gave and n, with a mark above all
 * of those, and gives what cb gave. */
XS(XS_LeakMark_evals) {
    dXSARGS;
    SV *code = ST(0), *cb = ST(2), *r;
    IV n = SvIV(ST(1)), i, total = 0;
    PERL_UNUSED_VAR(items);
    for (i = 0; i < n; i++) {
        ENTER;
        SAVETMPS;
        PUSHMARK(SP);
        PUTBACK;
        (void)eval_sv(sv_2mortal(newSVsv(code)), G_EVAL | G_SCALAR);
        SPAGAIN;
        total += POPi;
        PUTBACK;
        FREETMPS;
        LEAVE;
    }
    ENTER;
    SAVETMPS;
    PUSHMARK(SP);
    XPUSHs(sv_2mortal(newSViv(total)));
    XPUSHs(sv_2mortal(newSViv(n)));
    PUTBACK;
    (void)call_sv(cb, G_SCALAR);
    SPAGAIN;
    r = newSVsv(POPs);
    PUTBACK;
    FREETMPS;
    LEAVE;
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
XS_EXTERNAL(boot_LeakMark) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("LeakMark::evals", XS_LeakMark_evals, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $r = LeakMark::evals("2", 1000, sub { join "/", @_ }); $r`
	if got, want := requireModule(t, "LeakMark", so, program), "2000/1000"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
