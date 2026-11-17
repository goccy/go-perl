//go:build darwin || linux

package xs_test

import "testing"

// &PL_sv_undef, &PL_sv_yes and &PL_sv_no taken while the module is being
// loaded, before it is booted (a C++ static initializer: Geo::GDAL's
// `static SV *VSIStdoutSetRedirectionFct = &PL_sv_undef;`, or a
// constructor function) are the three when the module hands them to the
// interpreter later: returned, copied, compared with the default and
// replaced. Loading such a module crashed in the SDK (the frame of a
// module that is not booted yet: SIGSEGV, addr=0x88). What is expected is
// what perl 5.36 on Linux gives with the same module.
func TestNativeImmortalsTakenWhileLoading(t *testing.T) {
	so := buildModuleSource(t, "StaticSV", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* Defaults taken while the module is being loaded, as a C++ static
 * initializer takes them (Geo::GDAL: static SV *f = &PL_sv_undef;). */
static SV *undef_default, *yes_default, *no_default;
__attribute__((constructor)) static void take_defaults(void) {
    dTHX;
    undef_default = &PL_sv_undef;
    yes_default = &PL_sv_yes;
    no_default = &PL_sv_no;
}
XS(XS_StaticSV_undef_default) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = undef_default;
    XSRETURN(1);
}
XS(XS_StaticSV_yes_default) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = yes_default;
    XSRETURN(1);
}
XS(XS_StaticSV_no_default) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = no_default;
    XSRETURN(1);
}
/* copy(): a new scalar set from the undef default, as a copy */
XS(XS_StaticSV_copy) {
    dXSARGS;
    SV *sv = newSV(0);
    PERL_UNUSED_VAR(items);
    sv_setsv(sv, undef_default);
    ST(0) = sv_2mortal(sv);
    XSRETURN(1);
}
/* replace(code): what Geo::GDAL does with its default when it is given a
 * callback: drops what it had unless that is the default, keeps the new */
XS(XS_StaticSV_replace) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    if (undef_default != &PL_sv_undef) SvREFCNT_dec(undef_default);
    undef_default = newSVsv(ST(0));
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_StaticSV) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("StaticSV::undef_default", XS_StaticSV_undef_default, __FILE__);
    newXS("StaticSV::yes_default", XS_StaticSV_yes_default, __FILE__);
    newXS("StaticSV::no_default", XS_StaticSV_no_default, __FILE__);
    newXS("StaticSV::copy", XS_StaticSV_copy, __FILE__);
    newXS("StaticSV::replace", XS_StaticSV_replace, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my @got;
push @got, defined(StaticSV::undef_default()) ? "defined" : "undef";
push @got, StaticSV::yes_default() ? "true" : "false";
push @got, StaticSV::no_default() ? "true" : "false";
push @got, defined(StaticSV::copy()) ? "defined" : "undef";
StaticSV::replace("a callback");
push @got, StaticSV::undef_default();
join ",", @got`
	if got, want := requireModule(t, "StaticSV", so, program), "undef,true,false,undef,a callback"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
