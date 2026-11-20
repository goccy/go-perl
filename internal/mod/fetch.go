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
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
// directory a distribution archive wraps its files in. dst appears only
// once complete, so concurrent processes never see a partial tree.
func extract(archive, dst string) (err error) {
	tmp, err := os.MkdirTemp(filepath.Dir(dst), filepath.Base(dst)+".tmp-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmp)) }()
	if strings.HasSuffix(archive, ".zip") {
		err = extractZip(archive, tmp)
	} else {
		err = extractTar(archive, tmp)
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
	if err := os.Rename(root, dst); err != nil {
		if _, statErr := os.Stat(dst); statErr == nil {
			return nil // another process extracted it first
		}
		return err
	}
	return nil
}

// target resolves an archive member name below root, refusing names that
// escape it.
func target(root, name string) (string, error) {
	p := filepath.Join(root, filepath.FromSlash(name))
	if p != root && !strings.HasPrefix(p, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive member escapes the tree: %s", name)
	}
	return p, nil
}

func writeMember(p string, mode os.FileMode, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm()|0o600)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return errors.Join(err, f.Close())
}

func extractTar(archive, root string) (err error) {
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
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		p, err := target(root, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(p, 0o755)
		case tar.TypeReg:
			err = writeMember(p, hdr.FileInfo().Mode(), tr)
		case tar.TypeSymlink:
			// A link may only point inside the tree.
			if _, err = target(root, path.Join(path.Dir(hdr.Name), hdr.Linkname)); err == nil {
				if err = os.MkdirAll(filepath.Dir(p), 0o755); err == nil {
					err = os.Symlink(hdr.Linkname, p)
				}
			}
		}
		if err != nil {
			return err
		}
	}
}

func extractZip(archive, root string) (err error) {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, zr.Close()) }()
	for _, zf := range zr.File {
		p, err := target(root, zf.Name)
		if err != nil {
			return err
		}
		if zf.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		if err := errors.Join(writeMember(p, zf.Mode(), rc), rc.Close()); err != nil {
			return err
		}
	}
	return nil
}
