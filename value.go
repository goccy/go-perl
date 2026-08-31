package perl

// The Perl value system as seen from Go, following Perl's own type
// structure: SV (scalar value), RV (reference), AV (array value), HV (hash
// value), CV (code value), GV (glob value), IO. Every Value crossing the
// bridge is typed — scalars by value with their kind, references by handle —
// and nothing is stringified in transit.
//
// Two access paths, both type safe:
//
//   - runtime inspection: a Go type switch over the concrete types (or
//     Kind() when only the kind matters);
//   - a known type up front: As[T] extracts the concrete type or errors.

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync/atomic"

	"github.com/goccy/go-perl/internal"
)

// Kind is the runtime kind of a Value, one level finer than the concrete
// type for scalars (mirroring reflect.Value.Kind): a ScalarValue reports
// which representation its scalar holds, every other concrete type reports
// its single kind.
type Kind uint8

const (
	// KindUndef is an undefined scalar (SV: undef).
	KindUndef Kind = iota
	// KindBool is a core boolean scalar (SV: !!1 / !!0).
	KindBool
	// KindInt is an integer scalar (SV: IV).
	KindInt
	// KindFloat is a floating-point scalar (SV: NV).
	KindFloat
	// KindString is a string scalar (SV: PV — a byte string).
	KindString
	// KindRef is a reference (RV).
	KindRef
	// KindArray is an array (AV).
	KindArray
	// KindHash is a hash (HV).
	KindHash
	// KindCode is a subroutine (CV).
	KindCode
	// KindGlob is a typeglob (GV).
	KindGlob
	// KindIO is a filehandle body (IO).
	KindIO
)

var kindNames = [...]string{
	"undef", "bool", "int", "float", "string",
	"ref", "array", "hash", "code", "glob", "io",
}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind(" + strconv.Itoa(int(k)) + ")"
}

// Value is any Perl value. Sealed: the concrete types — ScalarValue,
// RefValue, ArrayValue, HashValue, CodeValue, GlobValue, IOValue — are the
// only implementations, so a type switch over them is exhaustive.
type Value interface {
	Kind() Kind
	sealed()
}

// PerlValue enumerates the concrete Value types, as the constraint behind
// the typed extraction path (As and friends).
type PerlValue interface {
	ScalarValue | RefValue | ArrayValue | HashValue | CodeValue | GlobValue | IOValue
}

// As extracts the concrete type T from v, for call sites that know the
// expected type up front; a kind mismatch is an error, never a panic. For
// runtime inspection use a type switch over the concrete types instead.
func As[T PerlValue](v Value) (T, error) {
	if t, ok := any(v).(T); ok {
		return t, nil
	}
	var zero T
	return zero, fmt.Errorf("perl: value is %s, not %T", v.Kind(), zero)
}

// ---- ScalarValue (SV) ----------------------------------------------------

// ScalarValue is one Perl scalar: undef, a boolean, an integer, a float, or
// a byte string. The accessors coerce the way Perl itself would, totally —
// they never panic. Kind reports which representation the scalar holds.
type ScalarValue struct {
	kind Kind
	b    bool
	i    int64
	f    float64
	s    []byte
	// utf8 mirrors the guest SvUTF8 flag so the encoding round-trips.
	utf8 bool
}

func (ScalarValue) sealed() {}

// Kind reports the scalar's representation: KindUndef, KindBool, KindInt,
// KindFloat, or KindString.
func (v ScalarValue) Kind() Kind { return v.kind }

// Bool applies Perl truth: undef, "", "0", and zero are false, everything
// else is true.
func (v ScalarValue) Bool() bool {
	switch v.kind {
	case KindBool:
		return v.b
	case KindInt:
		return v.i != 0
	case KindFloat:
		return v.f != 0
	case KindString:
		return len(v.s) != 0 && string(v.s) != "0"
	default:
		return false
	}
}

// Int numifies the scalar the way Perl does and truncates toward zero.
func (v ScalarValue) Int() int64 {
	switch v.kind {
	case KindBool:
		if v.b {
			return 1
		}
		return 0
	case KindInt:
		return v.i
	case KindFloat:
		return int64(v.f)
	case KindString:
		return int64(numify(string(v.s)))
	default:
		return 0
	}
}

// Float numifies the scalar the way Perl does (a leading numeric prefix
// counts, anything else is 0).
func (v ScalarValue) Float() float64 {
	switch v.kind {
	case KindBool:
		if v.b {
			return 1
		}
		return 0
	case KindInt:
		return float64(v.i)
	case KindFloat:
		return v.f
	case KindString:
		return numify(string(v.s))
	default:
		return 0
	}
}

// String is the scalar's string form, as Perl would stringify it (a boolean
// true is "1", false is ""; undef is "").
func (v ScalarValue) String() string {
	switch v.kind {
	case KindBool:
		if v.b {
			return "1"
		}
		return ""
	case KindInt:
		return strconv.FormatInt(v.i, 10)
	case KindFloat:
		return strconv.FormatFloat(v.f, 'g', 15, 64)
	case KindString:
		return string(v.s)
	default:
		return ""
	}
}

// Bytes is the scalar's raw byte string: for a string scalar the exact PV
// bytes (no re-encoding), otherwise the bytes of String().
func (v ScalarValue) Bytes() []byte {
	if v.kind == KindString {
		return v.s
	}
	return []byte(v.String())
}

// ScalarType constrains the Go types NewValue accepts as one Perl scalar.
type ScalarType interface {
	bool | int | int32 | int64 | uint32 | float64 | string | []byte
}

// NewValue builds the scalar for a Go value: a bool crosses as a Perl core
// boolean, integers as IVs, floats as NVs, a string as a utf8 character
// string, and a []byte as a raw byte string.
func NewValue[T ScalarType](v T) ScalarValue {
	switch x := any(v).(type) {
	case bool:
		return ScalarValue{kind: KindBool, b: x}
	case int:
		return ScalarValue{kind: KindInt, i: int64(x)}
	case int32:
		return ScalarValue{kind: KindInt, i: int64(x)}
	case int64:
		return ScalarValue{kind: KindInt, i: x}
	case uint32:
		return ScalarValue{kind: KindInt, i: int64(x)}
	case float64:
		return ScalarValue{kind: KindFloat, f: x}
	case string:
		return ScalarValue{kind: KindString, s: []byte(x), utf8: true}
	default:
		return ScalarValue{kind: KindString, s: any(v).([]byte)}
	}
}

// Undef is the undefined scalar.
func Undef() ScalarValue { return ScalarValue{kind: KindUndef} }

// ---- reference handles ---------------------------------------------------

// refHandle is the shared pin behind every handle-bearing Value. Views of
// one reference (the RefValue and the ArrayValue its Deref produced, for
// example) share the pin; the guest registry entry is released when the last
// view becomes unreachable, or when the instance closes.
type refHandle struct {
	p       *Perl
	id      uint64
	refkind uint8
	class   string
	// released flips once the release has been queued so views never
	// double-release.
	released atomic.Bool
}

func newRefHandle(p *Perl, id uint64, refkind uint8, class string) *refHandle {
	h := &refHandle{p: p, id: id, refkind: refkind, class: class}
	runtime.SetFinalizer(h, func(h *refHandle) {
		if h.released.CompareAndSwap(false, true) {
			h.p.raw.QueueRelease(h.id)
		}
	})
	return h
}

// RefValue is a reference (RV): a handle to a Perl reference held alive in
// the interpreter's registry. Identity is preserved — the same Perl
// reference always surfaces with the same handle, and sending it back
// dereferences to the same referent.
type RefValue struct {
	h *refHandle
}

func (RefValue) sealed() {}

// Kind returns KindRef.
func (v RefValue) Kind() Kind { return KindRef }

// Class returns the package the reference is blessed into, and whether it
// is blessed at all.
func (v RefValue) Class() (string, bool) { return v.h.class, v.h.class != "" }

// Equal reports whether two references designate the same Perl reference.
// The guest deduplicates handles by referent address, so identity
// comparison is handle comparison.
func (v RefValue) Equal(o RefValue) bool {
	return v.h != nil && o.h != nil && v.h.p == o.h.p && v.h.id == o.h.id
}

// Deref dereferences the reference — the reflect.Value.Elem analog:
//
//	\$x  -> the ScalarValue $x (or a fresh RefValue for a ref-to-ref)
//	\@a  -> the ArrayValue @a
//	\%h  -> the HashValue %h
//	\&f  -> the CodeValue &f
//	\*g  -> the GlobValue *g
//
// Aggregate derefs are view constructions (no guest call); a scalar deref
// reads the referent from the guest.
func (v RefValue) Deref(ctx context.Context) (Value, error) {
	switch v.h.refkind {
	case wireRefArray:
		return ArrayValue{h: v.h}, nil
	case wireRefHash:
		return HashValue{h: v.h}, nil
	case wireRefCode:
		return CodeValue{h: v.h}, nil
	case wireRefGlob:
		return GlobValue{h: v.h}, nil
	case wireRefIO:
		return IOValue{h: v.h}, nil
	default:
		resp, interrupted, err := v.h.p.raw.DerefOp(ctx, v.h.id)
		if err != nil {
			return nil, err
		}
		return v.h.p.decodeNodeResult(ctx, resp, interrupted)
	}
}

// MethodCall invokes $ref->method(args...) in list context and returns the
// return list, dispatched by Perl's own method resolution (inheritance,
// AUTOLOAD). A die comes back as *PerlError.
func (v RefValue) MethodCall(ctx context.Context, method string, args ...Value) ([]Value, error) {
	if method == "" {
		return nil, errors.New("perl: empty method name")
	}
	p := v.h.p
	enc, err := p.encodeArgs(args)
	if err != nil {
		return nil, err
	}
	resp, interrupted, err := p.raw.MethodCallOp(ctx, v.h.id, method, enc)
	if err != nil {
		return nil, err
	}
	return p.decodeListResult(ctx, resp, interrupted)
}

// Isa reports $ref->isa(class) — whether the reference's class is or
// inherits from class.
func (v RefValue) Isa(ctx context.Context, class string) (bool, error) {
	res, err := v.MethodCall(ctx, "isa", NewValue(class))
	if err != nil {
		return false, err
	}
	return len(res) > 0 && truthy(res[0]), nil
}

// Can reports $ref->can(method) — whether the class provides the method.
func (v RefValue) Can(ctx context.Context, method string) (bool, error) {
	res, err := v.MethodCall(ctx, "can", NewValue(method))
	if err != nil {
		return false, err
	}
	return len(res) > 0 && truthy(res[0]), nil
}

// truthy applies Perl truth to any Value (references are always true).
func truthy(v Value) bool {
	if s, ok := v.(ScalarValue); ok {
		return s.Bool()
	}
	return v != nil && v.Kind() != KindUndef
}

// ---- ArrayValue (AV) -----------------------------------------------------

// ArrayValue is a Perl array (AV), viewed through the reference that
// reached Go. Operations run real Perl element accesses, so ties and
// overloads behave as in plain Perl; a die comes back as *PerlError.
type ArrayValue struct {
	h *refHandle
}

func (ArrayValue) sealed() {}

// Kind returns KindArray.
func (v ArrayValue) Kind() Kind { return KindArray }

// Ref returns the reference to this array (\@a).
func (v ArrayValue) Ref() RefValue { return RefValue{h: v.h} }

// Len returns scalar @array.
func (v ArrayValue) Len(ctx context.Context) (int, error) {
	resp, interrupted, err := v.h.p.raw.ArrayLenOp(ctx, v.h.id)
	if err != nil {
		return 0, err
	}
	n, err := v.h.p.decodeI64Result(ctx, resp, interrupted)
	return int(n), err
}

// Index returns $array[i] (Perl indexing: negative counts from the end).
func (v ArrayValue) Index(ctx context.Context, i int) (Value, error) {
	resp, interrupted, err := v.h.p.raw.ArrayGetOp(ctx, v.h.id, int64(i))
	if err != nil {
		return nil, err
	}
	return v.h.p.decodeNodeResult(ctx, resp, interrupted)
}

// SetIndex performs $array[i] = val.
func (v ArrayValue) SetIndex(ctx context.Context, i int, val Value) error {
	enc, err := v.h.p.encodeSingle(val)
	if err != nil {
		return err
	}
	resp, interrupted, err := v.h.p.raw.ArraySetOp(ctx, v.h.id, int64(i), enc)
	if err != nil {
		return err
	}
	return v.h.p.decodeEmptyResult(ctx, resp, interrupted)
}

// Push appends vals to the array. An ArrayValue or HashValue among vals
// flattens into its contents, exactly like push @a, @b, %h in Perl.
func (v ArrayValue) Push(ctx context.Context, vals ...Value) error {
	enc, err := v.h.p.encodeArgs(vals)
	if err != nil {
		return err
	}
	resp, interrupted, err := v.h.p.raw.ArrayPushOp(ctx, v.h.id, enc)
	if err != nil {
		return err
	}
	return v.h.p.decodeEmptyResult(ctx, resp, interrupted)
}

// Values returns every element in one crossing.
func (v ArrayValue) Values(ctx context.Context) ([]Value, error) {
	resp, interrupted, err := v.h.p.raw.ArrayValuesOp(ctx, v.h.id)
	if err != nil {
		return nil, err
	}
	return v.h.p.decodeListResult(ctx, resp, interrupted)
}

// ---- HashValue (HV) ------------------------------------------------------

// HashValue is a Perl hash (HV), viewed through the reference that reached
// Go. Operations run real Perl accesses, so ties behave as in plain Perl.
type HashValue struct {
	h *refHandle
}

func (HashValue) sealed() {}

// Kind returns KindHash.
func (v HashValue) Kind() Kind { return KindHash }

// Ref returns the reference to this hash (\%h).
func (v HashValue) Ref() RefValue { return RefValue{h: v.h} }

// Get returns $hash{key} and whether the key exists.
func (v HashValue) Get(ctx context.Context, key string) (Value, bool, error) {
	resp, interrupted, err := v.h.p.raw.HashGetOp(ctx, v.h.id, encodeKey(key))
	if err != nil {
		return nil, false, err
	}
	return v.h.p.decodeExistsNodeResult(ctx, resp, interrupted)
}

// Set performs $hash{key} = val.
func (v HashValue) Set(ctx context.Context, key string, val Value) error {
	enc, err := v.h.p.encodeSingle(val)
	if err != nil {
		return err
	}
	resp, interrupted, err := v.h.p.raw.HashSetOp(ctx, v.h.id, encodeKey(key), enc)
	if err != nil {
		return err
	}
	return v.h.p.decodeEmptyResult(ctx, resp, interrupted)
}

// Delete performs delete $hash{key}.
func (v HashValue) Delete(ctx context.Context, key string) error {
	resp, interrupted, err := v.h.p.raw.HashDeleteOp(ctx, v.h.id, encodeKey(key))
	if err != nil {
		return err
	}
	return v.h.p.decodeEmptyResult(ctx, resp, interrupted)
}

// Keys returns keys %hash (Perl's hash order: unordered).
func (v HashValue) Keys(ctx context.Context) ([]string, error) {
	resp, interrupted, err := v.h.p.raw.HashKeysOp(ctx, v.h.id)
	if err != nil {
		return nil, err
	}
	vals, err := v.h.p.decodeListResult(ctx, resp, interrupted)
	if err != nil {
		return nil, err
	}
	keys := make([]string, len(vals))
	for i, kv := range vals {
		s, ok := kv.(ScalarValue)
		if !ok {
			return nil, fmt.Errorf("perl: hash key %d is %s, not a scalar", i, kv.Kind())
		}
		keys[i] = s.String()
	}
	return keys, nil
}

// ---- CodeValue (CV) ------------------------------------------------------

// CodeValue is a Perl subroutine (CV), viewed through the code reference
// that reached Go.
type CodeValue struct {
	h *refHandle
}

func (CodeValue) sealed() {}

// Kind returns KindCode.
func (v CodeValue) Kind() Kind { return KindCode }

// Ref returns the reference to this subroutine (\&f).
func (v CodeValue) Ref() RefValue { return RefValue{h: v.h} }

// Call invokes the subroutine in list context and returns its return list.
// An ArrayValue or HashValue among args flattens into the argument list,
// exactly like f(@a, %h) in Perl; pass Ref() to pass the reference itself.
// A die comes back as *PerlError.
func (v CodeValue) Call(ctx context.Context, args ...Value) ([]Value, error) {
	return v.call(ctx, false, args)
}

// CallScalar invokes the subroutine in scalar context and returns its
// single result.
func (v CodeValue) CallScalar(ctx context.Context, args ...Value) (Value, error) {
	res, err := v.call(ctx, true, args)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return Undef(), nil
	}
	return res[0], nil
}

func (v CodeValue) call(ctx context.Context, scalarCtx bool, args []Value) ([]Value, error) {
	p := v.h.p
	enc, err := p.encodeArgs(args)
	if err != nil {
		return nil, err
	}
	resp, interrupted, err := p.raw.InvokeOp(ctx, v.h.id, scalarCtx, enc)
	if err != nil {
		return nil, err
	}
	return p.decodeListResult(ctx, resp, interrupted)
}

// ---- GlobValue (GV) / IOValue (IO) ---------------------------------------

// GlobValue is a typeglob (GV) — a filehandle in most user code. Opaque for
// now: it can be inspected, held, and passed back to Perl; operations are
// added as needs arise.
type GlobValue struct {
	h *refHandle
}

func (GlobValue) sealed() {}

// Kind returns KindGlob.
func (v GlobValue) Kind() Kind { return KindGlob }

// Ref returns the reference to this glob (\*g).
func (v GlobValue) Ref() RefValue { return RefValue{h: v.h} }

// IOValue is a filehandle body (IO). Opaque for now, like GlobValue.
type IOValue struct {
	h *refHandle
}

func (IOValue) sealed() {}

// Kind returns KindIO.
func (v IOValue) Kind() Kind { return KindIO }

// Ref returns the reference to this IO (*g{IO} as a reference).
func (v IOValue) Ref() RefValue { return RefValue{h: v.h} }

// ---- construction of aggregates ------------------------------------------

// Pair is one key/value entry for NewHash.
type Pair struct {
	K string
	V Value
}

// NewArray materialises a fresh array in the guest holding vals (one
// crossing) and returns it. An ArrayValue or HashValue among vals flattens,
// like [@a, %h]; pass Ref() to nest a reference.
func (p *Perl) NewArray(ctx context.Context, vals ...Value) (ArrayValue, error) {
	enc, err := p.encodeArgs(vals)
	if err != nil {
		return ArrayValue{}, err
	}
	resp, interrupted, err := p.raw.NewArrayOp(ctx, enc)
	if err != nil {
		return ArrayValue{}, err
	}
	v, err := p.decodeNodeResult(ctx, resp, interrupted)
	if err != nil {
		return ArrayValue{}, err
	}
	ref, err := As[RefValue](v)
	if err != nil {
		return ArrayValue{}, err
	}
	return ArrayValue{h: ref.h}, nil
}

// NewHash materialises a fresh hash in the guest holding pairs (one
// crossing) and returns it.
func (p *Perl) NewHash(ctx context.Context, pairs ...Pair) (HashValue, error) {
	var enc []byte
	enc = appendU32(enc, uint32(len(pairs)*2))
	for _, pr := range pairs {
		enc = append(enc, encodeKey(pr.K)...)
		one, err := p.encodeSingle(pr.V)
		if err != nil {
			return HashValue{}, err
		}
		enc = append(enc, one...)
	}
	resp, interrupted, err := p.raw.NewHashOp(ctx, enc)
	if err != nil {
		return HashValue{}, err
	}
	v, err := p.decodeNodeResult(ctx, resp, interrupted)
	if err != nil {
		return HashValue{}, err
	}
	ref, err := As[RefValue](v)
	if err != nil {
		return HashValue{}, err
	}
	return HashValue{h: ref.h}, nil
}

// Adopt returns v as seen by c, a clone of the instance v was obtained
// from. Cloning copies the guest memory wholesale — the handle registry
// included — so a reference obtained in the prototype BEFORE its first
// Clone designates the same (copied) value in every clone under the same
// handle. Scalars are instance-free and pass through unchanged.
func Adopt[T PerlValue](c *Perl, v T) (T, error) {
	var zero T
	val := Value(v)
	switch x := val.(type) {
	case ScalarValue:
		return v, nil
	case RefValue:
		h, err := c.adoptHandle(x.h)
		if err != nil {
			return zero, err
		}
		return any(RefValue{h: h}).(T), nil
	case ArrayValue:
		h, err := c.adoptHandle(x.h)
		if err != nil {
			return zero, err
		}
		return any(ArrayValue{h: h}).(T), nil
	case HashValue:
		h, err := c.adoptHandle(x.h)
		if err != nil {
			return zero, err
		}
		return any(HashValue{h: h}).(T), nil
	case CodeValue:
		h, err := c.adoptHandle(x.h)
		if err != nil {
			return zero, err
		}
		return any(CodeValue{h: h}).(T), nil
	case GlobValue:
		h, err := c.adoptHandle(x.h)
		if err != nil {
			return zero, err
		}
		return any(GlobValue{h: h}).(T), nil
	case IOValue:
		h, err := c.adoptHandle(x.h)
		if err != nil {
			return zero, err
		}
		return any(IOValue{h: h}).(T), nil
	default:
		return zero, fmt.Errorf("perl: cannot adopt %T", val)
	}
}

func (c *Perl) adoptHandle(h *refHandle) (*refHandle, error) {
	if h == nil || h.released.Load() {
		return nil, errors.New("perl: Adopt of a released reference")
	}
	if c.raw.Closed() {
		return nil, internal.ErrClosed
	}
	return newRefHandle(c, h.id, h.refkind, h.class), nil
}

// numify converts a string to a number the way Perl does: skip leading
// whitespace, then read the longest prefix that parses as a decimal number
// (sign, digits, fraction, exponent); anything else contributes 0. This is
// a total scan — every input maps to a defined value.
func numify(s string) float64 {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r' || s[i] == '\f') {
		i++
	}
	start := i
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	digits := false
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
		digits = true
	}
	if i < len(s) && s[i] == '.' {
		j := i + 1
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
			digits = true
		}
		if j > i+1 || digits {
			i = j
		}
	}
	if !digits {
		return 0
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		j := i + 1
		if j < len(s) && (s[j] == '+' || s[j] == '-') {
			j++
		}
		expDigits := false
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
			expDigits = true
		}
		if expDigits {
			i = j
		}
	}
	f, err := strconv.ParseFloat(s[start:i], 64)
	if err != nil {
		return 0
	}
	return f
}
