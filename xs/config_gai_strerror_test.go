//go:build darwin || linux

package xs_test

import "testing"

// The SDK's config.h says the system has gai_strerror (HAS_GAI_STRERROR),
// as %Config's d_gai_strerror does: Socket from CPAN compiles a fallback
// of its own when config.h says it has none, which clashes with the
// declaration in <netdb.h>.
func TestNativeConfigHasGaiStrerror(t *testing.T) {
	buildModuleSource(t, "GaiStrerror", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <netdb.h>
#if defined(HAS_GETADDRINFO) && !defined(HAS_GAI_STRERROR)
static const char *gai_strerror(int err) { return "unknown"; }
#endif
const char *goperl_gai_text(int err) { return gai_strerror(err); }
`)
}
