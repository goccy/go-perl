//go:build darwin || linux

package xs_test

import "testing"

// A glob's GP is the interpreter's struct gp where it is, as Data::Alias
// uses it: GvGP gives it, its refcount is counted in place, GvGP_set gives
// it to a glob of the module's own (whose free counts it down again), and
// the glob's slots are its fields: &GvSV(gv) is the field's address, which
// a destructor that runs after the call writes through (Data::Alias's
// localization of a package variable; a scope an XSUB is called in ends
// with the call), and GvEGV and GvFORM are lvalues.
// Before, GvGP was an identity only, and the slots were copies whose
// address was good for the one call.
const globGPSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* share(\\*glob): the GP's count, the count with a second glob of the
 * module's own made to share it, whether that glob's scalar is the first's,
 * and the count once that glob is freed */
XS(XS_MyGp_share) {
    dXSARGS;
    GV *const gv = (GV *)SvRV(ST(0));
    GP *const gp = GvGP(gv);
    SV *const other = newSV(0);
    const U32 before = gp->gp_refcnt;
    SV *r;
    PERL_UNUSED_VAR(items);
    sv_upgrade(other, SVt_PVGV);
    isGV_with_GP_on(other);
    ++gp->gp_refcnt;
    GvGP_set(other, gp);
    r = newSVpvf("%u %u %d", (unsigned)before, (unsigned)gp->gp_refcnt,
                 GvSV((GV *)other) == GvSV(gv) && GvGP((GV *)other) == gp);
    sv_free(other);
    sv_catpvf(r, " %u", (unsigned)GvGP(gv)->gp_refcnt);
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
static void my_unlocalize(pTHX_ void *gp_v) {
    GP *const gp = (GP *)gp_v;
    SV *const value = (SV *)SSPOPPTR;
    SV **const sptr = (SV **)SSPOPPTR;
    SV *const old = *sptr;
    *sptr = value;
    SvREFCNT_dec(old);
    --gp->gp_refcnt;
}
/* localize(\\*glob, value, code): the glob's scalar is value while code
 * runs, until the call's scope ends, and is put back by a destructor
 * through its address then */
XS(XS_MyGp_localize) {
    dXSARGS;
    GV *const gv = (GV *)SvRV(ST(0));
    GP *const gp = GvGP(gv);
    SV **const sptr = &GvSV(gv);
    PERL_UNUSED_VAR(items);
    SSCHECK(2);
    SSPUSHPTR(sptr);
    SSPUSHPTR(*sptr);
    SAVEDESTRUCTOR_X(my_unlocalize, (void *)gp);
    ++gp->gp_refcnt;
    *sptr = newSVsv(ST(1));
    PUSHMARK(SP);
    call_sv(ST(2), G_VOID | G_DISCARD);
    XSRETURN_EMPTY;
}
/* egv(\\*glob, \\*other): GvEGV made other's glob and read back, and the
 * format slot's address, through which it is read */
XS(XS_MyGp_egv) {
    dXSARGS;
    GV *const gv = (GV *)SvRV(ST(0));
    GV *const other = (GV *)SvRV(ST(1));
    GV *const was = GvEGV(gv);
    CV **const formp = &GvFORM(gv);
    SV *r;
    PERL_UNUSED_VAR(items);
    GvEGV(gv) = other;
    r = newSVpvf("%d %d %d", was == gv, GvEGV(gv) == other, *formp == GvFORM(gv));
    GvEGV(gv) = was;
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyGp) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyGp::share", XS_MyGp_share);
    newXS_deffile("MyGp::localize", XS_MyGp_localize);
    newXS_deffile("MyGp::egv", XS_MyGp_egv);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeUsesTheGlobsGP(t *testing.T) {
	so := buildModuleSource(t, "MyGp", globGPSource)
	got := requireModule(t, "MyGp", so, `our $x = "outside";
my @r = (MyGp::share(\*x));
MyGp::localize(\*x, "inside", sub { push @r, $x });
push @r, $x, MyGp::share(\*x), MyGp::egv(\*x, \*STDOUT);
join "|", @r`)
	if want := "1 2 1 1|inside|outside|1 2 1 1|1 1 1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
