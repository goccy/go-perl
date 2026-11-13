//go:build darwin || linux

// Package xs loads host-native XS modules — shared libraries compiled
// with an ordinary C compiler against go-perl's XS SDK headers (sdk/include)
// — into a running Perl instance.
//
// The model is load-only: building the .so happens ahead of time (a CI step,
// a vendor step, the gperl CLI); at runtime this package only dlopens the
// prebuilt artifact, injects the API vtable (__goperl_xs_init), and runs the
// module's boot function, which registers its XSUBs as ordinary Perl subs.
// Perl then calls them like any XS: the guest's generic thunk forwards each
// call here, this loader invokes the native function, and every SV operation
// the native code performs travels back through the instance's XS helper.
//
// v3 speaks the SDK's pTHX model: every XSUB has perl's real signature
// (void xsub(pTHX_ CV* cv), where pTHX is the call frame), the frame carries
// a real perl stack (base/sp/marks over SV tokens), and the loader
// additionally maintains per-CV storage (CvXSUBANY/alias dispatch) and
// host-side MAGIC mirror chains whose lifetime the guest anchors.
//
// Requirements and caveats:
//   - The .so must be built against THIS SDK (system-perl XS binaries can
//     never load: their macros compiled into direct struct access against a
//     different memory model).
//   - Native code runs outside the wasm sandbox — loading a module is as
//     trusting as cgo.
//   - The instance must be on the copy-on-write snapshot path (the default),
//     whose linear memory never relocates: native code holds raw pointers
//     into it while a call is in flight.
//   - The SDK keeps some module state in process-wide statics; loading one
//     module into many instances in the same process shares those statics.
package xs

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/goccy/go-perl/internal"
)

// xsTrace logs every guest micro-op to stderr (GOPERL_XS_TRACE=1); the
// debugging aid of last resort when a native module misbehaves.
var xsTrace = os.Getenv("GOPERL_XS_TRACE") != ""

// abiVersion must match GOPERL_XS_ABI in sdk/include/perl.h.
const abiVersion = 16

// plErrno is the interpreter variable id of errno (GOPERL_PL_ERRNO).
const plErrno = 60

// failedThrown is the frame's failed value after croak_sv
// (GOPERL_FAILED_THROWN in the SDK), and respThrow the dispatch response
// that tells the interpreter to die with the value it was handed
// (GOPERL_RESP_THROW in perl-wasm).
const (
	failedThrown = 2
	respThrow    = 2
)

const (
	maxStack = 512 // GOPERL_XS_STACK: the slots a frame starts with
	// stackHeadroom is the room above the arguments a frame is given when
	// they do not fit the initial stack: return values an XSUB stores
	// without extending the stack first.
	stackHeadroom = 64
	// sharedBlockSize matches GOPERL_SHARED_RESERVED in the SDK.
	sharedBlockSize = 262144
	maxMarks        = 64 // GOPERL_XS_MARKS
	tmpSlots        = 16 // GOPERL_XS_TMPS
)

// Guest xs-helper ops the loader itself uses (perl.cc perl_xs_helper).
const (
	opAvFetch       = 12
	opRefcntInc     = 15
	opRefcntDec     = 78
	opNewXS         = 81
	opMagicAttach   = 70
	opMagicUnattach = 72
	opPlvarSet      = 74
	opPPHookSet     = 98
	opMagicSetHook  = 133
	opSvCall        = 187
	opGvCall        = 188
)

// gvCallCvNative is the GV_CALL selector that makes a CV call the native
// function of an id (GOPERL_GVCALL_CV_NATIVE in perl-wasm).
const gvCallCvNative = 37

// svCallMagicIsValue is the SV_CALL selector that asks whether a magic type
// is value magic (GOPERL_SVCALL_MAGIC_IS_VALUE in perl-wasm).
const svCallMagicIsValue = 55

// cFrame mirrors goperl_frame_t in sdk/include/perl.h (ABI). The trailing Go
// fields are invisible to C, which only knows the leading portion.
type cFrame struct {
	api          uintptr
	subname      uintptr
	jb           uintptr
	items        int32
	failed       int32
	psp          uintptr // SV** into st, maintained by the SDK macros
	markidx      int32
	hostsaveBase int32
	hookDirty    int32
	tmpidx       int32
	hookVal      [2]uint64
	imm          [3]uint64
	plop         uintptr
	hookOpTok    uint64  // pp-hook frames: the guest op being executed
	prevFrame    uintptr // frame nesting (SDK-maintained)
	instShared   uintptr // this instance's shared-state block (goperl_shared_t)
	// stBase and stCap describe the perl stack of the frame: st below to
	// begin with, a C heap block when the call has more arguments than st
	// holds or the SDK's EXTEND outgrew it. The block is freed when the
	// frame ends (releaseStack).
	stBase uintptr
	stCap  int64
	st     [maxStack]uint64
	marks  [maxMarks]int32
	tmp    [tmpSlots]uint64
	err    [512]byte
	// xerrno is 0, or 1 + the errno the XSUB left when it returned, in the
	// interpreter's numbering: the SDK notes it, dispatch hands it over.
	xerrno int32
	// memBase and memCap are the interpreter's linear memory as native
	// code sees it: where it starts and the size of the range reserved
	// for it. The memory grows inside that range and never moves (the
	// copy-on-write snapshot path), so a pointer is the interpreter's
	// exactly when it lies in the range.
	memBase uintptr
	memCap  uint64

	// Go-only:
	p *internal.Perl
}

// newFrame returns a frame of this instance with no stack yet.
func (s *state) newFrame() *cFrame {
	f := &cFrame{api: uintptr(unsafe.Pointer(vtable)), p: s.p, instShared: s.sharedRaw}
	if mem := s.p.RawMemory(); cap(mem) > 0 {
		f.memBase = uintptr(unsafe.Pointer(unsafe.SliceData(mem)))
		f.memCap = uint64(cap(mem))
	}
	return f
}

// cAPI mirrors goperl_api_t (ABI).
type cAPI struct {
	abi      uint32
	reserved uint32
	svIV     uintptr
	svPV     uintptr
	newIV    uintptr
	newPVN   uintptr
	svMortal uintptr
	regXS    uintptr
	xsOp     uintptr
	// v3:
	guestMem   uintptr
	newXS      uintptr
	cvAny      uintptr
	cvXsub     uintptr
	magicExt   uintptr
	magicChain uintptr
	magicDel   uintptr
	// v4:
	ppHookSet uintptr
	// v5:
	magicSetHead uintptr
	// v6: loader-owned cross-module shared state (zeroed 64KB block).
	sharedRaw uintptr
	// v7: host PerlIO layer registration.
	perlioDef uintptr
	perOpSet  uintptr
	// v11: the host descriptor behind a guest descriptor.
	hostFd uintptr
	// v14: `CvXSUB(cv) = fn`.
	cvSetXsub uintptr
}

// magicRec tracks one guest SV's host-side MAGIC chain.
type magicRec struct {
	svTok     uint64
	head      uintptr // C-heap MAGIC* chain head
	anchorObj uint64  // mg_obj whose refcount the guest anchor holds
	anchorID  uint32
}

func init() {
	// Back (*internal.Perl).AddXSDir with this package's directory loader.
	internal.RegisterXSDirLoader(loadDir)
	// Back DynaLoader's dl_* functions (Config.NativeModules).
	internal.RegisterXSDynamicLoader(internal.XSDynamicLoader{
		Open:    openLibrary,
		Symbol:  librarySymbol,
		Install: installXSub,
	})
}

// ArchTag names the per-architecture native-module directory
// (local/xs/<goos>_<goarch>). The tag follows the RUNNING binary — its
// dlopen must match the .so architecture.
func ArchTag() string { return internal.ArchTag() }

// LoadDir registers every native module under dir (a local/xs/<arch>
// directory, the `gperl xs build` output layout: <Module-Name>.so with the
// package separator spelled "-"). Registration is cheap — each module's
// boot runs lazily when Perl code first `use`s it. A missing dir is not an
// error (the project simply has no native modules).
func loadDir(p *internal.Perl, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".so" {
			continue
		}
		module := strings.ReplaceAll(strings.TrimSuffix(e.Name(), ".so"), "-", "::")
		if err := load(p, module, filepath.Join(dir, e.Name())); err != nil {
			return fmt.Errorf("load native module %s: %w", module, err)
		}
	}
	return nil
}

// state is the per-instance loader state.
type state struct {
	p  *internal.Perl
	mu sync.Mutex
	// libs and syms back the tokens DynaLoader's dl_load_file and
	// dl_find_symbol hand to Perl code (token = index+1).
	libs    []uintptr
	syms    []uintptr
	fns     []uintptr // fnID -> native XSUB pointer
	fnNames [][]byte  // fnID -> NUL-terminated sub name (frame.subname points here)
	// per-CV storage: CvXSUBANY slots (8 bytes of C heap each, written by
	// the module through the vtable) and the native fn behind each CV.
	cvAny map[uint64]uintptr

	// sharedRaw is this instance's C shared-state block (goperl_shared_t):
	// scope bookkeeping, save stacks, and mirrors, aliased by every loaded
	// .so through the frames the loader stamps (frame->inst). The vtable's
	// own shared_raw is the PROCESS block (per-thread current-frame table).
	sharedRaw uintptr
	// perOpFns maps a guest op token to the host pp function that one op
	// dispatches to (an op_ppaddr write on a single op); checked before
	// the per-type hookFns table.
	perOpFns map[uint64]uintptr
	// keywordInvoke is the SDK's __goperl_keyword_invoke entry.
	keywordInvoke uintptr
	// perlioFuncs maps layer funcs ids (handed out by perlioDef) to the
	// module's PerlIO_funcs table for method -6 dispatch.
	perlioFuncs  map[uint32]uintptr
	perlioNextID uint32
	cvFn         map[uint64]uintptr

	// host-side MAGIC mirrors.
	nextMagicID uint32
	magicBySV   map[uint64]*magicRec
	magicByID   map[uint32]*magicRec

	// pp hooks: host pp functions registered for op types (via the SDK's
	// PL_ppaddr proxy diff), and the SDK entry points that run them and
	// the save-stack destructors with croak protection.
	hookFns      map[int32]uintptr
	ppInvoke     uintptr // __goperl_pp_invoke
	dtorInvoke   uintptr // __goperl_dtor_invoke
	svtInvoke    uintptr // __goperl_svt_invoke
	xsEpilogue   uintptr // __goperl_xs_epilogue
	perlioInvoke uintptr // __goperl_perlio_invoke

	// cloneHookSet records that Load registered adoptClone with the
	// instance (once, however many modules load).
	cloneHookSet bool
}

// adoptClone builds the loader state for a copy of this state's instance:
// a perl.Clone, or the child of a fork. The copy's guest memory carries
// every loaded module — booted subs, fn ids, registered layers, compiled
// keyword ops — and every SV with its anchor of host-side magic, so the
// host tables those ids resolve through are carried over.
//
// What the loader itself keeps in C memory for an SV is copied, the way
// the copy of the guest memory copied the SV: each MAGIC of a mirror chain
// and each CvXSUBANY slot is the copy's own from here on. The references a
// chain holds on its mg_obj values were counted in guest memory and came
// along with it. An mg_ptr into the original's linear memory (the string
// buffer of an SV, say) points to the same place in the copy's. Memory a
// MODULE allocated (what an mg_ptr or a slot otherwise points at) cannot
// be copied and is shared by both instances; this SDK is
// a perl without ithreads, so no svt_dup or CLONE hook runs. What the
// loader can do for that memory it does: the copies of one MAGIC are
// released together, svt_free running once, for the last of them (see
// magicShares). Scope bookkeeping starts empty: the native activations it
// belongs to exist only in the original.
func (s *state) adoptClone(clone *internal.Perl) error {
	from, to := spanOf(s.p.RawMemory()), spanOf(clone.RawMemory())
	s.mu.Lock()
	statesMu.Lock()
	cs := &state{
		p:             clone,
		libs:          append([]uintptr(nil), s.libs...),
		syms:          append([]uintptr(nil), s.syms...),
		fns:           append([]uintptr(nil), s.fns...),
		fnNames:       append([][]byte(nil), s.fnNames...),
		cvAny:         make(map[uint64]uintptr, len(s.cvAny)),
		perlioFuncs:   make(map[uint32]uintptr, len(s.perlioFuncs)),
		perlioNextID:  s.perlioNextID,
		cvFn:          make(map[uint64]uintptr, len(s.cvFn)),
		nextMagicID:   s.nextMagicID,
		magicBySV:     make(map[uint64]*magicRec, len(s.magicBySV)),
		magicByID:     make(map[uint32]*magicRec, len(s.magicByID)),
		hookFns:       make(map[int32]uintptr, len(s.hookFns)),
		perOpFns:      make(map[uint64]uintptr, len(s.perOpFns)),
		sharedRaw:     cMalloc(sharedBlockSize),
		keywordInvoke: s.keywordInvoke,
		ppInvoke:      s.ppInvoke,
		dtorInvoke:    s.dtorInvoke,
		svtInvoke:     s.svtInvoke,
		xsEpilogue:    s.xsEpilogue,
		perlioInvoke:  s.perlioInvoke,
		cloneHookSet:  true,
	}
	pokeBytes(cs.sharedRaw, make([]byte, sharedBlockSize))
	for cv, slot := range s.cvAny {
		own := cMalloc(8)
		cMemcpy(own, slot, 8)
		cs.cvAny[cv] = own
	}
	for sv, rec := range s.magicBySV {
		own := &magicRec{svTok: sv, head: copyMagicChain(rec.head, from, to), anchorObj: rec.anchorObj, anchorID: rec.anchorID}
		cs.magicBySV[sv] = own
		cs.magicByID[own.anchorID] = own
	}
	for k, v := range s.perlioFuncs {
		cs.perlioFuncs[k] = v
	}
	for k, v := range s.cvFn {
		cs.cvFn[k] = v
	}
	for k, v := range s.hookFns {
		cs.hookFns[k] = v
	}
	for k, v := range s.perOpFns {
		cs.perOpFns[k] = v
	}
	states[clone] = cs
	statesMu.Unlock()
	s.mu.Unlock()

	clone.AddCloneHook(cs.adoptClone)
	if err := clone.SetNativeXSHandler(cs.dispatch); err != nil {
		return err
	}
	if err := clone.SetSetMagicHandler(cs.magicSetFired); err != nil {
		return err
	}
	if err := clone.SetMagicFreeHandler(cs.magicFreed); err != nil {
		return err
	}
	if err := clone.SetMagicLocalHandler(cs.magicLocalized); err != nil {
		return err
	}
	if err := clone.SetPPHookHandler(cs.ppHook); err != nil {
		return err
	}
	if err := clone.SetKeywordHandler(cs.keywordHook); err != nil {
		return err
	}
	if err := clone.SetPerlIOHandler(cs.perlioHook); err != nil {
		return err
	}
	return clone.SetDestructorHandler(cs.dtorFired)
}

var (
	statesMu sync.Mutex
	states   = map[*internal.Perl]*state{}

	vtableOnce sync.Once
	vtable     *cAPI

	// liveFrames pins every frame for the duration of its native call: it
	// keeps the frame reachable while only C holds its address, and it lets
	// vtable callbacks recover the *cFrame from the raw address without an
	// unsafe uintptr->pointer conversion.
	liveFrames sync.Map // uintptr -> *cFrame
)

func pinFrame(f *cFrame) uintptr {
	addr := uintptr(unsafe.Pointer(f))
	liveFrames.Store(addr, f)
	return addr
}

// unpinFrame ends a frame: its stack is released if it grew onto the heap.
func unpinFrame(addr uintptr) {
	if f := lookupFrame(addr); f != nil {
		f.releaseStack()
	}
	liveFrames.Delete(addr)
}

// lookupFrame resolves a callback's frame argument WITHOUT converting the
// uintptr back into a Go pointer: the pin registry keyed by address hands the
// original *cFrame back. (A direct conversion trips checkptr under -race on
// the purego callback path.)
func lookupFrame(addr uintptr) *cFrame {
	if v, ok := liveFrames.Load(addr); ok {
		return v.(*cFrame)
	}
	return nil
}

func stateFor(p *internal.Perl) *state {
	statesMu.Lock()
	defer statesMu.Unlock()
	s, ok := states[p]
	if !ok {
		s = &state{
			p:           p,
			cvAny:       map[uint64]uintptr{},
			perlioFuncs: map[uint32]uintptr{},
			cvFn:        map[uint64]uintptr{},
			magicBySV:   map[uint64]*magicRec{},
			magicByID:   map[uint32]*magicRec{},
			hookFns:     map[int32]uintptr{},
			perOpFns:    map[uint64]uintptr{},
			sharedRaw:   cMalloc(sharedBlockSize),
		}
		pokeBytes(s.sharedRaw, make([]byte, sharedBlockSize))
		states[p] = s
	}
	return s
}

// Load dlopens a native XS module built against the SDK, injects the vtable,
// and registers its boot function with the instance, so that a stock
// `use Module;` (whose .pm calls XSLoader::load) boots the native module
// lazily at use time. module is the Perl package the .xs declared (e.g.
// "Compiler::Lexer").
//
// Nothing is defined in the module's package until then, as for an
// extension linked into the interpreter: a program that asks whether the
// module is loaded must not find a sub in its package. The boot function
// waits under nativeBootPrefix, where the interpreter's loaders find it
// (DynaLoader::dl_static_boot) and install it as <Module>::bootstrap when
// the module is loaded.
func load(p *internal.Perl, module, path string) error {
	s, err := loaderState(p)
	if err != nil {
		return err
	}
	lib, err := s.open(path)
	if err != nil {
		return err
	}
	bootName := "boot_" + strings.ReplaceAll(module, "::", "__")
	bootFn, err := purego.Dlsym(lib, bootName)
	if err != nil {
		return fmt.Errorf("%s: no %s symbol: %w", path, bootName, err)
	}
	_, err = s.registerNative(nativeBootPrefix+module+"::bootstrap", false, bootFn)
	return err
}

// nativeBootPrefix is the package prefix under which the boot function of a
// registered native module is kept until the module is loaded; see perl-wasm's
// DynaLoader::dl_static_boot.
const nativeBootPrefix = "DynaLoader::native::"

// loaderState returns p's loader state once the C library the loader
// allocates from is bound.
func loaderState(p *internal.Perl) (*state, error) {
	libcInit()
	if libcErr != nil {
		return nil, libcErr
	}
	return stateFor(p), nil
}

// open dlopens a native XS module built against the SDK, injects the
// vtable, and picks up the SDK entry points the module carries.
func (s *state) open(path string) (lib uintptr, err error) {
	p := s.p
	if err := p.SetNativeXSHandler(s.dispatch); err != nil {
		return 0, err
	}
	if err := p.SetSetMagicHandler(s.magicSetFired); err != nil {
		return 0, err
	}
	if err := p.SetMagicFreeHandler(s.magicFreed); err != nil {
		return 0, err
	}
	if err := p.SetMagicLocalHandler(s.magicLocalized); err != nil {
		return 0, err
	}
	if err := p.SetPPHookHandler(s.ppHook); err != nil {
		return 0, err
	}
	if err := p.SetKeywordHandler(s.keywordHook); err != nil {
		return 0, err
	}
	if err := p.SetPerlIOHandler(s.perlioHook); err != nil {
		return 0, err
	}
	if err := p.SetDestructorHandler(s.dtorFired); err != nil {
		return 0, err
	}
	if !s.cloneHookSet {
		s.cloneHookSet = true
		p.AddCloneHook(s.adoptClone)
	}
	vtableOnce.Do(buildVtable)

	lib, err = purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return 0, fmt.Errorf("dlopen %s: %w", path, err)
	}
	initFn, err := purego.Dlsym(lib, "__goperl_xs_init")
	if err != nil {
		return 0, fmt.Errorf("%s is not a go-perl native XS module (no __goperl_xs_init): %w", path, err)
	}
	got, _, _ := purego.SyscallN(initFn, uintptr(unsafe.Pointer(vtable)))
	if uint32(got) != abiVersion {
		return 0, fmt.Errorf("%s: XS SDK ABI mismatch (module rejected vtable v%d)", path, abiVersion)
	}
	// The SDK's croak-guarded entry points for pp hooks and save-stack
	// destructors (weak, so any SDK .so provides them).
	if fn, err := purego.Dlsym(lib, "__goperl_pp_invoke"); err == nil {
		s.mu.Lock()
		s.ppInvoke = fn
		s.mu.Unlock()
	}
	if fn, err := purego.Dlsym(lib, "__goperl_svt_invoke"); err == nil {
		s.mu.Lock()
		s.svtInvoke = fn
		s.mu.Unlock()
	}
	if fn, err := purego.Dlsym(lib, "__goperl_xs_epilogue"); err == nil {
		s.mu.Lock()
		s.xsEpilogue = fn
		s.mu.Unlock()
	}
	if fn, err := purego.Dlsym(lib, "__goperl_keyword_invoke"); err == nil {
		s.keywordInvoke = fn
	}
	if fn, err := purego.Dlsym(lib, "__goperl_perlio_invoke"); err == nil {
		s.mu.Lock()
		s.perlioInvoke = fn
		s.mu.Unlock()
	}
	if fn, err := purego.Dlsym(lib, "__goperl_dtor_invoke"); err == nil {
		s.mu.Lock()
		s.dtorInvoke = fn
		s.mu.Unlock()
	}
	return lib, nil
}

// openLibrary is DynaLoader's dl_load_file: it loads the module at path
// and returns the token Perl code names it by.
func openLibrary(p *internal.Perl, path string) (uint64, error) {
	s, err := loaderState(p)
	if err != nil {
		return 0, err
	}
	lib, err := s.open(path)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.libs = append(s.libs, lib)
	return uint64(len(s.libs)), nil
}

// librarySymbol is DynaLoader's dl_find_symbol.
func librarySymbol(p *internal.Perl, lib uint64, name string) (uint64, error) {
	s := stateFor(p)
	s.mu.Lock()
	if lib == 0 || lib > uint64(len(s.libs)) {
		s.mu.Unlock()
		return 0, fmt.Errorf("unknown library handle %d", lib)
	}
	handle := s.libs[lib-1]
	s.mu.Unlock()
	fn, err := purego.Dlsym(handle, name)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syms = append(s.syms, fn)
	return uint64(len(s.syms)), nil
}

// installXSub is DynaLoader's dl_install_xsub: the function behind sym
// becomes the Perl sub name.
func installXSub(p *internal.Perl, name string, sym uint64) error {
	s := stateFor(p)
	s.mu.Lock()
	if sym == 0 || sym > uint64(len(s.syms)) {
		s.mu.Unlock()
		return fmt.Errorf("unknown symbol handle %d", sym)
	}
	fn := s.syms[sym-1]
	s.mu.Unlock()
	_, err := s.registerNative(name, false, fn)
	return err
}

// registerNative records fn and binds it as the Perl sub name via the
// guest's generic thunk, returning the guest CV token. An anonymous sub
// (newXS without a name) is in no glob: the reference the token stands for
// is the caller's, as with newXS(NULL, ...).
func (s *state) registerNative(name string, anonymous bool, fn uintptr) (uint64, error) {
	s.mu.Lock()
	s.fns = append(s.fns, fn)
	s.fnNames = append(s.fnNames, append([]byte(name), 0))
	id := int32(len(s.fns) - 1)
	s.mu.Unlock()
	var unnamed uint64
	if anonymous {
		unnamed = 1
	}
	cvTok, err := s.p.XSHelperOp(opNewXS, uint64(uint32(id)), unnamed, name)
	if err != nil {
		return 0, err
	}
	if cvTok == 0 {
		return 0, fmt.Errorf("register native XS %s: guest returned no CV", name)
	}
	s.mu.Lock()
	s.cvFn[cvTok] = fn
	s.mu.Unlock()
	return cvTok, nil
}

// setNative makes the existing CV cv call fn: what `CvXSUB(cv) = fn` means.
// The function gets an id of its own, under the name the CV's function had
// when the loader registered one for it.
func (s *state) setNative(cv uint64, fn uintptr) error {
	s.mu.Lock()
	name := []byte{0}
	if old, ok := s.cvFn[cv]; ok {
		for i, registered := range s.fns {
			if registered == old {
				name = s.fnNames[i]
				break
			}
		}
	}
	s.fns = append(s.fns, fn)
	s.fnNames = append(s.fnNames, name)
	id := int32(len(s.fns) - 1)
	s.cvFn[cv] = fn
	s.mu.Unlock()
	_, err := s.p.XSHelperOp(opGvCall, uint64(gvCallCvNative)<<32|cv, uint64(uint32(id)), "")
	return err
}

// dispatch handles one native XSUB call: payload [u32 fn_id][u32 cv]
// [u32 items][u32 tokens...]; response [1][u32 nret][u32 tokens...] or
// [0]+message.
func (s *state) dispatch(req []byte) []byte {
	fail := func(msg string) []byte { return append([]byte{0}, msg...) }
	if len(req) < 12 {
		return fail("native XS dispatch: short payload")
	}
	id := int32(binary.LittleEndian.Uint32(req[0:]))
	cvTok := uint64(binary.LittleEndian.Uint32(req[4:]))
	items := int(binary.LittleEndian.Uint32(req[8:]))
	if len(req) < 12+4*items {
		return fail("native XS dispatch: truncated payload")
	}
	s.mu.Lock()
	var fn, subname uintptr
	if id >= 0 && int(id) < len(s.fns) {
		fn = s.fns[id]
		subname = uintptr(unsafe.Pointer(&s.fnNames[id][0]))
	}
	s.mu.Unlock()
	if fn == 0 {
		return fail(fmt.Sprintf("native XS dispatch: unknown function id %d", id))
	}

	if xsTrace {
		fmt.Fprintf(os.Stderr, "[xsub] %s cv=%#x items=%d\n", cString(subname), cvTok, items)
	}
	f := s.newFrame()
	f.subname = subname
	f.items = int32(items)
	// Stack layout mirrors a real XS entry: a mark at offset 0, arguments
	// at st[1..items] (so ax == 1), sp at the last argument.
	f.marks[0] = 0
	f.markidx = 0
	f.initStack(items)
	for i := 0; i < items; i++ {
		f.setStack(1+i, uint64(binary.LittleEndian.Uint32(req[12+4*i:])))
	}
	f.psp = f.stBase + uintptr(items)*8
	fp := pinFrame(f)
	defer unpinFrame(fp)
	// The XSUB call and the epilogue that closes its activation must run
	// on ONE OS thread: the SDK's current-frame state is per-thread, and
	// the Go scheduler is otherwise free to migrate the goroutine between
	// the two purego calls.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	purego.SyscallN(fn, fp, uintptr(cvTok))
	// Close any activation levels the XSUB left open (xsubpp PPCODE exits
	// with a bare return, skipping XSRETURN's leave).
	s.mu.Lock()
	epi := s.xsEpilogue
	s.mu.Unlock()
	if epi != 0 {
		purego.SyscallN(epi, fp)
	}
	if f.xerrno != 0 {
		// What the C code left in errno is what $! says after the call.
		if _, err := s.p.XSHelperOp(opPlvarSet, plErrno, uint64(f.xerrno-1), ""); err != nil {
			return fail("native XS dispatch: " + err.Error())
		}
	}
	if f.failed == failedThrown {
		// croak_sv: the interpreter holds the thrown value and dies with
		// it when it sees this response.
		return []byte{respThrow}
	}
	if f.failed != 0 {
		return fail(f.errString())
	}
	top := int(int64(f.psp-f.stBase) / 8)
	if top < 0 || int64(top) >= f.stCap {
		return fail(fmt.Sprintf("native XS dispatch: corrupt stack pointer (top %d)", top))
	}
	n := top // ax == 1: return values are st[1..top]
	resp := make([]byte, 5+4*n)
	resp[0] = 1
	binary.LittleEndian.PutUint32(resp[1:], uint32(n))
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint32(resp[5+4*i:], uint32(f.stackAt(1+i)))
	}
	if xsTrace {
		fmt.Fprintf(os.Stderr, "[xsub] %s returns %d: % x\n", cString(subname), n, resp[5:])
	}
	return resp
}

// bootFrame builds a minimal frame for loader-driven native calls that
// happen outside an XSUB dispatch (svt_free teardown).
func (s *state) bootFrame() (*cFrame, uintptr) {
	f := s.newFrame()
	f.marks[0] = 0
	f.markidx = 0
	f.initStack(0)
	f.psp = f.stBase
	return f, pinFrame(f)
}

// ppHook runs a module's pp function for one guest op execution (reserved
// callback method -3). Payload: [u32 op][u32 op_type][u32 n][u32 stack-top
// tokens, bottom to top]; response [1][u32 next_op] or [0]+message.
func (s *state) ppHook(req []byte) []byte {
	fail := func(msg string) []byte { return append([]byte{0}, msg...) }
	if len(req) < 12 {
		return fail("pp hook: short payload")
	}
	opTok := uint64(binary.LittleEndian.Uint32(req[0:]))
	opType := int32(binary.LittleEndian.Uint32(req[4:]))
	n := int(binary.LittleEndian.Uint32(req[8:]))
	if n < 0 || n > 8 || len(req) < 12+4*n {
		return fail("pp hook: truncated payload")
	}
	s.mu.Lock()
	fn := s.perOpFns[opTok]
	if fn == 0 {
		fn = s.hookFns[opType]
	}
	invoke := s.ppInvoke
	s.mu.Unlock()
	if fn == 0 || invoke == 0 {
		return fail(fmt.Sprintf("pp hook: no host function for op %d", opType))
	}
	if xsTrace {
		fmt.Fprintf(os.Stderr, "[pphook] op=%#x type=%d n=%d\n", opTok, opType, n)
	}

	f := s.newFrame()
	f.hookOpTok = opTok
	f.items = int32(n)
	f.marks[0] = 0
	f.markidx = 0
	f.initStack(n)
	for i := 0; i < n; i++ {
		f.setStack(1+i, uint64(binary.LittleEndian.Uint32(req[12+4*i:])))
	}
	f.psp = f.stBase + uintptr(n)*8
	fp := pinFrame(f)
	next, _, _ := purego.SyscallN(invoke, fp, fn)
	unpinFrame(fp)
	if f.failed != 0 {
		if xsTrace {
			fmt.Fprintf(os.Stderr, "[pphook] FAILED: %s\n", f.errString())
		}
		return fail(f.errString())
	}
	if xsTrace {
		fmt.Fprintf(os.Stderr, "[pphook] -> next=%#x\n", next)
	}
	resp := make([]byte, 5)
	resp[0] = 1
	binary.LittleEndian.PutUint32(resp[1:], uint32(next))
	return resp
}

// dtorFired runs a save-stack destructor the guest scope pop released
// (reserved callback method -4).
func (s *state) dtorFired(id uint32) {
	s.mu.Lock()
	invoke := s.dtorInvoke
	s.mu.Unlock()
	if invoke == 0 {
		return
	}
	_, fp := s.bootFrame()
	purego.SyscallN(invoke, fp, uintptr(id))
	unpinFrame(fp)
}

// MAGIC mirror layout (must match struct magic in sdk/include/perl.h):
// mg_moremagic 0, mg_virtual 8, mg_private 16, mg_type 18, mg_flags 19,
// mg_len 20 (i32), mg_obj 24, mg_ptr 32; size 40. MGVTBL: 8 function
// pointers, svt_free at offset 32.
//
// A mirror MAGIC is allocated with a trailer the module does not know of:
// the id of the group of copies it belongs to (magicShares), 0 for a MAGIC
// that was never copied.
const (
	mgSize       = 40
	mgOffShare   = mgSize
	mgAllocSize  = mgSize + 8
	mgOffMore    = 0
	mgOffVirtual = 8
	mgOffType    = 18
	mgOffFlags   = 19
	mgOffLen     = 20
	mgOffObj     = 24
	mgOffPtr     = 32
	vtblOffSet   = 8
	vtblOffFree  = 32
	vtblOffLocal = 56
	mgfLocal     = 0x20 // MGf_LOCAL
	hefSVKey     = -2   // HEf_SVKEY: the name of a MAGIC is an SV
)

// magicShares counts, for every MAGIC that was copied to another instance
// (a Clone or a fork child), the instances that still hold a copy. The
// copies point at the same module memory, which the module's svt_free
// releases, so that hook must run exactly once: when the last copy goes.
// Which other instances still hold a copy is not something one instance's
// state can tell, hence the process-wide table.
var magicShares = struct {
	mu      sync.Mutex
	nextID  uint64
	holders map[uint64]int
}{holders: map[uint64]int{}}

// shareMagic makes the copy of the MAGIC at mg, which is one more holder of
// what the original holds.
func shareMagic(mg uintptr, node []byte) {
	magicShares.mu.Lock()
	defer magicShares.mu.Unlock()
	id := binary.LittleEndian.Uint64(node[mgOffShare:])
	if id == 0 {
		magicShares.nextID++
		id = magicShares.nextID
		magicShares.holders[id] = 1
		binary.LittleEndian.PutUint64(node[mgOffShare:], id)
		pokeBytes(mg+mgOffShare, node[mgOffShare:mgOffShare+8])
	}
	magicShares.holders[id]++
}

// lastMagicHolder ends this instance's hold on a MAGIC (node is its bytes)
// and reports whether no other instance holds a copy any more: whether
// svt_free is due.
func lastMagicHolder(node []byte) bool {
	id := binary.LittleEndian.Uint64(node[mgOffShare:])
	if id == 0 {
		return true
	}
	magicShares.mu.Lock()
	defer magicShares.mu.Unlock()
	magicShares.holders[id]--
	if magicShares.holders[id] > 0 {
		return false
	}
	delete(magicShares.holders, id)
	return true
}

func (s *state) magicExt(f *cFrame, sv, obj uint64, how int32, vtbl, namePtr uintptr, namLen int64) uintptr {
	mg := cMalloc(mgAllocSize)
	if mg == 0 {
		f.fail("sv_magicext: out of memory")
		return 0
	}
	s.mu.Lock()
	rec := s.magicBySV[sv]
	first := rec == nil
	var oldHead uintptr
	if rec != nil {
		oldHead = rec.head
	}
	s.mu.Unlock()

	// sv_magicext keeps the length it is given. A positive one makes the
	// MAGIC own a copy of the name; HEf_SVKEY says the name is an SV, which
	// the MAGIC then holds a reference to; with any other the pointer is
	// kept as it is.
	mgPtr := namePtr
	mgLen := int32(namLen)
	switch {
	case namePtr == 0:
	case namLen > 0:
		mgPtr = cMalloc(int(namLen) + 1)
		cMemcpy(mgPtr, namePtr, int(namLen))
		pokeBytes(mgPtr+uintptr(namLen), []byte{0})
	case namLen == hefSVKey:
		if _, err := s.p.XSHelperOp(opRefcntInc, uint64(namePtr), 0, ""); err != nil {
			f.fail("sv_magicext: " + err.Error())
			return 0
		}
	}
	var buf [mgAllocSize]byte
	binary.LittleEndian.PutUint64(buf[mgOffMore:], uint64(oldHead))
	binary.LittleEndian.PutUint64(buf[mgOffVirtual:], uint64(vtbl))
	buf[mgOffType] = byte(how)
	binary.LittleEndian.PutUint32(buf[mgOffLen:], uint32(mgLen))
	binary.LittleEndian.PutUint64(buf[mgOffObj:], obj)
	binary.LittleEndian.PutUint64(buf[mgOffPtr:], uint64(mgPtr))
	pokeBytes(mg, buf[:])

	// A vtbl with svt_set needs the guest anchor upgraded so assignments
	// to the SV are reported back (reserved method -5). Anchors start
	// set-less: most mirrors only carry svt_free, and a set-firing anchor
	// costs a host round trip on every write to the SV.
	wantSet := false
	if vtbl != 0 {
		setFn := copyIn(vtbl+vtblOffSet, 8)
		wantSet = binary.LittleEndian.Uint64(setFn) != 0
	}
	if xsTrace {
		fmt.Fprintf(os.Stderr, "[magicExt] sv=%#x how=%d vtbl=%#x wantSet=%v first=%v\n", sv, how, vtbl, wantSet, first)
	}

	if first {
		s.mu.Lock()
		s.nextMagicID++
		rec = &magicRec{svTok: sv, head: mg, anchorObj: obj, anchorID: s.nextMagicID}
		s.magicBySV[sv] = rec
		s.magicByID[rec.anchorID] = rec
		s.mu.Unlock()
		// The guest anchor holds the SV<->host-chain link and (perl
		// semantics) a reference on obj.
		if _, err := s.p.XSHelperOp(opMagicAttach, sv,
			uint64(rec.anchorID)<<32|obj&0xFFFFFFFF, ""); err != nil {
			f.fail("sv_magicext: " + err.Error())
			return 0
		}
		if wantSet {
			r, err := s.p.XSHelperOp(opMagicSetHook, sv, 0, "")
			if err != nil {
				f.fail("sv_magicext: " + err.Error())
			}
			if xsTrace {
				fmt.Fprintf(os.Stderr, "[magicExt] set-hook upgrade -> %d\n", r)
			}
		}
	} else {
		if wantSet {
			if _, err := s.p.XSHelperOp(opMagicSetHook, sv, 0, ""); err != nil {
				f.fail("sv_magicext: " + err.Error())
			}
		}
		s.mu.Lock()
		rec.head = mg
		s.mu.Unlock()
		if obj != 0 {
			// Real sv_magicext takes a reference on obj for every entry;
			// the anchor only covers the first, so hold this one manually.
			if _, err := s.p.XSHelperOp(opRefcntInc, obj, 0, ""); err != nil {
				f.fail("sv_magicext: " + err.Error())
			}
		}
	}
	return mg
}

// memSpan is the place of an instance's linear memory in the process.
type memSpan struct{ base, size uintptr }

func spanOf(mem []byte) memSpan {
	if len(mem) == 0 {
		return memSpan{}
	}
	return memSpan{base: uintptr(unsafe.Pointer(&mem[0])), size: uintptr(len(mem))}
}

// copyMagicChain returns a copy of the mirror chain at head, in the same
// order, for an instance whose linear memory is at to and is a copy of the
// memory at from: every MAGIC, and the name sv_magicext copied for it
// (mg_len > 0). An mg_ptr into the original's memory becomes the pointer
// to the same place in the copy's; any other is the module's and is the
// same pointer in the copy.
func copyMagicChain(head uintptr, from, to memSpan) uintptr {
	var first, last uintptr
	for mg := head; mg != 0; {
		node := copyIn(mg, mgAllocSize)
		next := uintptr(binary.LittleEndian.Uint64(node[mgOffMore:]))
		mgLen := int32(binary.LittleEndian.Uint32(node[mgOffLen:]))
		mgPtr := uintptr(binary.LittleEndian.Uint64(node[mgOffPtr:]))
		shareMagic(mg, node)
		if mgLen > 0 && mgPtr != 0 {
			name := cMalloc(int(mgLen) + 1)
			cMemcpy(name, mgPtr, int(mgLen)+1)
			binary.LittleEndian.PutUint64(node[mgOffPtr:], uint64(name))
		} else if mgPtr >= from.base && mgPtr-from.base < from.size {
			binary.LittleEndian.PutUint64(node[mgOffPtr:], uint64(to.base+(mgPtr-from.base)))
		}
		binary.LittleEndian.PutUint64(node[mgOffMore:], 0)
		own := cMalloc(mgAllocSize)
		pokeBytes(own, node)
		if last == 0 {
			first = own
		} else {
			var link [8]byte
			binary.LittleEndian.PutUint64(link[:], uint64(own))
			pokeBytes(last+mgOffMore, link[:])
		}
		last = own
		mg = next
	}
	return first
}

func (s *state) magicChainHead(sv uint64) uintptr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec := s.magicBySV[sv]; rec != nil {
		return rec.head
	}
	return 0
}

// callSvtHook invokes a node's vtbl entry at vtblOff — int (*)(pTHX_ SV*,
// MAGIC*), the shared shape of svt_set and svt_free — if present.
func (s *state) callSvtHook(sv uint64, mg, vtblOff uintptr) {
	node := copyIn(mg, mgSize)
	vtbl := uintptr(binary.LittleEndian.Uint64(node[mgOffVirtual:]))
	if vtbl == 0 {
		return
	}
	fnBytes := copyIn(vtbl+vtblOff, 8)
	fn := uintptr(binary.LittleEndian.Uint64(fnBytes))
	if fn == 0 {
		return
	}
	f, fp := s.bootFrame()
	s.mu.Lock()
	invoke := s.svtInvoke
	s.mu.Unlock()
	if invoke != 0 {
		// The SDK-side entry installs the current frame and a croak guard
		// before running the hook; interpreter macros inside it need both.
		purego.SyscallN(invoke, fp, fn, uintptr(sv), mg)
	} else {
		purego.SyscallN(fn, fp, uintptr(sv), mg)
	}
	_ = f
	unpinFrame(fp)
}

func (s *state) callSvtFree(sv uint64, mg uintptr) {
	s.callSvtHook(sv, mg, vtblOffFree)
}

// magicSetFired runs when the guest reports an assignment to an SV whose
// anchor was upgraded: run every mirror node's svt_set. Hooks may unlink
// themselves (SvMAGIC_set + Safefree, the Moose export-flag pattern), so
// each node's successor is read BEFORE the hook runs.
func (s *state) magicSetFired(id uint32) {
	s.mu.Lock()
	rec := s.magicByID[id]
	var head uintptr
	if rec != nil {
		head = rec.head
	}
	s.mu.Unlock()
	if rec == nil {
		return
	}
	for mg := head; mg != 0; {
		node := copyIn(mg, mgSize)
		next := uintptr(binary.LittleEndian.Uint64(node[mgOffMore:]))
		s.callSvtHook(rec.svTok, mg, vtblOffSet)
		mg = next
	}
}

// magicLocalized runs when `local` puts the SV nsv in the place of the
// variable whose anchor is id. It does for the mirror chain what perl's
// mg_localize does for an SV's own: value magic stays behind; an entry
// whose vtable has svt_local and asks for it (MGf_LOCAL) decides itself;
// every other entry is attached to nsv again with sv_magicext, which gives
// nsv a chain and an anchor of its own. The interpreter then calls set
// magic on nsv, and frees its magic with it when the scope ends.
func (s *state) magicLocalized(id uint32, nsv uint64) {
	s.mu.Lock()
	rec := s.magicByID[id]
	var head uintptr
	if rec != nil {
		head = rec.head
	}
	s.mu.Unlock()
	for mg := head; mg != 0; {
		node := copyIn(mg, mgSize)
		next := uintptr(binary.LittleEndian.Uint64(node[mgOffMore:]))
		how := int32(int8(node[mgOffType]))
		vtbl := uintptr(binary.LittleEndian.Uint64(node[mgOffVirtual:]))
		isValue, err := s.p.XSHelperOp(opSvCall, uint64(svCallMagicIsValue)<<32|uint64(node[mgOffType]), 0, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "goperl xs: local on a variable with magic: %v\n", err)
			return
		}
		switch {
		case isValue != 0:
		case node[mgOffFlags]&mgfLocal != 0 && vtbl != 0 && binary.LittleEndian.Uint64(copyIn(vtbl+vtblOffLocal, 8)) != 0:
			s.callSvtHook(nsv, mg, vtblOffLocal)
		default:
			f, fp := s.bootFrame()
			s.magicExt(f, nsv, binary.LittleEndian.Uint64(node[mgOffObj:]), how, vtbl,
				uintptr(binary.LittleEndian.Uint64(node[mgOffPtr:])),
				int64(int32(binary.LittleEndian.Uint32(node[mgOffLen:]))))
			if f.failed != 0 {
				fmt.Fprintf(os.Stderr, "goperl xs: local on a variable with magic: %s\n", f.errString())
			}
			unpinFrame(fp)
		}
		mg = next
	}
}

// perlioHook services reserved method -6: run one host layer hook (the
// SDK-side __goperl_perlio_invoke does the marshalling and croak guard).
func (s *state) perlioHook(req []byte) []byte {
	if len(req) < 16 {
		return append([]byte{0}, "perlio hook: short request"...)
	}
	id := binary.LittleEndian.Uint32(req[4:])
	s.mu.Lock()
	funcs := s.perlioFuncs[id]
	inv := s.perlioInvoke
	s.mu.Unlock()
	if funcs == 0 || inv == 0 {
		return append([]byte{0}, "perlio hook: unknown layer"...)
	}
	reqC := cMalloc(len(req))
	pokeBytes(reqC, req)
	respC := cMalloc(4096)
	f, fp := s.bootFrame()
	rlen, _, _ := purego.SyscallN(inv, fp, funcs, reqC, uintptr(len(req)), respC, 4096)
	_ = f
	unpinFrame(fp)
	n := int(int32(rlen))
	var out []byte
	if n <= 0 || n > 4096 {
		out = append([]byte{0}, "perlio hook: bad response"...)
	} else {
		out = copyIn(respC, n)
	}
	cFree(reqC)
	cFree(respC)
	return out
}

// keywordHook services reserved method -7: run the host keyword/infix
// plugin chain for one candidate word (the SDK's __goperl_keyword_invoke
// does the marshalling, parser sync, and croak guard). Response:
// [1][handled u8][ret u32][op token u32] or [0]+message.
func (s *state) keywordHook(req []byte) []byte {
	if len(req) < 5 {
		return append([]byte{0}, "keyword hook: short request"...)
	}
	kind := binary.LittleEndian.Uint32(req[0:])
	word := req[4:]
	s.mu.Lock()
	inv := s.keywordInvoke
	s.mu.Unlock()
	if inv == 0 {
		// no SDK module with a parse surface loaded: decline
		return []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}
	wordC := cMalloc(len(word) + 1)
	pokeBytes(wordC, word)
	pokeBytes(wordC+uintptr(len(word)), []byte{0})
	retC := cMalloc(4)
	tokC := cMalloc(8)
	f, fp := s.bootFrame()
	rc, _, _ := purego.SyscallN(inv, fp, wordC, uintptr(len(word)),
		uintptr(kind), retC, tokC)
	cFree(wordC)
	out := make([]byte, 10)
	switch int32(rc) {
	case -1:
		msg := f.errString()
		cFree(retC)
		cFree(tokC)
		unpinFrame(fp)
		return append([]byte{0}, msg...)
	case 1:
		out[0] = 1
		out[1] = 1
		copy(out[2:6], copyIn(retC, 4))
		tok := copyIn(tokC, 8)
		copy(out[6:10], tok[:4])
	default:
		out[0] = 1
		out[1] = 0
	}
	cFree(retC)
	cFree(tokC)
	unpinFrame(fp)
	return out
}

// magicSetHead implements SvMAGIC_set over the mirror chain: the module
// relinked the chain itself (typically to unlink the head) and hands us the
// new head. Unlinked nodes are the module's to free.
func (s *state) magicSetHead(sv uint64, mg uintptr) {
	s.mu.Lock()
	if rec := s.magicBySV[sv]; rec != nil {
		rec.head = mg
	}
	s.mu.Unlock()
}

// magicFreed runs when the guest frees an SV carrying the anchor magic:
// execute every mirror node's svt_free and drop the chain.
func (s *state) magicFreed(id uint32) {
	s.mu.Lock()
	rec := s.magicByID[id]
	if rec != nil {
		delete(s.magicByID, id)
		delete(s.magicBySV, rec.svTok)
	}
	s.mu.Unlock()
	if rec == nil {
		return
	}
	for mg := rec.head; mg != 0; {
		node := copyIn(mg, mgAllocSize)
		next := uintptr(binary.LittleEndian.Uint64(node[mgOffMore:]))
		obj := binary.LittleEndian.Uint64(node[mgOffObj:])
		if xsTrace {
			s.mu.Lock()
			slot, hasAny := s.cvAny[rec.svTok]
			s.mu.Unlock()
			var val uint64
			if hasAny {
				val = binary.LittleEndian.Uint64(copyIn(slot, 8))
			}
			vtblp := binary.LittleEndian.Uint64(node[mgOffVirtual:])
			fmt.Fprintf(os.Stderr, "[magicFreed] sv=%#x mg=%#x vtbl=%#x any=%#x\n", rec.svTok, mg, vtblp, val)
		}
		mgLen := int32(binary.LittleEndian.Uint32(node[mgOffLen:]))
		mgPtr := uintptr(binary.LittleEndian.Uint64(node[mgOffPtr:]))
		if lastMagicHolder(node) {
			s.callSvtFree(rec.svTok, mg)
		}
		// Every entry beyond the anchor's held one manually retained obj.
		if obj != 0 && obj != rec.anchorObj {
			_, _ = s.p.XSHelperOp(opRefcntDec, obj, 0, "")
		} else if obj != 0 && obj == rec.anchorObj {
			rec.anchorObj = 0 // the anchor releases its own reference
		}
		s.releaseMagicName(mgLen, mgPtr)
		cFree(mg)
		mg = next
	}
}

// releaseMagicName gives up what a MAGIC holds as its name: the copy
// sv_magicext made (a positive length), or its reference to the SV that is
// the name (HEf_SVKEY).
func (s *state) releaseMagicName(mgLen int32, mgPtr uintptr) {
	switch {
	case mgPtr == 0:
	case mgLen > 0:
		cFree(mgPtr)
	case mgLen == hefSVKey:
		_, _ = s.p.XSHelperOp(opRefcntDec, uint64(mgPtr), 0, "")
	}
}

// magicDel implements sv_unmagic/sv_unmagicext over the mirror chain.
func (s *state) magicDel(sv uint64, how int32, vtbl uintptr) {
	s.mu.Lock()
	rec := s.magicBySV[sv]
	s.mu.Unlock()
	if rec == nil {
		return
	}
	var prev uintptr
	mg := rec.head
	for mg != 0 {
		node := copyIn(mg, mgAllocSize)
		next := uintptr(binary.LittleEndian.Uint64(node[mgOffMore:]))
		typ := node[mgOffType]
		nodeVtbl := uintptr(binary.LittleEndian.Uint64(node[mgOffVirtual:]))
		if typ == byte(how) && (vtbl == 0 || nodeVtbl == vtbl) {
			obj := binary.LittleEndian.Uint64(node[mgOffObj:])
			mgLen := int32(binary.LittleEndian.Uint32(node[mgOffLen:]))
			mgPtr := uintptr(binary.LittleEndian.Uint64(node[mgOffPtr:]))
			if lastMagicHolder(node) {
				s.callSvtFree(sv, mg)
			}
			if obj != 0 && obj != rec.anchorObj {
				_, _ = s.p.XSHelperOp(opRefcntDec, obj, 0, "")
			}
			s.releaseMagicName(mgLen, mgPtr)
			if prev == 0 {
				rec.head = next
			} else {
				var b [8]byte
				binary.LittleEndian.PutUint64(b[:], uint64(next))
				pokeBytes(prev+mgOffMore, b[:])
			}
			cFree(mg)
			mg = next
			continue
		}
		prev = mg
		mg = next
	}
	if rec.head == 0 {
		s.mu.Lock()
		delete(s.magicBySV, sv)
		delete(s.magicByID, rec.anchorID)
		s.mu.Unlock()
		// Detaching the guest anchor would re-enter magicFreed; the maps
		// are already cleared, so that callback finds nothing to do.
		_, _ = s.p.XSHelperOp(opMagicUnattach, sv, 0, "")
	}
}

func (s *state) cvAnySlot(cv uint64) uintptr {
	s.mu.Lock()
	defer s.mu.Unlock()
	slot, ok := s.cvAny[cv]
	if !ok {
		slot = cMalloc(8)
		pokeBytes(slot, make([]byte, 8))
		s.cvAny[cv] = slot
	}
	return slot
}

// buildVtable creates the C-callable API table. Callbacks are process-global
// (purego.NewCallback registrations are permanent); the frame carries the
// instance.
func buildVtable() {
	vtable = &cAPI{abi: abiVersion}
	// The vtable's block is PROCESS-wide state: the per-thread current-
	// frame table every .so resolves _gof through. Instance state lives in
	// each state's own sharedRaw, stamped onto frames.
	vtable.sharedRaw = cMalloc(sharedBlockSize)
	pokeBytes(vtable.sharedRaw, make([]byte, sharedBlockSize))
	vtable.svIV = purego.NewCallback(func(fr, sv uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		v, err := f.p.XSHelperOp(1, uint64(sv), 0, "")
		if err != nil {
			f.fail("SvIV: " + err.Error())
			return 0
		}
		return uintptr(v)
	})
	vtable.svPV = purego.NewCallback(func(fr, sv, lenp uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		packed, err := f.p.XSHelperOp(2, uint64(sv), 0, "")
		if err != nil {
			f.fail("SvPV: " + err.Error())
			return 0
		}
		off := uint32(packed >> 32)
		ln := uint32(packed)
		if lenp != 0 {
			pokeU64(lenp, uint64(ln))
		}
		mem := f.p.RawMemory()
		if int(off)+int(ln) > len(mem) {
			f.fail("SvPV: string outside linear memory")
			return 0
		}
		return uintptr(unsafe.Pointer(&mem[off]))
	})
	vtable.newIV = purego.NewCallback(func(fr, v uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		sv, err := f.p.XSHelperOp(3, uint64(v), 0, "")
		if err != nil {
			f.fail("newSViv: " + err.Error())
			return 0
		}
		return uintptr(sv)
	})
	vtable.newPVN = purego.NewCallback(func(fr, p, ln uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		sv, err := f.p.XSHelperOp(4, 0, uint64(ln), goBytesString(p, int(ln)))
		if err != nil {
			f.fail("newSVpvn: " + err.Error())
			return 0
		}
		return uintptr(sv)
	})
	vtable.svMortal = purego.NewCallback(func(fr, sv uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		out, err := f.p.XSHelperOp(5, uint64(sv), 0, "")
		if err != nil {
			f.fail("sv_2mortal: " + err.Error())
			return 0
		}
		return uintptr(out)
	})
	vtable.xsOp = purego.NewCallback(func(fr uintptr, op int32, a, b, sPtr, sLen uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		v, err := f.p.XSHelperOp(op, uint64(a), uint64(b), goBytesString(sPtr, int(sLen)))
		if xsTrace {
			fmt.Fprintf(os.Stderr, "[xsop] op=%d a=%#x b=%#x slen=%d -> %#x\n", op, a, b, sLen, v)
		}
		if err != nil {
			// Loud on purpose: native code cannot be unwound from here, so
			// it continues with a zero result; the operator must see why.
			fmt.Fprintf(os.Stderr, "goperl xs: op %d failed: %v\n", op, err)
			f.fail(fmt.Sprintf("xs op %d: %v", op, err))
			return 0
		}
		return uintptr(v)
	})
	vtable.hostFd = purego.NewCallback(func(fr, fd uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return ^uintptr(0)
		}
		host, ok := f.p.HostDescriptor(int32(fd))
		if !ok {
			return ^uintptr(0) // -1: the handle has no descriptor of the host
		}
		return host
	})
	vtable.cvSetXsub = purego.NewCallback(func(fr, cv, fn uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		if err := stateFor(f.p).setNative(uint64(cv), fn); err != nil {
			f.fail("CvXSUB: " + err.Error())
		}
		return 0
	})
	vtable.regXS = purego.NewCallback(func(fr, name, fn uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		s := stateFor(f.p)
		if _, err := s.registerNative(cString(name), name == 0, fn); err != nil {
			f.fail("newXS: " + err.Error())
		}
		return 0
	})
	vtable.guestMem = purego.NewCallback(func(fr, gptr uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		mem := f.p.RawMemory()
		if gptr == 0 || int(gptr) >= len(mem) {
			f.fail("guest_mem: pointer outside linear memory")
			return 0
		}
		return uintptr(unsafe.Pointer(&mem[gptr]))
	})
	vtable.newXS = purego.NewCallback(func(fr, name, fn uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		s := stateFor(f.p)
		cvTok, err := s.registerNative(cString(name), name == 0, fn)
		if err != nil {
			f.fail("newXS: " + err.Error())
			return 0
		}
		return uintptr(cvTok)
	})
	vtable.cvAny = purego.NewCallback(func(fr, cv uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			if xsTrace {
				fmt.Fprintf(os.Stderr, "[cvAny] frame %#x not pinned (cv=%#x)\n", fr, cv)
			}
			return 0
		}
		return stateFor(f.p).cvAnySlot(uint64(cv))
	})
	vtable.cvXsub = purego.NewCallback(func(fr, cv uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		s := stateFor(f.p)
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.cvFn[uint64(cv)]
	})
	vtable.magicExt = purego.NewCallback(func(fr, sv, obj uintptr, how int32, vtbl, name, namlen uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		return stateFor(f.p).magicExt(f, uint64(sv), uint64(obj), how, vtbl, name, int64(namlen))
	})
	vtable.magicChain = purego.NewCallback(func(fr, sv uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		return stateFor(f.p).magicChainHead(uint64(sv))
	})
	vtable.magicDel = purego.NewCallback(func(fr, sv uintptr, how int32, vtbl uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		stateFor(f.p).magicDel(uint64(sv), how, vtbl)
		return 0
	})
	vtable.perlioDef = purego.NewCallback(func(fr, name, funcs uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		s := stateFor(f.p)
		s.mu.Lock()
		s.perlioNextID++
		id := s.perlioNextID
		s.perlioFuncs[id] = funcs
		s.mu.Unlock()
		_ = name
		return uintptr(id)
	})
	vtable.perOpSet = purego.NewCallback(func(fr uintptr, opTok uint64, fn uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		s := stateFor(f.p)
		s.mu.Lock()
		if fn != 0 {
			s.perOpFns[opTok] = fn
		} else {
			delete(s.perOpFns, opTok)
		}
		s.mu.Unlock()
		return 0
	})
	vtable.magicSetHead = purego.NewCallback(func(fr, sv, mg uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		stateFor(f.p).magicSetHead(uint64(sv), mg)
		return 0
	})
	vtable.ppHookSet = purego.NewCallback(func(fr uintptr, opType int32, fn uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		s := stateFor(f.p)
		s.mu.Lock()
		if fn != 0 {
			s.hookFns[opType] = fn
		} else {
			delete(s.hookFns, opType)
		}
		s.mu.Unlock()
		enable := uint64(0)
		if fn != 0 {
			enable = 1
		}
		if _, err := f.p.XSHelperOp(opPPHookSet, uint64(uint32(opType)), enable, ""); err != nil {
			f.fail("pp hook set: " + err.Error())
		}
		return 0
	})
}

// initStack gives the frame a stack with room for items arguments: its own
// array, or a C heap block when they do not fit.
func (f *cFrame) initStack(items int) {
	if items+stackHeadroom <= maxStack {
		f.stBase = uintptr(unsafe.Pointer(&f.st[0]))
		f.stCap = maxStack
		return
	}
	f.stCap = int64(items + stackHeadroom)
	f.stBase = cMalloc(int(f.stCap) * 8)
	pokeBytes(f.stBase, make([]byte, int(f.stCap)*8))
}

// inlineStack reports whether the stack is still the frame's own array.
func (f *cFrame) inlineStack() bool {
	return f.stBase == uintptr(unsafe.Pointer(&f.st[0]))
}

// setStack and stackAt write and read slot i of the stack, wherever it is.
func (f *cFrame) setStack(i int, tok uint64) {
	if f.inlineStack() {
		f.st[i] = tok
		return
	}
	pokeU64(f.stBase+uintptr(i)*8, tok)
}

func (f *cFrame) stackAt(i int) uint64 {
	if f.inlineStack() {
		return f.st[i]
	}
	return binary.LittleEndian.Uint64(copyIn(f.stBase+uintptr(i)*8, 8))
}

// releaseStack frees the stack when it is a heap block (the loader's for a
// long argument list, or the one the SDK moved it to).
func (f *cFrame) releaseStack() {
	if f.stBase != 0 && !f.inlineStack() {
		cFree(f.stBase)
		f.stBase = uintptr(unsafe.Pointer(&f.st[0]))
		f.stCap = maxStack
	}
}

// fail records an error on the frame; the dispatcher reports it after the
// XSUB returns. (Native code cannot be unwound from Go, so the failure is
// carried out-of-band.)
func (f *cFrame) fail(msg string) {
	f.failed = 1
	n := copy(f.err[:len(f.err)-1], msg)
	f.err[n] = 0
}

// Foreign-memory access. checkptr (enabled under -race) fatals on any Go
// dereference of memory it cannot attribute to a Go allocation — which is
// exactly what the dlopen'd library's strings and C stack are. So Go never
// dereferences foreign memory here: libc does the touching, driven through
// purego (strlen to size a C string, memcpy to move bytes across, malloc/
// free for the MAGIC mirrors C reads), and Go only ever reads its own
// buffers.
var (
	libcOnce sync.Once
	strlenFn uintptr
	memcpyFn uintptr
	mallocFn uintptr
	freeFn   uintptr
	libcErr  error
)

func libcInit() {
	libcOnce.Do(func() {
		// A perl executable is linked with the C math library, and XS
		// modules count on its functions being loaded already: many are
		// linked without -lm. This binary has no libm of its own, so the
		// library goes into the global namespace before the first module.
		// Where it is missing, a module that needs it fails to load with
		// the dynamic linker's own message.
		if runtime.GOOS == "linux" {
			_, _ = purego.Dlopen("libm.so.6", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		}
		for _, sym := range []struct {
			name string
			dst  *uintptr
		}{{"strlen", &strlenFn}, {"memcpy", &memcpyFn}, {"malloc", &mallocFn}, {"free", &freeFn}} {
			fn, err := purego.Dlsym(purego.RTLD_DEFAULT, sym.name)
			if err != nil {
				libcErr = fmt.Errorf("resolve libc %s: %w", sym.name, err)
				return
			}
			*sym.dst = fn
		}
	})
}

func cMalloc(n int) uintptr {
	p, _, _ := purego.SyscallN(mallocFn, uintptr(n))
	return p
}

func cFree(p uintptr) {
	if p != 0 {
		purego.SyscallN(freeFn, p)
	}
}

func cMemcpy(dst, src uintptr, n int) {
	if n > 0 {
		purego.SyscallN(memcpyFn, dst, src, uintptr(n))
	}
}

// copyIn copies n bytes of foreign memory at src into a fresh Go slice.
func copyIn(src uintptr, n int) []byte {
	if src == 0 || n <= 0 {
		return nil
	}
	buf := make([]byte, n)
	purego.SyscallN(memcpyFn, uintptr(unsafe.Pointer(&buf[0])), src, uintptr(n))
	runtime.KeepAlive(buf)
	return buf
}

// cString reads a NUL-terminated C string at ptr.
func cString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	n, _, _ := purego.SyscallN(strlenFn, ptr)
	return string(copyIn(ptr, int(n)))
}

// goBytesString copies n bytes of foreign memory at ptr into a Go string.
func goBytesString(ptr uintptr, n int) string {
	return string(copyIn(ptr, n))
}

// pokeU64 writes an 8-byte little-endian value into foreign memory (an
// out-parameter on the native caller's stack).
func pokeU64(dst uintptr, v uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], v)
	purego.SyscallN(memcpyFn, dst, uintptr(unsafe.Pointer(&buf[0])), 8)
	runtime.KeepAlive(&buf)
}

// pokeBytes writes a Go byte slice into foreign memory.
func pokeBytes(dst uintptr, b []byte) {
	if len(b) == 0 {
		return
	}
	purego.SyscallN(memcpyFn, dst, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	runtime.KeepAlive(b)
}

// errString reads the frame's croak message (Go memory; plain Go access).
func (f *cFrame) errString() string {
	b := f.err[:]
	if i := bytesIndexZero(b); i >= 0 {
		b = b[:i]
	}
	return string(b)
}

func bytesIndexZero(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return -1
}
