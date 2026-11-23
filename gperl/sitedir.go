package gperl

import (
	"os"
	"path/filepath"
	"runtime"
)

// siteDir returns the directory the site directories of the perl command
// are under: where `make install` and `./Build install` put the modules a
// user installs, and where every later run finds them. It is a directory of
// the user's own, under the place the system keeps a user's application
// data, so installing needs no privilege and does not touch a perl of the
// system:
//
//	Linux and other unix:  $XDG_DATA_HOME/go-perl, or ~/.local/share/go-perl
//	macOS:                 ~/Library/Application Support/go-perl
//	Windows:               %LocalAppData%\go-perl
//
// Below it the interpreter lays out site_perl/<version> (the modules), bin
// (their programs) and man. Nothing is created here: the directories come
// into being with the first installation. The result is empty when the user
// has no such place (no home directory); the program then runs without site
// directories.
func siteDir() string {
	base := userDataDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "go-perl")
}

// userDataDir returns the directory under which the system keeps a user's
// application data, or "" when there is none to name.
func userDataDir() string {
	switch runtime.GOOS {
	case "windows":
		return absoluteDir(os.Getenv("LocalAppData"))
	case "darwin", "ios":
		home := homeDir()
		if home == "" {
			return ""
		}
		return filepath.Join(home, "Library", "Application Support")
	default:
		// The XDG Base Directory Specification: a relative value is
		// invalid and to be ignored.
		if dir := absoluteDir(os.Getenv("XDG_DATA_HOME")); dir != "" {
			return dir
		}
		home := homeDir()
		if home == "" {
			return ""
		}
		return filepath.Join(home, ".local", "share")
	}
}

// homeDir returns the user's home directory when it is an absolute path.
func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return absoluteDir(home)
}

// absoluteDir returns dir when it is an absolute path, and "" otherwise.
func absoluteDir(dir string) string {
	if !filepath.IsAbs(dir) {
		return ""
	}
	return filepath.Clean(dir)
}
