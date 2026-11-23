//go:build darwin || linux

package xs_test

import "testing"

// CopLINE(PL_curcop) in a keyword plugin is the line the lexer is on after
// the lines it has read, and CopLINE_set of PL_curcop is the line a croak
// names, as Quote::Code numbers the lines of a here-doc and names the line
// one began on: the SDK kept the line it first read (the lines of a
// here-doc all came out as one) and the line it set never reached the
// interpreter (the croak named the line the lexer stopped on). What is
// expected is what perl 5.36 on Linux gives with the same module.
func TestNativeLexerLineOfPLCurcop(t *testing.T) {
	so := buildModuleSource(t, "CopLine", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static Perl_keyword_plugin_t next_keyword_plugin;
/* lines: the rest of the line and the lines after it up to one that is
 * "END" are read, as Quote::Code reads a here-doc; the expression is
 * "<line it began on>-<line the lexer is on after>". badlines reads the
 * same and croaks naming the line it began on, as Quote::Code does when
 * the here-doc has no end. */
static int keyword(pTHX_ char *kw, STRLEN len, OP **op_ptr) {
    const int bad = len == 8 && memEQ(kw, "badlines", 8);
    if (bad || (len == 5 && memEQ(kw, "lines", 5))) {
        const line_t start = CopLINE(PL_curcop);
        I32 c;
        line_t end;
        while ((c = lex_read_unichar(0)) != -1 && c != '\n')
            ;
        for (;;) {
            char buf[64];
            int n = 0;
            while ((c = lex_read_unichar(0)) != -1 && c != '\n')
                if (n < 63) buf[n++] = (char)c;
            buf[n] = 0;
            if (c == -1 || strEQ(buf, "END")) break;
        }
        end = CopLINE(PL_curcop);
        if (bad) {
            CopLINE_set(PL_curcop, start);
            croak("no end of what began on line %d", (int)start);
        }
        *op_ptr = newSVOP(OP_CONST, 0, newSVpvf("%d-%d", (int)start, (int)end));
        return KEYWORD_PLUGIN_EXPR;
    }
    return next_keyword_plugin(aTHX_ kw, len, op_ptr);
}
XS_EXTERNAL(boot_CopLine) {
    dXSBOOTARGSXSAPIVERCHK;
    wrap_keyword_plugin(keyword, &next_keyword_plugin);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $r = eval "\n\nmy \$x = lines\nfoo\nbar\nEND\n;\n\$x . ',' . __LINE__";
my $e = eval "\n\nbadlines\nfoo\nbar\nbaz\n";
my ($began, $at) = $@ =~ /began on line (\d+) at \(eval \d+\) line (\d+)/;
join "|", $r // "died: $@", $began // '?', $at // '?'`
	if got, want := requireModule(t, "CopLine", so, program), "3-7,8|3|3"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
