//go:build darwin || linux

package xs_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// In C++ the kinds of scalar are types of their own, as perl's struct sv,
// av, hv, cv, gv and io are: a module overloads functions on SV *, AV *,
// HV *, CV *, GV * and IO *, and the perl API gives each kind its type. The
// SDK made them one type, and Google::ProtocolBuffers::Dynamic's
// push_level(SV *) and push_level(HV *) were one function declared twice.
// What is expected is what perl 5.36 on Linux gives with the same module.
func TestNativeCxxKindsOfScalarAreTypesOfTheirOwn(t *testing.T) {
	cxx := findCXX(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "TypesXX.cc")
	if err := os.WriteFile(src, []byte(typesXXSource), 0o644); err != nil {
		t.Fatal(err)
	}
	so := filepath.Join(dir, "TypesXX.so")
	args := append(append([]string{}, cxx[1:]...), "-shared", "-fPIC", "-x", "c++", "-std=c++11",
		"-I", "sdk/include", "-o", so, src)
	if runtime.GOOS == "darwin" {
		args = append(args, "-undefined", "dynamic_lookup")
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	} else {
		args = append([]string{"-D_GNU_SOURCE"}, args...)
	}
	if out, err := exec.Command(cxx[0], args...).CombinedOutput(); err != nil {
		t.Fatalf("compile TypesXX.so: %v\n%s", err, out)
	}
	program := `our (@x, %x, $x);
sub x { 1 }
open x, "<", "/dev/null";
TypesXX::kinds(\@x, \%x, \&x, \*x)`
	if got, want := requireModule(t, "TypesXX", so, program), "AV HV CV GV SV AV HV CV - SV HV GV HV AV HV CV GV"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

const typesXXSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* The kinds of scalar as types of their own, as C++ code uses them:
 * overloads on SV *, AV *, HV *, CV *, GV * and IO *, and the perl API's
 * functions and macros giving each kind. kinds(\@a, \%h, \&c, \*g): the
 * overload each value picks. */
static const char *kind(pTHX_ SV *) { return "SV"; }
static const char *kind(pTHX_ AV *) { return "AV"; }
static const char *kind(pTHX_ HV *) { return "HV"; }
static const char *kind(pTHX_ CV *) { return "CV"; }
static const char *kind(pTHX_ GV *) { return "GV"; }
static const char *kind(pTHX_ IO *) { return "IO"; }
XS(XS_TypesXX_kinds) {
    dXSARGS;
    if (items != 4) croak_xs_usage(cv, "aref, href, cref, gref");
    AV *av = (AV *)SvRV(ST(0));
    HV *hv = (HV *)SvRV(ST(1));
    CV *code = (CV *)SvRV(ST(2));
    GV *gv = (GV *)SvRV(ST(3));
    AV *gav = GvAV(gv);
    HV *ghv = GvHV(gv);
    CV *gcv = GvCV(gv);
    IO *gio = GvIO(gv);
    SV *gsv = GvSV(gv);
    HV *stash = GvSTASH(gv);
    GV *cvgv = CvGV(code);
    HV *cvstash = CvSTASH(code);
    HV *defstash = PL_defstash;
    HV *mainstash = gv_stashpv("main", 0);
    AV *inc = get_av("INC", 0);
    HV *env = get_hv("ENV", 0);
    CV *cvc = get_cv("TypesXX::kinds", 0);
    GV *fetched = gv_fetchpv("main::x", GV_ADD, SVt_PV);
    AV *made = newAV();
    HV *madeh = newHV();
    CV *sub = sv_2cv(ST(2), &stash, &fetched, 0);
    HV *objstash = SvSTASH((SV *)hv);
    SV *out = newSVpvf("%s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s", kind(aTHX_ av),
                       kind(aTHX_ hv), kind(aTHX_ code), kind(aTHX_ gv), kind(aTHX_ ST(0)),
                       gav ? kind(aTHX_ gav) : "-", ghv ? kind(aTHX_ ghv) : "-",
                       gcv ? kind(aTHX_ gcv) : "-", gio ? kind(aTHX_ gio) : "-",
                       gsv ? kind(aTHX_ gsv) : "-", kind(aTHX_ stash), kind(aTHX_ cvgv),
                       kind(aTHX_ defstash), kind(aTHX_ inc), kind(aTHX_ env), kind(aTHX_ cvc),
                       kind(aTHX_ fetched));
    PERL_UNUSED_VAR(cvstash);
    PERL_UNUSED_VAR(mainstash);
    PERL_UNUSED_VAR(sub);
    PERL_UNUSED_VAR(objstash);
    /* the counts of each kind, as perl's macros take any kind */
    SvREFCNT_inc(code);
    SvREFCNT_dec(code);
    SvREFCNT_inc_simple_void(gv);
    SvREFCNT_dec(gv);
    SvREFCNT_dec(made);
    SvREFCNT_dec(madeh);
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
extern "C" XS_EXTERNAL(boot_TypesXX);
XS_EXTERNAL(boot_TypesXX) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("TypesXX::kinds", XS_TypesXX_kinds, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`
