//go:build darwin || linux

package xs_test

import "testing"

// An infinity and a NaN in the floating point conversions of the SDK's
// formatting (newSVpvf, croak, warn) are perl's Inf, -Inf and NaN, padded
// to the width as perl pads them (with the zeros of the 0 flag left of the
// sign), not the C library's inf and nan; other numbers are formatted as
// before. What is expected is what perl 5.36 on Linux gives with the same
// module.
func TestNativeFormattingOfInfinityAndNaN(t *testing.T) {
	so := buildModuleSource(t, "FmtInf", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* fmt(inf, nan, x): an infinity, a NaN and a plain number through the
 * floating point conversions of newSVpvf, with flags and widths */
XS(XS_FmtInf_fmt) {
    dXSARGS;
    NV i = SvNV(ST(0)), n = SvNV(ST(1)), x = SvNV(ST(2));
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("[%5.2f][%-6g][%+g][%05g][% g][%e][%G][%+G][%5g][%-5g][%.1f]"
                                "[%*g][%" NVgf "][%" NVff "][%.3" NVef "][%08.3f]",
                                i, i, i, -i, i, n, i, n, n, -n, -i, 7, -i, n, x, x, x));
    XSRETURN(1);
}
XS_EXTERNAL(boot_FmtInf) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("FmtInf::fmt", XS_FmtInf_fmt, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $i = 9**9**9;
FmtInf::fmt($i, -sin($i), -2.5)`
	if got, want := requireModule(t, "FmtInf", so, program), "[  Inf][Inf   ][+Inf][0-Inf][+Inf][NaN][Inf][NaN][  NaN][NaN  ][-Inf][   -Inf][NaN][-2.500000][-2.500e+00][-002.500]"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
