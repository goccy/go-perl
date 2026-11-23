package internal

// Low-level plumbing for host-native XSUBs (the native XS SDK) and the
// per-instance callback dispatcher. The higher-level loader — dlopen,
// vtable construction, frame dispatch — lives in the xs subpackage; this
// file exposes exactly the surface it needs on an instance.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync/atomic"
	"unsafe"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// nativeXSMethodID is the reserved callback method id the guest's generic
// native thunk dispatches on (perl.cc GOPERL_NATIVE_METHOD_ID). Bound Go
// functions use positive ids, so the spaces cannot collide.
const nativeXSMethodID = -1

// The requests of the reserved callback method ids below and the responses
// to them are made of 64-bit words, an address of the guest's memory one of
// them, whole (see perl-wasm's perl.cc).

// magicFreeMethodID is the reserved callback method id fired when a guest SV
// carrying the SDK's anchor magic is freed (perl.cc GOPERL_MG_FREE_METHOD_ID);
// the request is [the host magic id].
const magicFreeMethodID = -2

// ppHookMethodID is the reserved callback method id the guest run loop
// dispatches on for op types a native module claimed via the SDK's
// PL_ppaddr proxy (perl.cc GOPERL_PP_HOOK_METHOD_ID).
const ppHookMethodID = -3

// destructorMethodID is the reserved callback method id fired when a guest
// scope pops a save-stack destructor registered through the SDK (perl.cc
// GOPERL_DTOR_METHOD_ID); the request is [the host destructor id].
const destructorMethodID = -4

// magicLocalMethodID is the reserved callback method id fired when `local`
// puts a new SV in the place of one that carries the anchor magic (perl.cc
// GOPERL_MG_LOCAL_METHOD_ID): the request is [the anchor's host id][the
// new SV].
const magicLocalMethodID = -8

// perlioMethodID is the reserved callback method id the guest PerlIO proxy
// layer dispatches on for the layer slots a native module customized
// (perl.cc GOPERL_PERLIO_METHOD_ID).
const perlioMethodID = -6

// hostCallMethodID is the reserved callback method id of the calls of
// native code that share one channel (op checkers, peephole hooks,
// functions registered with call_atexit).
const hostCallMethodID = -9

// processMemoryMethodID is the reserved callback method id of reading
// memory of the process outside the guest's (perl.cc
// GOPERL_PROCMEM_METHOD_ID): the request is [address][length, or -1 for the
// string up to its NUL], the response [1] and the bytes, or [0] and a
// message to croak with.
const processMemoryMethodID = -10

// keywordMethodID is the reserved callback method id the guest keyword and
// infix plugin wrappers dispatch on while compiling (perl.cc
// GOPERL_KEYWORD_METHOD_ID).
const keywordMethodID = -7

// sockaddrMethodID is the reserved callback method id for turning a socket
// address of the guest's C library into the host's and back (perl.cc
// GOPERL_SOCKADDR_METHOD_ID; hostsockaddr.go).
const sockaddrMethodID = -11

// SetNativeXSHandler installs the dispatcher for native XSUB calls. req is
// the thunk's raw request ([fn_id][cv][items][svs...]) and the returned
// bytes are the raw response ([1][nret][svs...] on success, [0]+message on
// failure). Installing a
// handler registers the instance's callback dispatcher with the guest if
// that has not happened yet.
func (p *Perl) SetNativeXSHandler(fn func(req []byte) []byte) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	p.hookMu.Lock()
	p.nativeXS = fn
	p.hookMu.Unlock()
	return nil
}

// SetMagicFreeHandler installs the teardown hook for host-side MAGIC: when
// the guest frees an SV carrying the native SDK's anchor magic, id (the one
// the loader stored at attach time) is delivered here so the loader can run
// the native module's svt_free chain and release its mirrors. The hook's
// result is the response: [1], or [0]+message for the free to die with.
func (p *Perl) SetMagicFreeHandler(fn func(id uint32) []byte) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	p.hookMu.Lock()
	p.magicFree = fn
	p.hookMu.Unlock()
	return nil
}

// SetMagicLocalHandler installs the hook for `local` on an SV with
// host-side magic: id is the anchor of the variable, sv the token of the SV
// that takes its place for the scope and is to get the magic `local` gives.
// The hook's result is the response: [1], or [0]+message for `local` to die
// with.
func (p *Perl) SetMagicLocalHandler(fn func(id uint32, sv uint64) []byte) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	p.hookMu.Lock()
	p.magicLocal = fn
	p.hookMu.Unlock()
	return nil
}

// SetPPHookHandler installs the dispatcher for pp hooks: op types a native
// module claimed by writing the SDK's PL_ppaddr proxy. req is the guest's
// raw request ([op][op_type] and the perl stack) and the returned bytes are
// the raw response ([1][next_op] and what the function made of the stack on
// success, [0]+message to croak in the interpreter).
func (p *Perl) SetPPHookHandler(fn func(req []byte) []byte) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	p.hookMu.Lock()
	p.ppHook = fn
	p.hookMu.Unlock()
	return nil
}

// SetPerlIOHandler installs the dispatcher for PerlIO layer hooks: slots a
// native module customized in a layer it registered through the SDK's
// PerlIO_define_layer. req is the guest's raw request ([hook][funcs_id][f]
// [layer]+hook extras) and the returned bytes are the raw response
// ([1]+hook words on success, [0]+message to croak).
func (p *Perl) SetPerlIOHandler(fn func(req []byte) []byte) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	p.hookMu.Lock()
	p.perlioHook = fn
	p.hookMu.Unlock()
	return nil
}

// SetKeywordHandler installs the dispatcher for keyword/infix plugin
// forwarding: the guest parser offers candidate words to the host chain a
// native module registered through the SDK's wrap_keyword_plugin. req is
// the guest's raw request ([kind][word bytes]) and the returned bytes are
// the raw response ([1][handled][ret][op] on success, [0]+message to croak
// as a parse error).
func (p *Perl) SetKeywordHandler(fn func(req []byte) []byte) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	p.hookMu.Lock()
	p.keywordHook = fn
	p.hookMu.Unlock()
	return nil
}

// SetHostCallHandler installs the dispatcher of the host-call channel: the
// calls of native code whose request starts with a kind ([kind] and the
// kind's words). The returned bytes are the raw response ([1] and the
// result on success, [0]+message to croak in the interpreter).
func (p *Perl) SetHostCallHandler(fn func(req []byte) []byte) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	p.hookMu.Lock()
	p.hostCall = fn
	p.hookMu.Unlock()
	return nil
}

// SetProcessMemoryHandler installs the reader of the process's memory
// outside the guest's: n bytes at addr, or with n < 0 the bytes up to the
// NUL there.
func (p *Perl) SetProcessMemoryHandler(fn func(addr uint64, n int64) ([]byte, error)) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	p.hookMu.Lock()
	p.processMemory = fn
	p.hookMu.Unlock()
	return nil
}

// SetDestructorHandler installs the hook for save-stack destructors
// registered by native modules: id fires when the guest scope that holds
// it pops. The hook's result is the response: [1] when the destructor
// returned, [0]+message or the thrown-value response when it died, which
// the scope's end dies with in turn.
func (p *Perl) SetDestructorHandler(fn func(id uint32) []byte) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	p.hookMu.Lock()
	p.dtorFire = fn
	p.hookMu.Unlock()
	return nil
}

// RegisterNativeXS binds the guest's generic native thunk as the Perl sub
// name, dispatching to the native function the handler knows as fnID.
func (p *Perl) RegisterNativeXS(name string, fnID int32) error {
	if err := p.EnsureDispatcher(); err != nil {
		return err
	}
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendString(buf, 2, name)
	buf = pbAppendInt32(buf, 3, fnID)
	resp, err := p.invoke(0, midRegisterNativeXS, buf, wasm2go.Inv_0_19)
	if err != nil {
		return err
	}
	return pbExtractError(resp)
}

// XSHelperOp performs one SV micro-operation inside the guest: three
// operands and a string. It is the primitive the native SDK vtable is built
// from; see perl.cc perl_xs_helper for the opcode table. An operation that
// answers two values leaves the second in the operation box (XSOpBox).
//
// The operation may end in a non-local exit of the interpreter instead of a
// result (ErrGuestJump): see guestJump.
func (p *Perl) XSHelperOp(op int32, a, b, c uint64, s string) (uint64, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendInt32(buf, 2, op)
	buf = pbAppendUint64(buf, 3, a)
	buf = pbAppendUint64(buf, 4, b)
	buf = pbAppendUint64(buf, 5, c)
	buf = pbAppendString(buf, 6, s)
	buf = pbAppendUint64(buf, 7, uint64(len(s)))
	// The jump is taken off the module as soon as the operation returns,
	// inside the call: invoke frees the request in the guest after it,
	// which is guest code, and guest code run with the exception pending
	// stops at its first call that may raise one and leaves the rest
	// undone.
	parked := false
	resp, err := p.opInvoke(midXSHelper, buf, func(g *base.Module, req, n wptr) (int64, error) {
		r, err := wasm2go.Inv_0_23(g, req, n)
		parked = p.parkGuestJump()
		return r, err
	})
	if parked {
		return 0, ErrGuestJump
	}
	if err != nil {
		return 0, err
	}
	if e := pbExtractError(resp); e != nil {
		return 0, e
	}
	return readScalarAtField(resp, 1, (*pbReader).readUint64), nil
}

// XSOpBox answers the address of the instance's operation box in linear
// memory (pl.h perl_xs_opbox): native code writes an operation's arguments
// there itself, has XSOpRun run it, and reads its result there, with
// nothing encoded or allocated on the way.
func (p *Perl) XSOpBox() (uint64, error) {
	buf := pbAppendUint64(nil, 1, p.h)
	resp, err := p.invoke(0, midXSOpBox, buf, wasm2go.Inv_0_25)
	if err != nil {
		return 0, err
	}
	if e := pbExtractError(resp); e != nil {
		return 0, e
	}
	return readScalarAtField(resp, 1, (*pbReader).readUint64), nil
}

// XSOpRun runs the operation whose arguments are in the operation box, as
// XSHelperOp runs one; its result is in the box when it returns nil.
func (p *Perl) XSOpRun() error {
	if p.xsOpCall == nil {
		p.xsOpCall = func(g *base.Module, req, n wptr) (int64, error) {
			r, err := wasm2go.Inv_0_24(g, req, n)
			p.xsOpParked = p.parkGuestJump()
			return r, err
		}
	}
	p.xsOpParked = false
	_, err := p.opInvoke(midXSOp, nil, p.xsOpCall)
	if p.xsOpParked {
		p.xsOpParked = false
		return ErrGuestJump
	}
	return err
}

// opInvoke runs an operation of native code (XSHelperOp, XSOpRun). Native
// code makes its operations inside a callback of the instance, from the
// goroutine that runs the callback, which is on the instance's thread: while
// a callback is running the operation runs in place. Otherwise it is an
// entry like any other (invoke), which finds out which thread it is on.
// That costs a system call on Linux (gettid), and native code makes
// thousands of operations where a program makes one entry: every one of
// them paid it, and loading Moose took a sixth longer.
func (p *Perl) opInvoke(mid int32, req []byte, call invoker) ([]byte, error) {
	if p.callbackDepth > 0 {
		return p.m.invoke(0, mid, req, call)
	}
	return p.invoke(0, mid, req, call)
}

// RawMemory exposes the instance's linear memory. Native XSUBs read guest
// strings in place through pointers computed against this slice's base, so
// it must not be captured across calls that could grow the memory; on the
// copy-on-write snapshot path the mapping (and therefore the base address)
// is fixed for the instance's lifetime.
// HostDescriptor returns the descriptor of the host process behind the
// guest's descriptor fd; ok is false when fd is not open or is not backed
// by a descriptor of the host (a file of an in-memory filesystem, a
// standard stream supplied as a reader or writer).
func (p *Perl) HostDescriptor(fd int32) (uintptr, bool) {
	return p.wasi.HostDescriptor(fd)
}

// BeginOutside registers interrupt as what a signal for the program does
// while the instance runs native code (the runtime's
// WasiStubs.BeginOutside); the returned function ends the call.
func (p *Perl) BeginOutside(interrupt func()) (end func()) {
	return p.wasi.BeginOutside(p.m.g, interrupt).End
}

func (p *Perl) RawMemory() []byte {
	return wasm2go.Memory(p.m.g)
}

// MemorySize returns the size of the instance's linear memory in bytes: what
// the guest has grown it to, which is its peak, since a linear memory never
// shrinks. (RawMemory is that long; its capacity is the range reserved for
// the memory, whose pages past the size are not accessible.)
func (p *Perl) MemorySize() uint64 {
	return p.m.g.MemSize.Load()
}

// MemoryCeiling returns how far the instance's memory may grow (see
// memoryCeiling): every address of the memory is below it.
func (p *Perl) MemoryCeiling() uint64 {
	return uint64(memoryCeiling(p.opts, p.m.g.MemSize.Load()))
}

// MemorySizeWord returns where the instance keeps MemorySize: the word
// memory.grow stores the new size in, so a reader of it has the size as it
// is now, for the instance's whole life. Native code bounds the addresses
// of the interpreter's it turns into its own with it, since the memory
// grows while the interpreter runs, during a call of native code too.
func (p *Perl) MemorySizeWord() *uint64 {
	return (*uint64)(unsafe.Pointer(p.m.g.MemSize))
}

// An atomic.Uint64 is its 64-bit value and nothing else, so a pointer to
// one is a pointer to the value (MemorySizeWord).
var _ = [1]struct{}{}[unsafe.Sizeof(atomic.Uint64{})-8]

// EnsureDispatcher registers this instance's callback handler with the
// generated bridge machinery and tells the guest its callback id. Runs once
// per instance.
func (p *Perl) EnsureDispatcher() error {
	p.hookMu.Lock()
	if p.dispatcherSet {
		p.hookMu.Unlock()
		return nil
	}
	p.hookMu.Unlock()

	// Register on this instance's Module (NOT a process-global registry —
	// callback registries are per-module host state).
	p.m.cbMu.Lock()
	if p.m.callbacks == nil {
		p.m.callbacks = map[int32]CallbackHandler{}
	}
	p.m.nextCBID++
	cbID := p.m.nextCBID
	p.m.callbacks[cbID] = dispatcher{p: p}
	p.m.cbMu.Unlock()

	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendInt32(buf, 2, cbID)
	resp, err := p.invoke(0, midSetGoDispatcher, buf, wasm2go.Inv_0_22)
	if err != nil {
		return fmt.Errorf("set Go dispatcher: %w", err)
	}
	if e := pbExtractError(resp); e != nil {
		return fmt.Errorf("set Go dispatcher: %w", e)
	}

	p.hookMu.Lock()
	p.dispatcherSet = true
	p.hookMu.Unlock()
	return nil
}

// DispatcherSet reports whether the instance's callback dispatcher has been
// registered with the guest (Clone uses it to mirror the parent's state).
func (p *Perl) DispatcherSet() bool {
	p.hookMu.RLock()
	defer p.hookMu.RUnlock()
	return p.dispatcherSet
}

// ErrGuestJump is the result of an operation that the interpreter left by
// a non-local exit: it died (or exited) inside the operation, and the place
// that catches that is further out than the host code that asked for the
// operation.
var ErrGuestJump = errors.New("the interpreter is leaving the operation by a non-local exit")

// guestJump is a non-local exit of the interpreter (the longjmp of a die
// or an exit) on its way out through host code.
//
// The interpreter calls host code (a native XSUB, a keyword plugin, a
// magic hook), and that code asks the interpreter for operations, each a
// call into the guest of its own. When the interpreter dies inside such an
// operation and nothing inside the operation catches it, the jump has to
// get to a catch point of an activation of the guest that is below the
// host code on the stack. A jump in the guest is an exception the module
// keeps as pending while every function on the way returns; the operation
// therefore returns to the host with the exception still pending, and the
// guest would go on unwinding only when the host code has returned as
// well.
//
// The host code must not go on as if the operation had answered: on perl
// the jump would have passed through its frames. So the operation reports
// ErrGuestJump, the native code is unwound to its entry, and the jump goes
// on when the callback that ran it returns to the guest. Until then the
// exception is kept here, off the module: while it is pending there, no
// call into the guest does anything, and the native code that is being
// unwound still makes some (it leaves its scopes, frees what it holds).
//
// depth is the callback activation the jump is to resume from: the one
// that was running when it was parked. It cannot be derived: a callback
// that the unwinding itself causes (a destructor) is deeper and ends
// first.
type guestJump struct {
	parked bool
	depth  int
	tag    uint32
	vals   [1]uint64
}

// parkGuestJump takes a pending exception of the guest off the module and
// keeps it for resumeGuestJump; it reports whether there was one. Outside
// of a callback there is nothing to resume it for, and it is left alone.
func (p *Perl) parkGuestJump() bool {
	g := p.m.g
	if g.ExcPending == 0 || p.callbackDepth == 0 {
		return false
	}
	p.jump = guestJump{parked: true, depth: p.callbackDepth, tag: g.ExcTag, vals: g.ExcVals}
	g.ExcPending = 0
	return true
}

// resumeGuestJump puts a parked exception back on the module when the
// callback activation it was parked in returns to the guest, which then
// goes on unwinding; it reports whether it did.
func (p *Perl) resumeGuestJump() bool {
	if !p.jump.parked || p.jump.depth != p.callbackDepth {
		return false
	}
	g := p.m.g
	g.ExcTag, g.ExcVals = p.jump.tag, p.jump.vals
	g.ExcPending = 1
	p.jump = guestJump{}
	return true
}

// GuestJumpCaught drops the jump parked in the callback activation that is
// running: the native code it ran caught the jump itself and returned as
// usual (see the xs package's jumpCaught), so the guest does not go on
// unwinding when the callback returns.
func (p *Perl) GuestJumpCaught() {
	if p.jump.parked && p.jump.depth == p.callbackDepth {
		p.jump = guestJump{}
	}
}

// dispatcher is the per-instance CallbackHandler behind every guest->host
// call. Reserved (negative) method ids route to the native-XS hook surface;
// everything else is the Perl->Go function bridge, which the public package
// serves through UserHandler.
type dispatcher struct{ p *Perl }

func (d dispatcher) HandleCallback(methodID int32, req []byte) (resp []byte, err error) {
	d.p.callbackDepth++
	defer func() {
		if d.p.resumeGuestJump() {
			// The callback ends by the jump: the guest returns at once,
			// with nothing to read (errCallbackJumped).
			resp, err = nil, errCallbackJumped
		}
		d.p.callbackDepth--
	}()
	return d.handle(methodID, req)
}

func (d dispatcher) handle(methodID int32, req []byte) ([]byte, error) {
	if methodID == nativeXSMethodID {
		d.p.hookMu.RLock()
		native := d.p.nativeXS
		d.p.hookMu.RUnlock()
		if native == nil {
			return append([]byte{0}, "no native XS handler installed"...), nil
		}
		return native(req), nil
	}
	if methodID == magicFreeMethodID {
		d.p.hookMu.RLock()
		free := d.p.magicFree
		d.p.hookMu.RUnlock()
		if free != nil && len(req) >= 8 {
			return free(uint32(binary.LittleEndian.Uint64(req))), nil
		}
		return []byte{1}, nil
	}
	if methodID == magicLocalMethodID {
		d.p.hookMu.RLock()
		local := d.p.magicLocal
		d.p.hookMu.RUnlock()
		if local != nil && len(req) >= 16 {
			return local(uint32(binary.LittleEndian.Uint64(req)), binary.LittleEndian.Uint64(req[8:])), nil
		}
		return []byte{1}, nil
	}
	if methodID == ppHookMethodID {
		d.p.hookMu.RLock()
		hook := d.p.ppHook
		d.p.hookMu.RUnlock()
		if hook == nil {
			return append([]byte{0}, "no pp hook handler installed"...), nil
		}
		return hook(req), nil
	}
	if methodID == keywordMethodID {
		d.p.hookMu.RLock()
		hook := d.p.keywordHook
		d.p.hookMu.RUnlock()
		if hook == nil {
			// no parse-surface module loaded: decline ([1][handled 0]) so
			// the guest chain runs
			return []byte{1, 0, 0, 0, 0, 0, 0, 0, 0}, nil
		}
		return hook(req), nil
	}
	if methodID == perlioMethodID {
		d.p.hookMu.RLock()
		hook := d.p.perlioHook
		d.p.hookMu.RUnlock()
		if hook == nil {
			return append([]byte{0}, "no PerlIO layer handler installed"...), nil
		}
		return hook(req), nil
	}
	if methodID == processMemoryMethodID {
		d.p.hookMu.RLock()
		read := d.p.processMemory
		d.p.hookMu.RUnlock()
		if read == nil || len(req) < 16 {
			return append([]byte{0}, "no reader of the memory of the process"...), nil
		}
		b, err := read(binary.LittleEndian.Uint64(req), int64(binary.LittleEndian.Uint64(req[8:])))
		if err != nil {
			return append([]byte{0}, err.Error()...), nil
		}
		return append([]byte{1}, b...), nil
	}
	if methodID == hostCallMethodID {
		d.p.hookMu.RLock()
		call := d.p.hostCall
		d.p.hookMu.RUnlock()
		if call == nil {
			return append([]byte{0}, "no host call handler installed"...), nil
		}
		return call(req), nil
	}
	if methodID == sockaddrMethodID {
		return hostSockaddrAnswer(req), nil
	}
	if methodID == destructorMethodID {
		d.p.hookMu.RLock()
		fire := d.p.dtorFire
		d.p.hookMu.RUnlock()
		if fire != nil && len(req) >= 8 {
			return fire(uint32(binary.LittleEndian.Uint64(req))), nil
		}
		return []byte{1}, nil
	}
	if d.p.UserHandler == nil {
		return nil, fmt.Errorf("no handler bound for callback method id %d", methodID)
	}
	return d.p.callGo(methodID, req)
}
