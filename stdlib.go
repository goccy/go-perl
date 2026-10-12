package perl

// Self-contained Perl standard library.
//
// stdlib.zip is a trimmed copy of the assembled perl5/lib/ tree (no XS static
// archives — those are already linked into perl.wasm — no test suites, no pod)
// embedded into this module so an embedding application does not have to ship
// the lib/ directory alongside its binary. The library default serves it from
// memory, mounted read-only at StdlibDir (WithStdlib, what a nil Config.FS
// gets); ExtractStdlib is the host-filesystem alternative: it unpacks the tree
// once per process into a temp directory and returns that path, suitable for
// Config.StdlibDir when the FS is a host backend (fs.NewHostFS).

import (
	"archive/zip"
	"bytes"
	_ "embed"
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

// ExtractStdlib unpacks the embedded standard library into a temporary
// directory (once per process) and returns its path. The result is cached, so
// repeated calls — e.g. one per Interpreter — are cheap. The returned directory
// is the value to pass as Config.StdlibDir (it becomes @INC via perl_new).
func ExtractStdlib() (string, error) {
	stdlibOnce.Do(func() {
		stdlibPath, stdlibErr = extractStdlibTo("")
	})
	return stdlibPath, stdlibErr
}

// extractStdlibTo unpacks the embedded zip under parent (os.MkdirTemp default
// when empty). Exposed separately so tests can target an explicit location.
func extractStdlibTo(parent string) (string, error) {
	dir, err := os.MkdirTemp(parent, "go-perl-stdlib-")
	if err != nil {
		return "", err
	}
	zr, err := zip.NewReader(bytes.NewReader(stdlibZip), int64(len(stdlibZip)))
	if err != nil {
		return "", fmt.Errorf("open embedded stdlib: %w", err)
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
			return "", fmt.Errorf("illegal path in stdlib zip: %q", f.Name)
		}
		target := filepath.Join(dir, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(dir)+string(os.PathSeparator)) {
			return "", fmt.Errorf("illegal path in stdlib zip: %q", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", err
		}
		if err := extractStdlibFile(f, target); err != nil {
			return "", err
		}
	}
	return dir, nil
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
