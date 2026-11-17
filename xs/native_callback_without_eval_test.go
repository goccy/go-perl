//go:build darwin || linux

package xs_test

import "testing"

// call_sv of native code without G_EVAL is perl's: a die in the callback
// unwinds to the eval beyond the XSUB, $SIG{__DIE__} runs once for it, and
// caller(1) in the callback is the sub that called the XSUB. The bridge
// called the callback inside an eval of its own and the SDK raised $@
// again: the hook ran twice, and caller(1) was "(eval)" (Scope::Cleanup's
// t/caller.t). The exception object passes whole, local is undone once, and
// a callback that returns leaves $@ alone. What is expected is what perl
// 5.36 on Linux gives with the same module.
func TestNativeCallbackWithoutEvalIsPerls(t *testing.T) {
	so := buildModuleSource(t, "Callback", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* call(code): calls code without G_EVAL, as Scope::Cleanup runs a cleanup */
XS(XS_Callback_call) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    PUSHMARK(SP);
    PUTBACK;
    call_sv(ST(0), G_VOID | G_DISCARD);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_Callback) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("Callback::call", XS_Callback_call, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my @got;
{
    my $n = 0;
    local $SIG{__DIE__} = sub { $n++ };
    eval { Callback::call(sub { die "x\n" }) };
    push @got, "dies seen $n, error " . ($@ =~ s/\n\z//r);
}
sub outer { my @c; Callback::call(sub { @c = ((caller(0))[3], (caller(1))[3]) }); join "|", @c }
push @got, outer();
{
    my $obj = bless {}, 'E';
    eval { Callback::call(sub { die $obj }) };
    push @got, ref($@) . (($@ == $obj) ? " same" : " other");
}
our $g = 1;
eval { local $g = 2; Callback::call(sub { die "y\n" }) };
push @got, "g $g";
{
    eval { die "before\n" };
    Callback::call(sub { 1 });
    push @got, "errsv " . ($@ =~ s/\n\z//r);
}
join ",", @got`
	if got, want := requireModule(t, "Callback", so, program), "dies seen 1, error x,main::__ANON__|main::outer,E same,g 1,errsv before"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
