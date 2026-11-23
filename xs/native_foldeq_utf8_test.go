//go:build darwin || linux

package xs_test

import "testing"

// foldEQ_utf8 for native code, as perl's: full folds (sharp s, final sigma,
// the Kelvin sign, dotted capital I), a goal on one string with an end
// pointer on the other giving where the match ended, and a Latin-1 string
// against a UTF-8 one. Syntax::Infix::EqualityInsensitive did not load
// without it ("undefined symbol: foldEQ_utf8"). What is expected is what
// perl 5.36 on Linux gives with the same module.
func TestNativeFoldEQUtf8(t *testing.T) {
	so := buildModuleSource(t, "FoldEQ", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* eqi(a, b): the two strings the same case-insensitively, as
 * Syntax::Infix::EqualityInsensitive compares them */
XS(XS_FoldEQ_eqi) {
    dXSARGS;
    SV *a = ST(0), *b = ST(1);
    const char *sa, *sb;
    STRLEN la, lb;
    PERL_UNUSED_VAR(items);
    sa = SvPVutf8(a, la);
    sb = SvPVutf8(b, lb);
    ST(0) = foldEQ_utf8(sa, NULL, (UV)la, TRUE, sb, NULL, (UV)lb, TRUE) ? &PL_sv_yes : &PL_sv_no;
    XSRETURN(1);
}
/* prefix(a, b): whether a begins with b case-insensitively (a scanned up to
 * its end at most, b to its whole length), and where in a the match ended */
XS(XS_FoldEQ_prefix) {
    dXSARGS;
    const char *sa, *sb;
    STRLEN la, lb;
    char *pe;
    PERL_UNUSED_VAR(items);
    sa = SvPVutf8(ST(0), la);
    sb = SvPVutf8(ST(1), lb);
    pe = (char *)sa + la;
    if (foldEQ_utf8(sa, &pe, 0, TRUE, sb, NULL, (UV)lb, TRUE))
        ST(0) = sv_2mortal(newSViv((IV)(pe - sa)));
    else
        ST(0) = sv_2mortal(newSViv(-1));
    XSRETURN(1);
}
/* latin1(a, b): a as Latin-1 bytes against b as UTF-8 */
XS(XS_FoldEQ_latin1) {
    dXSARGS;
    const char *sa, *sb;
    STRLEN la, lb;
    PERL_UNUSED_VAR(items);
    sa = SvPVbyte(ST(0), la);
    sb = SvPVutf8(ST(1), lb);
    ST(0) = foldEQ_utf8(sa, NULL, (UV)la, FALSE, sb, NULL, (UV)lb, TRUE) ? &PL_sv_yes : &PL_sv_no;
    XSRETURN(1);
}
XS_EXTERNAL(boot_FoldEQ) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("FoldEQ::eqi", XS_FoldEQ_eqi, __FILE__);
    newXS("FoldEQ::prefix", XS_FoldEQ_prefix, __FILE__);
    newXS("FoldEQ::latin1", XS_FoldEQ_latin1, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my @pairs = (["Stra\x{df}e", "STRASSE"], ["\x{3a3}\x{391}\x{3a3}", "\x{3c3}\x{3b1}\x{3c2}"],
             ["abc", "ABD"], ["k", "\x{212a}"], ["abc", "ab"], ["\x{130}", "i\x{307}"]);
my @got = map { FoldEQ::eqi(@$_) ? 1 : 0 } @pairs;
push @got, FoldEQ::prefix("Stra\x{df}e und", "STRASSE"), FoldEQ::prefix("abc", "abd");
push @got, FoldEQ::latin1("CAF\xc9", "caf\x{e9}") ? 1 : 0, FoldEQ::latin1("\xdf", "ss") ? 1 : 0;
join ",", @got`
	if got, want := requireModule(t, "FoldEQ", so, program), "1,1,0,1,0,1,7,-1,1,1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
