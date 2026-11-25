package mod

import (
	"archive/tar"
	"archive/zip"
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
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Fetch makes the distribution at pathname present in the module directory
// — its tarball downloaded and verified, its tree extracted — and returns
// the tree's path. Work already done is not repeated.
func (m *Manager) Fetch(ctx context.Context, pathname string) (string, error) {
	tree, err := TreePath(m.Dir, pathname)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(tree); err == nil {
		return tree, nil
	}
	tarball, err := m.download(ctx, pathname)
	if err != nil {
		return "", err
	}
	if err := extract(tarball, tree); err != nil {
		return "", fmt.Errorf("extract %s: %w", pathname, err)
	}
	return tree, nil
}

// download fetches the tarball into the module directory, checking it
// against the author's CHECKSUMS file before it takes its final name.
func (m *Manager) download(ctx context.Context, pathname string) (_ string, err error) {
	dst := TarballPath(m.Dir, pathname)
	if _, err := os.Stat(dst); err == nil {
		return dst, nil
	}
	want, err := m.checksum(ctx, pathname)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), filepath.Base(dst)+".tmp-*")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tmp.Close(), os.Remove(tmp.Name()))
		}
	}()
	u := m.Mirror + "authors/id/" + pathname
	body, err := m.open(ctx, u)
	if err != nil {
		return "", err
	}
	sum := sha256.New()
	_, err = io.Copy(io.MultiWriter(tmp, sum), body)
	if err = errors.Join(err, body.Close()); err != nil {
		return "", fmt.Errorf("download %s: %w", u, err)
	}
	if got := hex.EncodeToString(sum.Sum(nil)); want != "" && got != want {
		return "", fmt.Errorf("%s: sha256 %s does not match CHECKSUMS (%s)", u, got, want)
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	return dst, os.Rename(tmp.Name(), dst)
}

// checksumEntry matches one file's record in a CHECKSUMS file and captures
// its sha256.
var checksumEntry = regexp.MustCompile(`'([^']+)'\s*=>\s*\{[^}]*?'sha256'\s*=>\s*'([0-9a-f]{64})'`)

// checksum returns the sha256 the author's CHECKSUMS file records for the
// distribution, or "" when the mirror publishes none for it (a private
// mirror need not keep CHECKSUMS; the download then goes unverified, and
// says so).
func (m *Manager) checksum(ctx context.Context, pathname string) (string, error) {
	author, file := path.Split(pathname)
	m.mu.Lock()
	set, ok := m.checksums[author]
	if !ok {
		set = &checksumSet{}
		m.checksums[author] = set
	}
	m.mu.Unlock()
	// One request per author directory, however many of its
	// distributions are being fetched at once.
	set.once.Do(func() {
		u := m.Mirror + "authors/id/" + author + "CHECKSUMS"
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
