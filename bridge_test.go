package perl_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

func TestCallNamedSub(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub add { my ($a, $b) = @_; $a + $b } 1;`); err != nil || !r.Ok {
		t.Fatalf("define add: err=%v error=%q", err, r.Error)
	}
	got, err := p.Call(ctx, "add", 40, 2)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(got) != 1 || got[0] != float64(42) {
		t.Fatalf("add(40,2) = %#v, want [42]", got)
	}
}

func TestCallListReturnAndStructuredArgs(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `
		sub shape {
			my ($items, $opts) = @_;
			return (scalar(@$items), $opts->{name}, [map { $_ * 2 } @$items]);
		} 1;`)
	if err != nil || !r.Ok {
		t.Fatalf("define shape: err=%v error=%q", err, r.Error)
	}
	got, err := p.Call(ctx, "shape", []any{1, 2, 3}, map[string]any{"name": "perl"})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("shape returned %#v, want 3 values", got)
	}
	if got[0] != float64(3) || got[1] != "perl" {
		t.Fatalf("shape scalars = %#v", got[:2])
	}
	doubled, ok := got[2].([]any)
	if !ok || len(doubled) != 3 || doubled[2] != float64(6) {
		t.Fatalf("shape list = %#v, want [2 4 6]", got[2])
	}
}

func TestCallDieIsPerlError(t *testing.T) {
	p := newPerl(t)
	_, err := p.Call(context.Background(), "no_such_sub_anywhere")
	if err == nil {
		t.Fatalf("expected error calling undefined sub")
	}
	var pe *perl.PerlError
	if !errors.As(err, &pe) {
		t.Fatalf("err = %T (%v), want *perl.PerlError", err, err)
	}
	if !strings.Contains(pe.Message, "Undefined subroutine") {
		t.Fatalf("message = %q, want it to mention the undefined subroutine", pe.Message)
	}
}

func TestBindGoFunction(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	err := p.Bind("go_upper", func(args []any) ([]any, error) {
		s, _ := args[0].(string)
		return []any{strings.ToUpper(s)}, nil
	})
	if err != nil {
		t.Fatalf("Bind: %v", err)
	}
	r, err := p.Eval(ctx, `go_upper("hello")`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if !r.Ok || r.Result != "HELLO" {
		t.Fatalf("go_upper = ok=%v result=%q error=%q", r.Ok, r.Result, r.Error)
	}
}

func TestBindErrorBecomesDie(t *testing.T) {
	p := newPerl(t)
	if err := p.Bind("go_fail", func(args []any) ([]any, error) {
		return nil, errors.New("intentional failure")
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	r, err := p.Eval(context.Background(), `eval { go_fail() }; $@`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if !r.Ok || !strings.Contains(r.Result, "intentional failure") {
		t.Fatalf("captured $@ = %q (ok=%v error=%q)", r.Result, r.Ok, r.Error)
	}
}

// TestBindReentrant drives the full round trip: Go calls Perl, the Perl sub
// calls back into a bound Go function, and that Go function evaluates Perl
// again on the same instance (the invoke lock is released around callback
// handlers, so the nested entry is legal).
func TestBindReentrant(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if err := p.Bind("go_nested", func(args []any) ([]any, error) {
		r, err := p.Eval(ctx, `10 + 5`)
		if err != nil {
			return nil, err
		}
		if !r.Ok {
			return nil, errors.New(r.Error)
		}
		return []any{r.Result}, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	if r, err := p.Eval(ctx, `sub relay { go_nested() . "-relayed" } 1;`); err != nil || !r.Ok {
		t.Fatalf("define relay: err=%v error=%q", err, r.Error)
	}
	got, err := p.Call(ctx, "relay")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(got) != 1 || got[0] != "15-relayed" {
		t.Fatalf("relay = %#v, want [15-relayed]", got)
	}
}

func TestBindRejectsInvalidName(t *testing.T) {
	p := newPerl(t)
	err := p.Bind(`x; system("rm")`, func(args []any) ([]any, error) { return nil, nil })
	if err == nil {
		t.Fatalf("expected invalid sub name to be rejected")
	}
}
