//go:build darwin || linux

package ccall

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

const fixtureC = `
typedef long (*cb8)(long, long, long, long, long, long, long, long);
typedef long (*cb2)(long, int);

/* Calls cb, then writes v through p: a pointer to Go memory, which the Go
   code cb runs moves unless the call keeps it in place. */
long storeAfter(cb2 cb, long *p, long v) {
	cb(0, 0);
	*p = v;
	return v;
}

long weigh6(long a, long b, long c, long d, long e, long f) {
	return a + 2*b + 3*c + 4*d + 5*e + 6*f;
}

/* Calls cb with eight arguments, the last ones on the stack where the
   calling convention has fewer argument registers. */
long call8(cb8 cb, long base) {
	return cb(base, base+1, base+2, base+3, base+4, base+5, base+6, base+7) + 1000;
}

/* A negative C int argument: the upper half of its register is not its
   sign (zero on arm64, where writing the 32-bit register clears it). */
long callint(cb2 cb) {
	return cb(7, -5);
}

/* Recursion through Go: C calls the callback, which calls back into C. */
long depth(cb2 cb, int n) {
	if (n == 0) return 0;
	return cb(n, n) + 1;
}
`

// errOtherArchitecture is the error of compileCommand for a test binary
// whose architecture the C compiler does not build for.
var errOtherArchitecture = errors.New("the C compiler does not build for the test's architecture")

// fixture compiles fixtureC into a shared library and loads it.
func fixture(t *testing.T) uintptr {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		// The calls under test go to C code and come from it, which only
		// a C compiler can make; a host without one has no native modules
		// to call either.
		t.Skipf("no C compiler to build the fixture: %v", err)
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "fixture.c")
	if err := os.WriteFile(src, []byte(fixtureC), 0o644); err != nil {
		t.Fatal(err)
	}
	lib := filepath.Join(dir, "fixture.so")
	name, args, err := compileCommand(cc, []string{"-shared", "-fPIC", "-O1", "-o", lib, src})
	if errors.Is(err, errOtherArchitecture) {
		// The test binary runs translated on a machine of another
		// architecture, whose C compiler cannot build the fixture for it.
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s", err, out)
	}
	h, err := purego.Dlopen(lib, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func sym(t *testing.T, h uintptr, name string) uintptr {
	t.Helper()
	p, err := purego.Dlsym(h, name)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// testCallbacks are the callbacks of TestCallsBothWays, made once:
// callbacks are made for good, and a test that runs again (-count) uses the
// ones it made before.
type testCallbacks struct{ cb8, cbInt, cbDepth, cbMove uintptr }

var (
	// seen is what the callback cb8 was called with last.
	seen Args
	// depthFn is the fixture's depth, which cbDepth calls back: that of
	// the run that calls it.
	depthFn atomic.Uintptr
	// moveFirst has cbDepth move the goroutine's stack in its first call
	// of a run.
	moveFirst atomic.Bool
	// makeTestCallbacks makes the testCallbacks the first time.
	makeTestCallbacks = sync.OnceValues(func() (*testCallbacks, error) {
		cbs := &testCallbacks{}
		var err error
		cbs.cb8, err = NewCallback(func(a *Args) uintptr {
			seen = *a
			return a[0] + a[7]
		})
		if err != nil {
			return nil, err
		}
		cbs.cbInt, err = NewCallback(func(a *Args) uintptr {
			return uintptr(int64(a[0]) * int64(int32(a[1])))
		})
		if err != nil {
			return nil, err
		}
		cbs.cbDepth, err = NewCallback(func(a *Args) uintptr {
			n := int32(a[1])
			if moveFirst.Swap(false) {
				moveStack()
			}
			runtime.GC()
			var c Call
			return c.Invoke(depthFn.Load(), cbs.cbDepth, uintptr(n-1))
		})
		if err != nil {
			return nil, err
		}
		cbs.cbMove, err = NewCallback(func(*Args) uintptr {
			moveStack()
			runtime.GC()
			return 0
		})
		if err != nil {
			return nil, err
		}
		return cbs, nil
	})
)

// grow uses n frames of the goroutine's stack.
func grow(n int) int {
	var pad [64]byte
	if n == 0 {
		return int(pad[0])
	}
	return grow(n-1) + int(pad[n%64])
}

// moveStack makes the runtime move the goroutine's stack: it uses more and
// more of it until the stack is copied to a larger one, which changes the
// address of a local.
func moveStack() {
	var mark byte
	at := uintptr(unsafe.Pointer(&mark))
	for n := 64; uintptr(unsafe.Pointer(&mark)) == at; n *= 2 {
		grow(n)
	}
}

// Calls into C get their arguments in order and return the result; a
// callback gets every one of eight arguments (some on the stack on amd64),
// its result goes back to C, and a C int argument reads right once
// converted; Go memory passed to Invoke stays where C writes into it while
// a callback moves the goroutine's stack; C can call Go which calls C
// again, many levels deep, with the stack moved and collections in
// between: the Calls stay where C wrote into them.
func TestCallsBothWays(t *testing.T) {
	h := fixture(t)
	weigh6, call8, callint := sym(t, h, "weigh6"), sym(t, h, "call8"), sym(t, h, "callint")
	depth, storeAfter := sym(t, h, "depth"), sym(t, h, "storeAfter")
	cbs, err := makeTestCallbacks()
	if err != nil {
		t.Fatal(err)
	}

	var c Call
	if got := c.Invoke(weigh6, 1, 10, 100, 1000, 10000, 100000); got != 1+20+300+4000+50000+600000 {
		t.Fatalf("weigh6: %d", got)
	}
	if got := Invoke(weigh6, 1, 1, 1, 1, 1, 1); got != 21 {
		t.Fatalf("weigh6 through Invoke: %d", got)
	}
	// stored is a local that nothing but the call makes escape: the
	// report gets a copy of it.
	var stored int64
	r := Invoke(storeAfter, cbs.cbMove, uintptr(unsafe.Pointer(&stored)), 42)
	if got := stored; r != 42 || got != 42 {
		t.Fatalf("storeAfter through Invoke: returned %d, stored %d", r, got)
	}

	if got := Invoke(call8, cbs.cb8, 40); got != 40+47+1000 {
		t.Fatalf("call8: %d", got)
	}
	for i, v := range seen {
		if v != 40+uintptr(i) {
			t.Fatalf("callback argument %d: %d, want %d (all: %v)", i, v, 40+i, seen)
		}
	}

	if got := int(Invoke(callint, cbs.cbInt)); got != -35 {
		t.Fatalf("callint: %d", got)
	}

	depthFn.Store(depth)
	moveFirst.Store(true)
	if got := Invoke(depth, cbs.cbDepth, 50); got != 50 {
		t.Fatalf("depth: %d", got)
	}
}

// newCallbackOrFail is NewCallback for a test, which fails when the
// callback cannot be made.
func newCallbackOrFail(t *testing.T, fn func(args *Args) uintptr) uintptr {
	t.Helper()
	entry, err := NewCallback(fn)
	if err != nil {
		t.Fatal(err)
	}
	return entry
}

// limitChildEnv marks the process TestNewCallbackLimit runs its test in:
// callbacks are made for good, so the test uses them all up in a process
// of its own, leaving the other tests theirs.
const limitChildEnv = "CCALL_TEST_LIMIT_CHILD"

// NewCallback makes MaxCallbacks callbacks in all and fails after them,
// leaving the ones it made working.
func TestNewCallbackLimit(t *testing.T) {
	if os.Getenv(limitChildEnv) == "" {
		// The child needs the fixture: a parent that cannot build it
		// skips, as the child would.
		fixture(t)
		cmd := exec.Command(os.Args[0], "-test.run=^TestNewCallbackLimit$", "-test.count=1")
		cmd.Env = append(os.Environ(), limitChildEnv+"=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
		return
	}
	h := fixture(t)
	callint := sym(t, h, "callint")
	var last uintptr
	for range MaxCallbacks {
		last = newCallbackOrFail(t, func(a *Args) uintptr { return a[0] + 1 })
	}
	if _, err := NewCallback(func(*Args) uintptr { return 0 }); err == nil {
		t.Fatalf("callback %d made", MaxCallbacks+1)
	}
	if got := Invoke(callint, last); got != 8 {
		t.Fatalf("last callback: %d", got)
	}
}
