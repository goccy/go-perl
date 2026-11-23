//go:build darwin || linux

package xs_test

import (
	"context"
	"fmt"
	"testing"
)

// BenchmarkNativeCall measures the call of an XSUB from Perl code: one
// iteration is a loop of xsubCalls calls, so the time of an iteration
// divided by that number is the cost of one call, the loop's own work
// included. The two functions differ in how they open their activation:
// one says dXSARGS, the other reads the stack with dMARK and dSP alone.
func BenchmarkNativeCall(b *testing.B) {
	const xsubCalls = 10000
	p := newHostPerl(b)
	perlAPIInstance(b, p)
	ctx := context.Background()
	want := fmt.Sprint(xsubCalls)
	for _, bc := range []struct{ name, call string }{
		{"dXSARGS", `PerlAPI::is_plain_value($_)`},
		{"dMARK_and_dSP", `PerlAPI::sum_without_dxsargs(-99)`},
	} {
		code := fmt.Sprintf(`my $n = 0; $n += %s for 1 .. %d; $n`, bc.call, xsubCalls)
		b.Run(bc.name, func(b *testing.B) {
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
