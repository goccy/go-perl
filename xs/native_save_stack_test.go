//go:build darwin || linux

package xs_test

import "testing"

// Native code reads and writes the save stack and the temporaries stack in
// place, as Future::AsyncAwait suspends an async sub (it takes a scope's
// entries off by type and puts them back later) and Data::Alias pushes
// entries of its own: an entry SAVEFREESV made is two words, the value and
// SAVEt_FREESV, below PL_savestack_ix; taking them off (PL_savestack_ix -=
// 2) and putting them back (save_pushptr, or dSS_ADD .. SS_ADD_END) leaves
// the scope to free the value as before; a mortal is the top of
// PL_tmps_stack, and taking it off (NULL in its slot, PL_tmps_ix down)
// makes it native code's own. Before, none of these names compiled.
const saveStackSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* entries(\\$v): SAVEFREESV of a reference more, read, taken off, put
 * back with save_pushptr, or pushed raw with dSS_ADD; the count is perl's
 * again after the scope */
XS(XS_MySs_entries) {
    dXSARGS;
    SV *sv = SvRV(ST(0));
    SV *out = sv_2mortal(newSVpvs(""));
    I32 ix;
    UV uv;
    void *p;
    U32 before = SvREFCNT(sv);
    PERL_UNUSED_VAR(items);
    ENTER;
    SAVEFREESV(SvREFCNT_inc(sv));
    ix = PL_savestack_ix;
    uv = PL_savestack[ix - 1].any_uv;
    p = PL_savestack[ix - 2].any_ptr;
    sv_catpvf(out, "type %s ptr %s held %d",
              (uv & SAVE_MASK) == SAVEt_FREESV ? "FREESV" : "other",
              p == (void *)sv ? "same" : "other", (int)(SvREFCNT(sv) - before));
    PL_savestack_ix -= 2;
    sv_catpvf(out, " off %d", (int)(PL_savestack_ix == ix - 2));
    save_pushptr(p, SAVEt_FREESV);
    sv_catpvf(out, " back %d", (int)(PL_savestack_ix == ix));
    LEAVE;
    sv_catpvf(out, " after %d", (int)(SvREFCNT(sv) - before));
    ENTER;
    {
        dSS_ADD;
        SS_ADD_PTR(SvREFCNT_inc(sv));
        SS_ADD_UV(SAVEt_FREESV);
        SS_ADD_END(2);
    }
    sv_catpvf(out, " raw %d", (int)(SvREFCNT(sv) - before));
    LEAVE;
    sv_catpvf(out, " freed %d", (int)(SvREFCNT(sv) - before));
    ST(0) = out;
    XSRETURN(1);
}
/* mortal(n): a mortal taken off the temporaries stack, then a mortal again */
XS(XS_MySs_mortal) {
    dXSARGS;
    SV *m = sv_2mortal(newSViv(SvIV(ST(0))));
    SSize_t ix = PL_tmps_ix;
    int top = PL_tmps_stack[ix] == m;
    PERL_UNUSED_VAR(items);
    PL_tmps_stack[ix] = NULL;
    PL_tmps_ix--;
    ST(0) = sv_2mortal(newSVpvf("top %d %s", top, PL_tmps_ix == ix - 1 ? "off" : "not off"));
    ST(1) = sv_2mortal(m);
    XSRETURN(2);
}
XS(XS_MySs_catch) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(CATCH_GET ? 1 : 0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MySs) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySs::entries", XS_MySs_entries);
    newXS_deffile("MySs::mortal", XS_MySs_mortal);
    newXS_deffile("MySs::catch", XS_MySs_catch);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeUsesTheSaveStack(t *testing.T) {
	so := buildModuleSource(t, "MySs", saveStackSource)
	got := requireModule(t, "MySs", so, `my $v = "x";
my $e = MySs::entries(\$v);
my ($m, $n) = MySs::mortal(42);
join "; ", $e, $m, $n, "catch " . MySs::catch()`)
	want := "type FREESV ptr same held 1 off 1 back 1 after 0 raw 1 freed 0; top 1 off; 42; catch 0"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
