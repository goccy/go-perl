//go:build linux

package xs

// The locale categories as glibc numbers them (locale.h).
const (
	lcCtype    = 0
	lcNumeric  = 1
	lcTime     = 2
	lcCollate  = 3
	lcMonetary = 4
	lcMessages = 5
)

// systemLocaleCategories are glibc's categories past the six of POSIX,
// which perl sets from the environment too.
var systemLocaleCategories = []localeCategory{
	{7, "LC_PAPER"}, {8, "LC_NAME"}, {9, "LC_ADDRESS"}, {10, "LC_TELEPHONE"},
	{11, "LC_MEASUREMENT"}, {12, "LC_IDENTIFICATION"},
}
