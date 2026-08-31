//go:build darwin || linux

package xs_test

import (
	"context"
	perl "github.com/goccy/go-perl"
	"os"
	"path/filepath"
	"testing"
)

// TestSyntaxKeywordMatch runs Syntax::Keyword::Match (on XS::Parse::Keyword)
// through the native SDK's parse surface: the guest keyword plugin forwards
// `match` to the host chain (reserved method -7), XPK drives the guest
// lexer/parser through the PL_parser shadow and the lex_/parse_ bridge, the
// dist builds its DISPATCHOP (a BASEOP-extended module op backed by NewOpSz)
// into a real guest optree, and each dispatch op — its op_ppaddr written to
// a host function — executes over the per-op pp hook, long after the
// compiling activation ended.
//
// It needs a prepared directory (opt-in via GOPERL_SKM_DIR) holding:
//
//	XS-Parse-Keyword.so, Syntax-Keyword-Match.so — `gperl xs build` output
//	lib/ — both dists' pure-Perl halves
func TestSyntaxKeywordMatch(t *testing.T) {
	dir := os.Getenv("GOPERL_SKM_DIR")
	if dir == "" {
		t.Skip("GOPERL_SKM_DIR not set; skipping the Syntax::Keyword::Match acceptance test")
	}
	for _, so := range []string{"XS-Parse-Keyword.so", "Syntax-Keyword-Match.so"} {
		if _, err := os.Stat(filepath.Join(dir, so)); err != nil {
			t.Fatalf("GOPERL_SKM_DIR is set but %s is missing: %v", so, err)
		}
	}

	p := newHostPerl(t)
	ctx := context.Background()

	loadXSModule(t, p, "XS::Parse::Keyword", filepath.Join(dir, "XS-Parse-Keyword.so"))
	loadXSModule(t, p, "Syntax::Keyword::Match", filepath.Join(dir, "Syntax-Keyword-Match.so"))

	if r, err := p.Eval(ctx, `sub __t_kinc { unshift @INC, @_; 1 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("inc helper: err=%v error=%v", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_kinc", perl.ValueOf(filepath.Join(dir, "lib"))); err != nil {
		t.Fatalf("add inc: %v", err)
	}

	mustEval := func(what, src, want string) {
		t.Helper()
		r, err := p.Eval(ctx, src)
		if err != nil || r.Error != nil {
			t.Fatalf("%s: err=%v ok=%v error=%v", what, err, (r.Error == nil), r.Error)
		}
		if want != "" && resultStr(r) != want {
			t.Fatalf("%s = %q, want %q", what, resultStr(r), want)
		}
	}

	// Compiling a match statement drives the whole parse bridge; the sub
	// is kept for later activations.
	mustEval("compile", `
		use Syntax::Keyword::Match;
		sub pick {
			my ($n) = @_;
			match($n : ==) {
				case(1) { return 'one' }
				case(2), case(4) { return 'even-small' }
				default { return 'other' }
			}
		}
		sub spick {
			my ($s) = @_;
			match($s : eq) {
				case("a") { return 'ay' }
				default { return 'na' }
			}
		}
		'compiled';
	`, "compiled")

	// Dispatch in the SAME activation.
	mustEval("dispatch", `join ',', pick(1), pick(2), pick(4), pick(9)`,
		"one,even-small,even-small,other")

	// Dispatch in LATER activations: the dispatch op's module state (case
	// values, branch targets) must survive the compiling activation's
	// registry teardown.
	for i := 0; i < 3; i++ {
		mustEval("later activation", `join ',', pick(1), spick('a'), spick('z')`,
			"one,ay,na")
	}

	// A match compiled by string-eval at runtime (the keyword plugin
	// firing inside an executing activation). The pragma is lexical, and a
	// top-level eval has no surrounding scope to inherit it from, so the
	// eval'd text imports it itself — same as real perl.
	mustEval("eval compile", `
		my $f = eval q{ use Syntax::Keyword::Match; sub { my ($n) = @_;
			match($n : ==) { case(7) { return 'seven' } default { return 'no' } }
		} } or die $@;
		join ',', $f->(7), $f->(8);
	`, "seven,no")

	// Errors: a die inside a case body propagates as a normal exception.
	mustEval("die in case", `
		use Syntax::Keyword::Match;
		my $err = '';
		eval { match(2 : ==) { case(2) { die "boom\n" } } 1 } or $err = $@;
		$err;
	`, "boom\n")
}
