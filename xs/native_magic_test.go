//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// A string constructor given no string (a NULL pointer) makes an undefined
// SV: newSVpvn(NULL, 0) is how XS code returns undef.
func TestNativeNullStringIsUndef(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"newSVpvn, newSVpv, newSVpvn_flags without a string; an empty string", `join ",", map { $_ ? "defined" : "undef" } PerlAPI::null_strings()`, "undef,undef,undef,defined"},
	})
}

// `local` on a variable with magic of a native module gives the new value
// the magic, as perl does: set magic sees the new value when the scope
// starts, every assignment in it, and the old value coming back; and the
// magic of the variable itself is still there afterwards.
func TestNativeMagicFollowsLocal(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"set magic through a local", `our $v = "a"; PerlAPI::watch($v); $v = "b"; { local $v = "c"; $v = "d" } $v = "e"; PerlAPI::watched()`, "|b|undef|c|d|b|e"},
		{"a local without a value", `our $w = "a"; PerlAPI::watch($w); { local $w; $w = "in" } $w = "out"; PerlAPI::watched()`, "|undef|in|a|out"},
	})
}

// sv_magicext keeps the length it is given: with HEf_SVKEY the name is an
// SV the magic holds a reference to, and gives up with the variable.
func TestNativeMagicNamedByAnSV(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the length and the reference", `our $key = "slot"; our @holder = ("h"); join ",", PerlAPI::keyed_attach($holder[0], $key)`, "1,-2"},
		{"the key read back", `PerlAPI::keyed_read($holder[0])`, "slot"},
		{"the reference given up", `my $before = Internals::SvREFCNT($key); @holder = (); $before - Internals::SvREFCNT($key)`, "1"},
	})
}

// SvTIED_mg says whether a container is tied and SvTIED_obj to what, so a
// module that treats tied containers apart (to count the keys of a tied
// hash by iterating) does.
func TestNativeTieMagic(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a plain hash", `my %h = (a => 1); PerlAPI::tied_to(\%h)`, "not tied"},
		{"a tied hash", `require Tie::Hash; tie my %h, "Tie::StdHash"; PerlAPI::tied_to(\%h)`, "Tie::StdHash"},
		{"a tied array", `require Tie::Array; tie my @a, "Tie::StdArray"; PerlAPI::tied_to(\@a)`, "Tie::StdArray"},
	})
}

// An overlong UTF-8 sequence is malformed unless the caller allows it.
func TestNativeOverlongUTF8(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a well-formed character", `PerlAPI::decode_first("\xc3\xa9", 0)`, "e9:2"},
		{"overlong, strict", `PerlAPI::decode_first("\xc0\xaf", 0)`, "malformed"},
		{"overlong, allowed", `PerlAPI::decode_first("\xc0\xaf", 1)`, "fffd:2"},
		{"overlong, with its value", `PerlAPI::decode_first("\xe0\x80\xaf", 2)`, "2f:3"},
	})
}

// PerlIO_binmode and PerlIO_apply_layers put layers on a handle of the
// program, and a handle operation that fails leaves its reason in errno
// for the native code that called it.
func TestNativeHandleLayersAndErrors(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "text.txt")
	if err := os.WriteFile(file, []byte("caf\xc3\xa9\r\nline\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `our $file = '`+file+`'; 1`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"PerlIO_binmode with a layer", `open my $fh, "<", $file or die $!; my $r = PerlAPI::pio_layers($fh, ":utf8", 0); my $l = <$fh>; "$r:" . length($l)`, "1:6"},
		{"PerlIO_apply_layers", `open my $fh, "<", $file or die $!; my $r = PerlAPI::pio_layers($fh, ":crlf", 1); my $l = <$fh>; "$r:" . length($l)`, "0:6"},
		{"an unknown layer", `open my $fh, "<", $file or die $!; no warnings; PerlAPI::pio_layers($fh, ":no_such_layer", 1)`, "-1"},
		// The seek fails with what the program's own seek on a pipe fails with.
		{"errno after an open and a seek that fail", `pipe(my $r, my $w) or die $!; my @f = PerlAPI::pio_failures("$file.missing", $r); seek($r, 0, 0); my $own = $!{ESPIPE} ? "ESPIPE" : $!{EINVAL} ? "EINVAL" : "$!"; join ",", @f[0 .. 2], ($f[3] eq $own ? "as the program's" : "$f[3] not $own")`, "1,ENOENT,1,as the program's"},
		{"errno left alone by an open and a seek that work", `open my $fh, "<", $file or die $!; join ",", PerlAPI::pio_failures($file, $fh)`, "0,none,0,none"},
	})
}
