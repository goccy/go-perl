package mod

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
)

// SnapshotFile is the lock file's name in a project directory.
const SnapshotFile = "cpanfile.snapshot"

// snapshotHeader is the first line of a carton snapshot.
const snapshotHeader = "# carton snapshot format: version 1.0"

// Package is one "Name version" line of a snapshot: a package a
// distribution provides, or a module it requires.
type Package struct {
	Name    string
	Version string // "undef" when the package declares none
}

// Distribution is one resolved distribution of a snapshot.
type Distribution struct {
	Name         string // "Cpanel-JSON-XS-4.38"
	Pathname     string // "R/RU/RURBAN/Cpanel-JSON-XS-4.38.tar.gz"
	Source              // where it comes from when not the CPAN mirror
	Provides     []Package
	Requirements []Package
}

// Snapshot is the content of a cpanfile.snapshot, in the carton format.
type Snapshot struct {
	Distributions []Distribution
}

// ReadSnapshot reads the snapshot at path.
func ReadSnapshot(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	snap, err := ParseSnapshot(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return snap, nil
}

// ParseSnapshot parses a carton snapshot. The format nests by indentation:
// a distribution at two spaces, its fields at four, list entries at six.
func ParseSnapshot(data []byte) (*Snapshot, error) {
	snap := &Snapshot{}
	var (
		dist *Distribution
		list *[]Package
	)
	sc := bufio.NewScanner(bytes.NewReader(data))
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		text := strings.TrimLeft(line, " ")
		indent := len(line) - len(text)
		switch {
		case text == "" || strings.HasPrefix(text, "#") || text == "DISTRIBUTIONS":
		case indent == 2:
			snap.Distributions = append(snap.Distributions, Distribution{Name: text})
			dist = &snap.Distributions[len(snap.Distributions)-1]
			list = nil
		case indent == 4 && dist != nil:
			key, value, ok := strings.Cut(text, ":")
			if !ok {
				return nil, fmt.Errorf("line %d: want \"key: value\", got %q", n, text)
			}
			switch key {
			case "pathname":
				dist.Pathname, list = strings.TrimSpace(value), nil
			case "mirror", "url", "git", "ref", "revision", "sha256":
				*dist.sourceField(key), list = strings.TrimSpace(value), nil
			case "provides":
				list = &dist.Provides
			case "requirements":
				list = &dist.Requirements
			default:
				return nil, fmt.Errorf("line %d: unknown field %q", n, key)
			}
		case indent == 6 && list != nil:
			name, version, ok := strings.Cut(text, " ")
			if !ok {
				return nil, fmt.Errorf("line %d: want \"Name version\", got %q", n, text)
			}
			*list = append(*list, Package{Name: name, Version: version})
		default:
			return nil, fmt.Errorf("line %d: unexpected indentation: %q", n, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	for _, d := range snap.Distributions {
		if d.Pathname == "" {
			return nil, fmt.Errorf("distribution %s has no pathname", d.Name)
		}
		if err := validateSource(&d); err != nil {
			return nil, fmt.Errorf("distribution %s: %w", d.Name, err)
		}
	}
	return snap, nil
}

// Format renders the snapshot the way carton writes it: distributions and
// their package lists sorted by name.
func (s *Snapshot) Format() []byte {
	dists := append([]Distribution(nil), s.Distributions...)
	sort.Slice(dists, func(i, j int) bool { return dists[i].Name < dists[j].Name })

	var b bytes.Buffer
	b.WriteString(snapshotHeader + "\nDISTRIBUTIONS\n")
	for _, d := range dists {
		fmt.Fprintf(&b, "  %s\n    pathname: %s\n", d.Name, d.Pathname)
		for _, key := range sourceFields {
			if value := *d.sourceField(key); value != "" {
				fmt.Fprintf(&b, "    %s: %s\n", key, value)
			}
		}
		for _, field := range []struct {
			name string
			list []Package
		}{{"provides", d.Provides}, {"requirements", d.Requirements}} {
			fmt.Fprintf(&b, "    %s:\n", field.name)
			list := append([]Package(nil), field.list...)
			sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
			for _, p := range list {
				fmt.Fprintf(&b, "      %s %s\n", p.Name, p.Version)
			}
		}
	}
	return b.Bytes()
}

// sourceFields are the snapshot fields of a Source, in the order they are
// written. A distribution from the CPAN mirror has none of them, so its
// entry stays as carton writes it.
var sourceFields = []string{"mirror", "url", "git", "ref", "revision", "sha256"}

// sourceField returns the Source field a snapshot key names.
func (d *Distribution) sourceField(key string) *string {
	return map[string]*string{"mirror": &d.Mirror, "url": &d.URL, "git": &d.Git, "ref": &d.Ref, "revision": &d.Revision, "sha256": &d.SHA256}[key]
}

// Provider returns the distribution that provides module, or nil.
func (s *Snapshot) Provider(module string) *Distribution {
	for i := range s.Distributions {
		for _, p := range s.Distributions[i].Provides {
			if p.Name == module {
				return &s.Distributions[i]
			}
		}
	}
	return nil
}
