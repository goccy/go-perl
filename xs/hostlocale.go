//go:build darwin || linux

package xs

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"github.com/goccy/go-perl/internal/ccall"
)

// The C library's locale for native code. perl sets the locale of its
// environment when it starts (init_i18nl10n) and keeps LC_NUMERIC "C" for
// itself, so the native code it loads runs in the user's locale: GLib
// converts the arguments it parses from the locale's character set,
// Text::CharWidth measures characters with mbtowc and wcwidth. The
// interpreter of gperl does that in its own C library; the C library of
// this process, which native code calls, would stay in "C". So before the
// first native module, the locale is set here as perl's start-up sets it:
// the locale the environment names for each category (LC_ALL, then the
// category's own variable, then LANG); when the system has no data for one
// of them, the locale LC_ALL names, then the one LANG names, for every
// category; else "C". The locale is the process's, so it is set once, from
// the environment of the first interpreter that loads native code.
var hostLocaleOnce sync.Once

func setHostLocale(env []string) {
	hostLocaleOnce.Do(func() {
		fn, err := libcSymbol("setlocale")
		if err != nil {
			// Loading native code goes on; it runs in the "C" locale.
			fmt.Fprintf(os.Stderr, "gperl: native code's locale not set: %v\n", err)
			return
		}
		set := func(category int, name string) bool {
			b := append([]byte(name), 0)
			r := ccall.Invoke(fn, uintptr(category), uintptr(unsafe.Pointer(&b[0])))
			runtime.KeepAlive(b)
			return r != 0
		}
		// every category, each in the locale names gives it; false when
		// the system has no data for one of them
		setAll := func(names func(variable string) string) bool {
			for _, c := range append(localeCategories, systemLocaleCategories...) {
				if !set(c.category, names(c.variable)) {
					return false
				}
			}
			return true
		}
		ok := setAll(func(variable string) string { return localeOfEnvironment(env, variable) })
		for _, fallback := range []string{envValue(env, "LC_ALL"), envValue(env, "LANG")} {
			if ok || fallback == "" {
				continue
			}
			ok = setAll(func(string) string { return fallback })
		}
		if !ok {
			setAll(func(string) string { return "C" })
		}
		set(lcNumeric, "C")
	})
}

// localeCategory is a category of the C library's locale and the
// environment variable that names its locale.
type localeCategory struct {
	category int
	variable string
}

// localeCategories are the six categories of POSIX.
var localeCategories = []localeCategory{
	{lcCtype, "LC_CTYPE"}, {lcNumeric, "LC_NUMERIC"}, {lcTime, "LC_TIME"},
	{lcCollate, "LC_COLLATE"}, {lcMonetary, "LC_MONETARY"}, {lcMessages, "LC_MESSAGES"},
}

// localeOfEnvironment is the locale env names for the category whose
// variable is given: LC_ALL, then the variable, then LANG, else "C".
func localeOfEnvironment(env []string, variable string) string {
	for _, name := range []string{"LC_ALL", variable, "LANG"} {
		if v := envValue(env, name); v != "" {
			return v
		}
	}
	return "C"
}

// envValue is the value of the variable in env, the first one of a name
// given twice, as getenv finds.
func envValue(env []string, name string) string {
	for _, kv := range env {
		if k, val, ok := strings.Cut(kv, "="); ok && k == name {
			return val
		}
	}
	return ""
}
