package perl_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// A guest read at an address the memory does not have — past its size
// but inside the instance's mapping, past the mapping, at 2^63, at the
// top of the address space — is a trap that ends the interpreter: never
// the bytes of whatever lies at that host address (zeros inside the
// mapping, the process's own memory past it), and never a fault that
// takes the process down. unpack "p" dereferences any address it is
// given, so it reaches all of them.
func TestReadPastTheMemoryTraps(t *testing.T) {
	t.Skip("the runtime does not confine guest memory accesses to the instance's memory yet")
	for _, addr := range []string{"3*2**30", "9*2**30", "2**63", "~0"} {
		t.Run(addr, func(t *testing.T) {
			cfg := hostConfig(t)
			var stdout, stderr bytes.Buffer
			cfg.Stdin, cfg.Stdout, cfg.Stderr = strings.NewReader(""), &stdout, &stderr
			p, err := perl.New(cfg)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			defer p.Close()
			status, err := p.RunProgram(context.Background(), perl.Program{Code: code(
				`$| = 1;`,
				`print "before\n";`,
				`print length(unpack("p", pack("Q", `+addr+`))), "\n";`,
				`print "after\n";`,
			)})
			if err == nil || status == 0 {
				t.Fatalf("RunProgram = (%d, %v), want a failure status and an error", status, err)
			}
			if !strings.Contains(err.Error(), "memory access out of bounds") {
				t.Fatalf("RunProgram error = %v, want the out-of-bounds trap", err)
			}
			if got, want := stdout.String(), "before\n"; got != want {
				t.Fatalf("the program printed %q, want %q", got, want)
			}
		})
	}
}
