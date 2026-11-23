//go:build darwin || linux

package xs_test

import "testing"

// The SDK's direct accessors (the name gv_stashsv reads, newSVpvn, ...)
// give the interpreter the XSUB's stack and top mark before they run, as
// every operation does. They take them back after, since Perl code may run
// in them (an overloaded "" of the name) and grow the stack. Before, they
// took back what the interpreter had before the XSUB's last moves: the mark
// dXSARGS popped came back as if it had not been popped, and the caller
// popped the wrong mark, so the values before the call were lost (a list
// with a constructor that blesses into gv_stashsv(ST(0), ...)). And Perl
// code run inside the accessor wrote over what the XSUB had pushed and put
// back.
const directAccessorStackSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* new(class): a blessed hash, the stash found by the name */
XS(XS_MyDa_new) {
    dXSARGS;
    HV *const stash = gv_stashsv(ST(0), GV_ADD);
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(sv_bless(newRV_noinc((SV *)newHV()), stash));
    XSRETURN(1);
}
/* around(name): 1, 2, the stash's name, 3; the name is read after 1 and 2
 * are pushed and put back */
XS(XS_MyDa_around) {
    dXSARGS;
    SV *const name = ST(0);
    HV *stash;
    PERL_UNUSED_VAR(items);
    SP -= items;
    XPUSHs(sv_2mortal(newSViv(1)));
    XPUSHs(sv_2mortal(newSViv(2)));
    PUTBACK;
    stash = gv_stashsv(name, 0);
    SPAGAIN;
    XPUSHs(sv_2mortal(newSVpv(stash ? HvNAME(stash) : "none", 0)));
    XPUSHs(sv_2mortal(newSViv(3)));
    PUTBACK;
}
XS_EXTERNAL(boot_MyDa) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyDa::new", XS_MyDa_new);
    newXS_deffile("MyDa::around", XS_MyDa_around);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeDirectAccessorKeepsTheStack(t *testing.T) {
	so := buildModuleSource(t, "MyDa", directAccessorStackSource)
	got := requireModule(t, "MyDa", so, `{ package Obj; sub x {} }
{ package Name; use overload '""' => sub { my @w = map { $_ * 2 } 1 .. 200; "Obj" } }
my @a = ("a", MyDa::new("Obj"), "b", MyDa::new("Obj"), "c");
my @b = (9, MyDa::around(bless {}, "Name"), 8);
join " ", scalar(@a), (map { ref($_) || $_ } @a), "|", @b`)
	if want := "5 a Obj b Obj c | 9 1 2 Obj 3 8"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
