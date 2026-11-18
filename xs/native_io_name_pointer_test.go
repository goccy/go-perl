//go:build darwin || linux

package xs_test

import "testing"

// A name word of an IO (IoFMT_NAME) set to an address keeps that address,
// as perl's char * field does: Filter::Util::Exec keeps a pointer into its
// buffer there and advances it as the bytes go down the pipe. Every store
// of a name word became a fresh copy of an empty string, the source the
// filter's command was given was garbage, and the filtered program died
// on an unrecognized character (Filter t/sh.t, t/exec.t). What is expected
// is what perl 5.36 on Linux gives with the same module.
func TestNativeIONameWordKeepsAnAddress(t *testing.T) {
	so := buildModuleSource(t, "IoBuf", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* Filter::Util::Exec's bookkeeping in the words of an IO: a buffer SV in
 * IoTOP_GV, a pointer into the buffer's string in IoFMT_NAME, advanced as
 * bytes are taken, an offset in IoPAGE_LEN. walk(text, step): the bytes
 * taken step at a time through the pointer, and whether it stays within */
#define BUF_SV(sv) IoTOP_GV(sv)
#define BUF_START(sv) SvPVX((SV *)BUF_SV(sv))
#define BUF_SIZE(sv) SvCUR((SV *)BUF_SV(sv))
#define BUF_NEXT(sv) IoFMT_NAME(sv)
#define BUF_END(sv) (BUF_START(sv) + BUF_SIZE(sv))
XS(XS_IoBuf_walk) {
    dXSARGS;
    SV *sv, *out;
    int step;
    if (items != 2) croak_xs_usage(cv, "text, step");
    step = (int)SvIV(ST(1));
    sv = newSV(0);
    (void)SvUPGRADE(sv, SVt_PVIO);
    BUF_SV(sv) = (GV *)newSV(1);
    (void)SvPOK_only((SV *)BUF_SV(sv));
    sv_setsv((SV *)BUF_SV(sv), ST(0));
    BUF_NEXT(sv) = NULL;
    IoPAGE_LEN(sv) = 0;
    out = newSVpvs("");
    if (!BUF_NEXT(sv)) BUF_NEXT(sv) = BUF_START(sv);
    while (BUF_NEXT(sv) < BUF_END(sv)) {
        int w = (int)(BUF_END(sv) - BUF_NEXT(sv));
        if (w > step) w = step;
        sv_catpvn(out, BUF_NEXT(sv), w);
        sv_catpvs(out, "|");
        BUF_NEXT(sv) += w;
        IoPAGE_LEN(sv) += w;
    }
    sv_catpvf(out, " taken %" IVdf " at end %d", (IV)IoPAGE_LEN(sv), BUF_NEXT(sv) == BUF_END(sv));
    SvREFCNT_dec((SV *)BUF_SV(sv));
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
XS_EXTERNAL(boot_IoBuf) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("IoBuf::walk", XS_IoBuf_walk, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `join " / ", IoBuf::walk("print 'hello';\n", 4), IoBuf::walk("abcdefghij", 3)`
	if got, want := requireModule(t, "IoBuf", so, program), "prin|t 'h|ello|';\n| taken 15 at end 1 / abc|def|ghi|j| taken 10 at end 1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
