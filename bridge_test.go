package perl_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

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
	// The returned arrayref crosses as a handle; Export materialises data.
	ref, ok := got[2].(*perl.Ref)
	if !ok || ref.Reftype() != "ARRAY" {
		t.Fatalf("shape list = %#v, want an ARRAY *perl.Ref", got[2])
	}
	defer ref.Free()
	exported, err := ref.Export(ctx)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	doubled, ok := exported.([]any)
	if !ok || len(doubled) != 3 || doubled[2] != float64(6) {
		t.Fatalf("shape list = %#v, want [2 4 6]", exported)
	}
}

// TestObjectIdentityRoundTrip is the pointer-semantics guarantee: a blessed
// Perl object crossing to Go and back must be THE SAME object — method calls
// from Go mutate the object Perl sees, the same object always surfaces with
// an Equal handle, and identity survives being passed back as an argument.
func TestObjectIdentityRoundTrip(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `
		package Counter;
		sub new { my ($c) = @_; bless { n => 0 }, $c }
		sub inc { my ($s) = @_; $s->{n}++; $s }
		sub n   { my ($s) = @_; $s->{n} }
		package main;
		our $counter = Counter->new;
		sub get_counter { $counter }
		sub same_as_ours { my ($x) = @_; $x == $counter ? "same" : "different" }
		1;`)
	if err != nil || !r.Ok {
		t.Fatalf("define Counter: err=%v error=%q", err, r.Error)
	}

	got, err := p.Call(ctx, "get_counter")
	if err != nil {
		t.Fatalf("Call get_counter: %v", err)
	}
	obj, ok := got[0].(*perl.Ref)
	if !ok {
		t.Fatalf("get_counter returned %#v, want *perl.Ref", got[0])
	}
	defer obj.Free()
	if obj.Class() != "Counter" || obj.Reftype() != "HASH" {
		t.Fatalf("Class=%q Reftype=%q, want Counter/HASH", obj.Class(), obj.Reftype())
	}

	// Mutate through a Go-side method call; Perl must observe it.
	if _, err := obj.MethodCall(ctx, "inc"); err != nil {
		t.Fatalf("MethodCall inc: %v", err)
	}
	if r, err := p.Eval(ctx, `$counter->{n}`); err != nil || !r.Ok || r.Result != "1" {
		t.Fatalf("Perl-side n after Go inc = %q (err=%v error=%q)", r.Result, err, r.Error)
	}

	// The same object surfaces with the same handle (dedup by refaddr).
	again, err := p.Call(ctx, "get_counter")
	if err != nil {
		t.Fatalf("Call get_counter again: %v", err)
	}
	obj2 := again[0].(*perl.Ref)
	defer obj2.Free()
	if !obj.Equal(obj2) {
		t.Fatalf("two crossings of the same object are not Equal")
	}

	// Passing the handle back dereferences to the same SV.
	same, err := p.Call(ctx, "same_as_ours", obj)
	if err != nil {
		t.Fatalf("Call same_as_ours: %v", err)
	}
	if same[0] != "same" {
		t.Fatalf("identity through an argument round trip = %#v, want same", same[0])
	}
}

// TestCodeRefInvoke drives a Perl closure from Go and checks the captured
// state advances — closures only work if the handle is the same code ref.
func TestCodeRefInvoke(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub make_adder { my ($n) = @_; my $sum = $n; sub { $sum += $_[0]; $sum } } 1;`); err != nil || !r.Ok {
		t.Fatalf("define make_adder: err=%v error=%q", err, r.Error)
	}
	res, err := p.Call(ctx, "make_adder", 10)
	if err != nil {
		t.Fatalf("Call make_adder: %v", err)
	}
	adder := res[0].(*perl.Ref)
	defer adder.Free()
	if adder.Reftype() != "CODE" {
		t.Fatalf("Reftype = %q, want CODE", adder.Reftype())
	}
	if out, err := adder.Invoke(ctx, 5); err != nil || out[0] != float64(15) {
		t.Fatalf("first Invoke = %#v err=%v, want 15", out, err)
	}
	if out, err := adder.Invoke(ctx, 7); err != nil || out[0] != float64(22) {
		t.Fatalf("second Invoke = %#v err=%v, want 22 (closure state must persist)", out, err)
	}
}

// TestBindReceivesRefs: a reference argument to a bound Go function arrives
// as a *Ref to the same object, is usable during the call, and — since the
// handler owns it like any other Ref — simply keeping it keeps the Perl
// object alive beyond the call.
func TestBindReceivesRefs(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	var kept *perl.Ref
	if err := p.Bind("go_take", func(args []any) ([]any, error) {
		ref, ok := args[0].(*perl.Ref)
		if !ok {
			return nil, errors.New("expected a reference argument")
		}
		kept = ref
		return []any{ref.Class()}, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	r, err := p.Eval(ctx, `package Box; sub new { bless {v=>"inside"}, shift } package main; go_take(Box->new)`)
	if err != nil || !r.Ok {
		t.Fatalf("Eval: err=%v error=%q", err, r.Error)
	}
	if r.Result != "Box" {
		t.Fatalf("bound fn saw class %q, want Box", r.Result)
	}
	if kept == nil {
		t.Fatalf("handler did not retain the ref")
	}
	defer kept.Free()
	// The retained handle outlives the call that delivered it.
	out, err := kept.Export(ctx)
	if err != nil {
		t.Fatalf("Export retained ref: %v", err)
	}
	if m, ok := out.(map[string]any); !ok || m["v"] != "inside" {
		t.Fatalf("retained ref exported %#v, want {v: inside}", out)
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

// TestGoFuncValueAsCallback passes Go function VALUES to Perl: they arrive
// as ordinary code refs Perl can call immediately, store, and call later.
func TestGoFuncValueAsCallback(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `
		sub apply     { my ($f, $x) = @_; $f->($x) + 1 }
		sub keep_cb   { $main::kept_cb = $_[0]; 1 }
		sub call_kept { $main::kept_cb->(@_) }
		1;`)
	if err != nil || !r.Ok {
		t.Fatalf("define helpers: err=%v error=%q", err, r.Error)
	}

	double := func(args []any) ([]any, error) {
		return []any{args[0].(float64) * 2}, nil
	}
	got, err := p.Call(ctx, "apply", double, 20)
	if err != nil {
		t.Fatalf("Call apply: %v", err)
	}
	if got[0] != float64(41) {
		t.Fatalf("apply(double, 20) = %#v, want 41", got[0])
	}

	// Perl keeps the callback and calls it in a later, unrelated call.
	joiner := func(args []any) ([]any, error) {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = fmt.Sprint(a)
		}
		return []any{strings.Join(parts, "-")}, nil
	}
	if _, err := p.Call(ctx, "keep_cb", joiner); err != nil {
		t.Fatalf("Call keep_cb: %v", err)
	}
	out, err := p.Call(ctx, "call_kept", "a", "b", "c")
	if err != nil {
		t.Fatalf("Call call_kept: %v", err)
	}
	if out[0] != "a-b-c" {
		t.Fatalf("stored Go callback returned %#v, want a-b-c", out[0])
	}
}

// TestBindClassFromGo defines a Perl class whose methods are Go functions:
// Perl constructs instances, calls methods (instance and class invocants),
// and subclasses it with plain @ISA inheritance.
func TestBindClassFromGo(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	err := p.BindClass("Shout", map[string]perl.GoFunc{
		"upper": func(args []any) ([]any, error) {
			self, ok := args[0].(*perl.Ref)
			if !ok {
				return nil, errors.New("upper is an instance method")
			}
			data, err := self.Export(ctx)
			if err != nil {
				return nil, err
			}
			word, _ := data.(map[string]any)["word"].(string)
			return []any{strings.ToUpper(word)}, nil
		},
		"who": func(args []any) ([]any, error) {
			switch inv := args[0].(type) {
			case *perl.Ref:
				return []any{"instance:" + inv.Class()}, nil
			case string:
				return []any{"class:" + inv}, nil
			default:
				return nil, fmt.Errorf("unexpected invocant %T", args[0])
			}
		},
	})
	if err != nil {
		t.Fatalf("BindClass: %v", err)
	}

	r, err := p.Eval(ctx, `Shout->new(word => "quiet")->upper`)
	if err != nil || !r.Ok || r.Result != "QUIET" {
		t.Fatalf("instance method = %q (err=%v error=%q)", r.Result, err, r.Error)
	}
	r, err = p.Eval(ctx, `Shout->who`)
	if err != nil || !r.Ok || r.Result != "class:Shout" {
		t.Fatalf("class invocant = %q (err=%v error=%q)", r.Result, err, r.Error)
	}
	// Perl-side subclassing of the Go-implemented class just works: method
	// resolution is Perl's own.
	r, err = p.Eval(ctx, `package Louder; our @ISA = ('Shout'); package main; Louder->new(word => "sub")->upper`)
	if err != nil || !r.Ok || r.Result != "SUB" {
		t.Fatalf("inherited method = %q (err=%v error=%q)", r.Result, err, r.Error)
	}
	r, err = p.Eval(ctx, `Louder->new->who`)
	if err != nil || !r.Ok || r.Result != "instance:Louder" {
		t.Fatalf("subclass invocant = %q (err=%v error=%q)", r.Result, err, r.Error)
	}
}

// TestRefKeepsPerlValueAlive is the GC-direction guarantee: while Go holds a
// *Ref, the registry pin holds a real Perl reference count, so dropping every
// Perl-side reference must NOT free the object under the host.
func TestRefKeepsPerlValueAlive(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `
		package Tracked;
		our $destroyed = 0;
		sub new { bless { alive => 1 }, shift }
		sub ping { "pong" }
		sub DESTROY { $destroyed++ }
		package main;
		our $t = Tracked->new;
		sub take_tracked { $t }
		1;`)
	if err != nil || !r.Ok {
		t.Fatalf("define Tracked: err=%v error=%q", err, r.Error)
	}
	got, err := p.Call(ctx, "take_tracked")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	obj := got[0].(*perl.Ref)

	// Drop the only Perl-side reference. The registry pin must keep the
	// object alive: DESTROY has not run and methods still work.
	if r, err := p.Eval(ctx, `undef $main::t; $Tracked::destroyed`); err != nil || !r.Ok || r.Result != "0" {
		t.Fatalf("destroyed after undef = %q (err=%v error=%q), want 0", r.Result, err, r.Error)
	}
	if out, err := obj.MethodCall(ctx, "ping"); err != nil || out[0] != "pong" {
		t.Fatalf("method on host-retained object = %#v err=%v, want pong", out, err)
	}

	// Releasing the last pin hands the object back to Perl's refcounting:
	// DESTROY runs.
	if err := obj.Free(); err != nil {
		t.Fatalf("Free: %v", err)
	}
	if r, err := p.Eval(ctx, `$Tracked::destroyed`); err != nil || !r.Ok || r.Result != "1" {
		t.Fatalf("destroyed after Free = %q (err=%v error=%q), want 1", r.Result, err, r.Error)
	}
	// The freed Ref refuses to cross again.
	if _, err := p.Call(ctx, "take_tracked", obj); err == nil {
		t.Fatalf("expected a freed Ref to be rejected as an argument")
	}
}

// TestRefFinalizerReleasesPins: *Ref wrappers dropped without Free are
// released by the garbage collector — the finalizer queues the pin and the
// next call drains it, so the guest registry does not grow without bound.
func TestRefFinalizerReleasesPins(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub fresh_ref { { n => $_[0] } } 1;`); err != nil || !r.Ok {
		t.Fatalf("define fresh_ref: err=%v error=%q", err, r.Error)
	}
	// __plwasm_release_all with no ids just reports the live-handle count.
	liveHandles := func() int {
		got, err := p.Call(ctx, "__plwasm_release_all")
		if err != nil {
			t.Fatalf("live-handle probe: %v", err)
		}
		return int(got[0].(float64))
	}

	base := liveHandles()
	for i := 0; i < 50; i++ {
		if _, err := p.Call(ctx, "fresh_ref", i); err != nil {
			t.Fatalf("Call fresh_ref: %v", err)
		}
	}
	// The 50 result Refs are unreachable now. Finalizers queue their pins;
	// the drain at the next call releases them. Loop because finalizers run
	// asynchronously after GC.
	for attempt := 0; attempt < 50; attempt++ {
		runtime.GC()
		if liveHandles() <= base {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("registry still holds %d handles (baseline %d) after GC", liveHandles(), base)
}

// TestUseAfterClose: every entry point fails cleanly on a closed instance,
// and a straggling Free is a no-op instead of a call into unmapped memory.
func TestUseAfterClose(t *testing.T) {
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	// Take a real Ref, then close.
	if r, err := p.Eval(ctx, `sub give_ref { {} } 1;`); err != nil || !r.Ok {
		t.Fatalf("define give_ref: err=%v error=%q", err, r.Error)
	}
	vals, err := p.Call(ctx, "give_ref")
	if err != nil {
		t.Fatalf("Call give_ref: %v", err)
	}
	ref := vals[0].(*perl.Ref)

	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if _, err := p.Eval(ctx, `1`); err == nil {
		t.Fatalf("Eval after Close should fail")
	}
	if _, err := p.Call(ctx, "give_ref"); err == nil {
		t.Fatalf("Call after Close should fail")
	}
	if err := ref.Free(); err != nil {
		t.Fatalf("Free after Close should be a no-op, got %v", err)
	}
}

// TestCallbackDoesNotDeadlockOtherGoroutines pins the locking contract: the
// instance lock is released while a bound Go function runs, so another
// goroutine's Eval proceeds as a nested top-level entry instead of
// deadlocking against the in-flight call that triggered the callback.
func TestCallbackDoesNotDeadlockOtherGoroutines(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	inHandler := make(chan struct{})
	otherDone := make(chan error, 1)
	if err := p.Bind("go_wait", func(args []any) ([]any, error) {
		close(inHandler)
		// Block until the OTHER goroutine has completed a full Eval on this
		// instance. If the lock were still held here, this would deadlock.
		if err := <-otherDone; err != nil {
			return nil, err
		}
		return []any{"released"}, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	go func() {
		<-inHandler
		r, err := p.Eval(ctx, `21 * 2`)
		if err == nil && (!r.Ok || r.Result != "42") {
			err = fmt.Errorf("nested eval ok=%v result=%q error=%q", r.Ok, r.Result, r.Error)
		}
		otherDone <- err
	}()
	r, err := p.Eval(ctx, `go_wait()`)
	if err != nil || !r.Ok || r.Result != "released" {
		t.Fatalf("go_wait = ok=%v result=%q error=%q err=%v", r.Ok, r.Result, r.Error, err)
	}
}

func TestBindRejectsInvalidName(t *testing.T) {
	p := newPerl(t)
	err := p.Bind(`x; system("rm")`, func(args []any) ([]any, error) { return nil, nil })
	if err == nil {
		t.Fatalf("expected invalid sub name to be rejected")
	}
}
