package main

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// maxRecordBytes bounds what is copied out of one container: its phase logs
// are cut at 256 MiB each, so more than this is not records.
const maxRecordBytes = 4 << 30

// copyOut copies the records of the ended container name (its /out) into
// the directory out, which must not exist. The container wrote them and
// may have been malicious; they are read on this machine. So docker cp
// hands them over as a tar stream, and only its regular files and
// directories are made (extractRecords): no link, fifo or device it
// planted ever exists here. out exists afterwards even when the copy
// failed.
func copyOut(name, out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("docker", "cp", name+":/out", "-")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	stream, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	extractErr := extractRecords(stream, out)
	// The rest of the stream, so that docker cp can end.
	io.Copy(io.Discard, stream)
	waitErr := cmd.Wait()
	// Whatever went wrong, nothing but files and directories stays.
	if err := removeSpecialFiles(out); err != nil {
		return err
	}
	if waitErr != nil {
		return fmt.Errorf("docker cp: %v: %s", waitErr, stderr.String())
	}
	return extractErr
}

// extractRecords makes, below dir, the regular files and directories of the
// tar stream r, whose entries are all below one top directory (docker cp
// names it after the directory copied), and skips every other kind of
// entry: symbolic and hard links, fifos, devices. An entry whose name
// leads out of the top directory is skipped too. Files are made new, with
// no mode bits but read and write for the owner and read for others, and
// never through an existing entry.
func extractRecords(r io.Reader, dir string) error {
	tr := tar.NewReader(r)
	var total int64
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		name := filepath.ToSlash(filepath.Clean(h.Name))
		_, rel, found := strings.Cut(name, "/")
		if !filepath.IsLocal(name) || !found || !filepath.IsLocal(rel) {
			// The top directory itself, or a name that leaves it.
			continue
		}
		path := filepath.Join(dir, filepath.FromSlash(rel))
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if total += h.Size; total > maxRecordBytes {
				return fmt.Errorf("the records pass %d bytes", int64(maxRecordBytes))
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0o644)
			if err != nil {
				return err
			}
			_, err = io.Copy(f, tr)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
			if err != nil {
				return err
			}
		}
	}
}

// removeSpecialFiles deletes every entry below dir that is neither a
// regular file nor a directory: symbolic links, fifos, sockets, devices.
// Nothing is followed: WalkDir reads the entries' own types.
func removeSpecialFiles(dir string) error {
	return filepath.WalkDir(dir, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if t := e.Type(); !t.IsRegular() && !t.IsDir() {
			return os.Remove(path)
		}
		return nil
	})
}
