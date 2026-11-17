//go:build darwin || linux

package xs_test

import "testing"

// A number native code hands the interpreter as a scalar, a sub or an op
// that is none of the interpreter's makes the operation die with an error
// that names it, which eval catches, and the interpreter goes on. Before,
// the operation read the number as a scalar's head and trapped outside
// the interpreter's memory: Data::Dump::Streamer reads a scalar's head as
// a pointer and hands over what it finds (SV_TYPE), XS::JIT mortalizes a
// stray pointer (SV_MORTAL), and B::CallChecker makes a B::OP object of an
// op on its own C stack and asks B for the next op. A value an operation
// only stores is not one it reads, and passes as it does in perl (a module
// keeps structures of its own in an array: memory_beyond_4g_test.go).
func TestNativeForeignValuesDie(t *testing.T) {
	so := buildModuleSource(t, "Foreign", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static SV *foreign_value(pTHX_ const char *what) {
    static char host_memory[64];
    if (strEQ(what, "host")) return (SV *)(void *)host_memory;
    if (strEQ(what, "inside")) {
        SV *sv = sv_2mortal(newSViv(1));
        return (SV *)((char *)sv + 8);
    }
    if (strEQ(what, "freed")) {
        SV *sv = newSViv(1);
        SvREFCNT_dec(sv);
        return sv;
    }
    return NULL;
}
XS(XS_Foreign_type) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_IV((IV)SvTYPE(foreign_value(aTHX_ SvPV_nolen(ST(0)))));
}
XS(XS_Foreign_mortal) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    sv_2mortal(foreign_value(aTHX_ SvPV_nolen(ST(0))));
    XSRETURN_EMPTY;
}
XS(XS_Foreign_set) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    sv_setiv(foreign_value(aTHX_ SvPV_nolen(ST(0))), 2);
    XSRETURN_EMPTY;
}
XS(XS_Foreign_store) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    av_store((AV *)foreign_value(aTHX_ SvPV_nolen(ST(0))), 0, newSViv(1));
    XSRETURN_EMPTY;
}
XS(XS_Foreign_call) {
    dXSARGS;
    SV *code = foreign_value(aTHX_ SvPV_nolen(ST(0)));
    PERL_UNUSED_VAR(items);
    PUSHMARK(SP);
    PUTBACK;
    call_sv(code, G_DISCARD);
    XSRETURN_EMPTY;
}
XS(XS_Foreign_call_with) {
    dXSARGS;
    SV *code = ST(0);
    SV *arg = foreign_value(aTHX_ SvPV_nolen(ST(1)));
    PERL_UNUSED_VAR(items);
    PUSHMARK(SP);
    XPUSHs(arg);
    PUTBACK;
    call_sv(code, G_DISCARD);
    XSRETURN_EMPTY;
}
XS(XS_Foreign_freed_flags) {
    dXSARGS;
    SV *sv = foreign_value(aTHX_ "freed");
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("rok:%d pok:%d ok:%d", SvROK(sv) ? 1 : 0, SvPOK(sv) ? 1 : 0,
                                SvOK(sv) ? 1 : 0));
    XSRETURN(1);
}
XS(XS_Foreign_bop) {
    dXSARGS;
    static OP op_of_its_own;
    SV *rv = newSV(0);
    PERL_UNUSED_VAR(items);
    sv_setiv(newSVrv(rv, "B::OP"), (IV)(uintptr_t)&op_of_its_own);
    ST(0) = sv_2mortal(rv);
    XSRETURN(1);
}
XS_EXTERNAL(boot_Foreign) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("Foreign::type", XS_Foreign_type);
    newXS_deffile("Foreign::mortal", XS_Foreign_mortal);
    newXS_deffile("Foreign::set", XS_Foreign_set);
    newXS_deffile("Foreign::store", XS_Foreign_store);
    newXS_deffile("Foreign::call", XS_Foreign_call);
    newXS_deffile("Foreign::call_with", XS_Foreign_call_with);
    newXS_deffile("Foreign::freed_flags", XS_Foreign_freed_flags);
    newXS_deffile("Foreign::bop", XS_Foreign_bop);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	const died = `sub died { my $code = shift;
    my $r = eval { $code->(); 1 } ? "lived" : $@;
    return $r =~ /\Anative code handed the interpreter 0x[0-9a-f]+ as (.+?) in the operation (\w+) \(\d+\), and it is not one of this interpreter's at /
        ? "$2: $1" : $r }
`
	for _, tc := range []struct {
		name, code, want string
	}{
		{"an address of the host process as a scalar", `Foreign::type("host")`, "SV_TYPE: a scalar"},
		{"an address inside a scalar's head", `Foreign::type("inside")`, "SV_TYPE: a scalar"},
		{"a stray pointer mortalized", `Foreign::mortal("host")`, "SV_MORTAL: a scalar"},
		{"a freed scalar set", `Foreign::set("freed")`, "SV_SETIV: a scalar"},
		{"a stray pointer as an array", `Foreign::store("host")`, "AV_STORE: a scalar"},
		{"a null sub called", `Foreign::call("null")`, "CALL_SV: the sub to call"},
		{"a stray pointer called", `Foreign::call("host")`, "CALL_SV: a scalar"},
		{"a stray pointer as an argument", `Foreign::call_with(sub {}, "host")`, "CALL_SV: a scalar"},
		{"a null argument", `Foreign::call_with(sub {}, "null")`, "CALL_SV: an argument"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := requireModule(t, "Foreign", so, died+`my $r = died(sub { `+tc.code+` });
my @a = (1, 2, 3);
"$r; then " . join(",", map { $_ * 2 } @a)`)
			if want := tc.want + "; then 2,4,6"; got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
	// the type of a freed scalar is perl's: reading the head alone is
	// allowed, and it says the scalar is freed
	if got, want := requireModule(t, "Foreign", so, `Foreign::type("freed")`), "255"; got != want {
		t.Errorf("type of a freed scalar: got %q, want %q", got, want)
	}
	// so are its flags, which it has none of left: Tk asks SvROK and SvPOK
	// of a scalar its Tcl_SetVar has just freed
	if got, want := requireModule(t, "Foreign", so, `Foreign::freed_flags()`), "rok:0 pok:0 ok:0"; got != want {
		t.Errorf("flags of a freed scalar: got %q, want %q", got, want)
	}
	if got, want := requireModule(t, "Foreign", so, `require B;
my $op = Foreign::bop();
my $r = eval { $op->next; 1 } ? "lived" : $@;
$r =~ /\Ago-perl gives native code no access to this interpreter internal: a B::OP object of an op of native code \(0x[0-9a-f]+\) at / ? "died" : $r`), "died"; got != want {
		t.Errorf("B::OP of an op of native code: got %q, want %q", got, want)
	}
}
