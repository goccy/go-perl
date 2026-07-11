package perl

// Interpreter is a hand-written, multi-interpreter API layered on top of the
// generated single-global binding in perl.go. Each Interpreter owns its own
// wasm module (independent linear memory + WASI host) and one Perl runtime, so
// several interpreters run concurrently and in isolation. It also exposes the
// WASI sandbox controls (preopen dir, env, filesystem/network policy hooks) and
// the interrupt primitive.

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// Method IDs of the perl bridge service (service 0), in the order the proto
// declares them: perl_close, perl_eval, perl_interrupt_addr, perl_new.
const (
	midClose         = 0
	midEval          = 1
	midInterruptAddr = 2
	midNew           = 3
)

// EvalResult is the decoded form of the JSON document perl_eval returns.
type EvalResult struct {
	// Ok is false when the eval died (a Perl-level die/croak, a compile
	// error, or a host interrupt); Error then holds $@.
	Ok bool `json:"ok"`
	// Result is the stringified value of the evaluated expression (the value
	// of the last statement), valid only when Ok is true.
	Result string `json:"result"`
	// Stdout / Stderr capture what the eval printed to STDOUT / STDERR (the
	// bridge redirects both onto in-memory scalars for the duration).
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	// Error is $@ when Ok is false.
	Error string `json:"error"`
}

// Config configures a new Interpreter's sandbox. The zero value is a usable
// interpreter: no host env is leaked and the embedded standard library is
// extracted and used automatically.
type Config struct {
	// StdlibDir is the directory holding the Perl standard library (the
	// assembled lib/ tree). It becomes @INC via perl_new. When empty, the
	// embedded stdlib is extracted to a temp dir and used.
	StdlibDir string
	// PreopenDir scopes the guest filesystem root ("/") to this host
	// directory. Empty leaves the host "/" visible (no scoping).
	PreopenDir string
	// Env is the environment the guest sees (%ENV). nil means an empty
	// environment — the host process os.Environ() is NOT leaked.
	Env []string
	// FSAccess, when non-nil, is a per-open/create/unlink whitelist. It
	// receives the guest path (relative to the preopen) and whether the
	// access is a write; returning false denies it (the guest open fails).
	FSAccess func(path string, write bool) bool
	// NetAccess, when non-nil, gates the socket accept/recv/send surface.
	// op is "accept"/"recv"/"send"; returning false denies it.
	NetAccess func(op string) bool
	// Dial, when non-nil, is the OUTBOUND-connection whitelist. It is called
	// with ("tcp", dotted-quad IP, port) before each connect; returning false
	// denies the connection. When nil, all outbound connections are allowed.
	Dial func(network, ip string, port int) bool
	// Resolve, when non-nil, is the name-resolution whitelist. It is called
	// with the host being resolved before each lookup; returning false denies
	// it. When nil, all lookups are allowed.
	Resolve func(host string) bool
	// Stdin, when non-nil, backs the guest's fd 0 (STDIN). Defaults to an
	// empty stream (the host process stdin is NOT used).
	Stdin io.Reader
	// Stdout, when non-nil, receives the guest's fd 1 writes (os-level
	// stdout). Note: Perl-level print output is also captured into
	// EvalResult.Stdout by the bridge; Stdout here is the raw fd sink.
	Stdout io.Writer
	// Stderr, when non-nil, receives the guest's fd 2 writes.
	Stderr io.Writer
	// MaxMemoryBytes, when > 0, caps this interpreter's wasm linear memory.
	// A guest allocation that would grow memory past this limit fails
	// (memory.grow returns -1) instead of growing the host process unbounded.
	// Rounded down to a multiple of the 64 KiB wasm page size; values below
	// the module's initial memory are ignored.
	MaxMemoryBytes int
	// MemoryReserveBytes, when > 0, is the initial linear-memory slice
	// capacity reserved for this interpreter. Reserving capacity up front
	// makes boot-time grows zero-copy reslices, dropping a freshly-booted
	// interpreter's resident memory. The reservation is virtual address space,
	// not resident memory. When 0, a default headroom is used.
	MemoryReserveBytes int
	// Exec, when non-nil, is the subprocess whitelist. It is called before
	// every process spawn with the executable path and full argv; returning
	// false denies the spawn. When nil, all spawns are permitted (subject to
	// host-subprocess being built in).
	Exec func(path string, argv []string) bool
	// FS, when non-nil, is the filesystem backend this interpreter sees as its
	// entire guest filesystem ("/"). Every file operation is routed to it, so
	// giving two interpreters separate FS values isolates them completely. Use
	// NewStdlibMemFS() for an in-memory FS pre-loaded with the standard
	// library. When FS is set, PreopenDir is ignored and StdlibDir defaults to
	// "/" (the FS root). When nil, the default os-backed filesystem (optionally
	// scoped by PreopenDir) is used.
	FS FS
}

// Interpreter is one isolated Perl interpreter.
type Interpreter struct {
	m    *Module
	wasi *base.WasiStubs
	h    uint64
}

// NewInterpreter builds a fresh wasm instance, applies the sandbox config,
// initializes the Perl runtime, and returns a ready interpreter.
func NewInterpreter(cfg Config) (inst *Interpreter, err error) {
	// Resolve the standard library location:
	//   - custom FS backend: the stdlib must live inside it; default the search
	//     path to its root ("/"). (Use NewStdlibMemFS to pre-load it.)
	//   - no FS and no StdlibDir: extract the embedded stdlib to a temp dir.
	if cfg.FS != nil {
		if cfg.StdlibDir == "" {
			cfg.StdlibDir = "/"
		}
	} else if cfg.StdlibDir == "" {
		dir, sErr := ExtractStdlib()
		if sErr != nil {
			return nil, fmt.Errorf("extract embedded stdlib: %w", sErr)
		}
		cfg.StdlibDir = dir
	}

	m := &Module{}
	wasi := base.DefaultWASI()
	// Sandbox by default: do not leak the host environment.
	wasi.SetEnv(cfg.Env)
	if cfg.FS != nil {
		wasi.SetFS(cfg.FS)
	} else if cfg.PreopenDir != "" {
		wasi.SetPreopenDir(cfg.PreopenDir)
	}
	if cfg.FSAccess != nil {
		wasi.SetFSAccessHook(cfg.FSAccess)
	}
	if cfg.NetAccess != nil {
		wasi.SetNetAccessHook(cfg.NetAccess)
	}
	if cfg.Dial != nil {
		wasi.SetDialHook(cfg.Dial)
	}
	if cfg.Resolve != nil {
		wasi.SetResolveHook(cfg.Resolve)
	}
	if cfg.Exec != nil {
		wasi.SetExecHook(cfg.Exec)
	}
	// Sandbox stdio by default: an unset stream does NOT fall through to the
	// host process stdio. Stdin defaults to empty (immediate EOF), stdout and
	// stderr to discard. (Perl-level print output is still captured into
	// EvalResult by the bridge; these back the raw guest fds 0/1/2.)
	if cfg.Stdin != nil {
		wasi.SetStdin(cfg.Stdin)
	} else {
		wasi.SetStdin(bytes.NewReader(nil))
	}
	if cfg.Stdout != nil {
		wasi.SetStdout(cfg.Stdout)
	} else {
		wasi.SetStdout(io.Discard)
	}
	if cfg.Stderr != nil {
		wasi.SetStderr(cfg.Stderr)
	} else {
		wasi.SetStderr(io.Discard)
	}
	if cfg.MemoryReserveBytes > 0 {
		m.g = wasm2go.NewWithWASIReserve(wasi, envStubs{m: m}, cfg.MemoryReserveBytes)
	} else {
		m.g = wasm2go.NewWithWASI(wasi, envStubs{m: m})
	}
	// Cap linear-memory growth so a runaway allocation fails in the guest
	// instead of growing the host process unbounded. Round down to a wasm
	// page; ignore values below the module's initial memory.
	if cfg.MaxMemoryBytes > 0 {
		const wasmPage = 65536
		max := uint64(cfg.MaxMemoryBytes) / wasmPage * wasmPage
		if max >= uint64(len(wasm2go.Memory(m.g))) {
			m.g.MaxMem = max
		}
	}

	// Run the reactor _initialize + wasmify init under a recover so a C++
	// static-initializer trap surfaces as an error.
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("instance init panicked: %v", r)
			}
		}()
		wasm2go.Initialize(m.g)
		_ = wasm2go.WasmInit(m.g)
	}()
	if err != nil {
		return nil, err
	}

	inst = &Interpreter{m: m, wasi: wasi}
	h, err := inst.perlNew(cfg.StdlibDir)
	if err != nil {
		return nil, fmt.Errorf("perl_new: %w", err)
	}
	if h == 0 {
		return nil, fmt.Errorf("perl_new returned 0 (interpreter init failed; check StdlibDir=%q)", cfg.StdlibDir)
	}
	inst.h = h
	return inst, nil
}

// Eval compiles and runs src in this interpreter's persistent package (its
// main::) and returns the structured result. A Perl-level die is reported via
// EvalResult.Ok=false / .Error, not as a Go error; a Go error indicates a
// host/transport failure (a wasm trap, encoding problem, ...).
func (i *Interpreter) Eval(src string) (EvalResult, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, i.h)
	buf = pbAppendString(buf, 2, src)
	resp, err := i.m.invoke(0, midEval, buf, wasm2go.Inv_0_1)
	if err != nil {
		return EvalResult{}, err
	}
	if e := pbExtractError(resp); e != nil {
		return EvalResult{}, e
	}
	js := readScalarAtField(resp, 1, (*pbReader).readString)
	var r EvalResult
	if err := json.Unmarshal([]byte(js), &r); err != nil {
		return EvalResult{}, fmt.Errorf("decode eval result %q: %w", js, err)
	}
	return r, nil
}

// Close finalizes the interpreter. The Interpreter must not be used afterward.
func (i *Interpreter) Close() error {
	var buf []byte
	buf = pbAppendUint64(buf, 1, i.h)
	resp, err := i.m.invoke(0, midClose, buf, wasm2go.Inv_0_0)
	if err != nil {
		return err
	}
	return pbExtractError(resp)
}

func (i *Interpreter) perlNew(stdlibDir string) (uint64, error) {
	var buf []byte
	buf = pbAppendString(buf, 1, stdlibDir)
	resp, err := i.m.invoke(0, midNew, buf, wasm2go.Inv_0_3)
	if err != nil {
		return 0, err
	}
	if e := pbExtractError(resp); e != nil {
		return 0, e
	}
	return readScalarAtField(resp, 1, (*pbReader).readUint64), nil
}

// interruptAddr fetches the address of this interpreter's interrupt-flag word
// in linear memory.
func (i *Interpreter) interruptAddr() (uint32, error) {
	var buf []byte
	buf = pbAppendUint64(buf, 1, i.h)
	resp, err := i.m.invoke(0, midInterruptAddr, buf, wasm2go.Inv_0_2)
	if err != nil {
		return 0, err
	}
	if e := pbExtractError(resp); e != nil {
		return 0, e
	}
	return readScalarAtField(resp, 1, (*pbReader).readUint32), nil
}

// Interrupt stops a running evaluation WITHOUT executing any wasm/C code on
// this instance. It writes the interrupt flag directly in linear memory; the
// interpreter's custom run loop (a pluggable PL_runops) checks the flag on
// every opcode and croaks at the next one, so Eval returns cleanly with
// Ok=false. Safe to call from another goroutine while Eval is running a
// long/infinite loop.
//
// The flag address is fetched here via a normal call, so callers that want to
// interrupt a loop must either call Interrupt from a separate goroutine (this
// call cannot acquire the per-instance lock while Eval holds it) or pre-fetch
// it with PrepareInterrupt before starting the loop.
func (i *Interpreter) Interrupt() error {
	ip, err := i.PrepareInterrupt()
	if err != nil {
		return err
	}
	ip.Fire()
	return nil
}

// Interrupter holds the pre-resolved interrupt-flag address. Resolve it with
// PrepareInterrupt BEFORE starting the loop you intend to interrupt (resolving
// needs the instance lock, which a running Eval holds), then call Fire from a
// watchdog goroutine.
type Interrupter struct {
	m    *Module
	addr uint32
}

// PrepareInterrupt resolves the interrupt-flag address up front so Fire can run
// concurrently with a busy Eval (which holds the instance lock).
func (i *Interpreter) PrepareInterrupt() (*Interrupter, error) {
	addr, err := i.interruptAddr()
	if err != nil {
		return nil, err
	}
	var addrErr error
	base.AccessMemory(i.m.g, func(mem []byte) {
		// The flag lives in the interpreter's heap, far below the initial
		// memory size; validate anyway so a surprising address surfaces here
		// rather than as a wild write in Fire.
		if uint64(addr)+4 > uint64(len(mem)) {
			addrErr = fmt.Errorf("interrupt address %#x out of range (memory is %d bytes)", addr, len(mem))
		}
	})
	if addrErr != nil {
		return nil, addrErr
	}
	return &Interrupter{m: i.m, addr: addr}, nil
}

// Fire trips the interrupt: it stores 1 into the interrupt-flag word. It does
// not take the instance lock — that is the point: it runs concurrently with a
// busy Eval.
//
// The write happens inside base.AccessMemory, which holds the same lock the
// runtime's memory.grow takes to mutate the memory slice header or relocate its
// backing array. That makes delivery deterministic: for the duration of the
// write the memory can neither be resliced nor relocated, so the flag lands in
// the array the guest observes. The run loop reads the word with a plain
// single-word load on its own goroutine, which is exactly the pluggable
// PL_runops contract.
func (ip *Interrupter) Fire() {
	base.AccessMemory(ip.m.g, func(mem []byte) {
		binary.LittleEndian.PutUint32(mem[ip.addr:], 1)
	})
}
