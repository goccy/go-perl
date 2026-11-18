//go:build darwin || linux

package xs_test

import "testing"

// The mro data of a stash is perl's struct mro_meta, as next::XS uses it:
// the algorithm in use (mro_which: its name and length), the c3 algorithm
// by name and its resolve function (the linearized @ISA), a cache native
// code makes in mro_nextmethod and finds there in a later call, and the
// generation that changes when a method is defined. Before, HvMROMETA gave
// two generation counters in one scratch struct.
const mroMetaSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* which(class): the name of the class's algorithm */
XS(XS_MyMro_which) {
    dXSARGS;
    HV *const stash = gv_stashsv(ST(0), 0);
    const struct mro_alg *const a = HvMROMETA(stash)->mro_which;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvn(a->name, a->length));
    XSRETURN(1);
}
/* c3(class): the class's c3 linearization, by the algorithm's resolve */
XS(XS_MyMro_c3) {
    dXSARGS;
    HV *const stash = gv_stashsv(ST(0), 0);
    const struct mro_alg *const algo = Perl_mro_get_from_name(aTHX_ sv_2mortal(newSVpvs("c3")));
    AV *const lin = algo->resolve(aTHX_ stash, 0);
    SV *const out = sv_2mortal(newSVpvs(""));
    SSize_t i;
    PERL_UNUSED_VAR(items);
    for (i = 0; i <= av_len(lin); i++)
        sv_catpvf(out, "%s%s", i ? "," : "", SvPV_nolen(*av_fetch(lin, i, 0)));
    ST(0) = out;
    XSRETURN(1);
}
/* cache(class, key, value): the value under the key in the class's
 * next::method cache, made if it has none; -> what was there before */
XS(XS_MyMro_cache) {
    dXSARGS;
    HV *const stash = gv_stashsv(ST(0), 0);
    struct mro_meta *const meta = HvMROMETA(stash);
    HV *nm = meta->mro_nextmethod;
    SV **was;
    SV *before;
    PERL_UNUSED_VAR(items);
    if (!nm) nm = meta->mro_nextmethod = newHV();
    was = hv_fetch(nm, SvPV_nolen(ST(1)), (I32)SvCUR(ST(1)), 0);
    before = was ? newSVsv(*was) : newSVpvs("none");
    hv_store(nm, SvPV_nolen(ST(1)), (I32)SvCUR(ST(1)), newSVsv(ST(2)), 0);
    ST(0) = sv_2mortal(before);
    XSRETURN(1);
}
/* gen(class): the generation of the class's own methods */
XS(XS_MyMro_gen) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVuv(HvMROMETA(gv_stashsv(ST(0), 0))->pkg_gen));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyMro) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyMro::which", XS_MyMro_which);
    newXS_deffile("MyMro::c3", XS_MyMro_c3);
    newXS_deffile("MyMro::cache", XS_MyMro_cache);
    newXS_deffile("MyMro::gen", XS_MyMro_gen);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeReadsTheMroMeta(t *testing.T) {
	so := buildModuleSource(t, "MyMro", mroMetaSource)
	got := requireModule(t, "MyMro", so, `{ package A; sub hi {} } { package B; our @ISA = ("A") } { package C; our @ISA = ("A") }
{ package D; use mro "c3"; our @ISA = ("B", "C") }
my @r = (MyMro::which("B"), MyMro::which("D"), MyMro::c3("D"));
push @r, MyMro::cache("D", "k", "v1"), MyMro::cache("D", "k", "v2");
my $g = MyMro::gen("D");
eval "sub D::added {}";
push @r, MyMro::gen("D") != $g ? "changed" : "same";
join "|", @r`)
	if want := "dfs|c3|D,B,C,A|none|v1|changed"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
