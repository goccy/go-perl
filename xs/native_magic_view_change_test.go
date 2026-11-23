//go:build darwin || linux

package xs_test

import "testing"

// A module that looks at a value's chain of magic (SvMAGIC), attaches
// magic of its own with sv_magicext and looks at the chain again, as
// Variable::Magic's cast does (vmg_find, then vmg_mg_magical), keeps that
// magic on the value: its free runs when the value goes. Before, the
// first look at a value with magic of the interpreter's (a tie, the hints
// of %^H) was written back over the chain sv_magicext had changed, which
// took the new magic off again: B::Hooks::EndOfScope's callbacks never
// ran (B-Hooks-Parser, BEGIN-Lift, B-Hooks-EndOfScope).
func TestNativeMagicAttachedAfterALookAtTheChain(t *testing.T) {
	so := buildModuleSource(t, "MagicView", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int view_freed;
static int view_free(pTHX_ SV *sv, MAGIC *mg) {
    PERL_UNUSED_VAR(sv);
    PERL_UNUSED_VAR(mg);
    view_freed++;
    return 0;
}
static MGVTBL view_vtbl = { NULL, NULL, NULL, NULL, view_free, NULL, NULL, NULL };
XS(XS_MagicView_cast) {
    dXSARGS;
    SV *sv = SvRV(ST(0));
    MAGIC *mg;
    int n = 0;
    PERL_UNUSED_VAR(items);
    for (mg = SvMAGIC(sv); mg; mg = mg->mg_moremagic) n++;
    sv_magicext(sv, NULL, PERL_MAGIC_ext, &view_vtbl, NULL, 0);
    for (mg = SvMAGIC(sv); mg; mg = mg->mg_moremagic) n++;
    XSRETURN_IV(n);
}
XS(XS_MagicView_freed) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_IV(view_freed);
}
XS_EXTERNAL(boot_MagicView) {
    dXSBOOTARGSXSAPIVERCHK;
    view_freed = 0; /* this interpreter's */
    newXS_deffile("MagicView::cast", XS_MagicView_cast);
    newXS_deffile("MagicView::freed", XS_MagicView_freed);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	const tie = `package Kept; sub TIESCALAR { my $v; bless \$v } sub FETCH { ${$_[0]} } sub STORE { ${$_[0]} = $_[1] }
package main;
`
	for _, tc := range []struct{ name, setup string }{
		{"a value with no magic", `my $x = 1;`},
		{"a tied value", tie + `tie my $x, "Kept";`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := requireModule(t, "MagicView", so, `my $n; { `+tc.setup+` $n = MagicView::cast(\$x); } "seen:" . ($n > 0 ? "yes" : "no") . " freed:" . MagicView::freed()`)
			if want := "seen:yes freed:1"; got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
