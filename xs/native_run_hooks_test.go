//go:build darwin || linux

package xs_test

import (
	"os"
	"path/filepath"
	"testing"
)

// A source filter of native code reads the file being compiled through the
// filters after it and changes the text; it finds itself by the function
// its data has and takes itself off at the end.
func TestNativeSourceFilter(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	dir := t.TempDir()
	src := "package FilterProbe;\nBEGIN { PerlAPI::filter_double() }\nour $v = 21;\nour $w = \"x21y\";\n1;\n"
	if err := os.WriteFile(filepath.Join(dir, "FilterProbe.pm"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	runEvalCases(t, p, []evalCase{
		{"filtered file", `unshift @INC, "` + dir + `"; require FilterProbe; "$FilterProbe::v $FilterProbe::w"`, "42 x42y"},
		{"later files unfiltered", `21`, "21"},
	})
}

// A keyword plugin reads the parser's fields (the word in PL_tokenbuf) and
// discards the rest of the buffer with lex_unstuff.
func TestNativeParserFields(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the word in the token buffer", `PerlAPI::install_kw_parser(); eval q{ goperl_tok }`, "1"},
		{"the rest of the buffer discarded", `PerlAPI::install_kw_parser(); my $r = eval q{ goperl_drop die "not reached" }; ($r // "undef") . "|$@"`, "7|"},
	})
}

// PL_runops given a run loop of native code: the code compiled and run
// after it runs op by op through that loop, with the same results; the
// loop it replaced comes back.
func TestNativeRunLoop(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"installed", `PerlAPI::runloop(1)`, "0"},
		{"code run through it", `my $s = 0; $s += $_ for 1 .. 10; $s`, "55"},
		{"ops counted", `PerlAPI::runloop(0) > 10 ? "counted" : "not counted"`, "counted"},
		{"the interpreter's own loop again", `my $s = 0; $s += $_ for 1 .. 3; $s`, "6"},
	})
}

// A block hook of native code is called for every block the compiler
// starts after it was registered.
func TestNativeBlockHook(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"registered", `PerlAPI::block_starts()`, "0"},
		{"blocks started", `my $n = PerlAPI::block_starts(); eval q{ { 1 } { 2 } 1 }; PerlAPI::block_starts() - $n >= 2 ? "called" : "not called"`, "called"},
	})
}

// An infix operator of native code (wrap_infix_plugin) is parsed with its
// precedence and built by its build_op; its parse runs after the operator,
// and what it leaves in the operator's data reaches build_op.
func TestNativeInfixOperators(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a named operator", `PerlAPI::enable_infix(); eval q{ join ",", ("a" equ "a") ? 1 : 0, ("a" equ "b") ? 1 : 0 } // $@`, "1,0"},
		{"a symbolic operator with a parse", `PerlAPI::enable_infix(); eval q{ join ",", 3 <~> 5, 5 <~> 3, 2 + 2 <~> 4 } // $@`, "-1,1,0"},
		{"other operators", `PerlAPI::enable_infix(); eval q{ my $x = 2; $x <=> 3 } // $@`, "-1"},
	})
}

// The scanning functions perl gives an extension that defines PERL_EXT
// (B::Hooks::Parser): skipspace_flags, scan_word and scan_str read the
// lexer's buffer from a place in it, and scan_str leaves the string in
// PL_lex_stuff.
func TestNativeLexerScanning(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a word and a string", `PerlAPI::install_kw_scan(); eval q{ goperl_scan Foo::bar "some text" } // $@`, "Foo::bar|some text"},
		{"the rest of the statement", `PerlAPI::install_kw_scan(); eval q{ my $v = goperl_scan w 'x y'; "$v!" } // $@`, "w|x y!"},
	})
}
