package perl

// The Go <-> Perl function bridge.
//
// A call's argument/return list crosses the boundary as a list of TAGGED
// value nodes (JSON is only the carrier; the tag is the semantics):
//
//   - plain scalars (numbers, strings, booleans, undef) cross BY VALUE —
//     Perl scalars are value-semantic anyway (assignment copies them), so
//     nothing is lost;
//   - Go composites ([]any, map[string]any, structs) cross BY VALUE and
//     materialise as fresh Perl structures each time — they carry data, not
//     identity;
//   - Perl REFERENCES — blessed objects, array/hash/code refs — are NEVER
//     serialised. They cross BY HANDLE: the guest pins the actual SV in a
//     per-interpreter registry (deduplicated by refaddr, refcount held) and
//     only the id crosses, surfacing in Go as a *Ref. Passing the *Ref back
//     dereferences to THE SAME SV, so object identity and aliasing survive
//     any number of round trips.
//
// Go -> Perl: Call invokes a named Perl sub through the perl_call bridge
// export. Perl -> Go: Bind registers a Go function under an id and evals a
// one-line Perl sub forwarding to main::__plwasm_go_call(id, @_) — glue the
// guest installed at boot (see perl-wasm's perl.cc for the codec, registry,
// and handle operations).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"sync/atomic"

	"github.com/goccy/go-perl/internal"
)

// PerlError is a Perl-level die surfaced by Call and the Ref operations:
// Message holds $@.
type PerlError struct {
	Message string
}

func (e *PerlError) Error() string { return e.Message }

// GoFunc is a Go function callable from Perl. args is the Perl call's
// argument list: plain scalars arrive as string/float64/bool/nil, references
// arrive as *Ref. The returned slice becomes the Perl call's return list
// (same conversions in reverse; a returned *Ref hands the same Perl
// reference back). Returning an error makes the Perl call die with the error
// text.
//
// Argument *Ref values are owned like any other Ref (keep them as long as
// needed; Free or the garbage collector releases them). fn may call back
// into the instance (Eval, Call, Ref methods) — the invoke lock is released
// for the callback's duration, and a nested entry from the handler's own
// goroutine is safe. What deadlocks is HOST-side lock cycles: a handler that
// blocks on something only the code that initiated the outer call can
// provide (a mutex it holds, a channel it will only service after the call
// returns) waits forever. Keep handlers self-contained or hand work to
// other goroutines without waiting on the outer caller.
type GoFunc func(args []any) ([]any, error)

// Ref is a handle to a Perl reference — a blessed object, an array/hash/code
// ref — held alive in the interpreter's registry. It preserves identity: the
// same Perl reference always surfaces with the same handle id, and sending a
// Ref back to Perl dereferences to the same SV.
//
// Lifecycle is aligned in both directions. While a Ref is reachable in Go,
// its registry pin holds a real Perl reference count, so Perl cannot garbage
// collect the value out from under the host. When the Ref becomes
// unreachable in Go, a finalizer queues the pin for release — the release
// itself runs inside the guest on the next Eval/Call, never on the finalizer
// goroutine (which could otherwise block behind a running Eval or touch a
// closed instance) — and once the last pin drops, Perl's own refcounting
// resumes. Call Free for prompt, deterministic release.
type Ref struct {
	p       *Perl
	id      uint64
	class   string
	reftype string
	// released flips on Free (or when the instance closes underneath); a
	// released Ref no longer owns its pin and refuses to cross the bridge.
	released atomic.Bool
}

// newRef wraps a decoded handle. Every wire crossing of an "r" node carries
// one guest pin the wrapper owns; the finalizer forwards that ownership to
// the release queue if the program never calls Free.
func newRef(p *Perl, id uint64, class, reftype string) *Ref {
	r := &Ref{p: p, id: id, class: class, reftype: reftype}
	runtime.SetFinalizer(r, func(r *Ref) {
		if r.released.CompareAndSwap(false, true) {
			r.p.raw.QueueRelease(r.id)
		}
	})
	return r
}

// AdoptRef returns r as seen by c, a clone of the instance r was obtained
// from. Cloning copies the guest memory wholesale — the handle registry
// included — so a reference obtained in the prototype BEFORE its first
// Clone designates the same (copied) value in every clone under the same
// handle. Each returned Ref is independently owned: instance registries
// are separate memories, so the prototype's wrapper and every adopted
// wrapper release against their own instance.
func (c *Perl) AdoptRef(r *Ref) (*Ref, error) {
	if r == nil || r.released.Load() {
		return nil, fmt.Errorf("perl: AdoptRef of a released reference")
	}
	if c.raw.Closed() {
		return nil, internal.ErrClosed
	}
	return newRef(c, r.id, r.class, r.reftype), nil
}

// Class returns the package the reference is blessed into, or "" for an
// unblessed reference.
func (r *Ref) Class() string { return r.class }

// Reftype returns the underlying reference type as Perl reports it
// ("HASH", "ARRAY", "CODE", "SCALAR", ...).
func (r *Ref) Reftype() string { return r.reftype }

// Equal reports whether two Refs designate the same Perl reference. The
// guest deduplicates handles by refaddr, so identity comparison is id
// comparison.
func (r *Ref) Equal(o *Ref) bool {
	return r != nil && o != nil && r.p == o.p && r.id == o.id
}

// MethodCall invokes $ref->method(args...) in list context and returns the
// return list. The method is dispatched by Perl's own method resolution
// (inheritance, AUTOLOAD), and a die comes back as *PerlError.
func (r *Ref) MethodCall(ctx context.Context, method string, args ...any) ([]any, error) {
	if method == "" {
		return nil, errors.New("empty method name")
	}
	return r.p.Call(ctx, "__plwasm_method_call", append([]any{r.id, method}, args...)...)
}

// Invoke calls the referenced CODE ref with args in list context. Calling a
// non-code reference fails with *PerlError.
func (r *Ref) Invoke(ctx context.Context, args ...any) ([]any, error) {
	return r.p.Call(ctx, "__plwasm_invoke_code", append([]any{r.id}, args...)...)
}

// Export deep-copies the referenced structure into Go data (the
// encoding/json shapes: map[string]any, []any, string, float64, bool, nil).
// It carries data, not identity: mutating the result does not touch the Perl
// side. Blessed nodes that provide TO_JSON convert; code/glob nodes become
// nil.
func (r *Ref) Export(ctx context.Context) (any, error) {
	res, err := r.p.Call(ctx, "__plwasm_export", r.id)
	if err != nil {
		return nil, err
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("export returned %d values, want 1", len(res))
	}
	js, ok := res[0].(string)
	if !ok {
		return nil, fmt.Errorf("export returned %T, want the JSON text", res[0])
	}
	var v any
	if err := json.Unmarshal([]byte(js), &v); err != nil {
		return nil, fmt.Errorf("decode exported structure: %w", err)
	}
	return v, nil
}

// Free releases this Ref's registry pin now instead of waiting for the
// garbage collector. When the last pin on a handle drops the guest releases
// its reference and Perl's own refcounting takes over. Idempotent; a no-op
// on a closed instance. Using the Ref afterwards fails.
func (r *Ref) Free() error {
	if !r.released.CompareAndSwap(false, true) {
		return nil
	}
	runtime.SetFinalizer(r, nil)
	return r.p.raw.ReleaseHandle(r.id)
}

// MarshalJSON refuses: a Ref inside composite data would silently serialise
// as an empty object and lose the very identity it exists to preserve. Pass
// Refs as top-level arguments instead.
func (r *Ref) MarshalJSON() ([]byte, error) {
	return nil, errors.New("perl.Ref cannot be embedded in composite data; pass it as its own argument")
}

// encodeValue converts one Go argument into its wire node.
func (p *Perl) encodeValue(v any) (internal.WireNode, error) {
	switch x := v.(type) {
	case nil:
		return internal.WireNode{K: "u"}, nil
	case *Ref:
		if x == nil {
			return internal.WireNode{K: "u"}, nil
		}
		if x.p != p {
			return internal.WireNode{}, errors.New("perl.Ref belongs to a different Perl instance")
		}
		if x.released.Load() {
			return internal.WireNode{}, errors.New("perl.Ref has been freed")
		}
		return internal.WireNode{K: "r", H: x.id}, nil
	case GoFunc:
		return p.encodeGoFunc(x)
	case func(args []any) ([]any, error):
		return p.encodeGoFunc(x)
	case bool, string,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, json.Number:
		raw, err := json.Marshal(x)
		if err != nil {
			return internal.WireNode{}, err
		}
		return internal.WireNode{K: "d", V: raw}, nil
	default:
		raw, err := json.Marshal(x)
		if err != nil {
			return internal.WireNode{}, fmt.Errorf("encode composite argument: %w", err)
		}
		return internal.WireNode{K: "j", V: raw}, nil
	}
}

// encodeGoFunc registers a Go function value and encodes it as an "f" node:
// the guest decodes it into an ordinary Perl code ref (a closure over the id)
// that Perl can store and call later. The registration lives for the
// instance's lifetime — every crossing of a func value allocates a fresh id,
// so avoid passing per-request closures in hot paths (bind a named function
// instead). The dispatcher must already be registered; Call arranges that
// before encoding.
func (p *Perl) encodeGoFunc(fn GoFunc) (internal.WireNode, error) {
	if fn == nil {
		return internal.WireNode{K: "u"}, nil
	}
	p.funcsMu.Lock()
	if p.funcs == nil {
		p.funcs = map[int32]GoFunc{}
	}
	p.nextFuncID++
	id := p.nextFuncID
	p.funcs[id] = fn
	p.funcsMu.Unlock()
	return internal.WireNode{K: "f", H: uint64(id)}, nil
}

// hasFuncArg reports whether any argument is a Go function value (which
// needs the callback dispatcher registered before it can cross).
func hasFuncArg(args []any) bool {
	for _, a := range args {
		switch a.(type) {
		case GoFunc, func(args []any) ([]any, error):
			return true
		}
	}
	return false
}

func (p *Perl) encodeArgs(args []any) ([]internal.WireNode, error) {
	nodes := make([]internal.WireNode, len(args))
	for i, a := range args {
		n, err := p.encodeValue(a)
		if err != nil {
			return nil, fmt.Errorf("argument %d: %w", i, err)
		}
		nodes[i] = n
	}
	return nodes, nil
}

// decodeValue converts one wire node into its Go value.
func (p *Perl) decodeValue(n internal.WireNode) (any, error) {
	switch n.K {
	case "u":
		return nil, nil
	case "d", "j":
		var v any
		if err := json.Unmarshal(n.V, &v); err != nil {
			return nil, err
		}
		return v, nil
	case "r":
		return newRef(p, n.H, n.C, n.T), nil
	default:
		return nil, fmt.Errorf("unknown bridge value kind %q", n.K)
	}
}

func (p *Perl) decodeValues(nodes []internal.WireNode) ([]any, error) {
	out := make([]any, len(nodes))
	for i, n := range nodes {
		v, err := p.decodeValue(n)
		if err != nil {
			return nil, fmt.Errorf("value %d: %w", i, err)
		}
		out[i] = v
	}
	return out, nil
}

// Call invokes the named Perl subroutine in list context and returns its
// return list. name is a fully qualified sub name ("My::App::handler") or a
// main:: sub name ("handler"). Scalar results arrive by value; reference
// results arrive as *Ref handles the caller should Free when done. A
// Perl-level die (including calling a sub that does not exist) is returned
// as a *PerlError; other errors are host/transport failures or a context
// cancellation, which behaves exactly like Eval's.
func (p *Perl) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	if hasFuncArg(args) {
		// A crossing Go function needs the Perl->Go dispatcher in place
		// before the guest can call it.
		if err := p.ensureDispatcher(); err != nil {
			return nil, err
		}
	}
	nodes, err := p.encodeArgs(args)
	if err != nil {
		return nil, err
	}
	res, err := p.raw.Call(ctx, name, nodes)
	if err != nil {
		var died *internal.PerlDied
		if errors.As(err, &died) {
			return nil, &PerlError{Message: died.Message}
		}
		return nil, err
	}
	return p.decodeValues(res)
}

// perlSubName pins what Bind accepts: a Perl identifier, optionally
// package-qualified. The name is interpolated into a generated sub
// definition, so anything else is rejected up front.
var perlSubName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(::[A-Za-z_][A-Za-z0-9_]*)*$`)

// Bind makes fn callable from Perl as the subroutine name (fully qualified,
// or defined into main:: when unqualified). The Perl side sees an ordinary
// sub: scalar arguments cross by value, references arrive in Go as *Ref, the
// returned slice becomes the return list, and a Go error — or a contained
// panic — becomes a Perl die. Binding the same name again replaces the sub
// (the previous function is unreferenced but its id stays allocated).
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
// method receives the invocant as args[0]: the class name (a string) for a
// class-method call, the blessed instance (a *Ref) for an instance call;
// invocant Refs follow the same borrowed-pin rule as other bound-function
// arguments.
//
// Unless methods defines "new", a default constructor is provided:
//
//	sub new { my $class = shift; bless { @_ }, $class }
//
// so instances are ordinary Perl hash-based objects (readable from Go via
// Ref.Export, extensible by subclassing in Perl). A Go method that needs
// per-instance Go state can capture it in a closure crossing as a value, or
// key a Go-side map by an id stored in the hash.
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
// the bound function's id; req/resp carry the tagged value lists. The
// response is ALWAYS the in-band JSON envelope — a Go error return would be
// protobuf-encoded by the generated dispatch and misparse guest-side, so
// failures are reported in-band instead.
func (p *Perl) handleUserCallback(methodID int32, req []byte) ([]byte, error) {
	p.funcsMu.RLock()
	fn, ok := p.funcs[methodID]
	p.funcsMu.RUnlock()
	if !ok {
		return p.callbackResponse(nil, fmt.Errorf("no Go function bound for id %d", methodID)), nil
	}
	var nodes []internal.WireNode
	if len(req) > 0 {
		if err := json.Unmarshal(req, &nodes); err != nil {
			return p.callbackResponse(nil, fmt.Errorf("decode arguments: %w", err)), nil
		}
	}
	args, err := p.decodeValues(nodes)
	if err != nil {
		return p.callbackResponse(nil, fmt.Errorf("decode arguments: %w", err)), nil
	}
	results, err := safeCall(fn, args)
	return p.callbackResponse(results, err), nil
}

// safeCall contains a panicking GoFunc so a guest-triggered call cannot take
// the whole process down; the panic surfaces as a Perl die.
func safeCall(fn GoFunc, args []any) (results []any, err error) {
	defer func() {
		if r := recover(); r != nil {
			results, err = nil, fmt.Errorf("Go function panicked: %v", r)
		}
	}()
	return fn(args)
}

// callbackResponse encodes the in-band Perl-facing response envelope.
func (p *Perl) callbackResponse(results []any, err error) []byte {
	type envelope struct {
		Ok     bool                `json:"ok"`
		Result []internal.WireNode `json:"result"`
		Error  string              `json:"error"`
	}
	env := envelope{Ok: err == nil, Result: []internal.WireNode{}}
	if err == nil {
		for i, v := range results {
			n, eErr := p.encodeValue(v)
			if eErr != nil {
				err = fmt.Errorf("encode Go result %d: %w", i, eErr)
				break
			}
			env.Result = append(env.Result, n)
		}
	}
	if err != nil {
		env = envelope{Ok: false, Result: []internal.WireNode{}, Error: err.Error()}
	}
	out, mErr := json.Marshal(env)
	if mErr != nil {
		out, _ = json.Marshal(envelope{Ok: false, Result: []internal.WireNode{},
			Error: fmt.Sprintf("encode Go response: %v", mErr)})
	}
	return out
}
