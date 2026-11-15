package perl_test

import (
	"context"
	"testing"
)

// A program's memory grows as far as a wasm32 memory can (just under 4
// GiB), as perl's does on the host as far as the machine's memory goes. An
// instance's memory used to be mapped 512 MiB long, and a program that
// needed more ran out of memory there ("Out of memory in
// perl:util:safesysrealloc"): a string of 520 MiB, a test suite that holds
// a large data set. Both kinds of instance are mapped alike: the first of a
// process, on memory of its own, and the ones built on the process's
// snapshot.
func TestProgramMemoryGrowsPastHalfAGigabyte(t *testing.T) {
	const size = 520 << 20
	for i := 0; i < 2; i++ {
		p := newPerl(t)
		r, err := p.Eval(context.Background(), `my $s = 'a' x (520 << 20); length $s`)
		if err != nil {
			t.Fatalf("instance %d: Eval: %v", i, err)
		}
		if r.Error != nil {
			t.Fatalf("instance %d: eval not ok: error=%v stderr=%q", i, r.Error, r.Stderr)
		}
		if got := resultStr(r); got != "545259520" {
			t.Fatalf("instance %d: length = %q, want %d", i, got, size)
		}
		p.Close()
	}
}
