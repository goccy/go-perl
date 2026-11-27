//go:build linux

package xs_test

import (
	"testing"

	"github.com/ebitengine/purego"
	"github.com/goccy/go-perl/xs"
)

// The C library functions that change the ids of the process come, like
// the others the runtime calls, from the entry table of a module linked
// into the binary: in a static binary, the C library linked into it. Here
// the module is loaded, so its table holds the process's C library.
func TestIDFunctionsFromTheEntryTable(t *testing.T) {
	registerDemo(t)
	for _, name := range []string{"setuid", "seteuid", "setreuid", "setresuid", "setgid", "setegid", "setregid", "setresgid", "setgroups", "__errno_location"} {
		want, err := purego.Dlsym(purego.RTLD_DEFAULT, name)
		if err != nil {
			t.Fatalf("dlsym %s: %v", name, err)
		}
		if got, err := xs.LibcSymbol(name); err != nil || got != want {
			t.Errorf("LibcSymbol(%s) = %#x, %v, want %#x from the entry table", name, got, err, want)
		}
	}
}
