//go:build !unix

package perl

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// stdlibTempRoot returns the directory go-perl in the system temporary
// directory, creating it: where the standard library goes when the user's
// cache directory cannot be used. On these systems the temporary directory
// is the user's own (Windows keeps one in each user's profile; WASI has the
// directories the host grants), and there is no owner to compare; the
// directory must still be a directory itself, not a link to one, which is
// not followed.
func stdlibTempRoot() (string, error) {
	root := filepath.Join(os.TempDir(), "go-perl")
	if err := os.Mkdir(root, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
		return "", err
	}
	fi, err := os.Lstat(root)
	if err != nil {
		return "", err
	}
	switch {
	case fi.Mode()&fs.ModeSymlink != 0:
		return "", fmt.Errorf("refusing %s: it is a symbolic link, not a directory of this user's", root)
	case fi.Mode().Type() != fs.ModeDir:
		return "", fmt.Errorf("refusing %s: it is not a directory", root)
	}
	return root, nil
}
