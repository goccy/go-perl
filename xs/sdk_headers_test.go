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
	const want = "go-perl gives native code no access to this interpreter internal: regcomp.h"
	if !strings.Contains(string(out), want) {
		t.Errorf("the compiler said:\n%s\nwant it to say %q", out, want)
	}
}

// A module that uses an internal of perl's that go-perl does not give
// native code (the save stack, the bodies of SVs, the tmps stack, the
// layers of a PerlIO handle) stops building with one fixed phrase and the
// name, for each use; #ifdef sees the name defined, and a module that does
// not use one builds (TestSDKHeadersCompileWherePerlsDo). Before, such a
// build failed with "implicit declaration" or "undeclared", which said
// nothing of why.
func TestSDKInternalsGoPerlDoesNotGive(t *testing.T) {
	_, out, err := compileModuleSource(t, "Internals", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"
#ifndef PL_main_root
#error "PL_main_root is not defined"
#endif
static int uses(SV *sv, PerlIO *f) {
    I32 base = PL_savestack_ix;
    void *root = PL_main_root;
    void *top = PL_eval_root;
    SSize_t n = sizeof(XPVMG);
    int t = CATCH_SET(1);
    PerlIO_funcs *tab = PerlIO_find_layer(aTHX_ "encoding", 8, 1);
    pad_push(NULL, 0);
    (void)sv; (void)f; (void)top; (void)n; (void)root; (void)tab;
    return (int)base + t;
}
`)
	if err == nil {
		t.Fatal("a module using perl's internals built")
	}
	const phrase = "go-perl gives native code no access to this interpreter internal: "
	for _, name := range []string{"PL_main_root", "PL_eval_root", "XPVMG", "CATCH_SET", "PerlIO_find_layer", "pad_push"} {
		if !strings.Contains(string(out), phrase+name) {
			t.Errorf("no %q in what the compiler said:\n%s", phrase+name, out)
		}
	}
	if strings.Contains(string(out), phrase+"PL_savestack_ix") {
		t.Errorf("PL_savestack_ix, which the SDK gives, was refused:\n%s", out)
	}
	if strings.Contains(string(out), "is not defined") {
		t.Errorf("#ifdef did not see the name defined:\n%s", out)
	}
}
