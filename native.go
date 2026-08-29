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

// magicFreeMethodID is the reserved callback method id fired when a guest SV
// carrying the SDK's anchor magic is freed (perl.cc GOPERL_MG_FREE_METHOD_ID);
// the payload is the u32 host magic id.
const magicFreeMethodID = -2

// ppHookMethodID is the reserved callback method id the guest run loop
// dispatches on for op types a native module claimed via the SDK's
// PL_ppaddr proxy (perl.cc GOPERL_PP_HOOK_METHOD_ID).
const ppHookMethodID = -3

// destructorMethodID is the reserved callback method id fired when a guest
// scope pops a save-stack destructor registered through the SDK (perl.cc
// GOPERL_DTOR_METHOD_ID); the payload is the u32 host destructor id.
const destructorMethodID = -4

// SetNativeXSHandler installs the dispatcher for native XSUB calls. req is
// the thunk's raw payload ([u32 fn_id][u32 cv_token][u32 items][u32
// sv_tokens...]) and the returned bytes are the raw response ([1][u32 nret]
// [u32 sv_tokens...] on success, [0]+message on failure). Installing a
// handler registers the instance's callback dispatcher with the guest if
// that has not happened yet.
func (p *Perl) SetNativeXSHandler(fn func(req []byte) []byte) error {
	if err := p.ensureDispatcher(); err != nil {
		return err
	}
	p.funcsMu.Lock()
	p.nativeXS = fn
	p.funcsMu.Unlock()
	return nil
}

// SetMagicFreeHandler installs the teardown hook for host-side MAGIC: when
// the guest frees an SV carrying the native SDK's anchor magic, id (the u32
// the loader stored at attach time) is delivered here so the loader can run
// the native module's svt_free chain and release its mirrors.
func (p *Perl) SetMagicFreeHandler(fn func(id uint32)) error {
	if err := p.ensureDispatcher(); err != nil {
		return err
	}
	p.funcsMu.Lock()
	p.magicFree = fn
	p.funcsMu.Unlock()
	return nil
}

// SetPPHookHandler installs the dispatcher for pp hooks: op types a native
// module claimed by writing the SDK's PL_ppaddr proxy. req is the guest's
// raw payload ([u32 op][u32 op_type][u32 n][u32 stack-top tokens]) and the
// returned bytes are the raw response ([1][u32 next_op] on success,
// [0]+message to croak in the interpreter).
func (p *Perl) SetPPHookHandler(fn func(req []byte) []byte) error {
	if err := p.ensureDispatcher(); err != nil {
		return err
	}
	p.funcsMu.Lock()
	p.ppHook = fn
	p.funcsMu.Unlock()
	return nil
}

// SetDestructorHandler installs the hook for save-stack destructors
// registered by native modules: id fires when the guest scope that holds
// it pops.
func (p *Perl) SetDestructorHandler(fn func(id uint32)) error {
	if err := p.ensureDispatcher(); err != nil {
		return err
	}
	p.funcsMu.Lock()
	p.dtorFire = fn
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