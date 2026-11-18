//go:build darwin || linux

package xs_test

import "testing"

// The internals Data::Alias's aliasing pp functions use, each as perl has
// it: a hash's placeholder count as an lvalue; an lvalue scalar's target
// offset and length as fields of its body, written through their address;
// av_reify of a sub's @_; vivify_defelem of the lvalue scalar of a missing
// hash element; save_delete, which deletes a key when the scope ends; marks
// pushed by hand up to the end of the mark stack, which markstack_grow
// extends; and DIE in a function that returns an op. Before, none of them
// was in the SDK (an XSUB that used one did not load).
const aliasInternalsSource = `/* av_reify and vivify_defelem are perl's to its core and its extensions
 * (Data::Alias defines PERL_CORE around perl.h) */
#define PERL_EXT
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* ph(\\%h): the placeholder count, the count one up, and down again */
XS(XS_MyAi_ph) {
    dXSARGS;
    HV *const hv = (HV *)SvRV(ST(0));
    const I32 a = HvPLACEHOLDERS(hv);
    I32 b;
    PERL_UNUSED_VAR(items);
    HvPLACEHOLDERS(hv)++;
    b = HvPLACEHOLDERS_get(hv);
    HvPLACEHOLDERS(hv)--;
    ST(0) = sv_2mortal(newSVpvf("%d %d %d", (int)a, (int)b, (int)HvPLACEHOLDERS(hv)));
    XSRETURN(1);
}
/* lv(substr): the lvalue's type, offset and length, then its value with the
 * length made 1 through the field's address */
XS(XS_MyAi_lv) {
    dXSARGS;
    SV *const sv = ST(0);
    SV *r;
    PERL_UNUSED_VAR(items);
    if (SvTYPE(sv) != SVt_PVLV) croak("no lvalue");
    r = newSVpvf("%c %d %d", LvTYPE(sv), (int)LvTARGOFF(sv), (int)LvTARGLEN(sv));
    {
        STRLEN *const lenp = &LvTARGLEN(sv);
        *lenp = 1;
    }
    sv_catpvf(r, " %s", SvPV_nolen(sv));
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
/* reify(): AvREAL and AvREIFY of the calling sub's @_ (GvAV(PL_defgv); a
 * reference to it would make it real), before and after av_reify */
XS(XS_MyAi_reify) {
    dXSARGS;
    AV *const av = GvAV(PL_defgv);
    const int real = !!AvREAL(av), reify = !!AvREIFY(av);
    PERL_UNUSED_VAR(items);
    if (!AvREAL(av) && AvREIFY(av)) av_reify(av);
    ST(0) = sv_2mortal(newSVpvf("%d%d %d%d", real, reify, !!AvREAL(av), !!AvREIFY(av)));
    XSRETURN(1);
}
/* viv($h{missing}) through a sub's @_: the lvalue's type and whether it
 * was made yet, then vivify_defelem, which makes the element */
XS(XS_MyAi_viv) {
    dXSARGS;
    SV *const sv = ST(0);
    SV *r;
    PERL_UNUSED_VAR(items);
    if (SvTYPE(sv) != SVt_PVLV) croak("no lvalue");
    r = newSVpvf("%c %d", LvTYPE(sv), LvTARGLEN(sv) ? 1 : 0);
    if (LvTYPE(sv) == 'y' && LvTARGLEN(sv)) vivify_defelem(sv);
    sv_catpvf(r, " %d", LvTARG(sv) ? 1 : 0);
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
/* sdel(\\%h, key, code): the key is deleted when the call's scope ends;
 * code runs before */
XS(XS_MyAi_sdel) {
    dXSARGS;
    HV *const hv = (HV *)SvRV(ST(0));
    STRLEN len;
    const char *const key = SvPV(ST(1), len);
    PERL_UNUSED_VAR(items);
    save_delete(hv, savepvn(key, len), (I32)len);
    PUSHMARK(SP);
    call_sv(ST(2), G_VOID | G_DISCARD);
    XSRETURN_EMPTY;
}
/* marks(n): n marks pushed by hand, the mark stack grown at its end; the
 * sum of what they hold, read back, and the pointer put back to the mark
 * that was on top */
XS(XS_MyAi_marks) {
    dXSARGS;
    const I32 n = (I32)SvIV(ST(0));
    const I32 top = TOPMARK;
    I32 i;
    IV sum = 0;
    PERL_UNUSED_VAR(items);
    for (i = 1; i <= n; i++) {
        if (++PL_markstack_ptr == PL_markstack_max) markstack_grow();
        *PL_markstack_ptr = i;
    }
    for (i = 0; i < n; i++) sum += PL_markstack_ptr[-i];
    PL_markstack_ptr -= n;
    ST(0) = sv_2mortal(newSVpvf("%" IVdf " %d", sum, TOPMARK == top ? 1 : 0));
    XSRETURN(1);
}
static OP *my_dier(pTHX_ int n) {
    if (n) DIE(aTHX_ "died %d", n);
    return NULL;
}
/* die_op(n): a function that returns an op dies with DIE */
XS(XS_MyAi_die_op) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    (void)my_dier(aTHX_ (int)SvIV(ST(0)));
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MyAi) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyAi::ph", XS_MyAi_ph);
    newXS_deffile("MyAi::lv", XS_MyAi_lv);
    newXS_deffile("MyAi::reify", XS_MyAi_reify);
    newXS_deffile("MyAi::viv", XS_MyAi_viv);
    newXS_deffile("MyAi::sdel", XS_MyAi_sdel);
    newXS_deffile("MyAi::marks", XS_MyAi_marks);
    newXS_deffile("MyAi::die_op", XS_MyAi_die_op);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeUsesAliasInternals(t *testing.T) {
	so := buildModuleSource(t, "MyAi", aliasInternalsSource)
	got := requireModule(t, "MyAi", so, `use Hash::Util qw(lock_keys);
my %h = (a => 1, b => 2);
lock_keys(%h);
delete $h{a};
my @r = (MyAi::ph(\%h));
my $s = "abcdef";
sub lv { MyAi::lv($_[0]) }
push @r, lv(substr($s, 2, 3));
sub reify { MyAi::reify() }
push @r, reify(1, 2);
my %v;
sub viv { MyAi::viv($_[0]) }
push @r, viv($v{k}), exists $v{k} ? "made" : "missing";
my %d = (k => 1);
MyAi::sdel(\%d, "k", sub { push @r, exists $d{k} ? "kept" : "gone" });
push @r, exists $d{k} ? "kept" : "gone";
push @r, MyAi::marks(1000);
push @r, eval { MyAi::die_op(7); 1 } ? "lived" : $@ =~ /^died 7/ ? "died" : $@;
$@ = "";
join "|", @r`)
	if want := "1 2 1|x 2 3 c|01 10|y 1 1|made|kept|gone|500500 1|died"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
