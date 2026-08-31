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

// scalarOf extracts a ScalarValue or fails the test.
func scalarOf(t *testing.T, v perl.Value) perl.ScalarValue {
	t.Helper()
	s, err := perl.As[perl.ScalarValue](v)
	if err != nil {
		t.Fatalf("As[ScalarValue]: %v", err)
	}
	return s
}

// refOf extracts a RefValue or fails the test.
func refOf(t *testing.T, v perl.Value) perl.RefValue {
	t.Helper()
	r, err := perl.As[perl.RefValue](v)
	if err != nil {
		t.Fatalf("As[RefValue]: %v", err)
	}
	return r
}

// derefAs dereferences a RefValue into the expected concrete type.
func derefAs[T perl.PerlValue](t *testing.T, ctx context.Context, v perl.Value) T {
	t.Helper()
	inner, err := refOf(t, v).Deref(ctx)
	if err != nil {
		t.Fatalf("Deref: %v", err)
	}
	out, err := perl.As[T](inner)
	if err != nil {
		t.Fatalf("As after Deref: %v", err)
	}
	return out
}

func TestCallNamedSub(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub add { my ($a, $b) = @_; $a + $b } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define add: err=%v error=%v", err, r.Error)
	}
	got, err := p.Call(ctx, "add", perl.ValueOf(40), perl.ValueOf(2))
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(got) != 1 || scalarOf(t, got[0]).Int() != 42 {
		t.Fatalf("add(40,2) = %#v, want [42]", got)
	}
}

func TestCallListReturnAndAggregateArgs(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `
		sub shape {
			my ($items, $opts) = @_;
			return (scalar(@$items), $opts->{name}, [map { $_ * 2 } @$items]);
		} 1;`)
	if err != nil || r.Error != nil {
		t.Fatalf("define shape: err=%v error=%v", err, r.Error)
	}
	items, err := p.NewArray(ctx, perl.ValueOf(1), perl.ValueOf(2), perl.ValueOf(3))
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	opts, err := p.NewHash(ctx, perl.Pair{K: "name", V: perl.ValueOf("perl")})
	if err != nil {
		t.Fatalf("NewHash: %v", err)
	}
	got, err := p.Call(ctx, "shape", items.Ref(), opts.Ref())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("shape returned %#v, want 3 values", got)
	}
	if scalarOf(t, got[0]).Int() != 3 || scalarOf(t, got[1]).String() != "perl" {
		t.Fatalf("shape scalars = %#v", got[:2])
	}
	// The returned arrayref crosses as a handle; the value API walks it.
	doubled := derefAs[perl.ArrayValue](t, ctx, got[2])
	vals, err := doubled.Values(ctx)
	if err != nil {
		t.Fatalf("Values: %v", err)
	}
	if len(vals) != 3 || scalarOf(t, vals[2]).Int() != 6 {
		t.Fatalf("shape list = %#v, want [2 4 6]", vals)
	}
}

// TestArgumentFlattening: an ArrayValue or HashValue in an argument list
// flattens into its contents — Perl's own calling convention for f(@a, %h).
func TestArgumentFlattening(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub count_args { scalar @_ } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define count_args: err=%v error=%v", err, r.Error)
	}
	arr, err := p.NewArray(ctx, perl.ValueOf(1), perl.ValueOf(2), perl.ValueOf(3))
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	// Flattened: 3 elements + 1 scalar = 4 arguments.
	got, err := p.Call(ctx, "count_args", arr, perl.ValueOf("x"))
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if scalarOf(t, got[0]).Int() != 4 {
		t.Fatalf("flattened arg count = %v, want 4", got[0])
	}
	// By reference: 1 arrayref + 1 scalar = 2 arguments.
	got, err = p.Call(ctx, "count_args", arr.Ref(), perl.ValueOf("x"))
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if scalarOf(t, got[0]).Int() != 2 {
		t.Fatalf("by-ref arg count = %v, want 2", got[0])
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
	if err != nil || r.Error != nil {
		t.Fatalf("define Counter: err=%v error=%v", err, r.Error)
	}

	got, err := p.Call(ctx, "get_counter")
	if err != nil {
		t.Fatalf("Call get_counter: %v", err)
	}
	obj := refOf(t, got[0])
	if class, blessed := obj.Class(); !blessed || class != "Counter" {
		t.Fatalf("Class = %q/%v, want Counter", class, blessed)
	}

	// Mutate through a Go-side method call; Perl must observe it.
	if _, err := obj.MethodCall(ctx, "inc"); err != nil {
		t.Fatalf("MethodCall inc: %v", err)
	}
	if r, err := p.Eval(ctx, `$counter->{n}`); err != nil || r.Error != nil || resultStr(r) != "1" {
		t.Fatalf("Perl-side n after Go inc = %q (err=%v error=%v)", resultStr(r), err, r.Error)
	}

	// The same object surfaces with the same handle (dedup by refaddr).
	again, err := p.Call(ctx, "get_counter")
	if err != nil {
		t.Fatalf("Call get_counter again: %v", err)
	}
	if !obj.Equal(refOf(t, again[0])) {
		t.Fatalf("two crossings of the same object are not Equal")
	}

	// Passing the handle back dereferences to the same SV.
	same, err := p.Call(ctx, "same_as_ours", obj)
	if err != nil {
		t.Fatalf("Call same_as_ours: %v", err)
	}
	if scalarOf(t, same[0]).String() != "same" {
		t.Fatalf("identity through an argument round trip = %#v, want same", same[0])
	}

	// Isa/Can consult Perl's own method resolution.
	if ok, err := obj.Isa(ctx, "Counter"); err != nil || !ok {
		t.Fatalf("Isa(Counter) = %v/%v, want true", ok, err)
	}
	if ok, err := obj.Can(ctx, "inc"); err != nil || !ok {
		t.Fatalf("Can(inc) = %v/%v, want true", ok, err)
	}
	if ok, err := obj.Can(ctx, "no_such_method"); err != nil || ok {
		t.Fatalf("Can(no_such_method) = %v/%v, want false", ok, err)
	}
}

// TestCodeValueCall drives a Perl closure from Go and checks the captured
// state advances — closures only work if the handle is the same code ref.
func TestCodeValueCall(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub make_adder { my ($n) = @_; my $sum = $n; sub { $sum += $_[0]; $sum } } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define make_adder: err=%v error=%v", err, r.Error)
	}
	res, err := p.Call(ctx, "make_adder", perl.ValueOf(10))
	if err != nil {
		t.Fatalf("Call make_adder: %v", err)
	}
	adder := derefAs[perl.CodeValue](t, ctx, res[0])
	if out, err := adder.Call(ctx, perl.ValueOf(5)); err != nil || scalarOf(t, out[0]).Int() != 15 {
		t.Fatalf("first Call = %#v err=%v, want 15", out, err)
	}
	if out, err := adder.CallScalar(ctx, perl.ValueOf(7)); err != nil || scalarOf(t, out).Int() != 22 {
		t.Fatalf("CallScalar = %#v err=%v, want 22 (closure state must persist)", out, err)
	}
}

// TestArrayAndHashOperations covers the AV/HV views end to end.
func TestArrayAndHashOperations(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	r, err := p.Eval(ctx, `our @list = (10, "twenty", 30.5); \@list`)
	if err != nil || r.Error != nil {
		t.Fatalf("build list: err=%v error=%v", err, r.Error)
	}
	arr := derefAs[perl.ArrayValue](t, ctx, r.Value)

	if n, err := arr.Len(ctx); err != nil || n != 3 {
		t.Fatalf("Len = %d/%v, want 3", n, err)
	}
	if v, err := arr.Index(ctx, 1); err != nil || scalarOf(t, v).String() != "twenty" {
		t.Fatalf("Index(1) = %#v/%v, want twenty", v, err)
	}
	if v, err := arr.Index(ctx, -1); err != nil || scalarOf(t, v).Float() != 30.5 {
		t.Fatalf("Index(-1) = %#v/%v, want 30.5", v, err)
	}
	if err := arr.SetIndex(ctx, 0, perl.ValueOf(11)); err != nil {
		t.Fatalf("SetIndex: %v", err)
	}
	if err := arr.Push(ctx, perl.ValueOf("tail")); err != nil {
		t.Fatalf("Push: %v", err)
	}
	// The Perl side observes every mutation (same array, not a copy).
	if r, err := p.Eval(ctx, `join ",", @list`); err != nil || r.Error != nil || resultStr(r) != "11,twenty,30.5,tail" {
		t.Fatalf("Perl view = %q (err=%v error=%v)", resultStr(r), err, r.Error)
	}

	r, err = p.Eval(ctx, `our %conf = (host => "example", port => 8080); \%conf`)
	if err != nil || r.Error != nil {
		t.Fatalf("build hash: err=%v error=%v", err, r.Error)
	}
	hash := derefAs[perl.HashValue](t, ctx, r.Value)

	if v, ok, err := hash.Get(ctx, "host"); err != nil || !ok || scalarOf(t, v).String() != "example" {
		t.Fatalf("Get(host) = %#v/%v/%v", v, ok, err)
	}
	if _, ok, err := hash.Get(ctx, "absent"); err != nil || ok {
		t.Fatalf("Get(absent) exists=%v err=%v, want false", ok, err)
	}
	if err := hash.Set(ctx, "scheme", perl.ValueOf("https")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := hash.Delete(ctx, "port"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	keys, err := hash.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("Keys = %v, want host+scheme", keys)
	}
	if r, err := p.Eval(ctx, `join ",", map { "$_=$conf{$_}" } sort keys %conf`); err != nil || r.Error != nil ||
		resultStr(r) != "host=example,scheme=https" {
		t.Fatalf("Perl view = %q (err=%v error=%v)", resultStr(r), err, r.Error)
	}
}

// TestByteStringRoundTrip: raw bytes — NUL bytes and non-UTF-8 sequences
// included — cross both directions unchanged. This is the reason Bytes()
// exists next to String().
func TestByteStringRoundTrip(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub mirror { $_[0] } sub bytelen { length $_[0] } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define mirror: err=%v error=%v", err, r.Error)
	}
	raw := []byte{0x00, 0xFF, 0xFE, 'a', 0x00, 0x80, 'z'}
	got, err := p.Call(ctx, "mirror", perl.ValueOf(raw))
	if err != nil {
		t.Fatalf("Call mirror: %v", err)
	}
	if string(scalarOf(t, got[0]).Bytes()) != string(raw) {
		t.Fatalf("mirror = %x, want %x", scalarOf(t, got[0]).Bytes(), raw)
	}
	n, err := p.Call(ctx, "bytelen", perl.ValueOf(raw))
	if err != nil {
		t.Fatalf("Call bytelen: %v", err)
	}
	if scalarOf(t, n[0]).Int() != int64(len(raw)) {
		t.Fatalf("guest length = %v, want %d", n[0], len(raw))
	}
}

// TestScalarKindsAcrossBridge: each scalar kind survives a round trip with
// its kind intact (nothing is stringified in transit).
func TestScalarKindsAcrossBridge(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub mirror2 { $_[0] } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define mirror2: err=%v error=%v", err, r.Error)
	}
	cases := []struct {
		in   perl.ScalarValue
		kind perl.Kind
	}{
		{perl.Undef(), perl.KindUndef},
		{perl.ValueOf(true), perl.KindBool},
		{perl.ValueOf(41), perl.KindInt},
		{perl.ValueOf(2.5), perl.KindFloat},
		{perl.ValueOf("text"), perl.KindString},
	}
	for _, c := range cases {
		got, err := p.Call(ctx, "mirror2", c.in)
		if err != nil {
			t.Fatalf("mirror2(%v): %v", c.in, err)
		}
		if got[0].Kind() != c.kind {
			t.Fatalf("kind of %v after round trip = %s, want %s", c.in, got[0].Kind(), c.kind)
		}
	}
	// The eval result path reports kinds too.
	r, err := p.Eval(ctx, `42`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval 42: err=%v error=%v", err, r.Error)
	}
	if r.Value.Kind() != perl.KindInt || scalarOf(t, r.Value).Int() != 42 {
		t.Fatalf("eval 42 = kind %s value %v", r.Value.Kind(), r.Value)
	}
}

// TestBindReceivesRefs: a reference argument to a bound Go function arrives
// as a handle to the same object, is usable during the call, and — since the
// handler owns it like any other value — simply keeping it keeps the Perl
// object alive beyond the call.
func TestBindReceivesRefs(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	var kept perl.RefValue
	if err := p.Bind("go_take", func(args []perl.Value) ([]perl.Value, error) {
		ref, err := perl.As[perl.RefValue](args[0])
		if err != nil {
			return nil, errors.New("expected a reference argument")
		}
		kept = ref
		class, _ := ref.Class()
		return []perl.Value{perl.ValueOf(class)}, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	r, err := p.Eval(ctx, `package Box; sub new { bless {v=>"inside"}, shift } package main; go_take(Box->new)`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: err=%v error=%v", err, r.Error)
	}
	if resultStr(r) != "Box" {
		t.Fatalf("bound fn saw class %q, want Box", resultStr(r))
	}
	// The retained handle outlives the call that delivered it.
	inner, err := kept.Deref(ctx)
	if err != nil {
		t.Fatalf("Deref retained ref: %v", err)
	}
	hv, err := perl.As[perl.HashValue](inner)
	if err != nil {
		t.Fatalf("retained ref is %s, want a hash", inner.Kind())
	}
	if v, ok, err := hv.Get(ctx, "v"); err != nil || !ok || scalarOf(t, v).String() != "inside" {
		t.Fatalf("retained ref field = %#v/%v/%v, want inside", v, ok, err)
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
	err := p.Bind("go_upper", func(args []perl.Value) ([]perl.Value, error) {
		s, err := perl.As[perl.ScalarValue](args[0])
		if err != nil {
			return nil, err
		}
		return []perl.Value{perl.ValueOf(strings.ToUpper(s.String()))}, nil
	})
	if err != nil {
		t.Fatalf("Bind: %v", err)
	}
	r, err := p.Eval(ctx, `go_upper("hello")`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil || resultStr(r) != "HELLO" {
		t.Fatalf("go_upper = ok=%v result=%q error=%v", (r.Error == nil), resultStr(r), r.Error)
	}
}

func TestBindErrorBecomesDie(t *testing.T) {
	p := newPerl(t)
	if err := p.Bind("go_fail", func(args []perl.Value) ([]perl.Value, error) {
		return nil, errors.New("intentional failure")
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	r, err := p.Eval(context.Background(), `eval { go_fail() }; $@`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil || !strings.Contains(resultStr(r), "intentional failure") {
		t.Fatalf("captured $@ = %q (ok=%v error=%v)", resultStr(r), (r.Error == nil), r.Error)
	}
}

// TestBindReentrant drives the full round trip: Go calls Perl, the Perl sub
// calls back into a bound Go function, and that Go function evaluates Perl
// again on the same instance (the invoke lock is released around callback
// handlers, so the nested entry is legal).
func TestBindReentrant(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if err := p.Bind("go_nested", func(args []perl.Value) ([]perl.Value, error) {
		r, err := p.Eval(ctx, `10 + 5`)
		if err != nil {
			return nil, err
		}
		if r.Error != nil {
			return nil, r.Error
		}
		return []perl.Value{r.Value}, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	if r, err := p.Eval(ctx, `sub relay { go_nested() . "-relayed" } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define relay: err=%v error=%v", err, r.Error)
	}
	got, err := p.Call(ctx, "relay")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(got) != 1 || scalarOf(t, got[0]).String() != "15-relayed" {
		t.Fatalf("relay = %#v, want [15-relayed]", got)
	}
}

// TestBindClassFromGo defines a Perl class whose methods are Go functions:
// Perl constructs instances, calls methods (instance and class invocants),
// and subclasses it with plain @ISA inheritance.
func TestBindClassFromGo(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	err := p.BindClass("Shout", map[string]perl.GoFunc{
		"upper": func(args []perl.Value) ([]perl.Value, error) {
			self, err := perl.As[perl.RefValue](args[0])
			if err != nil {
				return nil, errors.New("upper is an instance method")
			}
			inner, err := self.Deref(ctx)
			if err != nil {
				return nil, err
			}
			hv, err := perl.As[perl.HashValue](inner)
			if err != nil {
				return nil, err
			}
			word, _, err := hv.Get(ctx, "word")
			if err != nil {
				return nil, err
			}
			sv, err := perl.As[perl.ScalarValue](word)
			if err != nil {
				return nil, err
			}
			return []perl.Value{perl.ValueOf(strings.ToUpper(sv.String()))}, nil
		},
		"who": func(args []perl.Value) ([]perl.Value, error) {
			switch inv := args[0].(type) {
			case perl.RefValue:
				class, _ := inv.Class()
				return []perl.Value{perl.ValueOf("instance:" + class)}, nil
			case perl.ScalarValue:
				return []perl.Value{perl.ValueOf("class:" + inv.String())}, nil
			default:
				return nil, fmt.Errorf("unexpected invocant kind %s", args[0].Kind())
			}
		},
	})
	if err != nil {
		t.Fatalf("BindClass: %v", err)
	}

	r, err := p.Eval(ctx, `Shout->new(word => "quiet")->upper`)
	if err != nil || r.Error != nil || resultStr(r) != "QUIET" {
		t.Fatalf("instance method = %q (err=%v error=%v)", resultStr(r), err, r.Error)
	}
	r, err = p.Eval(ctx, `Shout->who`)
	if err != nil || r.Error != nil || resultStr(r) != "class:Shout" {
		t.Fatalf("class invocant = %q (err=%v error=%v)", resultStr(r), err, r.Error)
	}
	// Perl-side subclassing of the Go-implemented class just works: method
	// resolution is Perl's own.
	r, err = p.Eval(ctx, `package Louder; our @ISA = ('Shout'); package main; Louder->new(word => "sub")->upper`)
	if err != nil || r.Error != nil || resultStr(r) != "SUB" {
		t.Fatalf("inherited method = %q (err=%v error=%v)", resultStr(r), err, r.Error)
	}
	r, err = p.Eval(ctx, `Louder->new->who`)
	if err != nil || r.Error != nil || resultStr(r) != "instance:Louder" {
		t.Fatalf("subclass invocant = %q (err=%v error=%v)", resultStr(r), err, r.Error)
	}
}

// TestHandleKeepsPerlValueAlive is the GC-direction guarantee: while Go
// holds a handle-bearing Value, the registry pin holds a real Perl reference
// count, so dropping every Perl-side reference must NOT free the object
// under the host; dropping the Go value hands it back to Perl's refcounting
// (the finalizer queues the pin, the next call drains it, DESTROY runs).
func TestHandleKeepsPerlValueAlive(t *testing.T) {
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
	if err != nil || r.Error != nil {
		t.Fatalf("define Tracked: err=%v error=%v", err, r.Error)
	}

	// Scope the handle so it can be collected below.
	func() {
		got, err := p.Call(ctx, "take_tracked")
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		obj := refOf(t, got[0])

		// Drop the only Perl-side reference. The registry pin must keep the
		// object alive: DESTROY has not run and methods still work.
		if r, err := p.Eval(ctx, `undef $main::t; $Tracked::destroyed`); err != nil || r.Error != nil || resultStr(r) != "0" {
			t.Fatalf("destroyed after undef = %q (err=%v error=%v), want 0", resultStr(r), err, r.Error)
		}
		if out, err := obj.MethodCall(ctx, "ping"); err != nil || scalarOf(t, out[0]).String() != "pong" {
			t.Fatalf("method on host-retained object = %#v err=%v, want pong", out, err)
		}
	}()

	// The handle is unreachable now: the finalizer queues the pin, the next
	// call drains it, and Perl's own refcounting resumes — DESTROY runs.
	// Loop because finalizers run asynchronously after GC.
	destroyed := func() string {
		r, err := p.Eval(ctx, `$Tracked::destroyed`)
		if err != nil || r.Error != nil {
			t.Fatalf("probe: err=%v error=%v", err, r.Error)
		}
		return resultStr(r)
	}
	for attempt := 0; attempt < 50; attempt++ {
		runtime.GC()
		if destroyed() == "1" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("DESTROY did not run after the Go handle was collected")
}

// TestUseAfterClose: every entry point fails cleanly on a closed instance.
func TestUseAfterClose(t *testing.T) {
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	// Take a real handle, then close.
	if r, err := p.Eval(ctx, `sub give_ref { {} } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define give_ref: err=%v error=%v", err, r.Error)
	}
	vals, err := p.Call(ctx, "give_ref")
	if err != nil {
		t.Fatalf("Call give_ref: %v", err)
	}
	ref := refOf(t, vals[0])

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
	if _, err := ref.MethodCall(ctx, "anything"); err == nil {
		t.Fatalf("value operation after Close should fail")
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
	if err := p.Bind("go_wait", func(args []perl.Value) ([]perl.Value, error) {
		close(inHandler)
		// Block until the OTHER goroutine has completed a full Eval on this
		// instance. If the lock were still held here, this would deadlock.
		if err := <-otherDone; err != nil {
			return nil, err
		}
		return []perl.Value{perl.ValueOf("released")}, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	go func() {
		<-inHandler
		r, err := p.Eval(ctx, `21 * 2`)
		if err == nil && (r.Error != nil || resultStr(r) != "42") {
			err = fmt.Errorf("nested eval ok=%v result=%q error=%v", (r.Error == nil), resultStr(r), r.Error)
		}
		otherDone <- err
	}()
	r, err := p.Eval(ctx, `go_wait()`)
	if err != nil || r.Error != nil || resultStr(r) != "released" {
		t.Fatalf("go_wait = ok=%v result=%q error=%v err=%v", (r.Error == nil), resultStr(r), r.Error, err)
	}
}

func TestBindRejectsInvalidName(t *testing.T) {
	p := newPerl(t)
	err := p.Bind(`x; system("rm")`, func(args []perl.Value) ([]perl.Value, error) { return nil, nil })
	if err == nil {
		t.Fatalf("expected invalid sub name to be rejected")
	}
}

// TestHashArgumentFlattening: a HashValue in an argument list flattens to
// its key/value pairs, like f(%h).
func TestHashArgumentFlattening(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub kv { my %h = @_; join ",", map { "$_=$h{$_}" } sort keys %h } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define kv: err=%v error=%v", err, r.Error)
	}
	h, err := p.NewHash(ctx,
		perl.Pair{K: "a", V: perl.ValueOf(1)},
		perl.Pair{K: "b", V: perl.ValueOf(2)},
	)
	if err != nil {
		t.Fatalf("NewHash: %v", err)
	}
	got, err := p.Call(ctx, "kv", h)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if scalarOf(t, got[0]).String() != "a=1,b=2" {
		t.Fatalf("flattened hash = %#v, want a=1,b=2", got[0])
	}
}

// TestBindMultiValueReturn: a bound Go function's whole return slice
// becomes the Perl call's return list in list context.
func TestBindMultiValueReturn(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if err := p.Bind("go_three", func(args []perl.Value) ([]perl.Value, error) {
		return []perl.Value{perl.ValueOf("x"), perl.ValueOf(2), perl.ValueOf(true)}, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	r, err := p.Eval(ctx, `my @r = go_three(); scalar(@r) . ":" . join("|", @r)`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: err=%v error=%v", err, r.Error)
	}
	if resultStr(r) != "3:x|2|1" {
		t.Fatalf("list-context return = %q, want 3:x|2|1", resultStr(r))
	}
}

// TestExitDuringCall: a guest exit() inside Call unwinds cleanly and is the
// error ExitCode recognises; the interpreter is still destructible.
func TestExitDuringCall(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub bail { exit 7 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define bail: err=%v error=%v", err, r.Error)
	}
	_, err := p.Call(ctx, "bail")
	if code, ok := perl.ExitCode(err); !ok || code != 7 {
		t.Fatalf("Call bail = %v (code=%d ok=%v), want exit 7", err, code, ok)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close after exit: %v", err)
	}
}

// TestCallContextCancel: cancelling the context stops a runaway sub the
// same way it stops a runaway eval.
func TestCallContextCancel(t *testing.T) {
	p := newPerl(t)
	if r, err := p.Eval(context.Background(), `sub spin { my $n = 0; while (1) { $n++ } } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define spin: err=%v error=%v", err, r.Error)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, err := p.Call(ctx, "spin")
	if err == nil || err != ctx.Err() {
		t.Fatalf("Call spin = %v, want ctx.Err() (%v)", err, ctx.Err())
	}
	// The instance stays usable after a cancelled call.
	if r, err := p.Eval(context.Background(), `41 + 1`); err != nil || r.Error != nil || resultStr(r) != "42" {
		t.Fatalf("post-cancel eval: err=%v error=%v", err, r.Error)
	}
}
