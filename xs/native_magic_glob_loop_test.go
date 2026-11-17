//go:build darwin || linux

package xs_test

import "testing"

// Magic native code puts on what a slot of a glob holds, with the glob as
// its object (the isa magic of @ISA, as Package::Stash::XS adds a package's
// @ISA), holds no reference on the glob, as perl's sv_magicext takes none:
// the glob holds the array, and a reference back would be a loop that
// keeps both alive after the package is gone (Moose's anon class leak test
// found the glob and the array of every anonymous class leaked).
const magicGlobLoopSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* MyGl::add_isa(package): the package's @ISA made as Package::Stash::XS
 * makes it */
XS(XS_MyGl_add_isa) {
    dXSARGS;
    STRLEN len;
    const char *pkg;
    HV *stash;
    GV *gv;
    AV *av;
    PERL_UNUSED_VAR(items);
    pkg = SvPV(ST(0), len);
    stash = gv_stashpvn(pkg, (U32)len, GV_ADD);
    gv = (GV *)*hv_fetch(stash, "ISA", 3, 1);
    if (!isGV(gv))
        gv_init_pvn(gv, stash, "ISA", 3, GV_ADDMULTI);
    av = GvAVn(gv);
    sv_magic((SV *)av, (SV *)gv, PERL_MAGIC_isa, NULL, 0);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MyGl) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyGl::add_isa", XS_MyGl_add_isa);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeMagicOnAGlobsSlotHoldsNoGlob(t *testing.T) {
	so := buildModuleSource(t, "MyGl", magicGlobLoopSource)
	got := requireModule(t, "MyGl", so, `use B ();
use Scalar::Util ();
MyGl::add_isa("MyGl::Native");
@{"MyGl::Perl::ISA"} = ();
my @r = map { B::svref_2object(\*{"MyGl::${_}::ISA"})->REFCNT } qw(Native Perl);
# the glob goes with its package
my $glob = \*{"MyGl::Native::ISA"};
Scalar::Util::weaken($glob);
delete $MyGl::{"Native::"};
push @r, defined $glob ? "kept" : "freed";
"@r"`)
	if want := "2 2 freed"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
