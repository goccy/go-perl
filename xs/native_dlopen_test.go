//go:build darwin || linux

package xs_test

import (
	"strconv"
	"testing"
)

// A library compiled against perl.h that a module loads itself and calls
// with SVs (Affix compiles C that way and calls it from Perl) works on the
// interpreter as the module does: on perl it binds to the perl API of the
// process; here the module's dlopen hands it the module's vtable. Before,
// its first use of the API dereferenced a frame it never had (SIGSEGV,
// Affix t/018_sv_type.t). What is expected is what perl 5.36 on Linux gives
// with the same module and library.
func TestLibraryAModuleLoadsUsesTheInterpreter(t *testing.T) {
	lib := buildModuleSource(t, "SvCode", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
IV lib_next_iv(SV *sv) {
    dTHX;
    return SvIV(sv) + 1;
}
SV *lib_new_sv(int v) {
    dTHX;
    return sv_2mortal(newSViv(v * 2));
}
void lib_inc(SV *sv) {
    dTHX;
    if (SvIOK(sv)) sv_setiv(sv, SvIV(sv) + 1);
}
`)
	so := buildModuleSource(t, "DlLib", `#include <dlfcn.h>
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* call(path, sv): loads the library at path, calls its lib_inc with sv,
 * its lib_next_iv with sv and its lib_new_sv with 21 */
XS(XS_DlLib_call) {
    dXSARGS;
    void *h;
    IV (*next_iv)(SV *);
    SV *(*new_sv)(int);
    void (*inc)(SV *);
    IV next;
    SV *made;
    if (items != 2) croak_xs_usage(cv, "path, sv");
    h = dlopen(SvPV_nolen(ST(0)), RTLD_NOW | RTLD_LOCAL);
    if (!h) croak("dlopen: %s", dlerror());
    next_iv = (IV (*)(SV *))dlsym(h, "lib_next_iv");
    new_sv = (SV *(*)(int))dlsym(h, "lib_new_sv");
    inc = (void (*)(SV *))dlsym(h, "lib_inc");
    if (!next_iv || !new_sv || !inc) croak("dlsym failed");
    inc(ST(1));
    next = next_iv(ST(1));
    made = new_sv(21);
    ST(0) = sv_2mortal(newSVpvf("%" IVdf " %" IVdf " %" IVdf, SvIV(ST(1)), next, SvIV(made)));
    XSRETURN(1);
}
XS_EXTERNAL(boot_DlLib) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("DlLib::call", XS_DlLib_call, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $v = 41;
DlLib::call(` + strconv.Quote(lib) + `, $v) . " / $v"`
	if got, want := requireModule(t, "DlLib", so, program), "42 43 42 / 42"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
