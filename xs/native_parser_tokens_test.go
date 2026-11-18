//go:build darwin || linux

package xs_test

import "testing"

// Native code reads the next token of the source with yylex while the
// interpreter compiles, and gives it back to the lexer as the token to come
// (PL_parser's nexttype, nextval and nexttoke), with yylval as it was, as
// Data::Alias does to put a token of its own before a block: the program
// goes on as if nothing had been read. Before, yylex and those fields were
// not in the SDK.
const parserTokensSource = `/* the token kinds and yylex are perl's to its core: PERL_CORE around
 * perl.h, as Data::Alias has it */
#define PERL_CORE
#include "EXTERN.h"
#include "perl.h"
#undef PERL_CORE
#include "XSUB.h"
#if PERL_REVISION == 5 && PERL_VERSION < 37
#define KW_MY MY /* the token kind's older name */
#endif
/* peek(): at BEGIN, whether the next token is "my", and how many tokens are
 * to come once it is given back */
XS(XS_MyPx_peek) {
    dXSARGS;
    YYSTYPE save;
    int tok;
    PERL_UNUSED_VAR(items);
    save = PL_parser->yylval;
    tok = yylex();
    PL_parser->nexttype[PL_parser->nexttoke] = tok;
    PL_parser->nextval[PL_parser->nexttoke] = PL_parser->yylval;
    PL_parser->nexttoke++;
    PL_parser->yylval = save;
    ST(0) = sv_2mortal(Perl_newSVpvf(aTHX_ "%d %d", tok == KW_MY, (int)PL_parser->nexttoke));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyPx) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyPx::peek", XS_MyPx_peek);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeReadsTheNextToken(t *testing.T) {
	so := buildModuleSource(t, "MyPx", parserTokensSource)
	got := requireModule(t, "MyPx", so, `my $r = eval q{
our $seen;
BEGIN { $seen = MyPx::peek() }
my $x = 6 * 7;
"$seen|$x";
};
$r // "died: $@"`)
	if want := "1 1|42"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
