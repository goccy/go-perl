//go:build unix

package perl

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// stdlibTempRoot returns the directory of this user's in the system
// temporary directory (go-perl-<uid>), creating it with mode 0700: where the
// standard library goes when the user's cache directory cannot be used. The
// temporary directory is shared with every other user, so the directory is
// used only when it is this user's alone; see privateDir.
func stdlibTempRoot() (string, error) {
	uid := os.Getuid()
	root := filepath.Join(os.TempDir(), "go-perl-"+strconv.Itoa(uid))
	created := true
	if err := os.Mkdir(root, 0o700); err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return "", err
		}
		created = false
	}
	if err := privateDir(root, uid, created); err != nil {
		return "", err
	}
	return root, nil
}

// privateDir says why dir is not a directory only the user uid can use, or
// nil when it is one: a directory itself (not a symbolic link, which is not
// followed), owned by uid, with mode 0700, so that nobody else can have put
// anything in it. A directory this process has just created may have had
// its mode narrowed by the umask; it is given 0700.
func privateDir(dir string, uid int, created bool) error {
	fi, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	switch {
	case fi.Mode()&fs.ModeSymlink != 0:
		return fmt.Errorf("refusing %s: it is a symbolic link, not a directory of this user's", dir)
	case fi.Mode().Type() != fs.ModeDir:
		return fmt.Errorf("refusing %s: it is not a directory", dir)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("refusing %s: its owner cannot be told", dir)
	}
	if int(st.Uid) != uid {
		return fmt.Errorf("refusing %s: it belongs to user %d, not to this user (%d)", dir, st.Uid, uid)
	}
	if perm := fi.Mode().Perm(); perm != 0o700 {
		if !created {
			return fmt.Errorf("refusing %s: its mode is %#o, not 0700, so others may have changed what is in it", dir, perm)
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return err
		}
	}
	return nil
}
