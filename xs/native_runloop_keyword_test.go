//go:build darwin || linux

package xs_test

import "testing"

// A run loop of native code runs the ops of a block a keyword plugin built,
// or of code a checker of native code was handed, as the interpreter has
// them when they run: the peephole optimizer may have changed one after
// native code got its shadow, without native code being entered in between
// (an assignment to a lexical becomes a padsv_store). Syntax::Feature::Try's
// blocks under Test::LeakTrace's run loop died with "Modification of a
// read-only value attempted", and a string eval's assignments did nothing
// with B::Hooks::OP::Check::LeaveEval's checker installed: the loop ran the
// pp function of the type the op had when native code saw it.
const runloopKeywordSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int my_runops(pTHX) {
    while ((PL_op = PL_op->op_ppaddr(aTHX))) {
        PERL_ASYNC_CHECK();
    }
    TAINT_NOT;
    return 0;
}
static int (*next_kw)(pTHX_ char *, STRLEN, OP **);
/* mytry BLOCK: the block as an anonymous sub, with a call of
 * MyKw::end_of_block after it, handed to MyKw::run (as Syntax::Feature::Try
 * builds its blocks) */
static int my_kw(pTHX_ char *kw, STRLEN len, OP **op_ptr) {
    if (len == 5 && strnEQ(kw, "mytry", 5)) {
        I32 floor = start_subparse(0, CVf_ANON);
        OP *body = parse_block(0);
        GV *endgv = gv_fetchpvs("MyKw::end_of_block", GV_ADD, SVt_PVCV);
        GV *rungv = gv_fetchpvs("MyKw::run", GV_ADD, SVt_PVCV);
        body = op_append_elem(OP_LINESEQ, op_scope(body),
                              newUNOP(OP_ENTERSUB, OPf_STACKED, newGVOP(OP_GV, 0, endgv)));
        *op_ptr = newUNOP(OP_ENTERSUB, OPf_STACKED,
                          op_append_elem(OP_LIST, newANONSUB(floor, NULL, body),
                                         newGVOP(OP_GV, 0, rungv)));
        return KEYWORD_PLUGIN_STMT;
    }
    return next_kw(aTHX_ kw, len, op_ptr);
}
/* calls the code with G_EVAL, and answers $@ */
XS(XS_MyKw_xrun) {
    dXSARGS;
    SV *code = ST(0);
    I32 n;
    PERL_UNUSED_VAR(items);
    ENTER;
    SAVETMPS;
    PUSHMARK(SP);
    PUTBACK;
    n = call_sv(code, G_SCALAR | G_EVAL);
    SPAGAIN;
    SP -= n;
    PUTBACK;
    FREETMPS;
    LEAVE;
    ST(0) = sv_2mortal(newSVsv(ERRSV));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyKw) {
    dXSBOOTARGSXSAPIVERCHK;
    next_kw = PL_keyword_plugin;
    PL_keyword_plugin = my_kw;
    PL_runops = my_runops;
    newXS_deffile("MyKw::xrun", XS_MyKw_xrun);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeRunLoopRunsAKeywordsBlock(t *testing.T) {
	so := buildModuleSource(t, "MyKw", runloopKeywordSource)
	got := requireModule(t, "MyKw", so, `sub MyKw::run { my ($c) = @_; my $e = MyKw::xrun($c); die $e if $e; }
sub MyKw::end_of_block { 1 }
# the interpreter goes on after this eval with its next call of the loop
eval { die "start\n" };
eval q{ my $r = 0; mytry { $r = 1 }; my $s = 10; mytry { $s += 5 }; "r=$r s=$s" } // "died: $@"`)
	if want := "r=1 s=15"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

const runloopCheckerSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int my_runops(pTHX) {
    while ((PL_op = PL_op->op_ppaddr(aTHX))) {
        PERL_ASYNC_CHECK();
    }
    TAINT_NOT;
    return 0;
}
/* sees every eval's code when it is compiled, before the peephole
 * optimizer has it (as B::Hooks::OP::Check::LeaveEval's) */
static Perl_check_t prev_leaveeval;
static OP *seen_leaveeval(pTHX_ OP *o) {
    return prev_leaveeval(aTHX_ o);
}
XS_EXTERNAL(boot_MyCk) {
    dXSBOOTARGSXSAPIVERCHK;
    wrap_op_checker(OP_LEAVEEVAL, seen_leaveeval, &prev_leaveeval);
    PL_runops = my_runops;
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeRunLoopRunsCodeACheckerSaw(t *testing.T) {
	so := buildModuleSource(t, "MyCk", runloopCheckerSource)
	// cmp_ok of Test::Builder makes its comparison as a string eval does here
	got := requireModule(t, "MyCk", so, `sub cmp_le { my ($x, $y) = @_; my $check; my $ok = eval qq[\$check = sub { \$_[0] <= \$_[1] }; 1]; return "ok=$ok check=" . (defined $check ? $check->($x, $y) : "undef") . " err=$@" }
eval { die "start\n" };
my @r;
{ my $c; eval q{ $c = 5; 1 }; push @r, "lex=" . ($c // "undef"); }
{ our $g; eval q{ $g = 6; 1 }; push @r, "pkg=" . ($g // "undef"); }
{ my $c; my $v = eval q{ $c = 7; 8 }; push @r, "ret=$v lex=" . ($c // "undef"); }
{ my $c; eval q{ my $d = 9; $c = $d; 1 }; push @r, "via=" . ($c // "undef"); }
push @r, cmp_le(0, 0);
join " ", @r`)
	if want := "lex=5 pkg=6 ret=8 lex=7 via=9 ok=1 check=1 err="; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
