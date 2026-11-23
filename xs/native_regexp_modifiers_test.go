//go:build darwin || linux

package xs_test

import "testing"

// Native code rebuilds a pattern's modifiers from the flags it was compiled
// with, as Sereal::Encoder does when it writes a qr//: the flags masked by
// RXf_PMf_COMPILETIME, one bit for each letter of INT_PAT_MODS from
// RXf_PMf_STD_PMMOD_SHIFT on, and a buffer sized by INT_PAT_MODS. Before,
// neither name compiled.
const regexpModifiersSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MyRm_mods) {
    dXSARGS;
    REGEXP *const re = SvRX(ST(0));
    const U32 fl = re ? RX_EXTFLAGS(re) & RXf_PMf_COMPILETIME : 0;
    const char *const m = INT_PAT_MODS;
    char buf[sizeof(INT_PAT_MODS)];
    STRLEN n = 0;
    int i;
    PERL_UNUSED_VAR(items);
    for (i = 0; m[i]; i++)
        if (fl & (1U << (RXf_PMf_STD_PMMOD_SHIFT + i))) buf[n++] = m[i];
    ST(0) = sv_2mortal(newSVpvn(buf, n));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyRm) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyRm::mods", XS_MyRm_mods);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeRebuildsPatternModifiers(t *testing.T) {
	so := buildModuleSource(t, "MyRm", regexpModifiersSource)
	got := requireModule(t, "MyRm", so, `join ",", map { "[" . MyRm::mods($_) . "]" } qr/a/, qr/a/i, qr/a/msixn, qr/a/xx, qr/a/sm`)
	if want := "[],[i],[msixn],[xx],[ms]"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
