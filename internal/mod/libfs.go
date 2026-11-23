package mod

import (
	"io/fs"
	"os"
	"path/filepath"
)

// LibFiles merges the module directories of several distributions into one
// tree: slash path in the tree -> the host file it is. When two
// distributions carry the same file, the earlier directory wins. Directories
// that do not exist contribute nothing.
func LibFiles(dirs []string) (map[string]string, error) {
	files := map[string]string{}
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if p == dir && os.IsNotExist(err) {
					return fs.SkipAll
				}
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			if name := filepath.ToSlash(rel); files[name] == "" {
				files[name] = p
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}
