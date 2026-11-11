package main

// The native step: which distributions carry native (XS) code, and how
// much of CPAN stands on each of them.
//
// A distribution with XS code runs on go-perl only if it builds against
// the native XS SDK, so the ones many others depend on decide how far a
// survey can get. Whether a distribution has XS code is a fact about its
// archive — the names of the files in it — so the archives of the most
// depended-upon distributions are downloaded (and kept, the later steps
// need them too) and their file lists read. Nothing is unpacked.

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// nativeDist is one distribution's record in native.jsonl.
type nativeDist struct {
	Dist            string `json:"dist"`
	Release         string `json:"release"`
	DependentsTotal int    `json:"dependents_total"`
	// Bundled is set when the interpreter's standard library already
	// ships every module of the distribution at this version or newer, so
	// a dependency on it installs nothing.
	Bundled bool `json:"bundled,omitempty"`
	// XS lists the archive's .xs files, relative to the distribution
	// root, outside its test and bundled-installer directories.
	XS []string `json:"xs,omitempty"`
	// C counts the archive's other C and C++ sources and headers.
	C int `json:"c,omitempty"`
	// Error is set when the archive could not be fetched or read.
	Error string `json:"error,omitempty"`
}

const cpanMirror = "https://cpan.metacpan.org/authors/id/"

// scanNative writes native.jsonl for the top distributions by the number
// of distributions that need them, directly or indirectly.
func scanNative(dataDir string, top int) error {
	dists, err := readDists(filepath.Join(dataDir, "dists.jsonl"))
	if err != nil {
		return err
	}
	provided, err := readProvided(filepath.Join(dataDir, "provided.json"))
	if err != nil {
		return err
	}
	sort.SliceStable(dists, func(i, j int) bool {
		if dists[i].DependentsTotal != dists[j].DependentsTotal {
			return dists[i].DependentsTotal > dists[j].DependentsTotal
		}
		return dists[i].Dist < dists[j].Dist
	})
	if top > len(dists) {
		top = len(dists)
	}
	dists = dists[:top]

	rows := make([]nativeDist, len(dists))
	var wg sync.WaitGroup
	slots := make(chan struct{}, 8)
	for i := range dists {
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-slots }()
			d := &dists[i]
			row := nativeDist{Dist: d.Dist, Release: d.Release, DependentsTotal: d.DependentsTotal}
			row.Bundled = bundled(d, provided)
			archive, err := fetchArchive(dataDir, d)
			if err == nil {
				row.XS, row.C, err = nativeSources(d.Path, archive)
			}
			if err != nil {
				row.Error = err.Error()
			}
			rows[i] = row
		}()
	}
	wg.Wait()
	if err := writeJSONLines(filepath.Join(dataDir, "native.jsonl"), rows); err != nil {
		return err
	}

	withXS, bundledXS, failed := 0, 0, 0
	for _, r := range rows {
		switch {
		case r.Error != "":
			failed++
		case len(r.XS) > 0 && r.Bundled:
			bundledXS++
		case len(r.XS) > 0:
			withXS++
		}
	}
	fmt.Printf("%d distributions scanned: %d with XS code to build, %d with XS code already in the interpreter, %d unreadable\n",
		len(rows), withXS, bundledXS, failed)
	return nil
}

// bundled reports whether the interpreter ships every module of d at a
// version that satisfies d's own.
func bundled(d *dist, provided map[string]string) bool {
	if len(d.Modules) == 0 {
		return false
	}
	for _, m := range d.Modules {
		have, ok := provided[m]
		if !ok {
			return false
		}
		if enough, comparable := satisfies(have, d.Version); !comparable || !enough {
			return false
		}
	}
	return true
}

func readDists(path string) ([]dist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var dists []dist
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	for sc.Scan() {
		var d dist
		if err := json.Unmarshal(sc.Bytes(), &d); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		dists = append(dists, d)
	}
	return dists, sc.Err()
}

// fetchArchive returns the distribution's archive, from the data
// directory's archives/ tree when it is already there, verified against
// the recorded checksum either way.
func fetchArchive(dataDir string, d *dist) ([]byte, error) {
	local := filepath.Join(dataDir, "archives", filepath.FromSlash(d.Path))
	data, err := os.ReadFile(local)
	if err != nil {
		if data, err = httpGet(cpanMirror + d.Path); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(local+".tmp", data, 0o644); err != nil {
			return nil, err
		}
		if err := os.Rename(local+".tmp", local); err != nil {
			return nil, err
		}
	}
	if d.SHA256 != "" {
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != d.SHA256 {
			return nil, fmt.Errorf("%s: checksum %s, recorded %s", d.Path, got, d.SHA256)
		}
	}
	return data, nil
}

// nativeSources reads the archive's file list and returns its XS files
// and the number of its other C sources.
func nativeSources(name string, archive []byte) (xs []string, c int, err error) {
	names, err := archiveNames(name, archive)
	if err != nil {
		return nil, 0, err
	}
	for _, n := range names {
		// Entries sit below one top-level directory named after the
		// release.
		_, rel, ok := strings.Cut(path.Clean(n), "/")
		if !ok {
			continue
		}
		top, _, _ := strings.Cut(rel, "/")
		switch top {
		case "t", "xt", "inc", "eg", "examples":
			continue
		}
		switch path.Ext(rel) {
		case ".xs":
			xs = append(xs, rel)
		case ".c", ".h", ".cc", ".cpp", ".cxx", ".hpp":
			c++
		}
	}
	sort.Strings(xs)
	return xs, c, nil
}

// archiveNames lists the entries of a .tar.gz, .tgz, .tar.bz2 or .zip
// archive.
func archiveNames(name string, archive []byte) ([]string, error) {
	var names []string
	switch {
	case strings.HasSuffix(name, ".zip"):
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			names = append(names, f.Name)
		}
		return names, nil
	case strings.HasSuffix(name, ".tar.gz"), strings.HasSuffix(name, ".tgz"):
		gz, err := gzip.NewReader(bytes.NewReader(archive))
		if err != nil {
			return nil, err
		}
		return tarNames(gz)
	case strings.HasSuffix(name, ".tar.bz2"):
		return tarNames(bzip2.NewReader(bytes.NewReader(archive)))
	}
	return nil, fmt.Errorf("%s: unknown archive format", name)
}

func tarNames(r io.Reader) ([]string, error) {
	var names []string
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return names, nil
		}
		if err != nil {
			return nil, err
		}
		names = append(names, h.Name)
	}
}
