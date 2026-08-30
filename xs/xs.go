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
	perl "github.com/goccy/go-perl"
)

// xsTrace logs every guest micro-op to stderr (GOPERL_XS_TRACE=1); the
// debugging aid of last resort when a native module misbehaves.
var xsTrace = os.Getenv("GOPERL_XS_TRACE") != ""

// abiVersion must match GOPERL_XS_ABI in sdk/include/perl.h.
const abiVersion = 9

const (
	maxStack = 512 // GOPERL_XS_STACK
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
	opPPHookSet     = 98
	opMagicSetHook  = 133
)

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
	st           [maxStack]uint64
	marks        [maxMarks]int32
	tmp          [tmpSlots]uint64
	err          [512]byte

	// Go-only:
	p *perl.Perl
}

// cAPI mirrors goperl_api_t (ABI).
type cAPI struct {
	abi       uint32
	reserved  uint32
	svIV      uintptr
	svPV      uintptr
	newIV     uintptr
	newPVN    uintptr
	svMortal  uintptr
	regXS     uintptr
	xsOp      uintptr
	ptrEncode uintptr
	ptrDecode uintptr
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
}

// magicRec tracks one guest SV's host-side MAGIC chain.
type magicRec struct {
	svTok     uint64
	head      uintptr // C-heap MAGIC* chain head
	anchorObj uint64  // mg_obj whose refcount the guest anchor holds
	anchorID  uint32
}

func init() {
	// Back (*perl.Perl).AddXSDir with this package's directory loader.
	perl.RegisterXSDirLoader(LoadDir)
}

// ArchTag names the per-architecture native-module directory
// (local/xs/<goos>_<goarch>). The tag follows the RUNNING binary — its
// dlopen must match the .so architecture.
func ArchTag() string { return runtime.GOOS + "_" + runtime.GOARCH }

// LoadDir registers every native module under dir (a local/xs/<arch>
// directory, the `gperl xs build` output layout: <Module-Name>.so with the
// package separator spelled "-"). Registration is cheap — each module's
// boot runs lazily when Perl code first `use`s it. A missing dir is not an
// error (the project simply has no native modules).
func LoadDir(p *perl.Perl, dir string) error {
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
		if err := Load(p, module, filepath.Join(dir, e.Name())); err != nil {
			return fmt.Errorf("load native module %s: %w", module, err)
		}
	}
	return nil
}

// state is the per-instance loader state.
type state struct {
	p       *perl.Perl
	mu      sync.Mutex
	fns     []uintptr // fnID -> native XSUB pointer
	fnNames [][]byte  // fnID -> NUL-terminated sub name (frame.subname points here)
	// ptrs is the host-pointer registry backing T_PTROBJ: guest IVs are
	// 32-bit, so a native object pointer crosses as (index+1) here.
	ptrs []uintptr

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

// adoptClone builds the loader state for a perl.Clone of this state's
// instance. The clone's guest memory carries every loaded module — booted
// subs, fn ids, registered layers, compiled keyword ops — so the host
// tables those ids resolve through are carried over; runtime state (magic
// mirrors, scope bookkeeping) starts empty, exactly like the guest side
// does at rest.
func (s *state) adoptClone(clone *perl.Perl) error {
	s.mu.Lock()
	if len(s.magicBySV) > 0 || len(s.magicByID) > 0 {
		s.mu.Unlock()
		return fmt.Errorf("xs: Clone of an instance holding live host-side MAGIC mirrors is not supported; clone right after loading, before code that attaches magic runs")
	}
	statesMu.Lock()
	cs := &state{
		p:             clone,
		fns:           append([]uintptr(nil), s.fns...),
		fnNames:       append([][]byte(nil), s.fnNames...),
		ptrs:          append([]uintptr(nil), s.ptrs...),
		cvAny:         make(map[uint64]uintptr, len(s.cvAny)),
		perlioFuncs:   make(map[uint32]uintptr, len(s.perlioFuncs)),
		perlioNextID:  s.perlioNextID,
		cvFn:          make(map[uint64]uintptr, len(s.cvFn)),
		magicBySV:     map[uint64]*magicRec{},
		magicByID:     map[uint32]*magicRec{},
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
	for k, v := range s.cvAny {
		cs.cvAny[k] = v
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

// ptrIDBase splits the PTR2IV/INT2PTR value space: values below it are
// plain integers cast through pointers (XS code legitimately packs small
// ints into void*, e.g. savestack indexes) and round-trip as themselves;
// values at or above it that reach encodePtr are real host pointers, which
// cross the 32-bit guest IV as (base + registry index). 64-bit heap
// pointers are far above the base; a real mapping below 1<<30 would be
// misclassified, which is documented as out of scope.
const ptrIDBase = 1 << 30

func (s *state) encodePtr(p uintptr) uint64 {
	if p < ptrIDBase {
		return uint64(p) // plain integer in pointer clothing
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, q := range s.ptrs { // dedup: identity must survive round trips
		if q == p {
			return ptrIDBase + uint64(i)
		}
	}
	s.ptrs = append(s.ptrs, p)
	return ptrIDBase + uint64(len(s.ptrs)-1)
}

func (s *state) decodePtr(id uint64) uintptr {
	if id < ptrIDBase {
		return uintptr(id)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ix := id - ptrIDBase
	if ix >= uint64(len(s.ptrs)) {
		return 0
	}
	return s.ptrs[ix]
}

var (
	statesMu sync.Mutex
	states   = map[*perl.Perl]*state{}

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

func unpinFrame(addr uintptr) { liveFrames.Delete(addr) }

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

func stateFor(p *perl.Perl) *state {
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
// and registers its boot function as the Perl sub <Module>::bootstrap — the
// exact contract a statically linked perl's XSLoader/DynaLoader resolve, so
// a stock `use Module;` (whose .pm calls XSLoader::load) boots the native
// module lazily at use time with no loader-side patching. module is the Perl
// package the .xs declared (e.g. "Compiler::Lexer").
func Load(p *perl.Perl, module, path string) error {
	libcInit()
	if libcErr != nil {
		return libcErr
	}
	s := stateFor(p)
	if err := p.SetNativeXSHandler(s.dispatch); err != nil {
		return err
	}
	if err := p.SetSetMagicHandler(s.magicSetFired); err != nil {
		return err
	}
	if err := p.SetMagicFreeHandler(s.magicFreed); err != nil {
		return err
	}
	if err := p.SetPPHookHandler(s.ppHook); err != nil {
		return err
	}
	if err := p.SetKeywordHandler(s.keywordHook); err != nil {
		return err
	}
	if err := p.SetPerlIOHandler(s.perlioHook); err != nil {
		return err
	}
	if err := p.SetDestructorHandler(s.dtorFired); err != nil {
		return err
	}
	if !s.cloneHookSet {
		s.cloneHookSet = true
		p.AddCloneHook(s.adoptClone)
	}
	vtableOnce.Do(buildVtable)

	lib, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return fmt.Errorf("dlopen %s: %w", path, err)
	}
	initFn, err := purego.Dlsym(lib, "__goperl_xs_init")
	if err != nil {
		return fmt.Errorf("%s is not a go-perl native XS module (no __goperl_xs_init): %w", path, err)
	}
	got, _, _ := purego.SyscallN(initFn, uintptr(unsafe.Pointer(vtable)))
	if uint32(got) != abiVersion {
		return fmt.Errorf("%s: XS SDK ABI mismatch (module rejected vtable v%d)", path, abiVersion)
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

	bootName := "boot_" + strings.ReplaceAll(module, "::", "__")
	bootFn, err := purego.Dlsym(lib, bootName)
	if err != nil {
		return fmt.Errorf("%s: no %s symbol: %w", path, bootName, err)
	}
	// The static-perl bootstrap contract: DynaLoader::bootstrap (which
	// XSLoader falls back to on a perl without dynamic loading) resolves
	// and calls &{"${module}::bootstrap"}.
	_, err = s.registerNative(module+"::bootstrap", bootFn)
	return err
}

// registerNative records fn and binds it as the Perl sub name via the
// guest's generic thunk, returning the guest CV token.
func (s *state) registerNative(name string, fn uintptr) (uint64, error) {
	s.mu.Lock()
	s.fns = append(s.fns, fn)
	s.fnNames = append(s.fnNames, append([]byte(name), 0))
	id := int32(len(s.fns) - 1)
	s.mu.Unlock()
	cvTok, err := s.p.XSHelperOp(opNewXS, uint64(uint32(id)), 0, name)
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
	if items >= maxStack-1 {
		return fail(fmt.Sprintf("native XS dispatch: %d arguments exceed the SDK stack (%d)", items, maxStack))
	}
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

	f := &cFrame{api: uintptr(unsafe.Pointer(vtable)), p: s.p, subname: subname, instShared: s.sharedRaw}
	f.items = int32(items)
	// Stack layout mirrors a real XS entry: a mark at offset 0, arguments
	// at st[1..items] (so ax == 1), sp at the last argument.
	f.marks[0] = 0
	f.markidx = 0
	for i := 0; i < items; i++ {
		f.st[1+i] = uint64(binary.LittleEndian.Uint32(req[12+4*i:]))
	}
	base := uintptr(unsafe.Pointer(&f.st[0]))
	f.psp = base + uintptr(items)*8
	fp := pinFrame(f)
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
	unpinFrame(fp)
	if f.failed != 0 {
		return fail(f.errString())
	}
	top := int(int64(f.psp-base) / 8)
	if top < 0 || top >= maxStack {
		return fail(fmt.Sprintf("native XS dispatch: corrupt stack pointer (top %d)", top))
	}
	n := top // ax == 1: return values are st[1..top]
	resp := make([]byte, 5+4*n)
	resp[0] = 1
	binary.LittleEndian.PutUint32(resp[1:], uint32(n))
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint32(resp[5+4*i:], uint32(f.st[1+i]))
	}
	return resp
}

// bootFrame builds a minimal frame for loader-driven native calls that
// happen outside an XSUB dispatch (svt_free teardown).
func (s *state) bootFrame() (*cFrame, uintptr) {
	f := &cFrame{api: uintptr(unsafe.Pointer(vtable)), p: s.p, instShared: s.sharedRaw}
	f.marks[0] = 0
	f.markidx = 0
	f.psp = uintptr(unsafe.Pointer(&f.st[0]))
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

	f := &cFrame{api: uintptr(unsafe.Pointer(vtable)), p: s.p, instShared: s.sharedRaw}
	f.hookOpTok = opTok
	f.items = int32(n)
	f.marks[0] = 0
	f.markidx = 0
	for i := 0; i < n; i++ {
		f.st[1+i] = uint64(binary.LittleEndian.Uint32(req[12+4*i:]))
	}
	f.psp = uintptr(unsafe.Pointer(&f.st[0])) + uintptr(n)*8
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
const (
	mgSize       = 40
	mgOffMore    = 0
	mgOffVirtual = 8
	mgOffType    = 18
	mgOffLen     = 20
	mgOffObj     = 24
	mgOffPtr     = 32
	vtblOffSet   = 8
	vtblOffFree  = 32
)

func (s *state) magicExt(f *cFrame, sv, obj uint64, how int32, vtbl, namePtr uintptr, namLen int64) uintptr {
	mg := cMalloc(mgSize)
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

	mgPtr := namePtr
	mgLen := int32(0)
	if namLen > 0 { // real sv_magicext semantics: positive namlen copies
		mgPtr = cMalloc(int(namLen) + 1)
		cMemcpy(mgPtr, namePtr, int(namLen))
		pokeBytes(mgPtr+uintptr(namLen), []byte{0})
		mgLen = int32(namLen)
	}
	var buf [mgSize]byte
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
		node := copyIn(mg, mgSize)
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
		s.callSvtFree(rec.svTok, mg)
		// Every entry beyond the anchor's held one manually retained obj.
		if obj != 0 && obj != rec.anchorObj {
			_, _ = s.p.XSHelperOp(opRefcntDec, obj, 0, "")
		} else if obj != 0 && obj == rec.anchorObj {
			rec.anchorObj = 0 // the anchor releases its own reference
		}
		if mgLen > 0 && mgPtr != 0 {
			cFree(mgPtr)
		}
		cFree(mg)
		mg = next
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
		node := copyIn(mg, mgSize)
		next := uintptr(binary.LittleEndian.Uint64(node[mgOffMore:]))
		typ := node[mgOffType]
		nodeVtbl := uintptr(binary.LittleEndian.Uint64(node[mgOffVirtual:]))
		if typ == byte(how) && (vtbl == 0 || nodeVtbl == vtbl) {
			obj := binary.LittleEndian.Uint64(node[mgOffObj:])
			mgLen := int32(binary.LittleEndian.Uint32(node[mgOffLen:]))
			mgPtr := uintptr(binary.LittleEndian.Uint64(node[mgOffPtr:]))
			s.callSvtFree(sv, mg)
			if obj != 0 && obj != rec.anchorObj {
				_, _ = s.p.XSHelperOp(opRefcntDec, obj, 0, "")
			}
			if mgLen > 0 && mgPtr != 0 {
				cFree(mgPtr)
			}
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
	vtable.ptrEncode = purego.NewCallback(func(fr, p uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		return uintptr(stateFor(f.p).encodePtr(p))
	})
	vtable.ptrDecode = purego.NewCallback(func(fr, id uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		return stateFor(f.p).decodePtr(uint64(id))
	})
	vtable.regXS = purego.NewCallback(func(fr, name, fn uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		s := stateFor(f.p)
		if _, err := s.registerNative(cString(name), fn); err != nil {
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
		cvTok, err := s.registerNative(cString(name), fn)
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
