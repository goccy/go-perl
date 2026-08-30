package xs

// The SDK headers ship inside the library so build tooling (gperl xs
// build) is self-contained: no checkout of go-perl is needed on the
// machine that compiles an XS distribution.

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed sdk/include
var sdkFS embed.FS

// WriteSDK materializes the native XS SDK headers into dir (creating it),
// so a C compiler can be pointed at them with -I dir. Existing files are
// overwritten: the headers must match this library version.
func WriteSDK(dir string) error {
	sub, err := fs.Sub(sdkFS, "sdk/include")
	if err != nil {
		return err
	}
	return fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, path)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		data, err := fs.ReadFile(sub, path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
}
