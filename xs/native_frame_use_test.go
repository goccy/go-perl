//go:build darwin || linux

package xs_test

import (
	"fmt"
	"strings"
	"testing"
)

// constantsModule is a module whose BOOT makes n constants as OpenGL::Modern's
// does, newCONSTSUB(stash, name, newSViv(value)) one after another: n uses
// of the current frame and its operations in one function.
func constantsModule(name string, n int) string {
	var b strings.Builder
	b.WriteString("#include \"EXTERN.h\"\n#include \"perl.h\"\n#include \"XSUB.h\"\n")
	fmt.Fprintf(&b, "XS_EXTERNAL(boot_%s) {\n    dXSBOOTARGSXSAPIVERCHK;\n    HV *stash = gv_stashpv(\"%s\", GV_ADD);\n    PERL_UNUSED_VAR(items);\n", name, name)
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "    newCONSTSUB(stash, \"C%d\", newSViv((IV)%d));\n", i, i*7)
	}
	b.WriteString("    Perl_xs_boot_epilog(aTHX_ ax);\n}\n")
	return b.String()
}

// The current frame of native code is the thread's own slot, which the
// compiler asks for once in a function and keeps (it was looked up at every
// use, with the process block tested, which made a module's code twice as
// large). A function that uses it hundreds of times, built at -O2, works on
// it at each: the 400 constants a BOOT makes as OpenGL::Modern's does are
// what perl gives.
func TestNativeFrameUsedThroughoutAFunction(t *testing.T) {
	so, out, err := compileModuleSource(t, "Consts400", constantsModule("Consts400", 400), "-O2")
	if err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}
	program := `my $sum = 0; $sum += Consts400->can("C$_")->() for 0 .. 399; join ",", Consts400::C0(), Consts400::C399(), $sum`
	if got, want := requireModule(t, "Consts400", so, program), "0,2793,558600"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
