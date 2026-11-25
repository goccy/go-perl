package gperl

import (
	"debug/elf"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

// privatizeELF renames what the object defines but the kept symbol, lists
// the renamed definitions of default visibility (what the module's shared
// object exports), and gives the undefined symbols it is told the names
// they are given.
func TestPrivatizeELFExportsAndImports(t *testing.T) {
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("no cc available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "m.c")
	if err := os.WriteFile(src, []byte("int ext(void);\n__attribute__((visibility(\"hidden\"))) int bar(void) { return 1; }\n"+
		"int foo(void) { return bar() + ext(); }\nint entry = 0;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	obj := filepath.Join(dir, "m.o")
	if out, err := exec.Command("cc", "-c", "-fPIC", "-o", obj, src).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s", err, out)
	}
	exports, err := privatizeELF(obj, "entry", "..gperl.M", map[string]string{"ext": "ext..gperl.P"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(exports, []string{"foo"}) {
		t.Errorf("exports = %v, want [foo]", exports)
	}
	f, err := elf.Open(obj)
	if err != nil {
		t.Fatal(err)
	}
	syms, err := f.Symbols()
	if err := errors.Join(err, f.Close()); err != nil {
		t.Fatal(err)
	}
	defined := map[string]bool{}
	for _, s := range syms {
		defined[s.Name] = s.Section != elf.SHN_UNDEF
	}
	for name, want := range map[string]bool{"ext..gperl.P": false, "foo..gperl.M": true, "bar..gperl.M": true, "entry": true} {
		if got, ok := defined[name]; !ok || got != want {
			t.Errorf("symbol %s: present %v, defined %v; want defined %v", name, ok, got, want)
		}
	}
	for _, gone := range []string{"ext", "foo", "bar"} {
		if _, ok := defined[gone]; ok {
			t.Errorf("symbol %s keeps its name", gone)
		}
	}
}
