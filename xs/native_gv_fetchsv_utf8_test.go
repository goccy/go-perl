//go:build darwin || linux

package xs_test

import "testing"

// gv_fetchsv looks a name up as characters when its SV is (perl's passes
// SvUTF8 of the name with the flags), as Ex::Op does to install an exported
// sub of a name beyond ASCII: the glob is the one Perl code names. Before,
// the name's UTF-8 bytes made a glob of their own, and the sub was not
// found under its name.
const gvFetchsvUTF8Source = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* define(name, value): the scalar of the glob gv_fetchsv makes is value;
 * -> whether the glob's name is characters */
XS(XS_MyGf_define) {
    dXSARGS;
    GV *const gv = gv_fetchsv(ST(0), GV_ADD, SVt_PV);
    PERL_UNUSED_VAR(items);
    sv_setsv(GvSVn(gv), ST(1));
    ST(0) = sv_2mortal(newSViv(GvNAMEUTF8(gv) ? 1 : 0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyGf) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyGf::define", XS_MyGf_define);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeGvFetchsvTakesTheNamesCharacters(t *testing.T) {
	so := buildModuleSource(t, "MyGf", gvFetchsvUTF8Source)
	got := requireModule(t, "MyGf", so, `no strict 'refs';
my $name = "MyGfT::\x{e9}t\x{e9}\x{263a}";
my $u = MyGf::define($name, "chars");
my $plain = "MyGfT::plain";
my $p = MyGf::define($plain, "bytes");
my @keys = map { length } grep { /t/ } keys %MyGfT::;
join "|", $u, ${"MyGfT::\x{e9}t\x{e9}\x{263a}"} // "none", $p, ${"MyGfT::plain"} // "none", join(",", sort @keys)`)
	if want := "1|chars|0|bytes|4"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
