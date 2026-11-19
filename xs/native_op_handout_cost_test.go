//go:build darwin || linux

package xs_test

import (
	"strings"
	"testing"
)

// A native call that looks at PL_op (dXSTARG, GIMME_V) does work that does
// not grow with the program around the running op: once a tree is
// described, its descriptions last, and the baselines a hand-out gives
// (what a write into a shadow is told by) stop at the ops of finished
// trees. Before, every such call gave a baseline to every op the running
// op leads to (the whole program: 22034 shadows at 2000 statements), so a
// call took longer the bigger the program was. The module counts the
// shadows with a baseline after it looked at PL_op, an SDK internal: this
// is go-perl's own cost, which perl has no counterpart of.
const opHandoutCostSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

XS(XS_MyCost_shadows) {
    dXSARGS;
    int n = 0;
    goperl_opreg_ent_t *e;
    PERL_UNUSED_VAR(items);
    (void)PL_op->op_private;
    for (e = goperl_opreg_v ? goperl_opreg_v->active : 0; e; e = e->active_next) n++;
    ST(0) = sv_2mortal(newSViv(n));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyCost) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyCost::shadows", XS_MyCost_shadows);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCallCostDoesNotGrowWithTheProgram(t *testing.T) {
	so := buildModuleSource(t, "MyCost", opHandoutCostSource)
	// the counts of the calls after the first (which describes the tree),
	// from a loop in the program and from a loop in a sub
	counts := func(statements int) string {
		never := strings.Repeat("if ($never) { $never++ } ", statements)
		return requireModule(t, "MyCost", so, `my $r = eval q{
our $never;
my (@main, @sub);
for my $i (1..3) { push @main, MyCost::shadows() }
`+never+`
sub big { my @r; for my $i (1..3) { push @r, MyCost::shadows() } `+never+` @r }
@sub = big();
join "|", join(",", @main[1, 2]), join(",", @sub[1, 2]);
}; die $@ if $@; $r`)
	}
	small, big := counts(10), counts(2000)
	if small != big {
		t.Errorf("shadows given baselines per call: %q with 10 statements, %q with 2000", small, big)
	}
	for _, part := range strings.FieldsFunc(big, func(r rune) bool { return r == '|' || r == ',' }) {
		if len(part) > 2 {
			t.Errorf("a call gave %s shadows baselines (%q)", part, big)
		}
	}
}
