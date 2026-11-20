package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// UnpackTree writes tree into the user's cache directory, under a name made
// of prefix and a digest of the tree's content, and returns that directory.
// A directory of that name already there is the same content and is used as
// it is; one is only ever there complete, since the tree is written beside
// it and renamed into place.
func UnpackTree(tree fs.FS, prefix string) (string, error) {
	files := map[string][]byte{}
	sum := sha256.New()
	err := fs.WalkDir(tree, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(tree, name)
		if err != nil {
			return err
		}
		files[name] = data
		sum.Write([]byte(name))
		sum.Write([]byte{0})
		sum.Write(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	parent := filepath.Join(cache, "go-perl")
	dir := filepath.Join(parent, prefix+"-"+hex.EncodeToString(sum.Sum(nil)[:8]))
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(parent, prefix+"-tmp-")
	if err != nil {
		return "", err
	}
	for name, data := range files {
		dst := filepath.Join(tmp, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return "", errors.Join(err, os.RemoveAll(tmp))
		}
		if err := os.WriteFile(dst, data, 0o755); err != nil {
			return "", errors.Join(err, os.RemoveAll(tmp))
		}
	}
	if err := os.Rename(tmp, dir); err != nil {
		if _, statErr := os.Stat(dir); statErr != nil {
			return "", errors.Join(err, os.RemoveAll(tmp))
		}
		// Another process put the same tree in place first.
		return dir, os.RemoveAll(tmp)
	}
	return dir, nil
}
