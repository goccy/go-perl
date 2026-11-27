package perl_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	perl "github.com/goccy/go-perl"
)

// collectNow runs the garbage collector until the finalizers of what has
// become unreachable have run.
func collectNow() {
	for i := 0; i < 3; i++ {
		runtime.GC()
		time.Sleep(5 * time.Millisecond)
	}
}

// collectingContext collects garbage whenever it is asked whether it is
// done: an operation asks before it sends anything to the guest, so the
// collection happens while the operation is on its way.
type collectingContext struct{ context.Context }

func (c collectingContext) Err() error {
	collectNow()
	return c.Context.Err()
}

// callOnFreshValues makes a code reference and an array, keeps no Go value
// of either, and calls the code with a reference to the array: once the call
// is on its way, the call itself is all that holds their handles.
//
//go:noinline
func callOnFreshValues(ctx context.Context, p *perl.Perl, i int, how string) (string, error) {
	r, err := p.Eval(ctx, `sub { main::collect(); my ($list) = @_; join ",", ref($list), @$list }`)
	if err != nil {
		return "", err
	}
	if r.Error != nil {
		return "", r.Error
	}
	ref, err := perl.As[perl.RefValue](r.Value)
	if err != nil {
		return "", err
	}
	code, err := ref.Deref(ctx)
	if err != nil {
		return "", err
	}
	cv, err := perl.As[perl.CodeValue](code)
	if err != nil {
		return "", err
	}
	list, err := p.NewArray(ctx, perl.ValueOf(i), perl.ValueOf("x"))
	if err != nil {
		return "", err
	}
	var res []perl.Value
	switch how {
	case "on its way":
		res, err = cv.Call(collectingContext{ctx}, list.Ref())
	default:
		res, err = cv.Call(ctx, list.Ref())
	}
	if err != nil {
		return "", err
	}
	if len(res) != 1 {
		return "", fmt.Errorf("%d results", len(res))
	}
	s, err := perl.As[perl.ScalarValue](res[0])
	if err != nil {
		return "", err
	}
	return s.String(), nil
}

// The handles an operation sends — the code it calls and the references
// among its arguments — stay registered until the operation has ended, even
// when the operation is the last thing in Go that holds them: garbage
// collected while the operation is on its way to the guest, or while it
// runs (the called code runs a Go function that collects and then makes
// another operation, which sends the releases of what was found
// unreachable), must not release a handle the operation names.
func TestHandlesSentByAnOperationStayPinnedUntilItEnds(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	collectInCallback := false
	if err := p.Bind("main::collect", func(args []perl.Value) ([]perl.Value, error) {
		if !collectInCallback {
			return nil, nil
		}
		collectNow()
		r, err := p.Eval(ctx, "1")
		if err != nil {
			return nil, err
		}
		if r.Error != nil {
			return nil, r.Error
		}
		return nil, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	for _, how := range []string{"on its way", "while it runs"} {
		collectInCallback = how == "while it runs"
		for i := 0; i < 5; i++ {
			got, err := callOnFreshValues(ctx, p, i, how)
			if want := fmt.Sprintf("ARRAY,%d,x", i); err != nil || got != want {
				t.Fatalf("collected %s: call %d = (%q, %v), want (%q, nil)", how, i, got, err, want)
			}
		}
	}
}
