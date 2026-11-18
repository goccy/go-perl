//go:build darwin || linux

package xs_test

import "testing"

// The fields of ops of the PMOP, PADOP and COP classes that modules read as
// members (Loo, B::Generate, Devel::Size, Syntax::Infix::Smartmatch,
// warnings::unused), the roots of the main program, CATCH_SET, the
// constructor newMETHOP and the passes over a finished optree
// (List::Keywords). Before, none of them compiled. The expected strings are
// what perl prints with the same module code.
const opClassFieldsSource = `#define PERL_USE_VOLATILE_API
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#ifndef optimize_optree
#define optimize_optree(o) Perl_optimize_optree(aTHX_ o)
#endif
#ifndef finalize_optree
#define finalize_optree(o) Perl_finalize_optree(aTHX_ o)
#endif

/* the first op of the type at or below o, in the order of the tree */
static OP *MyOpf_find(pTHX_ OP *o, int type) {
    OP *k;
    if (!o) return NULL;
    if (o->op_type == type) return o;
    if (o->op_flags & OPf_KIDS)
        for (k = cUNOPo->op_first; k; k = OpSIBLING(k)) {
            OP *r = MyOpf_find(aTHX_ k, type);
            if (r) return r;
        }
    return NULL;
}

/* the match or substitution op of a sub: its pattern and flags, and for a
 * substitution the ops of its replacement */
XS(XS_MyOpf_pm) {
    dXSARGS;
    CV *c;
    OP *m;
    if (items != 1) croak_xs_usage(cv, "code");
    c = (CV *)SvRV(ST(0));
    m = MyOpf_find(aTHX_ CvROOT(c), OP_MATCH);
    if (!m) m = MyOpf_find(aTHX_ CvROOT(c), OP_SUBST);
    if (!m) croak("no match op");
    {
        PMOP *const pm = cPMOPx(m);
        REGEXP *const re = PM_GETRE(pm);
        SV *const out = sv_2mortal(newSVpvf("%s %.*s %s", OP_NAME(m), (int)RX_PRELEN(re),
                                            RX_PRECOMP(re),
                                            (pm->op_pmflags & PMf_FOLD) ? "fold" : "case"));
        if (m->op_type == OP_SUBST) {
            OP *const root = pm->op_pmreplrootu.op_pmreplroot;
            OP *const start = pm->op_pmstashstartu.op_pmreplstart;
            sv_catpvf(out, " root=%s start=%s", root ? OP_NAME(root) : "none",
                      start ? OP_NAME(start) : "none");
        }
        ST(0) = out;
    }
    XSRETURN(1);
}

/* a hint of the statement that called, through the COP's hints hash */
XS(XS_MyOpf_hint) {
    dXSARGS;
    STRLEN len;
    const char *k;
    SV *v;
    if (items != 1) croak_xs_usage(cv, "key");
    k = SvPV(ST(0), len);
    if (!PL_curcop->cop_hints_hash) {
        ST(0) = sv_2mortal(newSVpvs("no hints"));
        XSRETURN(1);
    }
    v = Perl_refcounted_he_fetch_pvn(aTHX_ PL_curcop->cop_hints_hash, k, len, 0, 0);
    ST(0) = sv_2mortal(newSVpvf("%s exists=%s", (v && v != &PL_sv_placeholder) ? SvPV_nolen(v) : "none",
                                Perl_refcounted_he_fetch_pvn(aTHX_ PL_curcop->cop_hints_hash, k, len,
                                                             0, REFCOUNTED_HE_EXISTS)
                                    ? "yes"
                                    : "no"));
    XSRETURN(1);
}

/* PL_main_root is the op B::main_root is; a store of another op is what
 * B::main_root says next, and the root is put back */
XS(XS_MyOpf_main_root) {
    dXSARGS;
    OP *b, *root, *other;
    SV *out;
    if (items != 2) croak_xs_usage(cv, "main_root, op");
    b = INT2PTR(OP *, SvIV(SvRV(ST(0))));
    other = INT2PTR(OP *, SvIV(SvRV(ST(1))));
    root = PL_main_root;
    out = sv_2mortal(newSVpvf("%s %s start=%s", root == b ? "same" : "other", OP_NAME(root),
                              OP_NAME(PL_main_start)));
    PL_main_root = other;
    PUSHMARK(SP);
    PUTBACK;
    call_pv("MyOpf::root_name", G_SCALAR);
    SPAGAIN;
    {
        SV *const name = POPs;
        sv_catpvf(out, " then=%s", SvPV_nolen(name));
    }
    PUTBACK;
    PL_main_root = root;
    ST(0) = out;
    XSRETURN(1);
}

XS(XS_MyOpf_catch) {
    dXSARGS;
    bool old, now;
    PERL_UNUSED_VAR(items);
    old = CATCH_GET;
    CATCH_SET(TRUE);
    now = CATCH_GET;
    CATCH_SET(old);
    ST(0) = sv_2mortal(newSVpvf("%d %d %d", (int)old, (int)now, (int)CATCH_GET));
    XSRETURN(1);
}

XS(XS_MyOpf_methop) {
    dXSARGS;
    OP *m;
    SV *out;
    PERL_UNUSED_VAR(items);
    /* a method named by $_ */
    m = newMETHOP(OP_METHOD, 0, newUNOP(OP_RV2SV, 0, newGVOP(OP_GV, 0, PL_defgv)));
    out = sv_2mortal(newSVpvf("%s first=%s", OP_NAME(m),
                              m->op_type == OP_METHOD ? OP_NAME(cUNOPx(m)->op_first) : "none"));
    op_free(m);
    ST(0) = out;
    XSRETURN(1);
}

/* opt_hidden { BLOCK }: the block, hidden from the peephole optimizer as
 * List::Keywords hides it, through the passes over a finished tree; the
 * value of the keyword says whether the concatenation in it became one
 * multiconcat, which optimize_optree makes */
static Perl_keyword_plugin_t myopf_next_keyword_plugin;
static int myopf_keyword_plugin(pTHX_ char *kw, STRLEN kwlen, OP **op_ptr) {
    if (kwlen == 10 && strnEQ(kw, "opt_hidden", 10)) {
        OP *block;
        int multi;
        lex_read_space(0);
        block = parse_block(0);
        (void)LINKLIST(block);
        optimize_optree(block);
        finalize_optree(block);
        multi = MyOpf_find(aTHX_ block, OP_MULTICONCAT) != NULL;
        op_free(block);
        *op_ptr = newSVOP(OP_CONST, 0, newSVpv(multi ? "multiconcat" : "concat", 0));
        return KEYWORD_PLUGIN_EXPR;
    }
    return (*myopf_next_keyword_plugin)(aTHX_ kw, kwlen, op_ptr);
}

XS(XS_MyOpf_keywords) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    if (!myopf_next_keyword_plugin)
        wrap_keyword_plugin(myopf_keyword_plugin, &myopf_next_keyword_plugin);
    XSRETURN_EMPTY;
}

XS_EXTERNAL(boot_MyOpf) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyOpf::pm", XS_MyOpf_pm);
    newXS_deffile("MyOpf::hint", XS_MyOpf_hint);
    newXS_deffile("MyOpf::main_root", XS_MyOpf_main_root);
    newXS_deffile("MyOpf::catch", XS_MyOpf_catch);
    newXS_deffile("MyOpf::methop", XS_MyOpf_methop);
    newXS_deffile("MyOpf::keywords", XS_MyOpf_keywords);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// The fields of the match and substitution ops, the hints hash of a COP,
// the roots of the main program, CATCH_SET, a method op built by newMETHOP
// and a block put through optimize_optree and finalize_optree.
func TestNativeFieldsOfMatchOpsAndStatements(t *testing.T) {
	so := buildModuleSource(t, "MyOpf", opClassFieldsSource)
	got := requireModule(t, "MyOpf", so, `require B;
sub MyOpf::root_name { B::main_root()->name }
MyOpf::keywords();
my $y = "Y";
my @r = (MyOpf::pm(sub { $_[0] =~ /ab+c/i }), MyOpf::pm(sub { $_[0] =~ s/x+/uc $y/e }),
    MyOpf::pm(sub { $_[0] =~ s/x/y/ }));
BEGIN { $^H{"myopf/key"} = 42 }
push @r, MyOpf::hint("myopf/key"), MyOpf::hint("myopf/none");
my $keep = sub { $_[0] + 1 };
my $op = B::svref_2object($keep)->ROOT;
push @r, MyOpf::main_root(B::main_root(), $op), B::main_root()->name, MyOpf::catch();
push @r, MyOpf::methop();
push @r, eval q{ opt_hidden { my $v = 1; my $s = "<" . $v . ">" . $v; } } // "died: $@";
join "|", @r`)
	want := "match ab+c fold|subst x+ case root=substcont start=padsv|subst x case root=none start=none" +
		"|42 exists=yes|none exists=no|same leave start=enter then=leavesub|leave|0 1 0" +
		"|method first=rv2sv|multiconcat"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
