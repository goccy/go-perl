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
// The SDK is a perl built without MULTIPLICITY: an XSUB has perl's
// signature there (void xsub(CV* cv)) and works on the frame the loader's
// entry makes current, which carries a real perl stack (base/sp/marks over
// SV tokens), and the loader
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
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/goccy/go-perl/internal"
)

// xsTrace logs every guest micro-op to stderr (GOPERL_XS_TRACE=1); the
// debugging aid of last resort when a native module misbehaves.
var xsTrace = os.Getenv("GOPERL_XS_TRACE") != ""

// abiVersion must match GOPERL_XS_ABI in sdk/include/perl.h.
const abiVersion = 35

// nativeHdrWords is the number of header words of a native-XSUB request
// (GOPERL_NATIVE_HDR_WORDS in perl-wasm): the function id, the CV, the
// item count, the interpreter's stack index of the first item (ax) and
// where its stack is; the items follow.
const nativeHdrWords = 5

// maxStackIndex bounds the stack index a request may name.
const maxStackIndex = 1 << 28

// plErrno is the interpreter variable id of errno (GOPERL_PL_ERRNO).
const plErrno = 60

// failedJump is the frame's failed value when the interpreter left an
// operation by a non-local exit (GOPERL_FAILED_JUMP in the SDK): the native
// call ends, and the interpreter goes on unwinding when it has returned.
const failedJump = 3

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
	opGvPtr         = 108
	opSvIsGVGP      = 110
	opMagicAttach   = 70
	opMagicUnattach = 72
	opPlvarSet      = 74
	opPPHookSet     = 98
	opMagicSetHook  = 133
	opSvCall        = 187
	opGvCall        = 188
	opInternal      = 204
)

// intLayout is the selector of opInternal that answers the address of the
// interpreter's table of the layout of its structures (GOPERL_INT_LAYOUT in
// perl-wasm): words of four bytes, layoutWords of them.
const (
	intLayout   = 0
	layoutWords = 32
)

// layoutNames names the words of the interpreter's layout table that a
// module is compared on (GOPERL_LAY_* in the SDK's perl.h).
var layoutNames = map[uint32]string{
	23: "offsetof(HE, hent_next)",
	24: "offsetof(HE, hent_hek)",
	25: "offsetof(HE, he_valu)",
	26: "sizeof(HE)",
	27: "offsetof(HEK, hek_hash)",
	28: "offsetof(HEK, hek_len)",
	29: "offsetof(HEK, hek_key)",
}

// gvCallCvNative is the GV_CALL selector that makes a CV call the native
// function of an id (GOPERL_GVCALL_CV_NATIVE in perl-wasm), and gvCallForm
// the one that answers a glob's format (GOPERL_GVCALL_FORM).
const (
	gvCallCvNative = 37
	gvCallForm     = 7
)

// The GV_PTR selectors of a glob's slots (GOPERL_GVPTR_* in perl-wasm).
const (
	gvPtrHV = 2
	gvPtrAV = 3
	gvPtrSV = 5
	gvPtrIO = 6
	gvPtrCV = 7
)

// svCallMagicIsValue is the SV_CALL selector that asks whether a magic type
// is value magic (GOPERL_SVCALL_MAGIC_IS_VALUE in perl-wasm).
const svCallMagicIsValue = 55

// opboxOffResult2 is where the box holds the second value of an operation
// that answers two (result2 of perl-wasm's struct goperl_opbox).
const opboxOffResult2 = 32

// opboxResult2 is the second value of the operation that has just answered,
// in the box at opbox of the instance's memory mem.
func opboxResult2(mem []byte, opbox uint64) uint64 {
	return binary.LittleEndian.Uint64(mem[opbox+opboxOffResult2:])
}

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
	// memBase and memSize are the interpreter's linear memory as native
	// code sees it: where it starts, and where the instance keeps its
	// size (internal.Perl.MemorySizeWord), which native code reads at
	// every use: the memory grows while the interpreter runs, during a
	// native call too, and never moves (the copy-on-write snapshot path),
	// so an address is the interpreter's exactly when it lies below the
	// size as it is then. The range reserved past the size is not
	// readable.
	memBase uintptr
	memSize uintptr
	// The perl stack of a pp function (ppFrame set): st is a copy of the
	// interpreter's stack and marks of its mark stack, gmarkLo the
	// interpreter's index of marks[0]. The SDK keeps what the copy was
	// when the two were last in line (stSeen, which it allocates and
	// frees) and whether the interpreter has moved on since (stStale); at
	// the end of the function it leaves what the function made of the
	// stack in ppState, a C heap block the loader frees.
	ppFrame    int32
	stStale    int32
	gmarkLo    int32
	markSeen   int32
	stSeen     uintptr
	stSeenN    int64
	marksSeen  [maxMarks]int32
	ppState    uintptr
	ppStateLen int64
	// nativeCall is set by the SDK in the frame of a native call the
	// interpreter made, as opposed to a function the loader runs on the
	// side (a destructor, a magic hook).
	nativeCall int32
	_          int32
	// entryJB is the SDK's: the catch point of its XSUB entry while no
	// XSUB has taken that activation over.
	entryJB uintptr
	// opbox is the address of the instance's operation box in its linear
	// memory (internal.Perl.XSOpBox): the SDK writes an operation's
	// operands there and has xsOpBox run it, and reads its results there.
	opbox uint64
	// tmpMore is the SDK's last block of SV** result slots beyond tmp
	// (goperl_tmpblock_t: the block before it, then the slots), which the
	// SDK mallocs and releaseStack frees with the frame.
	tmpMore uintptr
	// The stack of an XSUB is the interpreter's by index: gstack is where
	// the interpreter's stack is in its memory, and below the number of
	// its values under the arguments (ax), which the SDK reads into
	// st[0..below-1] when native code may index them, setting below to 0.
	gstack uint64
	below  int32
	// markv is the block the SDK moved the marks to once they outgrew
	// marks (markcap of them), a C heap block releaseStack frees; 0 before.
	markcap int32
	markv   uintptr

	// Go-only:
	p *internal.Perl
}

// newFrame returns a frame of this instance with no stack yet.
func (s *state) newFrame() *cFrame {
	f := &cFrame{api: uintptr(unsafe.Pointer(vtable)), p: s.p, instShared: s.sharedRaw, opbox: s.opbox}
	f.memSize = uintptr(unsafe.Pointer(s.p.MemorySizeWord()))
	if mem := s.p.RawMemory(); cap(mem) > 0 {
		f.memBase = uintptr(unsafe.Pointer(unsafe.SliceData(mem)))
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
	// v20: a pointer of the module's own in the buffer slot of an SV.
	pvHost uintptr
	// v27: run the operation in the frame's operation box; 1 when it
	// returned, 0 when it failed (the frame says how).
	xsOpBox uintptr
}

// magicRec tracks one guest SV's host-side MAGIC chain.
type magicRec struct {
	svTok    uint64
	head     uintptr // C-heap MAGIC* chain head
	anchorID uint32
	// freeing is set while the interpreter frees the anchor (magicFreed).
	// perl's mg_free keeps the anchor on the SV's chain while its free
	// runs, and frees it after: a module that empties its chain from a
	// svt_free then (SvMAGIC_set to NULL, sv_unmagic of the last MAGIC)
	// must not have the anchor taken off under that free. Nothing else
	// tells that the free is running: the interpreter's chain looks as it
	// does at any other time, and the maps stay for mg_find in the hooks.
	freeing bool
}

func init() {
	// Back (*internal.Perl).AddXSDir with this package's directory loader.
	internal.RegisterXSDirLoader(loadDir)
	// Back DynaLoader's dl_* functions (Config.NativeModules).
	internal.RegisterXSDynamicLoader(internal.XSDynamicLoader{
		Open:    openLibrary,
		Symbol:  librarySymbol,
		Install: installXSub,
		Release: releaseState,
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
		// native code runs in the instance from here on, on its thread
		// (internal.Perl.KeepOnOneThread): the module is opened there too
		p.KeepOnOneThread()
		p.OnThread(func() { err = load(p, module, filepath.Join(dir, e.Name())) })
		if err != nil {
			return fmt.Errorf("load native module %s: %w", module, err)
		}
	}
	return nil
}

// state is the per-instance loader state.
type state struct {
	p  *internal.Perl
	mu sync.Mutex
	// libs and syms are what DynaLoader's dl_load_file and dl_find_symbol
	// handed to Perl code: the library handles and the addresses of the
	// symbols themselves, as perl's dl_dlopen.xs hands them (SPVM passes
	// a symbol's address to native code of its own, which calls it). The
	// loader refuses a handle or an address it did not hand out.
	libs    map[uintptr]struct{}
	syms    map[uintptr]struct{}
	fns     []uintptr // fnID -> native XSUB pointer
	fnNames [][]byte  // fnID -> NUL-terminated sub name (frame.subname points here)
	// per-CV storage: CvXSUBANY slots (8 bytes of C heap each, written by
	// the module through the vtable) and the native fn behind each CV.
	cvAny map[uint64]uintptr
	// pvHost holds, per SV, the pointer a module put in the SV's
	// string-buffer slot when it is not one into the interpreter's memory
	// and the room the module says the buffer has (two words of C heap
	// each; the interpreter's slot holds GOPERL_PV_HOST). An entry outlives
	// its SV and counts again only for an SV at the same address that is
	// given such a pointer, which overwrites it.
	pvHost map[uint64]uintptr
	// opbox is the address of the instance's operation box in its linear
	// memory (see cFrame.opbox): asked of the interpreter once, as the
	// interrupt flag's is, since only the interpreter knows it.
	opbox uint64

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
	xsInvoke     uintptr // __goperl_xs_invoke
	xsEpilogue   uintptr // __goperl_xs_epilogue
	perlioInvoke uintptr // __goperl_perlio_invoke
	hostInvoke   uintptr // __goperl_host_invoke
	// sharedClone and sharedRelease are the SDK's __goperl_shared_clone,
	// which gives the block of a copy of the instance what modules
	// registered with the interpreter, and __goperl_shared_release, which
	// lets go of what the block shares with copies when the instance goes.
	sharedClone   uintptr
	sharedRelease uintptr

	// cloneHookSet records that Load registered adoptClone with the
	// instance (once, however many modules load).
	cloneHookSet bool

	// nativeThread is the thread running native code for the instance
	// right now (0: none), for a signal to interrupt it from another
	// goroutine (see native_interrupt.go): which thread the instance's
	// goroutine is on is known only on that goroutine.
	nativeThread atomic.Uint64
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
// magicShares). The SDK's block of the instance gets what modules
// registered with the interpreter (op checkers, keyword plugins, ...: see
// __goperl_shared_clone); scope bookkeeping starts empty: the native
// activations it belongs to exist only in the original.
func (s *state) adoptClone(clone *internal.Perl) error {
	if clone.ForkChild() {
		return s.adoptForkChild(clone)
	}
	if err := memoryAboveCeiling(clone); err != nil {
		return err
	}
	from, to := spanOf(s.p.RawMemory()), spanOf(clone.RawMemory())
	s.mu.Lock()
	statesMu.Lock()
	cs := &state{
		p:             clone,
		libs:          maps.Clone(s.libs),
		syms:          maps.Clone(s.syms),
		fns:           append([]uintptr(nil), s.fns...),
		fnNames:       append([][]byte(nil), s.fnNames...),
		cvAny:         make(map[uint64]uintptr, len(s.cvAny)),
		pvHost:        make(map[uint64]uintptr, len(s.pvHost)),
		perlioFuncs:   make(map[uint32]uintptr, len(s.perlioFuncs)),
		perlioNextID:  s.perlioNextID,
		cvFn:          make(map[uint64]uintptr, len(s.cvFn)),
		nextMagicID:   s.nextMagicID,
		magicBySV:     make(map[uint64]*magicRec, len(s.magicBySV)),
		magicByID:     make(map[uint32]*magicRec, len(s.magicByID)),
		hookFns:       make(map[int32]uintptr, len(s.hookFns)),
		perOpFns:      make(map[uint64]uintptr, len(s.perOpFns)),
		sharedRaw:     cMalloc(sharedBlockSize),
		opbox:         s.opbox, // the copy's memory has it where the original's has it
		sharedClone:   s.sharedClone,
		sharedRelease: s.sharedRelease,
		keywordInvoke: s.keywordInvoke,
		ppInvoke:      s.ppInvoke,
		dtorInvoke:    s.dtorInvoke,
		svtInvoke:     s.svtInvoke,
		xsInvoke:      s.xsInvoke,
		xsEpilogue:    s.xsEpilogue,
		perlioInvoke:  s.perlioInvoke,
		hostInvoke:    s.hostInvoke,
		cloneHookSet:  true,
	}
	pokeBytes(cs.sharedRaw, make([]byte, sharedBlockSize))

	if s.sharedClone != 0 {
		if rc, _, _ := purego.SyscallN(s.sharedClone, cs.sharedRaw, s.sharedRaw); int32(rc) != 0 {
			statesMu.Unlock()
			s.mu.Unlock()
			return errors.New("xs: out of memory copying the native modules' state")
		}
	}
	for cv, slot := range s.cvAny {
		own := cMalloc(8)
		cMemcpy(own, slot, 8)
		cs.cvAny[cv] = own
	}
	for sv, slot := range s.pvHost {
		own := cMalloc(pvHostSize)
		cMemcpy(own, slot, pvHostSize)
		cs.pvHost[sv] = own
	}
	for sv, rec := range s.magicBySV {
		own := &magicRec{svTok: sv, head: copyMagicChain(rec.head, from, to), anchorID: rec.anchorID}
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
	if err := clone.SetHostCallHandler(cs.hostCall); err != nil {
		return err
	}
	if err := clone.SetProcessMemoryHandler(readProcessMemory); err != nil {
		return err
	}
	return clone.SetDestructorHandler(cs.dtorFired)
}

// readProcessMemory reads memory of this process that is not the
// interpreter's, for unpack 'p' and 'P' of an address native code made (a
// buffer a C function returned): n bytes at addr, or with n < 0 the bytes up
// to the NUL there. As on perl, the address is trusted: memory that is not
// there faults.
func readProcessMemory(addr uint64, n int64) ([]byte, error) {
	if addr == 0 {
		return nil, errors.New("read of the memory of the process at address 0")
	}
	if n < 0 {
		l, _, _ := purego.SyscallN(strlenFn, uintptr(addr))
		n = int64(l)
	}
	return copyIn(uintptr(addr), int(n)), nil
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
			libs:        map[uintptr]struct{}{},
			syms:        map[uintptr]struct{}{},
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
		// the box is required (loaderState says so when there is none)
		if box, err := p.XSOpBox(); err == nil {
			s.opbox = box
		}
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
	lib, err := s.open(path, false)
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
	setHostLocale(p.Environment())
	if err := memoryAboveCeiling(p); err != nil {
		return nil, err
	}
	s := stateFor(p)
	if s.opbox == 0 {
		// the second value of an operation that answers two is only ever
		// in the box
		return nil, errors.New("xs: the interpreter has no operation box")
	}
	return s, nil
}

// memoryAboveCeiling refuses native code to an instance whose memory does
// not lie at or above its ceiling in this process. Native code reads some
// of the interpreter's structures where they lie (the entries of hashes),
// and a pointer such a structure holds may be an address of the memory or,
// where a module stored one there, a pointer of this process into the
// memory: the SDK tells them apart by the number alone (goperl_he_ptr).
// Every address of the memory is below its ceiling (it never grows past
// it), and every pointer into it is at or above its base (it never moves);
// with the base at or above the ceiling the two never meet, so the rule is
// total. The reservations of the 64-bit hosts are mapped above their
// ceilings (Linux maps them below the top of the address space; macOS
// gives one past the zero page and the program a place at or above its
// length, which is longer than the ceiling), so this refuses nothing
// there; but no system promises it, and this makes it a condition.
func memoryAboveCeiling(p *internal.Perl) error {
	mem := p.RawMemory()
	base := uintptr(unsafe.Pointer(unsafe.SliceData(mem)))
	if ceiling := p.MemoryCeiling(); cap(mem) == 0 || uint64(base) < ceiling {
		return fmt.Errorf("xs: the interpreter's memory lies at %#x, below its ceiling of %#x bytes: native code cannot tell its addresses from pointers into it", base, ceiling)
	}
	return nil
}

// layoutAgrees compares the layout of the interpreter's structures that
// the module at path was compiled with (its __goperl_xs_layout: pairs of a
// word of the interpreter's layout table and the value the module has for
// it) with the interpreter's own, and refuses a module that disagrees: it
// reads and writes those structures where they lie. A module built before
// the SDK read any of them in place has no such table, and nothing to
// compare.
func (s *state) layoutAgrees(lib uintptr, path string) error {
	fn, err := purego.Dlsym(lib, "__goperl_xs_layout")
	if err != nil {
		return nil
	}
	np := cMalloc(8)
	defer cFree(np)
	pokeU64(np, 0)
	pairs, _, _ := purego.SyscallN(fn, np)
	n := int(binary.LittleEndian.Uint32(copyIn(np, 4)))
	if pairs == 0 || n <= 0 || n > layoutWords {
		return fmt.Errorf("%s: __goperl_xs_layout describes no layout", path)
	}
	want := copyIn(pairs, 8*n)
	at, err := s.p.XSHelperOp(opInternal, 0, 0, intLayout, "")
	if err != nil {
		return fmt.Errorf("%s: the interpreter's layout: %w", path, err)
	}
	mem := s.p.RawMemory()
	if at == 0 || at > uint64(len(mem)) || uint64(len(mem))-at < 4*layoutWords {
		return fmt.Errorf("%s: the interpreter did not describe its layout", path)
	}
	for i := 0; i < n; i++ {
		word := binary.LittleEndian.Uint32(want[8*i:])
		val := binary.LittleEndian.Uint32(want[8*i+4:])
		name, known := layoutNames[word]
		if !known || word >= layoutWords {
			return fmt.Errorf("%s: the module was compiled with a layout of word %d, which this loader does not know", path, word)
		}
		if got := binary.LittleEndian.Uint32(mem[at+4*uint64(word):]); got != val {
			return fmt.Errorf("%s: the module was compiled with %s = %d, the interpreter has %d", path, name, val, got)
		}
	}
	return nil
}

// open dlopens a native XS module built against the SDK, injects the
// vtable, and picks up the SDK entry points the module carries. global
// makes the library's symbols available to the libraries loaded after it
// (RTLD_GLOBAL), as DynaLoader's dl_load_file does when a module's
// dl_load_flags ask for it; otherwise they stay the library's own.
func (s *state) open(path string, global bool) (lib uintptr, err error) {
	p := s.p
	if err := p.SetNativeXSHandler(s.dispatch); err != nil {
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
	if err := p.SetHostCallHandler(s.hostCall); err != nil {
		return 0, err
	}
	if err := p.SetProcessMemoryHandler(readProcessMemory); err != nil {
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

	mode := purego.RTLD_NOW | purego.RTLD_LOCAL
	if global {
		mode = purego.RTLD_NOW | purego.RTLD_GLOBAL
	}
	lib, err = purego.Dlopen(path, mode)
	if err != nil {
		return 0, fmt.Errorf("dlopen %s: %w", path, err)
	}
	initFn, err := purego.Dlsym(lib, "__goperl_xs_init")
	if err != nil {
		return lib, &plainLibraryError{path: path, err: err}
	}
	got, _, _ := purego.SyscallN(initFn, uintptr(unsafe.Pointer(vtable)))
	if uint32(got) != abiVersion {
		return 0, fmt.Errorf("%s: XS SDK ABI mismatch (module rejected vtable v%d)", path, abiVersion)
	}
	if err := s.layoutAgrees(lib, path); err != nil {
		return 0, err
	}
	// The SDK's croak-guarded entry points for pp hooks and save-stack
	// destructors (weak, so any SDK .so provides them).
	if fn, err := purego.Dlsym(lib, "__goperl_pp_invoke"); err == nil {
		s.mu.Lock()
		s.ppInvoke = fn
		s.mu.Unlock()
	}
	if fn, err := purego.Dlsym(lib, "__goperl_shared_clone"); err == nil {
		s.mu.Lock()
		s.sharedClone = fn
		s.mu.Unlock()
	}
	if fn, err := purego.Dlsym(lib, "__goperl_shared_release"); err == nil {
		s.mu.Lock()
		s.sharedRelease = fn
		s.mu.Unlock()
	}
	if fn, err := purego.Dlsym(lib, "__goperl_svt_invoke"); err == nil {
		s.mu.Lock()
		s.svtInvoke = fn
		s.mu.Unlock()
	}
	if fn, err := purego.Dlsym(lib, "__goperl_xs_invoke"); err == nil {
		s.mu.Lock()
		s.xsInvoke = fn
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
	if fn, err := purego.Dlsym(lib, "__goperl_host_invoke"); err == nil {
		s.mu.Lock()
		s.hostInvoke = fn
		s.mu.Unlock()
	}
	return lib, nil
}

// dlGlobalSymbols is the bit of dl_load_file's flags that makes a library's
// symbols global (perl's dl_dlopen.xs: RTLD_GLOBAL).
const dlGlobalSymbols = 0x01

// plainLibraryError is the failure of open for a shared library that is
// not a native XS module: it has none of the SDK's entry points. The
// library is loaded all the same (lib of open's result).
type plainLibraryError struct {
	path string
	err  error
}

func (e *plainLibraryError) Error() string {
	return fmt.Sprintf("%s is not a go-perl native XS module (no __goperl_xs_init): %v", e.path, e.err)
}

func (e *plainLibraryError) Unwrap() error { return e.err }

// openLibrary is DynaLoader's dl_load_file: it loads the library at path
// and returns its handle, as perl's does. Like dlopen, which perl's
// dl_load_file is, it loads any shared library, not only a native module:
// code that looks for a C library's functions through DynaLoader
// (FFI::CheckLib, for one) loads the library to find its symbols. Only a
// module is initialized for the SDK. flags are dl_load_file's: with bit
// 0x01 (a module's dl_load_flags returning it) the library's symbols are
// global, so that the modules loaded after it link against it — XS::Loader
// loads every module that way, and a module built on another one's library
// (Test::Catch on XS::libcatch) needs it.
func openLibrary(p *internal.Perl, path string, flags int) (uint64, error) {
	if p.ForkChild() {
		return 0, errors.New(unavailable(path, 0))
	}
	s, err := loaderState(p)
	if err != nil {
		return 0, err
	}
	// dl_load_file is a Go function the program calls, which runs off the
	// instance's thread (internal.Perl's callGo); loading runs the
	// library's initializers, native code, on the thread.
	var lib uintptr
	p.OnThread(func() { lib, err = s.open(path, flags&dlGlobalSymbols != 0) })
	var plain *plainLibraryError
	if errors.As(err, &plain) {
		err = nil
	}
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.libs[lib] = struct{}{}
	return uint64(lib), nil
}

// librarySymbol is DynaLoader's dl_find_symbol: the address of the symbol
// in the library whose handle dl_load_file returned.
func librarySymbol(p *internal.Perl, lib uint64, name string) (uint64, error) {
	if p.ForkChild() {
		return 0, errors.New(unavailable(name, 0))
	}
	s := stateFor(p)
	handle := uintptr(lib)
	s.mu.Lock()
	_, known := s.libs[handle]
	s.mu.Unlock()
	if !known {
		return 0, fmt.Errorf("unknown library handle %#x", lib)
	}
	fn, err := purego.Dlsym(handle, name)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syms[fn] = struct{}{}
	return uint64(fn), nil
}

// installXSub is DynaLoader's dl_install_xsub: the function at the address
// dl_find_symbol returned becomes the Perl sub name.
func installXSub(p *internal.Perl, name string, sym uint64) error {
	if p.ForkChild() {
		return errors.New(unavailable(name, 0))
	}
	s := stateFor(p)
	fn := uintptr(sym)
	s.mu.Lock()
	_, known := s.syms[fn]
	s.mu.Unlock()
	if !known {
		return fmt.Errorf("unknown symbol address %#x", sym)
	}
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
	cvTok, err := s.p.XSHelperOp(opNewXS, uint64(uint32(id)), unnamed, 0, name)
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
	_, err := s.p.XSHelperOp(opGvCall, cv, uint64(uint32(id)), gvCallCvNative, "")
	return err
}

// The requests and responses of the reserved callback method ids are made
// of 64-bit words (an address of the interpreter's memory is one, whole):
// a request is a string of words (a keyword's ends in the bytes of the
// word), a response a status byte and then words, or the message of a
// failure (see perl-wasm's perl.cc).

// wordAt is the i-th 64-bit word of a request.
func wordAt(req []byte, i int) uint64 {
	return binary.LittleEndian.Uint64(req[8*i:])
}

// dispatch handles one native XSUB call: request [fn_id][cv][items]
// [svs...]; response [1][nret][svs...] or [0]+message.
func (s *state) dispatch(req []byte) []byte {
	fail := func(msg string) []byte { return append([]byte{0}, msg...) }
	if len(req) < 3*8 {
		return fail("native XS dispatch: short payload")
	}
	id := int32(int64(wordAt(req, 0)))
	cvTok := wordAt(req, 1)
	nitems := wordAt(req, 2)
	if len(req) < nativeHdrWords*8 || nitems > uint64(len(req)/8) || len(req) < 8*(nativeHdrWords+int(nitems)) {
		return fail("native XS dispatch: truncated payload")
	}
	items := int(nitems)
	ax := int(int64(wordAt(req, 3)))
	if ax < 1 || ax > maxStackIndex {
		return fail(fmt.Sprintf("native XS dispatch: no stack index %d", ax))
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
	// The stack is the interpreter's by index, as a real XS entry's is:
	// the arguments at st[ax..ax+items-1], ax being the interpreter's
	// index of the first, a mark at ax-1, sp at the last argument. The
	// values below the arguments are the interpreter's, at gstack; the SDK
	// reads them in when native code may look at them.
	f.marks[0] = int32(ax - 1)
	f.markidx = 0
	f.initStack(ax - 1 + items)
	for i := 0; i < items; i++ {
		f.setStack(ax+i, wordAt(req, nativeHdrWords+i))
	}
	f.psp = f.stBase + uintptr(ax-1+items)*8
	f.gstack = wordAt(req, 4)
	f.below = int32(ax)
	fp := pinFrame(f)
	defer unpinFrame(fp)
	// The XSUB call and the epilogue that closes its activation must run
	// on ONE OS thread: the SDK's current-frame state is per-thread, and
	// the Go scheduler is otherwise free to migrate the goroutine between
	// the two purego calls.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	s.mu.Lock()
	inv, epi := s.xsInvoke, s.xsEpilogue
	s.mu.Unlock()
	if inv == 0 {
		return fail("native XS dispatch: no native module provides the entry point")
	}
	prevThread, prevInterrupt, entered := s.enterNative()
	// The SDK's entry makes the frame current before the XSUB's first
	// statement: an XSUB takes no context argument, as in perl without
	// MULTIPLICITY, and works on the current frame.
	purego.SyscallN(inv, fp, fn, uintptr(cvTok))
	// Close any activation levels the XSUB left open (xsubpp PPCODE exits
	// with a bare return, skipping XSRETURN's leave).
	if epi != 0 {
		purego.SyscallN(epi, fp)
	}
	s.leaveNative(prevThread, prevInterrupt, entered)
	if f.xerrno != 0 {
		// What the C code left in errno is what $! says after the call.
		if _, err := s.p.XSHelperOp(opPlvarSet, plErrno, uint64(f.xerrno-1), 0, ""); err != nil {
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
	n := top - (ax - 1) // the return values are st[ax..top]
	if n < 0 {
		return fail(fmt.Sprintf("native XS dispatch: corrupt stack pointer (top %d below ax %d)", top, ax))
	}
	resp := make([]byte, 1+8+8*n)
	resp[0] = 1
	binary.LittleEndian.PutUint64(resp[1:], uint64(n))
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint64(resp[1+8+8*i:], f.stackAt(ax+i))
	}
	if xsTrace {
		fmt.Fprintf(os.Stderr, "[xsub] %s returns %d: % x\n", cString(subname), n, resp[1+8:])
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

// ppMarksMax is the number of marks the guest hands over with the stack of
// a pp function (GOPERL_PP_MARKS_MAX in perl-wasm): the top ones.
const ppMarksMax = 32

// ppHook runs a module's pp function for one guest op execution (reserved
// callback method -3). Request: [op][op_type] and the perl stack as the
// interpreter has it: [depth][top mark][n][n marks, the top ones from the
// lowest][depth values from index 1]. Response: [1][next_op] and, when the
// function changed its copy of the stack, what it became (the SDK builds
// that part: see goperl_pp_state_build); or [0]+message.
func (s *state) ppHook(req []byte) []byte {
	fail := func(msg string) []byte { return append([]byte{0}, msg...) }
	if len(req) < 5*8 {
		return fail("pp hook: short payload")
	}
	opTok := wordAt(req, 0)
	opType := int32(int64(wordAt(req, 1)))
	udepth := wordAt(req, 2)
	topMark := int32(int64(wordAt(req, 3)))
	unmarks := wordAt(req, 4)
	if unmarks > ppMarksMax || udepth > uint64(len(req)/8) ||
		len(req) < 8*(5+int(unmarks)+int(udepth)) {
		return fail("pp hook: truncated payload")
	}
	depth, nmarks := int(udepth), int(unmarks)
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
		fmt.Fprintf(os.Stderr, "[pphook] op=%#x type=%d depth=%d marks=%d\n", opTok, opType, depth, topMark)
	}

	f := s.newFrame()
	f.hookOpTok = opTok
	f.items = int32(depth)
	f.ppFrame = 1
	f.initStack(depth)
	for i := 0; i < depth; i++ {
		f.setStack(1+i, wordAt(req, 5+nmarks+i))
	}
	f.psp = f.stBase + uintptr(depth)*8
	// marks[0] is the mark below the ones handed over (the interpreter's
	// mark stack starts with an entry that is never a mark)
	f.gmarkLo = topMark - int32(nmarks)
	for i := 0; i < nmarks; i++ {
		f.marks[1+i] = int32(int64(wordAt(req, 5+i)))
	}
	f.markidx = int32(nmarks)
	fp := pinFrame(f)
	runtime.LockOSThread()
	prevThread, prevInterrupt, entered := s.enterNative()
	next, _, _ := purego.SyscallN(invoke, fp, fn)
	s.leaveNative(prevThread, prevInterrupt, entered)
	runtime.UnlockOSThread()
	unpinFrame(fp)
	var stackState []byte
	if f.ppState != 0 {
		stackState = copyIn(f.ppState, int(f.ppStateLen))
		cFree(f.ppState)
		f.ppState = 0
	}
	f.releaseStack()
	if f.failed == failedThrown {
		return []byte{respThrow}
	}
	if f.failed != 0 {
		if xsTrace {
			fmt.Fprintf(os.Stderr, "[pphook] FAILED: %s\n", f.errString())
		}
		return fail(f.errString())
	}
	if xsTrace {
		fmt.Fprintf(os.Stderr, "[pphook] -> next=%#x state=%d bytes\n", next, len(stackState))
	}
	resp := make([]byte, 1+8, 1+8+len(stackState))
	resp[0] = 1
	binary.LittleEndian.PutUint64(resp[1:], uint64(next))
	return append(resp, stackState...)
}

// hostCall services reserved method -9: one call of native code through
// the host-call channel (the SDK's __goperl_host_invoke runs it under a
// croak guard). req is [kind] and the kind's words; the response is [1]
// and the words the SDK answered with, or [0]+message.
// boxedOp describes the operation in the instance's box (perl-wasm's struct
// goperl_opbox: a, b and c at 0, 8 and 16, the op number at 40), for the
// message of an operation that failed.
func (s *state) boxedOp() string {
	mem := s.p.RawMemory()
	at := s.opbox
	if at == 0 || at+48 > uint64(len(mem)) {
		return "(no box)"
	}
	w := mem[at:]
	return fmt.Sprintf("%d (a=%#x b=%#x c=%#x)", int32(binary.LittleEndian.Uint32(w[40:])),
		binary.LittleEndian.Uint64(w[0:]), binary.LittleEndian.Uint64(w[8:]), binary.LittleEndian.Uint64(w[16:]))
}

func (s *state) hostCall(req []byte) []byte {
	s.mu.Lock()
	inv := s.hostInvoke
	s.mu.Unlock()
	if inv == 0 {
		return append([]byte{0}, "host call: no native module provides the entry point"...)
	}
	const respCap = 64
	reqC := cMalloc(len(req) + 1)
	pokeBytes(reqC, req)
	respC := cMalloc(respCap)
	f, fp := s.bootFrame()
	// One OS thread for the whole call: the SDK's current frame is per
	// thread.
	runtime.LockOSThread()
	prevThread, prevInterrupt, entered := s.enterNative()
	n, _, _ := purego.SyscallN(inv, fp, reqC, uintptr(len(req)), respC, respCap)
	s.leaveNative(prevThread, prevInterrupt, entered)
	runtime.UnlockOSThread()
	unpinFrame(fp)
	var out []byte
	switch rlen := int(int32(n)); {
	case f.failed == failedThrown:
		out = []byte{respThrow}
	case f.failed != 0:
		out = append([]byte{0}, f.errString()...)
	case rlen < 0 || rlen > respCap:
		out = append([]byte{0}, "host call: bad response"...)
	default:
		out = append([]byte{1}, copyIn(respC, rlen)...)
	}
	f.releaseStack()
	cFree(reqC)
	cFree(respC)
	return out
}

// dtorFired runs a save-stack destructor the guest scope pop released
// (reserved callback method -4). A destructor that died answers as an XSUB
// that died does: the interpreter dies with its message, or with the value
// it threw.
func (s *state) dtorFired(id uint32) []byte {
	s.mu.Lock()
	invoke := s.dtorInvoke
	s.mu.Unlock()
	if invoke == 0 {
		return []byte{1}
	}
	f, fp := s.bootFrame()
	runtime.LockOSThread()
	prevThread, prevInterrupt, entered := s.enterNative()
	purego.SyscallN(invoke, fp, uintptr(id))
	s.leaveNative(prevThread, prevInterrupt, entered)
	runtime.UnlockOSThread()
	unpinFrame(fp)
	defer f.releaseStack()
	switch {
	case f.failed == failedThrown:
		return []byte{respThrow}
	case f.failed != 0:
		return append([]byte{0}, f.errString()...)
	}
	return []byte{1}
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
	vtblOffGet   = 0
	vtblOffSet   = 8
	vtblOffLen   = 16
	vtblOffClear = 24
	vtblOffFree  = 32
	vtblOffCopy  = 40
	vtblOffLocal = 56
	mgfLocal     = 0x20 // MGf_LOCAL
	// MGf_REFCOUNTED: the MAGIC holds a reference on its mg_obj.
	mgfRefcounted = 0x02
	hefSVKey      = -2 // HEf_SVKEY: the name of a MAGIC is an SV
)

// struct Perl_custom_infix of sdk/include/perl.h: prec (an enum, 4 bytes)
// at 0, parse at 8, build_op at 16.
const (
	infixSize     = 24
	infixOffPrec  = 0
	infixOffParse = 8
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

// releaseState gives up the loader state of an instance that closes. Its
// interpreter was destroyed the way perl's main() destroys one, which frees
// no value that is left (a global with native magic): its memory goes with
// the instance, as a process's memory goes with the process. So it no
// longer holds the copies of MAGIC it shares with the instances it was
// copied from or to, and no svt_free runs for them here: the last instance
// that frees its copy runs it. So it is with a function given to
// call_atexit that it shares and did not run (see __goperl_shared_release).
func releaseState(p *internal.Perl) {
	statesMu.Lock()
	s := states[p]
	delete(states, p)
	statesMu.Unlock()
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sharedRelease != 0 {
		purego.SyscallN(s.sharedRelease, s.sharedRaw)
	}
	for _, rec := range s.magicBySV {
		for mg := rec.head; mg != 0; {
			node := copyIn(mg, mgAllocSize)
			lastMagicHolder(node)
			mg = uintptr(binary.LittleEndian.Uint64(node[mgOffMore:]))
		}
	}
	flushNativeStdio()
}

// flushNativeStdio writes out what native code left in the C library's
// stdio buffers (printf to stdout): perl's exit flushes them, and a gperl
// program ends without the C library's exit. The interpreter's own output
// went out before, in its teardown, as on perl.
func flushNativeStdio() {
	fflushOnce.Do(func() {
		fflushFn, _ = purego.Dlsym(purego.RTLD_DEFAULT, "fflush")
	})
	if fflushFn != 0 {
		purego.SyscallN(fflushFn, 0)
	}
}

var (
	fflushOnce sync.Once
	fflushFn   uintptr
)

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
		if _, err := s.p.XSHelperOp(opRefcntInc, uint64(namePtr), 0, 0, ""); err != nil {
			f.fail("sv_magicext: " + err.Error())
			return 0
		}
	}
	takesObj := magicTakesObj(sv, obj, how) && !s.globHolds(obj, sv)
	var buf [mgAllocSize]byte
	binary.LittleEndian.PutUint64(buf[mgOffMore:], uint64(oldHead))
	binary.LittleEndian.PutUint64(buf[mgOffVirtual:], uint64(vtbl))
	buf[mgOffType] = byte(how)
	if takesObj {
		buf[mgOffFlags] = mgfRefcounted
	}
	binary.LittleEndian.PutUint32(buf[mgOffLen:], uint32(mgLen))
	binary.LittleEndian.PutUint64(buf[mgOffObj:], obj)
	binary.LittleEndian.PutUint64(buf[mgOffPtr:], uint64(mgPtr))
	pokeBytes(mg, buf[:])

	// A vtbl with svt_get, svt_set, svt_len, svt_clear or svt_copy needs
	// the guest anchor to have that hook too, so that the interpreter
	// reports the access (through the host-call channel, where a hook
	// that croaks dies in the interpreter). Anchors start without: most
	// mirrors only carry svt_free, and a hook on the anchor costs a host
	// round trip on every such access to the SV.
	var hooks uint64
	if vtbl != 0 {
		fns := copyIn(vtbl, 8*8)
		for bit, off := range []int{vtblOffGet, vtblOffSet, vtblOffLen, vtblOffClear, vtblOffCopy} {
			if binary.LittleEndian.Uint64(fns[off:]) != 0 {
				hooks |= 1 << bit
			}
		}
	}
	wantSet := hooks != 0
	if xsTrace {
		fmt.Fprintf(os.Stderr, "[magicExt] sv=%#x how=%d vtbl=%#x hooks=%#x first=%v\n", sv, how, vtbl, hooks, first)
	}

	if first {
		s.mu.Lock()
		s.nextMagicID++
		rec = &magicRec{svTok: sv, head: mg, anchorID: s.nextMagicID}
		s.magicBySV[sv] = rec
		s.magicByID[rec.anchorID] = rec
		s.mu.Unlock()
		// The guest anchor holds the SV<->host-chain link, and nothing
		// else: what a MAGIC holds is the MAGIC's (below).
		if _, err := s.p.XSHelperOp(opMagicAttach, sv, 0, uint64(rec.anchorID), ""); err != nil {
			f.fail("sv_magicext: " + err.Error())
			return 0
		}
		if wantSet {
			r, err := s.p.XSHelperOp(opMagicSetHook, sv, hooks, 0, "")
			if err != nil {
				f.fail("sv_magicext: " + err.Error())
			}
			if xsTrace {
				fmt.Fprintf(os.Stderr, "[magicExt] set-hook upgrade -> %d\n", r)
			}
		}
	} else {
		if wantSet {
			if _, err := s.p.XSHelperOp(opMagicSetHook, sv, hooks, 0, ""); err != nil {
				f.fail("sv_magicext: " + err.Error())
			}
		}
		s.mu.Lock()
		rec.head = mg
		s.mu.Unlock()
	}
	if takesObj {
		// sv_magicext takes a reference on obj for the MAGIC
		// (MGf_REFCOUNTED). Whoever frees the MAGIC gives it up: mg_free
		// here (magicFreed, magicDel), or a module that takes the MAGIC off
		// the chain itself (Object::Pad::FieldAttr::LazyInit), as on perl.
		if _, err := s.p.XSHelperOp(opRefcntInc, obj, 0, 0, ""); err != nil {
			f.fail("sv_magicext: " + err.Error())
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

// callSvtHook invokes a node's vtbl entry at vtblOff — int (*)(SV*,
// MAGIC*), the shared shape of svt_set and svt_free — if present. died
// says the hook died, with the message msg.
func (s *state) callSvtHook(sv uint64, mg, vtblOff uintptr) (msg string, died bool) {
	node := copyIn(mg, mgSize)
	vtbl := uintptr(binary.LittleEndian.Uint64(node[mgOffVirtual:]))
	if vtbl == 0 {
		return "", false
	}
	fnBytes := copyIn(vtbl+vtblOff, 8)
	fn := uintptr(binary.LittleEndian.Uint64(fnBytes))
	if fn == 0 {
		return "", false
	}
	s.mu.Lock()
	invoke := s.svtInvoke
	s.mu.Unlock()
	if invoke == 0 {
		return "", false
	}
	f, fp := s.bootFrame()
	defer unpinFrame(fp)
	runtime.LockOSThread()
	prevThread, prevInterrupt, entered := s.enterNative()
	// The SDK-side entry installs the current frame and a croak guard
	// before running the hook; interpreter macros inside it need both. It
	// returns 0 when the hook died, with the message in the frame.
	ran, _, _ := purego.SyscallN(invoke, fp, fn, uintptr(sv), mg)
	s.leaveNative(prevThread, prevInterrupt, entered)
	runtime.UnlockOSThread()
	if uint32(ran) == 0 {
		return f.errString(), true
	}
	return "", false
}

func (s *state) callSvtFree(sv uint64, mg uintptr) (msg string, died bool) {
	return s.callSvtHook(sv, mg, vtblOffFree)
}

// magicLocalized runs when `local` puts the SV nsv in the place of the
// variable whose anchor is id. It does for the mirror chain what perl's
// mg_localize does for an SV's own: value magic stays behind; an entry
// whose vtable has svt_local and asks for it (MGf_LOCAL) decides itself;
// every other entry is attached to nsv again with sv_magicext, which gives
// nsv a chain and an anchor of its own. The interpreter then calls set
// magic on nsv, and frees its magic with it when the scope ends.
func (s *state) magicLocalized(id uint32, nsv uint64) []byte {
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
		isValue, err := s.p.XSHelperOp(opSvCall, uint64(node[mgOffType]), 0, svCallMagicIsValue, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "goperl xs: local on a variable with magic: %v\n", err)
			return []byte{1}
		}
		switch {
		case isValue != 0:
		case node[mgOffFlags]&mgfLocal != 0 && vtbl != 0 && binary.LittleEndian.Uint64(copyIn(vtbl+vtblOffLocal, 8)) != 0:
			if msg, died := s.callSvtHook(nsv, mg, vtblOffLocal); died {
				// `local` dies with it, as perl's mg_localize lets it out
				return append([]byte{0}, msg...)
			}
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
	return []byte{1}
}

// perlioHook services reserved method -6: run one host layer hook (the
// SDK-side __goperl_perlio_invoke does the marshalling and croak guard).
// The request starts [hook][funcs id][f][layer].
func (s *state) perlioHook(req []byte) []byte {
	if len(req) < 4*8 {
		return append([]byte{0}, "perlio hook: short request"...)
	}
	id := uint32(wordAt(req, 1))
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
	runtime.LockOSThread()
	prevThread, prevInterrupt, entered := s.enterNative()
	rlen, _, _ := purego.SyscallN(inv, fp, funcs, reqC, uintptr(len(req)), respC, 4096)
	s.leaveNative(prevThread, prevInterrupt, entered)
	runtime.UnlockOSThread()
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
// does the marshalling, parser sync, and croak guard). Request: [kind] and
// the bytes of the word. Response: [1][handled][ret][op] or [0]+message;
// for an infix operator a module claimed, [1][1][its length][the module's
// struct][its precedence][has parse].
func (s *state) keywordHook(req []byte) []byte {
	if len(req) < 8+1 {
		return append([]byte{0}, "keyword hook: short request"...)
	}
	kind := uint32(wordAt(req, 0))
	word := req[8:]
	s.mu.Lock()
	inv := s.keywordInvoke
	s.mu.Unlock()
	if inv == 0 {
		// no SDK module with a parse surface loaded: decline
		return []byte{1, 0, 0, 0, 0, 0, 0, 0, 0}
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
	words := func(w ...uint64) []byte {
		out := make([]byte, 1+8*len(w))
		out[0] = 1
		for i, v := range w {
			binary.LittleEndian.PutUint64(out[1+8*i:], v)
		}
		return out
	}
	var out []byte
	switch int32(rc) {
	case -1:
		msg := f.errString()
		thrown := f.failed == failedThrown
		cFree(retC)
		cFree(tokC)
		unpinFrame(fp)
		if thrown {
			// the interpreter holds the value the plugin threw
			return []byte{respThrow}
		}
		return append([]byte{0}, msg...)
	case 1:
		ret := binary.LittleEndian.Uint32(copyIn(retC, 4))
		tok := binary.LittleEndian.Uint64(copyIn(tokC, 8))
		if kind == 1 {
			// An infix operator: ret is its length and tok the module's
			// struct Perl_custom_infix {prec int32; parse; build_op},
			// whose precedence and whether it has a parse go to the
			// interpreter with it.
			fields := copyIn(uintptr(tok), infixSize)
			var hasParse uint64
			if binary.LittleEndian.Uint64(fields[infixOffParse:]) != 0 {
				hasParse = 1
			}
			prec := uint64(binary.LittleEndian.Uint32(fields[infixOffPrec:]))
			out = words(1, uint64(ret), tok, prec, hasParse)
			break
		}
		// the plugin's result (a KEYWORD_PLUGIN_ value) and the op
		out = words(1, uint64(int64(int32(ret))), tok)
	default:
		out = words(0)
	}
	cFree(retC)
	cFree(tokC)
	unpinFrame(fp)
	return out
}

// magicSetHead implements SvMAGIC_set over the mirror chain: the module
// relinked the chain itself (typically to unlink the head) and hands us the
// new head. Unlinked nodes are the module's to free.
//
// An SV that had no chain gets one: a module links a MAGIC of its own onto
// the empty chain (Event's watchers, Tk's traces), and the SV has that MAGIC
// from then on, anchored in the interpreter as one sv_magicext attaches is,
// so that the chain goes when the SV is freed. Its nodes are the module's:
// it frees what it allocated, as on perl.
//
// A chain the module empties (Variable::Magic's dispell unlinks the last
// MAGIC it cast) leaves the SV with no magic of native code's: the
// interpreter's anchor goes, as with sv_unmagic of the last MAGIC
// (magicDel), and mg_magical finds no magic of it any more. While the
// anchor is being freed, the free takes it off (see magicRec.freeing).
func (s *state) magicSetHead(f *cFrame, sv uint64, mg uintptr) {
	s.mu.Lock()
	if rec := s.magicBySV[sv]; mg == 0 {
		detach := rec != nil && !rec.freeing
		if rec != nil {
			rec.head = 0
		}
		if detach {
			delete(s.magicBySV, sv)
			delete(s.magicByID, rec.anchorID)
		}
		s.mu.Unlock()
		if detach {
			// As in magicDel: the anchor's free finds nothing to do.
			if _, err := s.p.XSHelperOp(opMagicUnattach, sv, 0, 0, ""); err != nil {
				f.fail("SvMAGIC_set: " + err.Error())
			}
		}
		return
	} else if rec != nil {
		rec.head = mg
		s.mu.Unlock()
	} else {
		s.nextMagicID++
		rec := &magicRec{svTok: sv, head: mg, anchorID: s.nextMagicID}
		s.magicBySV[sv] = rec
		s.magicByID[rec.anchorID] = rec
		s.mu.Unlock()
		if _, err := s.p.XSHelperOp(opMagicAttach, sv, 0, uint64(rec.anchorID), ""); err != nil {
			f.fail("SvMAGIC_set: " + err.Error())
			return
		}
	}
	// The interpreter reports the accesses the vtables of the chain have
	// hooks for, as for the magic sv_magicext attaches: a chain spliced
	// with a MAGIC of another vtable, or one whose MAGIC was given a vtable
	// after it was attached (mg_magical of the SDK sets the head the chain
	// has for this look), gets the hooks of its vtables.
	var hooks uint64
	for node := mg; node != 0; {
		buf := copyIn(node, mgAllocSize)
		if vtbl := uintptr(binary.LittleEndian.Uint64(buf[mgOffVirtual:])); vtbl != 0 {
			fns := copyIn(vtbl, 8*8)
			for bit, off := range []int{vtblOffGet, vtblOffSet, vtblOffLen, vtblOffClear, vtblOffCopy} {
				if binary.LittleEndian.Uint64(fns[off:]) != 0 {
					hooks |= 1 << bit
				}
			}
		}
		node = uintptr(binary.LittleEndian.Uint64(buf[mgOffMore:]))
	}
	if hooks != 0 {
		if _, err := s.p.XSHelperOp(opMagicSetHook, sv, hooks, 0, ""); err != nil {
			f.fail("SvMAGIC_set: " + err.Error())
		}
	}
}

// magicFreed runs when the guest frees an SV carrying the anchor magic:
// execute every mirror node's svt_free and drop the chain. As in perl's
// mg_free, the SV has its chain while a MAGIC's svt_free runs, starting at
// that MAGIC (a hook that looks its MAGIC up with mg_findext finds it, as
// Linux::Epoll's does), and loses it once every MAGIC is freed.
//
// A hook that dies ends it, as a die ends perl's mg_free: the response is
// [0] and the message, which the guest's anchor dies with; the MAGIC is
// not freed (perl's mg_free never gets to it), nor what is after it:
// Variable::Magic's free takes its MAGIC off and frees it, has the SV
// freed once more for the rest of the chain, and dies. A hook that
// returns has its MAGIC freed and the chain go on after it, as in perl's
// mg_free, whatever the hook did to the chain (Variable::Magic's dispell
// of the MAGIC being freed leaves the MAGIC to mg_free; magic cast from
// the hook is dropped). The response is [1] otherwise.
func (s *state) magicFreed(id uint32) []byte {
	s.mu.Lock()
	rec := s.magicByID[id]
	if rec != nil {
		rec.freeing = true
	}
	s.mu.Unlock()
	if rec == nil {
		return []byte{1}
	}
	defer func() {
		s.mu.Lock()
		if s.magicByID[id] == rec {
			delete(s.magicByID, id)
		}
		if s.magicBySV[rec.svTok] == rec {
			delete(s.magicBySV, rec.svTok)
		}
		s.mu.Unlock()
	}()
	for mg := rec.head; mg != 0; {
		node := copyIn(mg, mgAllocSize)
		next := uintptr(binary.LittleEndian.Uint64(node[mgOffMore:]))
		s.mu.Lock()
		rec.head = mg
		s.mu.Unlock()
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
		if lastMagicHolder(node) {
			if msg, died := s.callSvtFree(rec.svTok, mg); died {
				return append([]byte{0}, msg...)
			}
			// What svt_free left in the MAGIC is what is given up.
			node = copyIn(mg, mgAllocSize)
		}
		s.releaseMagicObj(node)
		s.releaseMagicName(int32(binary.LittleEndian.Uint32(node[mgOffLen:])),
			uintptr(binary.LittleEndian.Uint64(node[mgOffPtr:])))
		s.mu.Lock()
		rec.head = next
		s.mu.Unlock()
		cFree(mg)
		mg = next
	}
	return []byte{1}
}

// magicTakesObj reports whether sv_magicext takes a reference on obj for a
// MAGIC of type how on sv, as perl's does (and marks the MAGIC
// MGf_REFCOUNTED): not for no obj, not for the variable itself, nor for
// the types whose obj is the variable's own container.
func magicTakesObj(sv, obj uint64, how int32) bool {
	if obj == 0 || obj == sv {
		return false
	}
	switch how {
	case '#', 'D', 'd', ':': // arylen, regdata, regdatum, symtab
		return false
	}
	return true
}

// globHolds reports whether obj is a glob one of whose slots (scalar,
// array, hash, sub, handle, format) holds sv. sv_magicext takes no
// reference on such an obj: the glob holds sv, and a reference back would
// be a loop that keeps both alive (the isa magic of @ISA, whose obj is its
// glob, as Package::Stash::XS adds a package's @ISA).
func (s *state) globHolds(obj, sv uint64) bool {
	if obj == 0 || sv == 0 {
		return false
	}
	if isGV, err := s.p.XSHelperOp(opSvIsGVGP, obj, 0, 0, ""); err != nil || isGV == 0 {
		return false
	}
	for _, sel := range []uint64{gvPtrSV, gvPtrAV, gvPtrHV, gvPtrCV, gvPtrIO} {
		if got, err := s.p.XSHelperOp(opGvPtr, obj, sel, 0, ""); err == nil && got == sv {
			return true
		}
	}
	form, err := s.p.XSHelperOp(opGvCall, obj, 0, gvCallForm, "")
	return err == nil && form == sv
}

// releaseMagicObj gives up the reference a MAGIC holds on its mg_obj, as
// mg_free does: only a MAGIC that took one (MGf_REFCOUNTED) holds one —
// a module that sets mg_obj itself gives its reference up in its own
// svt_free.
func (s *state) releaseMagicObj(node []byte) {
	obj := binary.LittleEndian.Uint64(node[mgOffObj:])
	if obj == 0 || node[mgOffFlags]&mgfRefcounted == 0 {
		return
	}
	_, _ = s.p.XSHelperOp(opRefcntDec, obj, 0, 0, "")
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
		_, _ = s.p.XSHelperOp(opRefcntDec, uint64(mgPtr), 0, 0, "")
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
			if lastMagicHolder(node) {
				if msg, died := s.callSvtFree(sv, mg); died {
					// sv_unmagic of native code goes on: its caller is
					// native code, which a die cannot reach from here
					fmt.Fprintf(os.Stderr, "goperl xs: svt_free died in sv_unmagic: %s\n", msg)
				}
				node = copyIn(mg, mgAllocSize)
			}
			s.releaseMagicObj(node)
			s.releaseMagicName(int32(binary.LittleEndian.Uint32(node[mgOffLen:])),
				uintptr(binary.LittleEndian.Uint64(node[mgOffPtr:])))
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
	s.mu.Lock()
	detach := rec.head == 0 && !rec.freeing
	s.mu.Unlock()
	if detach {
		s.mu.Lock()
		delete(s.magicBySV, sv)
		delete(s.magicByID, rec.anchorID)
		s.mu.Unlock()
		// Detaching the guest anchor would re-enter magicFreed; the maps
		// are already cleared, so that callback finds nothing to do.
		_, _ = s.p.XSHelperOp(opMagicUnattach, sv, 0, 0, "")
	}
}

// pvHostSize is the size of what is kept for an SV whose buffer slot holds
// a pointer of the module's own: the pointer and the buffer's room.
const pvHostSize = 16

// pvHostSlot returns the storage of what is kept for sv's buffer slot,
// making it when asked to; 0 when sv has none.
func (s *state) pvHostSlot(sv uint64, create bool) uintptr {
	s.mu.Lock()
	defer s.mu.Unlock()
	slot, ok := s.pvHost[sv]
	if !ok && create {
		slot = cMalloc(pvHostSize)
		pokeBytes(slot, make([]byte, pvHostSize))
		if s.pvHost == nil {
			s.pvHost = map[uint64]uintptr{}
		}
		s.pvHost[sv] = slot
	}
	return slot
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
		v, err := f.p.XSHelperOp(1, uint64(sv), 0, 0, "")
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
		// SV_PV: the address of the string; second: its length
		off, err := f.p.XSHelperOp(2, uint64(sv), 0, 0, "")
		if err != nil {
			f.fail("SvPV: " + err.Error())
			return 0
		}
		mem := f.p.RawMemory()
		ln := opboxResult2(mem, f.opbox)
		if lenp != 0 {
			pokeU64(lenp, ln)
		}
		if off > uint64(len(mem)) || ln > uint64(len(mem))-off {
			f.fail("SvPV: string outside linear memory")
			return 0
		}
		return uintptr(unsafe.Pointer(unsafe.SliceData(mem))) + uintptr(off)
	})
	vtable.newIV = purego.NewCallback(func(fr, v uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		sv, err := f.p.XSHelperOp(3, uint64(v), 0, 0, "")
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
		sv, err := f.p.XSHelperOp(4, 0, uint64(ln), 0, goBytesString(p, int(ln)))
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
		out, err := f.p.XSHelperOp(5, uint64(sv), 0, 0, "")
		if err != nil {
			f.fail("sv_2mortal: " + err.Error())
			return 0
		}
		return uintptr(out)
	})
	vtable.xsOp = purego.NewCallback(func(fr uintptr, op int32, a, b, c, sPtr, sLen uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		v, err := f.p.XSHelperOp(op, uint64(a), uint64(b), uint64(c), goBytesString(sPtr, int(sLen)))
		if xsTrace {
			fmt.Fprintf(os.Stderr, "[xsop] op=%d a=%#x b=%#x c=%#x slen=%d -> %#x\n", op, a, b, c, sLen, v)
		}
		if errors.Is(err, internal.ErrGuestJump) {
			// The interpreter died inside the operation and is unwinding
			// to a catch point of its own: the native call ends here (the
			// SDK jumps to its entry), and says nothing of its own.
			if f.failed == 0 {
				f.failed = failedJump
				f.err[0] = 0
			}
			return 0
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
	vtable.pvHost = purego.NewCallback(func(fr, sv uintptr, create int32) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		return stateFor(f.p).pvHostSlot(uint64(sv), create != 0)
	})
	vtable.xsOpBox = purego.NewCallback(func(fr uintptr) uintptr {
		f := lookupFrame(fr)
		if f == nil {
			return 0
		}
		err := f.p.XSOpRun()
		if errors.Is(err, internal.ErrGuestJump) {
			// as in xsOp: the interpreter is unwinding to a catch point
			// of its own, and the native call ends here
			if f.failed == 0 {
				f.failed = failedJump
				f.err[0] = 0
			}
			return 0
		}
		if err != nil {
			op := stateFor(f.p).boxedOp()
			fmt.Fprintf(os.Stderr, "goperl xs: op in the box failed: %s: %v\n", op, err)
			f.fail(fmt.Sprintf("xs op %s: %v", op, err))
			return 0
		}
		return 1
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
		if gptr == 0 || uint64(gptr) >= uint64(len(mem)) {
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
		stateFor(f.p).magicSetHead(f, uint64(sv), mg)
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
		if _, err := f.p.XSHelperOp(opPPHookSet, uint64(uint32(opType)), enable, 0, ""); err != nil {
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

// releaseStack frees what the frame holds on the C heap when it ends: the
// stack when it is a heap block (the loader's for a long argument list, or
// the one the SDK moved it to), and the SDK's blocks of SV** result slots.
func (f *cFrame) releaseStack() {
	if f.stBase != 0 && !f.inlineStack() {
		cFree(f.stBase)
		f.stBase = uintptr(unsafe.Pointer(&f.st[0]))
		f.stCap = maxStack
	}
	if f.markv != 0 {
		cFree(f.markv)
		f.markv, f.markcap = 0, 0
	}
	for f.tmpMore != 0 {
		b := f.tmpMore
		f.tmpMore = uintptr(binary.LittleEndian.Uint64(copyIn(b, 8)))
		cFree(b)
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
	dladdrFn uintptr
	memcpyFn uintptr
	mallocFn uintptr
	freeFn   uintptr
	libcErr  error
)

func libcInit() {
	libcOnce.Do(func() {
		// A perl executable is linked with the C libraries of its
		// configuration (perllibs: -lpthread -ldl -lm -lcrypt -lutil -lc),
		// and XS modules count on their functions being loaded already:
		// many are linked without -lm, and a module that calls crypt or
		// fcrypt (Crypt::Passwd) without -lcrypt. Of these, the math
		// library and libcrypt are libraries of their own on Linux (the
		// others are in the C library); this binary has neither, so they go
		// into the global namespace before the first module. Where one is
		// missing, a module that needs it fails to load with the dynamic
		// linker's own message.
		if runtime.GOOS == "linux" {
			for _, lib := range []string{"libm.so.6", "libcrypt.so.1"} {
				_, _ = purego.Dlopen(lib, purego.RTLD_NOW|purego.RTLD_GLOBAL)
			}
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
		// only for messages (libraryOf): none where the C library lacks it
		if fn, err := purego.Dlsym(purego.RTLD_DEFAULT, "dladdr"); err == nil {
			dladdrFn = fn
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
