//go:build darwin || linux

package xs_test

import "testing"

// Raw words native code pushes with SSPUSH* right before SAVEDESTRUCTOR_X
// are that destructor's data, which it pops itself, as Data::Alias has
// them: the save stack is read as perl's leave_scope reads it, which never
// takes them for a group. The words here are chosen to look like groups:
// a type word of SAVEt_ALLOC claiming a thousand words below it, and one
// of SAVEt_INT with a word below it that is no address. They reach the
// destructor as they were pushed, also when another native call reads the
// save stack while they are on it. Before, the view took the top raw word
// for a type word and croaked, and the croak's own flush walked the view
// again, until the native stack overflowed (SIGSEGV); with the bits of an
// address, which word misread depended on where it was allocated.
const ssRawWordsSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static SV *popped;
/* pops its three raw words */
static void my_dtor(pTHX_ void *p) {
    const UV alloc = SSPOPUV;
    const UV type = SSPOPUV;
    const IV value = SSPOPIV;
    PERL_UNUSED_VAR(p);
    sv_setpvf(popped, "%" IVdf " %d %d", value, type == (UV)SAVEt_INT,
              alloc == (((UV)1000 << SAVE_TIGHT_SHIFT) | SAVEt_ALLOC));
}
/* raw(code): three raw words and the destructor that pops them; code runs
 * while they are on the save stack */
XS(XS_MyRw_raw) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    if (!popped) popped = newSVpvs("");
    sv_setpvs(popped, "not run");
    SSCHECK(3);
    SSPUSHIV(-7);
    SSPUSHUV((UV)SAVEt_INT);
    SSPUSHUV(((UV)1000 << SAVE_TIGHT_SHIFT) | SAVEt_ALLOC);
    SAVEDESTRUCTOR_X(my_dtor, NULL);
    PUSHMARK(SP);
    call_sv(ST(0), G_VOID | G_DISCARD);
    XSRETURN_EMPTY;
}
/* top(): whether the entry on top of the save stack is a destructor's (it
 * is not: call_sv's scope is above it); a native call of its own reads the
 * save stack while the raw words are on it */
XS(XS_MyRw_top) {
    dXSARGS;
    const UV t = PL_savestack[PL_savestack_ix - 1].any_uv & SAVE_MASK;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(t == SAVEt_DESTRUCTOR_X ? 1 : 0));
    XSRETURN(1);
}
XS(XS_MyRw_popped) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = popped ? popped : &PL_sv_undef;
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyRw) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyRw::raw", XS_MyRw_raw);
    newXS_deffile("MyRw::top", XS_MyRw_top);
    newXS_deffile("MyRw::popped", XS_MyRw_popped);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeRawSaveStackWordsAreTheDestructors(t *testing.T) {
	so := buildModuleSource(t, "MyRw", ssRawWordsSource)
	got := requireModule(t, "MyRw", so, `my @r;
MyRw::raw(sub { push @r, MyRw::top(), MyRw::popped() });
push @r, MyRw::popped();
MyRw::raw(sub { my @x = map { $_ * 2 } 1 .. 3 });
push @r, MyRw::popped();
join "|", @r`)
	if want := "0|not run|-7 1 1|-7 1 1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
