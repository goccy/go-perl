//go:build darwin || linux

package xs_test

import "testing"

// A module adds a name to the pad of the code being compiled the way
// Lexical::Var does: a new pad name whose fields it sets straight away,
// marked "our" with the names pad.h kept from before pad names were
// structures of their own (SvPAD_OUR_on, SvOURSTASH_set), stored with
// padnamelist_store. Before, the SDK had none of those names: the module's
// fallbacks did SvFLAGS and av_store on the SDK's pad name, which is no
// scalar, and the operations trapped (Lexical-Var, Lexical-Import).
func TestNativeStoresAPadName(t *testing.T) {
	so := buildModuleSource(t, "PadName", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_PadName_add) {
    dXSARGS;
    PADLIST *padlist = CvPADLIST(PL_compcv);
    PADNAMELIST *names = PadlistNAMES(padlist);
    PAD *pad = PadlistARRAY(padlist)[1];
    PADNAME *name = newPADNAMEpvn("$padname_test", 13);
    PADOFFSET off;
    PADNAME *back;
    PERL_UNUSED_VAR(items);
    /* never in scope: nothing compiled after this finds it */
    name->xpadn_low = 0;
    name->xpadn_high = 0;
    SvPAD_OUR_on(name);
    SvOURSTASH_set(name, (HV *)SvREFCNT_inc((SV *)gv_stashpv("main", 0)));
    off = AvFILLp(pad) + 1;
    padnamelist_store(names, off, name);
    av_store(pad, off, newSV(0));
    back = padnamelist_fetch(names, off);
    ST(0) = sv_2mortal(newSVpvf("%s our:%d flags:%d low:%d max:%d",
                                PadnamePV(back), (int)PadnameIsOUR(back),
                                (int)(PadnameFLAGS(back) & PADNAMEf_OUR),
                                (int)COP_SEQ_RANGE_LOW(back),
                                (int)(PadnamelistMAX(names) == (SSize_t)off)));
    XSRETURN(1);
}
XS_EXTERNAL(boot_PadName) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("PadName::add", XS_PadName_add);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	got := requireModule(t, "PadName", so, `my $r;
eval q{ sub compiled { my $x = 1; BEGIN { $r = PadName::add() } $x + 1 } 1 } or die $@;
"$r; " . compiled()`)
	if want := "$padname_test our:1 flags:16 low:0 max:1; 2"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
