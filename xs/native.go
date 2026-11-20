//go:build darwin || linux

package xs

import (
	"path/filepath"
	"sync"

	"github.com/goccy/go-perl/internal"
)

// A binary built without cgo carries its native modules as files
// (internal.SetXSNative). The dynamic loader needs them on disk, so they
// are unpacked once per user into the cache directory (internal.UnpackTree),
// and every later process loads from there.
var (
	nativeOnce sync.Once
	nativeDir  string
	nativeErr  error
)

// loadNative registers the native modules the binary carries as files.
func loadNative(p *internal.Perl) error {
	tree := internal.XSNative()
	if tree == nil {
		return nil
	}
	nativeOnce.Do(func() { nativeDir, nativeErr = internal.UnpackTree(tree, "native") })
	if nativeErr != nil {
		return nativeErr
	}
	return loadAutoDir(p, filepath.Join(nativeDir, "auto"))
}
