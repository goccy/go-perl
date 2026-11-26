package gperl

import (
	"debug/elf"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
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

// A wrapper's reference to the C library function it stands in for
// (GOPERL_LIBC in perl.h) becomes that function, bound global, while the
// wrapper gets a name of its own: a static binary calls the wrapper from the
// module and the C library's function from it.
func TestPrivatizeELFBindsTheCLibrary(t *testing.T) {
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("no cc available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "m.c")
	if err := os.WriteFile(src, []byte("#include <string.h>\n#include <unistd.h>\n"+
		"extern __typeof__(write) goperl_libc_real_write __attribute__((weak));\n"+
		"static int wrapped;\n"+
		"__attribute__((weak, visibility(\"hidden\"))) ssize_t write(int fd, const void *b, size_t n) {\n"+
		"    wrapped++;\n    return goperl_libc_real_write(fd, b, n);\n}\n"+
		"int entry(const char *s) { write(1, s, strlen(s)); return wrapped; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	obj := filepath.Join(dir, "m.o")
	if out, err := exec.Command("cc", "-c", "-fPIC", "-o", obj, src).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s", err, out)
	}
	if _, err := privatizeELF(obj, "entry", "..gperl.M", nil); err != nil {
		t.Fatal(err)
	}
	f, err := elf.Open(obj)
	if err != nil {
		t.Fatal(err)
	}
	syms, err := f.Symbols()
	if err := errors.Join(err, f.Close()); err != nil {
		t.Fatal(err)
	}
	bind := map[string]elf.SymBind{}
	defined := map[string]bool{}
	for _, s := range syms {
		bind[s.Name], defined[s.Name] = elf.ST_BIND(s.Info), s.Section != elf.SHN_UNDEF
	}
	if b, ok := bind["write"]; !ok || defined["write"] || b != elf.STB_GLOBAL {
		t.Errorf("write: present %v, defined %v, bind %v; want an undefined global", ok, defined["write"], b)
	}
	if !defined["write..gperl.M"] {
		t.Error("the wrapper is not write..gperl.M")
	}
	if _, ok := bind["goperl_libc_real_write"]; ok {
		t.Error("goperl_libc_real_write keeps its name")
	}

	libc, err := exec.Command("cc", "-print-file-name=libc.a").Output()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(strings.TrimSpace(string(libc))) {
		t.Skip("the C compiler has no static C library")
	}
	main := filepath.Join(dir, "main.c")
	if err := os.WriteFile(main, []byte("#include <stdio.h>\nint entry(const char *);\n"+
		"int main(void) { int n = entry(\"from the module\\n\"); printf(\"wrapped %d\\n\", n); return 0; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "static")
	if out, err := exec.Command("cc", "-static", "-o", bin, main, obj).CombinedOutput(); err != nil {
		t.Fatalf("cc -static: %v\n%s", err, out)
	}
	out, err := exec.Command(bin).CombinedOutput()
	if err != nil || string(out) != "from the module\nwrapped 1\n" {
		t.Fatalf("static binary: %v\n%s", err, out)
	}
}
