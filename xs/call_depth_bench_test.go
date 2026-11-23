//go:build darwin || linux

package xs_test

import (
	"context"
	"fmt"
	"testing"
)

// BenchmarkNativeCallDepth is BenchmarkNativeCall with values below the
// call on the interpreter's stack: the loop runs inside a foreach over a
// list of that many values, which stay on the stack while it runs, as
// those of a map or a sort over a long list do. An XSUB's stack is the
// interpreter's by index, so its arguments are that far up; what lies
// below them is read only when native code looks there, and a call costs
// what one at the bottom costs (the foreach's own steps over the list
// included).
func BenchmarkNativeCallDepth(b *testing.B) {
	const xsubCalls = 10000
	p := newHostPerl(b)
	perlAPIInstance(b, p)
	ctx := context.Background()
	want := fmt.Sprint(xsubCalls)
	for _, below := range []int{0, 100000} {
		code := fmt.Sprintf(`my $n = 0; for my $x (((0) x %d), 1) { next unless $x; $n += PerlAPI::is_plain_value($_) for 1 .. %d } $n`, below, xsubCalls)
		b.Run(fmt.Sprintf("%d_below", below), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r, err := p.Eval(ctx, code)
				if err != nil || r.Error != nil {
					b.Fatalf("err=%v error=%v", err, r.Error)
				}
				if got := resultStr(r); got != want {
					b.Fatalf("got %q, want %q", got, want)
				}
			}
		})
	}
}
