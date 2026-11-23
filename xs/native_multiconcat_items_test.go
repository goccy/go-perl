//go:build darwin || linux

package xs_test

import "testing"

// The items of a multiconcat op, as Loo deparses them: the number of
// arguments, the constant string and the lengths of its segments, at the
// indexes of perl 5.44's header (three pad indexes come before the lengths
// there). Before, the op had no items (Loo's deparse ended the process) and
// the SDK's indexes were an older perl's. The expected string is what perl
// prints with the same module code.
const multiconcatItemsSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

static OP *MyCat_find(pTHX_ OP *o, int type) {
    OP *k;
    if (!o) return NULL;
    if (o->op_type == type) return o;
    if (o->op_flags & OPf_KIDS)
        for (k = cUNOPo->op_first; k; k = OpSIBLING(k)) {
            OP *r = MyCat_find(aTHX_ k, type);
            if (r) return r;
        }
    return NULL;
}

/* the arguments and the constant segments of the multiconcat op of a sub */
XS(XS_MyCat_concat) {
    dXSARGS;
    OP *m;
    if (items != 1) croak_xs_usage(cv, "code");
    m = MyCat_find(aTHX_ CvROOT((CV *)SvRV(ST(0))), OP_MULTICONCAT);
    if (!m) croak("no multiconcat op");
    {
        UNOP_AUX_item *const aux = cUNOP_AUXx(m)->op_aux;
        const SSize_t nargs = aux[PERL_MULTICONCAT_IX_NARGS].ssize;
        const char *pv = aux[PERL_MULTICONCAT_IX_PLAIN_PV].pv;
        SV *const out = sv_2mortal(newSVpvf("nargs=%d len=%d segs=", (int)nargs,
                                            (int)aux[PERL_MULTICONCAT_IX_PLAIN_LEN].ssize));
        SSize_t seg;
        for (seg = 0; seg <= nargs; seg++) {
            const SSize_t len = aux[PERL_MULTICONCAT_IX_LENGTHS + seg].ssize;
            if (seg) sv_catpvs(out, ",");
            if (len >= 0 && pv) {
                sv_catpvn(out, pv, len);
                pv += len;
            }
        }
        ST(0) = out;
    }
    XSRETURN(1);
}

XS_EXTERNAL(boot_MyCat) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyCat::concat", XS_MyCat_concat);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeItemsOfAMulticoncat(t *testing.T) {
	so := buildModuleSource(t, "MyCat", multiconcatItemsSource)
	got := requireModule(t, "MyCat", so, `join "|", MyCat::concat(sub { "<" . $_[0] . "|" . $_[1] . ">" }),
    MyCat::concat(sub { my $s = "x"; $s .= $_[0] . "-" . $_[1]; $s })`)
	want := "nargs=2 len=3 segs=<,|,>|nargs=2 len=1 segs=,-,"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
