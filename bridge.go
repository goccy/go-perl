package perl

// The Go <-> Perl function bridge.
//
// Values cross the boundary as JSON, encoded/decoded guest-side by JSON::PP
// (loaded lazily on first use) and host-side by encoding/json: Perl scalars
// map to JSON strings/numbers/bools, array refs to arrays, hash refs to
// objects, undef to null. On the Go side that means string, float64, bool,
// []any, map[string]any and nil — the encoding/json defaults.
//
// Go -> Perl: Call invokes a named Perl sub through the perl_call bridge
// export, which routes the JSON-decoded arguments through a dispatcher sub
// the guest installed at boot.
//
// Perl -> Go: Bind registers a Go function under an id and evals a one-line
// Perl sub that forwards its arguments to main::__plwasm_go_call(id, @_) —
// glue the guest installed at boot that JSON-encodes the arguments, calls the
// __plwasm_go_invoke XS (which forwards the bytes to this package through the
// wasmify callback import), and decodes the response, dying with the error
// text when the Go function failed.

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	wasm2go "github.com/goccy/perlwasm2go"
)

// PerlError is a Perl-level die surfaced by Call: the sub (or the argument
// decode) died and Message holds $@.
type PerlError struct {
	Message string
}

func (e *PerlError) Error() string { return e.Message }

// GoFunc is a Go function callable from Perl. args is the Perl call's
// argument list decoded from JSON; the returned slice becomes the Perl
// call's return list. Returning an error makes the Perl call die with the
// error text.
//
// fn may call back into the instance (Eval, Call) — the instance lock is
// released for the callback's duration — but must not use it from other
// goroutines concurrently.
type GoFunc func(args []any) ([]any, error)

// callEnvelope is the decoded perl_call response document.
type callEnvelope struct {
	Ok     bool   `json:"ok"`
	Result []any  `json:"result"`
	Error  string `json:"error"`
}

// Call invokes the named Perl subroutine in list context and returns its
// return list. name is a fully qualified sub name ("My::App::handler") or a
// main:: sub name ("handler"). A Perl-level die (including calling a sub
// that does not exist) is returned as a *PerlError; other errors are
// host/transport failures or a context cancellation, which behaves exactly
// like Eval's.
func (p *Perl) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	argsJSON := []byte("[]")
	if len(args) > 0 {
		var err error
		if argsJSON, err = json.Marshal(args); err != nil {
			return nil, fmt.Errorf("encode arguments: %w", err)
		}
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
		if interrupted {
			return nil, ctx.Err()
		}
		return nil, &PerlError{Message: env.Error}
	}
	return env.Result, nil
}

// perlSubName pins what Bind accepts: a Perl identifier, optionally
// package-qualified. The name is interpolated into a generated sub
// definition, so anything else is rejected up front.
var perlSubName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(::[A-Za-z_][A-Za-z0-9_]*)*$`)

// Bind makes fn callable from Perl as the subroutine name (fully qualified,
// or defined into main:: when unqualified). The Perl side sees an ordinary
// sub: arguments cross as JSON, the returned slice becomes the return list,
// and a Go error becomes a Perl die. Binding the same name again replaces
// the sub (the previous function is unreferenced but its id stays
// allocated).
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
	if !r.Ok {
		return fmt.Errorf("define %s: %s", name, r.Error)
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
	resp, err := p.m.invoke(0, midSetGoDispatcher, buf, wasm2go.Inv_0_5)
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
// sub. methodID carries the bound function's id; req/resp carry JSON. The
// response is ALWAYS a JSON document ({"ok":...,"result":...,"error":...}) —
// a Go error return would be protobuf-encoded by the generated dispatch and
// misparse guest-side, so failures are reported in-band instead.
type goDispatcher struct{ p *Perl }

func (d goDispatcher) HandleCallback(methodID int32, req []byte) ([]byte, error) {
	d.p.funcsMu.RLock()
	fn, ok := d.p.funcs[methodID]
	d.p.funcsMu.RUnlock()
	if !ok {
		return goCallResponse(nil, fmt.Errorf("no Go function bound for id %d", methodID)), nil
	}
	var args []any
	if len(req) > 0 {
		if err := json.Unmarshal(req, &args); err != nil {
			return goCallResponse(nil, fmt.Errorf("decode arguments: %w", err)), nil
		}
	}
	results, err := safeCall(fn, args)
	return goCallResponse(results, err), nil
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

// goCallResponse encodes the in-band Perl-facing response envelope.
func goCallResponse(results []any, err error) []byte {
	type envelope struct {
		Ok     bool   `json:"ok"`
		Result []any  `json:"result"`
		Error  string `json:"error"`
	}
	env := envelope{Ok: err == nil, Result: results}
	if err != nil {
		env.Error = err.Error()
	}
	if env.Result == nil {
		env.Result = []any{}
	}
	out, mErr := json.Marshal(env)
	if mErr != nil {
		// A result the host cannot encode (NaN, a channel, ...) — report
		// that instead of failing the transport.
		out, _ = json.Marshal(envelope{Ok: false, Result: []any{},
			Error: fmt.Sprintf("encode Go result: %v", mErr)})
	}
	return out
}
