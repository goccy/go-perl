//go:build darwin || linux

package xs_test

import "testing"

// What Devel::Size reads to reckon the memory of values: AvALLOC, where an
// array's allocation starts (a shift moves AvARRAY on and leaves AvALLOC),
// the static pad names PL_padname_undef and PL_padname_const that entries of
// a list of names are, and the body of a REGEXP through SvANY, typed as
// perl's struct regexp. Before, none of them compiled. The expected string
// is what perl prints with the same module code.
const sizeViewsSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

/* how many slots a shift left before AvARRAY, and the first value there */
XS(XS_MySzv_gap) {
    dXSARGS;
    AV *av;
    SV **alloc;
    if (items != 1) croak_xs_usage(cv, "aref");
    av = (AV *)SvRV(ST(0));
    alloc = AvALLOC(av);
    ST(0) = sv_2mortal(newSVpvf("%d %s", (int)(AvARRAY(av) - alloc),
                                alloc && AvARRAY(av)[0] ? SvPV_nolen(AvARRAY(av)[0]) : "none"));
    XSRETURN(1);
}

/* the names of a sub's pad: u for PL_padname_undef, c for PL_padname_const,
 * the name of a named one */
XS(XS_MySzv_names) {
    dXSARGS;
    CV *c;
    PADNAMELIST *pnl;
    SV *out;
    SSize_t i;
    if (items != 1) croak_xs_usage(cv, "code");
    c = (CV *)SvRV(ST(0));
    pnl = PadlistNAMES(CvPADLIST(c));
    out = sv_2mortal(newSVpvs(""));
    for (i = 0; i <= PadnamelistMAX(pnl); i++) {
        PADNAME *pn = PadnamelistARRAY(pnl)[i];
        if (i) sv_catpvs(out, " ");
        if (pn == &PL_padname_undef) sv_catpvs(out, "u");
        else if (pn == &PL_padname_const) sv_catpvs(out, "c");
        else if (!pn) sv_catpvs(out, "-");
        else if (PadnamePV(pn)) sv_catpv(out, PadnamePV(pn));
        else sv_catpvs(out, "?");
    }
    sv_catpvf(out, "|%s", &PL_padname_undef == &PL_padname_const ? "same" : "distinct");
    ST(0) = out;
    XSRETURN(1);
}

/* the groups of a pattern, read from its body as Devel::Size does */
XS(XS_MySzv_parens) {
    dXSARGS;
    REGEXP *rx;
    regexp *body;
    if (items != 1) croak_xs_usage(cv, "qr");
    rx = SvRX(ST(0));
    if (!rx) croak("not a pattern");
    body = SvANY(rx);
    ST(0) = sv_2mortal(newSVpvf("%d %d %d", (int)body->nparens, (int)SvANY(rx)->nparens,
                                (int)RX_NPARENS(rx)));
    XSRETURN(1);
}

XS_EXTERNAL(boot_MySzv) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySzv::gap", XS_MySzv_gap);
    newXS_deffile("MySzv::names", XS_MySzv_names);
    newXS_deffile("MySzv::parens", XS_MySzv_parens);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeViewsDevelSizeReads(t *testing.T) {
	so := buildModuleSource(t, "MySzv", sizeViewsSource)
	got := requireModule(t, "MySzv", so, `my @a = map { "v$_" } 1 .. 5;
shift @a; shift @a;
my @b = (1, 2);
my $f = sub { my $x; $x };
join "|", MySzv::gap(\@a), MySzv::gap(\@b), MySzv::names($f),
    MySzv::parens(qr/(a)(b)?c/), MySzv::parens(qr/abc/)`)
	want := "2 v3|0 1|u $x|distinct|2 2 2|0 0 0"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
