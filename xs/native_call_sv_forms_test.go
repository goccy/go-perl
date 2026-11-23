//go:build darwin || linux

package xs_test

import "testing"

// call_sv of native code takes the code in every form perl's does (a CV, a
// reference to one, a glob, a sub's name, a method name with G_METHOD or
// G_METHOD_NAMED) in every context, with and without G_EVAL and
// G_DISCARD, and leaves the stack where perl leaves it; and Tk's
// Lang_catch, a mortal CV of the module's own whose XSUB is native code,
// called with G_VOID|G_EVAL and no mark of its own, leaves the stack where
// it was. What is expected is what perl 5.36 on Linux gives with the same
// module.
func TestNativeCallSvForms(t *testing.T) {
	so := buildModuleSource(t, "CallForms", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* the flags by name */
static I32 flags_named(const char *s) {
    I32 f = strstr(s, "list") ? G_LIST : strstr(s, "void") ? G_VOID : G_SCALAR;
    if (strstr(s, "eval")) f |= G_EVAL;
    if (strstr(s, "discard")) f |= G_DISCARD;
    if (strstr(s, "method")) f |= G_METHOD;
    return f;
}
/* call(form, flags): main::target (or the method "meth" of "Klass") called
 * through call_sv with the code as the form says: cv, rv, gv, name, method,
 * method_named; arguments 1, 2 (after the invocant for a method). ->
 * "count:results:moved" where moved is how far the stack pointer is from
 * where it was before the arguments. */
XS(XS_CallForms_call) {
    dXSARGS;
    const char *form = SvPV_nolen(ST(0));
    const I32 flags = flags_named(SvPV_nolen(ST(1)));
    SV *code;
    SV **before;
    I32 count, i;
    SV *out = newSVpvs("");
    PERL_UNUSED_VAR(items);
    if (strEQ(form, "cv")) code = (SV *)get_cv("main::target", 0);
    else if (strEQ(form, "rv")) code = sv_2mortal(newRV_inc((SV *)get_cv("main::target", 0)));
    else if (strEQ(form, "gv")) code = (SV *)gv_fetchpvs("main::target", 0, SVt_PVCV);
    else if (strEQ(form, "name")) code = sv_2mortal(newSVpvs("main::target"));
    else code = sv_2mortal(newSVpvs("meth"));
    SP -= items;
    PUTBACK;
    ENTER;
    SAVETMPS;
    before = PL_stack_sp;
    PUSHMARK(SP);
    if (strncmp(form, "method", 6) == 0) XPUSHs(sv_2mortal(newSVpvs("Klass")));
    mXPUSHi(1);
    mXPUSHi(2);
    PUTBACK;
    if (strEQ(form, "method_named"))
        count = call_sv(code, (flags & ~G_METHOD) | G_METHOD_NAMED);
    else
        count = call_sv(code, flags | (strEQ(form, "method") ? G_METHOD : 0));
    SPAGAIN;
    sv_catpvf(out, "%d:", (int)count);
    for (i = count - 1; i >= 0; i--) {
        SV *r = *(SP - i);
        sv_catpvf(out, "%s%s", SvOK(r) ? SvPV_nolen(r) : "undef", i ? "," : "");
    }
    sv_catpvf(out, ":%d", (int)(SP - before) - (int)count);
    SP -= count;
    PUTBACK;
    if ((flags & G_EVAL) && SvTRUE(ERRSV)) sv_catpvf(out, ":died %s", SvPV_nolen(ERRSV));
    FREETMPS;
    LEAVE;
    SPAGAIN;
    XPUSHs(sv_2mortal(out));
    XSRETURN(1);
}
/* the native code Tk runs with G_EVAL through a CV of its own making */
static void quiet(pTHX_ CV *cv) {
    dMARK;
    dAX;
    PERL_UNUSED_VAR(cv);
    PERL_UNUSED_VAR(mark);
    XSRETURN(0);
}
/* catch(): Tk's Lang_catch: a mortal CV whose XSUB is native code, called
 * with call_sv(cv, G_VOID|G_EVAL) and no mark of its own pushed; -> whether
 * the stack pointer moved, and the count */
XS(XS_CallForms_catch) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    SP -= items;
    PUTBACK;
    {
        dSP;
        CV *cv = (CV *)sv_newmortal();
        int count;
        SV **oldSP = sp;
        sv_upgrade((SV *)cv, SVt_PVCV);
        CvXSUB(cv) = quiet;
        CvXSUBANY(cv).any_ptr = NULL;
        CvISXSUB_on(cv);
        count = call_sv((SV *)cv, G_VOID | G_EVAL);
        SPAGAIN;
        XPUSHs(sv_2mortal(newSVpvf("count %d, %s", count, sp != oldSP ? "moved" : "stayed")));
        PUTBACK;
    }
    XSRETURN(1);
}
XS_EXTERNAL(boot_CallForms) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("CallForms::call", XS_CallForms_call, __FILE__);
    newXS("CallForms::catch", XS_CallForms_catch, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `sub target { my $w = wantarray; defined $w ? ($w ? ("l", @_) : "s" . join("", @_)) : () }
package Klass;
sub meth { my $class = shift; my $w = wantarray; defined $w ? ($w ? ("m$class", @_) : "m" . join("", @_)) : () }
package main;
my @got;
for my $form (qw(cv rv gv name method method_named)) {
    push @got, join " ", "$form:", map { CallForms::call($form, $_) } "scalar", "scalar eval", "list", "list eval", "void", "void discard", "scalar discard eval";
}
my $c = CallForms::catch();
push @got, "catch: $c";
join " | ", @got`
	if got, want := requireModule(t, "CallForms", so, program), "cv: 1:s12:0 1:s12:0 3:l,1,2:0 3:l,1,2:0 0::0 0::0 0::0 | rv: 1:s12:0 1:s12:0 3:l,1,2:0 3:l,1,2:0 0::0 0::0 0::0 | gv: 1:s12:0 1:s12:0 3:l,1,2:0 3:l,1,2:0 0::0 0::0 0::0 | name: 1:s12:0 1:s12:0 3:l,1,2:0 3:l,1,2:0 0::0 0::0 0::0 | method: 1:m12:0 1:m12:0 3:mKlass,1,2:0 3:mKlass,1,2:0 0::0 0::0 0::0 | method_named: 1:m12:0 1:m12:0 3:mKlass,1,2:0 3:mKlass,1,2:0 0::0 0::0 0::0 | catch: count 0, stayed"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
