package perl

// The Go <-> Perl function bridge.
//
// A call's argument/return list crosses the boundary as TYPED BINARY nodes
// (see wire.go): scalars by value with their kind, references by handle —
// the guest pins the actual SV in a per-interpreter registry (deduplicated
// by referent address, refcount held) and only the id crosses, surfacing in
// Go as a handle-bearing Value. Passing it back dereferences to THE SAME SV,
// so object identity and aliasing survive any number of round trips.
//
// Go -> Perl: Call invokes a named Perl sub through the perl_call bridge
// export (CodeValue.Call is the by-reference form). Perl -> Go: Bind
// registers a Go function under an id and evals a one-line Perl sub
// forwarding to main::__plwasm_go_call(id, @_) — an XS the guest installed
// at boot that speaks the same node protocol.

import (
	"context"
	"fmt"
	"regexp"
)

// PerlError is a Perl-level die surfaced by Eval, Call, and the value
// operations: Message holds $@.
type PerlError struct {
	Message string
}

func (e *PerlError) Error() string { return e.Message }

// GoFunc is a Go function callable from Perl. args is the Perl call's
// argument list as typed Values (references arrive as handle-bearing
// Values); the returned slice becomes the Perl call's return list.
// Returning an error makes the Perl call die with the error text.
//
// fn may call back into the instance (Eval, Call, value operations) — the
// invoke lock is released for the callback's duration, and a nested entry
// from the handler's own goroutine is safe. What deadlocks is HOST-side
// lock cycles: a handler that blocks on something only the code that
// initiated the outer call can provide (a mutex it holds, a channel it will
// only service after the call returns) waits forever. Keep handlers
// self-contained or hand work to other goroutines without waiting on the
// outer caller.
type GoFunc func(args []Value) ([]Value, error)

// Call invokes the named Perl subroutine in list context and returns its
// return list. name is a fully qualified sub name ("My::App::handler") or a
// main:: sub name ("handler"). An ArrayValue or HashValue among args
// flattens into the argument list, exactly like f(@a, %h) in Perl (pass
// Ref() to pass the reference itself). A Perl-level die (including calling
// a sub that does not exist) is returned as a *PerlError; other errors are
// host/transport failures or a context cancellation, which behaves exactly
// like Eval's.
func (p *Perl) Call(ctx context.Context, name string, args ...Value) ([]Value, error) {
	enc, err := p.encodeArgs(args)
	if err != nil {
		return nil, err
	}
	resp, interrupted, err := p.raw.CallOp(ctx, name, enc)
	if err != nil {
		return nil, err
	}
	return p.decodeListResult(ctx, resp, interrupted)
}

// perlSubName pins what Bind accepts: a Perl identifier, optionally
// package-qualified. The name is interpolated into a generated sub
// definition, so anything else is rejected up front.
var perlSubName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(::[A-Za-z_][A-Za-z0-9_]*)*$`)

// Bind makes fn callable from Perl as the subroutine name (fully qualified,
// or defined into main:: when unqualified). The Perl side sees an ordinary
// sub: arguments arrive in Go as typed Values, the returned slice becomes
// the return list, and a Go error — or a contained panic — becomes a Perl
// die. Binding the same name again replaces the sub (the previous function
// is unreferenced but its id stays allocated).
func (p *Perl) Bind(name string, fn GoFunc) error {
	if fn == nil {
		return fmt.Errorf("nil GoFunc for %q", name)
	}
	if !perlSubName.MatchString(name) {
		return fmt.Errorf("invalid Perl sub name %q", name)
	}
	if err := p.ensureDispatcher(); err != nil {
		return err
	}

	p.funcsMu.Lock()
	p.nextFuncID++
	id := p.nextFuncID
	p.funcs[id] = fn
	p.funcsMu.Unlock()

	r, err := p.Eval(context.Background(),
		fmt.Sprintf("sub %s { main::__plwasm_go_call(%d, @_) } 1;", name, id))
	if err != nil {
		return err
	}
	if r.Error != nil {
		return fmt.Errorf("define %s: %s", name, r.Error)
	}
	return nil
}

// perlMethodName pins what BindClass accepts as a method key: one plain
// identifier (the package half comes from the class name).
var perlMethodName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// BindClass defines a Perl class whose methods are implemented in Go: each
// entry becomes the sub name::method, so Perl's own method resolution —
// inheritance via @ISA, class methods, SUPER:: — applies unchanged. Every
// method receives the invocant as args[0]: the class name (a string scalar)
// for a class-method call, the blessed instance (a RefValue) for an
// instance call.
//
// Unless methods defines "new", a default constructor is provided:
//
//	sub new { my $class = shift; bless { @_ }, $class }
//
// so instances are ordinary Perl hash-based objects (readable from Go
// through the value operations, extensible by subclassing in Perl). A Go
// method that needs per-instance Go state can key a Go-side map by an id
// stored in the hash.
func (p *Perl) BindClass(name string, methods map[string]GoFunc) error {
	if !perlSubName.MatchString(name) {
		return fmt.Errorf("invalid Perl class name %q", name)
	}
	for m := range methods {
		if !perlMethodName.MatchString(m) {
			return fmt.Errorf("invalid method name %q", m)
		}
	}
	if _, hasNew := methods["new"]; !hasNew {
		r, err := p.Eval(context.Background(),
			fmt.Sprintf("sub %s::new { my $class = shift; bless { @_ }, $class } 1;", name))
		if err != nil {
			return err
		}
		if r.Error != nil {
			return fmt.Errorf("define %s::new: %s", name, r.Error)
		}
	}
	for m, fn := range methods {
		if err := p.Bind(name+"::"+m, fn); err != nil {
			return err
		}
	}
	return nil
}

// ensureDispatcher makes sure the instance's guest-side callback dispatcher
// is registered and this wrapper's function table exists.
func (p *Perl) ensureDispatcher() error {
	p.funcsMu.Lock()
	if p.funcs == nil {
		p.funcs = map[int32]GoFunc{}
	}
	p.funcsMu.Unlock()
	return p.raw.EnsureDispatcher()
}

// handleUserCallback serves the Perl->Go function bridge: methodID carries
// the bound function's id, req is the guest-encoded argument node list, and
// the returned bytes are the response envelope. Failures are ALWAYS
// reported in-band (a Go error return would be protobuf-encoded by the
// generated dispatch and misparse guest-side).
func (p *Perl) handleUserCallback(methodID int32, req []byte) ([]byte, error) {
	p.funcsMu.RLock()
	fn, ok := p.funcs[methodID]
	p.funcsMu.RUnlock()
	if !ok {
		return callbackDie(fmt.Errorf("no Go function bound for id %d", methodID)), nil
	}
	r := &nodeReader{b: req}
	count := int(r.u32())
	args := make([]Value, 0, count)
	for i := 0; i < count; i++ {
		v, err := p.decodeNode(r)
		if err != nil {
			return callbackDie(fmt.Errorf("decode arguments: %w", err)), nil
		}
		args = append(args, v)
	}
	results, err := safeCall(fn, args)
	if err != nil {
		return callbackDie(err), nil
	}
	enc, err := p.encodeArgs(results)
	if err != nil {
		return callbackDie(fmt.Errorf("encode Go results: %w", err)), nil
	}
	return append([]byte{wireOK}, enc...), nil
}

// callbackDie encodes a die envelope for the Perl->Go response direction.
func callbackDie(err error) []byte {
	msg := err.Error()
	b := make([]byte, 0, 5+len(msg))
	b = append(b, wireDie)
	b = appendU32(b, uint32(len(msg)))
	return append(b, msg...)
}

// safeCall contains a panicking GoFunc so a guest-triggered call cannot take
// the whole process down; the panic surfaces as a Perl die.
func safeCall(fn GoFunc, args []Value) (results []Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			results, err = nil, fmt.Errorf("Go function panicked: %v", r)
		}
	}()
	return fn(args)
}
