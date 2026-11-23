//go:build darwin || linux

package xs_test

import "testing"

// A destructor entry native code takes off the save stack and pushes again
// with SAVEDESTRUCTOR_X, from the function and argument the entry holds, is
// the same destructor: it runs once, with its argument, when the scope
// ends, as Future::AsyncAwait relies on when it suspends a sub with
// Syntax::Keyword::Try's finally block in scope and resumes it. Before, the
// function the entry holds was taken for one of native code's, and the
// call of it crashed.
const destructorPutbackSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int runs;
static void count_dtor(pTHX_ void *p) {
    PERL_UNUSED_CONTEXT;
    runs += (int)(IV)p;
}
XS(XS_MyDp_run) {
    dXSARGS;
    void (*fn)(pTHX_ void *);
    void *arg;
    int before;
    PERL_UNUSED_VAR(items);
    runs = 0;
    ENTER;
    SAVEDESTRUCTOR_X(count_dtor, (void *)(IV)5);
    PL_savestack_ix -= 3;
    fn = PL_savestack[PL_savestack_ix].any_dxptr;
    arg = PL_savestack[PL_savestack_ix + 1].any_ptr;
    SAVEDESTRUCTOR_X(fn, arg);
    before = runs;
    LEAVE;
    ST(0) = sv_2mortal(newSVpvf("before %d after %d", before, runs));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyDp) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyDp::run", XS_MyDp_run);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodePutsADestructorEntryBack(t *testing.T) {
	so := buildModuleSource(t, "MyDp", destructorPutbackSource)
	got := requireModule(t, "MyDp", so, `MyDp::run()`)
	if want := "before 0 after 5"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
