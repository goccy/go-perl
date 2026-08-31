package perl_test

// Ties and magic through the value operations: the bridge's array/hash
// operations must run REAL Perl element accesses — tied FETCH/STORE/DELETE
// and iteration magic fire, and a die inside them surfaces as *PerlError,
// never as a crash or a wrong classification.

import (
	"context"
	"errors"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// tieSetup defines pure-Perl tie classes (no CPAN modules) and returns
// references to a tied array and a tied hash whose operations log
// themselves; the "bomb" variants die inside FETCH.
const tieSetup = `
package LogTie::Array;
sub TIEARRAY  { bless { d => [], log => $_[1] }, $_[0] }
sub FETCH     { my ($s,$i) = @_; push @{$s->{log}}, "fetch:$i"; $s->{d}[$i] }
sub STORE     { my ($s,$i,$v) = @_; push @{$s->{log}}, "store:$i"; $s->{d}[$i] = $v }
sub FETCHSIZE { scalar @{$_[0]{d}} }
sub STORESIZE { }
sub PUSH      { my $s = shift; push @{$s->{log}}, "push"; push @{$s->{d}}, @_ }
sub CLEAR     { $_[0]{d} = [] }
sub EXTEND    { }

package BombTie::Array;
our @ISA = ('LogTie::Array');
sub FETCH { die "tied fetch exploded\n" }

package LogTie::Hash;
sub TIEHASH  { bless { d => {} }, $_[0] }
sub FETCH    { $_[0]{d}{$_[1]} }
sub STORE    { $_[0]{d}{$_[1]} = $_[2] }
sub DELETE   { delete $_[0]{d}{$_[1]} }
sub EXISTS   { exists $_[0]{d}{$_[1]} }
sub FIRSTKEY { my @k = sort keys %{$_[0]{d}}; $k[0] }
sub NEXTKEY  { my ($s, $last) = @_; my @k = sort keys %{$_[0]{d}}; for my $i (0..$#k) { return $k[$i+1] if $k[$i] eq $last } undef }
sub CLEAR    { $_[0]{d} = {} }
package main;
our @tie_log;
tie our @tied, 'LogTie::Array', \@tie_log;
@tied = (10, 20, 30);
@tie_log = (); # keep only what the Go-side operations trigger
tie our @bomb, 'BombTie::Array', [];
$#bomb; # size stays 0; FETCH is the bomb
push @bomb, 1;
tie our %tiedh, 'LogTie::Hash';
%tiedh = (a => 1, b => 2);
[\@tied, \@bomb, \%tiedh];
`

func tieValues(t *testing.T, p *perl.Perl, ctx context.Context) (perl.ArrayValue, perl.ArrayValue, perl.HashValue) {
	t.Helper()
	r, err := p.Eval(ctx, tieSetup)
	if err != nil || r.Error != nil {
		t.Fatalf("tie setup: err=%v error=%v", err, r.Error)
	}
	views := derefAs[perl.ArrayValue](t, ctx, r.Value)
	vals, err := views.Values(ctx)
	if err != nil || len(vals) != 3 {
		t.Fatalf("tie setup values: %#v/%v", vals, err)
	}
	return derefAs[perl.ArrayValue](t, ctx, vals[0]),
		derefAs[perl.ArrayValue](t, ctx, vals[1]),
		derefAs[perl.HashValue](t, ctx, vals[2])
}

// TestTiedArrayOperations: Index/SetIndex/Len/Push on a tied array drive
// the tie's FETCH/STORE/FETCHSIZE/PUSH methods.
func TestTiedArrayOperations(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	tied, _, _ := tieValues(t, p, ctx)

	if n, err := tied.Len(ctx); err != nil || n != 3 {
		t.Fatalf("tied Len = %d/%v, want 3", n, err)
	}
	if v, err := tied.Index(ctx, 1); err != nil || scalarOf(t, v).Int() != 20 {
		t.Fatalf("tied Index(1) = %#v/%v, want 20", v, err)
	}
	if err := tied.SetIndex(ctx, 1, perl.ValueOf(21)); err != nil {
		t.Fatalf("tied SetIndex: %v", err)
	}
	if err := tied.Push(ctx, perl.ValueOf(40)); err != nil {
		t.Fatalf("tied Push: %v", err)
	}
	// The tie's methods observed every operation.
	r, err := p.Eval(ctx, `join ",", $tied[1], $tied[3], "log=" . join("|", grep { /store|push/ } @tie_log)`)
	if err != nil || r.Error != nil {
		t.Fatalf("probe: err=%v error=%v", err, r.Error)
	}
	if got := resultStr(r); got != "21,40,log=store:1|push" {
		t.Fatalf("tie observation = %q", got)
	}
}

// TestTiedFetchDieIsPerlError: a die inside a tied FETCH — triggered from
// Go through Index — comes back as *PerlError, and the instance stays
// usable.
func TestTiedFetchDieIsPerlError(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	_, bomb, _ := tieValues(t, p, ctx)

	_, err := bomb.Index(ctx, 0)
	var pe *perl.PerlError
	if !errors.As(err, &pe) || !strings.Contains(pe.Message, "tied fetch exploded") {
		t.Fatalf("Index on bomb = %v, want *PerlError with the tied message", err)
	}
	// Values drains through FETCH too.
	if _, err := bomb.Values(ctx); !errors.As(err, &pe) {
		t.Fatalf("Values on bomb = %v, want *PerlError", err)
	}
	// The instance survives the die.
	if r, err := p.Eval(ctx, `1 + 1`); err != nil || r.Error != nil || resultStr(r) != "2" {
		t.Fatalf("post-die eval: err=%v error=%v", err, r.Error)
	}
}

// TestTiedHashOperations: Get/Set/Delete/Keys on a tied hash drive the
// tie's methods (EXISTS/FETCH/STORE/DELETE/FIRSTKEY/NEXTKEY).
func TestTiedHashOperations(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	_, _, tiedh := tieValues(t, p, ctx)

	if v, ok, err := tiedh.Get(ctx, "a"); err != nil || !ok || scalarOf(t, v).Int() != 1 {
		t.Fatalf("tied Get(a) = %#v/%v/%v", v, ok, err)
	}
	if _, ok, err := tiedh.Get(ctx, "zz"); err != nil || ok {
		t.Fatalf("tied Get(zz) exists=%v err=%v, want false", ok, err)
	}
	if err := tiedh.Set(ctx, "c", perl.ValueOf(3)); err != nil {
		t.Fatalf("tied Set: %v", err)
	}
	if err := tiedh.Delete(ctx, "a"); err != nil {
		t.Fatalf("tied Delete: %v", err)
	}
	keys, err := tiedh.Keys(ctx)
	if err != nil || len(keys) != 2 {
		t.Fatalf("tied Keys = %v/%v, want [b c]", keys, err)
	}
}

// TestUtf8HashKeys: non-ASCII keys round-trip through Set/Get/Keys.
func TestUtf8HashKeys(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	h, err := p.NewHash(ctx, perl.Pair{K: "キー", V: perl.ValueOf("値")})
	if err != nil {
		t.Fatalf("NewHash: %v", err)
	}
	if v, ok, err := h.Get(ctx, "キー"); err != nil || !ok || scalarOf(t, v).String() != "値" {
		t.Fatalf("Get(utf8) = %#v/%v/%v", v, ok, err)
	}
	if err := h.Set(ctx, "鍵", perl.ValueOf(2)); err != nil {
		t.Fatalf("Set(utf8): %v", err)
	}
	keys, err := h.Keys(ctx)
	if err != nil || len(keys) != 2 {
		t.Fatalf("Keys = %v/%v", keys, err)
	}
	found := map[string]bool{}
	for _, k := range keys {
		found[k] = true
	}
	if !found["キー"] || !found["鍵"] {
		t.Fatalf("utf8 keys = %v", keys)
	}
}

// TestRegexpRefDeref: a qr// reference (refkind REGEXP) dereferences to the
// pattern's scalar form.
func TestRegexpRefDeref(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `\qr/ab+c/i`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: err=%v error=%v", err, r.Error)
	}
	// \qr// is a ref to the (blessed) Regexp object.
	ref := refOf(t, r.Value)
	inner, err := ref.Deref(ctx)
	if err != nil {
		t.Fatalf("Deref: %v", err)
	}
	rref, err := perl.As[perl.RefValue](inner)
	if err != nil {
		t.Fatalf("inner = %s, want the Regexp ref", inner.Kind())
	}
	if class, blessed := rref.Class(); !blessed || class != "Regexp" {
		t.Fatalf("Class = %q/%v, want Regexp", class, blessed)
	}
	pat, err := rref.Deref(ctx)
	if err != nil {
		t.Fatalf("Deref regexp: %v", err)
	}
	if s, err := perl.As[perl.ScalarValue](pat); err != nil || !strings.Contains(s.String(), "ab+c") {
		t.Fatalf("regexp deref = %#v/%v, want the pattern text", pat, err)
	}
}

// TestIsaOnUnblessedRefFails: method dispatch on an unblessed reference is
// a Perl-level error, reported as *PerlError.
func TestIsaOnUnblessedRefFails(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `{}`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: err=%v error=%v", err, r.Error)
	}
	ref := refOf(t, r.Value)
	_, err = ref.Isa(ctx, "HASH")
	var pe *perl.PerlError
	if !errors.As(err, &pe) || !strings.Contains(pe.Message, "unblessed") {
		t.Fatalf("Isa on unblessed = %v, want *PerlError (unblessed reference)", err)
	}
}
