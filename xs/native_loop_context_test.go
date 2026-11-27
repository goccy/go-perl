//go:build darwin || linux

package xs_test

import "testing"

// Native code reads a loop context as perl's struct block_loop has it, as
// Future::AsyncAwait does when it suspends a foreach: a counting loop's
// state (state_u.lazyiv cur and end), that its iteration variable is a
// lexical's pad slot (CxPADLOOP), and the value the variable had before
// (itersave); an eval context's JMPENV is the interpreter's top one inside
// it; PL_scopestack_ix is a store; CLEAR_ERRSV empties $@. A STRLEN is put
// back by save_strlen; save_freepv frees a string of the module's and one
// of the interpreter's, also when the entry is taken off the save stack
// and pushed again (Future::AsyncAwait's suspend and resume);
// PL_restartop and PL_restartjmpenv are NULL with no die going on; PL_curpm
// is the last match, and storing it back keeps $1. Before, struct
// block_loop had the loop op alone, and none of the other names compiled.
const loopContextSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* loop(): the innermost loop context, as the words "cur end pad saved" */
XS(XS_MyLc_loop) {
    dXSARGS;
    I32 ix;
    PERL_UNUSED_VAR(items);
    for (ix = cxstack_ix; ix >= 0; ix--) {
        const PERL_CONTEXT *cx = &cxstack[ix];
        if (CxTYPE(cx) == CXt_LOOP_LAZYIV) {
            ST(0) = sv_2mortal(newSVpvf("%d %d %s %s", (int)cx->blk_loop.state_u.lazyiv.cur,
                                        (int)cx->blk_loop.state_u.lazyiv.end,
                                        CxPADLOOP(cx) ? "pad" : "glob",
                                        cx->blk_loop.itersave ? "saved" : "none"));
            XSRETURN(1);
        }
    }
    ST(0) = sv_2mortal(newSVpvs("no loop"));
    XSRETURN(1);
}
/* in_eval(): whether the innermost eval context was entered under the
 * interpreter's top JMPENV */
XS(XS_MyLc_in_eval) {
    dXSARGS;
    I32 ix;
    PERL_UNUSED_VAR(items);
    for (ix = cxstack_ix; ix >= 0; ix--) {
        const PERL_CONTEXT *cx = &cxstack[ix];
        if (CxTYPE(cx) == CXt_EVAL) {
            ST(0) = sv_2mortal(newSViv(cx->blk_eval.cur_top_env == PL_top_env ? 1 : 0));
            XSRETURN(1);
        }
    }
    ST(0) = sv_2mortal(newSViv(-1));
    XSRETURN(1);
}
/* scopes(): PL_scopestack_ix moved down and back, and $@ emptied */
XS(XS_MyLc_scopes) {
    dXSARGS;
    const I32 was = PL_scopestack_ix;
    int back;
    PERL_UNUSED_VAR(items);
    PL_scopestack_ix = was - 1;
    PL_scopestack_ix = was;
    back = PL_scopestack_ix == was;
    sv_setpvs(ERRSV, "x");
    CLEAR_ERRSV();
    ST(0) = sv_2mortal(newSVpvf("back %d errsv [%s]", back, SvPV_nolen(ERRSV)));
    XSRETURN(1);
}
/* saves(): what the saves left, PL_restartop and PL_restartjmpenv, and
 * PL_curpm stored back */
static STRLEN lc_len;
XS(XS_MyLc_saves) {
    dXSARGS;
    char *own = savepv("own");
    SV *sv = sv_2mortal(newSVpvs("theirs"));
    char *theirs = SvPVX(sv);
    char *pv;
    PMOP *const was = PL_curpm;
    PERL_UNUSED_VAR(items);
    SvPV_set(sv, NULL);
    SvLEN_set(sv, 0);
    SvCUR_set(sv, 0);
    SvPOK_off(sv);
    lc_len = 5;
    ENTER;
    Perl_save_strlen(aTHX_ &lc_len);
    lc_len = 9;
    save_freepv(own);
    save_freepv(theirs);
    PL_savestack_ix -= 2;
    pv = (char *)PL_savestack[PL_savestack_ix].any_ptr;
    save_freepv(pv);
    LEAVE;
    PL_curpm = NULL;
    PL_curpm = was;
    ST(0) = sv_2mortal(newSVpvf("len %d restart %s %s curpm %s", (int)lc_len,
                                PL_restartop ? "set" : "null",
                                PL_restartjmpenv ? "set" : "null", was ? "set" : "null"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyLc) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyLc::loop", XS_MyLc_loop);
    newXS_deffile("MyLc::in_eval", XS_MyLc_in_eval);
    newXS_deffile("MyLc::scopes", XS_MyLc_scopes);
    newXS_deffile("MyLc::saves", XS_MyLc_saves);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeReadsALoopContext(t *testing.T) {
	so := buildModuleSource(t, "MyLc", loopContextSource)
	got := requireModule(t, "MyLc", so, `my @l;
for my $i (3 .. 7) { push @l, MyLc::loop() if $i == 5 }
my $e = eval { MyLc::in_eval() };
"ab" =~ /(b)/;
my $s = MyLc::saves();
join "; ", @l, "eval $e", MyLc::scopes(), $s, "match $1"`)
	if want := "6 7 pad saved; eval 1; back 1 errsv []; len 5 restart null null curpm set; match b"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
