package gperl

import (
	"os"
	"path/filepath"
)

// RemoveSelfPerl removes the name the test binary has as the perl command:
// a test binary is gone after its run, and its name would point nowhere.
func RemoveSelfPerl() {
	link, _, _, ok := selfPerlPath()
	if !ok {
		return
	}
	if _, err := os.Lstat(link); err == nil {
		os.RemoveAll(filepath.Dir(link))
	}
}
