//go:build darwin || linux

package xs_test

import "testing"

// Native code walks the SV arenas as Devel::Leak does: from
// PL_sv_arenaroot, the slots of an arena are sva + 1 up to
// &sva[SvREFCNT(sva)], a slot whose SvTYPE is SVTYPEMASK is free, and
// SvANY(sva) is the next arena. Counting the scalars in use before and
// after making some finds them (Search-Xapian's leak tests count so).
// Before, PL_sv_arenaroot died saying go-perl gives no access to it (and
// before that, it was NULL and a walk counted nothing).
func TestNativeWalkOfTheSVArenas(t *testing.T) {
	so := buildModuleSource(t, "ArenaWalk", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static IV in_use(pTHX_ IV *arenas, SV *find, int *found) {
    SV *sva;
    IV n = 0;
    *arenas = 0;
    for (sva = PL_sv_arenaroot; sva; sva = (SV *)SvANY(sva)) {
        SV *sv = sva + 1;
        SV *svend = &sva[SvREFCNT(sva)];
        (*arenas)++;
        while (sv < svend) {
            if (SvTYPE(sv) != SVTYPEMASK) {
                n++;
                if (sv == find) *found = 1;
            }
            ++sv;
        }
    }
    return n;
}
XS(XS_ArenaWalk_count) {
    dXSARGS;
    IV arenas;
    int found = 0;
    IV n = in_use(aTHX_ &arenas, items > 0 && SvROK(ST(0)) ? SvRV(ST(0)) : NULL, &found);
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%" IVdf " %" IVdf " %d", n, arenas, found));
    XSRETURN(1);
}
XS_EXTERNAL(boot_ArenaWalk) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("ArenaWalk::count", XS_ArenaWalk_count);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	got := requireModule(t, "ArenaWalk", so, `my ($before) = split / /, ArenaWalk::count();
my @made = map { "scalar $_" } 1 .. 500;
my $mine = "find me";
my ($after, $arenas, $found) = split / /, ArenaWalk::count(\$mine);
my $grew = $after - $before;
($grew >= 500 && $grew < 1500 ? "grew by the scalars made" : "grew by $grew") . ", " . ($arenas >= 1 ? "arenas" : "no arenas") . ", " . ($found ? "found mine" : "not found")`)
	if want := "grew by the scalars made, arenas, found mine"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
