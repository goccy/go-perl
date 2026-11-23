//go:build darwin || linux

package xs_test

import "testing"

// Native code reads the save stack's entries in its own pointers, as
// Future::AsyncAwait does when it suspends an async sub and puts the
// entries back or runs them: the location SAVESPTR named is the pointer
// native code gave (a slot of an array's body here, SAVESPTR(PAD_SVl(ix))
// in Future::AsyncAwait), the string of a SAVEt_FREEPV entry is the string,
// and the function of a SAVEDESTRUCTOR_X entry can be called with its
// argument, which runs the destructor once: taken off the save stack and
// called, it does not run again when the scope ends. Entries taken off and
// pushed again with save_pushptrptr are the same saves. Before, the entries
// were the interpreter's words: the location and the string were its
// addresses, and calling the function crashed.
const saveStackViewSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int runs;
static void count_dtor(pTHX_ void *p) {
    PERL_UNUSED_CONTEXT;
    runs += (int)(IV)p;
}
/* sptr(\\@a): SAVESPTR of the array's first slot, as the entry names it */
XS(XS_MySv_sptr) {
    dXSARGS;
    AV *const av = (AV *)SvRV(ST(0));
    SV **const slot = av_fetch(av, 0, 1);
    SV *const was = *slot;
    I32 ix;
    const char *loc, *val;
    PERL_UNUSED_VAR(items);
    ENTER;
    SAVESPTR(*slot);
    ix = PL_savestack_ix;
    loc = PL_savestack[ix - 2].any_ptr == (void *)slot ? "same" : "other";
    val = PL_savestack[ix - 3].any_ptr == (void *)was ? "same" : "other";
    *slot = newSVpvs("inside");
    LEAVE;
    ST(0) = sv_2mortal(newSVpvf("loc %s val %s back %s", loc, val, SvPV_nolen(*av_fetch(av, 0, 0))));
    XSRETURN(1);
}
/* freepv(): the string of a SAVEt_FREEPV entry */
XS(XS_MySv_freepv) {
    dXSARGS;
    SV *const sv = sv_2mortal(newSVpvs("theirs"));
    char *const pv = SvPVX(sv);
    I32 ix;
    const char *same;
    PERL_UNUSED_VAR(items);
    SvPV_set(sv, NULL);
    SvLEN_set(sv, 0);
    SvCUR_set(sv, 0);
    SvPOK_off(sv);
    ENTER;
    save_freepv(pv);
    ix = PL_savestack_ix;
    same = PL_savestack[ix - 2].any_ptr == (void *)pv ? "same" : "other";
    LEAVE;
    ST(0) = sv_2mortal(newSVpvf("pv %s", same));
    XSRETURN(1);
}
/* dtor(): a destructor entry's function called with its argument after the
 * entry is taken off, and one taken off and pushed again with
 * save_pushptrptr */
XS(XS_MySv_dtor) {
    dXSARGS;
    void (*fn)(pTHX_ void *);
    void *arg;
    int called, after, pushed;
    PERL_UNUSED_VAR(items);
    runs = 0;
    ENTER;
    SAVEDESTRUCTOR_X(count_dtor, (void *)(IV)5);
    PL_savestack_ix -= 3;
    fn = PL_savestack[PL_savestack_ix].any_dxptr;
    arg = PL_savestack[PL_savestack_ix + 1].any_ptr;
    (*fn)(aTHX_ arg);
    called = runs;
    LEAVE;
    after = runs;
    runs = 0;
    ENTER;
    SAVEDESTRUCTOR_X(count_dtor, (void *)(IV)7);
    PL_savestack_ix -= 3;
    fn = PL_savestack[PL_savestack_ix].any_dxptr;
    arg = PL_savestack[PL_savestack_ix + 1].any_ptr;
    save_pushptrptr((void *)fn, arg, SAVEt_DESTRUCTOR_X);
    LEAVE;
    pushed = runs;
    ST(0) = sv_2mortal(newSVpvf("called %d after %d pushed %d", called, after, pushed));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MySv) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySv::sptr", XS_MySv_sptr);
    newXS_deffile("MySv::freepv", XS_MySv_freepv);
    newXS_deffile("MySv::dtor", XS_MySv_dtor);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeReadsTheSaveStackInItsOwnPointers(t *testing.T) {
	so := buildModuleSource(t, "MySv", saveStackViewSource)
	got := requireModule(t, "MySv", so, `my @a = ("before");
join "; ", MySv::sptr(\@a), MySv::freepv(), MySv::dtor()`)
	if want := "loc same val same back before; pv same; called 5 after 5 pushed 7"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
