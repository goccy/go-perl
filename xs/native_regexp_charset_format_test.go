//go:build darwin || linux

package xs_test

import "testing"

// get_regex_charset reads the character set of a pattern's flags from the
// three bits perl keeps it in (from bit 7), and set_regex_charset writes it
// there and leaves the other flags alone; the SDK read and wrote them two
// bits lower, so /aa came out as no charset and /l as /aa. And a format is a
// CV to CvGV, CvDEPTH, CvFILE and CvSTASH as a sub is: the SDK answered
// nothing for one, so its glob was not found. Struct::Codec encodes both
// (a regexp with its modifiers; a format by its name). What is expected is
// what perl 5.36 on Linux gives with the same module.
func TestNativeRegexpCharsetAndFormatGlob(t *testing.T) {
	so := buildModuleSource(t, "RxFmt", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* charset(qr): the character set of a compiled pattern, as get_regex_charset
 * reads it from the pattern's flags */
XS(XS_RxFmt_charset) {
    dXSARGS;
    REGEXP *rx = SvRX(ST(0));
    const char *name = "?";
    PERL_UNUSED_VAR(items);
    switch (get_regex_charset(RX_EXTFLAGS(rx))) {
    case REGEX_DEPENDS_CHARSET: name = "d"; break;
    case REGEX_LOCALE_CHARSET: name = "l"; break;
    case REGEX_UNICODE_CHARSET: name = "u"; break;
    case REGEX_ASCII_RESTRICTED_CHARSET: name = "a"; break;
    case REGEX_ASCII_MORE_RESTRICTED_CHARSET: name = "aa"; break;
    }
    ST(0) = sv_2mortal(newSVpv(name, 0));
    XSRETURN(1);
}
/* setcs(flags of a qr, which): set_regex_charset of the flags to /aa (1)
 * or /a (0), read back, and whether the other flags stayed */
XS(XS_RxFmt_setcs) {
    dXSARGS;
    U32 fl = RX_EXTFLAGS(SvRX(ST(0)));
    const U32 before = fl & ~RXf_PMf_CHARSET;
    PERL_UNUSED_VAR(items);
    set_regex_charset(&fl, SvIV(ST(1)) ? REGEX_ASCII_MORE_RESTRICTED_CHARSET
                                       : REGEX_ASCII_RESTRICTED_CHARSET);
    ST(0) = sv_2mortal(newSVpvf("%d %s", (int)get_regex_charset(fl),
                                (fl & ~RXf_PMf_CHARSET) == before ? "kept" : "changed"));
    XSRETURN(1);
}
/* fmt(\*glob{FORMAT}): the glob a format is in (CvGV), as package::name, its
 * depth and whether it has a file */
XS(XS_RxFmt_fmt) {
    dXSARGS;
    CV *fm = (CV *)SvRV(ST(0));
    GV *gv = CvGV(fm);
    PERL_UNUSED_VAR(items);
    if (!gv) {
        ST(0) = sv_2mortal(newSVpvs("no glob"));
    } else {
        ST(0) = sv_2mortal(newSVpvf("%s::%s %d %s %s", HvNAME(GvSTASH(gv)), GvNAME(gv), (int)CvDEPTH(fm),
                                    CvFILE(fm) ? "file" : "nofile", CvSTASH(fm) ? HvNAME(CvSTASH(fm)) : "nostash"));
    }
    XSRETURN(1);
}
XS_EXTERNAL(boot_RxFmt) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("RxFmt::charset", XS_RxFmt_charset, __FILE__);
    newXS("RxFmt::setcs", XS_RxFmt_setcs, __FILE__);
    newXS("RxFmt::fmt", XS_RxFmt_fmt, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `package Report;
format SUMMARY =
@<<<<< @>>>>
$a, $b
.
package main;
my @got = map { RxFmt::charset($_) } qr/x/, qr/x/l, qr/x/u, qr/x/a, qr/x/aa, qr/x/aai, qr/x/msixpn;
push @got, RxFmt::setcs(qr/x/msixn, 1), RxFmt::setcs(qr/x/msixn, 0), RxFmt::setcs(qr/x/u, 1);
my $f = *Report::SUMMARY{FORMAT}; push @got, ref($f), RxFmt::fmt($f);
join ",", @got`
	if got, want := requireModule(t, "RxFmt", so, program), "d,l,u,a,aa,aa,d,4 kept,3 kept,4 kept,FORMAT,Report::SUMMARY 0 file nostash"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
