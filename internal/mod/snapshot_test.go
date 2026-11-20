package mod

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A snapshot written by carton survives a parse/format round trip byte
// for byte.
func TestSnapshotRoundTrip(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "plack", SnapshotFile)
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := ReadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := snap.Format(); !bytes.Equal(got, want) {
		t.Fatalf("round trip changed the snapshot:\n%s", firstDifference(got, want))
	}

	dist := snap.Provider("Plack::Request")
	if dist == nil || !strings.HasPrefix(dist.Name, "Plack-") || !strings.HasSuffix(dist.Pathname, "/"+dist.Name+".tar.gz") {
		t.Fatalf("Provider(Plack::Request) = %+v", dist)
	}
	if snap.Provider("No::Such::Module") != nil {
		t.Fatal("Provider found a module the snapshot does not provide")
	}
}

func firstDifference(got, want []byte) string {
	g, w := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(g) && i < len(w); i++ {
		if g[i] != w[i] {
			return fmt.Sprintf("line %d: got %q, want %q", i+1, g[i], w[i])
		}
	}
	return "lengths differ"
}

// Format sorts what the resolver appended in discovery order.
func TestSnapshotFormatSorts(t *testing.T) {
	snap := &Snapshot{Distributions: []Distribution{
		{Name: "B-2", Pathname: "X/XX/XXX/B-2.tar.gz", Provides: []Package{{"B::Z", "2"}, {"B", "2"}}},
		{Name: "A-1", Pathname: "X/XX/XXX/A-1.tar.gz", Requirements: []Package{{"B", "1"}}},
	}}
	want := "# carton snapshot format: version 1.0\nDISTRIBUTIONS\n" +
		"  A-1\n    pathname: X/XX/XXX/A-1.tar.gz\n    provides:\n    requirements:\n      B 1\n" +
		"  B-2\n    pathname: X/XX/XXX/B-2.tar.gz\n    provides:\n      B 2\n      B::Z 2\n    requirements:\n"
	if got := string(snap.Format()); got != want {
		t.Fatalf("Format =\n%s\nwant\n%s", got, want)
	}
	if snap.Distributions[0].Name != "B-2" {
		t.Fatal("Format reordered the snapshot it was called on")
	}
}

func TestParseSnapshotRejectsMalformedInput(t *testing.T) {
	for name, src := range map[string]string{
		"field without colon":  "DISTRIBUTIONS\n  A-1\n    pathname\n",
		"unknown field":        "DISTRIBUTIONS\n  A-1\n    checksum: x\n",
		"entry without name":   "DISTRIBUTIONS\n  A-1\n    pathname: X/A-1.tar.gz\n    provides:\n      A\n",
		"entry outside a list": "DISTRIBUTIONS\n  A-1\n    pathname: X/A-1.tar.gz\n      A 1\n",
		"field before a dist":  "DISTRIBUTIONS\n    pathname: X/A-1.tar.gz\n",
		"missing pathname":     "DISTRIBUTIONS\n  A-1\n    provides:\n      A 1\n",
	} {
		if _, err := ParseSnapshot([]byte(src)); err == nil {
			t.Errorf("%s: parsed without error", name)
		}
	}
	if _, err := ReadSnapshot(filepath.Join(t.TempDir(), SnapshotFile)); !os.IsNotExist(err) {
		t.Errorf("ReadSnapshot of a missing file: %v", err)
	}
	bad := filepath.Join(t.TempDir(), SnapshotFile)
	if err := os.WriteFile(bad, []byte("DISTRIBUTIONS\n  A-1\n    checksum: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSnapshot(bad); err == nil || !strings.Contains(err.Error(), bad) {
		t.Errorf("ReadSnapshot error does not name the file: %v", err)
	}
}

func TestDirAndMirrorSettings(t *testing.T) {
	t.Setenv("GPERLMODCACHE", "")
	t.Setenv("HOME", "/home/gopher")
	if dir, err := Dir(); err != nil || dir != filepath.Join("/home/gopher", "gperl", "pkg", "mod") {
		t.Fatalf("default Dir = %q, %v", dir, err)
	}
	custom := t.TempDir()
	t.Setenv("GPERLMODCACHE", custom)
	if dir, err := Dir(); err != nil || dir != custom {
		t.Fatalf("Dir with GPERLMODCACHE = %q, %v", dir, err)
	}

	t.Setenv("GPERLMIRROR", "")
	if got := Mirror(); got != DefaultMirror {
		t.Fatalf("default Mirror = %q", got)
	}
	t.Setenv("GPERLMIRROR", "https://cpan.example.com/pub")
	if got := Mirror(); got != "https://cpan.example.com/pub/" {
		t.Fatalf("Mirror = %q", got)
	}
	for value, want := range map[string]bool{"": false, "0": false, "false": false, "1": true, "yes": true} {
		t.Setenv("GPERLMIRRORONLY", value)
		if got := MirrorOnly(); got != want {
			t.Errorf("MirrorOnly with %q = %v", value, got)
		}
	}
}

func TestDistributionPaths(t *testing.T) {
	for pathname, want := range map[string][2]string{
		"R/RU/RURBAN/Cpanel-JSON-XS-4.38.tar.gz": {"Cpanel-JSON-XS", "4.38"},
		"M/MI/MIYAGAWA/Plack-1.0050.tar.gz":      {"Plack", "1.0050"},
		"X/XX/XXX/Some-Dist-v1.2.3.tgz":          {"Some-Dist", "v1.2.3"},
		"X/XX/XXX/Dev-Dist-0.01_02.zip":          {"Dev-Dist", "0.01_02"},
	} {
		name, version, err := SplitPathname(pathname)
		if err != nil || name != want[0] || version != want[1] {
			t.Errorf("SplitPathname(%s) = %q, %q, %v", pathname, name, version, err)
		}
	}
	if _, _, err := SplitPathname("X/XX/XXX/README"); err == nil {
		t.Error("SplitPathname accepted a non-archive name")
	}
	if _, err := TreePath("/mod", "X/XX/XXX/README"); err == nil {
		t.Error("TreePath accepted a non-archive name")
	}
	tree, err := TreePath("/mod", "R/RU/RURBAN/Cpanel-JSON-XS-4.38.tar.gz")
	if err != nil || tree != filepath.Join("/mod", "Cpanel-JSON-XS@4.38") {
		t.Errorf("TreePath = %q, %v", tree, err)
	}
	if got, want := TarballPath("/mod", "R/RU/RURBAN/Cpanel-JSON-XS-4.38.tar.gz"),
		filepath.Join("/mod", "cache", "download", "authors", "id", "R", "RU", "RURBAN", "Cpanel-JSON-XS-4.38.tar.gz"); got != want {
		t.Errorf("TarballPath = %q", got)
	}
}
