//go:build darwin || linux

package xs_test

import "testing"

// A scalar native code points at memory of its own and makes read-only only
// after setting its magic, as File::Map does for a read-only map
// (map_file($x, $file, '<'), which PDL::PP compares a file with): the
// string reaches the interpreter as a read-only copy. SvSETMAGIC calls the
// set magic only of a scalar that has some, as perl's does, so it is no
// operation that ends the scalar's making before it is read-only. A copy
// does not follow later changes to the mapped memory (or the file), as a
// map would. A scalar left writable is still refused
// (TestNativeStringInNativeMemory).
const readonlyMapSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static char region[] = "mapped bytes";
static MGVTBL map_vtbl;
/* map(\\$x, readonly): $x's buffer is the region, then its set magic, ext
 * magic of the module's, and read-only when asked */
XS(XS_MyRm_map) {
    dXSARGS;
    SV *const var = SvRV(ST(0));
    const int readonly = (int)SvIV(ST(1));
    MAGIC *mg;
    PERL_UNUSED_VAR(items);
    if (SvTYPE(var) < SVt_PV) sv_upgrade(var, SVt_PV);
    if (SvPVX(var) && SvLEN(var)) Safefree(SvPVX(var));
    SvPV_set(var, region);
    SvLEN_set(var, 0);
    SvCUR_set(var, sizeof region - 1);
    SvPOK_only(var);
    SvSETMAGIC(var);
    mg = sv_magicext(var, NULL, PERL_MAGIC_ext, &map_vtbl, NULL, 0);
    mg->mg_private = 0x4d50;
    if (readonly) SvREADONLY_on(var);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MyRm) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyRm::map", XS_MyRm_map);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeReadOnlyMapReachesTheInterpreter(t *testing.T) {
	so := buildModuleSource(t, "MyRm", readonlyMapSource)
	got := requireModule(t, "MyRm", so, `my $x;
MyRm::map(\$x, 1);
my $r = join ",", $x, length($x), Internals::SvREADONLY($x) ? "ro" : "rw";
$r`)
	if want := "mapped bytes,12,ro"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
