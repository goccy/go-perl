//go:build darwin || linux

package xs_test

import "testing"

// The variables of the parser native code reads by perl's names beyond
// ppport.h's (PL_oldbufptr, PL_oldoldbufptr, PL_linestart and the like,
// which perl.h gives the core only, so a module defines them itself over
// PL_parser or ppport.h's D_PPP_my_PL_parser_var, as signatures.xs does):
// where the lexer is while a BEGIN block runs. signatures did not build
// ("oldbufptr undeclared"). What is expected is what perl 5.36 on Linux
// gives with the same module.
func TestNativeCodeReadsTheParsersVariables(t *testing.T) {
	so := buildModuleSource(t, "ParsePos", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* perl's own headers give these names to the core only (ppport.h gives
 * PL_bufptr) */
#ifndef PL_bufptr
#define PL_bufptr (PL_parser->bufptr)
#endif
#ifndef PL_oldbufptr
#define PL_oldbufptr (PL_parser->oldbufptr)
#endif
#ifndef PL_oldoldbufptr
#define PL_oldoldbufptr (PL_parser->oldoldbufptr)
#endif
#ifndef PL_linestart
#define PL_linestart (PL_parser->linestart)
#endif
/* pos(): where the lexer is in the line it reads, called while a BEGIN
 * block runs: bufptr, oldbufptr and oldoldbufptr from the start of the
 * line, and the text from oldbufptr to bufptr, as signatures.xs reads
 * PL_oldbufptr */
XS(XS_ParsePos_pos) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    if (!PL_parser) {
        ST(0) = sv_2mortal(newSVpvs("no parser"));
        XSRETURN(1);
    }
    ST(0) = sv_2mortal(newSVpvf("%d %d %d [%.*s]", (int)(PL_bufptr - PL_linestart),
                                (int)(PL_oldbufptr - PL_linestart),
                                (int)(PL_oldoldbufptr - PL_linestart),
                                (int)(PL_bufptr - PL_oldbufptr), PL_oldbufptr));
    XSRETURN(1);
}
XS_EXTERNAL(boot_ParsePos) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("ParsePos::pos", XS_ParsePos_pos, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my ($r1, $r2);
eval q{my $x = 1; BEGIN { $r1 = ParsePos::pos() } my $y = 2;
  sub foo { BEGIN { $r2 = ParsePos::pos() } } 1} or die $@;
"$r1 | $r2"`
	if got, want := requireModule(t, "ParsePos", so, program), "42 41 39 [}] | 97 96 94 [}]"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
