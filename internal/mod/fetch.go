package mod

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"cmp"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Fetch makes the distribution d present in the module directory — its
// archive downloaded and verified, its tree extracted — and returns the
// tree's path. Work already done is not repeated. A distribution from
// another mirror, a URL, or a git repository that is not pinned yet gets
// pinned: d records the digest or the commit fetched.
func (m *Manager) Fetch(ctx context.Context, d *Distribution) (string, error) {
	if d.Git == "" && d.URL == "" && d.Mirror == "" || d.Revision != "" || d.SHA256 != "" {
		tree, err := TreePath(m.Dir, d)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(tree); err == nil {
			return tree, nil
		}
	}
	var archive string
	var err error
	switch {
	case d.Git != "":
		archive, err = m.fetchGit(ctx, d)
	case d.URL != "" || d.Mirror != "":
		archive, err = m.fetchArchive(ctx, d)
	default:
		archive, err = m.download(ctx, d.Pathname)
	}
	if err != nil {
		return "", err
	}
	tree, err := TreePath(m.Dir, d)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(tree); err == nil {
		return tree, nil
	}
	if err := extract(archive, tree); err != nil {
		return "", fmt.Errorf("extract %s: %w", d.Pathname, err)
	}
	return tree, nil
}

// download fetches the tarball into the module directory, checking it
// against the author's CHECKSUMS file before it takes its final name.
func (m *Manager) download(ctx context.Context, pathname string) (string, error) {
	dst := TarballPath(m.Dir, pathname)
	if _, err := os.Stat(dst); err == nil {
		return dst, nil
	}
	want, err := m.checksum(ctx, m.Mirror, pathname)
	if err != nil {
		return "", err
	}
	tmp, _, err := m.save(ctx, m.Mirror+"authors/id/"+pathname, filepath.Dir(dst), want, "CHECKSUMS")
	if err != nil {
		return "", err
	}
	if err := os.Rename(tmp, dst); err != nil {
		return "", errors.Join(err, os.Remove(tmp))
	}
	return dst, nil
}

// fetchArchive returns the archive of a distribution from another mirror
// or a URL, kept under its digest, downloading it when the module
// directory lacks it. The digest the snapshot pins must match; without
// one, the mirror's CHECKSUMS (a URL has none) must, and d records the
// digest.
func (m *Manager) fetchArchive(ctx context.Context, d *Distribution) (string, error) {
	dir := filepath.Join(m.Dir, "cache", "download", "sha256")
	base := path.Base(d.Pathname)
	if d.SHA256 != "" {
		if p := filepath.Join(dir, d.SHA256, base); exists(p) {
			return p, nil
		}
	}
	u, want, against := d.URL, d.SHA256, "the snapshot"
	if d.Mirror != "" {
		u = d.Mirror + "authors/id/" + d.Pathname
		if want == "" {
			var err error
			if want, err = m.checksum(ctx, d.Mirror, d.Pathname); err != nil {
				return "", err
			}
			against = "CHECKSUMS"
		}
	}
	tmp, sum, err := m.save(ctx, u, dir, want, against)
	if err != nil {
		return "", err
	}
	dst := filepath.Join(dir, sum, base)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", errors.Join(err, os.Remove(tmp))
	}
	if err := os.Rename(tmp, dst); err != nil {
		return "", errors.Join(err, os.Remove(tmp))
	}
	d.SHA256 = sum
	return dst, nil
}

// save downloads u into a temporary file in dir and returns the file and
// its sha256, which must be want unless want is empty (against names where
// want comes from).
func (m *Manager) save(ctx context.Context, u, dir, want, against string) (_, _ string, err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	tmp, err := os.CreateTemp(dir, path.Base(u)+".tmp-*")
	if err != nil {
		return "", "", err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tmp.Close(), os.Remove(tmp.Name()))
		}
	}()
	body, err := m.open(ctx, u)
	if err != nil {
		return "", "", err
	}
	sum := sha256.New()
	_, err = io.Copy(io.MultiWriter(tmp, sum), body)
	if err = errors.Join(err, body.Close()); err != nil {
		return "", "", fmt.Errorf("download %s: %w", u, err)
	}
	got := hex.EncodeToString(sum.Sum(nil))
	if want != "" && got != want {
		return "", "", fmt.Errorf("%s: sha256 %s does not match %s (%s)", u, got, against, want)
	}
	if err := tmp.Close(); err != nil {
		return "", "", err
	}
	return tmp.Name(), got, nil
}

// fetchGit returns the archive of a distribution from a git repository,
// kept under its commit, cloning the repository when the module directory
// lacks it. Without a pinned revision, d.Ref (the default branch when
// empty) is resolved, and d records the commit and a pathname naming it.
func (m *Manager) fetchGit(ctx context.Context, d *Distribution) (_ string, err error) {
	dir := filepath.Join(m.Dir, "cache", "download", "git")
	if d.Revision != "" {
		if p := filepath.Join(dir, d.Revision+".tar.gz"); exists(p) {
			return p, nil
		}
	}
	name, err := repoName(d.Git)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	clone, err := os.MkdirTemp(dir, name+".clone-*")
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(clone)) }()
	// Plain http is allowed only for a repository on the loopback
	// interface (validateGitURL), so that no redirect leaves https.
	allow := "https:ssh"
	if strings.HasPrefix(d.Git, "http://") {
		allow += ":http"
	}
	if _, err := git(ctx, dir, allow, "clone", "--bare", "--quiet", "--", d.Git, clone); err != nil {
		return "", err
	}
	ref := cmp.Or(d.Revision, d.Ref, "HEAD")
	out, err := git(ctx, dir, allow, "--git-dir="+clone, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("%s: ref %s: %w", d.Git, ref, err)
	}
	rev := strings.TrimSpace(out)
	dst := filepath.Join(dir, rev+".tar.gz")
	tmp := filepath.Join(clone, "archive.tar.gz")
	if _, err := git(ctx, dir, allow, "--git-dir="+clone, "archive", "--format=tar.gz", "--prefix="+name+"/", "-o", tmp, rev); err != nil {
		return "", fmt.Errorf("%s: %w", d.Git, err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		return "", err
	}
	d.Revision = rev
	d.Pathname = name + "-" + rev[:12] + ".tar.gz"
	return dst, nil
}

// git runs git and returns what it prints. git's own credential prompt is
// off (ssh may still ask on the terminal), and only the transports allow
// lists are allowed, whatever git's configuration says.
func git(ctx context.Context, dir, allow string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ALLOW_PROTOCOL="+allow)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// checksumEntry matches one file's record in a CHECKSUMS file and captures
// its sha256.
var checksumEntry = regexp.MustCompile(`'([^']+)'\s*=>\s*\{[^}]*?'sha256'\s*=>\s*'([0-9a-f]{64})'`)

// checksum returns the sha256 the author's CHECKSUMS file records for the
// distribution, or "" when the mirror publishes none for it (a private
// mirror need not keep CHECKSUMS; the download then goes unverified, and
// says so).
func (m *Manager) checksum(ctx context.Context, mirror, pathname string) (string, error) {
	author, file := path.Split(pathname)
	u := mirror + "authors/id/" + author + "CHECKSUMS"
	m.mu.Lock()
	set, ok := m.checksums[u]
	if !ok {
		set = &checksumSet{}
		m.checksums[u] = set
	}
	m.mu.Unlock()
	// One request per author directory, however many of its
	// distributions are being fetched at once.
	set.once.Do(func() {
		body, err := m.get(ctx, u)
		switch {
		case errors.Is(err, errNotFound):
			m.logf("%s is missing; distributions under %s are not verified", u, author)
		case err != nil:
			set.err = err
			return
		}
		set.sums = map[string]string{}
		for _, match := range checksumEntry.FindAllSubmatch(body, -1) {
			set.sums[string(match[1])] = string(match[2])
		}
	})
	return set.sums[file], set.err
}

// checksumSet is one author directory's CHECKSUMS, loaded once.
type checksumSet struct {
	once sync.Once
	sums map[string]string
	err  error
}

// extract unpacks the archive into dst, dropping the single top-level
// directory a distribution archive wraps its files in; a symbolic link that
// leads outside the tree that remains is an error. dst appears only once
// complete, so concurrent processes never see a partial tree.
func extract(archive, dst string) (err error) {
	tmp, err := os.MkdirTemp(filepath.Dir(dst), filepath.Base(dst)+".tmp-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmp)) }()
	if strings.HasSuffix(archive, ".zip") {
		err = extractZip(archive, tmp)
	} else {
		err = ExtractTar(archive, tmp)
	}
	if err != nil {
		return err
	}
	root := tmp
	if entries, err := os.ReadDir(tmp); err != nil {
		return err
	} else if len(entries) == 1 && entries[0].IsDir() {
		root = filepath.Join(tmp, entries[0].Name())
	}
	if err := treeLinksInside(root); err != nil {
		return err
	}
	if err := os.Rename(root, dst); err != nil {
		if _, statErr := os.Stat(dst); statErr == nil {
			return nil // another process extracted it first
		}
		return err
	}
	return nil
}

// member is an archive member's name as a path below the tree; a name that
// leaves the tree is an error.
func member(name string) (string, error) {
	p := filepath.FromSlash(name)
	if !filepath.IsLocal(p) {
		return "", fmt.Errorf("archive member escapes the tree: %s", name)
	}
	return p, nil
}

// writeMember writes the file name of the tree rooted at root, with r's
// content and the modification time modTime, as tar does (a build's make
// compares the times). The root keeps every step inside the tree, symbolic
// links included.
func writeMember(root *os.Root, name string, mode os.FileMode, modTime time.Time, r io.Reader) error {
	p, err := member(name)
	if err != nil {
		return err
	}
	if err := root.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := root.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm()|0o600)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	if err = errors.Join(err, f.Close()); err != nil {
		return err
	}
	return root.Chtimes(p, time.Time{}, modTime)
}

// writeHardLink makes name a hard link to the member target written before
// it, as tar does. A link to itself (tar archives so a file with several
// links that it is given twice) is the file already written.
func writeHardLink(root *os.Root, name, target string) error {
	p, err := member(name)
	if err != nil {
		return err
	}
	old, err := member(target)
	if err != nil {
		return err
	}
	if old == p {
		return nil
	}
	if err := root.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return root.Link(old, p)
}

// writeSymlink makes name a symbolic link to link, which must name a path
// inside the tree as it is written (linksInside checks where the links lead
// once all are in place).
func writeSymlink(root *os.Root, name, link string) error {
	p, err := member(name)
	if err != nil {
		return err
	}
	if linkLeaves(name, link) {
		return fmt.Errorf("archive member %s links outside the tree: %s", name, link)
	}
	if err := root.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return root.Symlink(link, p)
}

// linkLeaves reports whether link, the target of the symbolic link name (a
// slash-separated path below a tree), names a path outside the tree.
func linkLeaves(name, link string) bool {
	return path.IsAbs(link) || !filepath.IsLocal(filepath.FromSlash(path.Join(path.Dir(name), link)))
}

// treeLinksInside fails for a symbolic link of the tree at dir whose target
// names a path outside the tree, or that linksInside refuses.
func treeLinksInside(dir string) (err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	var links []string
	err = fs.WalkDir(root.FS(), ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.Type()&fs.ModeSymlink == 0 {
			return err
		}
		link, err := root.Readlink(filepath.FromSlash(name))
		if err != nil {
			return err
		}
		if linkLeaves(name, link) {
			return fmt.Errorf("archive member %s links outside the tree: %s", name, link)
		}
		links = append(links, name)
		return nil
	})
	if err != nil {
		return err
	}
	return linksInside(root, links)
}

// linksInside fails for a symbolic link of the tree at root that cannot be
// followed inside the tree: one that leads outside it through other links,
// a loop of links, or a link below a file. A link to nothing is no error.
func linksInside(root *os.Root, links []string) error {
	for _, name := range links {
		if _, err := root.Stat(filepath.FromSlash(name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("archive member %s: a link that cannot be followed inside the tree: %w", name, err)
		}
	}
	return nil
}

// ExtractTar unpacks the tar archive at archive (gzip-compressed, or bzip2
// for a name ending in .bz2) into the directory dir. A member, or the
// target of a symbolic link, that leaves dir is an error.
func ExtractTar(archive, dir string) (err error) {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	var r io.Reader
	if strings.HasSuffix(archive, ".bz2") {
		r = bzip2.NewReader(f)
	} else if r, err = gzip.NewReader(f); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	tr := tar.NewReader(r)
	var links []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return linksInside(root, links)
		}
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			var p string
			if p, err = member(hdr.Name); err == nil {
				err = root.MkdirAll(p, 0o755)
			}
		case tar.TypeReg:
			err = writeMember(root, hdr.Name, hdr.FileInfo().Mode(), hdr.ModTime, tr)
		case tar.TypeSymlink:
			err = writeSymlink(root, hdr.Name, hdr.Linkname)
			links = append(links, hdr.Name)
		case tar.TypeLink:
			// A hard link to a symbolic link is that link again, so it is
			// checked as one.
			err = writeHardLink(root, hdr.Name, hdr.Linkname)
			links = append(links, hdr.Name)
		}
		if err != nil {
			return err
		}
	}
}

func extractZip(archive, dir string) (err error) {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, zr.Close()) }()
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			p, err := member(zf.Name)
			if err != nil {
				return err
			}
			if err := root.MkdirAll(p, 0o755); err != nil {
				return err
			}
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		if err := errors.Join(writeMember(root, zf.Name, zf.Mode(), zf.Modified, rc), rc.Close()); err != nil {
			return err
		}
	}
	return nil
}
