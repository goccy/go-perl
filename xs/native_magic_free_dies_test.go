//go:build darwin || linux

package xs_test

import (
	"regexp"
	"testing"
)

// A svt_free of native code that dies: the die reaches the program, as
// perl's mg_free lets it out, and a MAGIC the hook took off the value's
// chain and freed itself before it died is not freed again. That is what
// Variable::Magic's free does when its callback dies (vmg_svt_free's
// cleanup: it takes its MAGIC off, frees it, has the value freed once more
// for the rest of the chain, and dies). Before, the die was printed and
// dropped, and the MAGIC freed a second time (free(): double free, abort).
func TestNativeMagicFreeThatDies(t *testing.T) {
	so := buildModuleSource(t, "MagicDie", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int freed;
static int die_free(pTHX_ SV *sv, MAGIC *mg) {
    freed++;
    switch (mg->mg_private) {
    case 1:
        croak("died in free");
    case 2: {
        /* Variable::Magic's way: the MAGIC off and freed, the value freed
         * once more for the rest of its chain, then the die */
        SvREFCNT_inc_simple_void(sv);
        if (SvMAGIC(sv) == mg) SvMAGIC_set(sv, mg->mg_moremagic);
        Safefree(mg);
        SvREFCNT_dec(sv);
        croak("died in free, its magic gone");
    }
    }
    return 0;
}
static MGVTBL die_vtbl = { NULL, NULL, NULL, NULL, die_free, NULL, NULL, NULL };
XS(XS_MagicDie_cast) {
    dXSARGS;
    MAGIC *mg;
    PERL_UNUSED_VAR(items);
    mg = sv_magicext(SvRV(ST(0)), NULL, PERL_MAGIC_ext, &die_vtbl, NULL, 0);
    mg->mg_private = (U16)SvIV(ST(1));
    XSRETURN_EMPTY;
}
XS(XS_MagicDie_freed) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_IV(freed);
}
XS_EXTERNAL(boot_MagicDie) {
    dXSBOOTARGSXSAPIVERCHK;
    freed = 0; /* this interpreter's */
    newXS_deffile("MagicDie::cast", XS_MagicDie_cast);
    newXS_deffile("MagicDie::freed", XS_MagicDie_freed);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	for _, tc := range []struct {
		name, mode string
		want       *regexp.Regexp
	}{
		{"a free that returns", "0", regexp.MustCompile(`\Alived, freed 1; then 2,4,6\z`)},
		{"a free that dies", "1", regexp.MustCompile(`\Adied in free at .* line \d+\.\n, freed 1; then 2,4,6\z`)},
		{"a free that takes its magic off and dies", "2", regexp.MustCompile(`\Adied in free, its magic gone at .* line \d+\.\n, freed 1; then 2,4,6\z`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := requireModule(t, "MagicDie", so, `my $r = eval { my $x = 1; MagicDie::cast(\$x, `+tc.mode+`); 1 } ? "lived" : $@;
my @a = (1, 2, 3);
"$r, freed " . MagicDie::freed() . "; then " . join(",", map { $_ * 2 } @a)`)
			if !tc.want.MatchString(got) {
				t.Errorf("got %q, want %v", got, tc.want)
			}
		})
	}
}
