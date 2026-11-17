//go:build darwin || linux

package xs_test

import "testing"

// liveSlotsModule uses the SV** of av_fetch, hv_fetch and av_store as perl
// code does: it swaps elements through them (Algorithm::FEC), replaces a
// hash's entry through one (Crypt::Sodium::XS's clone), reads one after an
// av_store of the same element (Language::Befunge's Vec::XS), keeps an
// lvalue fetch's slot and writes it after other calls (XML::Hash::XS), and
// localizes an entry with save_svref (Tk).
const liveSlotsModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_LiveSlots_swap) {
    dXSARGS;
    if (items != 3) croak_xs_usage(cv, "av, i, j");
    {
        AV *av = (AV *)SvRV(ST(0));
        SV **a = av_fetch(av, SvIV(ST(1)), 0);
        SV **b = av_fetch(av, SvIV(ST(2)), 0);
        SV *t;
        if (!a || !b) croak("no such element");
        t = *a; *a = *b; *b = t;
        XSRETURN_EMPTY;
    }
}
XS(XS_LiveSlots_replace) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "hv, value");
    {
        HV *hv = (HV *)SvRV(ST(0));
        SV **svp = hv_fetchs(hv, "state", 0);
        if (!svp) croak("no state");
        SvREFCNT_dec(*svp);
        *svp = newSVsv(ST(1));
        XSRETURN_EMPTY;
    }
}
XS(XS_LiveSlots_store_then_read) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "av");
    {
        AV *av = (AV *)SvRV(ST(0));
        SV **svp = av_fetch(av, 0, 1);
        av_store(av, 0, newSVpvs("stored"));
        ST(0) = sv_2mortal(newSVsv(*svp));
        XSRETURN(1);
    }
}
XS(XS_LiveSlots_kept_lvalue) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "hv");
    {
        HV *hv = (HV *)SvRV(ST(0));
        SV **a = hv_fetchs(hv, "a", 1);
        SV **b = hv_fetchs(hv, "b", 1);
        (void)hv_stores(hv, "c", newSViv(3));
        sv_setiv(*a, 1);
        SvREFCNT_dec(*b);
        *b = newSViv(2);
        XSRETURN_EMPTY;
    }
}
XS(XS_LiveSlots_localize) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "hv, code");
    {
        HV *hv = (HV *)SvRV(ST(0));
        SV **x;
        SV *sv;
        ENTER;
        x = hv_fetchs(hv, "var", 1);
        sv = save_svref(x);
        sv_setpvs(sv, "temporary");
        PUSHMARK(SP);
        PUTBACK;
        call_sv(ST(1), G_DISCARD);
        LEAVE;
        XSRETURN_EMPTY;
    }
}
XS(XS_LiveSlots_tied_fetch) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "av");
    {
        SV **svp = av_fetch((AV *)SvRV(ST(0)), 1, 0);
        if (!svp) croak("no element");
        SvGETMAGIC(*svp);
        ST(0) = sv_2mortal(newSVsv(*svp));
        XSRETURN(1);
    }
}
XS_EXTERNAL(boot_LiveSlots) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("LiveSlots::swap", XS_LiveSlots_swap);
    newXS_deffile("LiveSlots::replace", XS_LiveSlots_replace);
    newXS_deffile("LiveSlots::store_then_read", XS_LiveSlots_store_then_read);
    newXS_deffile("LiveSlots::kept_lvalue", XS_LiveSlots_kept_lvalue);
    newXS_deffile("LiveSlots::localize", XS_LiveSlots_localize);
    newXS_deffile("LiveSlots::tied_fetch", XS_LiveSlots_tied_fetch);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// The SV** av_fetch, hv_fetch and av_store give is the slot of the array or
// the hash, as on perl: what native code stores through it is the
// container's, what it reads through it is what the container holds now,
// and save_svref saves it where the interpreter keeps it. Before, it was a
// copy of the element's address: a swap or a replacement through it never
// reached the container, a read after av_store saw the old (freed) element,
// and save_svref died ("the slot is this process's copy of a value"), which
// failed Tk's test harness and nearly every Tk dependent (Tk/Derived.pm).
func TestNativeFetchSlotsAreTheContainers(t *testing.T) {
	so := buildModuleSource(t, "LiveSlots", liveSlotsModule)
	for _, tc := range []struct{ name, code, want string }{
		{"a swap through two slots", `my @a = (1, 2, 3); LiveSlots::swap(\@a, 0, 2); "@a"`, "3 2 1"},
		{"an entry replaced through its slot", `my %h = (state => "old"); LiveSlots::replace(\%h, "new"); $h{state}`, "new"},
		{"a slot read after av_store", `my @a = ("before"); LiveSlots::store_then_read(\@a) . "," . $a[0]`, "stored,stored"},
		{"lvalue slots kept and written later", `my %h; LiveSlots::kept_lvalue(\%h); join ",", map { "$_=$h{$_}" } sort keys %h`, "a=1,b=2,c=3"},
		{"save_svref of an entry", `my %h = (var => "original"); my $in; LiveSlots::localize(\%h, sub { $in = $h{var} }); "$in,$h{var}"`, "temporary,original"},
		{"an element of a tied array", `require Tie::Array; tie my @a, "Tie::StdArray"; @a = (5, 6); LiveSlots::tied_fetch(\@a)`, "6"},
	} {
		if got := requireModule(t, "LiveSlots", so, tc.code); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
