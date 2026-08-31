package perl

// The host half of the typed value protocol (see perl-wasm's pl.h for the
// authoritative wire description): little-endian binary nodes for values,
// result envelopes for operations. Nothing is stringified in transit —
// scalars cross by value with their kind (byte strings raw), references
// cross by handle.

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/goccy/go-perl/internal"
)

// Node tags.
const (
	wireUndef    = 0
	wireBool     = 1
	wireInt      = 2
	wireFloat    = 3
	wireString   = 4
	wireRef      = 5
	wireHostFunc = 6
	wireFlatten  = 7
)

// Reference kinds (what the reference's referent is).
const (
	wireRefScalar = 0
	wireRefArray  = 1
	wireRefHash   = 2
	wireRefCode   = 3
	wireRefGlob   = 4
	wireRefIO     = 5
	wireRefFormat = 6
	wireRefRegexp = 7
	wireRefOther  = 8
)

// Envelope statuses.
const (
	wireOK   = 0
	wireDie  = 1
	wireExit = 2
)

func appendU32(b []byte, v uint32) []byte { return binary.LittleEndian.AppendUint32(b, v) }
func appendU64(b []byte, v uint64) []byte { return binary.LittleEndian.AppendUint64(b, v) }

// encodeKey encodes one Go string as a string node (utf8-flagged, matching
// New's string convention; Perl normalises hash keys, so ASCII keys compare
// equal either way).
func encodeKey(key string) []byte {
	b := make([]byte, 0, 6+len(key))
	b = append(b, wireString, 1)
	b = appendU32(b, uint32(len(key)))
	return append(b, key...)
}

// encodeValue appends v's node. In list positions (listPos true — argument
// lists, Push, NewArray) an ArrayValue/HashValue flattens into its contents,
// Perl's own calling convention; in single-value positions it is an error —
// pass Ref() to mean the reference.
func (p *Perl) encodeValue(b []byte, v Value, listPos bool) ([]byte, error) {
	switch x := v.(type) {
	case nil:
		return append(b, wireUndef), nil
	case ScalarValue:
		switch x.kind {
		case KindUndef:
			return append(b, wireUndef), nil
		case KindBool:
			b = append(b, wireBool)
			if x.b {
				return append(b, 1), nil
			}
			return append(b, 0), nil
		case KindInt:
			b = append(b, wireInt)
			return appendU64(b, uint64(x.i)), nil
		case KindFloat:
			b = append(b, wireFloat)
			return appendU64(b, math.Float64bits(x.f)), nil
		default:
			b = append(b, wireString)
			if x.utf8 {
				b = append(b, 1)
			} else {
				b = append(b, 0)
			}
			b = appendU32(b, uint32(len(x.s)))
			return append(b, x.s...), nil
		}
	case RefValue:
		return p.encodeHandle(b, x.h, wireRef)
	case CodeValue:
		return p.encodeHandle(b, x.h, wireRef)
	case GlobValue:
		return p.encodeHandle(b, x.h, wireRef)
	case IOValue:
		return p.encodeHandle(b, x.h, wireRef)
	case ArrayValue:
		if !listPos {
			return nil, fmt.Errorf("perl: an array cannot fill a single-value slot; pass Ref() for the reference")
		}
		return p.encodeHandle(b, x.h, wireFlatten)
	case HashValue:
		if !listPos {
			return nil, fmt.Errorf("perl: a hash cannot fill a single-value slot; pass Ref() for the reference")
		}
		return p.encodeHandle(b, x.h, wireFlatten)
	default:
		return nil, fmt.Errorf("perl: cannot encode %T", v)
	}
}

// encodeHandle appends a handle-bearing node (a ref node or a flatten node).
func (p *Perl) encodeHandle(b []byte, h *refHandle, tag byte) ([]byte, error) {
	if h == nil || h.released.Load() {
		return nil, fmt.Errorf("perl: value's reference has been released")
	}
	if h.p != p {
		return nil, fmt.Errorf("perl: value belongs to a different Perl instance")
	}
	b = append(b, tag)
	b = appendU64(b, h.id)
	if tag == wireRef {
		b = append(b, h.refkind)
		b = appendU16(b, 0) // class travels guest->host only
	}
	return b, nil
}

func appendU16(b []byte, v uint16) []byte { return binary.LittleEndian.AppendUint16(b, v) }

// encodeArgs encodes a list-position node list ([u32 count] nodes).
func (p *Perl) encodeArgs(args []Value) ([]byte, error) {
	b := appendU32(nil, uint32(len(args)))
	var err error
	for i, a := range args {
		b, err = p.encodeValue(b, a, true)
		if err != nil {
			return nil, fmt.Errorf("argument %d: %w", i, err)
		}
	}
	return b, nil
}

// encodeSingle encodes one single-position node.
func (p *Perl) encodeSingle(v Value) ([]byte, error) {
	return p.encodeValue(nil, v, false)
}

// nodeReader walks a response buffer with bounds checking; any overrun
// flips fail so decoding is total.
type nodeReader struct {
	b    []byte
	off  int
	fail bool
}

func (r *nodeReader) need(n int) bool {
	if r.fail || len(r.b)-r.off < n {
		r.fail = true
		return false
	}
	return true
}

func (r *nodeReader) u8() byte {
	if !r.need(1) {
		return 0
	}
	v := r.b[r.off]
	r.off++
	return v
}

func (r *nodeReader) u16() uint16 {
	if !r.need(2) {
		return 0
	}
	v := binary.LittleEndian.Uint16(r.b[r.off:])
	r.off += 2
	return v
}

func (r *nodeReader) u32() uint32 {
	if !r.need(4) {
		return 0
	}
	v := binary.LittleEndian.Uint32(r.b[r.off:])
	r.off += 4
	return v
}

func (r *nodeReader) u64() uint64 {
	if !r.need(8) {
		return 0
	}
	v := binary.LittleEndian.Uint64(r.b[r.off:])
	r.off += 8
	return v
}

func (r *nodeReader) bytes(n int) []byte {
	if !r.need(n) {
		return nil
	}
	v := r.b[r.off : r.off+n]
	r.off += n
	return v
}

// lenBytes reads a u32-length-prefixed byte string.
func (r *nodeReader) lenBytes() []byte {
	n := r.u32()
	return r.bytes(int(n))
}

// decodeNode reads one node into a Value. References surface as RefValue
// handles owning one registry pin each.
func (p *Perl) decodeNode(r *nodeReader) (Value, error) {
	tag := r.u8()
	if r.fail {
		return nil, fmt.Errorf("perl: malformed value node")
	}
	switch tag {
	case wireUndef:
		return Undef(), nil
	case wireBool:
		return ScalarValue{kind: KindBool, b: r.u8() != 0}, nil
	case wireInt:
		return ScalarValue{kind: KindInt, i: int64(r.u64())}, nil
	case wireFloat:
		return ScalarValue{kind: KindFloat, f: math.Float64frombits(r.u64())}, nil
	case wireString:
		utf8 := r.u8() != 0
		raw := r.lenBytes()
		if r.fail {
			return nil, fmt.Errorf("perl: malformed string node")
		}
		s := make([]byte, len(raw))
		copy(s, raw)
		return ScalarValue{kind: KindString, s: s, utf8: utf8}, nil
	case wireRef:
		id := r.u64()
		refkind := r.u8()
		class := string(r.bytes(int(r.u16())))
		if r.fail {
			return nil, fmt.Errorf("perl: malformed ref node")
		}
		return RefValue{h: newRefHandle(p, id, refkind, class)}, nil
	default:
		return nil, fmt.Errorf("perl: unknown value node tag %d", tag)
	}
}

// decodeEnvelope consumes the status byte; a die becomes *PerlError (or the
// cancellation when the interrupt watchdog fired), a caught guest exit()
// becomes the error ExitCode recognises. The reader is positioned at the ok
// payload on nil error.
func decodeEnvelope(ctx context.Context, r *nodeReader, interrupted bool) error {
	switch r.u8() {
	case wireOK:
		return nil
	case wireDie:
		msg := r.lenBytes()
		if interrupted {
			return ctx.Err()
		}
		return &PerlError{Message: string(msg)}
	case wireExit:
		return &internal.ExitError{Code: int(int32(r.u32()))}
	default:
		return fmt.Errorf("perl: malformed result envelope")
	}
}

// decodeListResult decodes an ok envelope carrying a node list.
func (p *Perl) decodeListResult(ctx context.Context, resp []byte, interrupted bool) ([]Value, error) {
	r := &nodeReader{b: resp}
	if err := decodeEnvelope(ctx, r, interrupted); err != nil {
		return nil, err
	}
	count := int(r.u32())
	out := make([]Value, 0, count)
	for i := 0; i < count; i++ {
		v, err := p.decodeNode(r)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// decodeNodeResult decodes an ok envelope carrying one node.
func (p *Perl) decodeNodeResult(ctx context.Context, resp []byte, interrupted bool) (Value, error) {
	r := &nodeReader{b: resp}
	if err := decodeEnvelope(ctx, r, interrupted); err != nil {
		return nil, err
	}
	return p.decodeNode(r)
}

// decodeI64Result decodes an ok envelope carrying one i64.
func (p *Perl) decodeI64Result(ctx context.Context, resp []byte, interrupted bool) (int64, error) {
	r := &nodeReader{b: resp}
	if err := decodeEnvelope(ctx, r, interrupted); err != nil {
		return 0, err
	}
	v := int64(r.u64())
	if r.fail {
		return 0, fmt.Errorf("perl: malformed result envelope")
	}
	return v, nil
}

// decodeEmptyResult decodes an ok envelope with no payload.
func (p *Perl) decodeEmptyResult(ctx context.Context, resp []byte, interrupted bool) error {
	r := &nodeReader{b: resp}
	return decodeEnvelope(ctx, r, interrupted)
}

// decodeExistsNodeResult decodes an ok envelope carrying u8 exists + node.
func (p *Perl) decodeExistsNodeResult(ctx context.Context, resp []byte, interrupted bool) (Value, bool, error) {
	r := &nodeReader{b: resp}
	if err := decodeEnvelope(ctx, r, interrupted); err != nil {
		return nil, false, err
	}
	exists := r.u8() != 0
	v, err := p.decodeNode(r)
	if err != nil {
		return nil, false, err
	}
	return v, exists, nil
}

// decodeEvalResult decodes perl_eval's envelope: the result node (on ok) or
// $@ (on die), followed either way by the captured stdout/stderr.
func (p *Perl) decodeEvalResult(ctx context.Context, resp []byte, interrupted bool) (Result, error) {
	r := &nodeReader{b: resp}
	res := Result{Value: Undef()}
	switch r.u8() {
	case wireOK:
		v, err := p.decodeNode(r)
		if err != nil {
			return Result{}, err
		}
		res.Value = v
	case wireDie:
		msg := r.lenBytes()
		if interrupted {
			return Result{}, ctx.Err()
		}
		res.Error = &PerlError{Message: string(msg)}
	default:
		return Result{}, fmt.Errorf("perl: malformed eval envelope")
	}
	res.Stdout = string(r.lenBytes())
	res.Stderr = string(r.lenBytes())
	if r.fail {
		return Result{}, fmt.Errorf("perl: malformed eval envelope")
	}
	return res, nil
}
