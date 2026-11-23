//go:build darwin || linux

package xs_test

import "testing"

// Names of perl's API native code lacked, each against perl: my_atof3
// (FU), which leaves the value as it was where it answers NULL; call_list
// (the boot code of an older ExtUtils::ParseXS, Memcached::libmemcached);
// hv_scalar of a hash and a tied one (Audio::FLAC::Header); packlist and
// unpackstring, whose values come on the stack (Crypt::OpenSSL3,
// Thread::Channel); sv_utf8_downgrade_flags and sv_utf8_upgrade_flags_grow
// (LMDB_File, Syntax::Feature::QwComments); the strict UTF-8 string checks;
// a list of pad names made, filled and freed (B::Flags); the classes of
// code points below and above 256, under the locale and at a UTF-8 pointer
// (re::engine::GNU); and nBIT_MASK. What is expected is what perl 5.36 on
// Linux gives with the same module.
func TestNativeCodeUsesMorePerlAPINames(t *testing.T) {
	so := buildModuleSource(t, "ApiNames", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* atof3(str, len): my_atof3 of the first len bytes: the number, and how
 * many bytes it took (-1 for NULL) */
XS(XS_ApiNames_atof3) {
    dXSARGS;
    STRLEN len;
    const char *s = SvPV(ST(0), len);
    STRLEN n = (STRLEN)SvUV(ST(1));
    NV v = -1;
    char *end = my_atof3(s, &v, n);
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%" NVgf "|%d", v, end ? (int)(end - s) : -1));
    XSRETURN(1);
}
/* calllist(name...): the named subs run through call_list; how many are
 * left on the array after */
XS(XS_ApiNames_calllist) {
    dXSARGS;
    AV *av = newAV();
    I32 i;
    SSize_t left;
    for (i = 0; i < items; i++)
        av_push(av, SvREFCNT_inc((SV *)get_cv(SvPV_nolen(ST(i)), 0)));
    call_list(PL_scopestack_ix, av);
    left = av_top_index(av) + 1;
    SvREFCNT_dec((SV *)av);
    ST(0) = sv_2mortal(newSViv((IV)left));
    XSRETURN(1);
}
/* hscalar(\%h): hv_scalar */
XS(XS_ApiNames_hscalar) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = hv_scalar((HV *)SvRV(ST(0)));
    XSRETURN(1);
}
/* pack(pattern, values...): packlist */
XS(XS_ApiNames_pack) {
    dXSARGS;
    STRLEN plen;
    const char *pat = SvPV(ST(0), plen);
    SV *cat = sv_2mortal(newSVpvs(""));
    packlist(cat, pat, pat + plen, &ST(1), &ST(items));
    ST(0) = cat;
    XSRETURN(1);
}
/* unpack(pattern, string): unpackstring's values */
XS(XS_ApiNames_unpack) {
    dXSARGS;
    STRLEN plen, slen;
    const char *pat = SvPV(ST(0), plen);
    const char *str = SvPV(ST(1), slen);
    SSize_t count;
    PERL_UNUSED_VAR(items);
    SP -= items;
    PUTBACK;
    count = unpackstring(pat, pat + plen, str, str + slen, 0);
    SPAGAIN;
    XSRETURN(count);
}
/* utf8(str, extra): downgraded (failing quietly) with the flags of
 * sv_utf8_downgrade_flags, and a copy upgraded with room for extra more */
XS(XS_ApiNames_utf8) {
    dXSARGS;
    SV *d = sv_2mortal(newSVsv(ST(0)));
    SV *u = sv_2mortal(newSVsv(ST(0)));
    STRLEN extra = (STRLEN)SvUV(ST(1)), n;
    bool ok = sv_utf8_downgrade_flags(d, TRUE, 0);
    bool ok2 = sv_utf8_downgrade_nomg(sv_2mortal(newSVsv(ST(0))), TRUE);
    PERL_UNUSED_VAR(items);
    n = sv_utf8_upgrade_flags_grow(u, SV_GMAGIC, extra);
    ST(0) = sv_2mortal(newSVpvf("%d%d:%d:%d:%d", ok ? 1 : 0, ok2 ? 1 : 0, (int)SvCUR(d), (int)n,
                                SvLEN(u) >= SvCUR(u) + extra + 1 ? 1 : 0));
    XSRETURN(1);
}
/* strict(bytes): is_strict_utf8_string, is_c9strict_utf8_string,
 * is_utf8_string, is_utf8_fixed_width_buf_flags; and where the strict one
 * stopped */
XS(XS_ApiNames_strict) {
    dXSARGS;
    STRLEN len, chars = 0;
    const U8 *s = (const U8 *)SvPVbyte(ST(0), len);
    const U8 *ep = NULL;
    PERL_UNUSED_VAR(items);
    (void)is_strict_utf8_string_loclen(s, len, &ep, &chars);
    ST(0) = sv_2mortal(newSVpvf("%d%d%d%d:%d:%d", is_strict_utf8_string(s, len) ? 1 : 0,
                                is_c9strict_utf8_string(s, len) ? 1 : 0,
                                is_utf8_string(s, len) ? 1 : 0,
                                is_utf8_fixed_width_buf_flags(s, len, 0) ? 1 : 0, (int)(ep - s),
                                (int)chars));
    XSRETURN(1);
}
/* names(): a list of pad names made, filled with one and freed */
XS(XS_ApiNames_names) {
    dXSARGS;
    PADNAMELIST *pnl = newPADNAMELIST(3);
    SSize_t before = PadnamelistMAX(pnl);
    PADNAME *pn = newPADNAMEpvn("$x", 2);
    SV *out;
    PERL_UNUSED_VAR(items);
    padnamelist_store(pnl, 0, pn);
    out = newSVpvf("%d,%d,%s", (int)before, (int)PadnamelistMAX(pnl),
                   PadnamePV(padnamelist_fetch(pnl, 0)));
    PadnamelistREFCNT_dec(pnl);
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
/* classes(cp): the classes of a code point, then under the locale, then
 * of its UTF-8, then its lower case */
XS(XS_ApiNames_classes) {
    dXSARGS;
    UV c = SvUV(ST(0));
    U8 buf[UTF8_MAXBYTES_CASE + 1], *e;
    STRLEN l;
    char out[64];
    int i = 0;
    PERL_UNUSED_VAR(items);
    out[i++] = isLOWER_uvchr(c) ? 'l' : '-';
    out[i++] = isCNTRL_uvchr(c) ? 'c' : '-';
    out[i++] = isGRAPH_uvchr(c) ? 'g' : '-';
    out[i++] = isPRINT_uvchr(c) ? 'p' : '-';
    out[i++] = isPUNCT_uvchr(c) ? 'u' : '-';
    out[i++] = isXDIGIT_uvchr(c) ? 'x' : '-';
    out[i++] = isALPHANUMERIC_uvchr(c) ? 'a' : '-';
    out[i++] = isBLANK_uvchr(c) ? 'b' : '-';
    out[i++] = isPSXSPC_uvchr(c) ? 's' : '-';
    out[i++] = isASCII_uvchr(c) ? 'A' : '-';
    out[i++] = '/';
    out[i++] = isLOWER_LC_uvchr(c) ? 'l' : '-';
    out[i++] = isALPHA_LC_uvchr(c) ? 'a' : '-';
    out[i++] = isWORDCHAR_LC_uvchr(c) ? 'w' : '-';
    out[i++] = isSPACE_LC_uvchr(c) ? 's' : '-';
    out[i++] = isPUNCT_LC_uvchr(c) ? 'u' : '-';
    out[i++] = '/';
    e = uvchr_to_utf8(buf, c);
    out[i++] = isLOWER_utf8_safe(buf, e) ? 'l' : '-';
    out[i++] = isPRINT_utf8_safe(buf, e) ? 'p' : '-';
    out[i++] = isPUNCT_utf8_safe(buf, e) ? 'u' : '-';
    out[i++] = isALPHA_utf8(buf, e) ? 'a' : '-';
    out[i++] = isSPACE_utf8_safe(buf, e) ? 's' : '-';
    out[i++] = isXDIGIT_utf8_safe(buf, e) ? 'x' : '-';
    out[i++] = isUPPER_LC_utf8_safe(buf, e) ? 'U' : '-';
    out[i] = '\0';
    ST(0) = sv_2mortal(newSVpvf("%s/%" UVxf, out, toLOWER_uni(c, buf, &l)));
    XSRETURN(1);
}
/* mask(n): nBIT_MASK */
XS(XS_ApiNames_mask) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVuv((UV)nBIT_MASK(SvUV(ST(0)))));
    XSRETURN(1);
}
XS_EXTERNAL(boot_ApiNames) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("ApiNames::atof3", XS_ApiNames_atof3, __FILE__);
    newXS("ApiNames::calllist", XS_ApiNames_calllist, __FILE__);
    newXS("ApiNames::hscalar", XS_ApiNames_hscalar, __FILE__);
    newXS("ApiNames::pack", XS_ApiNames_pack, __FILE__);
    newXS("ApiNames::unpack", XS_ApiNames_unpack, __FILE__);
    newXS("ApiNames::utf8", XS_ApiNames_utf8, __FILE__);
    newXS("ApiNames::strict", XS_ApiNames_strict, __FILE__);
    newXS("ApiNames::names", XS_ApiNames_names, __FILE__);
    newXS("ApiNames::classes", XS_ApiNames_classes, __FILE__);
    newXS("ApiNames::mask", XS_ApiNames_mask, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `our @log;
sub a1 { push @log, "a1" }
sub a2 { push @log, "a2" }
package TH { sub TIEHASH { bless {}, shift } sub SCALAR { "tied-scalar" } sub FETCH {} sub FIRSTKEY {} }
package main;
tie my %t, 'TH';
my @r;
push @r, join ";", map { ApiNames::atof3(@$_) } ["3.25abc", 7], ["1e3", 2], ["1e3", 3], [" -12.5e-1x", 10],
  ["Inf", 3], ["nanx", 4], ["abc", 3], ["0x10", 4];
push @r, ApiNames::calllist("main::a1", "main::a2") . "[@log]";
push @r, ApiNames::hscalar({a => 1, b => 2}) . "," . ApiNames::hscalar({}) . "," . ApiNames::hscalar(\%t);
my $p = ApiNames::pack("NnC/a*", 1, 2, "xyz");
push @r, unpack("H*", $p) . "=" . join(",", ApiNames::unpack("NnC/a*", $p)) . "=" . scalar(() = ApiNames::unpack("(a1)*", "abcde"));
push @r, ApiNames::utf8("caf\x{e9}", 10) . "," . ApiNames::utf8("\x{263a}x", 3);
push @r, join ",", map { ApiNames::strict($_) } "abc", "\xed\xa0\x80", "\xef\xbf\xbe", "abc\xe2\x98", "\xf4\x90\x80\x80", "a\xffb";
push @r, ApiNames::names();
push @r, join ",", map { ApiNames::classes($_) } 0x41, 0x61, 0x09, 0x20, 0xA0, 0xE9, 0x2D, 0xFF10, 0x3B1, 0x2028, 0x7F, 0x1D7CE;
push @r, join ",", map { ApiNames::mask($_) } 0, 5, 63;
join "\n", @r`
	if got, want := requireModule(t, "ApiNames", so, program), `3.25|4;1|1;1000|3;-1.25|9;Inf|3;NaN|3;-1|-1;0|1
0[a1 a2]
2,0,tied-scalar
0000000100020378797a=1,2,xyz=5
11:4:5:1,00:4:4:1
1111:3:3,0011:0:0,0111:0:0,0001:3:3,0011:0:0,0000:1:1
-1,0,$x
--gp-xa--A/-aw--/-p-a-xU/61,l-gp-xa--A/law--/lp-a-x-/61,-c-----bsA/---s-/----s--/9,---p---bsA/---s-/-p--s--/20,---p---bs-/-----/-p--s--/a0,l-gp--a---/-----/lp-a---/e9,--gpu----A/----u/-pu----/2d,--gp-xa---/--w--/-p---x-/ff10,l-gp--a---/law--/lp-a---/3b1,--------s-/---s-/----s--/2028,-c-------A/-----/-------/7f,--gp--a---/--w--/-p-----/1d7ce
0,31,9223372036854775807`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
