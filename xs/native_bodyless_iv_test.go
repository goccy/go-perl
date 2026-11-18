//go:build darwin || linux

package xs_test

import "testing"

// Sereal::Decoder unrolls sv_setiv for a new scalar: it makes the scalar a
// bodyless IV itself, SvANY(sv) = (XPVIV *)((char *)&sv->sv_u.svu_iv -
// STRUCT_OFFSET(XPVIV, xiv_iv)), sets the type and the IOK flags in
// SvFLAGS and stores the integer with SvIV_set; the scalar is then an
// integer as sv_setiv makes it. And modules take the address of
// PL_sig_pending to watch it across a wait (Data::HashMap::Shared,
// Data::ReqRep::Shared): it is the interpreter's own variable. Before, XPVIV
// and sv_u did not compile, SvANY could not be stored into, and
// &PL_sig_pending was no lvalue.
const bodylessIVSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* iv(n): a new scalar made an IV as Sereal::Decoder makes one */
XS(XS_MyBi_iv) {
    dXSARGS;
    IV iv = SvIV(ST(0));
    SV *into = newSV(0);
    PERL_UNUSED_VAR(items);
    if (SvTYPE(into) == SVt_NULL) {
        SvANY(into) = (XPVIV*)((char*)&(into->sv_u.svu_iv) - STRUCT_OFFSET(XPVIV, xiv_iv));
        SvFLAGS(into) |= (SVt_IV | SVf_IOK | SVp_IOK);
        SvIV_set(into, iv);
    } else {
        sv_setiv(into, iv);
    }
    ST(0) = sv_2mortal(into);
    XSRETURN(1);
}
/* sig_pending(): PL_sig_pending through the address a module keeps */
XS(XS_MyBi_sig_pending) {
    dXSARGS;
    volatile int *watched = (volatile int *)&PL_sig_pending;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(*watched == PL_sig_pending ? 1 : 0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyBi) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyBi::iv", XS_MyBi_iv);
    newXS_deffile("MyBi::sig_pending", XS_MyBi_sig_pending);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeMakesABodylessIV(t *testing.T) {
	so := buildModuleSource(t, "MyBi", bodylessIVSource)
	got := requireModule(t, "MyBi", so, `use B ();
my @r;
for my $n (42, -7, 0) {
    my $ref = \MyBi::iv($n);
    my $o = B::svref_2object($ref);
    push @r, sprintf "%d %s %s", $$ref, ref($o), ($o->FLAGS & B::SVf_IOK()) ? "IOK" : "-";
}
my $w = MyBi::iv(5);
$w .= "x";
push @r, $w;
push @r, "watch " . MyBi::sig_pending();
join "; ", @r`)
	// perl: the scalar is an IV (B::IV) with IOK, and takes a string as one
	if want := "42 B::IV IOK; -7 B::IV IOK; 0 B::IV IOK; 5x; watch 1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
