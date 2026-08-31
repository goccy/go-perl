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
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"sync/atomic"

	wasm2go "github.com/goccy/perlwasm2go"
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

// wireNode is one tagged value on the bridge:
//
//	k="u"                 undef / nil
//	k="d" v=<scalar>      plain scalar by value
//	k="j" v=<structure>   composite data by value (fresh structures on decode)
//	k="r" h=<id> t/c      a Perl reference by handle (t=reftype, c=class)
type wireNode struct {
	K string          `json:"k"`
	V json.RawMessage `json:"v,omitempty"`
	H uint64          `json:"h,omitempty"`
	T string          `json:"t,omitempty"`
	C string          `json:"c,omitempty"`
}

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
			r.p.queueRelease(r.id)
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
	if c.closed.Load() {
		return nil, errClosed
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
	if r.p.closed.Load() {
		return nil
	}
	_, err := r.p.call(context.Background(), "__plwasm_release", []any{r.id}, false)
	return err
}

// MarshalJSON refuses: a Ref inside composite data would silently serialise
// as an empty object and lose the very identity it exists to preserve. Pass
// Refs as top-level arguments instead.
func (r *Ref) MarshalJSON() ([]byte, error) {
	return nil, errors.New("perl.Ref cannot be embedded in composite data; pass it as its own argument")
}

// encodeValue converts one Go argument into its wire node.
func (p *Perl) encodeValue(v any) (wireNode, error) {
	switch x := v.(type) {
	case nil:
		return wireNode{K: "u"}, nil
	case *Ref:
		if x == nil {
			return wireNode{K: "u"}, nil
		}
		if x.p != p {
			return wireNode{}, errors.New("perl.Ref belongs to a different Perl instance")
		}
		if x.released.Load() {
			return wireNode{}, errors.New("perl.Ref has been freed")
		}
		return wireNode{K: "r", H: x.id}, nil
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
			return wireNode{}, err
		}
		return wireNode{K: "d", V: raw}, nil
	default:
		raw, err := json.Marshal(x)
		if err != nil {
			return wireNode{}, fmt.Errorf("encode composite argument: %w", err)
		}
		return wireNode{K: "j", V: raw}, nil
	}
}

// encodeGoFunc registers a Go function value and encodes it as an "f" node:
// the guest decodes it into an ordinary Perl code ref (a closure over the id)
// that Perl can store and call later. The registration lives for the
// instance's lifetime — every crossing of a func value allocates a fresh id,
// so avoid passing per-request closures in hot paths (bind a named function
// instead). The dispatcher must already be registered; Call arranges that
// before encoding.
func (p *Perl) encodeGoFunc(fn GoFunc) (wireNode, error) {
	if fn == nil {
		return wireNode{K: "u"}, nil
	}
	p.funcsMu.Lock()
	if p.funcs == nil {
		p.funcs = map[int32]GoFunc{}
	}
	p.nextFuncID++
	id := p.nextFuncID
	p.funcs[id] = fn
	p.funcsMu.Unlock()
	return wireNode{K: "f", H: uint64(id)}, nil
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

func (p *Perl) encodeArgs(args []any) ([]byte, error) {
	nodes := make([]wireNode, len(args))
	for i, a := range args {
		n, err := p.encodeValue(a)
		if err != nil {
			return nil, fmt.Errorf("argument %d: %w", i, err)
		}
		nodes[i] = n
	}
	return json.Marshal(nodes)
}

// decodeValue converts one wire node into its Go value.
func (p *Perl) decodeValue(n wireNode) (any, error) {
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

func (p *Perl) decodeValues(nodes []wireNode) ([]any, error) {
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

// callEnvelope is the decoded perl_call response document. Exit is set (and
// Ok false) when the guest called exit() and the bridge caught the unwind.
type callEnvelope struct {
	Ok     bool       `json:"ok"`
	Result []wireNode `json:"result"`
	Exit   *int       `json:"exit"`
	Error  string     `json:"error"`
}

// Call invokes the named Perl subroutine in list context and returns its
// return list. name is a fully qualified sub name ("My::App::handler") or a
// main:: sub name ("handler"). Scalar results arrive by value; reference
// results arrive as *Ref handles the caller should Free when done. A
// Perl-level die (including calling a sub that does not exist) is returned
// as a *PerlError; other errors are host/transport failures or a context
// cancellation, which behaves exactly like Eval's.
func (p *Perl) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	return p.call(ctx, name, args, true)
}

// call is Call's engine; drain=false skips the release-queue flush (the
// flush itself and Ref.Free use that path, so a drain cannot recurse).
func (p *Perl) call(ctx context.Context, name string, args []any, drain bool) ([]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.closed.Load() {
		return nil, errClosed
	}
	if drain {
		p.drainReleases(ctx)
	}
	if hasFuncArg(args) {
		// A crossing Go function needs the Perl->Go dispatcher in place
		// before the guest can call it.
		if err := p.ensureDispatcher(); err != nil {
			return nil, err
		}
	}
	argsJSON, err := p.encodeArgs(args)
	if err != nil {
		return nil, err
	}

	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendString(buf, 2, name)
	buf = pbAppendString(buf, 3, string(argsJSON))

	disarm := p.armInterrupt(ctx)
	resp, invokeErr := p.m.invoke(0, midCall, buf, wasm2go.Inv_0_0)
	interrupted := disarm()

	if invokeErr != nil {
		return nil, invokeErr
	}
	if e := pbExtractError(resp); e != nil {
		return nil, e
	}
	js := readScalarAtField(resp, 1, (*pbReader).readString)
	var env callEnvelope
	if err := json.Unmarshal([]byte(js), &env); err != nil {
		return nil, fmt.Errorf("decode call result %q: %w", js, err)
	}
	if !env.Ok {
		if env.Exit != nil {
			return nil, &exitStatusError{code: *env.Exit}
		}
		if interrupted {
			return nil, ctx.Err()
		}
		return nil, &PerlError{Message: env.Error}
	}
	return p.decodeValues(env.Result)
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

// ensureDispatcher registers this instance's callback handler with the
// generated bridge machinery and tells the guest its callback id. Runs once
// per instance.
func (p *Perl) ensureDispatcher() error {
	p.funcsMu.Lock()
	if p.dispatcherSet {
		p.funcsMu.Unlock()
		return nil
	}
	if p.funcs == nil {
		p.funcs = map[int32]GoFunc{}
	}
	p.funcsMu.Unlock()

	// Register on this instance's Module (NOT the package-level
	// RegisterCallback, which addresses the global single-instance binding).
	p.m.cbMu.Lock()
	if p.m.callbacks == nil {
		p.m.callbacks = map[int32]CallbackHandler{}
	}
	p.m.nextCBID++
	cbID := p.m.nextCBID
	p.m.callbacks[cbID] = goDispatcher{p: p}
	p.m.cbMu.Unlock()

	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendInt32(buf, 2, cbID)
	resp, err := p.m.invoke(0, midSetGoDispatcher, buf, wasm2go.Inv_0_6)
	if err != nil {
		return fmt.Errorf("set Go dispatcher: %w", err)
	}
	if e := pbExtractError(resp); e != nil {
		return fmt.Errorf("set Go dispatcher: %w", e)
	}

	p.funcsMu.Lock()
	p.dispatcherSet = true
	p.funcsMu.Unlock()
	return nil
}

// goDispatcher is the per-instance CallbackHandler behind every bound Perl
// sub. methodID carries the bound function's id; req/resp carry the tagged
// value lists. The response is ALWAYS the in-band JSON envelope — a Go error
// return would be protobuf-encoded by the generated dispatch and misparse
// guest-side, so failures are reported in-band instead.
type goDispatcher struct{ p *Perl }

func (d goDispatcher) HandleCallback(methodID int32, req []byte) ([]byte, error) {
	if methodID == nativeXSMethodID {
		d.p.funcsMu.RLock()
		native := d.p.nativeXS
		d.p.funcsMu.RUnlock()
		if native == nil {
			return append([]byte{0}, "no native XS handler installed"...), nil
		}
		return native(req), nil
	}
	if methodID == magicFreeMethodID {
		d.p.funcsMu.RLock()
		free := d.p.magicFree
		d.p.funcsMu.RUnlock()
		if free != nil && len(req) >= 4 {
			free(binary.LittleEndian.Uint32(req))
		}
		return []byte{1}, nil
	}
	if methodID == setMagicMethodID {
		d.p.funcsMu.RLock()
		set := d.p.magicSet
		d.p.funcsMu.RUnlock()
		if set != nil && len(req) >= 4 {
			set(binary.LittleEndian.Uint32(req))
		}
		return []byte{1}, nil
	}
	if methodID == ppHookMethodID {
		d.p.funcsMu.RLock()
		hook := d.p.ppHook
		d.p.funcsMu.RUnlock()
		if hook == nil {
			return append([]byte{0}, "no pp hook handler installed"...), nil
		}
		return hook(req), nil
	}
	if methodID == keywordMethodID {
		d.p.funcsMu.RLock()
		hook := d.p.keywordHook
		d.p.funcsMu.RUnlock()
		if hook == nil {
			// no parse-surface module loaded: decline so the guest chain runs
			return []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil
		}
		return hook(req), nil
	}
	if methodID == perlioMethodID {
		d.p.funcsMu.RLock()
		hook := d.p.perlioHook
		d.p.funcsMu.RUnlock()
		if hook == nil {
			return append([]byte{0}, "no PerlIO layer handler installed"...), nil
		}
		return hook(req), nil
	}
	if methodID == destructorMethodID {
		d.p.funcsMu.RLock()
		fire := d.p.dtorFire
		d.p.funcsMu.RUnlock()
		if fire != nil && len(req) >= 4 {
			fire(binary.LittleEndian.Uint32(req))
		}
		return []byte{1}, nil
	}
	d.p.funcsMu.RLock()
	fn, ok := d.p.funcs[methodID]
	d.p.funcsMu.RUnlock()
	if !ok {
		return d.response(nil, fmt.Errorf("no Go function bound for id %d", methodID)), nil
	}
	var nodes []wireNode
	if len(req) > 0 {
		if err := json.Unmarshal(req, &nodes); err != nil {
			return d.response(nil, fmt.Errorf("decode arguments: %w", err)), nil
		}
	}
	args, err := d.p.decodeValues(nodes)
	if err != nil {
		return d.response(nil, fmt.Errorf("decode arguments: %w", err)), nil
	}
	results, err := safeCall(fn, args)
	return d.response(results, err), nil
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

// response encodes the in-band Perl-facing response envelope.
func (d goDispatcher) response(results []any, err error) []byte {
	type envelope struct {
		Ok     bool       `json:"ok"`
		Result []wireNode `json:"result"`
		Error  string     `json:"error"`
	}
	env := envelope{Ok: err == nil, Result: []wireNode{}}
	if err == nil {
		for i, v := range results {
			n, eErr := d.p.encodeValue(v)
			if eErr != nil {
				err = fmt.Errorf("encode Go result %d: %w", i, eErr)
				break
			}
			env.Result = append(env.Result, n)
		}
	}
	if err != nil {
		env = envelope{Ok: false, Result: []wireNode{}, Error: err.Error()}
	}
	out, mErr := json.Marshal(env)
	if mErr != nil {
		out, _ = json.Marshal(envelope{Ok: false, Result: []wireNode{},
			Error: fmt.Sprintf("encode Go response: %v", mErr)})
	}
	return out
}
