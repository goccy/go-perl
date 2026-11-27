package internal

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// InCache runs place on the directory of go-perl in the user's cache
// directory, and when that cannot be used (no cache directory, or place
// fails there), on this user's directory in the system temporary directory
// (TempRoot). what names the content in the error of both failing.
func InCache(what string, place func(root string) (string, error)) (string, error) {
	base, err := os.UserCacheDir()
	if err == nil {
		var dir string
		if dir, err = place(filepath.Join(base, "go-perl")); err == nil {
			return dir, nil
		}
	}
	root, terr := TempRoot()
	if terr != nil {
		return "", fmt.Errorf("no directory for %s: the user cache directory: %v; the temporary directory: %w", what, err, terr)
	}
	return place(root)
}

// UnpackTree writes tree into the cache (InCache), under a name made of
// prefix and a digest of the tree's content, and returns that directory.
// A directory of that name already there is the same content and is used as
// it is; one is only ever there complete, since the tree is written beside
// it and renamed into place.
func UnpackTree(tree fs.FS, prefix string) (string, error) {
	sum, err := treeDigest(tree)
	if err != nil {
		return "", err
	}
	name := prefix + "-" + hex.EncodeToString(sum[:8])
	return InCache(prefix, func(root string) (string, error) {
		dir := filepath.Join(root, name)
		if _, err := os.Stat(dir); err == nil {
			return dir, nil
		}
		if err := os.MkdirAll(root, 0o755); err != nil {
			return "", err
		}
		tmp, err := os.MkdirTemp(root, prefix+"-tmp-")
		if err != nil {
			return "", err
		}
		if err := os.CopyFS(tmp, tree); err != nil {
			return "", errors.Join(err, os.RemoveAll(tmp))
		}
		if err := os.Rename(tmp, dir); err != nil {
			if _, statErr := os.Stat(dir); statErr != nil {
				return "", errors.Join(err, os.RemoveAll(tmp))
			}
			// Another process put the same tree in place first.
			return dir, os.RemoveAll(tmp)
		}
		return dir, nil
	})
}

// treeDigest is a SHA-256 of the names, modes and contents of tree's files,
// each field preceded by its length.
func treeDigest(tree fs.FS) ([sha256.Size]byte, error) {
	h := sha256.New()
	var n [8]byte
	field := func(b []byte) {
		binary.LittleEndian.PutUint64(n[:], uint64(len(b)))
		h.Write(n[:])
		h.Write(b)
	}
	err := fs.WalkDir(tree, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := fs.ReadFile(tree, name)
		if err != nil {
			return err
		}
		field([]byte(name))
		field(binary.LittleEndian.AppendUint32(nil, uint32(info.Mode().Perm())))
		field(data)
		return nil
	})
	var sum [sha256.Size]byte
	h.Sum(sum[:0])
	return sum, err
}
