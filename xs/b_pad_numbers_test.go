//go:build darwin || linux

package xs_test

import "testing"

// The pad structures of B objects, which native code takes as perl's
// typemaps take them: a B::PADLIST and a B::PADNAMELIST by INT2PTR of their
// number, and a B::PADNAME through the scalar pointer B::SV's typemap makes
// of it, as B::Flags does. B tags these numbers as it tags an op's, so the
// SDK tells them from its own addresses (B-Flags died of SIGSEGV reading
// PadnameFLAGS of a B::PADNAME). What is expected is what perl 5.36 on
// Linux gives with the same module.
func TestNativeCodeTakesPadStructuresOfBObjects(t *testing.T) {
	so := buildModuleSource(t, "BPad", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* Pad structures of B objects, taken as modules take them: the number of a
 * B::PADLIST or B::PADNAMELIST object as perl's typemaps take it (INT2PTR of
 * SvIV(SvRV(obj))), and that of a B::PADNAME object through a scalar
 * pointer, as B::Flags's typemap for B::SV gives it. */
static IV num(pTHX_ SV *obj) {
    if (!SvROK(obj)) croak("not a B object");
    return SvIV(SvRV(obj));
}
static void name_into(pTHX_ SV *out, PADNAME *pn) {
    if (PadnamePV(pn)) sv_catpvn(out, PadnamePV(pn), PadnameLEN(pn));
    else sv_catpvs(out, "-");
    if (PadnameOUTER(pn)) sv_catpvs(out, ":OUTER");
    if (PadnameIsSTATE(pn)) sv_catpvs(out, ":STATE");
    if (SvPAD_OUR(pn)) sv_catpvs(out, ":OUR");
    if (SvPAD_TYPED(pn)) sv_catpvs(out, ":TYPED");
    sv_catpvf(out, ":%d", (int)PadnameFLAGS(pn) & (PADNAMEt_OUTER | PADNAMEt_STATE));
}
XS(XS_BPad_padlist_max) {
    dXSARGS;
    PADLIST *pl;
    if (items != 1) croak_xs_usage(cv, "padlist");
    pl = INT2PTR(PADLIST *, num(aTHX_ ST(0)));
    ST(0) = sv_2mortal(newSViv((IV)PadlistMAX(pl)));
    XSRETURN(1);
}
XS(XS_BPad_names_max) {
    dXSARGS;
    PADNAMELIST *pnl;
    if (items != 1) croak_xs_usage(cv, "padnamelist");
    pnl = INT2PTR(PADNAMELIST *, num(aTHX_ ST(0)));
    ST(0) = sv_2mortal(newSViv((IV)PadnamelistMAX(pnl)));
    XSRETURN(1);
}
/* the names of a B::PADNAMELIST, from its array */
XS(XS_BPad_list_names) {
    dXSARGS;
    PADNAMELIST *pnl;
    SV *out = sv_2mortal(newSVpvs(""));
    SSize_t i;
    if (items != 1) croak_xs_usage(cv, "padnamelist");
    pnl = INT2PTR(PADNAMELIST *, num(aTHX_ ST(0)));
    for (i = 1; i <= PadnamelistMAX(pnl); i++) {
        PADNAME *const pn = PadnamelistARRAY(pnl)[i];
        /* the names of the program's variables */
        if (!pn || !PadnamePV(pn) || PadnameLEN(pn) < 2) continue;
        if (SvCUR(out)) sv_catpvs(out, ",");
        name_into(aTHX_ out, pn);
    }
    ST(0) = out;
    XSRETURN(1);
}
/* a B::PADNAME, as B::Flags reads one */
XS(XS_BPad_name) {
    dXSARGS;
    SV *sv;
    SV *out = sv_2mortal(newSVpvs(""));
    if (items != 1) croak_xs_usage(cv, "padname");
    sv = INT2PTR(SV *, num(aTHX_ ST(0)));
    name_into(aTHX_ out, (PADNAME *)sv);
    ST(0) = out;
    XSRETURN(1);
}
XS_EXTERNAL(boot_BPad) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("BPad::padlist_max", XS_BPad_padlist_max, __FILE__);
    newXS("BPad::names_max", XS_BPad_names_max, __FILE__);
    newXS("BPad::list_names", XS_BPad_list_names, __FILE__);
    newXS("BPad::name", XS_BPad_name, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `use B;
use feature 'state';
package Some::Class;
package main;
my $outer = 1;
sub f { my $x; our $o; state $s; my Some::Class $t; my ($y, $z) = @_; $outer++; $x + $y + $z + $s }
my $cv = B::svref_2object(\&f);
my $pl = $cv->PADLIST;
my ($pnl) = $pl->ARRAY;
my @r;
push @r, BPad::padlist_max($pl) == $pl->MAX ? "padlist" : "padlist differs";
push @r, BPad::names_max($pnl) == $pnl->MAX ? "list" : "list differs";
my @names = $pnl->ARRAY;
push @r, join ",", map { BPad::name($_) } grep { ref $_ eq "B::PADNAME" && length($_->PV // "") > 1 } @names[1 .. $#names];
push @r, BPad::list_names($pnl);
join " ", @r`
	if got, want := requireModule(t, "BPad", so, program), "padlist list $x:0,$o:OUR:0,$s:STATE:2,$t:TYPED:0,$y:0,$z:0,$outer:OUTER:1 $x:0,$o:OUR:0,$s:STATE:2,$t:TYPED:0,$y:0,$z:0,$outer:OUTER:1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
