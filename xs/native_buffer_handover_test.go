//go:build darwin || linux

package xs_test

import "testing"

// handOverModule hands blocks of its own to SVs as their buffers, the
// sv_usepvn idiom: CBOR::Free's encoder (SvPV_set, SvPOK_on, SvCUR_set,
// SvLEN_set), JSON::XS::ByteString's strings (SvPOK_on first), a block
// Renew grew, a savepv copy, and the buffer slot as an lvalue.
const handOverModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static SV *handed(SV *src, int how) {
    STRLEN len;
    const char *text = SvPV(src, len);
    SV *sv = newSV(0);
    char *buf;
    if (how == 2) {
        Newx(buf, 2, char);
        Renew(buf, len + 1, char);
    } else if (how == 3) {
        buf = savepvn(text, len);
    } else {
        Newx(buf, len + 8, char);
    }
    Copy(text, buf, len, char);
    buf[len] = '\0';
    if (how == 1) { /* JSON::XS::ByteString */
        sv_upgrade(sv, SVt_PV);
        SvPOK_on(sv);
        SvPV_set(sv, buf);
        SvCUR_set(sv, len);
        SvLEN_set(sv, len + 8);
    } else if (how == 4) { /* the slot as an lvalue */
        SvUPGRADE(sv, SVt_PV);
        SvPVX(sv) = buf;
        SvLEN(sv) = len + 8;
        SvCUR_set(sv, len);
        SvPOK_only(sv);
    } else { /* CBOR::Free */
        SvUPGRADE(sv, SVt_PV);
        SvPV_set(sv, buf);
        SvPOK_on(sv);
        SvCUR_set(sv, len);
        SvLEN_set(sv, how == 0 ? len + 8 : len + 1);
    }
    return sv;
}
XS(XS_HandOver_handed) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "text, how");
    ST(0) = sv_2mortal(handed(ST(0), (int)SvIV(ST(1))));
    XSRETURN(1);
}
XS(XS_HandOver_many) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        AV *av = newAV();
        IV i, n = SvIV(ST(0));
        for (i = 0; i < n; i++) {
            SV *num = sv_2mortal(newSViv(i));
            av_push(av, handed(num, (int)(i % 5)));
        }
        ST(0) = sv_2mortal(newRV_noinc((SV *)av));
        XSRETURN(1);
    }
}
XS_EXTERNAL(boot_HandOver) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("HandOver::handed", XS_HandOver_handed);
    newXS_deffile("HandOver::many", XS_HandOver_many);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// A block the module's own allocators gave (Newx, Renew, savepv) that it
// hands to an SV as its buffer is the SV's from then on, as on perl: the
// string reads as the bytes the module wrote, and grows and is freed as
// any of the interpreter's. Before, the string read as zeros or blanks of
// the right length (CBOR-Free, JSON-XS-ByteString), and since the refusal
// of a string in native memory, the call died.
func TestNativeBufferHandedToAnSV(t *testing.T) {
	so := buildModuleSource(t, "HandOver", handOverModule)
	for _, tc := range []struct{ name, code, want string }{
		{"the CBOR::Free order", `my $s = HandOver::handed("encoded", 0); $s .= " and more"; $s`, "encoded and more"},
		{"the JSON::XS::ByteString order", `HandOver::handed("Cindy \x{597d}", 1) eq "Cindy \xe5\xa5\xbd" ? "same bytes" : "other"`, "same bytes"},
		{"a block Renew grew", `HandOver::handed("renewed", 2)`, "renewed"},
		{"a savepvn copy", `HandOver::handed("saved", 3)`, "saved"},
		{"the buffer slot as an lvalue", `HandOver::handed("lvalue", 4)`, "lvalue"},
		{"many, read back", `my $a = HandOver::many(2000); my $bad = grep { $a->[$_] ne "$_" } 0 .. 1999; "$bad bad of " . scalar(@$a)`, "0 bad of 2000"},
	} {
		if got := requireModule(t, "HandOver", so, tc.code); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
