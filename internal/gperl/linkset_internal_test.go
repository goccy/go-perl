package gperl

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-perl/internal/mod"
)

// What gperl generates from the dependencies — the merged module tree and
// the link set — is generated again by another gperl.
func TestGeneratedDirectoriesAreKeyedByGperl(t *testing.T) {
	m := &mod.Manager{Dir: t.TempDir()}
	saved := mod.GperlVersion
	t.Cleanup(func() { mod.GperlVersion = saved })
	type dirs struct{ lib, linkSet string }
	generate := func(version string) dirs {
		t.Helper()
		mod.GperlVersion = func() string { return version }
		lib, err := mergedLib(m, nil, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		links, err := generateLinkSet(m, &projectDeps{lib: map[string]string{}})
		if err != nil {
			t.Fatal(err)
		}
		return dirs{lib: lib, linkSet: links.cpanDir}
	}
	a, b := generate("gperl a"), generate("gperl b")
	if a.lib == b.lib || a.linkSet == b.linkSet {
		t.Fatalf("two gperls share %+v and %+v", a, b)
	}
	if again := generate("gperl a"); again != a {
		t.Fatalf("one gperl generated %+v, then %+v", a, again)
	}
}

// The link set carries the module tree as one stored zip archive, in name
// order, so two names that differ only in case both reach the binary.
func TestLinkSetCarriesNamesDifferingInCase(t *testing.T) {
	m := &mod.Manager{Dir: t.TempDir()}
	lib := map[string]string{}
	for name, content := range map[string]string{"Case/Name.pm": "1;\n", "Case/NAME.pm": "2;\n"} {
		host := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(host, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		lib[name] = host
	}
	links, err := generateLinkSet(m, &projectDeps{lib: lib})
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(filepath.Join(links.cpanDir, "lib.zip"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := zr.Close(); err != nil {
			t.Error(err)
		}
	})
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
		if f.Method != zip.Store {
			t.Errorf("%s: method %d, want stored", f.Name, f.Method)
		}
	}
	if want := []string{"Case/NAME.pm", "Case/Name.pm"}; !slices.Equal(names, want) {
		t.Fatalf("lib.zip holds %q, want %q", names, want)
	}
	for name, want := range map[string]string{"Case/Name.pm": "1;\n", "Case/NAME.pm": "2;\n"} {
		if got, err := fs.ReadFile(zr, name); err != nil || string(got) != want {
			t.Errorf("%s = %q, %v; want %q", name, got, err, want)
		}
	}
	src, err := os.ReadFile(filepath.Join(links.cpanDir, "cpan.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "//go:embed lib.zip\n") {
		t.Fatalf("cpan.go does not embed lib.zip:\n%s", src)
	}
}

// A program the module tree holds (an Alien distribution's) stays one where
// a binary unpacks the tree; the archive carries no times.
func TestLibZipKeepsModes(t *testing.T) {
	dir := t.TempDir()
	lib := map[string]string{}
	for name, mode := range map[string]os.FileMode{"auto/share/dist/A/bin/tool": 0o755, "A.pm": 0o644} {
		host := filepath.Join(dir, strings.ReplaceAll(name, "/", "_"))
		if err := os.WriteFile(host, []byte("x\n"), mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(host, mode); err != nil {
			t.Fatal(err)
		}
		lib[name] = host
	}
	archive := filepath.Join(dir, "lib.zip")
	if err := writeLibZip(lib, archive); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := zr.Close(); err != nil {
			t.Error(err)
		}
	})
	// Other modification times make the same archive.
	old := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, host := range lib {
		if err := os.Chtimes(host, old, old); err != nil {
			t.Fatal(err)
		}
	}
	again := filepath.Join(dir, "again.zip")
	if err := writeLibZip(lib, again); err != nil {
		t.Fatal(err)
	}
	a, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(again)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Error("the archive changed with the files' modification times")
	}
	out := filepath.Join(t.TempDir(), "tree")
	if err := os.CopyFS(out, zr); err != nil {
		t.Fatal(err)
	}
	for name, exec := range map[string]bool{"auto/share/dist/A/bin/tool": true, "A.pm": false} {
		info, err := os.Stat(filepath.Join(out, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm()&0o100 != 0; got != exec {
			t.Errorf("%s unpacked as %v, want executable %v", name, info.Mode(), exec)
		}
	}
}
