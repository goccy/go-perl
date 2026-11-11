package perl

// Self-contained Perl standard library.
//
// stdlib.zip is a trimmed copy of the assembled perl5/lib/ tree (no XS static
// archives — those are already linked into perl.wasm — no test suites, no pod)
// embedded into this module so an embedding application does not have to ship
// the lib/ directory alongside its binary. ExtractStdlib unpacks it onto the
// host filesystem — once per user, not once per process — and returns that
// path, suitable for Config.StdlibDir.

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
// read-only. When the user cache directory cannot be used, the library is
// unpacked into a temporary directory for this process instead.
func ExtractStdlib() (string, error) {
	stdlibOnce.Do(func() {
		stdlibPath, stdlibErr = cachedStdlib()
	})
	return stdlibPath, stdlibErr
}

// cachedStdlib resolves the shared copy, unpacking it when this is the first
// process to need it.
func cachedStdlib() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return extractStdlibTo("")
	}
	dir, err := stdlibInCache(filepath.Join(base, "go-perl"))
	if err != nil {
		return extractStdlibTo("")
	}
	return dir, nil
}

// stdlibInCache returns the library's directory under root, creating it when
// absent. The directory only ever exists complete: it is unpacked under a
// temporary name next to its final one and renamed into place, so processes
// racing to create it all end up on one whole copy.
func stdlibInCache(root string) (string, error) {
	sum := sha256.Sum256(stdlibZip)
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

// extractStdlibTo unpacks the embedded zip into a new directory under parent
// (the system temporary directory when empty) and returns it.
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
