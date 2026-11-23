//go:build darwin

package xs

// The locale categories as macOS numbers them (locale.h).
const (
	lcCollate  = 1
	lcCtype    = 2
	lcMonetary = 3
	lcNumeric  = 4
	lcTime     = 5
	lcMessages = 6
)

// systemLocaleCategories are the system's categories past the six of
// POSIX: macOS has none.
var systemLocaleCategories []localeCategory
