package gperl

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// The running executable as the perl command.
//
// A perl knows the path it can be run by: $^X, and %Config's perlpath and
// startperl, which is what an installation writes on the #! line of the
// programs it installs. The executable that runs the interpreter here is
// not a perl command by its own command line (gperl wants `run` first), so
// it has a second name that is one: a link called perl, in a directory of
// its own in the user's cache, that points at the executable. Started by
// that name the executable takes its whole command line as perl's (see
// reexec.go). The place of the link follows from the place of the
// executable alone, so a link that was removed is the same link again after
// the next run.

// selfPerlPath returns where the link to this executable is (whether or not
// it exists) and the executable it points at; ok is false when the process
// cannot name its executable or the user has no cache directory.
func selfPerlPath() (link, exe string, ok bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", "", false
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return "", "", false
	}
	cache, err := os.UserCacheDir()
	if err != nil || !filepath.IsAbs(cache) {
		return "", "", false
	}
	sum := sha256.Sum256([]byte(exe))
	return filepath.Join(cache, "go-perl", "perl-"+hex.EncodeToString(sum[:8]), "perl"), exe, true
}

// selfPerl returns the path that runs this executable as the perl command,
// making the link when it is not there; "" when there is no such path on
// this system.
func selfPerl() string {
	link, exe, ok := selfPerlPath()
	if !ok {
		return ""
	}
	if target, err := os.Readlink(link); err == nil && target == exe {
		return link
	}
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return ""
	}
	// Whatever is at the name is not the link: a new one is made beside it
	// and renamed over it, so that the name runs an executable at every
	// moment, also while several processes do this at once.
	tmp, err := os.MkdirTemp(filepath.Dir(link), "new-")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(tmp)
	fresh := filepath.Join(tmp, "perl")
	if err := os.Symlink(exe, fresh); err != nil {
		return ""
	}
	if err := os.Rename(fresh, link); err != nil {
		return ""
	}
	return link
}

// startedAsPerl reports whether the process was started by the name
// selfPerl gives the executable: the name the system passed as the
// command's own (the path on a #! line, or the path a program ran).
func startedAsPerl() bool {
	if len(os.Args) == 0 || !filepath.IsAbs(os.Args[0]) || filepath.Base(os.Args[0]) != "perl" {
		return false
	}
	link, _, ok := selfPerlPath()
	return ok && filepath.Clean(os.Args[0]) == link
}
