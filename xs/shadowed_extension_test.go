//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// The version the shadowing Digest::MD5 of testdata/shadow declares, in its
// module file and in its shared object alike.
const shadowVersion = "9.99"

// buildShadowMD5 compiles testdata/shadow/MD5.c, a native Digest::MD5 whose
// md5_hex answers "shadow:<data>".
func buildShadowMD5(t *testing.T) string {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	so := filepath.Join(t.TempDir(), "MD5.so")
	args := []string{"-shared", "-fPIC", "-I", "sdk/include", `-DXS_VERSION="` + shadowVersion + `"`, "-o", so, "testdata/shadow/MD5.c"}
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile MD5.so: %v\n%s", err, out)
	}
	return so
}

// installShadowMD5 lays the shadowing module out the way an installation
// does: Digest/MD5.pm under libDir and auto/Digest/MD5/MD5.so under archDir
// (the same directory for an installed module, two directories for a built
// one that is run from its blib). The module file loads the extension with
// the loader statement given.
func installShadowMD5(t *testing.T, so, libDir, archDir, load string) {
	t.Helper()
	pm := "package Digest::MD5;\nour $VERSION = '" + shadowVersion + "';\n" + load + "\n1;\n"
	if err := os.MkdirAll(filepath.Join(libDir, "Digest"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libDir, "Digest", "MD5.pm"), []byte(pm), 0o644); err != nil {
		t.Fatal(err)
	}
	auto := filepath.Join(archDir, "auto", "Digest", "MD5")
	if err := os.MkdirAll(auto, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(so, filepath.Join(auto, "MD5.so")); err != nil {
		t.Fatal(err)
	}
}

// evalWithInc evaluates code on a fresh instance after the directories were
// put on @INC: ahead of the interpreter's library, or behind it.
func evalWithInc(t *testing.T, native bool, ahead bool, dirs []string, code string) string {
	t.Helper()
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, NativeModules: native})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	ctx := context.Background()
	vals := make([]perl.Value, len(dirs))
	for i, d := range dirs {
		vals[i] = perl.ValueOf(d)
	}
	if r, err := p.Eval(ctx, `sub __t_inc_ahead { unshift @INC, @_; 1 } sub __t_inc_behind { push @INC, @_; 1 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("define inc helpers: err=%v error=%v", err, r.Error)
	}
	helper := "__t_inc_behind"
	if ahead {
		helper = "__t_inc_ahead"
	}
	if _, err := p.Call(ctx, helper, vals...); err != nil {
		t.Fatalf("extend @INC: %v", err)
	}
	r, err := p.Eval(ctx, code)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("Eval died: %v", r.Error)
	}
	return resultStr(r)
}

const (
	digestOfABC  = "900150983cd24fb0d6963f7d28e17f72"
	shadowOfABC  = "shadow:abc " + shadowVersion
	useDigestMD5 = `use Digest::MD5; Digest::MD5::md5_hex("abc") . " " . Digest::MD5->VERSION`
)

// A module installed ahead of the interpreter's library replaces the
// linked-in extension of the same name: its module file is the one that is
// loaded, and the shared object installed with it is the one that is booted,
// whichever loader the module file uses and whether the shared object is
// beside the module file or in a directory of its own.
func TestInstalledExtensionShadowsLinkedInOne(t *testing.T) {
	so := buildShadowMD5(t)
	loaders := map[string]string{
		"XSLoader":   "require XSLoader;\nXSLoader::load('Digest::MD5', $VERSION);",
		"DynaLoader": "require DynaLoader;\nour @ISA = ('DynaLoader');\n__PACKAGE__->bootstrap($VERSION);",
	}
	for name, load := range loaders {
		t.Run(name+"/installed", func(t *testing.T) {
			dir := t.TempDir()
			installShadowMD5(t, so, dir, dir, load)
			if got := evalWithInc(t, true, true, []string{dir}, useDigestMD5); got != shadowOfABC {
				t.Fatalf("got %q, want %q", got, shadowOfABC)
			}
		})
		t.Run(name+"/blib", func(t *testing.T) {
			lib, arch := t.TempDir(), t.TempDir()
			installShadowMD5(t, so, lib, arch, load)
			if got := evalWithInc(t, true, true, []string{lib, arch}, useDigestMD5); got != shadowOfABC {
				t.Fatalf("got %q, want %q", got, shadowOfABC)
			}
		})
	}
}

// The linked-in extension is the one the interpreter's own module file
// gets: a shared object behind the library in @INC is not reached, and one
// ahead of it is not used by an interpreter that loads no native code.
func TestLinkedInExtensionWhereNoSharedObjectShadowsIt(t *testing.T) {
	so := buildShadowMD5(t)
	want := digestOfABC
	code := `use Digest::MD5; Digest::MD5::md5_hex("abc")`

	t.Run("behind the library", func(t *testing.T) {
		dir := t.TempDir()
		installShadowMD5(t, so, dir, dir, "require XSLoader;\nXSLoader::load('Digest::MD5', $VERSION);")
		if got := evalWithInc(t, true, false, []string{dir}, code); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
	t.Run("no native modules", func(t *testing.T) {
		// Only the shared object is ahead of the library: the module file
		// is the interpreter's own.
		lib, arch := t.TempDir(), t.TempDir()
		installShadowMD5(t, so, lib, arch, "")
		if got := evalWithInc(t, false, true, []string{arch}, code); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}
