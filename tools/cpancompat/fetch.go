package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	packagesURL = "https://cpan.metacpan.org/modules/02packages.details.txt.gz"
	metacpanAPI = "https://fastapi.metacpan.org/v1"

	packagesFile = "02packages.details.txt.gz"
	releasesFile = "releases.jsonl"
	// indexedOnlyFile holds releases that are not latest but still own
	// modules in the package index.
	indexedOnlyFile = "releases-indexed-only.jsonl"
	snapshotFile    = "snapshot.json"
)

// snapshot pins what a data directory was fetched from, so every later
// number can be traced to one state of CPAN.
type snapshot struct {
	FetchedAt      time.Time `json:"fetched_at"`
	PackagesURL    string    `json:"packages_url"`
	PackagesSHA256 string    `json:"packages_sha256"`
	Releases       int       `json:"releases"`
}

// releaseFields are the MetaCPAN release fields kept per distribution.
// tests is the CPAN Testers summary (pass/fail/na/unknown report counts).
var releaseFields = []string{
	"distribution", "name", "version", "author", "archive", "download_url",
	"checksum_sha256", "date", "maturity", "authorized", "dependency", "tests", "stat.size",
}

var httpClient = &http.Client{Timeout: 2 * time.Minute}

// fetch downloads the package index (module -> distribution archive, the
// index CPAN clients resolve against) and the metadata of every release
// MetaCPAN marks latest (dependencies and CPAN Testers summary).
func fetch(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "fetching", packagesURL)
	body, err := httpGet(packagesURL)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dataDir, packagesFile), body, 0o644); err != nil {
		return err
	}
	sum := sha256.Sum256(body)

	latest := map[string]any{"term": map[string]any{"status": "latest"}}
	n, err := fetchReleases(filepath.Join(dataDir, releasesFile), []map[string]any{latest})
	if err != nil {
		return err
	}
	if err := fetchIndexedOnly(dataDir); err != nil {
		return err
	}
	snap := snapshot{
		FetchedAt:      time.Now().UTC().Truncate(time.Second),
		PackagesURL:    packagesURL,
		PackagesSHA256: hex.EncodeToString(sum[:]),
		Releases:       n,
	}
	out, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, snapshotFile), append(out, '\n'), 0o644)
}

type scrollPage struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Total int `json:"total"`
		Hits  []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// fetchReleases scrolls through every latest release and writes one JSON
// document per line. It writes to a temporary file and renames, so an
// interrupted run never leaves a truncated result behind.
func fetchReleases(path string, queries []map[string]any) (int, error) {
	tmp, err := os.Create(path + ".tmp")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	n := 0
	for _, q := range queries {
		query, err := json.Marshal(map[string]any{"size": 1000, "query": q, "_source": releaseFields})
		if err != nil {
			return 0, err
		}
		page, err := scrollRequest(http.MethodPost, metacpanAPI+"/release/_search?scroll=5m", query)
		if err != nil {
			return 0, err
		}
		total, got := page.Hits.Total, 0
		for len(page.Hits.Hits) > 0 {
			for _, h := range page.Hits.Hits {
				var compact bytes.Buffer
				if err := json.Compact(&compact, h.Source); err != nil {
					return 0, err
				}
				compact.WriteByte('\n')
				if _, err := tmp.Write(compact.Bytes()); err != nil {
					return 0, err
				}
				got++
			}
			fmt.Fprintf(os.Stderr, "\rreleases: %d", n+got)
			next := metacpanAPI + "/_search/scroll?scroll=5m&scroll_id=" + url.QueryEscape(page.ScrollID)
			if page, err = scrollRequest(http.MethodGet, next, nil); err != nil {
				return 0, err
			}
		}
		if got != total {
			return 0, fmt.Errorf("release scroll ended at %d of %d", got, total)
		}
		n += got
	}
	fmt.Fprintln(os.Stderr)
	if err := tmp.Close(); err != nil {
		return 0, err
	}
	return n, os.Rename(tmp.Name(), path)
}

// fetchIndexedOnly fetches the releases the package index still points
// modules at although MetaCPAN no longer marks them latest: older perl
// releases (the index keeps a core module at the last release that
// changed it) and superseded uploads that still own a module. Without
// them a required module would resolve to an archive of unknown identity.
func fetchIndexedOnly(dataDir string) error {
	moduleArchive, err := readPackages(filepath.Join(dataDir, packagesFile))
	if err != nil {
		return err
	}
	releases, err := readReleases(filepath.Join(dataDir, releasesFile))
	if err != nil {
		return err
	}
	known := map[string]bool{}
	for _, r := range releases {
		if _, path, ok := strings.Cut(r.DownloadURL, "/authors/id/"); ok {
			known[path] = true
		}
	}
	archives := map[string]bool{}
	for _, p := range moduleArchive {
		if !known[p] {
			archives[path.Base(p)] = true
		}
	}
	names := make([]string, 0, len(archives))
	for a := range archives {
		names = append(names, a)
	}
	sort.Strings(names)
	var queries []map[string]any
	for len(names) > 0 {
		batch := names[:min(200, len(names))]
		names = names[len(batch):]
		queries = append(queries, map[string]any{"terms": map[string]any{"archive": batch}})
	}
	_, err = fetchReleases(filepath.Join(dataDir, indexedOnlyFile), queries)
	return err
}

func scrollRequest(method, u string, body []byte) (*scrollPage, error) {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 3 * time.Second)
		}
		req, err := http.NewRequest(method, u, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "cpancompat (go-perl CPAN compatibility survey)")
		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("%s: %s", u, resp.Status)
			continue
		}
		var page scrollPage
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, err
		}
		return &page, nil
	}
	return nil, lastErr
}

func httpGet(u string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "cpancompat (go-perl CPAN compatibility survey)")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", u, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
