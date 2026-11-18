//go:build darwin || linux

package xs_test

import "testing"

// Names and sizes modules take from perl.h: the sizes of the bodies of SVs
// and of the op structs (Sort::External, Hash::SharedMem, Devel::Size), GP
// (Stats::LikeR), NV_MAX_10_EXP, HEK_BASESIZE, LvTARGOFF/LvTARGLEN
// (Devel::FindRef), perl_alloc and perl_free as perl without MULTIPLICITY
// has them (DBD::Firebird's event callback), sv_mortalcopy_flags
// (Stats::LikeR) and save_sptr (List::Keywords). Before, none of them
// compiled. The expected strings are what perl prints with the same module
// code.
const perlhNamesSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

XS(XS_MyNames_sizes) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("sv+xpv=%d pv-body=%d xpvmg=%d xpvnv=%d xpvav=%d xpvhv=%d unop=%d "
                                "binop=%d logop=%d listop=%d svop=%d loop=%d hek=%d",
                                (int)(sizeof(SV) + sizeof(struct xpv)),
                                (int)(sizeof(XPV) - STRUCT_OFFSET(XPV, xpv_cur)), (int)sizeof(XPVMG),
                                (int)sizeof(XPVNV), (int)sizeof(XPVAV), (int)sizeof(XPVHV),
                                (int)sizeof(struct unop), (int)sizeof(struct binop),
                                (int)sizeof(struct logop), (int)sizeof(struct listop),
                                (int)sizeof(struct svop), (int)sizeof(struct loop),
                                (int)HEK_BASESIZE));
    XSRETURN(1);
}

XS(XS_MyNames_limits) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%d %d", (int)NV_MAX_10_EXP, (int)NV_MIN_10_EXP));
    XSRETURN(1);
}

/* whether two globs share their GP, kept as perl's type */
XS(XS_MyNames_gp_same) {
    dXSARGS;
    GP *a, *b;
    if (items != 2) croak_xs_usage(cv, "a, b");
    a = GvGP((GV *)SvRV(ST(0)));
    b = GvGP((GV *)SvRV(ST(1)));
    ST(0) = sv_2mortal(newSVpv(a == b ? "same" : "other", 0));
    XSRETURN(1);
}

/* the type, the offset and the length of an lvalue scalar's target */
XS(XS_MyNames_lv) {
    dXSARGS;
    SV *lv;
    if (items != 1) croak_xs_usage(cv, "ref");
    lv = SvRV(ST(0));
    ST(0) = sv_2mortal(newSVpvf("%c %d %d", LvTYPE(lv), (int)LvTARGOFF(lv), (int)LvTARGLEN(lv)));
    XSRETURN(1);
}

/* an interpreter allocated, made the context, and freed after the context
 * is put back, as DBD::Firebird's event callback does */
XS(XS_MyNames_alloc_free) {
    dXSARGS;
    void *const context = PERL_GET_CONTEXT;
    PerlInterpreter *i;
    PERL_UNUSED_VAR(items);
    i = perl_alloc();
    if (!i) croak("no interpreter allocated");
    PERL_SET_CONTEXT(i);
    PERL_SET_CONTEXT(context);
    perl_free(i);
    ST(0) = sv_2mortal(newSVpvs("freed"));
    XSRETURN(1);
}

/* mortal copies of an argument, without and with its get magic run, as
 * Stats::LikeR copies the tied arguments of a call */
XS(XS_MyNames_copies) {
    dXSARGS;
    SV *plain, *got;
    if (items != 1) croak_xs_usage(cv, "value");
    plain = sv_mortalcopy_flags(ST(0), SV_NOSTEAL);
    got = sv_mortalcopy_flags(ST(0), SV_GMAGIC | SV_NOSTEAL);
    ST(0) = sv_2mortal(newSVpvf("%s/%s temp=%d", SvOK(plain) ? SvPV_nolen(plain) : "undef",
                                SvPV_nolen(got), SvTEMP(got) ? 1 : 0));
    XSRETURN(1);
}

/* $MyNames::x replaced for a call of MyNames::peek inside a scope of its
 * own, the slot saved with save_sptr */
XS(XS_MyNames_inner) {
    dXSARGS;
    GV *const gv = gv_fetchpvs("MyNames::x", GV_ADD, SVt_PV);
    SV *inner;
    PERL_UNUSED_VAR(items);
    ENTER;
    save_sptr(&GvSV(gv));
    GvSV(gv) = sv_2mortal(newSVpvs("inner"));
    PUSHMARK(SP);
    PUTBACK;
    call_pv("MyNames::peek", G_SCALAR);
    SPAGAIN;
    inner = newSVsv(POPs);
    PUTBACK;
    LEAVE;
    ST(0) = sv_2mortal(newSVpvf("%s then %s", SvPV_nolen(inner), SvPV_nolen(GvSV(gv))));
    SvREFCNT_dec(inner);
    XSRETURN(1);
}

XS_EXTERNAL(boot_MyNames) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyNames::sizes", XS_MyNames_sizes);
    newXS_deffile("MyNames::limits", XS_MyNames_limits);
    newXS_deffile("MyNames::gp_same", XS_MyNames_gp_same);
    newXS_deffile("MyNames::lv", XS_MyNames_lv);
    newXS_deffile("MyNames::alloc_free", XS_MyNames_alloc_free);
    newXS_deffile("MyNames::copies", XS_MyNames_copies);
    newXS_deffile("MyNames::inner", XS_MyNames_inner);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeNamesAndSizesOfPerlH(t *testing.T) {
	so := buildModuleSource(t, "MyNames", perlhNamesSource)
	got := requireModule(t, "MyNames", so, `our ($ga, $gb, $gc); *ga = *gb;
my $s = "abcdef"; my $r = \substr($s, 2, 3);
{ package MyNames::Tie; sub TIESCALAR { my $n = 0; bless \$n } sub FETCH { ++${$_[0]} } }
tie my $t, "MyNames::Tie";
$MyNames::x = "outer";
sub MyNames::peek { $MyNames::x }
join "|", MyNames::sizes(), MyNames::limits(), MyNames::gp_same(\*ga, \*gb),
    MyNames::gp_same(\*ga, \*gc), MyNames::lv($r), MyNames::alloc_free(), MyNames::copies($t),
    MyNames::inner()`)
	want := "sv+xpv=56 pv-body=16 xpvmg=48 xpvnv=48 xpvav=40 xpvhv=32 unop=48 binop=56 logop=56 " +
		"listop=56 svop=48 loop=80 hek=8|308 -307|same|other|x 2 3|freed|undef/1 temp=1|inner then outer"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
