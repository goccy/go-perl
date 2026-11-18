//go:build darwin || linux

package ccall

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ebitengine/purego"
)

const fixtureC = `
typedef long (*cb8)(long, long, long, long, long, long, long, long);
typedef long (*cb2)(long, int);

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

// fixture compiles fixtureC into a shared library and loads it.
func fixture(t *testing.T) uintptr {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no C compiler")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "fixture.c")
	if err := os.WriteFile(src, []byte(fixtureC), 0o644); err != nil {
		t.Fatal(err)
	}
	lib := filepath.Join(dir, "fixture.so")
	args := []string{"-shared", "-fPIC", "-O1", "-o", lib, src}
	if runtime.GOOS == "darwin" {
		// the compiler's own architecture is not the test's under Rosetta
		arch := map[string]string{"amd64": "x86_64", "arm64": "arm64"}[runtime.GOARCH]
		args = append([]string{"-arch", arch}, args...)
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
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

// Calls into C get their arguments in order and return the result; a
// callback gets every one of eight arguments (some on the stack on amd64),
// its result goes back to C, and a C int argument reads right once
// converted; C can call Go which calls C again, deeper than the
// goroutine's stack holds at first, with collections in between: the
// Calls stay where C wrote into them.
func TestCallsBothWays(t *testing.T) {
	h := fixture(t)
	weigh6, call8, callint, depth := sym(t, h, "weigh6"), sym(t, h, "call8"), sym(t, h, "callint"), sym(t, h, "depth")

	var c Call
	if got := c.Invoke(weigh6, 1, 10, 100, 1000, 10000, 100000); got != 1+20+300+4000+50000+600000 {
		t.Fatalf("weigh6: %d", got)
	}
	if got := Invoke(weigh6, 1, 1, 1, 1, 1, 1); got != 21 {
		t.Fatalf("weigh6 through Invoke: %d", got)
	}

	var seen Args
	cb8 := NewCallback(func(a *Args) uintptr {
		seen = *a
		return a[0] + a[7]
	})
	if got := Invoke(call8, cb8, 40); got != 40+47+1000 {
		t.Fatalf("call8: %d", got)
	}
	for i, v := range seen {
		if v != 40+uintptr(i) {
			t.Fatalf("callback argument %d: %d, want %d (all: %v)", i, v, 40+i, seen)
		}
	}

	cbInt := NewCallback(func(a *Args) uintptr {
		return uintptr(int64(a[0]) * int64(int32(a[1])))
	})
	if got := int64(Invoke(callint, cbInt)); got != -35 {
		t.Fatalf("callint: %d", got)
	}

	var cbDepth uintptr
	var grow func(n int) int
	grow = func(n int) int {
		var pad [64]byte
		if n == 0 {
			return int(pad[0])
		}
		return grow(n-1) + int(pad[n%64])
	}
	type held struct{ call Call }
	cbDepth = NewCallback(func(a *Args) uintptr {
		n := int32(a[1])
		grow(2000) // moves the goroutine's stack while C waits
		runtime.GC()
		h := &held{}
		return h.call.Invoke(depth, cbDepth, uintptr(n-1))
	})
	if got := Invoke(depth, cbDepth, 50); got != 50 {
		t.Fatalf("depth: %d", got)
	}
}

// A Call does not allocate: the Call is the caller's.
func TestInvokeAllocatesNothing(t *testing.T) {
	h := fixture(t)
	weigh6 := sym(t, h, "weigh6")
	c := new(Call)
	if n := testing.AllocsPerRun(100, func() { c.Invoke(weigh6, 1, 2, 3, 4, 5, 6) }); n != 0 {
		t.Fatalf("%v allocations a call", n)
	}
}
