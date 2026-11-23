//go:build darwin || linux

package xs_test

import "testing"

// inlineModule is what Inline::C and Inline::CPP generate around a C
// function: the XSUB pushes a mark of its own (`temp = PL_markstack_ptr++`),
// calls the function, and returns nothing when the function did not take
// the mark (a void function), or the values the function left when it did
// (a function using dXSARGS, Inline_Stack_Vars).
const inlineModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int calls;
static void plain(int a) { calls += a; }
static int with_stack(int a) {
    dXSARGS;
    (void)items;
    sp = mark;
    XPUSHs(sv_2mortal(newSViv(a * 2)));
    XPUSHs(sv_2mortal(newSViv(a * 3)));
    PUTBACK;
    return 2;
}
#define INLINE_CALL(call)                         \
    {                                             \
        I32 *temp;                                \
        SP -= items;                              \
        temp = PL_markstack_ptr++;                \
        call;                                     \
        if (PL_markstack_ptr != temp) {           \
            PL_markstack_ptr = temp;              \
            XSRETURN_EMPTY;                       \
        }                                         \
        return;                                   \
    }
XS(XS_InlineShapes_void_fn) {
    dXSARGS;
    INLINE_CALL(plain((int)SvIV(ST(0))))
}
XS(XS_InlineShapes_stack_fn) {
    dXSARGS;
    INLINE_CALL(with_stack((int)SvIV(ST(0))))
}
XS(XS_InlineShapes_calls) {
    dXSARGS;
    (void)items;
    ST(0) = sv_2mortal(newSViv(calls));
    XSRETURN(1);
}
XS_EXTERNAL(boot_InlineShapes) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("InlineShapes::void_fn", XS_InlineShapes_void_fn);
    newXS_deffile("InlineShapes::stack_fn", XS_InlineShapes_stack_fn);
    newXS_deffile("InlineShapes::calls", XS_InlineShapes_calls);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// The XSUBs Inline generates around a void function return nothing, and
// around a function that uses dXSARGS the values it left. Before, the mark
// pointer read after `PL_markstack_ptr++` was the frame's own again, the
// XSUB took the void function for one that used the stack, and returned
// its argument (Inline-CPP's t/grammar/02scope.t got 10 for a void member
// function).
func TestNativeInlineVoidAndStackFunctions(t *testing.T) {
	so := buildModuleSource(t, "InlineShapes", inlineModule)
	code := `my @v = InlineShapes::void_fn(10); my @s = InlineShapes::stack_fn(5); join ";", scalar(@v), "@s", InlineShapes::calls()`
	if got, want := requireModule(t, "InlineShapes", so, code), "0;10 15;10"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// dXSARGS in a C++ function that returns a value compiles (Inline::CPP's
// functions with variable arguments use it): C++ makes `return;` an error
// there, which the SDK's dXSARGS had in its croak branch.
func TestNativeDXSARGSInACPlusPlusFunction(t *testing.T) {
	_, out, err := compileModuleSource(t, "CxxStack", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int count_args(int unused, ...) {
    dXSARGS;
    (void)unused;
    (void)sp;
    return (int)items;
}
extern "C" XS_EXTERNAL(boot_CxxStack) {
    dXSBOOTARGSXSAPIVERCHK;
    (void)count_args;
    Perl_xs_boot_epilog(aTHX_ ax);
}
`, "-x", "c++", "-std=gnu++17")
	if err != nil {
		t.Fatalf("a C++ function with dXSARGS did not compile: %v\n%s", err, out)
	}
}
