//go:build darwin || linux

package xs_test

import "testing"

// pad.h's names of the pad being compiled, by offset (PAD_COMPNAME_*), as
// Switch::Declare reads them while code is compiled to tell an `our`
// variable from a `my` one: the name, whether it is an `our`, and its
// package. Before, they were not in the SDK, and Switch::Declare did not
// load.
const padCompnameSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* name(var): at BEGIN, the compiled pad's name of var, whether it is an
 * our, and the package of an our */
XS(XS_MyPc_name) {
    dXSARGS;
    STRLEN len;
    const char *const var = SvPV(ST(0), len);
    const PADOFFSET po = pad_findmy_pvn(var, len, 0);
    SV *r;
    PERL_UNUSED_VAR(items);
    if (po == NOT_IN_PAD) {
        ST(0) = sv_2mortal(newSVpvs("none"));
        XSRETURN(1);
    }
    r = newSVpvf("%s:%d", PAD_COMPNAME_PV(po), PAD_COMPNAME_FLAGS_isOUR(po) ? 1 : 0);
    if (PAD_COMPNAME_FLAGS_isOUR(po)) sv_catpvf(r, ":%s", HvNAME(PAD_COMPNAME_OURSTASH(po)));
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyPc) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyPc::name", XS_MyPc_name);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeReadsTheCompiledPadsNames(t *testing.T) {
	so := buildModuleSource(t, "MyPc", padCompnameSource)
	got := requireModule(t, "MyPc", so, `my $r = eval q{
package MyPcT;
our $o = 1;
my $m = 2;
our $seen;
BEGIN { $seen = join "|", MyPc::name('$o'), MyPc::name('$m'), MyPc::name('$nope') }
$seen;
};
$r // "died: $@"`)
	if want := "$o:1:MyPcT|$m:0|none"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
