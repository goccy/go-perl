//go:build darwin || linux

package xs

// Native code in a fork child.
//
// A fork makes a pseudo-process: a copy of the instance's memory that runs
// in this same process (see internal/fork.go). The interpreter's memory is
// copied; the memory of native code is not, and cannot be: the C heap, the
// modules' statics and the state of the libraries they use belong to the
// process, so parent and child would share every structure native code
// keeps (a database handle, an event loop, a cache's lock state), each
// changing and freeing what the other still uses. A real fork gives each
// process its own copy; this one cannot.
//
// So native code is not available in a fork child. Every way the child's
// interpreter enters native code ends in the child's handlers below,
// installed once, when the child is made (adoptForkChild): a call of an
// XSUB, a pp function, a hook of the compiler or of magic dies with a Perl
// error that names the code and says why; a destructor (DESTROY, svt_free,
// a save-stack destructor, a layer's Popped, a call_atexit function) is not
// run, so that the child frees nothing of the parent's; and in the child's
// teardown (END blocks, global destruction) a call is skipped instead of
// dying, where skipping it leaves the program meaning what it means. A hook
// that decides what code means (a keyword plugin, an op checker, a pp
// function, the run loop) dies there too. dl_load_file refuses to load a
// library in the child.
//
// Native threads and signal handlers belong to the process. A thread a
// module started keeps running for the parent; nothing of the child's runs
// on it. A signal sent to a pseudo-process is posted to its interpreter,
// never raised in the process, so no native handler runs for a child.

import (
	"encoding/binary"
	"fmt"

	"github.com/goccy/go-perl/internal"
	"github.com/goccy/go-perl/internal/ccall"
)

// The interpreter's state a fork child's handlers decide by: PL_phase
// (GOPERL_PL_PHASE, through GOPERL_OP_PLVAR_GET) and the type of the stack
// the interpreter runs on (GOPERL_SI_TYPE of PL_curstackinfo, through
// GOPERL_OP_SI_GET); and the name of an op type, for a message
// (GOPERL_OP_OP_NAME_STR).
const (
	opPlvarGet    = 73
	opSIGet       = 112
	opOpNameStr   = 115
	plPhase       = 36
	siType        = 2
	phaseEnd      = 5 // PERL_PHASE_END
	phaseDestruct = 6 // PERL_PHASE_DESTRUCT
	siDestroy     = 6 // PERLSI_DESTROY: the stack a destructor runs on
)

// The kinds of the host-call channel (GOPERL_HOST_* in perl-wasm's perl.cc).
const (
	hostNewOp  = 0
	hostPeep   = 1
	hostAtexit = 2
	hostCallCk = 3
	hostMagic  = 4
	hostUvar   = 5
	hostFilter = 6
	hostRunops = 7
	hostBhk    = 8
	hostInfix  = 9
)

// The hooks of the PerlIO proxy layer (GOPERLIO_HOOK_* in perl.cc).
const (
	perlioPushed = 1
	perlioFill   = 2
	perlioPopped = 3
)

// adoptForkChild is adoptClone for a fork child: the child gets a loader
// state that runs no native code. It keeps what names the code the child's
// interpreter may try to enter (the functions of the subs, of the pp
// hooks, the layers); nothing native code keeps is copied, and nothing of
// the modules' is told about the child.
func (s *state) adoptForkChild(child *internal.Perl) error {
	s.mu.Lock()
	cs := &state{
		p:            child,
		libs:         map[uintptr]struct{}{},
		syms:         map[uintptr]struct{}{},
		fns:          append([]uintptr(nil), s.fns...),
		fnNames:      append([][]byte(nil), s.fnNames...),
		cvAny:        map[uint64]uintptr{},
		pvHost:       map[uint64]uintptr{},
		perlioFuncs:  make(map[uint32]uintptr, len(s.perlioFuncs)),
		cvFn:         map[uint64]uintptr{},
		magicBySV:    map[uint64]*magicRec{},
		magicByID:    map[uint32]*magicRec{},
		hookFns:      make(map[int32]uintptr, len(s.hookFns)),
		perOpFns:     make(map[uint64]uintptr, len(s.perOpFns)),
		opbox:        s.opbox,
		cloneHookSet: true,
	}
	for k, v := range s.perlioFuncs {
		cs.perlioFuncs[k] = v
	}
	for k, v := range s.hookFns {
		cs.hookFns[k] = v
	}
	for k, v := range s.perOpFns {
		cs.perOpFns[k] = v
	}
	s.mu.Unlock()
	statesMu.Lock()
	states[child] = cs
	statesMu.Unlock()

	child.AddCloneHook(cs.adoptForkChild)
	if err := child.SetNativeXSHandler(cs.childXSUB); err != nil {
		return err
	}
	if err := child.SetMagicFreeHandler(func(uint32) []byte { return []byte{1} }); err != nil {
		return err
	}
	if err := child.SetMagicLocalHandler(cs.childMagicLocal); err != nil {
		return err
	}
	if err := child.SetPPHookHandler(cs.childPPHook); err != nil {
		return err
	}
	if err := child.SetKeywordHandler(cs.childKeyword); err != nil {
		return err
	}
	if err := child.SetPerlIOHandler(cs.childPerlIO); err != nil {
		return err
	}
	if err := child.SetHostCallHandler(cs.childHostCall); err != nil {
		return err
	}
	if err := child.SetProcessMemoryHandler(childProcessMemory); err != nil {
		return err
	}
	return child.SetDestructorHandler(func(uint32) []byte { return []byte{1} })
}

// unavailable is the message a fork child's interpreter dies with when it
// was to enter native code: what it was to run and, where the code's
// function is known, the library the function is in.
func unavailable(what string, fn uintptr) string {
	msg := what + ": native modules cannot be used in a forked child"
	if lib := libraryOf(fn); lib != "" {
		msg += " (" + lib + ")"
	}
	return msg
}

// refuse is the response that makes the interpreter die with
// unavailable's message.
func refuse(what string, fn uintptr) []byte {
	return append([]byte{0}, unavailable(what, fn)...)
}

// tearingDown reports whether the child's interpreter is being destroyed:
// its END blocks or its global destruction run.
func (s *state) tearingDown() bool {
	phase, err := s.p.XSHelperOp(opPlvarGet, plPhase, 0, 0, "")
	return err == nil && (phase == phaseEnd || phase == phaseDestruct)
}

// inDestructor reports whether the interpreter runs a destructor: DESTROY
// of an object, and what it calls (perl runs it on a stack of its own).
func (s *state) inDestructor() bool {
	t, err := s.p.XSHelperOp(opSIGet, 0, siType, 0, "")
	return err == nil && int64(t) == siDestroy
}

// childXSUB answers a call of an XSUB in a fork child (method -1): a
// destructor, and any call in the child's teardown, is not run and returns
// nothing; any other call dies.
func (s *state) childXSUB(req []byte) []byte {
	if len(req) < 3*8 {
		return append([]byte{0}, "native XS dispatch: short payload"...)
	}
	if s.inDestructor() || s.tearingDown() {
		return []byte{1, 0, 0, 0, 0, 0, 0, 0, 0} // [1][no values]
	}
	id := int64(wordAt(req, 0))
	what, fn := "a native sub", uintptr(0)
	s.mu.Lock()
	if id >= 0 && id < int64(len(s.fns)) {
		fn = s.fns[id]
		if name := s.fnNames[id]; len(name) > 1 {
			what = string(name[:len(name)-1]) // NUL-terminated
		}
	}
	s.mu.Unlock()
	return refuse(what, fn)
}

// childMagicLocal answers `local` on a variable with magic of native code
// (method -8): the magic `local` gives the new value is native code's to
// make, so `local` dies, or, in the teardown, gives none.
func (s *state) childMagicLocal(uint32, uint64) []byte {
	if s.tearingDown() {
		return []byte{1}
	}
	return refuse("local on a variable with magic of a native module", 0)
}

// childPPHook answers the run of an op whose pp function is native code
// (method -3): it dies, in the teardown too, since skipping the op would
// change what the program does.
func (s *state) childPPHook(req []byte) []byte {
	if len(req) < 2*8 {
		return append([]byte{0}, "pp hook: short payload"...)
	}
	opTok, opType := wordAt(req, 0), int32(int64(wordAt(req, 1)))
	s.mu.Lock()
	fn := s.perOpFns[opTok]
	if fn == 0 {
		fn = s.hookFns[opType]
	}
	s.mu.Unlock()
	return refuse(fmt.Sprintf("an op (%s) of a native module", s.opName(opType)), fn)
}

// opName is the interpreter's name of an op type (PL_op_name: "custom" for
// a custom op), or its number when that is not to be had.
func (s *state) opName(opType int32) string {
	at, err := s.p.XSHelperOp(opOpNameStr, uint64(uint32(opType)), 0, 0, "")
	mem := s.p.RawMemory()
	if err != nil || at == 0 || s.opbox == 0 {
		return fmt.Sprint(opType)
	}
	n := opboxResult2(mem, s.opbox)
	if at > uint64(len(mem)) || n > uint64(len(mem))-at {
		return fmt.Sprint(opType)
	}
	return string(mem[at : at+n])
}

// childKeyword answers a word the parser offers the keyword or infix
// plugins of native code (method -7), which the parent registered: what
// the word means is theirs to say, so the compilation dies. (An infix
// plugin is offered the rest of the line, not a word.)
func (s *state) childKeyword(req []byte) []byte {
	if len(req) < 8+1 {
		return append([]byte{0}, "keyword hook: short request"...)
	}
	if wordAt(req, 0) == 1 {
		return refuse("the infix operator plugin of a native module", 0)
	}
	return refuse(fmt.Sprintf("the keyword plugin of a native module (compiling %q)", req[8:]), 0)
}

// childPerlIO answers a slot of a layer of native code (method -6): Popped
// is the layer's destructor and is not run; Pushed and Fill die, or, in
// the teardown, do nothing (Fill then finds the end of the file).
func (s *state) childPerlIO(req []byte) []byte {
	if len(req) < 4*8 {
		return append([]byte{0}, "perlio hook: short request"...)
	}
	hook := wordAt(req, 0)
	if hook == perlioPopped || s.tearingDown() {
		return []byte{1}
	}
	s.mu.Lock()
	funcs := s.perlioFuncs[uint32(wordAt(req, 1))]
	s.mu.Unlock()
	what := "a PerlIO layer of a native module"
	if funcs != 0 {
		// PerlIO_funcs begins with its size and then the layer's name
		if name := cString(uintptr(binary.LittleEndian.Uint64(copyIn(funcs+8, 8)))); name != "" {
			what = fmt.Sprintf("the PerlIO layer :%s", name)
		}
	}
	return refuse(what, funcs)
}

// childHostCall answers a call of native code through the host-call
// channel (method -9). A call_atexit function is not run (it releases what
// the module keeps). Magic hooks die, and in the teardown do nothing. The
// compiler's hooks (op checkers, peephole hooks, call checkers, block
// hooks, infix operators, source filters) and a run loop of native code
// die in the teardown too: what they do is what the code means.
func (s *state) childHostCall(req []byte) []byte {
	if len(req) < 8 {
		return append([]byte{0}, "host call: short request"...)
	}
	var what string
	switch kind := wordAt(req, 0); kind {
	case hostAtexit:
		return []byte{1}
	case hostMagic, hostUvar:
		if s.tearingDown() {
			return []byte{1}
		}
		what = "magic of a native module"
	case hostNewOp:
		what = "the op checker of a native module"
	case hostPeep:
		what = "the peephole optimizer of a native module"
	case hostCallCk:
		what = "the call checker of a native module"
	case hostFilter:
		what = "the source filter of a native module"
	case hostRunops:
		what = "the run loop of a native module"
	case hostBhk:
		what = "the block hook of a native module"
	case hostInfix:
		what = "the infix operator of a native module"
	default:
		what = fmt.Sprintf("native code (host call %d)", kind)
	}
	return refuse(what, 0)
}

// childProcessMemory refuses unpack 'p' and 'P' of an address outside the
// interpreter's memory in a fork child: that memory is the parent's native
// memory, which the parent changes and frees as it goes.
func childProcessMemory(addr uint64, _ int64) ([]byte, error) {
	return nil, fmt.Errorf("%s", unavailable(fmt.Sprintf("memory of native code at %#x", addr), 0))
}

// libraryOf names the shared library the function at fn is in (dladdr's
// file name), or "" when that is not known.
func libraryOf(fn uintptr) string {
	if fn == 0 || dladdrFn == 0 {
		return ""
	}
	// Dl_info: dli_fname, dli_fbase, dli_sname, dli_saddr
	info := cMalloc(32)
	defer cFree(info)
	pokeBytes(info, make([]byte, 32))
	if ok := ccall.Invoke(dladdrFn, fn, info); ok == 0 {
		return ""
	}
	return cString(uintptr(binary.LittleEndian.Uint64(copyIn(info, 8))))
}
