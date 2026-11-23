//go:build darwin || linux

package xs_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// buildGlobalSymbolsModule compiles testdata/globalsyms/<name>.c into
// auto/<name>/<name>.so under dir. Undefined symbols are left for the
// dynamic linker to find when the module is loaded.
func buildGlobalSymbolsModule(t *testing.T, cc, dir, name string) {
	t.Helper()
	auto := filepath.Join(dir, "auto", name)
	if err := os.MkdirAll(auto, 0o755); err != nil {
		t.Fatal(err)
	}
	args := []string{"-shared", "-fPIC", "-I", "sdk/include", `-DXS_VERSION="1.00"`,
		"-o", filepath.Join(auto, name+".so"), filepath.Join("testdata", "globalsyms", name+".c")}
	if runtime.GOOS == "darwin" {
		args = append(args, "-undefined", "dynamic_lookup")
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile %s.so: %v\n%s", name, err, out)
	}
}

// A module whose dl_load_flags ask for global symbols (0x01) lends the
// functions it exports to the modules loaded after it, which link against
// them, as DynaLoader does on perl; and the module loaded after it keeps
// its own SDK state: its subs and the provider's both work.
func TestModuleLoadedWithGlobalSymbols(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	dir := t.TempDir()
	buildGlobalSymbolsModule(t, cc, dir, "Provider")
	buildGlobalSymbolsModule(t, cc, dir, "Consumer")
	files := map[string]string{
		"Provider.pm": "package Provider;\nrequire DynaLoader;\nour @ISA = ('DynaLoader');\n" +
			"sub dl_load_flags { 0x01 }\n__PACKAGE__->bootstrap('1.00');\n1;\n",
		"Consumer.pm": "package Consumer;\nrequire DynaLoader;\nour @ISA = ('DynaLoader');\n" +
			"__PACKAGE__->bootstrap('1.00');\n1;\n",
	}
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got := evalWithInc(t, true, true, []string{dir},
		`use Provider; use Consumer; join ",", Consumer::twice(21), Consumer::twice(4), Provider::calls()`)
	if want := "42,8,2"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
