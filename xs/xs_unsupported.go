//go:build !darwin && !linux

package xs

// Native XS loading rides purego's cgo-free dlopen, which exists only on
// the platforms the SDK's toolchain presets target (darwin, linux). On
// every other platform the package still builds — the SDK headers
// (WriteSDK) and the layout helpers stay useful, and loading reports a
// clear error instead of failing the build of everything that links the
// gperl tooling.

import (
	"fmt"
	"os"
	"runtime"

	perl "github.com/goccy/go-perl"
)

// ArchTag names the per-architecture native-module directory
// (local/xs/<tag>), following the running binary.
func ArchTag() string { return runtime.GOOS + "_" + runtime.GOARCH }

// LoadDir keeps the real implementation's contract — a missing dir means
// the project simply has no native modules and is not an error — and
// reports the platform gap only when there actually is something to load.
func LoadDir(p *perl.Perl, dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	return fmt.Errorf("xs: native XS modules are not supported on %s/%s (loading needs dlopen; supported: darwin, linux)",
		runtime.GOOS, runtime.GOARCH)
}
