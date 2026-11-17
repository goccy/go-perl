//go:build darwin || linux

package xs_test

import "testing"

// A module that takes the last MAGIC of its own off a variable by relinking
// the chain itself (SvMAGIC_set, as Variable::Magic's dispell does) and
// calls mg_magical leaves the variable with no magic: perl sees no MAGIC
// and none of the magic flags (Variable::Magic t/32-hash.t: "no longer
// has the RMG flag set"). Before, the interpreter kept its anchor of the
// module's chain, and with it the flags. A svt_free that empties the chain
// while the variable is freed is no different from perl's either: the
// free goes on and the variable goes.
const chainEmptiedModuleSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int my_clear(pTHX_ SV *sv, MAGIC *mg) { PERL_UNUSED_ARG(sv); PERL_UNUSED_ARG(mg); return 0; }
static int my_get(pTHX_ SV *sv, MAGIC *mg) { PERL_UNUSED_ARG(sv); PERL_UNUSED_ARG(mg); return 0; }
static int frees = 0;
static int my_free_unlink(pTHX_ SV *sv, MAGIC *mg) {
    /* take the MAGIC being freed off the chain, as a module that keeps
     * the chain itself may */
    if (SvMAGIC(sv) == mg) SvMAGIC_set(sv, mg->mg_moremagic);
    frees++;
    return 0;
}
static MGVTBL vt_clear = {0, 0, 0, my_clear, 0, 0, 0, 0};
static MGVTBL vt_get = {my_get, 0, 0, 0, 0, 0, 0, 0};
static MGVTBL vt_unlink = {0, 0, 0, 0, my_free_unlink, 0, 0, 0};
static MGVTBL *vtbl_of(const char *which) {
    return strEQ(which, "clear") ? &vt_clear : strEQ(which, "get") ? &vt_get : &vt_unlink;
}
XS(XS_MyChain_cast) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "ref, which");
    sv_magicext(SvRV(ST(0)), NULL, PERL_MAGIC_ext, vtbl_of(SvPV_nolen(ST(1))), NULL, 0);
    XSRETURN_EMPTY;
}
/* takes off every MAGIC of the vtable by relinking the chain, frees them,
 * and has the flags computed again */
XS(XS_MyChain_dispell) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "ref, which");
    SV *sv = SvRV(ST(0));
    MGVTBL *vt = vtbl_of(SvPV_nolen(ST(1)));
    MAGIC *prev = NULL, *mg = SvMAGIC(sv), *more;
    int n = 0;
    for (; mg; mg = more) {
        more = mg->mg_moremagic;
        if (mg->mg_type == PERL_MAGIC_ext && mg->mg_virtual == vt) {
            if (prev) prev->mg_moremagic = more;
            else SvMAGIC_set(sv, more);
            Safefree(mg);
            n++;
        } else {
            prev = mg;
        }
    }
    mg_magical(sv);
    XSRETURN_IV(n);
}
XS(XS_MyChain_frees) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_IV(frees);
}
XS_EXTERNAL(boot_MyChain) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyChain::cast", XS_MyChain_cast);
    newXS_deffile("MyChain::dispell", XS_MyChain_dispell);
    newXS_deffile("MyChain::frees", XS_MyChain_frees);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeMagicChainEmptied(t *testing.T) {
	so := buildModuleSource(t, "MyChain", chainEmptiedModuleSource)
	cases := []struct{ name, code, want string }{
		{"a hash after its last MAGIC is taken off", `
require B;
sub magic_of { my $o = B::svref_2object($_[0]); my $f = $o->FLAGS;
  join "", map { $f & $_->[1] ? $_->[0] : "-" } [R => B::SVs_RMG()], [G => B::SVs_GMG()], [S => B::SVs_SMG()] }
my @r;
for my $which (qw(clear get)) {
  my %h = (a => 1);
  MyChain::cast(\%h, $which);
  push @r, "$which cast " . magic_of(\%h);
  my $n = MyChain::dispell(\%h, $which);
  push @r, "dispelled $n " . magic_of(\%h);
  $h{b} = 2;
  push @r, "used " . join(",", map { "$_=$h{$_}" } sort keys %h);
}
join "; ", @r`, "clear cast R--; dispelled 1 ---; used a=1,b=2; get cast -G-; dispelled 1 ---; used a=1,b=2"},
		{"a scalar after its last MAGIC is taken off, cast again", `
require B;
my $x = 5;
MyChain::cast(\$x, "get");
MyChain::dispell(\$x, "get");
my $plain = (B::svref_2object(\$x)->FLAGS & (B::SVs_RMG() | B::SVs_GMG() | B::SVs_SMG())) ? "magical" : "plain";
MyChain::cast(\$x, "get");
my $again = (B::svref_2object(\$x)->FLAGS & B::SVs_GMG()) ? "magical" : "plain";
MyChain::dispell(\$x, "get");
"$plain $again $x"`, "plain magical 5"},
		{"a svt_free that empties the chain while the variable is freed", `
{ my %h = (a => 1); MyChain::cast(\%h, "unlink"); }
{ my $s = 3; MyChain::cast(\$s, "unlink"); MyChain::cast(\$s, "get"); }
"frees " . MyChain::frees()`, "frees 2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := requireModule(t, "MyChain", so, tc.code); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
