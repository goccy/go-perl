//go:build darwin || linux

package xs_test

import (
	"strings"
	"testing"
)

// headersModule includes perl's headers by the names a distribution uses,
// proto.h among them, and uses the common macros of an XSUB.
const headersModule = `#include "EXTERN.h"
#include "perl.h"
#include "proto.h"
#include "embed.h"
#include "XSUB.h"
#define MY_CXT_KEY "Headers::_guts" XS_VERSION
typedef struct { int calls; } my_cxt_t;
START_MY_CXT
static SV *call_back(SV *code, SV *arg) {
    dTHX;
    dSP;
    int n;
    SV *ret;
    ENTER;
    SAVETMPS;
    PUSHMARK(SP);
    XPUSHs(arg);
    PUTBACK;
    n = call_sv(code, G_SCALAR);
    SPAGAIN;
    ret = n == 1 ? newSVsv(POPs) : newSV(0);
    PUTBACK;
    FREETMPS;
    LEAVE;
    return ret;
}
XS(XS_Headers_called) {
    dXSARGS;
    dXSTARG;
    dXSI32;
    dMY_CXT;
    if (items != 2) croak_xs_usage(cv, "code, arg");
    MY_CXT.calls += 1 + ix;
    sv_setsv(TARG, sv_2mortal(call_back(ST(0), ST(1))));
    sv_catpvf(TARG, " %d", MY_CXT.calls);
    ST(0) = TARG;
    XSRETURN(1);
}
XS(XS_Headers_joined) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "a, b");
    {
        STRLEN la, lb;
        const char *a = SvPV(ST(0), la);
        const char *b = SvPV(ST(1), lb);
        SV *out = newSVpvn(a, la);
        HV *hv = newHV();
        SV **svp;
        sv_catpvn(out, b, lb);
        (void)hv_stores(hv, "joined", out);
        svp = hv_fetchs(hv, "joined", 0);
        ST(0) = sv_2mortal(svp ? newSVsv(*svp) : &PL_sv_undef);
        SvREFCNT_dec((SV *)hv);
        XSRETURN(1);
    }
}
XS_EXTERNAL(boot_Headers) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("Headers::joined", XS_Headers_joined);
    newXS_deffile("Headers::called", XS_Headers_called);
    {
        MY_CXT_INIT;
        MY_CXT.calls = 0;
    }
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// The SDK's headers compile wherever perl's do: under the warnings
// distributions make errors (Alien-SNMP builds with
// -Werror=declaration-after-statement, JavaBin with
// -Werror=missing-field-initializers), and by the names of perl's CORE
// headers a distribution includes (proto.h: Net-Z3950-SimpleServer,
// Sys-Ibam, Wifi-WpaCtrl). Before, the SDK's own perl.h failed those
// builds, and proto.h was not there.
func TestSDKHeadersCompileWherePerlsDo(t *testing.T) {
	so, out, err := compileModuleSource(t, "Headers", headersModule,
		"-Wall", "-Werror=declaration-after-statement", "-Werror=missing-field-initializers")
	if err != nil {
		t.Fatalf("compile Headers.so: %v\n%s", err, out)
	}
	const code = `join ",", Headers::joined("left ", "right"), Headers::called(sub { "<$_[0]>" }, "a"), Headers::called(sub { uc $_[0] }, "b")`
	if got, want := requireModule(t, "Headers", so, code), "left right,<a> 1,B 2"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A header of perl's CORE whose structures go-perl cannot give stops the
// build with a message naming what native code cannot have, instead of a
// missing file or code that reads them wrongly: regcomp.h, the program of a
// compiled regexp (Regexp-MinLength walks it).
func TestSDKHeaderThatCannotBeGiven(t *testing.T) {
	_, out, err := compileModuleSource(t, "Regcomp", `#include "EXTERN.h"
#include "perl.h"
#include "regcomp.h"
`)
	if err == nil {
		t.Fatal("a module including regcomp.h built")
	}
	const want = "regcomp.h: the program of a compiled regexp (its regnodes) is the interpreter's own on go-perl"
	if !strings.Contains(string(out), want) {
		t.Errorf("the compiler said:\n%s\nwant it to say %q", out, want)
	}
}
