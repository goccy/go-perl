//go:build darwin || linux

package xs_test

import "testing"

// A croak of native code runs $SIG{__DIE__} where perl's croak runs it:
// while the native code is still running, inside the scopes it opened.
// Text::Xslate sets PL_diehook for a render and its hook names Text::Xslate
// in what dies there; the hook ran after the scopes were left, so the hook
// outside ran instead. A hook that dies makes its exception the one thrown,
// and no hook runs again on the way out. What is expected is what perl
// 5.36 on Linux gives with the same module.
func TestNativeCroakRunsTheDieHookInPlace(t *testing.T) {
	so := buildModuleSource(t, "DieHook", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* croak_in(hook, msg): croaks with $SIG{__DIE__} set to hook for the call,
 * as Text::Xslate sets PL_diehook for a render */
XS(XS_DieHook_croak_in) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ENTER;
    SAVESPTR(PL_diehook);
    PL_diehook = ST(0);
    croak("%s", SvPV_nolen(ST(1)));
    LEAVE;
    XSRETURN_EMPTY;
}
/* croak_plain(msg): croaks */
XS(XS_DieHook_croak_plain) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    croak("%s", SvPV_nolen(ST(0)));
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_DieHook) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("DieHook::croak_in", XS_DieHook_croak_in, __FILE__);
    newXS("DieHook::croak_plain", XS_DieHook_croak_plain, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my @outer;
local $SIG{__DIE__} = sub { push @outer, $_[0]; die "outer: $_[0]" };
my $inner = sub { die "inner: $_[0]" };
eval { DieHook::croak_in($inner, "deep") };
my $e1 = $@;
$e1 =~ s/ at .*//s;
eval { DieHook::croak_plain("plain") };
my $e2 = $@;
$e2 =~ s/ at .*//s;
my $seen = join "/", map { (my $m = $_) =~ s/ at .*//s; $m } @outer;
join ",", $e1, $e2, scalar(@outer), $seen`
	if got, want := requireModule(t, "DieHook", so, program), "inner: deep,outer: plain,1,plain"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
