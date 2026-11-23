package mod

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// fakeDist is one release on the fake CPAN.
type fakeDist struct {
	pathname string // below authors/id
	module   string // the module MetaDB indexes it under
	version  string
	requires map[string]string // runtime requires
	tests    map[string]string // test requires
}

// fakeCPAN serves the three services the manager talks to: a mirror
// (tarballs, CHECKSUMS, 02packages), MetaDB, and MetaCPAN.
type fakeCPAN struct {
	t     *testing.T
	srv   *httptest.Server
	dists []fakeDist

	mu           sync.Mutex
	tarballs     map[string][]byte
	hits         map[string]int
	noMetaDB     map[string]bool // modules MetaDB does not know
	corrupt      map[string]bool // pathnames whose CHECKSUMS entry is wrong
	noChecksums  bool
	metaCPANDown bool
}

func newFakeCPAN(t *testing.T, dists ...fakeDist) *fakeCPAN {
	c := &fakeCPAN{t: t, dists: dists, tarballs: map[string][]byte{}, hits: map[string]int{}, noMetaDB: map[string]bool{}, corrupt: map[string]bool{}}
	c.srv = httptest.NewServer(http.HandlerFunc(c.serve))
	t.Cleanup(c.srv.Close)
	return c
}

// tarball returns the release's archive, built once so that its checksum
// is stable.
func (c *fakeCPAN) tarball(d fakeDist) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	if data, ok := c.tarballs[d.pathname]; ok {
		return data
	}
	data := c.buildTarball(d)
	c.tarballs[d.pathname] = data
	return data
}

func (c *fakeCPAN) buildTarball(d fakeDist) []byte {
	base := strings.TrimSuffix(path.Base(d.pathname), ".tar.gz")
	prereq := func(reqs map[string]string) string {
		var parts []string
		for m, v := range reqs {
			parts = append(parts, fmt.Sprintf("%q: %q", m, v))
		}
		sort.Strings(parts)
		return "{" + strings.Join(parts, ", ") + "}"
	}
	meta := fmt.Sprintf(`{"abstract":"x","author":["a"],"dynamic_config":0,"generated_by":"t","license":["perl_5"],
"meta-spec":{"version":2},"name":%q,"version":%q,
"prereqs":{"runtime":{"requires":%s},"test":{"requires":%s}}}`,
		strings.ReplaceAll(d.module, "::", "-"), d.version, prereq(d.requires), prereq(d.tests))
	pm := fmt.Sprintf("package %s;\nour $VERSION = '%s';\n1;\n", d.module, d.version)
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zw)
	for _, member := range [][2]string{
		{base + "/META.json", meta},
		{base + "/lib/" + strings.ReplaceAll(d.module, "::", "/") + ".pm", pm},
	} {
		name, body := member[0], member[1]
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			c.t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			c.t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		c.t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		c.t.Fatal(err)
	}
	return buf.Bytes()
}

func (c *fakeCPAN) serve(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	c.hits[r.URL.Path]++
	c.mu.Unlock()
	p := r.URL.Path
	switch {
	case strings.HasPrefix(p, "/metadb/package/"):
		module := strings.TrimPrefix(p, "/metadb/package/")
		var latest *fakeDist
		for i := range c.dists {
			if c.dists[i].module == module {
				latest = &c.dists[i]
			}
		}
		if latest == nil || c.noMetaDB[module] {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "---\ndistfile: %s\nprovides:\n  %s: %s\nversion: %s\n", latest.pathname, module, latest.version, latest.version)
	case strings.HasPrefix(p, "/metadb/history/"):
		module := strings.TrimPrefix(p, "/metadb/history/")
		found := false
		for _, d := range c.dists {
			if d.module == module && !c.noMetaDB[module] {
				found = true
				fmt.Fprintf(w, "%s %s %s\n", d.module, d.version, d.pathname)
			}
		}
		if !found {
			http.NotFound(w, r)
		}
	case strings.HasPrefix(p, "/metacpan/"):
		module := strings.TrimPrefix(p, "/metacpan/")
		if c.metaCPANDown {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		want := strings.TrimPrefix(r.URL.Query().Get("version"), "== ")
		for i := len(c.dists) - 1; i >= 0; i-- {
			if d := c.dists[i]; d.module == module && (want == "" || want == d.version) {
				fmt.Fprintf(w, `{"download_url":"https://cpan.metacpan.org/authors/id/%s","version":%q}`, d.pathname, d.version)
				return
			}
		}
		http.NotFound(w, r)
	case p == "/mirror/modules/02packages.details.txt.gz":
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		fmt.Fprint(zw, "File: 02packages.details.txt\nLine-Count: 0\n\n")
		latest := map[string]fakeDist{}
		for _, d := range c.dists {
			latest[d.module] = d
		}
		for _, d := range latest {
			fmt.Fprintf(zw, "%-30s %8s  %s\n", d.module, d.version, d.pathname)
		}
		if err := zw.Close(); err != nil {
			c.t.Error(err)
		}
		if _, err := w.Write(buf.Bytes()); err != nil {
			c.t.Error(err)
		}
	case strings.HasSuffix(p, "/CHECKSUMS"):
		if c.noChecksums {
			http.NotFound(w, r)
			return
		}
		author := strings.TrimSuffix(strings.TrimPrefix(p, "/mirror/authors/id/"), "CHECKSUMS")
		fmt.Fprint(w, "# CHECKSUMS file\n$cksum = {\n")
		for _, d := range c.dists {
			if !strings.HasPrefix(d.pathname, author) {
				continue
			}
			sum := sha256.Sum256(c.tarball(d))
			hexsum := hex.EncodeToString(sum[:])
			if c.corrupt[d.pathname] {
				hexsum = strings.Repeat("0", 64)
			}
			fmt.Fprintf(w, "  '%s' => {\n    'md5' => 'x',\n    'sha256' => '%s',\n    'size' => 1\n  },\n", path.Base(d.pathname), hexsum)
		}
		fmt.Fprint(w, "};\n")
	case strings.HasPrefix(p, "/mirror/authors/id/"):
		for _, d := range c.dists {
			if "/mirror/authors/id/"+d.pathname == p {
				if _, err := w.Write(c.tarball(d)); err != nil {
					c.t.Error(err)
				}
				return
			}
		}
		http.NotFound(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (c *fakeCPAN) manager(t *testing.T) *Manager {
	t.Helper()
	return &Manager{
		Dir:       filepath.Join(t.TempDir(), "mod"),
		Mirror:    c.srv.URL + "/mirror/",
		MetaDB:    c.srv.URL + "/metadb/",
		MetaCPAN:  c.srv.URL + "/metacpan/",
		HTTP:      c.srv.Client(),
		Log:       io.Discard,
		helper:    sharedHelper,
		checksums: map[string]*checksumSet{},
		bundled:   map[string]bundledModule{},
	}
}

func (c *fakeCPAN) count(prefix string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for p, hits := range c.hits {
		if strings.HasPrefix(p, prefix) {
			n += hits
		}
	}
	return n
}

// The fake CPAN every test starts from: App::A needs B::Lib (two releases)
// and a bundled module; its test prereq must not be pulled in.
func standardCPAN(t *testing.T) *fakeCPAN {
	return newFakeCPAN(t,
		fakeDist{pathname: "A/AA/AAA/App-A-1.00.tar.gz", module: "App::A", version: "1.00",
			requires: map[string]string{"B::Lib": "1.0", "File::Spec": "0", "perl": "5.008"},
			tests:    map[string]string{"Test::Never::Fetched": "0"}},
		fakeDist{pathname: "B/BB/BBB/B-Lib-1.50.tar.gz", module: "B::Lib", version: "1.50"},
		fakeDist{pathname: "B/BB/BBB/B-Lib-2.00.tar.gz", module: "B::Lib", version: "2.00"},
		fakeDist{pathname: "C/CC/CCC/C-Only-0.10.tar.gz", module: "C::Only", version: "0.10"},
	)
}

func project(t *testing.T, cpanfile string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, CPANfile), cpanfile)
	return dir
}

func distNames(s *Snapshot) string {
	if s == nil {
		return "<no snapshot>"
	}
	var names []string
	for _, d := range s.Distributions {
		names = append(names, d.Name)
	}
	return strings.Join(names, " ")
}

func TestTidyResolvesTheClosure(t *testing.T) {
	c := standardCPAN(t)
	m := c.manager(t)
	dir := project(t, "requires 'App::A';\n")
	snap, err := m.Tidy(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := distNames(snap); got != "App-A-1.00 B-Lib-2.00" {
		t.Fatalf("resolved %q", got)
	}
	want := "# carton snapshot format: version 1.0\nDISTRIBUTIONS\n" +
		"  App-A-1.00\n    pathname: A/AA/AAA/App-A-1.00.tar.gz\n    provides:\n      App::A 1.00\n" +
		"    requirements:\n      B::Lib 1.0\n      File::Spec 0\n      perl 5.008\n" +
		"  B-Lib-2.00\n    pathname: B/BB/BBB/B-Lib-2.00.tar.gz\n    provides:\n      B::Lib 2.00\n    requirements:\n"
	got, err := os.ReadFile(filepath.Join(dir, SnapshotFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("snapshot =\n%s\nwant\n%s", got, want)
	}
	for _, f := range []string{"App-A@1.00/lib/App/A.pm", "B-Lib@2.00/META.json", "cache/download/authors/id/B/BB/BBB/B-Lib-2.00.tar.gz"} {
		if _, err := os.Stat(filepath.Join(m.Dir, filepath.FromSlash(f))); err != nil {
			t.Errorf("module directory: %v", err)
		}
	}

	// A second run finds everything pinned and present: no lookups, no
	// downloads.
	before := c.count("/")
	if _, err := m.Tidy(context.Background(), dir); err != nil {
		t.Fatal(err)
	}
	if after := c.count("/"); after != before {
		t.Fatalf("a tidy with nothing to do made %d requests", after-before)
	}
}

func TestTidyVersionRangesAndPins(t *testing.T) {
	c := standardCPAN(t)
	m := c.manager(t)
	dir := project(t, "requires 'B::Lib', '< 2.0';\n")
	snap, err := m.Tidy(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := distNames(snap); got != "B-Lib-1.50" {
		t.Fatalf("with < 2.0 resolved %q", got)
	}

	// Loosening the cpanfile keeps the pinned release ...
	writeFile(t, filepath.Join(dir, CPANfile), "requires 'B::Lib';\n")
	if snap, err = m.Tidy(context.Background(), dir); err != nil || distNames(snap) != "B-Lib-1.50" {
		t.Fatalf("pin not kept: %q, %v", distNames(snap), err)
	}
	// ... until the module is asked for again.
	if snap, err = m.Get(context.Background(), dir, "B::Lib"); err != nil || distNames(snap) != "B-Lib-2.00" {
		t.Fatalf("Get did not upgrade: %q, %v", distNames(snap), err)
	}
	if snap, err = m.Get(context.Background(), dir, "B::Lib@1.50", "C::Only"); err != nil || distNames(snap) != "B-Lib-1.50 C-Only-0.10" {
		t.Fatalf("Get with a version: %q, %v", distNames(snap), err)
	}
	cpanfile, err := os.ReadFile(filepath.Join(dir, CPANfile))
	if err != nil {
		t.Fatal(err)
	}
	if want := "requires 'B::Lib', '== 1.50';\nrequires 'C::Only';\n"; string(cpanfile) != want {
		t.Fatalf("cpanfile = %q", cpanfile)
	}
	if _, err := m.Get(context.Background(), dir, "not a module"); err == nil {
		t.Fatal("Get accepted a malformed module spec")
	}
}

func TestGetCreatesTheCPANfile(t *testing.T) {
	c := standardCPAN(t)
	m := c.manager(t)
	dir := t.TempDir()
	snap, err := m.Get(context.Background(), dir, "C::Only")
	if err != nil || distNames(snap) != "C-Only-0.10" {
		t.Fatalf("Get = %q, %v", distNames(snap), err)
	}
}

func TestResolveFallsBackToMetaCPAN(t *testing.T) {
	c := standardCPAN(t)
	c.noMetaDB["C::Only"] = true
	m := c.manager(t)
	snap, err := m.Tidy(context.Background(), project(t, "requires 'C::Only';\nrequires 'B::Lib', '== 1.50';\n"))
	if err != nil || distNames(snap) != "B-Lib-1.50 C-Only-0.10" {
		t.Fatalf("resolved %q, %v", distNames(snap), err)
	}
	if c.count("/metacpan/C::Only") == 0 {
		t.Fatal("MetaCPAN was not asked for the module MetaDB lacks")
	}

	c.metaCPANDown = true
	if _, err := c.manager(t).Tidy(context.Background(), project(t, "requires 'C::Only';\n")); err == nil {
		t.Fatal("resolution succeeded with MetaDB lacking the module and MetaCPAN down")
	}
}

func TestResolveErrors(t *testing.T) {
	c := standardCPAN(t)
	for name, cpanfile := range map[string]string{
		"unknown module":      "requires 'No::Such::Thing';\n",
		"unsatisfiable range": "requires 'B::Lib', '> 9';\n",
		"range with no match": "requires 'B::Lib', '== 7.77';\n",
		"newer perl":          "requires 'perl', '7.0';\n",
		"conflict":            "requires 'B::Lib', '== 1.50';\nrequires 'App::A';\nrequires 'B::Lib', '>= 2.0';\n",
	} {
		if _, err := c.manager(t).Tidy(context.Background(), project(t, cpanfile)); err == nil {
			t.Errorf("%s: Tidy succeeded", name)
		}
	}
	if _, err := c.manager(t).Tidy(context.Background(), t.TempDir()); err == nil {
		t.Error("Tidy succeeded without a cpanfile")
	}
	dir := project(t, "requires 'C::Only';\n")
	writeFile(t, filepath.Join(dir, SnapshotFile), "DISTRIBUTIONS\n  X-1\n    bogus: 1\n")
	if _, err := c.manager(t).Tidy(context.Background(), dir); err == nil {
		t.Error("Tidy succeeded over a malformed snapshot")
	}
	if _, err := c.manager(t).Download(context.Background(), dir); err == nil {
		t.Error("Download succeeded over a malformed snapshot")
	}
}

func TestMirrorOnlyUsesTheIndex(t *testing.T) {
	c := standardCPAN(t)
	m := c.manager(t)
	m.MirrorOnly = true
	snap, err := m.Tidy(context.Background(), project(t, "requires 'App::A';\n"))
	if err != nil || distNames(snap) != "App-A-1.00 B-Lib-2.00" {
		t.Fatalf("resolved %q, %v", distNames(snap), err)
	}
	if n := c.count("/metadb/") + c.count("/metacpan/"); n != 0 {
		t.Fatalf("mirror-only resolution made %d requests to MetaDB/MetaCPAN", n)
	}
	if n := c.count("/mirror/modules/"); n != 1 {
		t.Fatalf("the index was fetched %d times", n)
	}
	for _, cpanfile := range []string{"requires 'B::Lib', '< 2.0';\n", "requires 'No::Such::Thing';\n"} {
		if _, err := m.Tidy(context.Background(), project(t, cpanfile)); err == nil {
			t.Errorf("mirror-only Tidy of %q succeeded", cpanfile)
		}
	}
}

func TestDownloadFollowsTheSnapshot(t *testing.T) {
	c := standardCPAN(t)
	dir := project(t, "requires 'App::A';\n")
	if _, err := c.manager(t).Tidy(context.Background(), dir); err != nil {
		t.Fatal(err)
	}

	// A fresh module directory is filled from the snapshot alone.
	m := c.manager(t)
	lookups := c.count("/metadb/") + c.count("/metacpan/")
	snap, err := m.Download(context.Background(), dir)
	if err != nil || len(snap.Distributions) != 2 {
		t.Fatalf("Download = %v, %v", snap, err)
	}
	if n := c.count("/metadb/") + c.count("/metacpan/"); n != lookups {
		t.Fatalf("Download resolved again (%d lookups)", n-lookups)
	}
	if _, err := os.Stat(filepath.Join(m.Dir, "B-Lib@2.00", "lib", "B", "Lib.pm")); err != nil {
		t.Fatal(err)
	}
	requests := c.count("/")
	if _, err := m.Download(context.Background(), dir); err != nil || c.count("/") != requests {
		t.Fatalf("a Download with nothing to do: err=%v, %d requests", err, c.count("/")-requests)
	}

	// Without a snapshot Download resolves; without a cpanfile there is
	// nothing to do.
	fresh := project(t, "requires 'C::Only';\n")
	if snap, err := c.manager(t).Download(context.Background(), fresh); err != nil || distNames(snap) != "C-Only-0.10" {
		t.Fatalf("Download without a snapshot = %q, %v", distNames(snap), err)
	}
	if snap, err := c.manager(t).Download(context.Background(), t.TempDir()); err != nil || len(snap.Distributions) != 0 {
		t.Fatalf("Download without a cpanfile = %v, %v", snap, err)
	}
}

func TestChecksumVerification(t *testing.T) {
	c := standardCPAN(t)
	c.corrupt["C/CC/CCC/C-Only-0.10.tar.gz"] = true
	m := c.manager(t)
	_, err := m.Tidy(context.Background(), project(t, "requires 'C::Only';\n"))
	if err == nil || !strings.Contains(err.Error(), "CHECKSUMS") {
		t.Fatalf("a tarball that fails its checksum was accepted: %v", err)
	}
	entries, err := os.ReadDir(filepath.Dir(TarballPath(m.Dir, "C/CC/CCC/C-Only-0.10.tar.gz")))
	if err != nil || len(entries) != 0 {
		t.Fatalf("the rejected download left files behind: %v, %v", entries, err)
	}

	// A mirror without CHECKSUMS is used unverified, and says so.
	c.noChecksums = true
	var log bytes.Buffer
	m = c.manager(t)
	m.Log = &log
	if _, err := m.Tidy(context.Background(), project(t, "requires 'C::Only';\n")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(log.String(), "not verified") {
		t.Fatalf("no notice about the missing CHECKSUMS: %q", log.String())
	}
}

func TestClean(t *testing.T) {
	c := standardCPAN(t)
	m := c.manager(t)
	if _, err := m.Tidy(context.Background(), project(t, "requires 'C::Only';\n")); err != nil {
		t.Fatal(err)
	}
	if err := m.Clean(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(m.Dir); !os.IsNotExist(err) {
		t.Fatalf("module directory still exists: %v", err)
	}
}

func TestNewManagerReadsTheEnvironment(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GPERLMODCACHE", dir)
	t.Setenv("GPERLMIRROR", "https://cpan.example.com")
	t.Setenv("GPERLMIRRORONLY", "1")
	m, err := NewManager()
	if err != nil {
		t.Fatal(err)
	}
	if m.Dir != dir || m.Mirror != "https://cpan.example.com/" || !m.MirrorOnly {
		t.Fatalf("NewManager = %+v", m)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := m.helper.Best("0", "1"); err == nil {
		t.Fatal("a closed manager's helper still answers")
	}
}

func TestExtractArchiveFormats(t *testing.T) {
	dir := t.TempDir()

	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	for name, body := range map[string]string{"Zip-Dist-1.0/": "", "Zip-Dist-1.0/lib/Zip.pm": "package Zip;1;"} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(dir, "Zip-Dist-1.0.zip")
	if err := os.WriteFile(zipPath, zbuf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := extract(zipPath, filepath.Join(dir, "Zip-Dist@1.0")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "Zip-Dist@1.0", "lib", "Zip.pm")); err != nil {
		t.Fatal(err)
	}
	// Extracting over a finished tree leaves it alone.
	if err := extract(zipPath, filepath.Join(dir, "Zip-Dist@1.0")); err != nil {
		t.Fatalf("second extract: %v", err)
	}

	tarWith := func(name string, hdrs ...*tar.Header) string {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gz)
		for _, h := range hdrs {
			if err := tw.WriteHeader(h); err != nil {
				t.Fatal(err)
			}
			if h.Typeflag == tar.TypeReg {
				if _, err := tw.Write(make([]byte, h.Size)); err != nil {
					t.Fatal(err)
				}
			}
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	ok := tarWith("ok.tar.gz",
		&tar.Header{Name: "D-1/", Typeflag: tar.TypeDir, Mode: 0o755},
		&tar.Header{Name: "D-1/bin/tool", Typeflag: tar.TypeReg, Mode: 0o755, Size: 3},
		&tar.Header{Name: "D-1/bin/alias", Typeflag: tar.TypeSymlink, Linkname: "tool"})
	if err := extract(ok, filepath.Join(dir, "D@1")); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(filepath.Join(dir, "D@1", "bin", "alias")); err != nil || fi.Mode()&0o100 == 0 {
		t.Fatalf("symlinked executable: %v, %v", fi, err)
	}
	for name, hdr := range map[string]*tar.Header{
		"escaping file": {Name: "../evil", Typeflag: tar.TypeReg, Mode: 0o644, Size: 1},
		"escaping link": {Name: "D-1/link", Typeflag: tar.TypeSymlink, Linkname: "../../../etc/passwd"},
	} {
		bad := tarWith("bad.tar.gz", hdr)
		if err := extract(bad, filepath.Join(dir, "Bad@1")); err == nil {
			t.Errorf("%s: extracted", name)
		}
		if _, err := os.Stat(filepath.Join(dir, "Bad@1")); !os.IsNotExist(err) {
			t.Errorf("%s: left a tree behind", name)
		}
	}
	// A link to an absolute path, and a file written through it, write
	// nothing outside the tree.
	outside := t.TempDir()
	through := tarWith("through.tar.gz",
		&tar.Header{Name: "D-1/abs", Typeflag: tar.TypeSymlink, Linkname: outside},
		&tar.Header{Name: "D-1/abs/pwned", Typeflag: tar.TypeReg, Mode: 0o644, Size: 1})
	if err := extract(through, filepath.Join(dir, "Through@1")); err == nil {
		t.Error("absolute link: extracted")
	}
	if _, err := os.Stat(filepath.Join(outside, "pwned")); !os.IsNotExist(err) {
		t.Errorf("a file was written outside the tree: %v", err)
	}
	// Links that each name a path inside, but lead outside one through the
	// other, are refused.
	chain := tarWith("chain.tar.gz",
		&tar.Header{Name: "D-1/x", Typeflag: tar.TypeSymlink, Linkname: "."},
		&tar.Header{Name: "D-1/y", Typeflag: tar.TypeSymlink, Linkname: "x/../.."})
	if err := extract(chain, filepath.Join(dir, "Chain@1")); err == nil {
		t.Error("chained links: extracted")
	}
	if err := extract(filepath.Join(dir, "missing.tar.gz"), filepath.Join(dir, "Missing@1")); err == nil {
		t.Error("extracted a missing archive")
	}
	if err := os.WriteFile(filepath.Join(dir, "garbage.zip"), []byte("not a zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := extract(filepath.Join(dir, "garbage.zip"), filepath.Join(dir, "Garbage@1")); err == nil {
		t.Error("extracted a corrupt zip")
	}
}
