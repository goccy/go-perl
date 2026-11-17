//go:build darwin || linux

package xs_test

import "testing"

// A run loop of native code (PL_runops, as Test::LeakTrace installs and
// leaves installed, with its per-interpreter context read on every op) runs
// the ops of code that dies inside an eval, with $@, $! and
// $SIG{__DIE__} localized around it as Test::Builder's _try does, and goes
// on: what the eval catches, and every value around it, is as with perl's
// own loop.
const runopsModuleSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <stdlib.h>
#define MY_CXT_KEY "MyRunops::_guts" XS_VERSION
typedef struct { int enabled; int need_stateinfo; } my_cxt_t;
START_MY_CXT
static int my_runops(pTHX) {
    dMY_CXT;
    COP *last_cop = PL_curcop;
    while ((PL_op = PL_op->op_ppaddr(aTHX))) {
        PERL_ASYNC_CHECK();
        if (!MY_CXT.need_stateinfo) continue;
        if (last_cop != PL_curcop) last_cop = PL_curcop;
    }
    TAINT_NOT;
    return 0;
}
XS(XS_MyRunops_install) {
    dXSARGS;
    dMY_CXT;
    PERL_UNUSED_VAR(items);
    MY_CXT.enabled = 1;
    PL_runops = my_runops;
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MyRunops) {
    dXSBOOTARGSXSAPIVERCHK;
    MY_CXT_INIT;
    MY_CXT.enabled = 0;
    MY_CXT.need_stateinfo = 0;
    if (getenv("MYRUNOPS_AT_BOOT")) PL_runops = my_runops;
    newXS_deffile("MyRunops::install", XS_MyRunops_install);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeRunLoopThroughADie(t *testing.T) {
	so := buildModuleSource(t, "MyRunops", runopsModuleSource)
	got := requireModule(t, "MyRunops", so, `my @w; local $SIG{__WARN__} = sub { push @w, $_[0] };
MyRunops::install();
sub try_it {
    my ($code) = @_;
    my ($return, $error);
    {
        local $!;
        local $@;
        local $SIG{__DIE__};
        $return = eval { $code->() };
        $error = $@;
    }
    return ($return, $error);
}
my @r;
for my $x ([1], {a => 2}, \3) {
    my ($ret, $err) = try_it(sub { $x->isa('UNIVERSAL') });
    push @r, defined $ret ? "lived" : ($err =~ /unblessed reference/ ? "died" : $err);
}
sub f { my @args = @_; $_ = ref $_ for @args; join "/", @args }
push @r, f(\1, [2]);
push @r, "warnings: " . join("|", @w);
join ",", @r`)
	if want := "died,died,died,SCALAR/ARRAY,warnings: "; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// The run loop installed when the module is loaded (Test::LeakTrace's BOOT
// does so) runs everything after, in one call of the loop: a die inside it
// that an eval inside it catches leaves the scopes up to that eval only,
// and the code after the eval still has its own (the values `local` gave,
// $@ as the eval left it), as Test::More's comparisons need: they used to
// warn "Attempt to free unreferenced scalar".
func TestNativeRunLoopInstalledAtLoad(t *testing.T) {
	t.Setenv("MYRUNOPS_AT_BOOT", "1")
	so := buildModuleSource(t, "MyRunops", runopsModuleSource)
	got := requireModule(t, "MyRunops", so, `my @w; local $SIG{__WARN__} = sub { push @w, $_[0] };
our $g = "outer";
sub try_it {
    my ($code) = @_;
    my ($return, $error, $inside);
    {
        local $g = "inner";
        local $@;
        $return = eval { $code->() };
        $error = $@;
        $inside = $g;
    }
    return ($return, $error =~ /^(\w+)/ ? $1 : $error, $inside, $g);
}
# the interpreter goes on after this eval with its next call of the loop
eval { die "start\n" };
my @r = (join "/", map { defined ? $_ : "undef" } try_it(sub { die "boom\n" }));
require Test::More;
my %a = (foo => 42);
push @r, Test::More::eq_array([\$a{foo}], [\42]) ? 1 : 0, Test::More::eq_array([1, [2]], [1, [2]]) ? 1 : 0;
push @r, "warnings: " . join("|", @w);
join ",", @r`)
	if want := "undef/boom/inner/outer,1,1,warnings: "; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
