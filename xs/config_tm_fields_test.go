//go:build darwin || linux

package xs_test

import "testing"

// The SDK's config.h says struct tm has tm_zone and tm_gmtoff
// (HAS_TM_TM_ZONE, HAS_TM_TM_GMTOFF), as the C libraries of Linux and
// darwin have them and %Config says: native code that formats a time with
// %z takes the offset from tm_gmtoff, which it fills with mktime only when
// config.h says the field is there (POSIX::strftime::GNU's XS printed the
// offset of the time it ran at for a summer date).
func TestNativeConfigHasTheZoneFieldsOfTm(t *testing.T) {
	buildModuleSource(t, "TmFields", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <time.h>
#if !defined(HAS_TM_TM_ZONE) || !defined(HAS_TM_TM_GMTOFF)
#error "config.h does not say struct tm has tm_zone and tm_gmtoff"
#endif
long goperl_tm_offset(struct tm *tm) { return tm->tm_gmtoff; }
const char *goperl_tm_zone(struct tm *tm) { return tm->tm_zone; }
`)
}
