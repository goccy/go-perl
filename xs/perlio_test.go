//go:build darwin || linux

package xs_test

import (
	"context"
	perl "github.com/goccy/go-perl"
	"os"
	"path/filepath"
	"testing"
)

// TestPerlIOUtf8Strict runs PerlIO::utf8_strict — a dist whose whole point
// is a custom PerlIO layer — through the native SDK's layer bridge: the
// guest pushes a PerlIOBuf-derived proxy under the dist's layer name and
// only the slots the dist customized (Pushed and Fill here) round-trip to
// the host, where they run against a shadow instance whose tail forwards
// PerlIO_read/eof/get_ptr/... back to the guest layer below the proxy.
//
// It needs a prepared directory (opt-in via GOPERL_UTF8STRICT_DIR) holding:
//
//	PerlIO-utf8_strict.so — `gperl xs build` output for the dist
//	lib/                  — the dist's pure-Perl half (blib/lib)
func TestPerlIOUtf8Strict(t *testing.T) {
	dir := os.Getenv("GOPERL_UTF8STRICT_DIR")
	if dir == "" {
		t.Skip("GOPERL_UTF8STRICT_DIR not set; skipping the PerlIO::utf8_strict acceptance test")
	}
	so := filepath.Join(dir, "PerlIO-utf8_strict.so")
	if _, err := os.Stat(so); err != nil {
		t.Fatalf("GOPERL_UTF8STRICT_DIR is set but %s is missing: %v", so, err)
	}

	tmp := t.TempDir()
	valid := filepath.Join(tmp, "valid.txt")
	// Latin-1 supplement, hiragana, and an astral camel: 1-, 2-, 3- and
	// 4-byte sequences, some spanning readline calls.
	if err := os.WriteFile(valid,
		[]byte("abc\xc3\xa9def\n\xe3\x81\x82\xe3\x81\x84\n\xf0\x9f\x90\xaaend\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(tmp, "bad.txt")
	if err := os.WriteFile(bad, []byte("ok\xc0\xafrest"), 0o644); err != nil {
		t.Fatal(err)
	}

	p := newHostPerl(t)
	ctx := context.Background()

	loadXSModule(t, p, "PerlIO::utf8_strict", so)

	if r, err := p.Eval(ctx, `sub __t_pinc { unshift @INC, @_; 1 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("inc helper: err=%v error=%v", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_pinc", perl.NewValue(filepath.Join(dir, "lib"))); err != nil {
		t.Fatalf("add inc: %v", err)
	}
	if r, err := p.Eval(ctx,
		`sub __t_pfiles { our ($VALID, $BAD) = @_; 1 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("file helper: err=%v error=%v", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_pfiles", perl.NewValue(valid), perl.NewValue(bad)); err != nil {
		t.Fatalf("pass files: %v", err)
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

	mustEval("load", `require PerlIO::utf8_strict; 'loaded'`, "loaded")

	// Valid UTF-8 through readline: decoded characters with the UTF-8
	// flag on, across all sequence widths.
	mustEval("read valid", `
		open my $r, '<:utf8_strict', $main::VALID or die $!;
		my @lines = <$r>;
		close $r;
		join ',',
			scalar @lines,
			$lines[0] eq "abc\x{e9}def\n" ? 'latin' : 'L!',
			$lines[1] eq "\x{3042}\x{3044}\n" ? 'kana' : 'K!',
			$lines[2] eq "\x{1f42a}end\n" ? 'astral' : 'A!',
			utf8::is_utf8($lines[1]) ? 'flagged' : 'F!';
	`, "3,latin,kana,astral,flagged")

	// A slurp drives the block-read path rather than fast_gets.
	mustEval("slurp", `
		open my $r, '<:utf8_strict', $main::VALID or die $!;
		local $/;
		my $all = <$r>;
		close $r;
		length $all;
	`, "16")

	// Malformed input (an overlong encoding) must croak with the dist's
	// message, raised from the host Fill hook through the layer bridge.
	mustEval("malformed", `
		open my $r, '<:utf8_strict', $main::BAD or die $!;
		my $line = eval { scalar <$r> };
		my $err = $@ // '';
		close $r;
		!defined($line) && $err =~ /Can't decode ill-formed UTF-8 octet sequence/
			? 'croaked' : "unexpected: $err";
	`, "croaked")

	// Pushing onto an already-open handle exercises binmode-time Pushed.
	mustEval("binmode push", `
		open my $r, '<:raw', $main::VALID or die $!;
		binmode($r, ':utf8_strict') or die 'binmode failed';
		my $line = <$r>;
		close $r;
		$line eq "abc\x{e9}def\n" ? 'pushed' : 'wrong';
	`, "pushed")
}
