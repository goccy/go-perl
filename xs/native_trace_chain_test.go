//go:build darwin || linux

package xs_test

import "testing"

// Tk traces a variable the way this module does (tkGlue.c's
// Tcl_TraceVar2): it takes the value's chain of magic off whole
// (SvMAGIC(sv) = NULL), attaches uvar magic of its own with sv_magic,
// which would not add a second of its type, puts its ufuncs in, and puts
// the chain back with the new magic at its end. A variable may have many
// traces (the widgets of a radio group share one) and the magic of the
// interpreter too (a tie). Every trace sees an assignment, in order; an
// untraced one (taken off the chain as Tcl_UntraceVar2 does) no more; and
// a variable with no trace left is assigned without one.
//
// Before, the store of the chain without the interpreter's magic failed
// ("native code unlinked magic 'q' ..."), the interpreter called only the
// first uvar magic of the chain, whichever trace that was (so a value went
// to another widget's option: "unknown color name"), and an assignment
// after the last trace was taken off died ("no uvar magic on the value").
func TestNativeTraceTakesTheChainOffAndBack(t *testing.T) {
	so := buildModuleSource(t, "Trace", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static SV *trace_log;
static I32 trace_val(pTHX_ IV ix, SV *sv) { PERL_UNUSED_VAR(ix); PERL_UNUSED_VAR(sv); return 0; }
static I32 trace_set(pTHX_ IV ix, SV *sv) {
    sv_catpvf(trace_log, "%s%s=%s", SvCUR(trace_log) ? "," : "", (const char *)INT2PTR(char *, ix),
              SvPV_nolen(sv));
    return 0;
}
XS(XS_Trace_trace) {
    dXSARGS;
    SV *sv = SvRV(ST(0));
    const char *name = savepv(SvPV_nolen(ST(1)));
    struct ufuncs *ufp;
    MAGIC *mg, *mg_list, **mgp;
    PERL_UNUSED_VAR(items);
    (void)SvUPGRADE(sv, SVt_PVMG);
    mg_list = SvMAGIC(sv);
    SvMAGIC(sv) = NULL;
    sv_magic(sv, 0, PERL_MAGIC_uvar, 0, 0);
    Newz(666, ufp, 1, struct ufuncs);
    ufp->uf_val = trace_val;
    ufp->uf_set = trace_set;
    ufp->uf_index = PTR2IV(name);
    mg = SvMAGIC(sv);
    mg->mg_ptr = (char *)ufp;
    mg->mg_len = sizeof(struct ufuncs);
    SvMAGIC(sv) = mg_list;
    mgp = &SvMAGIC(sv);
    while ((mg_list = *mgp)) mgp = &mg_list->mg_moremagic;
    *mgp = mg;
    if (!SvMAGICAL(sv)) croak("not magical");
    XSRETURN_EMPTY;
}
XS(XS_Trace_untrace) {
    dXSARGS;
    SV *sv = SvRV(ST(0));
    const char *name = SvPV_nolen(ST(1));
    MAGIC **mgp, *mg;
    PERL_UNUSED_VAR(items);
    if (SvMAGICAL(sv) && (mgp = &SvMAGIC(sv))) {
        for (mg = *mgp; mg; mg = *mgp) {
            if (mg->mg_type == PERL_MAGIC_uvar && mg->mg_ptr &&
                mg->mg_len == sizeof(struct ufuncs) &&
                ((struct ufuncs *)mg->mg_ptr)->uf_set == trace_set &&
                strEQ(INT2PTR(char *, ((struct ufuncs *)mg->mg_ptr)->uf_index), name)) {
                *mgp = mg->mg_moremagic;
                Safefree(mg->mg_ptr);
                mg->mg_ptr = NULL;
                Safefree(mg);
            } else {
                mgp = &mg->mg_moremagic;
            }
        }
    }
    XSRETURN_EMPTY;
}
XS(XS_Trace_log) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVsv(trace_log));
    sv_setpvs(trace_log, "");
    XSRETURN(1);
}
XS_EXTERNAL(boot_Trace) {
    dXSBOOTARGSXSAPIVERCHK;
    trace_log = newSVpvs(""); /* this interpreter's */
    newXS_deffile("Trace::trace", XS_Trace_trace);
    newXS_deffile("Trace::untrace", XS_Trace_untrace);
    newXS_deffile("Trace::log", XS_Trace_log);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	const tied = `package Kept; sub TIESCALAR { my $v; bless \$v } sub FETCH { ${$_[0]} } sub STORE { ${$_[0]} = "<$_[1]>" }
package main;
`
	for _, tc := range []struct{ name, setup string }{
		{"a plain variable", `my $v = 0;`},
		{"a tied variable", tied + `tie my $v, "Kept";`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := requireModule(t, "Trace", so, tc.setup+`
my @r;
Trace::trace(\$v, "a"); Trace::trace(\$v, "b");
$v = 1; push @r, Trace::log();
Trace::untrace(\$v, "a");
$v = 2; push @r, Trace::log();
Trace::untrace(\$v, "b");
$v = 3; push @r, Trace::log() || "none";
push @r, "v=$v", tied($v) ? "tied" : "plain";
join " ", @r`)
			want := "a=1,b=1 b=2 none v=3 plain"
			if tc.name == "a tied variable" {
				// the magic of the value is off while a magic's set runs
				// (perl's save_magic), so a trace reads the value stored
				want = "a=1,b=1 b=2 none v=<3> tied"
			}
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
