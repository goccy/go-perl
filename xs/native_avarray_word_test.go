//go:build darwin || linux

package xs_test

import "testing"

// AvARRAY is an lvalue, and a word native code keeps in it that is no body
// of the array's is what it reads back in a later call, as on perl:
// Math::Pari chains its objects through the slot (the previous object's
// SV there, the level in AvFILLp), and empties the array before perl frees
// it. Before, AvARRAY was no lvalue, and a store through its slot failed.
// The body of an array whose elements were all shifted off ends where its
// block does, which is the array's body still: AvARRAY is where av_unshift
// puts the next element, one slot back.
const avarrayWordSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* keep(\\@a, \\$x): $x's SV in @a's AvARRAY, the level 3 in AvFILLp */
XS(XS_MyAw_keep) {
    dXSARGS;
    AV *const av = (AV *)SvRV(ST(0));
    PERL_UNUSED_VAR(items);
    AvARRAY(av) = (SV **)SvRV(ST(1));
    AvFILLp(av) = 3;
    XSRETURN_EMPTY;
}
/* kept(\\@a, \\$x): whether the slot still holds $x's SV and the level, and
 * the array emptied again */
XS(XS_MyAw_kept) {
    dXSARGS;
    AV *const av = (AV *)SvRV(ST(0));
    const int same = (SV *)AvARRAY(av) == SvRV(ST(1)) && AvFILLp(av) == 3;
    PERL_UNUSED_VAR(items);
    AvFILLp(av) = -1;
    AvARRAY(av) = NULL;
    ST(0) = sv_2mortal(newSVpv(same ? "kept" : "lost", 0));
    XSRETURN(1);
}
/* front(\\@a) of an array shifted empty: whether AvARRAY is its body, the
 * slot after the one av_unshift then gives */
XS(XS_MyAw_front) {
    dXSARGS;
    AV *const av = (AV *)SvRV(ST(0));
    SV **const before = AvARRAY(av);
    const SSize_t fill = AvFILLp(av);
    PERL_UNUSED_VAR(items);
    av_unshift(av, 1);
    ST(0) = sv_2mortal(newSVpv(fill == -1 && AvARRAY(av) + 1 == before ? "body" : "word", 0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyAw) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyAw::keep", XS_MyAw_keep);
    newXS_deffile("MyAw::kept", XS_MyAw_kept);
    newXS_deffile("MyAw::front", XS_MyAw_front);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeKeepsAWordInAvARRAY(t *testing.T) {
	so := buildModuleSource(t, "MyAw", avarrayWordSource)
	got := requireModule(t, "MyAw", so, `my $x = 7;
my $r = do { my @a; MyAw::keep(\@a, \$x); my $k = MyAw::kept(\@a, \$x); push @a, 1, 2; "$k " . scalar(@a) };
my @word = grep { my $b = [1 .. $_]; shift @$b for 1 .. $_; MyAw::front($b) ne "body" } 1 .. 24;
"$r $x [@word]"`)
	if want := "kept 2 7 []"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
