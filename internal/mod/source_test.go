package mod

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// gitRepo makes a repository whose two commits carry Fork 1.0 (tagged
// v1.0) and then Fork 2.0, both with Fork::Util, and serves a bare clone of
// it over git's dumb HTTP protocol. It returns the clone's URL and a count
// of the clones made from it.
func gitRepo(t *testing.T) (string, func() int) {
	t.Helper()
	root := t.TempDir()
	work := filepath.Join(root, "work")
	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	run(work, "init", "--quiet", "--initial-branch=main")
	for _, version := range []string{"1.0", "2.0"} {
		writeFile(t, filepath.Join(work, "lib", "Fork.pm"), "package Fork;\nour $VERSION = '"+version+"';\n1;\n")
		writeFile(t, filepath.Join(work, "lib", "Fork", "Util.pm"), "package Fork::Util;\nour $VERSION = '"+version+"';\n1;\n")
		writeFile(t, filepath.Join(work, "META.json"), `{"abstract":"x","author":["a"],"dynamic_config":0,"generated_by":"t","license":["perl_5"],
"meta-spec":{"version":2},"name":"Fork","version":"`+version+`","prereqs":{"runtime":{"requires":{"B::Lib":"1.0"}}}}`)
		run(work, "add", ".")
		run(work, "commit", "--quiet", "-m", version)
		if version == "1.0" {
			run(work, "tag", "v1.0")
		}
	}
	run(root, "clone", "--quiet", "--bare", work, "Fork.git")
	run(filepath.Join(root, "Fork.git"), "update-server-info")
	var mu sync.Mutex
	clones := 0
	files := http.FileServer(http.Dir(root))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A clone starts by reading the repository's refs.
		if r.URL.Path == "/Fork.git/info/refs" {
			mu.Lock()
			clones++
			mu.Unlock()
		}
		files.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/Fork.git", func() int {
		mu.Lock()
		defer mu.Unlock()
		return clones
	}
}

func TestTidyPinnedSources(t *testing.T) {
	c := standardCPAN(t)
	// A second mirror, as a DarkPAN would be: its releases are not on
	// the CPAN MetaDB and MetaCPAN know.
	dark := newFakeCPAN(t,
		fakeDist{pathname: "M/MY/MYCO/Dark-Lib-0.10.tar.gz", module: "Dark::Lib", version: "0.10"},
		fakeDist{pathname: "U/UU/UUU/U-Dist-0.50.tar.gz", module: "U::Dist", version: "0.50"},
	)
	repo, clones := gitRepo(t)
	m := c.manager(t)
	// Fork::Util, which is on no CPAN, comes with the distribution the
	// cpanfile names for Fork.
	dir := project(t, `
requires 'B::Lib', dist => 'BBB/B-Lib-1.50.tar.gz';
requires 'Dark::Lib', dist => 'MYCO/Dark-Lib-0.10.tar.gz', mirror => '`+dark.srv.URL+`/mirror';
requires 'U::Dist', '0.5', url => '`+dark.srv.URL+`/mirror/authors/id/U/UU/UUU/U-Dist-0.50.tar.gz';
requires 'Fork', git => '`+repo+`', ref => 'v1.0';
requires 'Fork::Util';
`)
	snap, err := m.Tidy(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, SnapshotFile))
	if err != nil {
		t.Fatal(err)
	}
	// Fork requires B::Lib 1.0, which the release the cpanfile pins
	// meets.
	want := regexp.MustCompile(`^# carton snapshot format: version 1\.0
DISTRIBUTIONS
  B-Lib-1\.50
    pathname: B/BB/BBB/B-Lib-1\.50\.tar\.gz
    provides:
      B::Lib 1\.50
    requirements:
  Dark-Lib-0\.10
    pathname: M/MY/MYCO/Dark-Lib-0\.10\.tar\.gz
    mirror: http://127\.0\.0\.1:\d+/mirror/
    sha256: [0-9a-f]{64}
    provides:
      Dark::Lib 0\.10
    requirements:
  Fork-([0-9a-f]{12})
    pathname: Fork-([0-9a-f]{12})\.tar\.gz
    git: http://127\.0\.0\.1:\d+/Fork\.git
    ref: v1\.0
    revision: ([0-9a-f]{40})
    provides:
      Fork 1\.0
      Fork::Util 1\.0
    requirements:
      B::Lib 1\.0
  U-Dist-0\.50
    pathname: U-Dist-0\.50\.tar\.gz
    url: http://127\.0\.0\.1:\d+/mirror/authors/id/U/UU/UUU/U-Dist-0\.50\.tar\.gz
    sha256: [0-9a-f]{64}
    provides:
      U::Dist 0\.50
    requirements:
$`)
	match := want.FindStringSubmatch(string(got))
	if match == nil || match[1] != match[2] || !strings.HasPrefix(match[3], match[1]) {
		t.Fatalf("snapshot =\n%s", got)
	}
	parsed, err := ParseSnapshot(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(parsed.Format()) != string(got) {
		t.Fatalf("the snapshot does not survive a round trip:\n%s", parsed.Format())
	}
	for _, d := range snap.Distributions {
		tree, err := TreePath(m.Dir, &d)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(tree, "META.json")); err != nil {
			t.Errorf("%s: %v", d.Name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(m.Dir, "Fork@git."+match[1], ".git")); err == nil {
		t.Error("the tree of a git distribution carries the repository")
	}

	// A second run keeps every pin and makes no request.
	before := c.count("/") + dark.count("/")
	if _, err := m.Tidy(context.Background(), dir); err != nil {
		t.Fatal(err)
	}
	if after := c.count("/") + dark.count("/"); after != before {
		t.Fatalf("a tidy with nothing to do made %d requests", after-before)
	}

	// A fresh module directory is filled from the snapshot, every archive
	// checked against it.
	fresh := c.manager(t)
	if _, err := fresh.Download(context.Background(), dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(fresh.Dir, "Fork@git."+match[1], "lib", "Fork.pm")); err != nil {
		t.Fatal(err)
	}
	tampered := regexp.MustCompile(`(url: \S+\n    sha256: )[0-9a-f]{64}`).ReplaceAllString(string(got), "${1}"+strings.Repeat("0", 64))
	writeFile(t, filepath.Join(dir, SnapshotFile), tampered)
	if _, err := c.manager(t).Download(context.Background(), dir); err == nil || !strings.Contains(err.Error(), "does not match the snapshot") {
		t.Fatalf("Download of an archive that does not match the snapshot: %v", err)
	}

	// Two modules the cpanfile takes from one repository clone it once.
	before = clones()
	both := project(t, "requires 'Fork', git => '"+repo+"';\nrequires 'Fork::Util', '2.0', git => '"+repo+"';\n")
	if _, err := c.manager(t).Tidy(context.Background(), both); err != nil {
		t.Fatal(err)
	}
	if n := clones() - before; n != 1 {
		t.Fatalf("two modules of one repository cloned it %d times", n)
	}

	// The modules of one repository move together: the old snapshot's
	// commit stays for all of them, or upgrading one fetches it for all.
	moved := project(t, "requires 'Fork', git => '"+repo+"', ref => 'v1.0';\nrequires 'Fork::Util', git => '"+repo+"', ref => 'v1.0';\n")
	mm := c.manager(t)
	if _, err := mm.Tidy(context.Background(), moved); err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(filepath.Join(moved, SnapshotFile))
	if err != nil {
		t.Fatal(err)
	}
	// As if the default branch had been at v1.0 when the snapshot was made.
	writeFile(t, filepath.Join(moved, SnapshotFile), strings.Replace(string(old), "    ref: v1.0\n", "", 1))
	writeFile(t, filepath.Join(moved, CPANfile), "requires 'Fork', git => '"+repo+"';\nrequires 'Fork::Util', git => '"+repo+"';\n")
	for _, c := range []struct {
		upgrade []string
		want    string
	}{{nil, "1.0"}, {[]string{"Fork"}, "2.0"}} {
		snap, err := mm.Tidy(context.Background(), moved, c.upgrade...)
		if err != nil {
			t.Fatalf("Tidy upgrading %v: %v", c.upgrade, err)
		}
		if v := providedVersion(snap.Provider("Fork::Util"), "Fork::Util"); v != c.want {
			t.Fatalf("Tidy upgrading %v gave Fork::Util %s, want %s", c.upgrade, v, c.want)
		}
	}

	// Without the ref, the default branch is resolved.
	writeFile(t, filepath.Join(dir, SnapshotFile), string(got))
	writeFile(t, filepath.Join(dir, CPANfile), "requires 'Fork', '2.0', git => '"+repo+"';\n")
	if snap, err = m.Tidy(context.Background(), dir); err != nil || !strings.HasPrefix(distNames(snap), "B-Lib-1.50 Fork-") {
		t.Fatalf("Tidy with the default branch = %q, %v", distNames(snap), err)
	}
	if v := providedVersion(snap.Provider("Fork"), "Fork"); v != "2.0" {
		t.Fatalf("the default branch gave Fork %s", v)
	}
}

func TestTidyPinnedSourceErrors(t *testing.T) {
	c := standardCPAN(t)
	url := c.srv.URL + "/mirror/authors/id/C/CC/CCC/C-Only-0.10.tar.gz"
	for _, tc := range []struct{ cpanfile, want string }{
		{"requires 'X', dev => 1;", "cpanfile line 1: X: option dev is not supported"},
		{"requires 'X', git => 'https://h/x.git', url => 'https://h/X-1.tar.gz';", "option git cannot be combined"},
		{"requires 'X', ref => 'main';", "option ref needs git"},
		{"requires 'X', mirror => 'https://h/';", "option mirror needs dist"},
		{"requires 'X', url => 'https://h/X-1.tar.gz', dist => 'AB/X-1.tar.gz';", "option url cannot be combined"},
		{"requires 'X', url => 'http://example.com/X-1.tar.gz';", "want an https URL"},
		{"requires 'X', url => 'file:///tmp/X-1.tar.gz';", "want an https URL"},
		{"requires 'X', url => 'https://h/README';", "not a distribution archive name"},
		{"requires 'X', dist => 'AB/X-1.tar.gz', mirror => 'ftp://h/';", "want an https URL"},
		{"requires 'X', dist => 'X-1.tar.gz';", "want AUTHOR/Dist-1.0.tar.gz"},
		{"requires 'X', dist => 'AB/../X-1.tar.gz';", "want AUTHOR/Dist-1.0.tar.gz"},
		{"requires 'X', dist => 'A/AB/AC/X-1.tar.gz';", "want AUTHOR/Dist-1.0.tar.gz"},
		{"requires 'X', git => 'file:///tmp/x.git';", "want an https or ssh URL"},
		{"requires 'X', git => '--upload-pack=x';", "want an https or ssh URL"},
		{"requires 'X', git => 'git://h/x.git';", "want an https or ssh URL"},
		{"requires 'X', git => 'ext::sh';", "want an https or ssh URL"},
		{"requires 'X', git => 'https://h/x.git', ref => \"v1\\n    url: https://h/\";", "option ref \"v1\\n    url: https://h/\": an empty value or white space"},
		{"requires 'X', git => 'https://h/x.git', ref => '';", "option ref \"\": an empty value or white space"},
		{"requires 'C::Only';\nrequires 'C::Only', url => '" + url + "';", "cpanfile lines 1 and 2 name different distributions for C::Only"},
		{"requires 'C::Only', '1.0', url => '" + url + "';", "cpanfile line 1: C::Only 1.0: C-Only-0.10 has version 0.10"},
		{"requires 'Other', url => '" + url + "';", "cpanfile line 1: Other: C-Only-0.10 does not provide it"},
		{"requires 'C::Only';\nconflicts 'C::Only', '< 1.0';", "cpanfile line 2: conflicts C::Only < 1.0, but version 0.10 is selected"},
		{"conflicts 'File::Spec';", "cpanfile line 1: conflicts File::Spec 0, but version"},
	} {
		_, err := c.manager(t).Tidy(context.Background(), project(t, tc.cpanfile+"\n"))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("Tidy(%q) error = %v, want %q", tc.cpanfile, err, tc.want)
		}
	}

	// Options are checked in optional features too, and a conflict with a
	// module nothing selects is no error.
	if _, err := c.manager(t).Tidy(context.Background(), project(t, "feature 'f' => sub { requires 'X', dev => 1 };\n")); err == nil {
		t.Error("an option inside a feature was not checked")
	}
	if _, err := c.manager(t).Tidy(context.Background(), project(t, "requires 'C::Only';\nconflicts 'No::Such::Module';\n")); err != nil {
		t.Error(err)
	}
}

func TestPinnedSourceSyntax(t *testing.T) {
	for _, c := range []struct {
		opts map[string]string
		want Distribution
	}{
		{map[string]string{"dist": "MIYAGAWA/Plack-1.0.tar.gz"}, Distribution{Pathname: "M/MI/MIYAGAWA/Plack-1.0.tar.gz"}},
		{map[string]string{"dist": "M/MI/MIYAGAWA/Plack-1.0.tar.gz"}, Distribution{Pathname: "M/MI/MIYAGAWA/Plack-1.0.tar.gz"}},
		{map[string]string{"dist": "MYCO/Plack-1.0.tar.gz", "mirror": "https://pause.local"}, Distribution{Pathname: "M/MY/MYCO/Plack-1.0.tar.gz", Source: Source{Mirror: "https://pause.local/"}}},
		{map[string]string{"url": "https://h/dl/Plack-1.0.tar.gz?x=1"}, Distribution{Pathname: "Plack-1.0.tar.gz", Source: Source{URL: "https://h/dl/Plack-1.0.tar.gz?x=1"}}},
		{map[string]string{"git": "git@github.com:x/Plack.git", "ref": "main"}, Distribution{Source: Source{Git: "git@github.com:x/Plack.git", Ref: "main"}}},
		{map[string]string{"git": "ssh://git@h/x/Plack"}, Distribution{Source: Source{Git: "ssh://git@h/x/Plack"}}},
		{map[string]string{"git": "git@[::1]:x/Plack.git"}, Distribution{Source: Source{Git: "git@[::1]:x/Plack.git"}}},
		{map[string]string{"git": "http://localhost:8080/Plack.git"}, Distribution{Source: Source{Git: "http://localhost:8080/Plack.git"}}},
	} {
		d, err := pinned(Prereq{Module: "Plack", Options: c.opts})
		if err != nil || d == nil || !reflect.DeepEqual(*d, c.want) {
			t.Errorf("pinned(%v) = %+v, %v, want %+v", c.opts, d, err, c.want)
		}
	}
	for repo, want := range map[string]string{
		"git@github.com:x/Plack.git":  "Plack",
		"https://h/x/p5-Foo+bar.git":  "p5-Foo-bar",
		"ssh://git@h:22/x/Plack/":     "Plack",
		"http://127.0.0.1:1/Fork.git": "Fork",
	} {
		if got, err := repoName(repo); err != nil || got != want {
			t.Errorf("repoName(%s) = %q, %v, want %s", repo, got, err, want)
		}
	}
	if _, err := repoName("https://h/.git"); err == nil {
		t.Error("repoName accepted a URL with no repository name")
	}
}

func TestParseSnapshotSources(t *testing.T) {
	sum, rev := strings.Repeat("ab", 32), strings.Repeat("cd", 20)
	head := "# carton snapshot format: version 1.0\nDISTRIBUTIONS\n  X-1\n    pathname: X-1.tar.gz\n"
	for _, body := range []string{
		"    url: https://h/X-1.tar.gz\n",
		"    url: https://h/X-1.tar.gz\n    sha256: abc\n",
		"    mirror: https://h/\n",
		"    git: https://h/X.git\n",
		"    git: https://h/X.git\n    revision: abc\n",
		"    git: https://h/X.git\n    revision: " + rev + "\n    url: https://h/X-1.tar.gz\n    sha256: " + sum + "\n",
		"    url: https://h/X-1.tar.gz\n    mirror: https://h/\n    sha256: " + sum + "\n",
		"    revision: " + rev + "\n",
		"    ref: main\n",
		"    sha256: " + sum + "\n",
		"    git: file:///tmp/X.git\n    revision: " + rev + "\n",
		"    url: http://example.com/X-1.tar.gz\n    sha256: " + sum + "\n",
	} {
		if _, err := ParseSnapshot([]byte(head + body)); err == nil {
			t.Errorf("ParseSnapshot accepted\n%s", body)
		}
	}
}
