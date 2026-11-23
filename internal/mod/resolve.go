package mod

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// release is a resolved distribution release.
type release struct {
	Pathname string // below authors/id
	Version  string // the version of the requested module in it
}

// resolve finds the release that provides module at a version rng accepts:
// the mirror's index when only the mirror may be used, otherwise CPAN
// MetaDB and then MetaCPAN — the order cpanm and cpm use.
func (m *Manager) resolve(ctx context.Context, module, rng string) (*release, error) {
	if m.MirrorOnly {
		return m.resolveIndex(ctx, module, rng)
	}
	rel, err := m.resolveMetaDB(ctx, module, rng)
	if err == nil || !errors.Is(err, errNotFound) {
		return rel, err
	}
	rel, cpanErr := m.resolveMetaCPAN(ctx, module, rng)
	if errors.Is(cpanErr, errNotFound) {
		return nil, fmt.Errorf("%s (%s): no release on CPAN satisfies it: %w", module, rng, errNotFound)
	}
	return rel, cpanErr
}

// needsHistory reports whether rng can exclude the latest release, so that
// older releases have to be considered.
func needsHistory(rng string) bool {
	return strings.Contains(rng, "<") || strings.Contains(rng, "!=") || strings.Contains(rng, "==")
}

func (m *Manager) resolveMetaDB(ctx context.Context, module, rng string) (*release, error) {
	if needsHistory(rng) {
		return m.resolveMetaDBHistory(ctx, module, rng)
	}
	body, err := m.get(ctx, m.MetaDB+"package/"+url.PathEscape(module))
	if err != nil {
		return nil, err
	}
	// The answer is a flat YAML mapping; only two top-level keys matter.
	rel := &release{}
	sc := bufio.NewScanner(bytes.NewReader(body))
	for sc.Scan() {
		key, value, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		switch key {
		case "distfile":
			rel.Pathname = strings.TrimSpace(value)
		case "version":
			rel.Version = strings.Trim(strings.TrimSpace(value), `'"`)
		}
	}
	if rel.Pathname == "" {
		return nil, fmt.Errorf("%s: MetaDB answer has no distfile: %w", module, errNotFound)
	}
	ok, err := m.helper.Accepts(rng, rel.Version)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%s: latest release %s does not satisfy %q: %w", module, rel.Version, rng, errNotFound)
	}
	return rel, nil
}

// resolveMetaDBHistory picks the highest release in the module's history
// that rng accepts. Each history line is "Module version distfile".
func (m *Manager) resolveMetaDBHistory(ctx context.Context, module, rng string) (*release, error) {
	body, err := m.get(ctx, m.MetaDB+"history/"+url.PathEscape(module))
	if err != nil {
		return nil, err
	}
	var versions, pathnames []string
	sc := bufio.NewScanner(bytes.NewReader(body))
	for sc.Scan() {
		if f := strings.Fields(sc.Text()); len(f) == 3 && f[0] == module {
			versions, pathnames = append(versions, f[1]), append(pathnames, f[2])
		}
	}
	i, err := m.helper.Best(rng, versions...)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		return nil, fmt.Errorf("%s: none of %d releases satisfies %q: %w", module, len(versions), rng, errNotFound)
	}
	return &release{Pathname: pathnames[i], Version: versions[i]}, nil
}

func (m *Manager) resolveMetaCPAN(ctx context.Context, module, rng string) (*release, error) {
	u := m.MetaCPAN + url.PathEscape(module)
	if rng != "" && rng != "0" {
		u += "?version=" + url.QueryEscape(rng)
	}
	body, err := m.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var answer struct {
		DownloadURL string `json:"download_url"`
		Version     any    `json:"version"`
	}
	if err := json.Unmarshal(body, &answer); err != nil {
		return nil, fmt.Errorf("%s: MetaCPAN answer: %w", module, err)
	}
	_, pathname, ok := strings.Cut(answer.DownloadURL, "/authors/id/")
	if !ok {
		return nil, fmt.Errorf("%s: MetaCPAN answer has no CPAN download URL: %w", module, errNotFound)
	}
	return &release{Pathname: pathname, Version: fmt.Sprint(answer.Version)}, nil
}

// indexEntry is one line of a mirror's 02packages index.
type indexEntry struct {
	version  string
	pathname string
}

func (m *Manager) resolveIndex(ctx context.Context, module, rng string) (*release, error) {
	index, err := m.mirrorIndex(ctx)
	if err != nil {
		return nil, err
	}
	entry, ok := index[module]
	if !ok {
		return nil, fmt.Errorf("%s: not in the index of %s: %w", module, m.Mirror, errNotFound)
	}
	accepted, err := m.helper.Accepts(rng, entry.version)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return nil, fmt.Errorf("%s: %s indexes %s, which does not satisfy %q: %w", module, m.Mirror, entry.version, rng, errNotFound)
	}
	return &release{Pathname: entry.pathname, Version: entry.version}, nil
}

// mirrorIndex loads the mirror's modules/02packages.details.txt.gz once:
// a header, a blank line, then "Module version pathname" lines.
func (m *Manager) mirrorIndex(ctx context.Context) (map[string]indexEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.index != nil {
		return m.index, nil
	}
	u := m.Mirror + "modules/02packages.details.txt.gz"
	body, err := m.get(ctx, u)
	if err != nil {
		return nil, err
	}
	zr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", u, err)
	}
	text, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", u, err)
	}
	index := map[string]indexEntry{}
	inBody := false
	sc := bufio.NewScanner(bytes.NewReader(text))
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		if !inBody {
			inBody = sc.Text() == ""
			continue
		}
		if f := strings.Fields(sc.Text()); len(f) == 3 {
			index[f[0]] = indexEntry{version: f[1], pathname: f[2]}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", u, err)
	}
	m.index = index
	return index, nil
}
