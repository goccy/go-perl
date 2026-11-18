//go:build darwin || linux

package xs_test

import "testing"

// A pp function of native code works on the interpreter's own stack, as
// perl's pp functions do: deep in a long list (a map over thousands of
// values) it sees the whole stack below it at the interpreter's indices,
// PL_markstack_ptr - PL_markstack is the interpreter's index of the top
// mark (Future::AsyncAwait counts the marks of a context so), EXTEND grows
// the stack by thousands of values and a pp function pushes a hundred
// marks, more than a copy held. Before, the stack was a copy handed over
// for every op (the whole list, again for each value) and given back
// after it, its marks the top 32 of the interpreter's, and a pp function
// that pushed more marks died.
const ppZeroCopySource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static Perl_ppaddr_t orig_chr;
static IV max_depth, top_ok = 1, marks_ok = 1, calls;
static OP *my_chr(pTHX) {
    dSP;
    const SSize_t depth = SP - PL_stack_base;
    const SSize_t marks = PL_markstack_ptr - PL_markstack;
    int i;
    calls++;
    if (depth > max_depth) max_depth = depth;
    if (PL_stack_base[depth] != TOPs) top_ok = 0;
    if (marks < 1 || TOPMARK > depth) marks_ok = 0;
    EXTEND(SP, 5000);
    for (i = 0; i < 5000; i++) PUSHs(&PL_sv_undef);
    for (i = 0; i < 100; i++) PUSHMARK(SP);
    for (i = 0; i < 100; i++) (void)POPMARK;
    if (PL_markstack_ptr - PL_markstack != marks) marks_ok = 0;
    SP -= 5000;
    PUTBACK;
    return orig_chr(aTHX);
}
XS(XS_MyZc_hook) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    orig_chr = PL_ppaddr[OP_CHR];
    PL_ppaddr[OP_CHR] = my_chr;
    XSRETURN_EMPTY;
}
XS(XS_MyZc_unhook) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    PL_ppaddr[OP_CHR] = orig_chr;
    ST(0) = sv_2mortal(newSVpvf("calls %d depth %s top %d marks %d", (int)calls,
                                max_depth >= 3000 ? "deep" : "shallow", (int)top_ok, (int)marks_ok));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyZc) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyZc::hook", XS_MyZc_hook);
    newXS_deffile("MyZc::unhook", XS_MyZc_unhook);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativePPFunctionWorksOnTheInterpretersStack(t *testing.T) {
	so := buildModuleSource(t, "MyZc", ppZeroCopySource)
	got := requireModule(t, "MyZc", so, `MyZc::hook();
my @r = eval q{ my @big = (1 .. 3000); map { chr(65 + $_ % 26) } @big };
die $@ if $@;
my $s = MyZc::unhook();
join "; ", scalar(@r), join("", @r[0 .. 4]), $s`)
	if want := "3000; BCDEF; calls 3000 depth deep top 1 marks 1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
