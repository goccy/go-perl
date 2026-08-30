package perl_test

import (
	"context"
	"testing"

	perl "github.com/goccy/go-perl"
)

// TestValueCoercions pins the Perl coercion semantics of the Value
// accessors over Eval results.
func TestValueCoercions(t *testing.T) {
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	ctx := context.Background()

	eval := func(src string) perl.Value {
		t.Helper()
		r, err := p.Eval(ctx, src)
		if err != nil {
			t.Fatalf("Eval %q: %v", src, err)
		}
		if r.Error != nil {
			t.Fatalf("Eval %q died: %v", src, r.Error)
		}
		return r.Value
	}

	if v := eval(`21 * 2`); v.Int() != 42 || v.String() != "42" || !v.Bool() {
		t.Fatalf("42: Int=%d String=%q Bool=%v", v.Int(), v.String(), v.Bool())
	}
	if v := eval(`3.5 + 0.25`); v.Float() != 3.75 {
		t.Fatalf("3.75: Float=%v", v.Float())
	}
	if v := eval(`"  -12abc"`); v.Int() != -12 {
		t.Fatalf("perl numify: Int=%d", v.Int())
	}
	if v := eval(`"1.5e2xyz"`); v.Float() != 150 {
		t.Fatalf("perl numify exp: Float=%v", v.Float())
	}
	// Perl truth: "0" and "" are false; "0.0" and "00" are true.
	if v := eval(`"0"`); v.Bool() {
		t.Fatal(`"0" must be false`)
	}
	if v := eval(`""`); v.Bool() {
		t.Fatal(`"" must be false`)
	}
	if v := eval(`"0.0"`); !v.Bool() {
		t.Fatal(`"0.0" must be true`)
	}
	if v := eval(`"abc"`); v.Float() != 0 || v.Bool() != true {
		t.Fatalf("abc: Float=%v Bool=%v", v.Float(), v.Bool())
	}

	// A die surfaces as Result.Error, not as Eval's error.
	r, err := p.Eval(ctx, `die "nope\n"`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error == nil || r.Error.Error() != "nope\n" {
		t.Fatalf("die: %v", r.Error)
	}
}
