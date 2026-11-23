//go:build darwin || linux

package xs_test

import "testing"

// A module that takes the last MAGIC of its own off a variable by relinking
// the chain itself (SvMAGIC_set, as Variable::Magic's dispell does) and
// calls mg_magical leaves the variable with no magic: perl sees no MAGIC
// and none of the magic flags (Variable::Magic t/32-hash.t: "no longer
// has the RMG flag set"). Before, the interpreter kept its anchor of the
// module's chain, and with it the flags. The same holds for uvar magic
// the module attached with sv_magic: once it takes that MAGIC off (and
// mg_magical), the hash has no uvar magic left, as on perl, and its
// accesses call nothing of the module's; a callback of the uvar magic that
// takes it off (Variable::Magic t/15-self.t dispells from a fetch callback)
// has the access it runs for go on as on perl. A svt_free that empties the chain
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
static int uvar_calls = 0;
static I32 my_uvar(pTHX_ IV action, SV *sv) { PERL_UNUSED_ARG(action); PERL_UNUSED_ARG(sv); uvar_calls++; return 0; }
XS(XS_MyChain_uvar_cast) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    struct ufuncs uf;
    uf.uf_val = my_uvar;
    uf.uf_set = NULL;
    uf.uf_index = 0;
    sv_magic(SvRV(ST(0)), NULL, PERL_MAGIC_uvar, (const char *)&uf, sizeof uf);
    XSRETURN_EMPTY;
}
/* takes off the uvar MAGIC by relinking the chain, as Variable::Magic's
 * dispell does, frees it with its copy of the ufuncs, and has the flags
 * computed again */
XS(XS_MyChain_uvar_dispell) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    SV *sv = SvRV(ST(0));
    MAGIC *prev = NULL, *mg = SvMAGIC(sv);
    for (; mg; prev = mg, mg = mg->mg_moremagic)
        if (mg->mg_type == PERL_MAGIC_uvar) break;
    if (!mg) XSRETURN_IV(0);
    if (prev) prev->mg_moremagic = mg->mg_moremagic;
    else SvMAGIC_set(sv, mg->mg_moremagic);
    Safefree(mg->mg_ptr);
    Safefree(mg);
    mg_magical(sv);
    XSRETURN_IV(1);
}
/* a uvar callback that takes its own MAGIC off the hash it runs for, and
 * frees it once the call is over, as Variable::Magic's dispell from a
 * callback does (its freed tokens) */
static MAGIC *self_freed = NULL;
static I32 my_uvar_self(pTHX_ IV action, SV *sv) {
    PERL_UNUSED_ARG(action);
    MAGIC *prev = NULL, *mg = SvMAGIC(sv);
    for (; mg; prev = mg, mg = mg->mg_moremagic)
        if (mg->mg_type == PERL_MAGIC_uvar) break;
    if (mg) {
        if (prev) prev->mg_moremagic = mg->mg_moremagic;
        else SvMAGIC_set(sv, mg->mg_moremagic);
        mg->mg_moremagic = NULL;
        self_freed = mg;
        mg_magical(sv);
    }
    uvar_calls++;
    return 0;
}
XS(XS_MyChain_uvar_cast_self) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    struct ufuncs uf;
    uf.uf_val = my_uvar_self;
    uf.uf_set = NULL;
    uf.uf_index = 0;
    sv_magic(SvRV(ST(0)), NULL, PERL_MAGIC_uvar, (const char *)&uf, sizeof uf);
    XSRETURN_EMPTY;
}
XS(XS_MyChain_uvar_free_self) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    if (self_freed) {
        Safefree(self_freed->mg_ptr);
        Safefree(self_freed);
        self_freed = NULL;
    }
    XSRETURN_EMPTY;
}
XS(XS_MyChain_uvar_calls) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_IV(uvar_calls);
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
    newXS_deffile("MyChain::uvar_cast", XS_MyChain_uvar_cast);
    newXS_deffile("MyChain::uvar_dispell", XS_MyChain_uvar_dispell);
    newXS_deffile("MyChain::uvar_calls", XS_MyChain_uvar_calls);
    newXS_deffile("MyChain::uvar_cast_self", XS_MyChain_uvar_cast_self);
    newXS_deffile("MyChain::uvar_free_self", XS_MyChain_uvar_free_self);
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
		{"a hash after its uvar MAGIC is taken off", `
require B;
my %h = (a => 1);
MyChain::uvar_cast(\%h);
my $v = $h{a};
my $during = MyChain::uvar_calls() ? "called" : "not called";
my $n = MyChain::uvar_dispell(\%h);
my $before = MyChain::uvar_calls();
$v = $h{a}; $h{b} = 2;
my $o = B::svref_2object(\%h);
my $flags = ($o->FLAGS & (B::SVs_RMG() | B::SVs_GMG() | B::SVs_SMG())) ? "magical" : "plain";
my @types = $flags eq "magical" ? map { $_->TYPE } $o->MAGIC : ();
join " ", $during, "dispelled $n", $flags, "[@types]", (MyChain::uvar_calls() == $before ? "no calls after" : "called after"), join(",", map { "$_=$h{$_}" } sort keys %h)`,
			"called dispelled 1 plain [] no calls after a=1,b=2"},
		{"a uvar callback that takes its own MAGIC off", `
my %h = (a => "bar", b => "baz");
MyChain::uvar_cast_self(\%h);
my $before = MyChain::uvar_calls();
my $got = $h{a};
MyChain::uvar_free_self();
my $calls = MyChain::uvar_calls() - $before;
my $again = $h{b};
my $after = MyChain::uvar_calls() - $before;
$h{c} = "new";
join " ", $got, $calls, $again, $after, join(",", map { "$_=$h{$_}" } sort keys %h)`,
			"bar 1 baz 1 a=bar,b=baz,c=new"},
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
