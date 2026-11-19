//go:build darwin || linux

package xs_test

import (
	"regexp"
	"testing"
)

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

// File::Map's own steps for a map (mmap_impl, reset_var and add_magic of
// File::Map 0.71): the buffer stored through the SvPVX/SvLEN/SvCUR
// lvalues, SvPOK_only_UTF8, SvSETMAGIC, ext magic by sv_magicext with
// MGf_LOCAL set on it, SvTAINTED_on, and SvREADONLY_on last for a map that
// is not writable. Each of the steps after the string is an operation of
// the interpreter that is not one that makes the scalar, and the read-only
// flag comes after them: a read-only map reaches the interpreter as a copy,
// decided at the end of the native call. A writable one is refused there,
// with the scalar left no string: perl would write into the map through
// it, and gperl's copy would not follow (perl maps it).
const fileMapSequenceSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static char fm_region[] = "mapped bytes";
static int fm_write(pTHX_ SV *sv, MAGIC *mg) {
    PERL_UNUSED_VAR(sv);
    PERL_UNUSED_VAR(mg);
    return 0;
}
static MGVTBL fm_vtbl = {0, fm_write, 0, 0, 0, 0, 0, 0};
/* map(\\$x, writable) */
XS(XS_MyFm_map) {
    dXSARGS;
    SV *const var = SvRV(ST(0));
    const int writable = (int)SvIV(ST(1));
    MAGIC *mg;
    PERL_UNUSED_VAR(items);
    if (SvPOK(var)) SvPV_free(var);
    SvUPGRADE(var, SVt_PVMG);
    SvPVX(var) = fm_region;
    SvLEN(var) = 0;
    SvCUR(var) = sizeof fm_region - 1;
    SvPOK_only_UTF8(var);
    SvSETMAGIC(var);
    mg = sv_magicext(var, NULL, PERL_MAGIC_ext, &fm_vtbl, (const char *)fm_region, 0);
    mg->mg_flags |= MGf_LOCAL;
    SvTAINTED_on(var);
    if (!writable) SvREADONLY_on(var);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MyFm) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyFm::map", XS_MyFm_map);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeFileMapSequenceMapsReadOnly(t *testing.T) {
	so := buildModuleSource(t, "MyFm", fileMapSequenceSource)
	got := requireModule(t, "MyFm", so, `my $x;
MyFm::map(\$x, 0);
join ",", $x, length($x), Internals::SvREADONLY($x) ? "ro" : "rw"`)
	if want := "mapped bytes,12,ro"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	got = requireModule(t, "MyFm", so, `my $y;
my $r = eval { MyFm::map(\$y, 1); 1 } ? "mapped" : $@;
join ",", $r, defined $y ? "string" : "no string"`)
	if !regexp.MustCompile(`(?s)^the string of a scalar cannot be in native memory on go-perl: .* only when it is read-only at .*\n,no string$`).MatchString(got) {
		t.Errorf("writable map: got %q, want the refusal and no string", got)
	}
}

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
