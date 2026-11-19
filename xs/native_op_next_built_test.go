//go:build darwin || linux

package xs_test

import "testing"

// A sub that native code builds of ops it made itself, as
// XS::Parse::Sublike builds one with an empty signature: the ops' op_next,
// read in the same native call after newATTRSUB linked them, lead from
// CvSTART through the sub as perl's do (Syntax::Keyword::MultiSub walks
// them for the signature's op). Before, the first nextstate's op_next was
// the op itself, and the walk never ended.
const opNextBuiltSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* build(): the names of the ops from CvSTART of an anonymous sub of
 * natively made ops, by op_next, and a reference to the sub */
XS(XS_MyOn_build) {
    dXSARGS;
    char buf[256];
    size_t n = 0;
    I32 floor_ix;
    OP *sigop, *body, *o;
    CV *sub;
    int i;
    struct op_argcheck_aux *aux;
    PERL_UNUSED_VAR(items);
    ENTER;
    floor_ix = start_subparse(FALSE, CVf_ANON);
    aux = (struct op_argcheck_aux *)PerlMemShared_malloc(sizeof *aux);
    aux->params = 0;
    aux->opt_params = 0;
    aux->slurpy = 0;
    sigop = op_prepend_elem(OP_LINESEQ, newSTATEOP(0, NULL, NULL),
                            newUNOP_AUX(OP_ARGCHECK, 0, NULL, (UNOP_AUX_item *)aux));
    sigop = op_append_elem(OP_LINESEQ, sigop, newSTATEOP(0, NULL, NULL));
    sigop = newUNOP_AUX(OP_ARGCHECK, 0, sigop, NULL);
    op_null(sigop);
    body = op_append_list(OP_LINESEQ, sigop,
                          newSTATEOP(0, NULL, newSVOP(OP_CONST, 0, newSVpvs("one"))));
    sub = newATTRSUB(floor_ix, NULL, NULL, NULL, body);
    buf[0] = 0;
    for (o = CvSTART(sub), i = 0; o && i < 10; o = o->op_next, i++)
        n += snprintf(buf + n, sizeof buf - n, "%s%s", i ? " " : "", OP_NAME(o));
    LEAVE;
    ST(0) = sv_2mortal(newSVpv(buf, 0));
    ST(1) = sv_2mortal(newRV_noinc((SV *)sub));
    XSRETURN(2);
}
XS_EXTERNAL(boot_MyOn) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyOn::build", XS_MyOn_build);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeWalksTheOpsOfASubItBuilt(t *testing.T) {
	so := buildModuleSource(t, "MyOn", opNextBuiltSource)
	got := requireModule(t, "MyOn", so, `my ($names, $cv);
eval q{ BEGIN { ($names, $cv) = MyOn::build() } 1 } or die $@;
join "|", $names, $cv->()`)
	if want := "nextstate argcheck nextstate const leavesub|one"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
