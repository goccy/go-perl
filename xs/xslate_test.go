//go:build darwin || linux

package xs_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-perl/xs"
)

// TestTextXslate runs a real-world XS dist — Text::Xslate — through the
// native SDK: MAGIC-backed engine state, the unified perl stack, alias
// XSUBs, the PL_ppaddr scratch-op path (the range operator), macro frames,
// builtin methods, and the die-hook error path.
//
// It needs a prepared directory (opt-in via GOPERL_XSLATE_DIR) holding:
//
//	TextXslate.so  — the dist's two xsubpp outputs compiled against
//	                 sdk/include with the host C compiler, matching this
//	                 test binary's architecture:
//	                   cc -shared -fPIC -O2 -I <go-perl>/xs/sdk/include \
//	                      -I <dist>/src -I <dist> \
//	                      -o TextXslate.so Text-Xslate.c xslate_methods.c
//	lib/           — the dist's pure-Perl half (Text-Xslate-*/lib)
//	deps/          — cpanm -L output for the runtime deps, vendored
//	                 pure-Perl (MOUSE_PUREPERL=1 PUREPERL_ONLY=1 cpanm -L
//	                 deps --notest Mouse Data::MessagePack), i.e. holding
//	                 lib/perl5[/<archname>]
func TestTextXslate(t *testing.T) {
	dir := os.Getenv("GOPERL_XSLATE_DIR")
	if dir == "" {
		t.Skip("GOPERL_XSLATE_DIR not set; skipping the Text::Xslate acceptance test")
	}
	so := filepath.Join(dir, "TextXslate.so")
	if _, err := os.Stat(so); err != nil {
		t.Fatalf("GOPERL_XSLATE_DIR is set but %s is missing: %v", so, err)
	}

	p := newHostPerl(t)
	ctx := context.Background()

	if err := xs.Load(p, "Text::Xslate", so); err != nil {
		t.Fatalf("Load: %v", err)
	}

	inc := []any{filepath.Join(dir, "lib"), filepath.Join(dir, "deps", "lib", "perl5")}
	entries, _ := os.ReadDir(filepath.Join(dir, "deps", "lib", "perl5"))
	for _, e := range entries {
		if e.IsDir() && strings.Contains(e.Name(), "-") { // archname dirs
			inc = append(inc, filepath.Join(dir, "deps", "lib", "perl5", e.Name()))
		}
	}
	if r, err := p.Eval(ctx, `sub __t_xinc { unshift @INC, @_; 1 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("inc helper: err=%v error=%v", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_xinc", inc...); err != nil {
		t.Fatalf("add inc: %v", err)
	}

	mustEval := func(what, src, want string) {
		t.Helper()
		r, err := p.Eval(ctx, src)
		if err != nil || r.Error != nil {
			t.Fatalf("%s: err=%v ok=%v error=%v", what, err, (r.Error == nil), r.Error)
		}
		if want != "" && r.Value.String() != want {
			t.Fatalf("%s = %q, want %q", what, r.Value.String(), want)
		}
	}

	mustEval("load", `
		$ENV{MOUSE_PUREPERL} = 1;
		$ENV{PERL_DATA_MESSAGEPACK} = 'pp';
		require Text::Xslate;
		Text::Xslate::Engine->can('render_string') ? 'XS' : 'PP';
	`, "XS")

	mustEval("render", `
		my $tx = Text::Xslate->new();
		$tx->render_string(
			'Hello, <: $name :>! <: for $items -> $i { :>[<: $i :>]<: } :> total=<: $items.size() :>',
			{ name => 'go-perl', items => [1, 2, 3] });
	`, "Hello, go-perl! [1][2][3] total=3")

	mustEval("escape", `
		Text::Xslate->new->render_string('<: $v :>', { v => '<b>&"</b>' });
	`, "&lt;b&gt;&amp;&quot;&lt;/b&gt;")

	mustEval("function", `
		Text::Xslate->new(function => { double => sub { $_[0] * 2 } })
			->render_string('<: double(21) :>', {});
	`, "42")

	// The range operator compiles to the scratch-OP pp_flop path.
	mustEval("range", `
		Text::Xslate->new->render_string('<: for [ 1 .. 4 ] -> $i { :>(<: $i :>)<: } :>', {});
	`, "(1)(2)(3)(4)")

	mustEval("macro", `
		my $out = Text::Xslate->new->render_string(
			join("\n", ': macro add1 -> $x {', ': $x + 1;', ': }', '<: add1(41) :>'), {});
		$out =~ s/\s+//g;
		$out;
	`, "42")

	mustEval("methods", `
		Text::Xslate->new->render_string('<: $items.join(",") :>|<: $h.keys().join("-") :>',
			{ items => [ 5, 6, 7 ], h => { b => 1, a => 2, c => 3 } });
	`, "5,6,7|a-b-c")

	r, err := p.Eval(ctx, `
		my $tx = Text::Xslate->new(verbose => 2);
		my $out = eval { $tx->render_string('<: no_such_fn() :>', {}) };
		$@ ? "died: $@" : "no death: $out";
	`)
	if err != nil || r.Error != nil {
		t.Fatalf("error path: err=%v error=%v", err, r.Error)
	}
	if !strings.HasPrefix(fmt.Sprint(r.Value.String()), "died: Undefined symbol") {
		t.Fatalf("error path = %q, want a died: Undefined symbol message", r.Value.String())
	}

	mustEval("two engines", `
		my $a = Text::Xslate->new();
		my $b = Text::Xslate->new();
		$a->render_string('<: $v :>', { v => 'A' }) . $b->render_string('<: $v :>', { v => 'B' });
	`, "AB")
}
