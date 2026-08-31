package perl_test

import (
	"context"
	"testing"

	perl "github.com/goccy/go-perl"
)

// TestScalarCoercions pins the Perl-semantics coercions on ScalarValue.
func TestScalarCoercions(t *testing.T) {
	cases := []struct {
		name string
		v    perl.ScalarValue
		str  string
		i    int64
		f    float64
		b    bool
	}{
		{"undef", perl.Undef(), "", 0, 0, false},
		{"true", perl.ValueOf(true), "1", 1, 1, true},
		{"false", perl.ValueOf(false), "", 0, 0, false},
		{"int", perl.ValueOf(42), "42", 42, 42, true},
		{"zero", perl.ValueOf(0), "0", 0, 0, false},
		{"float", perl.ValueOf(2.5), "2.5", 2, 2.5, true},
		{"string", perl.ValueOf("perl"), "perl", 0, 0, true},
		{"numeric prefix", perl.ValueOf("42abc"), "42abc", 42, 42, true},
		{"exponent", perl.ValueOf("1.5e2xyz"), "1.5e2xyz", 150, 150, true},
		{"zero string", perl.ValueOf("0"), "0", 0, 0, false},
		{"empty string", perl.ValueOf(""), "", 0, 0, false},
		{"bytes", perl.ValueOf([]byte{'4', '2'}), "42", 42, 42, true},
	}
	for _, c := range cases {
		if got := c.v.String(); got != c.str {
			t.Errorf("%s: String() = %q, want %q", c.name, got, c.str)
		}
		if got := c.v.Int(); got != c.i {
			t.Errorf("%s: Int() = %d, want %d", c.name, got, c.i)
		}
		if got := c.v.Float(); got != c.f {
			t.Errorf("%s: Float() = %v, want %v", c.name, got, c.f)
		}
		if got := c.v.Bool(); got != c.b {
			t.Errorf("%s: Bool() = %v, want %v", c.name, got, c.b)
		}
	}
}

// TestScalarBytes: Bytes returns the raw byte string without re-encoding.
func TestScalarBytes(t *testing.T) {
	raw := []byte{0x00, 0xFF, 'x'}
	v := perl.ValueOf(raw)
	if got := v.Bytes(); string(got) != string(raw) {
		t.Fatalf("Bytes = %x, want %x", got, raw)
	}
	if v.Kind() != perl.KindString {
		t.Fatalf("Kind = %s, want string", v.Kind())
	}
}

// TestKindReporting: every concrete type reports its kind, Kind renders a
// name, and the numeric ValueOf variants map to KindInt.
func TestKindReporting(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `[\my @a2, \my %h2, sub { 5 }, *STDOUT{IO}]`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: err=%v error=%v", err, r.Error)
	}
	parts, err := derefAs[perl.ArrayValue](t, ctx, r.Value).Values(ctx)
	if err != nil || len(parts) != 4 {
		t.Fatalf("setup: %#v/%v", parts, err)
	}
	arr := derefAs[perl.ArrayValue](t, ctx, parts[0])
	hash := derefAs[perl.HashValue](t, ctx, parts[1])
	code := derefAs[perl.CodeValue](t, ctx, parts[2])
	io := derefAs[perl.IOValue](t, ctx, parts[3])
	if arr.Kind() != perl.KindArray || hash.Kind() != perl.KindHash ||
		code.Kind() != perl.KindCode || io.Kind() != perl.KindIO {
		t.Fatalf("view kinds = %s/%s/%s/%s", arr.Kind(), hash.Kind(), code.Kind(), io.Kind())
	}
	if io.Ref().Kind() != perl.KindRef || hash.Ref().Kind() != perl.KindRef {
		t.Fatalf("Ref kinds wrong")
	}
	// A code ref rebuilt from the view is the same subroutine.
	if out, err := p.Call(ctx, "__t_kind_probe", code.Ref()); err == nil {
		_ = out
		t.Fatalf("expected undefined sub to die")
	}
	if out, err := code.Call(ctx); err != nil || scalarOf(t, out[0]).Int() != 5 {
		t.Fatalf("code via view = %#v/%v, want 5", out, err)
	}
	if perl.KindUndef.String() != "undef" || perl.Kind(200).String() == "" {
		t.Fatalf("Kind.String rendering broken")
	}
	if perl.ValueOf(int32(7)).Int() != 7 || perl.ValueOf(uint32(8)).Int() != 8 ||
		perl.ValueOf(int64(9)).Int() != 9 {
		t.Fatalf("numeric ValueOf variants broken")
	}
	if string(perl.ValueOf(42).Bytes()) != "42" {
		t.Fatalf("Bytes of a non-string scalar = %q", perl.ValueOf(42).Bytes())
	}
}

// TestAsExtraction: As succeeds on the matching concrete type and errors —
// never panics — on a mismatch.
func TestAsExtraction(t *testing.T) {
	var v perl.Value = perl.ValueOf(7)
	if s, err := perl.As[perl.ScalarValue](v); err != nil || s.Int() != 7 {
		t.Fatalf("As[ScalarValue] = %v/%v", s, err)
	}
	if _, err := perl.As[perl.RefValue](v); err == nil {
		t.Fatalf("As[RefValue] on a scalar should error")
	}
}

// TestValueTypeSwitch: the runtime-inspection path — a type switch over the
// sealed implementations — sees the concrete types.
func TestValueTypeSwitch(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `[1, 2]`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: err=%v error=%v", err, r.Error)
	}
	switch v := r.Value.(type) {
	case perl.RefValue:
		if v.Kind() != perl.KindRef {
			t.Fatalf("RefValue.Kind = %s", v.Kind())
		}
		inner, err := v.Deref(ctx)
		if err != nil {
			t.Fatalf("Deref: %v", err)
		}
		if _, ok := inner.(perl.ArrayValue); !ok {
			t.Fatalf("deref of an arrayref is %T, want ArrayValue", inner)
		}
	default:
		t.Fatalf("eval of an arrayref returned %T, want RefValue", r.Value)
	}
}

// TestScalarRefDeref: Deref on a scalar reference reads the referent; a
// ref-to-ref derefs one level at a time.
func TestScalarRefDeref(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `my $x = "inner"; \\$x`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: err=%v error=%v", err, r.Error)
	}
	outer, err := perl.As[perl.RefValue](r.Value)
	if err != nil {
		t.Fatalf("outer: %v", err)
	}
	mid, err := outer.Deref(ctx)
	if err != nil {
		t.Fatalf("Deref outer: %v", err)
	}
	midRef, err := perl.As[perl.RefValue](mid)
	if err != nil {
		t.Fatalf("mid is %s, want ref", mid.Kind())
	}
	inner, err := midRef.Deref(ctx)
	if err != nil {
		t.Fatalf("Deref mid: %v", err)
	}
	if s, err := perl.As[perl.ScalarValue](inner); err != nil || s.String() != "inner" {
		t.Fatalf("innermost = %#v/%v, want inner", inner, err)
	}
}

// TestUtf8StringRoundTrip: a Go string crosses as a utf8 character string,
// so Perl sees characters (length counts them), and it comes back byte-equal.
func TestUtf8StringRoundTrip(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub charlen { length $_[0] } sub mirror3 { $_[0] } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define: err=%v error=%v", err, r.Error)
	}
	const s = "こんにちは" // 5 characters, 15 UTF-8 bytes
	got, err := p.Call(ctx, "charlen", perl.ValueOf(s))
	if err != nil {
		t.Fatalf("charlen: %v", err)
	}
	if scalarOf(t, got[0]).Int() != 5 {
		t.Fatalf("character length = %v, want 5", got[0])
	}
	back, err := p.Call(ctx, "mirror3", perl.ValueOf(s))
	if err != nil {
		t.Fatalf("mirror3: %v", err)
	}
	if scalarOf(t, back[0]).String() != s {
		t.Fatalf("round trip = %q, want %q", scalarOf(t, back[0]).String(), s)
	}
}

// TestNewArrayNewHashRoundTrip: Go-built aggregates materialise guest-side
// and read back consistently.
func TestNewArrayNewHashRoundTrip(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()

	arr, err := p.NewArray(ctx, perl.ValueOf("a"), perl.ValueOf(2), perl.ValueOf(true))
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	if n, err := arr.Len(ctx); err != nil || n != 3 {
		t.Fatalf("Len = %d/%v, want 3", n, err)
	}
	vals, err := arr.Values(ctx)
	if err != nil || len(vals) != 3 {
		t.Fatalf("Values = %#v/%v", vals, err)
	}
	if scalarOf(t, vals[0]).String() != "a" || scalarOf(t, vals[1]).Int() != 2 || !scalarOf(t, vals[2]).Bool() {
		t.Fatalf("Values = %#v", vals)
	}

	h, err := p.NewHash(ctx,
		perl.Pair{K: "name", V: perl.ValueOf("go-perl")},
		perl.Pair{K: "nested", V: arr.Ref()},
	)
	if err != nil {
		t.Fatalf("NewHash: %v", err)
	}
	if v, ok, err := h.Get(ctx, "name"); err != nil || !ok || scalarOf(t, v).String() != "go-perl" {
		t.Fatalf("Get(name) = %#v/%v/%v", v, ok, err)
	}
	nested, ok, err := h.Get(ctx, "nested")
	if err != nil || !ok {
		t.Fatalf("Get(nested): %v/%v", ok, err)
	}
	nref, err := perl.As[perl.RefValue](nested)
	if err != nil {
		t.Fatalf("nested is %s, want ref", nested.Kind())
	}
	if !nref.Equal(arr.Ref()) {
		t.Fatalf("nested arrayref lost identity")
	}
}

// TestSingleSlotRejectsBareAggregate: an ArrayValue/HashValue cannot fill a
// single-value slot (Perl would want an explicit reference there) — the
// encoder reports it instead of guessing.
func TestSingleSlotRejectsBareAggregate(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	arr, err := p.NewArray(ctx, perl.ValueOf(1))
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	h, err := p.NewHash(ctx)
	if err != nil {
		t.Fatalf("NewHash: %v", err)
	}
	if err := h.Set(ctx, "k", arr); err == nil {
		t.Fatalf("Set with a bare array should error (pass Ref())")
	}
	if err := h.Set(ctx, "k", arr.Ref()); err != nil {
		t.Fatalf("Set with the reference: %v", err)
	}
}

// TestAdoptAcrossClone: a handle obtained in the prototype BEFORE its first
// Clone designates the same (copied) value in every clone.
func TestAdoptAcrossClone(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `my $n = 100; sub { $n += $_[0]; $n }`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: err=%v error=%v", err, r.Error)
	}
	ref, err := perl.As[perl.RefValue](r.Value)
	if err != nil {
		t.Fatalf("As: %v", err)
	}
	inner, err := ref.Deref(ctx)
	if err != nil {
		t.Fatalf("Deref: %v", err)
	}
	code, err := perl.As[perl.CodeValue](inner)
	if err != nil {
		t.Fatalf("code: %v", err)
	}

	c, err := p.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	defer c.Close()
	adopted, err := perl.Adopt(c, code)
	if err != nil {
		t.Fatalf("Adopt: %v", err)
	}
	// The clone's copy advances independently of the prototype's.
	if out, err := adopted.CallScalar(ctx, perl.ValueOf(1)); err != nil || scalarOf(t, out).Int() != 101 {
		t.Fatalf("clone call = %#v/%v, want 101", out, err)
	}
	if out, err := code.CallScalar(ctx, perl.ValueOf(2)); err != nil || scalarOf(t, out).Int() != 102 {
		t.Fatalf("prototype call = %#v/%v, want 102 (independent state)", out, err)
	}
}

// TestAdoptEveryKind: Adopt rebinds each handle-bearing concrete type, and
// passes scalars through unchanged.
func TestAdoptEveryKind(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `our @aa = (1); our %hh = (k => "v"); [\@aa, \%hh, \*STDOUT]`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: err=%v error=%v", err, r.Error)
	}
	parts, err := derefAs[perl.ArrayValue](t, ctx, r.Value).Values(ctx)
	if err != nil || len(parts) != 3 {
		t.Fatalf("setup: %#v/%v", parts, err)
	}
	arr := derefAs[perl.ArrayValue](t, ctx, parts[0])
	hash := derefAs[perl.HashValue](t, ctx, parts[1])
	glob := derefAs[perl.GlobValue](t, ctx, parts[2])
	if glob.Kind() != perl.KindGlob || glob.Ref().Kind() != perl.KindRef {
		t.Fatalf("glob view kind = %s", glob.Kind())
	}

	c, err := p.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	defer c.Close()

	if s, err := perl.Adopt(c, perl.ValueOf("plain")); err != nil || s.String() != "plain" {
		t.Fatalf("Adopt scalar = %#v/%v", s, err)
	}
	ca, err := perl.Adopt(c, arr)
	if err != nil {
		t.Fatalf("Adopt array: %v", err)
	}
	if n, err := ca.Len(ctx); err != nil || n != 1 {
		t.Fatalf("adopted array Len = %d/%v", n, err)
	}
	ch, err := perl.Adopt(c, hash)
	if err != nil {
		t.Fatalf("Adopt hash: %v", err)
	}
	if v, ok, err := ch.Get(ctx, "k"); err != nil || !ok || scalarOf(t, v).String() != "v" {
		t.Fatalf("adopted hash Get = %#v/%v/%v", v, ok, err)
	}
	if _, err := perl.Adopt(c, glob); err != nil {
		t.Fatalf("Adopt glob: %v", err)
	}
	if _, err := perl.Adopt(c, arr.Ref()); err != nil {
		t.Fatalf("Adopt ref: %v", err)
	}
}
