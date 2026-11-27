//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/xs"
)

// demoEntryFlags compiles the SDK's entry table for Demo::XS into the module.
var demoEntryFlags = []string{"-DGOPERL_ENTRY_BOOT=boot_Demo__XS", "-DGOPERL_ENTRY_TABLE=goperl_entry_Demo__XS", "sdk/include/goperl_entry.c"}

// evalDemo runs Demo::XS through a .pm that boots it the stock way.
func evalDemo(t *testing.T, p *perl.Perl) {
	t.Helper()
	libdir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(libdir, "Demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, "Demo", "XS.pm"),
		[]byte("package Demo::XS;\nour $VERSION = '0.01';\nuse XSLoader;\nXSLoader::load(__PACKAGE__, $VERSION);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := p.AddInc(context.Background(), libdir); err != nil {
		t.Fatalf("AddInc: %v", err)
	}
	r, err := p.Eval(context.Background(), `use Demo::XS; Demo::XS::greet("gopher") . Demo::XS::add(40, 2)`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if want := "hello, gopher42"; r.Error != nil || resultStr(r) != want {
		t.Fatalf("Demo::XS = %q (error=%v), want %q", resultStr(r), r.Error, want)
	}
}

// A module registered as linked into the binary reaches every instance
// created afterwards, with no per-instance loading.
func TestRegisteredModule(t *testing.T) {
	registerDemo(t)
	evalDemo(t, newHostPerl(t))
	evalDemo(t, newHostPerl(t))
}

// registerDemo builds Demo::XS with its entry table and registers the
// loaded module as linked into the binary, for the rest of the test.
func registerDemo(t *testing.T) {
	t.Helper()
	so := buildDemo(t, demoEntryFlags...)
	lib, err := purego.Dlopen(so, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		t.Fatalf("dlopen: %v", err)
	}
	entry, err := purego.Dlsym(lib, xs.EntrySymbol("Demo::XS"))
	if err != nil {
		t.Fatalf("entry table: %v", err)
	}
	// The address is C memory handed back as a uintptr.
	xs.Register("Demo::XS", *(*unsafe.Pointer)(unsafe.Pointer(&entry)))
	t.Cleanup(func() { xs.Unregister("Demo::XS") })
}
