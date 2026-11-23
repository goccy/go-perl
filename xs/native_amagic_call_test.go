//go:build darwin || linux

package xs_test

import "testing"

// amagic_call of native code gives the overloaded operator's answer, as
// perl's: cmp and <=> of two objects either way round and of an object and
// a plain string, gt by way of cmp, and none for operands that overload
// nothing. Its context is PL_op's, as perl takes it: undef in an XSUB
// called in void context, and a value under an op of native code's own
// made PL_op with OPf_WANT_SCALAR, which Algorithm::Heapify::XS does to
// order its heaps by an overloaded cmp; GIMME_V reads that op too, and an
// op without want bits gives the context of the enclosing sub. What is
// expected is what perl 5.36 on Linux gives with the same module.
func TestNativeAmagicCall(t *testing.T) {
	so := buildModuleSource(t, "Amagic", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* call(left, right, which): the overloaded cmp, <=>, gt or > of the two, as
 * Algorithm::Heapify::XS asks amagic_call for it; "none" when neither
 * overloads it */
XS(XS_Amagic_call) {
    dXSARGS;
    SV *left = ST(0), *right = ST(1), *res;
    const char *which = SvPV_nolen(ST(2));
    int method = strEQ(which, "cmp") ? scmp_amg : strEQ(which, "<=>") ? ncmp_amg
               : strEQ(which, "gt") ? sgt_amg : gt_amg;
    PERL_UNUSED_VAR(items);
    res = amagic_call(left, right, method, 0);
    ST(0) = res ? sv_2mortal(newSVsv(res)) : sv_2mortal(newSVpvs("none"));
    XSRETURN(1);
}
/* keep(left, right): amagic_call's cmp of the two kept in $main::kept, so
 * that it can be called in void context too */
XS(XS_Amagic_keep) {
    dXSARGS;
    SV *res = amagic_call(ST(0), ST(1), scmp_amg, 0);
    PERL_UNUSED_VAR(items);
    sv_setsv(get_sv("main::kept", GV_ADD), res ? res : sv_2mortal(newSVpvs("none")));
    XSRETURN_EMPTY;
}
/* forced(left, right, want): with a copy of PL_op whose want bits are want
 * PL_op, as Algorithm::Heapify::XS forces scalar context: amagic_call's cmp
 * of the two and GIMME_V, kept in $main::kept as "cmp/gimme" */
XS(XS_Amagic_forced) {
    dXSARGS;
    OP fakeop;
    SV *res;
    U8 gimme;
    const U8 want = (U8)SvUV(ST(2));
    PERL_UNUSED_VAR(items);
    ENTER;
    SAVEOP();
    Copy(PL_op, &fakeop, 1, OP);
    fakeop.op_flags = want;
    PL_op = &fakeop;
    res = amagic_call(ST(0), ST(1), scmp_amg, 0);
    gimme = GIMME_V;
    sv_setpvf(get_sv("main::kept", GV_ADD), "%s/%d",
              res ? (SvOK(res) ? SvPV_nolen(res) : "undef") : "none", (int)gimme);
    LEAVE;
    XSRETURN_EMPTY;
}
/* method(obj, name): the sub an overload method name ("(gt") of the class
 * of obj names, as package::name, as Algorithm::Heapify::XS looks for the
 * methods a class really overloads ("none" when there is none) */
XS(XS_Amagic_method) {
    dXSARGS;
    STRLEN len;
    const char *name = SvPV(ST(1), len);
    HV *stash = SvSTASH(SvRV(ST(0)));
    GV *gv = gv_fetchmeth_pvn(stash, name, len, -1, 0);
    CV *mcv = gv ? GvCV(gv) : NULL;
    GV *cvgv = mcv ? CvGV(mcv) : NULL;
    PERL_UNUSED_VAR(items);
    if (!cvgv) {
        ST(0) = sv_2mortal(newSVpvs("none"));
    } else {
        const HEK *gvhek = GvNAME_HEK(cvgv);
        const HEK *stashek = HvNAME_HEK(GvSTASH(cvgv));
        ST(0) = sv_2mortal(newSVpvf("%.*s::%.*s", stashek ? (int)HEK_LEN(stashek) : 1,
                                    stashek ? HEK_KEY(stashek) : "?", gvhek ? (int)HEK_LEN(gvhek) : 1,
                                    gvhek ? HEK_KEY(gvhek) : "?"));
    }
    XSRETURN(1);
}
XS_EXTERNAL(boot_Amagic) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("Amagic::call", XS_Amagic_call, __FILE__);
    newXS("Amagic::method", XS_Amagic_method, __FILE__);
    newXS("Amagic::keep", XS_Amagic_keep, __FILE__);
    newXS("Amagic::forced", XS_Amagic_forced, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `package Str;
use overload 'cmp' => sub { my ($a, $b, $swap) = @_; my $r = $$a cmp (ref $b ? $$b : $b); $swap ? -$r : $r },
             '""' => sub { ${$_[0]} };
package Num;
use overload '<=>' => sub { my ($a, $b, $swap) = @_; my $r = $$a <=> (ref $b ? $$b : $b); $swap ? -$r : $r };
package Big;
use overload 'cmp' => sub { my ($l, $r, $swap) = @_; ($l, $r) = ($r, $l) if $swap; $$l cmp $$r },
             '""' => sub { ${$_[0]} }, fallback => 1;
package main;
my ($a, $b) = (bless(\(my $x = "pear"), 'Str'), bless(\(my $y = "apple"), 'Str'));
my ($n, $m) = (bless(\(my $p = 3), 'Num'), bless(\(my $q = 10), 'Num'));
my ($u, $v) = (bless(\(my $s = "18446744073709551615"), 'Big'), bless(\(my $t = "18446744073709551611"), 'Big'));
our $kept;
Amagic::keep($u, $v);
my $void = defined $kept ? $kept : "undef";
my $scalar = do { my $x = Amagic::keep($v, $u); defined $kept ? $kept : "undef" };
Amagic::forced($u, $v, 2);
my $forced = $kept;
Amagic::forced($v, $u, 1);
my $forced_void = $kept;
sub inlist { Amagic::forced($u, $v, 0); $kept }
my ($forced_none) = inlist();
join ",", $void, $scalar, $forced, $forced_void, $forced_none, map { defined $_ ? $_ : "undef" }
    Amagic::call($a, $b, "cmp"), Amagic::call($b, $a, "cmp"), Amagic::call($a, "zebra", "cmp"),
    Amagic::call($n, $m, "<=>"), Amagic::call($m, $n, "<=>"), Amagic::call($a, $b, "gt"),
    Amagic::call(1, 2, "cmp"),
    Amagic::call($u, $v, "cmp"), Amagic::call($v, $u, "cmp"), Amagic::call($u, $v, "gt"),
    Amagic::method($u, "(cmp"), Amagic::method($u, "(gt"), Amagic::method($a, "(gt")`
	if got, want := requireModule(t, "Amagic", so, program), "undef,-1,1/2,undef/1,1/3,1,-1,-1,-1,1,1,none,1,-1,1,Big::__ANON__,none,none"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
