//go:build darwin || linux

// Package xsnative loads host-native XS modules — shared libraries compiled
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
// Requirements and caveats:
//   - The .so must be built against THIS SDK (system-perl XS binaries can
//     never load: their macros compiled into direct struct access against a
//     different memory model).
//   - Native code runs outside the wasm sandbox — loading a module is as
//     trusting as cgo.
//   - The instance must be on the copy-on-write snapshot path (the default),
//     whose linear memory never relocates: native code holds raw pointers
//     into it while a call is in flight.
package xsnative

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	perl "github.com/goccy/go-perl"
)

// abiVersion must match GOPERL_XS_ABI in sdk/include/perl.h.
const abiVersion = 2

const maxStack = 64

// cFrame mirrors goperl_frame_t in sdk/include/perl.h (ABI). The trailing Go
// fields are invisible to C, which only knows the leading 1064 bytes.
type cFrame struct {
	api      uintptr
	subname  uintptr
	jb       uintptr
	items    int32
	nret     int32
	failed   int32
	reserved int32
	st       [maxStack]uint64
	err      [512]byte
	tmp      uint64 // scratch slot for the SDK's pointer-returning fetch macros

	// Go-only:
	p *perl.Perl
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
	// v2:
	xsOp      uintptr
	ptrEncode uintptr
	ptrDecode uintptr
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
}

func (s *state) encodePtr(p uintptr) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ptrs = append(s.ptrs, p)
	return uint64(len(s.ptrs))
}

func (s *state) decodePtr(id uint64) uintptr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == 0 || id > uint64(len(s.ptrs)) {
		return 0
	}
	return s.ptrs[id-1]
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
		s = &state{p: p}
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
	bootName := "boot_" + strings.ReplaceAll(module, "::", "__")
	bootFn, err := purego.Dlsym(lib, bootName)
	if err != nil {
		return fmt.Errorf("%s: no %s symbol: %w", path, bootName, err)
	}
	// The static-perl bootstrap contract: DynaLoader::bootstrap (which
	// XSLoader falls back to on a perl without dynamic loading) resolves
	// and calls &{"${module}::bootstrap"}.
	return s.registerNative(module+"::bootstrap", bootFn)
}

// registerNative records fn and binds it as the Perl sub name via the
// guest's generic thunk. Called from the vtable during boot.
func (s *state) registerNative(name string, fn uintptr) error {
	s.mu.Lock()
	s.fns = append(s.fns, fn)
	s.fnNames = append(s.fnNames, append([]byte(name), 0))
	id := int32(len(s.fns) - 1)
	s.mu.Unlock()
	return s.p.RegisterNativeXS(name, id)
}

// dispatch handles one native XSUB call: payload [u32 fn_id][u32 items]
// [u32 tokens...]; response [1][u32 nret][u32 tokens...] or [0]+message.
func (s *state) dispatch(req []byte) []byte {
	fail := func(msg string) []byte { return append([]byte{0}, msg...) }
	if len(req) < 8 {
		return fail("native XS dispatch: short payload")
	}
	id := int32(binary.LittleEndian.Uint32(req[0:]))
	items := int(binary.LittleEndian.Uint32(req[4:]))
	if items > maxStack {
		return fail(fmt.Sprintf("native XS dispatch: %d arguments exceed the SDK stack (%d)", items, maxStack))
	}
	if len(req) < 8+4*items {
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

	f := &cFrame{api: uintptr(unsafe.Pointer(vtable)), p: s.p, subname: subname}
	f.items = int32(items)
	for i := 0; i < items; i++ {
		f.st[i] = uint64(binary.LittleEndian.Uint32(req[8+4*i:]))
	}
	fp := pinFrame(f)
	purego.SyscallN(fn, fp)
	unpinFrame(fp)
	if f.failed != 0 {
		return fail(f.errString())
	}
	n := int(f.nret)
	if n < 0 || n > maxStack {
		return fail(fmt.Sprintf("native XS dispatch: bad return count %d", n))
	}
	resp := make([]byte, 5+4*n)
	resp[0] = 1
	binary.LittleEndian.PutUint32(resp[1:], uint32(n))
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint32(resp[5+4*i:], uint32(f.st[i]))
	}
	return resp
}

// buildVtable creates the C-callable API table. Callbacks are process-global
// (purego.NewCallback registrations are permanent); the frame carries the
// instance.
func buildVtable() {
	vtable = &cAPI{abi: abiVersion}
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
		if err != nil {
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
		if err := s.registerNative(cString(name), fn); err != nil {
			f.fail("newXS: " + err.Error())
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
// purego (strlen to size a C string, memcpy to move bytes across), and Go
// only ever reads its own buffers.
var (
	libcOnce sync.Once
	strlenFn uintptr
	memcpyFn uintptr
	libcErr  error
)

func libcInit() {
	libcOnce.Do(func() {
		for _, sym := range []struct {
			name string
			dst  *uintptr
		}{{"strlen", &strlenFn}, {"memcpy", &memcpyFn}} {
			fn, err := purego.Dlsym(purego.RTLD_DEFAULT, sym.name)
			if err != nil {
				libcErr = fmt.Errorf("resolve libc %s: %w", sym.name, err)
				return
			}
			*sym.dst = fn
		}
	})
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
