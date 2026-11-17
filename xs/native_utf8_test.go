//go:build darwin || linux

package xs_test

import "testing"

// UTF-8 of native code is perl's. Decoding: well-formed characters, each
// malformation (a continuation byte alone, a short or interrupted
// sequence, an overlong one) with and without UTF8_CHECK_ONLY and
// UTF8_ALLOW_ANY, surrogates, noncharacters and code points above Unicode
// allowed, disallowed and warned about, with the code point, the length
// and the errors perl gives and its warnings ("Malformed UTF-8 character",
// which Text::Minify::XS relies on); utf8_to_uvchr_buf with the utf8
// warnings on and off, and the messages of utf8n_to_uvchr_msgs. Encoding
// under the UNICODE_ flags (warned, refused, with a message), the length of
// a character under flags, partial characters and strings under flags,
// which Encode uses. What is expected is what perl 5.36 on Linux gives with
// the same module.
func TestNativeUTF8IsPerls(t *testing.T) {
	so := buildModuleSource(t, "Utf8Dec", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* the flags by name, for they are numbered differently from one perl to
 * another */
static U32 flags_named(const char *name) {
    U32 flags = 0;
    if (strstr(name, "check")) flags |= UTF8_CHECK_ONLY;
    if (strstr(name, "any")) flags |= UTF8_ALLOW_ANY;
    if (strstr(name, "nosur")) flags |= UTF8_DISALLOW_SURROGATE;
    if (strstr(name, "warnsur")) flags |= UTF8_WARN_SURROGATE;
    if (strstr(name, "warnnon")) flags |= UTF8_WARN_NONCHAR;
    if (strstr(name, "nosuper")) flags |= UTF8_DISALLOW_SUPER;
    if (strstr(name, "illegal")) flags |= UTF8_DISALLOW_ILLEGAL_INTERCHANGE;
    return flags;
}
static void errors_named(SV *out, U32 errors) {
    if (errors & UTF8_GOT_CONTINUATION) sv_catpvs(out, " continuation");
    if (errors & UTF8_GOT_SHORT) sv_catpvs(out, " short");
    if (errors & UTF8_GOT_NON_CONTINUATION) sv_catpvs(out, " non-continuation");
    if (errors & UTF8_GOT_LONG) sv_catpvs(out, " long");
    if (errors & UTF8_GOT_SURROGATE) sv_catpvs(out, " surrogate");
    if (errors & UTF8_GOT_NONCHAR) sv_catpvs(out, " nonchar");
    if (errors & UTF8_GOT_SUPER) sv_catpvs(out, " super");
}
/* decode(bytes, flags): utf8n_to_uvchr_error of the first character, as
 * "code point/length/errors" */
XS(XS_Utf8Dec_decode) {
    dXSARGS;
    STRLEN len, used = 0;
    U32 errors = 0;
    const U8 *p = (const U8 *)SvPVbyte(ST(0), len);
    const UV cp = utf8n_to_uvchr_error(p, len, &used, flags_named(SvPV_nolen(ST(1))), &errors);
    SV *const out = sv_2mortal(newSVpvf("%lx/%ld/", (unsigned long)cp,
                                        used == (STRLEN)-1 ? -1L : (long)used));
    PERL_UNUSED_VAR(items);
    errors_named(out, errors);
    ST(0) = out;
    XSRETURN(1);
}
/* buf(bytes): utf8_to_uvchr_buf of the first character, as
 * "code point/length" */
XS(XS_Utf8Dec_buf) {
    dXSARGS;
    STRLEN len, used = 0;
    const U8 *p = (const U8 *)SvPVbyte(ST(0), len);
    const UV cp = utf8_to_uvchr_buf(p, p + len, &used);
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%lx/%ld", (unsigned long)cp, used == (STRLEN)-1 ? -1L : (long)used));
    XSRETURN(1);
}
/* msgs(bytes): utf8n_to_uvchr_msgs of the first character with the
 * messages taken, as "code point/number of messages/errors" */
XS(XS_Utf8Dec_msgs) {
    dXSARGS;
    STRLEN len, used = 0;
    U32 errors = 0;
    AV *msgs = NULL;
    const U8 *p = (const U8 *)SvPVbyte(ST(0), len);
    const UV cp = utf8n_to_uvchr_msgs(p, len, &used, 0, &errors, &msgs);
    SV *const out = sv_2mortal(newSVpvf("%lx/%ld/", (unsigned long)cp, msgs ? (long)(av_top_index(msgs) + 1) : 0L));
    PERL_UNUSED_VAR(items);
    errors_named(out, errors);
    if (msgs) SvREFCNT_dec((SV *)msgs);
    ST(0) = out;
    XSRETURN(1);
}
/* the encoder's flags by name */
static UV uflags_named(const char *name) {
    UV flags = 0;
    if (strstr(name, "warnsur")) flags |= UNICODE_WARN_SURROGATE;
    if (strstr(name, "nosur")) flags |= UNICODE_DISALLOW_SURROGATE;
    if (strstr(name, "warnnon")) flags |= UNICODE_WARN_NONCHAR;
    if (strstr(name, "nononchar")) flags |= UNICODE_DISALLOW_NONCHAR;
    if (strstr(name, "nosuper")) flags |= UNICODE_DISALLOW_SUPER;
    return flags;
}
/* encode(cp, flags, msgs): uvchr_to_utf8_flags_msgs, as the bytes in hex
 * (or NULL) and with msgs the number of messages */
XS(XS_Utf8Dec_encode) {
    dXSARGS;
    U8 buf[UTF8_MAXBYTES + 1];
    HV *msgs = NULL;
    const bool want = SvTRUE(ST(2));
    U8 *const end = uvchr_to_utf8_flags_msgs(buf, SvUV(ST(0)), uflags_named(SvPV_nolen(ST(1))),
                                             want ? &msgs : NULL);
    SV *const out = sv_2mortal(newSVpvs(""));
    PERL_UNUSED_VAR(items);
    if (!end) {
        sv_catpvs(out, "NULL");
    } else {
        for (const U8 *q = buf; q < end; q++) sv_catpvf(out, "%02x", *q);
    }
    if (want) {
        sv_catpvf(out, " msgs %d", msgs ? 1 : 0);
        if (msgs) SvREFCNT_dec((SV *)msgs);
    }
    ST(0) = out;
    XSRETURN(1);
}
/* charlen(bytes, flags): isUTF8_CHAR_flags */
XS(XS_Utf8Dec_charlen) {
    dXSARGS;
    STRLEN len;
    const U8 *p = (const U8 *)SvPVbyte(ST(0), len);
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVuv(isUTF8_CHAR_flags(p, p + len, flags_named(SvPV_nolen(ST(1))))));
    XSRETURN(1);
}
/* partial(bytes, flags): is_utf8_valid_partial_char_flags */
XS(XS_Utf8Dec_partial) {
    dXSARGS;
    STRLEN len;
    const U8 *p = (const U8 *)SvPVbyte(ST(0), len);
    PERL_UNUSED_VAR(items);
    ST(0) = is_utf8_valid_partial_char_flags(p, p + len, flags_named(SvPV_nolen(ST(1)))) ? &PL_sv_yes
                                                                                         : &PL_sv_no;
    XSRETURN(1);
}
/* strloc(bytes, flags): is_utf8_string_loclen_flags, as "valid/where it
 * stopped/characters before" */
XS(XS_Utf8Dec_strloc) {
    dXSARGS;
    STRLEN len, count = 0;
    const U8 *p = (const U8 *)SvPVbyte(ST(0), len), *ep = NULL;
    const bool ok = is_utf8_string_loclen_flags(p, len, &ep, &count, flags_named(SvPV_nolen(ST(1))));
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%d/%ld/%ld", ok ? 1 : 0, (long)(ep - p), (long)count));
    XSRETURN(1);
}
XS_EXTERNAL(boot_Utf8Dec) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("Utf8Dec::decode", XS_Utf8Dec_decode, __FILE__);
    newXS("Utf8Dec::buf", XS_Utf8Dec_buf, __FILE__);
    newXS("Utf8Dec::msgs", XS_Utf8Dec_msgs, __FILE__);
    newXS("Utf8Dec::encode", XS_Utf8Dec_encode, __FILE__);
    newXS("Utf8Dec::charlen", XS_Utf8Dec_charlen, __FILE__);
    newXS("Utf8Dec::partial", XS_Utf8Dec_partial, __FILE__);
    newXS("Utf8Dec::strloc", XS_Utf8Dec_strloc, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my @warned;
local $SIG{__WARN__} = sub { my $w = shift; $w =~ s/ at (?:\(eval \d+\)|\S+) line \d+\.\n\z//; push @warned, $w };
my @got;
for my $case (["\xc3\xa9", ""], ["\xf0\x9f\x98\x80", ""], ["\xa0", ""], ["\xa0", "check"],
              ["\xa0", "any"], ["\xe2\x82", "check"], ["\xe2\x41", "check"], ["\xc0\xaf", "check"],
              ["\xed\xa0\x80", ""], ["\xed\xa0\x80", "check nosur"], ["\xed\xa0\x80", "warnsur"],
              ["\xef\xbf\xbf", ""], ["\xef\xbf\xbf", "warnnon"], ["\xef\xbf\xbf", "check illegal"],
              ["\xf4\x90\x80\x80", ""], ["\xf4\x90\x80\x80", "check nosuper"]) {
    my ($bytes, $flags) = @$case;
    @warned = ();
    my $r = Utf8Dec::decode($bytes, $flags);
    push @got, $r . (@warned ? " [@warned]" : "");
}
@warned = ();
push @got, Utf8Dec::buf("\xa0") . (@warned ? " [@warned]" : "");
{
    no warnings 'utf8';
    @warned = ();
    push @got, Utf8Dec::buf("\xa0") . (@warned ? " [@warned]" : "");
}
@warned = ();
push @got, Utf8Dec::buf("\xe3\x81\x82") . (@warned ? " [@warned]" : "");
@warned = ();
push @got, Utf8Dec::msgs("\xa0") . (@warned ? " [@warned]" : "");
for my $case ([0x263A, "", 0], [0x1F600, "", 0], [0xD800, "", 0], [0xD800, "warnsur", 0],
              [0xD800, "nosur", 0], [0xD800, "warnsur", 1], [0xFFFF, "warnnon", 0], [0xFFFF, "nononchar", 0],
              [0x110000, "", 0], [0x110000, "nosuper", 0]) {
    @warned = ();
    push @got, Utf8Dec::encode(@$case) . (@warned ? " [@warned]" : "");
}
push @got, join ",", map { Utf8Dec::charlen(@$_) } ["\xed\xa0\x80", ""], ["\xed\xa0\x80", "illegal"],
    ["\xe2\x82", ""], ["\xf4\x90\x80\x80", ""], ["\xf4\x90\x80\x80", "illegal"], ["\xef\xbf\xbe", "illegal"];
push @got, join ",", map { Utf8Dec::partial(@$_) ? 1 : 0 } ["\xe2\x82", ""], ["\xed\xa0", "illegal"],
    ["\xed\xa0", ""], ["\xe2\x82\xac", ""], ["\xf4\x90", "illegal"], ["\xa0", ""];
push @got, join ",", map { Utf8Dec::strloc(@$_) } ["ab\xed\xa0\x80cd", "illegal"], ["ab\xed\xa0\x80cd", ""],
    ["a\xe3\x81\x82\xe2\x82", ""];
join "|", @got`
	if got, want := requireModule(t, "Utf8Dec", so, program), `e9/2/|1f600/4/|0/1/ continuation [Malformed UTF-8 character: \xa0 (unexpected continuation byte 0xa0, with no preceding start byte) in subroutine entry]|0/-1/ continuation|fffd/1/ continuation|0/-1/ short|0/-1/ short non-continuation|0/-1/ long|d800/3/|0/-1/ surrogate|d800/3/ surrogate [UTF-16 surrogate U+D800 in subroutine entry]|ffff/3/|ffff/3/ nonchar [Unicode non-character U+FFFF is not recommended for open interchange in subroutine entry]|0/-1/ nonchar|110000/4/|0/-1/ super|0/-1 [Malformed UTF-8 character: \xa0 (unexpected continuation byte 0xa0, with no preceding start byte) in subroutine entry]|fffd/1|3042/3|0/1/ continuation|e298ba|f09f9880|eda080|eda080 [UTF-16 surrogate U+D800]|NULL|eda080 msgs 1|efbfbf [Unicode non-character U+FFFF is not recommended for open interchange]|NULL|f4908080|NULL|3,0,0,4,0,0|1,0,1,0,0,0|0/2/2,1/7/5,0/4/2`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
