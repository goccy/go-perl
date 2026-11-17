//go:build darwin || linux

package xs_test

import "testing"

// An encoding made by enc2xs (Encode::EUCJPASCII, Encode::HanExtra) is an
// Encode::XS object holding the address of a table of the module's own,
// in native memory, which Encode in the interpreter cannot read: using it
// dies saying so. Before, Encode read the address as one of the
// interpreter's memory and the interpreter trapped.
func TestNativeEncodeTableDies(t *testing.T) {
	so := buildModuleSource(t, "NativeTable", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_NativeTable_encoding) {
    dXSARGS;
    static char table_of_its_own[64];
    SV *iv = newSViv((IV)(uintptr_t)table_of_its_own);
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(sv_bless(newRV_noinc(iv), gv_stashpv("Encode::XS", GV_ADD)));
    XSRETURN(1);
}
XS_EXTERNAL(boot_NativeTable) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("NativeTable::encoding", XS_NativeTable_encoding);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	for _, tc := range []struct{ name, code string }{
		{"decode", `$e->decode("abc")`},
		{"encode", `$e->encode("abc")`},
		{"name", `$e->name`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := requireModule(t, "NativeTable", so, `require Encode;
my $e = NativeTable::encoding();
my $r = eval { `+tc.code+`; 1 } ? "lived" : $@;
$r =~ /\Ago-perl gives native code no access to this interpreter internal: an Encode::XS table of native code \(0x[0-9a-f]+\) at / ? "died" : $r`)
			if got != "died" {
				t.Errorf("got %q, want died", got)
			}
		})
	}
}
