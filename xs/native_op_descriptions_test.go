//go:build darwin || linux

package xs_test

import "testing"

// Native code walks the links of ops in its own memory, and finds every op
// it reaches with its links: from the op a pp function of its own runs
// (List::Keywords' first returns PL_op->op_next->op_next), from an op a
// number names (B::Generate's find_cv_by_root walks op_next to the end of
// the sub), and from an op it kept from an earlier native call. An op
// reached in two ways, in two native calls, has the same bytes
// (Devel::Cover keys a condition by a copy of the op). What perl writes
// into ops at run time is seen: a match op's regexp, which pp_regcomp
// replaces when the pattern changes, and the reference count in a sub's
// root that each closure made of the sub counts. Ops of subs that were freed are not taken for the ops of
// later subs at their addresses. Before, the ops after an op were at the
// edge of what was described and had no links: the program ended where the
// pp function returned NULL, and the walks stopped at the second op. The
// expected string is what perl prints with the same module code.
const opDescriptionsSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <string.h>

/* the pp function a call of MyDesc::peek() is compiled to: the name of the
 * op after the next one, kept in $MyDesc::seen */
static OP *MyDesc_pp_peek(pTHX) {
    OP *const after = PL_op->op_next ? PL_op->op_next->op_next : NULL;
    sv_setpv(get_sv("MyDesc::seen", GV_ADD), after ? OP_NAME(after) : "none");
    return NORMAL;
}
static OP *MyDesc_ck(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    OP *o;
    PERL_UNUSED_ARG(namegv);
    PERL_UNUSED_ARG(ckobj);
    op_free(entersubop);
    o = newOP(OP_CUSTOM, 0);
    o->op_ppaddr = MyDesc_pp_peek;
    return o;
}
XS(XS_MyDesc_peek) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    croak("MyDesc::peek was not compiled to its op");
}
XS(XS_MyDesc_install) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    cv_set_call_checker(get_cv("MyDesc::peek", 0), MyDesc_ck, &PL_sv_undef);
    XSRETURN_EMPTY;
}

/* the names of the ops from o to the end of its chain */
static SV *MyDesc_names(pTHX_ const OP *o) {
    SV *const out = newSVpvs("");
    for (; o; o = o->op_next) {
        if (SvCUR(out)) sv_catpvs(out, ",");
        sv_catpv(out, OP_NAME(o));
    }
    return out;
}

/* the last op of the chain from the op a number names (a B::OP object's),
 * and whether it is the sub's root */
XS(XS_MyDesc_chain_end) {
    dXSARGS;
    OP *o;
    int n = 0;
    if (items != 2) croak_xs_usage(cv, "code, opnumber");
    o = INT2PTR(OP *, SvIV(ST(1)));
    while (o->op_next) {
        o = o->op_next;
        n++;
    }
    ST(0) = sv_2mortal(newSVpvf("%s %s %s", OP_NAME(o),
                                o == CvROOT((CV *)SvRV(ST(0))) ? "root" : "other",
                                n > 1 ? "walked" : "short"));
    XSRETURN(1);
}

/* the op after the call of keep(), kept for a later call */
static const OP *MyDesc_kept;
XS(XS_MyDesc_keep) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    MyDesc_kept = PL_op->op_next;
    XSRETURN_EMPTY;
}
XS(XS_MyDesc_walk_kept) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(MyDesc_names(aTHX_ MyDesc_kept));
    XSRETURN(1);
}

/* the bytes of the op after the call of keep_bytes(code), and how far it
 * is from the start of the sub; compared later with the op that far from
 * the start */
static OP MyDesc_bytes;
static int MyDesc_steps;
XS(XS_MyDesc_keep_bytes) {
    dXSARGS;
    const OP *o, *const after = PL_op->op_next;
    int n = 0;
    if (items != 1) croak_xs_usage(cv, "code");
    for (o = CvSTART((CV *)SvRV(ST(0))); o && o != after; o = o->op_next) n++;
    MyDesc_steps = o ? n : -1;
    memcpy(&MyDesc_bytes, after, sizeof(OP));
    XSRETURN_EMPTY;
}
XS(XS_MyDesc_bytes_same) {
    dXSARGS;
    const OP *o;
    int n;
    if (items != 1) croak_xs_usage(cv, "code");
    o = CvSTART((CV *)SvRV(ST(0)));
    for (n = 0; o && n < MyDesc_steps; n++) o = o->op_next;
    ST(0) = sv_2mortal(newSVpv(MyDesc_steps < 0 ? "lost" :
                               !o ? "short" :
                               memcmp(&MyDesc_bytes, o, sizeof(OP)) ? "differ" : "same", 0));
    XSRETURN(1);
}

static OP *MyDesc_find(pTHX_ OP *o, int type) {
    OP *k;
    if (!o) return NULL;
    if (o->op_type == type) return o;
    if (o->op_flags & OPf_KIDS)
        for (k = cUNOPo->op_first; k; k = OpSIBLING(k)) {
            OP *r = MyDesc_find(aTHX_ k, type);
            if (r) return r;
        }
    return NULL;
}
/* the pattern of the match op of a sub */
XS(XS_MyDesc_pattern) {
    dXSARGS;
    OP *o;
    REGEXP *rx;
    if (items != 1) croak_xs_usage(cv, "code");
    o = MyDesc_find(aTHX_ CvROOT((CV *)SvRV(ST(0))), OP_MATCH);
    rx = o ? PM_GETRE(cPMOPx(o)) : NULL;
    ST(0) = sv_2mortal(rx ? newSVpvn(RX_PRECOMP(rx), RX_PRELEN(rx)) : newSVpvs("none"));
    XSRETURN(1);
}
/* the reference count of a sub's root, which closures share */
XS(XS_MyDesc_refcnt) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "code");
    ST(0) = sv_2mortal(newSVuv((UV)CvROOT((CV *)SvRV(ST(0)))->op_targ));
    XSRETURN(1);
}
/* the names of the ops after the call of names_here() */
XS(XS_MyDesc_names_here) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    sv_setsv(get_sv("MyDesc::names", GV_ADD), sv_2mortal(MyDesc_names(aTHX_ PL_op->op_next)));
    XSRETURN_EMPTY;
}

XS_EXTERNAL(boot_MyDesc) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyDesc::peek", XS_MyDesc_peek);
    newXS_deffile("MyDesc::install", XS_MyDesc_install);
    newXS_deffile("MyDesc::chain_end", XS_MyDesc_chain_end);
    newXS_deffile("MyDesc::keep", XS_MyDesc_keep);
    newXS_deffile("MyDesc::walk_kept", XS_MyDesc_walk_kept);
    newXS_deffile("MyDesc::keep_bytes", XS_MyDesc_keep_bytes);
    newXS_deffile("MyDesc::bytes_same", XS_MyDesc_bytes_same);
    newXS_deffile("MyDesc::pattern", XS_MyDesc_pattern);
    newXS_deffile("MyDesc::refcnt", XS_MyDesc_refcnt);
    newXS_deffile("MyDesc::names_here", XS_MyDesc_names_here);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeWalksOfOpsAcrossCalls(t *testing.T) {
	so := buildModuleSource(t, "MyDesc", opDescriptionsSource)
	got := requireModule(t, "MyDesc", so, `MyDesc::install();
my $f = eval q{ sub { MyDesc::peek(); my $z = 5; $MyDesc::seen } } or die $@;
use B;
my $h = sub { my ($p, $q) = @_; $p ? $q + 1 : $q * 2 };
my $k = sub { MyDesc::keep(); return 7 };
$k->();
my $g = sub { MyDesc::keep_bytes($_[0]); return 3 };
$g->($g);
join "|", $f->(), MyDesc::chain_end($h, ${B::svref_2object($h)->START}),
    MyDesc::walk_kept(), MyDesc::bytes_same($g)`)
	want := "const|leavesub root walked|nextstate,const,leavesub|same"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestNativeSeesWhatPerlWritesIntoOps(t *testing.T) {
	so := buildModuleSource(t, "MyDesc", opDescriptionsSource)
	got := requireModule(t, "MyDesc", so, `my @r;
my $re = "ab";
my $m = sub { "xxabcd" =~ m/$re/ ? "y" : "n" };
push @r, $m->(), MyDesc::pattern($m);
$re = "cd";
push @r, $m->(), MyDesc::pattern($m);
sub mk { my $x = shift; sub { $x } }
my $c1 = mk(1);
push @r, MyDesc::refcnt($c1);
my $c2 = mk(2);
push @r, MyDesc::refcnt($c1);
undef $c2;
push @r, MyDesc::refcnt($c1);
for my $code ('MyDesc::names_here(); return 1 + $_[0]', 'MyDesc::names_here(); return join "-", @_') {
    my $s = eval "sub { $code }" or die $@;
    $s->(1);
    push @r, $MyDesc::names;
    undef $s;
}
join "|", @r`)
	want := "y|ab|y|cd|2|3|2|" +
		"nextstate,const,aelemfast,add,leavesub|" +
		"nextstate,pushmark,const,gv,rv2av,join,leavesub"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
