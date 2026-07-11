package perl_test

import (
	"os/exec"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// The classic naive recursive Fibonacci: fib(36) drives ~29.9M calls
// (14930352 leaf additions), a deep-recursion, function-call-bound workload
// that stresses the interpreter's op dispatch and sub-call machinery.
const (
	fibDef   = `sub fib { my $n = shift; $n < 2 ? $n : fib($n-1) + fib($n-2) }`
	fib36Val = "14930352"
)

// BenchmarkFib36 measures go-perl — the wasm2go-transpiled Perl interpreter
// running entirely in Go — computing fib(36). The interpreter and the sub
// definition are set up once (persistent state), so the timed region is the
// fib(36) evaluation itself, not interpreter boot.
func BenchmarkFib36(b *testing.B) {
	i, err := perl.NewInterpreter(perl.Config{})
	if err != nil {
		b.Fatalf("NewInterpreter: %v", err)
	}
	defer i.Close()
	if _, err := i.Eval(fibDef); err != nil {
		b.Fatalf("define fib: %v", err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, err := i.Eval(`fib(36)`)
		if err != nil {
			b.Fatalf("Eval: %v", err)
		}
		if !r.Ok || r.Result != fib36Val {
			b.Fatalf("fib(36) = %q ok=%v error=%q", r.Result, r.Ok, r.Error)
		}
	}
}

// BenchmarkFib36NativePerl measures the host's /usr/bin/perl on the equivalent
// script, as a baseline to compare go-perl against. Each iteration spawns a
// fresh perl process; its interpreter boot (single-digit ms) is negligible
// next to the multi-second fib(36) run. Skipped if /usr/bin/perl is absent.
func BenchmarkFib36NativePerl(b *testing.B) {
	const perlBin = "/usr/bin/perl"
	if _, err := exec.LookPath(perlBin); err != nil {
		b.Skipf("%s not available: %v", perlBin, err)
	}
	script := fibDef + `; print fib(36)`
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err := exec.Command(perlBin, "-e", script).Output()
		if err != nil {
			b.Fatalf("run %s: %v", perlBin, err)
		}
		if got := strings.TrimSpace(string(out)); got != fib36Val {
			b.Fatalf("native fib(36) = %q, want %q", got, fib36Val)
		}
	}
}
