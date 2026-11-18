//go:build darwin || linux

package xs_test

import "testing"

// Names of perl's handy.h that modules use: READ_XDIGIT and XDIGIT_VALUE
// (Google::ProtocolBuffers::Dynamic's src/mapper.cpp did not compile
// without READ_XDIGIT), OCTAL_VALUE, isALPHA_FOLD_EQ and _NE, withinCOUNT,
// inRANGE, toTITLE, memLT and its kin, AV_FROM_REF, the limits of the small
// integer types and MUTABLE_IO. What is expected is what perl 5.36 on Linux
// gives with the same module.
func TestNativeHandyNames(t *testing.T) {
	so := buildModuleSource(t, "HandyNames", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* Names of perl's handy.h public to modules: names(\@array, "hex digits")
 * says what each gives. AV_FROM_REF is perl 5.38's; a perl before has none. */
#ifndef AV_FROM_REF
#define AV_FROM_REF(ref) ((AV *)SvRV(ref))
#endif
XS(XS_HandyNames_names) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "aref, hex");
    AV *av = AV_FROM_REF(ST(0));
    const char *s = SvPV_nolen(ST(1));
    UV hex = 0;
    while (isXDIGIT(*s)) hex = hex * 16 + READ_XDIGIT(s);
    char *copy;
    Newx(copy, 7, char);
    Copy("copied", copy, 7, char);
    SV *out = newSVpvf(
        "hex %" UVuf " xd %d oct %d fold %d%d count %d%d range %d%d title %c mem %d%d%d%d "
        "aref %d copy %s max %u %u %d %u io %d",
        hex, (int)XDIGIT_VALUE('b'), (int)OCTAL_VALUE('7'), isALPHA_FOLD_EQ('a', 'A') ? 1 : 0,
        isALPHA_FOLD_NE('a', 'b') ? 1 : 0, withinCOUNT('c', 'a', 2) ? 1 : 0,
        withinCOUNT('d', 'a', 2) ? 1 : 0, inRANGE('5', '0', '9') ? 1 : 0,
        inRANGE('x', '0', '9') ? 1 : 0, toTITLE('q'), memLT("ab", "ac", 2) ? 1 : 0,
        memLE("ab", "ab", 2) ? 1 : 0, memGT("b", "a", 1) ? 1 : 0, memGE("a", "b", 1) ? 1 : 0,
        (int)av_len(av), copy, (unsigned)U8_MAX, (unsigned)U16_MAX, (int)I16_MIN,
        (unsigned)U32_MIN, MUTABLE_IO(GvIO(gv_fetchpvs("STDOUT", 0, SVt_PVIO))) != NULL);
    Safefree(copy);
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
XS_EXTERNAL(boot_HandyNames) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("HandyNames::names", XS_HandyNames_names, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `HandyNames::names([1, 2, 3], "fF9")`
	want := "hex 4089 xd 11 oct 7 fold 11 count 10 range 10 title Q mem 1110 aref 2 copy copied max 255 65535 -32768 0 io 1"
	if got := requireModule(t, "HandyNames", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
