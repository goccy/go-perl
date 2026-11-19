//go:build darwin || linux

package xs_test

import "testing"

// SvANY of a scalar is its body, where native code reaches a field by the
// offset the body struct names, as LMDB_File hands LMDB the address of a
// key's integer (&((XPVIV*)SvANY(sv))->xiv_iv): read through it, and
// written, it is the scalar's integer, in a bodyless IV, in a scalar with a
// string as well, and in a blessed one. Before, SvANY was a page that
// cannot be read, and LMDB_File died of SIGSEGV.
const svanyBodySource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* iv(sv, n): the integer read through the body's xiv_iv, then n written
 * there */
XS(XS_MySb_iv) {
    dXSARGS;
    SV *const sv = ST(0);
    IV *p;
    IV was;
    PERL_UNUSED_VAR(items);
    SvIV_please(sv);
    p = (IV *)&(((XPVIV *)SvANY(sv))->xiv_iv);
    was = *p;
    *p = SvIV(ST(1));
    ST(0) = sv_2mortal(newSViv(was));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MySb) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySb::iv", XS_MySb_iv);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeReachesAScalarsBody(t *testing.T) {
	so := buildModuleSource(t, "MySb", svanyBodySource)
	got := requireModule(t, "MySb", so, `my $i = 41;
my $s = "7"; my $x = $s + 0;
my $o = bless \(my $v = 5), "MySbT";
my @r = (MySb::iv($i, 42), $i, MySb::iv($s, 8), $s + 0, MySb::iv($$o, 6), $$o);
join "|", @r`)
	if want := "41|42|7|8|5|6"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
