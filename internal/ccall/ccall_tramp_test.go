//go:build (darwin || linux) && (arm64 || amd64)

package ccall

import "testing"

// A Call does not allocate: the Call is the caller's.
func TestInvokeAllocatesNothing(t *testing.T) {
	h := fixture(t)
	weigh6 := sym(t, h, "weigh6")
	c := new(Call)
	if n := testing.AllocsPerRun(100, func() { c.Invoke(weigh6, 1, 2, 3, 4, 5, 6) }); n != 0 {
		t.Fatalf("%v allocations a call", n)
	}
}
