//go:build darwin || linux

package xs_test

import "testing"

// A croak in an XSUB of a module compiled as C++ dies as perl's croak
// does, which eval catches, both in the XSUB the interpreter called and in
// a function of the module with dXSARGS of its own that returns a value
// (Inline::CPP's functions with variable arguments). Before, the croak of
// the XSUB the interpreter called went on to an outer activation there was
// none of, and the process crashed (Cavil-Matcher, a C++ module, croaks on
// a corrupt segment file).
func TestNativeCxxModuleCroaks(t *testing.T) {
	so, out, err := compileModuleSource(t, "CxxCroak", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int inner(pTHX_ CV *cv) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    PERL_UNUSED_VAR(cv);
    croak("inner says no");
    return 1;
}
XS(XS_CxxCroak_croak) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    croak("outer says no");
}
XS(XS_CxxCroak_nested) {
    dXSARGS;
    int r;
    PERL_UNUSED_VAR(items);
    PUSHMARK(SP);
    PUTBACK;
    r = inner(aTHX_ cv);
    XSRETURN_IV(r);
}
extern "C" XS_EXTERNAL(boot_CxxCroak) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("CxxCroak::croak", XS_CxxCroak_croak);
    newXS_deffile("CxxCroak::nested", XS_CxxCroak_nested);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`, "-x", "c++", "-std=gnu++17", "-fno-exceptions") // no C++ runtime to link
	if err != nil {
		t.Fatalf("compile CxxCroak.so: %v\n%s", err, out)
	}
	for _, tc := range []struct{ name, code, want string }{
		{"the XSUB the interpreter called", `my $r = eval { CxxCroak::croak(); 1 } ? "lived" : $@; my @a = (1, 2); ($r =~ /^outer says no at / ? "died " : $r) . join(",", @a)`,
			"died 1,2"},
		{"a function with dXSARGS of its own", `my $r = eval { CxxCroak::nested(); 1 } ? "lived" : $@; my @a = (3, 4); ($r =~ /^inner says no at / ? "died " : $r) . join(",", @a)`,
			"died 3,4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := requireModule(t, "CxxCroak", so, tc.code); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
