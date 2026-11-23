//go:build darwin || linux

package xs_test

import "testing"

// The host save stack grows as perl's save stack does. Every native call
// keeps three entries there while it runs, so native calls nested through
// perl code three hundred deep (Promise::XS recursing until its own limit
// of 254 croaks: t/infinite_recursion.t) ran out of the SDK's 256, and the
// croak of that came before the call could catch it, which aborted the
// process. A croak at the bottom unwinds all of them, putting back what each
// saved. What is expected is what perl 5.36 on Linux gives with the module.
func TestNativeHostSavesGrowAsPerls(t *testing.T) {
	so := buildModuleSource(t, "Deep", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int depth = -1;
/* Calls cb with cb and n - 1, for n down to 0: as many native calls nested
 * through perl code, each of which holds depth (a variable of the module's)
 * at its own n with SAVEINT while it runs. Gives the number of the calls
 * that found depth at their n again once the call below them returned; at
 * 0, croaks if asked to. */
XS(XS_Deep_through) {
    dXSARGS;
    SV *cb = ST(0);
    IV n = SvIV(ST(1)), r = 0;
    if (n == 0 && items > 2 && SvTRUE(ST(2))) croak("the bottom");
    if (n > 0) {
        ENTER;
        SAVETMPS;
        SAVEINT(depth);
        depth = (int)n;
        PUSHMARK(SP);
        XPUSHs(cb);
        XPUSHs(sv_2mortal(newSViv(n - 1)));
        if (items > 2) XPUSHs(ST(2));
        PUTBACK;
        (void)call_sv(cb, G_SCALAR);
        SPAGAIN;
        r = POPi + (depth == n ? 1 : 0);
        PUTBACK;
        FREETMPS;
        LEAVE;
    }
    ST(0) = sv_2mortal(newSViv(r));
    XSRETURN(1);
}
XS(XS_Deep_depth) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(depth));
    XSRETURN(1);
}
XS_EXTERNAL(boot_Deep) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("Deep::through", XS_Deep_through, __FILE__);
    newXS("Deep::depth", XS_Deep_depth, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `
my $cb; $cb = sub { Deep::through(@_) };
my $r = Deep::through($cb, 300);
my $died = eval { Deep::through($cb, 300, 1); 1 } ? "lived" : $@ =~ /^the bottom/ ? "died" : "died: $@";
join ",", $r, Deep::depth(), $died, Deep::depth()`
	if got, want := requireModule(t, "Deep", so, program), "300,-1,died,-1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
