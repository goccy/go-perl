//go:build darwin || linux

package xs_test

import "testing"

// utf8_hop goes back for a negative count, past the continuation bytes of
// each character, as perl's does. pTk's Tcl_UtfPrev is utf8_hop(s, -1),
// which the Text widget steps back through its lines with: before, a
// negative count did not move, so delete('1.0', 'end') of a Text widget
// ended at a byte past its last line and walked off the B-tree (SIGSEGV in
// TkBTreeGetTags: Tk's t/errordialog.t, Tk-ObjScanner, Tk-ROSyntaxText).
func TestNativeUTF8HopBack(t *testing.T) {
	so := buildModuleSource(t, "Hop", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_Hop_offsets) {
    dXSARGS;
    STRLEN len;
    const U8 *s = (const U8 *)SvPV(ST(0), len);
    const U8 *end = s + len;
    const U8 *p = end;
    SV *out = newSVpvs("");
    int steps = 0;
    PERL_UNUSED_VAR(items);
    while (p > s && steps++ < 8) {
        const U8 *q = utf8_hop(p, -1);
        sv_catpvf(out, "%d,", (int)(q - s));
        if (q == p) break; /* did not move */
        p = q;
    }
    sv_catpvf(out, "fwd:%d", (int)(utf8_hop(s, 2) - s));
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
XS_EXTERNAL(boot_Hop) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("Hop::offsets", XS_Hop_offsets);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	// a, e-acute (2 bytes), euro sign (3 bytes), b
	got := requireModule(t, "Hop", so, `my $s = "a\x{e9}\x{20ac}b"; utf8::encode($s); Hop::offsets($s)`)
	if want := "6,3,1,0,fwd:3"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
