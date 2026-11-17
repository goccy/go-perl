//go:build darwin || linux

package xs_test

import (
	"regexp"
	"testing"
)

// A read-only scalar whose buffer is native code's own memory (SvPV_set to
// a string of a library's, SvLEN 0) is a string perl reads: the interpreter
// gets a copy of the bytes when the scalar becomes a string. MarpaX-ESLIF
// makes it so in this order: SvPOK_only, SvPV_set, SvLEN_set(0),
// SvCUR_set, SvREADONLY_on; SWISH-3 sets the read-only flag first and the
// flag of a string last, and hands the scalar to the interpreter. A view
// that stays writable (File::Map's mapping, which perl writes through)
// still fails, and a structure kept behind a scalar that is no string is
// left where it is. Before, every view made a string was refused ("the
// string of a scalar cannot be in native memory").
func TestNativeReadOnlyViewOfNativeMemory(t *testing.T) {
	so := buildModuleSource(t, "PvView", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <string.h>
static char text[] = "the bytes of a library";
static int structure[4] = {1, 2, 3, 4};
XS(XS_PvView_marpa) {
    dXSARGS;
    SV *sv = newSV(0);
    PERL_UNUSED_VAR(items);
    SvUPGRADE(sv, SVt_PV);
    SvPOK_only(sv);
    SvPV_set(sv, text);
    SvLEN_set(sv, 0);
    SvCUR_set(sv, strlen(text));
    SvREADONLY_on(sv);
    ST(0) = sv_2mortal(sv);
    XSRETURN(1);
}
XS(XS_PvView_swish) {
    dXSARGS;
    SV *wrapper = sv_newmortal();
    SV *out = newSVpvs("[");
    PERL_UNUSED_VAR(items);
    sv_upgrade(wrapper, SVt_PV);
    SvREADONLY_on(wrapper);
    SvLEN_set(wrapper, 0);
    SvPVX(wrapper) = text + 4;
    SvCUR_set(wrapper, 5);
    SvPOK_on(wrapper);
    sv_catsv(out, wrapper);
    sv_catpvs(out, "]");
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
XS(XS_PvView_writable) {
    dXSARGS;
    SV *sv = newSV(0);
    PERL_UNUSED_VAR(items);
    SvUPGRADE(sv, SVt_PV);
    SvPV_set(sv, text);
    SvLEN_set(sv, 0);
    SvCUR_set(sv, strlen(text));
    SvPOK_only(sv);
    ST(0) = sv_2mortal(sv);
    XSRETURN(1);
}
XS(XS_PvView_structure) {
    dXSARGS;
    SV *sv = newSV(0);
    PERL_UNUSED_VAR(items);
    sv_upgrade(sv, SVt_PV);
    SvPVX(sv) = (char *)structure;
    ST(0) = sv_2mortal(newSViv(((int *)SvPVX(sv))[2] == 3 && SvPVX(sv) == (char *)structure));
    SvPVX(sv) = NULL;
    SvREFCNT_dec(sv);
    XSRETURN(1);
}
XS_EXTERNAL(boot_PvView) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("PvView::marpa", XS_PvView_marpa);
    newXS_deffile("PvView::swish", XS_PvView_swish);
    newXS_deffile("PvView::writable", XS_PvView_writable);
    newXS_deffile("PvView::structure", XS_PvView_structure);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	for _, tc := range []struct {
		name, code string
		want       *regexp.Regexp
	}{
		{"made read-only after its length", `my $v = PvView::marpa(); join "|", $v, length($v), uc($v)`,
			regexp.MustCompile(`\Athe bytes of a library\|22\|THE BYTES OF A LIBRARY\z`)},
		{"read-only first, handed to the interpreter", `PvView::swish()`, regexp.MustCompile(`\A\[bytes\]\z`)},
		{"a view perl could write through", `my $r = eval { my $v = PvView::writable(); 1 } ? "lived" : $@; $r`,
			regexp.MustCompile(`\Athe string of a scalar cannot be in native memory on go-perl: .* made it a string that perl can write to; .* only when it is read-only at `)},
		{"a structure behind a scalar", `PvView::structure()`, regexp.MustCompile(`\A1\z`)},
	} {
		if got := requireModule(t, "PvView", so, tc.code); !tc.want.MatchString(got) {
			t.Errorf("%s: got %q, want %v", tc.name, got, tc.want)
		}
	}
}
