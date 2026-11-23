//go:build darwin || linux

package xs_test

import "testing"

// MULTICALL from the pp function of a custom op, as Heap::PQ calls its
// comparator: the calls run on a stackinfo of the interpreter's pushed for
// them, each one's value is at the top of that stack, and the pp function
// is back on its own stack after POP_MULTICALL. Before, the SDK refused
// MULTICALL in a pp function. The expected string is what perl prints with
// the same module code.
const multicallPPSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

/* the pp function a call of MyMcp::fold() is compiled to: $MyMcp::block
 * called for every element of @MyMcp::vals as $_, the values it returns
 * joined; and the number of values on the stack before and after */
static OP *MyMcp_pp_fold(pTHX) {
    dSP;
    AV *const vals = get_av("MyMcp::vals", 0);
    CV *const block = (CV *)SvRV(get_sv("MyMcp::block", 0));
    SV *const out = sv_2mortal(newSVpvs(""));
    const SSize_t depth = SP - PL_stack_base;
    SSize_t i;
    {
        dMULTICALL;
        I32 gimme = G_SCALAR;
        SAVESPTR(GvSV(PL_defgv));
        PUSH_MULTICALL(block);
        for (i = 0; i <= av_len(vals); i++) {
            GvSV(PL_defgv) = *av_fetch(vals, i, 0);
            MULTICALL;
            if (i) sv_catpvs(out, ",");
            sv_catsv(out, *PL_stack_sp);
        }
        POP_MULTICALL;
    }
    SPAGAIN;
    sv_catpvf(out, " depth %s", SP - PL_stack_base == depth ? "kept" : "moved");
    XPUSHs(out);
    RETURN;
}
static OP *MyMcp_ck(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    OP *o;
    PERL_UNUSED_ARG(namegv);
    PERL_UNUSED_ARG(ckobj);
    op_free(entersubop);
    o = newOP(OP_CUSTOM, 0);
    o->op_ppaddr = MyMcp_pp_fold;
    return o;
}
XS(XS_MyMcp_fold) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    croak("MyMcp::fold was not compiled to its op");
}
XS(XS_MyMcp_install) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    cv_set_call_checker(get_cv("MyMcp::fold", 0), MyMcp_ck, &PL_sv_undef);
    XSRETURN_EMPTY;
}

XS_EXTERNAL(boot_MyMcp) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyMcp::fold", XS_MyMcp_fold);
    newXS_deffile("MyMcp::install", XS_MyMcp_install);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeMulticallFromAPPFunction(t *testing.T) {
	so := buildModuleSource(t, "MyMcp", multicallPPSource)
	got := requireModule(t, "MyMcp", so, `MyMcp::install();
my $f = eval q{ sub { local $MyMcp::block = shift; local @MyMcp::vals = @_; my @l = ("a", MyMcp::fold(), "z"); "@l" } }
    or die $@;
$_ = "outer";
join "|", $f->(sub { $_ * 2 }, 1, 2, 3), $f->(sub { my $x = uc; "<$x>" }, "p", "q"), $_`)
	want := "a 2,4,6 depth kept z|a <P>,<Q> depth kept z|outer"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
