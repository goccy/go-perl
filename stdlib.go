package perl

// Self-contained Perl standard library.
//
// stdlib.zip is a trimmed copy of the assembled perl5/lib/ tree (no XS static
// archives — those are already linked into perl.wasm — no test suites, no pod)
// embedded into this module so an embedding application does not have to ship
// the lib/ directory alongside its binary. The library default serves it from
// memory, mounted read-only at StdlibDir (WithStdlib, what a nil Config.FS
// gets); ExtractStdlib is the host-filesystem alternative: it unpacks the tree
// onto the host filesystem — once per user, not once per process — and
// returns that path, suitable for Config.StdlibDir when the FS is a host
// backend (fs.NewHostFS).

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goccy/go-perl/internal"
)

//go:embed stdlib.zip
var stdlibZip []byte

var (
	stdlibOnce sync.Once
	stdlibPath string
	stdlibErr  error
)

// ExtractStdlib returns a host directory holding the embedded standard
// library, the value to pass as Config.StdlibDir for an instance on the host
// filesystem (it becomes @INC via perl_new).
//
// The library is unpacked once into the user's cache directory, under a name
// derived from its content, and every later process — and every later call —
// reuses that directory. A different go-perl version carries a different
// library and gets its own directory. The directory is shared: treat it as
// read-only. When the user cache directory cannot be used (HOME is not set,
// say), the library goes the same way into a directory of the user's own in
// the system temporary directory, which is used only when it is the user's
// alone (see internal.TempRoot); otherwise ExtractStdlib fails and says why.
func ExtractStdlib() (string, error) {
	stdlibOnce.Do(func() {
		stdlibPath, stdlibErr = cachedStdlib()
	})
	return stdlibPath, stdlibErr
}

// cachedStdlib resolves the shared copy, unpacking it when this is the first
// process to need it.
func cachedStdlib() (string, error) {
	return internal.InCache("the standard library", stdlibInCache)
}

// stdlibInCache returns the library's directory under root, creating it when
// absent. The directory only ever exists complete: it is unpacked under a
// temporary name next to its final one and renamed into place, so processes
// racing to create it all end up on one whole copy.
func stdlibInCache(root string) (string, error) {
	sum, err := stdlibDigest(stdlibZip)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "stdlib-"+hex.EncodeToString(sum[:16]))
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	tmp, err := extractStdlibTo(root)
	if err != nil {
		return "", err
	}
	if err := os.Rename(tmp, dir); err != nil {
		// Another process put its copy in place first; use that one.
		os.RemoveAll(tmp)
		if _, serr := os.Stat(dir); serr != nil {
			return "", errors.Join(err, serr)
		}
	}
	return dir, nil
}

// stdlibDigest names the content of the library archive lib: a SHA-256 of
// what the archive's central directory records of each entry, in its order
// (its name, whether it is a directory, its uncompressed size and the
// CRC-32 of its content, the check the zip format itself relies on), which
// is what an unpacked copy is made of. It reads the directory alone, not
// the archive's ten megabytes: every process start names the directory,
// and a SHA-256 of the whole archive was a third of what starting an
// interpreter cost.
func stdlibDigest(lib []byte) ([sha256.Size]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(lib), int64(len(lib)))
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("open embedded stdlib: %w", err)
	}
	h := sha256.New()
	var rec [13]byte
	for _, f := range zr.File {
		binary.LittleEndian.PutUint64(rec[0:8], uint64(len(f.Name)))
		h.Write(rec[:8])
		io.WriteString(h, f.Name)
		rec[0] = 0
		if f.FileInfo().IsDir() {
			rec[0] = 1
		}
		binary.LittleEndian.PutUint64(rec[1:9], f.UncompressedSize64)
		binary.LittleEndian.PutUint32(rec[9:13], f.CRC32)
		h.Write(rec[:])
	}
	var sum [sha256.Size]byte
	h.Sum(sum[:0])
	return sum, nil
}

// extractStdlibTo unpacks the embedded zip into a new directory under parent
// and returns it.
func extractStdlibTo(parent string) (string, error) {
	dir, err := os.MkdirTemp(parent, "go-perl-stdlib-")
	if err != nil {
		return "", err
	}
	if err := unpackStdlib(dir); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

func unpackStdlib(dir string) error {
	zr, err := zip.NewReader(bytes.NewReader(stdlibZip), int64(len(stdlibZip)))
	if err != nil {
		return fmt.Errorf("open embedded stdlib: %w", err)
	}
	for _, f := range zr.File {
		if filepath.Clean(f.Name) == "." {
			continue
		}
		// Guard against zip-slip. Entry names are external text (the zip
		// format defines them as such), so the name itself is checked: no
		// parent-directory segments, no absolute paths — the stdlib zip we
		// build never contains either. The joined target is then re-checked
		// to stay under dir (defense in depth).
		if strings.Contains(f.Name, "..") || strings.HasPrefix(f.Name, "/") {
			return fmt.Errorf("illegal path in stdlib zip: %q", f.Name)
		}
		target := filepath.Join(dir, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(dir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal path in stdlib zip: %q", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := extractStdlibFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func extractStdlibFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, rc); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
