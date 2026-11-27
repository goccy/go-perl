//go:build darwin || linux

package xs_test

import "testing"

// Native code works on the interpreter's own stack and mark stack, as on
// perl: an XSUB's arguments and what it returns are there, a destructor
// that runs inside an XSUB's call pushes onto the stack the XSUB finds when
// it looks again (SPAGAIN), and a stack pointer kept from one native call
// is the stack pointer in another (Scope::Upper's unwind keeps the one its
// XSUB had and puts it back in a destructor). The rest is what any XSUB
// does with the stack, which has to stay as it is: EXTEND far past the
// stack's room, call_sv inside call_sv, a croak deep inside both, an XSUB
// that returns fewer or more values than it was given, in list and in
// scalar context, called deep in the interpreter's stack. Before, each
// native call had a copy of the stack of its own: what a destructor pushed
// was gone, and a kept stack pointer pointed into a copy that was gone.
const frameStackSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* many(n): 0 .. n-1, pushed one at a time */
XS(XS_MyFs_many) {
    dXSARGS;
    const IV n = SvIV(ST(0));
    IV i;
    SP -= items;
    for (i = 0; i < n; i++) mXPUSHi(i);
    PUTBACK;
}
/* fewer(...): the number of arguments, alone */
XS(XS_MyFs_fewer) {
    dXSARGS;
    ST(0) = sv_2mortal(newSViv(items));
    XSRETURN(1);
}
/* more(...): the number of arguments, then x and y */
XS(XS_MyFs_more) {
    dXSARGS;
    EXTEND(SP, 3);
    ST(0) = sv_2mortal(newSViv(items));
    ST(1) = sv_2mortal(newSVpvs("x"));
    ST(2) = sv_2mortal(newSVpvs("y"));
    XSRETURN(3);
}
/* call(code, args...): what the code returns in list context, and how many */
XS(XS_MyFs_call) {
    dXSARGS;
    SV *const code = ST(0);
    I32 i, n;
    SV **res;
    PUSHMARK(SP);
    for (i = 1; i < items; i++) XPUSHs(ST(i));
    PUTBACK;
    n = call_sv(code, G_ARRAY);
    SPAGAIN;
    res = SP - n + 1;
    for (i = 0; i < n; i++) ST(i) = res[i];
    ST(n) = sv_2mortal(newSViv(n));
    XSRETURN(n + 1);
}
XS(XS_MyFs_boom) {
    dXSARGS;
    croak("%s", items ? SvPV_nolen(ST(0)) : "boom");
}
/* dtor_push(): a destructor of a scope the XSUB opens and closes pushes
 * 42, and the XSUB pushes 43 after it */
static void push42(pTHX_ void *p) {
    dSP;
    PERL_UNUSED_VAR(p);
    XPUSHs(sv_2mortal(newSViv(42)));
    PUTBACK;
}
XS(XS_MyFs_dtor_push) {
    dXSARGS;
    SP -= items;
    PUTBACK;
    ENTER;
    SAVEDESTRUCTOR_X(push42, NULL);
    LEAVE;
    SPAGAIN;
    XPUSHs(sv_2mortal(newSViv(43)));
    PUTBACK;
}
/* kept(): the stack pointer as it is, kept; back(...): the stack pointer
 * is the kept one again, put back by a destructor of a scope it opens and
 * closes, and the three values there are what it reads */
static SV **kept_sp;
XS(XS_MyFs_kept) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    kept_sp = PL_stack_sp;
    XSRETURN(0);
}
static void back_dtor(pTHX_ void *p) {
    PERL_UNUSED_VAR(p);
    PL_stack_sp = kept_sp;
}
XS(XS_MyFs_back) {
    dXSARGS;
    SV **base;
    I32 i, n = 3;
    SV *out = sv_2mortal(newSVpvs(""));
    PERL_UNUSED_VAR(items);
    ENTER;
    SAVEDESTRUCTOR_X(back_dtor, NULL);
    LEAVE;
    base = PL_stack_sp - n + 1;
    for (i = 0; i < n; i++) sv_catpvf(out, "%s%s", i ? "," : "", SvPV_nolen(base[i]));
    ST(0) = out;
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyFs) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyFs::many", XS_MyFs_many);
    newXS_deffile("MyFs::fewer", XS_MyFs_fewer);
    newXS_deffile("MyFs::more", XS_MyFs_more);
    newXS_deffile("MyFs::call", XS_MyFs_call);
    newXS_deffile("MyFs::boom", XS_MyFs_boom);
    newXS_deffile("MyFs::dtor_push", XS_MyFs_dtor_push);
    newXS_deffile("MyFs::kept", XS_MyFs_kept);
    newXS_deffile("MyFs::back", XS_MyFs_back);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeWorksOnTheInterpretersStack(t *testing.T) {
	so := buildModuleSource(t, "MyFs", frameStackSource)
	for _, tc := range []struct{ name, code, want string }{
		{"EXTEND far past the stack's room",
			`my @m = MyFs::many(200000); my $n = () = MyFs::many(300000); join ",", scalar(@m), $m[-1], $n`,
			"200000,199999,300000"},
		{"fewer and more values than arguments",
			`my $s = MyFs::more(1); my $t = MyFs::fewer(4, 5, 6); join ",", 0, MyFs::fewer(5, 6, 7), 8, MyFs::more(1), 9, $s, $t`,
			"0,3,8,1,x,y,9,y,3"},
		{"call_sv inside call_sv",
			`join ",", MyFs::call(sub { map { $_ * 2 } MyFs::call(sub { @_, 3 }, 1, 2) }, 9)`,
			"2,4,6,6,4"},
		{"a croak deep inside",
			`my @r = (1, (eval { MyFs::call(sub { MyFs::call(sub { MyFs::boom("deep") }) }); 1 } ? "lived" : $@ =~ /^deep at / ? "died" : $@), 2, 3); join ",", @r, MyFs::fewer(@r)`,
			"1,died,2,3,4"},
		{"deep in the interpreter's stack",
			`my @x = map { MyFs::fewer(1 .. $_ % 5) } 1 .. 50000; my $sum = 0; $sum += $_ for @x; $sum`,
			"100000"},
		{"a destructor pushes onto the stack",
			`join ",", 1, MyFs::dtor_push(7, 8), 2`,
			"1,42,43,2"},
		{"a stack pointer kept from another call",
			`join ",", "a", "b", "c", MyFs::kept(), MyFs::back(1, 2)`,
			"a,b,c,a,b,c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := requireModule(t, "MyFs", so, tc.code); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
