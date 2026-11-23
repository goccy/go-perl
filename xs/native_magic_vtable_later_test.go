//go:build darwin || linux

package xs_test

import "testing"

// A module that attaches magic without a vtable, finds it with mg_find,
// gives it a vtable and calls mg_magical has the hooks of that vtable
// called, as pTk does for the internal representation it keeps on a
// Tcl_Obj (Tcl_ObjMagic: sv_magic(obj, data, '~', NULL, 0), then
// mg->mg_virtual = &TclObj_vtab and mg_magical). Its set hook drops the
// representation when the value is assigned. Before, the interpreter was
// never told of the hooks: an assignment kept the representation of the
// old value, and Tk read an option name as the option it named before
// (Can't set -relief to `flat': unknown color name).
func TestNativeMagicGivenAVtableAfterItWasAttached(t *testing.T) {
	so := buildModuleSource(t, "LateVtbl", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int sets;
static int late_set(pTHX_ SV *sv, MAGIC *mg) {
    PERL_UNUSED_VAR(mg);
    sets++;
    sv_unmagic(sv, PERL_MAGIC_ext);
    return 0;
}
static MGVTBL late_vtbl = { NULL, late_set, NULL, NULL, NULL, NULL, NULL, NULL };
XS(XS_LateVtbl_attach) {
    dXSARGS;
    SV *sv = ST(0);
    SV *data = newSV(16);
    MAGIC *mg;
    PERL_UNUSED_VAR(items);
    sv_upgrade(sv, SVt_PVMG);
    sv_magic(sv, data, PERL_MAGIC_ext, NULL, 0);
    SvREFCNT_dec(data);
    SvRMAGICAL_off(sv);
    mg = mg_find(sv, PERL_MAGIC_ext);
    if (!mg || mg->mg_obj != data) croak("no magic found");
    mg->mg_virtual = &late_vtbl;
    mg_magical(sv);
    XSRETURN_EMPTY;
}
XS(XS_LateVtbl_has) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_IV(mg_find(ST(0), PERL_MAGIC_ext) ? 1 : 0);
}
XS(XS_LateVtbl_sets) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_IV(sets);
}
XS_EXTERNAL(boot_LateVtbl) {
    dXSBOOTARGSXSAPIVERCHK;
    sets = 0; /* this interpreter's */
    newXS_deffile("LateVtbl::attach", XS_LateVtbl_attach);
    newXS_deffile("LateVtbl::has", XS_LateVtbl_has);
    newXS_deffile("LateVtbl::sets", XS_LateVtbl_sets);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	got := requireModule(t, "LateVtbl", so, `my $x = "-background";
LateVtbl::attach($x);
my @seen = (LateVtbl::has($x));
$x = "-relief";
push @seen, LateVtbl::sets(), LateVtbl::has($x);
$x = "-text";
push @seen, LateVtbl::sets(), $x;
join ",", @seen`)
	if want := "1,1,0,1,-text"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
