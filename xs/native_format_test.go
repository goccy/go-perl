//go:build darwin || linux

package xs_test

import (
	"context"
	"testing"
)

// The text perl's formatting functions make is in the encoding of its
// arguments: an SV argument that is UTF-8 makes the whole text UTF-8, the
// bytes around it becoming the characters they were, and the text is as
// long as its parts. warn and croak hand the interpreter that text.
func TestNativeFormattedTextKeepsItsEncoding(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `sub shape { my $s = shift; (utf8::is_utf8($s) ? "utf8" : "bytes") . ":" . join(".", map { sprintf "%x", ord } split //, $s) } 1`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"newSVpvf of a UTF-8 string", `shape(PerlAPI::formatted("new", "\x{263a}"))`, "utf8:3c.263a.3e"},
		{"newSVpvf of a byte string", `shape(PerlAPI::formatted("new", "\xe9"))`, "bytes:3c.e9.3e"},
		{"sv_setpvf", `shape(PerlAPI::formatted("set", "\x{263a}"))`, "utf8:37.3a.263a"},
		{"sv_catpvf of UTF-8 onto bytes", `shape(PerlAPI::formatted("cat", "\x{263a}", "\xe9"))`, "utf8:e9.2b.263a"},
		{"sv_catpvf of bytes onto UTF-8", `shape(PerlAPI::formatted("cat", "\xe9", "\x{263a}"))`, "utf8:263a.2b.e9"},
		{"bytes of the format beside a UTF-8 argument", `shape(PerlAPI::formatted("mixed", "\x{263a}"))`, "utf8:e9.ea.7c.263a.7c.e8"},
		{"bytes of the format beside a byte argument", `shape(PerlAPI::formatted("mixed", "z"))`, "bytes:e9.ea.7c.7a.7c.e8"},
		{"an argument longer than any buffer", `length(PerlAPI::formatted("long", 100000))`, "100002"},
		{"warn", `my $w; local $SIG{__WARN__} = sub { $w = shift }; PerlAPI::formatted("warn", "\x{263a}"); $w eq "warned: \x{263a}\n" ? "same" : shape($w)`, "same"},
		{"croak", `eval { PerlAPI::formatted("croak", "\x{263a}\n") }; $@ eq "\x{263a}\n" ? "same" : shape($@)`, "same"},
		{"croak with a long message", `eval { PerlAPI::formatted("croak", ("y" x 5000) . "\n") }; length $@`, "5001"},
		{"croak names the place of the caller", `eval { PerlAPI::formatted("croak", "no newline") }; $@ =~ /^no newline at .* line \d+\.$/ ? "placed" : $@`, "placed"},
	})
}
