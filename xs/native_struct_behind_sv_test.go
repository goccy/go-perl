//go:build darwin || linux

package xs_test

import "testing"

// A structure of native code behind an SV that is no string, with the size
// of its buffer: Hash::SharedMem's handle (newSV_type(SVt_PVMG), Newxz,
// SvPV_set, SvLEN_set, no flag of a string). The module goes on writing
// through its own pointer after SvLEN_set, and reads the structure back with
// SvPVX in later calls, and frees it in DESTROY. Before, SvLEN_set handed the
// block to the interpreter at once: SvPVX gave a copy taken at that moment
// (mode 0, no name), and the module's own block was freed when the call
// returned. What is expected is what perl 5.36 on Linux gives with the same
// module.
func TestNativeStructureWithItsSizeBehindAnSV(t *testing.T) {
	so := buildModuleSource(t, "KeepStruct", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* Hash::SharedMem's handle: a structure behind an SV that is no string,
 * with the size of its buffer (SvPV_set, SvLEN_set). The module goes on
 * writing through its own pointer after SvLEN_set, and reads the structure
 * back with SvPVX in later calls. */
struct handle {
    IV mode;
    IV writes;
    SV *name;
};
static struct handle *handle_of(pTHX_ SV *ref) {
    if (!SvROK(ref) || !sv_derived_from(ref, "KeepStruct")) croak("not a handle");
    return (struct handle *)SvPVX(SvRV(ref));
}
XS(XS_KeepStruct_new) {
    dXSARGS;
    SV *sv;
    struct handle *h;
    PERL_UNUSED_VAR(items);
    sv = newSV_type(SVt_PVMG);
    Newxz(h, 1, struct handle);
    SvPV_set(sv, (char *)h);
    SvLEN_set(sv, sizeof(struct handle));
    h->mode = SvIV(ST(0));
    h->name = newSVsv(ST(1));
    ST(0) = sv_2mortal(sv_bless(newRV_noinc(sv), gv_stashpv("KeepStruct", GV_ADD)));
    XSRETURN(1);
}
XS(XS_KeepStruct_write) {
    dXSARGS;
    struct handle *h;
    PERL_UNUSED_VAR(items);
    h = handle_of(aTHX_ ST(0));
    if (!h->mode) croak("can't write %s: opened in unwritable mode", SvPV_nolen(h->name));
    h->writes++;
    ST(0) = sv_2mortal(newSViv(h->writes));
    XSRETURN(1);
}
XS(XS_KeepStruct_info) {
    dXSARGS;
    struct handle *h;
    SV *sv;
    PERL_UNUSED_VAR(items);
    h = handle_of(aTHX_ ST(0));
    sv = SvRV(ST(0));
    ST(0) = sv_2mortal(newSVpvf("%d:%d:%s:%s:%d", (int)h->mode, (int)h->writes,
                                SvPV_nolen(h->name), SvPOK(sv) ? "string" : "no string",
                                (int)SvLEN(sv)));
    XSRETURN(1);
}
XS(XS_KeepStruct_DESTROY) {
    dXSARGS;
    SV *sv;
    struct handle *h;
    PERL_UNUSED_VAR(items);
    sv = SvRV(ST(0));
    h = (struct handle *)SvPVX(sv);
    if (h) {
        SvREFCNT_dec(h->name);
        Safefree(h);
        SvPV_set(sv, NULL);
        SvLEN_set(sv, 0);
    }
    XSRETURN(0);
}
XS_EXTERNAL(boot_KeepStruct) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("KeepStruct::new", XS_KeepStruct_new, __FILE__);
    newXS("KeepStruct::write", XS_KeepStruct_write, __FILE__);
    newXS("KeepStruct::info", XS_KeepStruct_info, __FILE__);
    newXS("KeepStruct::DESTROY", XS_KeepStruct_DESTROY, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my @r;
my $h = KeepStruct::new(1, "first");
KeepStruct::write($h) for 1 .. 3;
push @r, KeepStruct::info($h);
my $ro = KeepStruct::new(0, "ro");
push @r, eval { KeepStruct::write($ro); 1 } ? "written" : ($@ =~ /unwritable mode/ ? "refused" : $@);
my $bad = 0;
for my $i (1 .. 300) {
    my $t = KeepStruct::new(2, "h$i");
    KeepStruct::write($t) for 1 .. $i % 5;
    $bad++ if KeepStruct::info($t) ne "2:" . ($i % 5) . ":h$i:no string:24";
}
push @r, "bad=$bad";
undef $h;
push @r, KeepStruct::info($ro);
join ",", @r`
	const want = "1:3:first:no string:24,refused,bad=0,0:0:ro:no string:24"
	if got := requireModule(t, "KeepStruct", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
