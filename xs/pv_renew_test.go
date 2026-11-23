//go:build darwin || linux

package xs_test

import "testing"

// SvPV_renew for native code, as sv.h has it: the buffer of a string made
// exactly n bytes long, larger or smaller, with what it holds kept, and the
// buffer of a chopped string whose offset was taken off renewed to a
// smaller size, as Feersum keeps its read buffer (Feersum did not load
// without it: "undefined symbol: SvPV_renew"). What is expected is what
// perl 5.36 on Linux gives with the same module.
func TestNativeCodeRenewsTheBufferOfAString(t *testing.T) {
	so := buildModuleSource(t, "PvRenew", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* renew(str, grow, n): a copy of str grown to grow bytes, then renewed to
 * n with SvPV_renew; the length the SV has then, and what it holds after
 * appending "+" */
XS(XS_PvRenew_renew) {
    dXSARGS;
    STRLEN grow = (STRLEN)SvUV(ST(1)), n = (STRLEN)SvUV(ST(2)), len;
    SV *sv = sv_2mortal(newSVsv(ST(0)));
    PERL_UNUSED_VAR(items);
    SvGROW(sv, grow);
    if (SvCUR(sv) >= n) SvCUR_set(sv, n - 1);
    SvPV_renew(sv, n);
    len = SvLEN(sv);
    *SvEND(sv) = '\0';
    sv_catpvs(sv, "+");
    ST(0) = sv_2mortal(newSVpvf("%d:%" SVf, len == n ? 1 : 0, SVfARG(sv)));
    XSRETURN(1);
}
/* chopped(str, off, n): the buffer of a string chopped at off, with its
 * offset taken off and its length set to 0, renewed to n, as Feersum keeps
 * its read buffer; the length then and the string after appending "ab" */
XS(XS_PvRenew_chopped) {
    dXSARGS;
    STRLEN off = (STRLEN)SvUV(ST(1)), n = (STRLEN)SvUV(ST(2)), len;
    SV *sv = sv_2mortal(newSVsv(ST(0)));
    PERL_UNUSED_VAR(items);
    SvGROW(sv, 4096);
    sv_chop(sv, SvPVX(sv) + off);
    SvOOK_off(sv);
    SvCUR_set(sv, 0);
    if (SvLEN(sv) > n) SvPV_renew(sv, n);
    len = SvLEN(sv);
    sv_catpvs(sv, "ab");
    ST(0) = sv_2mortal(newSVpvf("%d:%" SVf, len == n ? 1 : 0, SVfARG(sv)));
    XSRETURN(1);
}
XS_EXTERNAL(boot_PvRenew) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("PvRenew::renew", XS_PvRenew_renew, __FILE__);
    newXS("PvRenew::chopped", XS_PvRenew_chopped, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `join ",", PvRenew::renew("hello", 4096, 32), PvRenew::renew("hello world", 64, 6),
  PvRenew::renew("x" x 100, 200, 1000), PvRenew::chopped("hello world", 6, 16)`
	if got, want := requireModule(t, "PvRenew", so, program), "1:hello+,1:hello+,1:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx+,1:ab"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
