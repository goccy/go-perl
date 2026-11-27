//go:build darwin

package xs

// The host's CLOCK_MONOTONIC, read through the C library: the Go runtime
// reads mach_absolute_time, which stands still while the system sleeps.

import (
	"github.com/ebitengine/purego"
	"github.com/goccy/go-perl/internal/ccall"

	"github.com/goccy/go-perl/internal"
)

// clockMonotonic is darwin's CLOCK_MONOTONIC.
const clockMonotonic = 6

func init() {
	// clock_gettime_nsec_np is in the C library of every macOS the Go
	// toolchain supports; without it the program reads the runtime's clock.
	fn, err := purego.Dlsym(purego.RTLD_DEFAULT, "clock_gettime_nsec_np")
	if err != nil {
		return
	}
	internal.RegisterHostMonotonicClock(func() uint64 { return uint64(ccall.Invoke(fn, clockMonotonic)) })
}
