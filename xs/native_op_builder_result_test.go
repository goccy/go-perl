//go:build linux

package xs_test

import "testing"

// A builder of ops that ends without a return statement after its last
// call hands back what that call returned, as with perl's functions:
// Syntax::Operator::ExistsOr's new_op ends with op_contextualize. The
// SDK's builders were inlined into such a function, and its result was not
// where the caller read it (the stack protector of an inlined buffer
// cleared it): the operator's whole expression was lost from the program.
// The module is built optimised and with the stack protector, as a
// distribution's build makes it, by gcc, which leaves the last call's
// result in place (clang takes the end of such a function for unreachable,
// with perl as with the SDK). What is expected is what perl 5.36 on Linux
// gives with the same module.
func TestNativeOpBuilderResultOfTheLastCall(t *testing.T) {
	so, out, err := compileModuleSource(t, "FallRet", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* A builder of ops that leaves its result where the last call left it, as
 * Syntax::Operator::ExistsOr's does: no return statement after the call
 * that builds the op. */
#pragma GCC diagnostic ignored "-Wreturn-type"
static OP *build_op(pTHX_ IV n) __attribute__((noinline));
static OP *build_op(pTHX_ IV n) {
    op_contextualize(newUNOP(OP_NEGATE, 0, newSVOP(OP_CONST, 0, newSViv(n))), G_SCALAR);
}
static OP *(*volatile builder)(pTHX_ IV) = build_op;
static Perl_keyword_plugin_t next_keyword_plugin;
/* the keyword fortytwo is the expression -42, as the builder makes it */
static int keyword(pTHX_ char *kw, STRLEN len, OP **op_ptr) {
    if (len == 8 && memEQ(kw, "fortytwo", 8)) {
        *op_ptr = builder(aTHX_ 42);
        return KEYWORD_PLUGIN_EXPR;
    }
    return next_keyword_plugin(aTHX_ kw, len, op_ptr);
}
XS_EXTERNAL(boot_FallRet) {
    dXSBOOTARGSXSAPIVERCHK;
    wrap_keyword_plugin(keyword, &next_keyword_plugin);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`, "-O2", "-fstack-protector-strong", "-Wno-return-type")
	if err != nil {
		t.Fatalf("compile FallRet.so: %v\n%s", err, out)
	}
	program := `my $r = eval q{ my $x = fortytwo; $x + 1 };
defined $r ? $r : "died: $@"`
	if got, want := requireModule(t, "FallRet", so, program), "-41"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
