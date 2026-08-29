package perl

// Low-level plumbing for host-native XSUBs (the native XS SDK). The
// higher-level loader — dlopen, vtable construction, frame dispatch — lives
// in the xsnative subpackage; this file exposes exactly the surface it needs
// on a Perl instance. These APIs are building blocks, not application API.

import (
	"fmt"

	wasm2go "github.com/goccy/perlwasm2go"
)

// nativeXSMethodID is the reserved callback method id the guest's generic
// native thunk dispatches on (perl.cc GOPERL_NATIVE_METHOD_ID). Bound Go
// functions use positive ids, so the spaces cannot collide.
const nativeXSMethodID = -1

// SetNativeXSHandler installs the dispatcher for native XSUB calls. req is
// the thunk's raw payload ([u32 fn_id][u32 items][u32 sv_tokens...]) and the
// returned bytes are the raw response ([1][u32 nret][u32 sv_tokens...] on
// success, [0]+message on failure). Installing a handler registers the
// instance's callback dispatcher with the guest if that has not happened yet.
func (p *Perl) SetNativeXSHandler(fn func(req []byte) []byte) error {
	if err := p.ensureDispatcher(); err != nil {
		return err
	}
	p.funcsMu.Lock()
	p.nativeXS = fn
	p.funcsMu.Unlock()
	return nil
}

// RegisterNativeXS binds the guest's generic native thunk as the Perl sub
// name, dispatching to the native function the handler knows as fnID.
func (p *Perl) RegisterNativeXS(name string, fnID int32) error {
	if !perlSubName.MatchString(name) {
		return fmt.Errorf("invalid Perl sub name %q", name)
	}
	if err := p.ensureDispatcher(); err != nil {
		return err
	}
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendString(buf, 2, name)
	buf = pbAppendInt32(buf, 3, fnID)
	resp, err := p.m.invoke(0, midRegisterNativeXS, buf, wasm2go.Inv_0_5)
	if err != nil {
		return err
	}
	return pbExtractError(resp)
}

// XS helper opcodes (perl.cc perl_xs_helper).
const (
	xsOpSvIV     = 1
	xsOpSvPV     = 2
	xsOpNewIV    = 3
	xsOpNewPVN   = 4
	xsOpSvMortal = 5
)

// XSHelperOp performs one SV micro-operation inside the guest. It is the
// primitive the native SDK vtable is built from; see perl.cc perl_xs_helper
// for the opcode table.
func (p *Perl) XSHelperOp(op int32, a, b uint64, s string) (uint64, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendInt32(buf, 2, op)
	buf = pbAppendUint64(buf, 3, a)
	buf = pbAppendUint64(buf, 4, b)
	buf = pbAppendString(buf, 5, s)
	resp, err := p.m.invoke(0, midXSHelper, buf, wasm2go.Inv_0_7)
	if err != nil {
		return 0, err
	}
	if e := pbExtractError(resp); e != nil {
		return 0, e
	}
	return readScalarAtField(resp, 1, (*pbReader).readUint64), nil
}

// RawMemory exposes the instance's linear memory. Native XSUBs read guest
// strings in place through pointers computed against this slice's base, so
// it must not be captured across calls that could grow the memory; on the
// copy-on-write snapshot path the mapping (and therefore the base address)
// is fixed for the instance's lifetime.
func (p *Perl) RawMemory() []byte {
	return wasm2go.Memory(p.m.g)
}