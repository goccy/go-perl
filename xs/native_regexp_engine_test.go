//go:build darwin || linux

package xs_test

import "testing"

// A regexp engine of native code (re::engine::RE2, re::engine::GNU) is
// perl's regexp_engine: ENGINE gives PTR2IV of it for $^H{regcomp}, perl
// calls its comp and exec, and they fill the regexp's body (the wrapped
// pattern, the private pointer, the groups' offsets and names, the string
// matched, pointed at or copied). The functions it takes from perl
// (Perl_reg_numbered_buff_fetch and the rest) are perl's, and native code
// may call them too. Before, regexp_engine had no members and an engine
// did not compile. The expected string is what perl prints with the same
// module code.
const regexpEngineSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <string.h>

/* Two engines that find the pattern as a literal string: MyEng::A points
 * the regexp at the string matched, as re::engine::RE2 does; MyEng::B
 * copies it, as re::engine::GNU does, so that $& outlives a change of the
 * string. A pattern that starts with "(" is perl's to compile, as
 * re::engine::RE2 leaves one it cannot compile. */
typedef struct {
    STRLEN len;
    char pat[1];
} MyEng_pat;

static REGEXP *MyEng_comp_a(pTHX_ SV *const pattern, U32 flags);
static REGEXP *MyEng_comp_b(pTHX_ SV *const pattern, U32 flags);
static I32 MyEng_exec(pTHX_ REGEXP *const rx_sv, char *stringarg, char *strend, char *strbeg,
                      SSize_t minend, SV *sv, void *data, U32 flags);
static char *MyEng_intuit(pTHX_ REGEXP *const rx, SV *sv, const char *const strbeg, char *strpos,
                          char *strend, const U32 flags, re_scream_pos_data *data) {
    PERL_UNUSED_ARG(rx); PERL_UNUSED_ARG(sv); PERL_UNUSED_ARG(strbeg); PERL_UNUSED_ARG(strpos);
    PERL_UNUSED_ARG(strend); PERL_UNUSED_ARG(flags); PERL_UNUSED_ARG(data);
    return NULL;
}
static SV *MyEng_checkstr(pTHX_ REGEXP *const rx) {
    PERL_UNUSED_ARG(rx);
    return NULL;
}
static void MyEng_free(pTHX_ REGEXP *const rx_sv) {
    regexp *const rx = (regexp *)SvANY(rx_sv);
    Safefree(rx->pprivate);
}
static SV *MyEng_package_a(pTHX_ REGEXP *const rx) {
    PERL_UNUSED_ARG(rx);
    return newSVpvs("MyEng::A");
}
static SV *MyEng_package_b(pTHX_ REGEXP *const rx) {
    PERL_UNUSED_ARG(rx);
    return newSVpvs("MyEng::B");
}
static regexp_engine MyEng_a = {
    .comp = MyEng_comp_a, .exec = MyEng_exec, .intuit = MyEng_intuit,
    .checkstr = MyEng_checkstr, .rxfree = MyEng_free,
    .numbered_buff_FETCH = Perl_reg_numbered_buff_fetch,
    .numbered_buff_STORE = Perl_reg_numbered_buff_store,
    .numbered_buff_LENGTH = Perl_reg_numbered_buff_length,
    .named_buff = Perl_reg_named_buff, .named_buff_iter = Perl_reg_named_buff_iter,
    .qr_package = MyEng_package_a,
};
static regexp_engine MyEng_b = {
    .comp = MyEng_comp_b, .exec = MyEng_exec, .intuit = MyEng_intuit,
    .checkstr = MyEng_checkstr, .rxfree = MyEng_free,
    .numbered_buff_FETCH = Perl_reg_numbered_buff_fetch,
    .numbered_buff_STORE = Perl_reg_numbered_buff_store,
    .numbered_buff_LENGTH = Perl_reg_numbered_buff_length,
    .named_buff = Perl_reg_named_buff, .named_buff_iter = Perl_reg_named_buff_iter,
    .qr_package = MyEng_package_b,
};

static REGEXP *MyEng_comp(pTHX_ SV *const pattern, U32 flags, const regexp_engine *eng) {
    STRLEN plen;
    const char *exp = SvPV(pattern, plen);
    REGEXP *rx_sv;
    regexp *rx;
    MyEng_pat *ri;
    SV *wrapped, *value;
    I32 group = 1;
    if (plen == 0 || exp[0] == '(') return Perl_re_compile(aTHX_ pattern, flags);
    rx_sv = (REGEXP *)newSV_type(SVt_REGEXP);
    rx = (regexp *)SvANY(rx_sv);
    Zero(rx, 1, regexp);
    rx->extflags = flags;
    rx->engine = eng;
    wrapped = newSVpvs("(?^:");
    sv_catpvn(wrapped, exp, plen);
    sv_catpvs(wrapped, ")");
    rx->pre_prefix = 4;
    RX_WRAPPED(rx_sv) = savepvn(SvPVX(wrapped), SvCUR(wrapped));
    RX_WRAPLEN(rx_sv) = SvCUR(wrapped);
    SvREFCNT_dec(wrapped);
    ri = (MyEng_pat *)safemalloc(sizeof(MyEng_pat) + plen);
    ri->len = plen;
    memcpy(ri->pat, exp, plen);
    rx->pprivate = ri;
    /* one group, the whole match, named "all" */
    value = newSV_type(SVt_PVNV);
    sv_setpvn(value, (char *)&group, sizeof group);
    SvIV_set(value, 1);
    rx->paren_names = newHV();
    (void)hv_stores(rx->paren_names, "all", value);
    rx->nparens = 1;
#if PERL_VERSION >= 38
    rx->logical_nparens = 1;
#endif
    Newxz(rx->offs, 2, regexp_paren_pair);
    return rx_sv;
}
static REGEXP *MyEng_comp_a(pTHX_ SV *const pattern, U32 flags) {
    return MyEng_comp(aTHX_ pattern, flags, &MyEng_a);
}
static REGEXP *MyEng_comp_b(pTHX_ SV *const pattern, U32 flags) {
    return MyEng_comp(aTHX_ pattern, flags, &MyEng_b);
}
static I32 MyEng_exec(pTHX_ REGEXP *const rx_sv, char *stringarg, char *strend, char *strbeg,
                      SSize_t minend, SV *sv, void *data, U32 flags) {
    regexp *const rx = (regexp *)SvANY(rx_sv);
    const MyEng_pat *const ri = (const MyEng_pat *)rx->pprivate;
    char *at;
    PERL_UNUSED_ARG(minend); PERL_UNUSED_ARG(sv); PERL_UNUSED_ARG(data); PERL_UNUSED_ARG(flags);
    for (at = stringarg; at + ri->len <= strend; at++)
        if (memcmp(at, ri->pat, ri->len) == 0) break;
    if (at + ri->len > strend) return 0;
    rx->offs[0].start = rx->offs[1].start = at - strbeg;
    rx->offs[0].end = rx->offs[1].end = at - strbeg + ri->len;
    rx->lastparen = rx->lastcloseparen = 1;
    if (rx->engine == &MyEng_b) {
        RXp_MATCH_COPY_FREE(rx);
        rx->subbeg = (char *)safemalloc(strend - strbeg + 1);
        Copy(strbeg, rx->subbeg, strend - strbeg, char);
        rx->subbeg[strend - strbeg] = 0;
        RXp_MATCH_COPIED_on(rx);
    } else {
        rx->subbeg = strbeg;
    }
    rx->sublen = strend - strbeg;
    rx->suboffset = rx->subcoffset = 0;
    return 1;
}

XS(XS_MyEng_engine_a) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(PTR2IV(&MyEng_a)));
    XSRETURN(1);
}
XS(XS_MyEng_engine_b) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(PTR2IV(&MyEng_b)));
    XSRETURN(1);
}
/* which engine a regexp has, as native code compares them */
XS(XS_MyEng_engine_name) {
    dXSARGS;
    const regexp_engine *e;
    if (items != 1) croak_xs_usage(cv, "qr");
    e = RX_ENGINE(SvRX(ST(0)));
    ST(0) = sv_2mortal(newSVpv(e == &MyEng_a ? "a" : e == &MyEng_b ? "b" : "other", 0));
    XSRETURN(1);
}
XS(XS_MyEng_engine_number) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "qr");
    ST(0) = sv_2mortal(newSViv(PTR2IV(RX_ENGINE(SvRX(ST(0))))));
    XSRETURN(1);
}
XS(XS_MyEng_round_trip) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(INT2PTR(regexp_engine *, PTR2IV(&MyEng_a)) == &MyEng_a));
    XSRETURN(1);
}
/* a match through the engine (CALLREGEXEC) and a group of it by perl's own
 * function, called by native code */
XS(XS_MyEng_match_group) {
    dXSARGS;
    REGEXP *rx;
    STRLEN len;
    char *s;
    SV *out;
    if (items != 3) croak_xs_usage(cv, "qr, string, n");
    rx = SvRX(ST(0));
    s = SvPV(ST(1), len);
    if (CALLREGEXEC(rx, s, s + len, s, 0, ST(1), NULL, 0) <= 0) XSRETURN_UNDEF;
    out = sv_newmortal();
    Perl_reg_numbered_buff_fetch(aTHX_ rx, (I32)SvIV(ST(2)), out);
    ST(0) = out;
    XSRETURN(1);
}

XS_EXTERNAL(boot_MyEng) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyEng::engine_a", XS_MyEng_engine_a);
    newXS_deffile("MyEng::engine_b", XS_MyEng_engine_b);
    newXS_deffile("MyEng::engine_name", XS_MyEng_engine_name);
    newXS_deffile("MyEng::engine_number", XS_MyEng_engine_number);
    newXS_deffile("MyEng::round_trip", XS_MyEng_round_trip);
    newXS_deffile("MyEng::match_group", XS_MyEng_match_group);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeRegexpEngine(t *testing.T) {
	so := buildModuleSource(t, "MyEng", regexpEngineSource)
	got := requireModule(t, "MyEng", so, `my $r = eval q{
my @r;
{
    BEGIN { $^H{regcomp} = MyEng::engine_a() }
    my $re = qr/b.c/;
    push @r, ref($re), "$re", MyEng::engine_name($re);
    push @r, ("xxab.cyy" =~ $re) ? "match" : "none";
    push @r, "$&,$1,$+{all},$-[0],$+[0],$#+";
    push @r, ("xxabzcyy" =~ /b.c/) ? "match" : "none";
    push @r, MyEng::match_group($re, "--b.c--", 1);
    push @r, MyEng::engine_number($re) == MyEng::engine_a() ? "same" : "other";
    push @r, ref(qr/(\d)/), ("a42" =~ /(\d+)/ ? $1 : "none");
}
{
    BEGIN { $^H{regcomp} = MyEng::engine_b() }
    my $s = "hello world";
    if ($s =~ /o w/) { $s = "changed"; push @r, $& }
    push @r, ref(qr/x/), join(",", map { /a/ ? "y" : "n" } "bab", "bbb");
}
push @r, ("xxabzcyy" =~ /b.c/) ? "core" : "none";
push @r, MyEng::engine_name(qr/x/), MyEng::round_trip();
join "|", @r
}; die $@ if $@; $r`)
	want := "MyEng::A|(?^:b.c)|a|match|b.c,b.c,b.c,3,6,1|none|b.c|same|Regexp|42|o w|MyEng::B|y,n|core|other|1"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
