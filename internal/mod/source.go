package mod

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// Source is where a distribution comes from when it is not the CPAN
// mirror: what a cpanfile requirement's options name, and what pins it.
type Source struct {
	Mirror   string // the mirror the release at Pathname is on
	URL      string // the URL of the distribution's archive
	Git      string // the URL of the distribution's git repository
	Ref      string // the branch, tag, or commit Git names; "" for its default branch
	Revision string // the commit Git was resolved to
	SHA256   string // the archive's digest, recorded on its first download (Mirror, URL)
}

var (
	sha256Hex   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	revisionHex = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
	pauseID     = regexp.MustCompile(`^[A-Z][A-Z0-9-]+$`)
	notInName   = regexp.MustCompile(`[^A-Za-z0-9_.-]`)
	// transportName is what git takes for a transport before "::".
	transportName = regexp.MustCompile(`^[A-Za-z0-9+.-]+$`)
)

// options is every option of a cpanfile requirement gperl knows.
var options = []string{"dist", "git", "mirror", "ref", "url"}

// pinned returns the distribution the options of a cpanfile requirement
// name — its revision or digest still unknown — or nil when they name none.
func pinned(p Prereq) (*Distribution, error) {
	opts := p.Options
	for key, value := range opts {
		if !slices.Contains(options, key) {
			return nil, fmt.Errorf("option %s is not supported", key)
		}
		// No URL, release, or git ref has white space; the snapshot
		// keeps each value on a line of its own.
		if value == "" || strings.ContainsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
			return nil, fmt.Errorf("option %s %q: an empty value or white space", key, value)
		}
	}
	has := func(key string) bool { _, ok := opts[key]; return ok }
	switch {
	case has("git"):
		if has("url") || has("dist") || has("mirror") {
			return nil, errors.New("option git cannot be combined with url, dist, or mirror")
		}
		if err := validateGitURL(opts["git"]); err != nil {
			return nil, err
		}
		return &Distribution{Source: Source{Git: opts["git"], Ref: opts["ref"]}}, nil
	case has("ref"):
		return nil, errors.New("option ref needs git")
	case has("url"):
		if has("dist") || has("mirror") {
			return nil, errors.New("option url cannot be combined with dist or mirror")
		}
		u, err := archiveURL(opts["url"])
		if err != nil {
			return nil, err
		}
		d := &Distribution{Pathname: path.Base(u.Path), Source: Source{URL: opts["url"]}}
		if _, _, err := SplitPathname(d.Pathname); err != nil {
			return nil, err
		}
		return d, nil
	case has("dist"):
		pathname, err := distPathname(opts["dist"])
		if err != nil {
			return nil, err
		}
		d := &Distribution{Pathname: pathname}
		if has("mirror") {
			if _, err := archiveURL(opts["mirror"]); err != nil {
				return nil, err
			}
			d.Mirror = strings.TrimSuffix(opts["mirror"], "/") + "/"
		}
		return d, nil
	case has("mirror"):
		return nil, errors.New("option mirror needs dist")
	}
	return nil, nil
}

// distPathname expands a dist option ("AUTHOR/Dist-1.0.tar.gz", or the
// path below authors/id itself) into the path below authors/id, as cpanm
// does.
func distPathname(dist string) (string, error) {
	parts := strings.Split(dist, "/")
	if len(parts) >= 2 && len(parts[0]) > 1 {
		parts = append([]string{parts[0][:1], parts[0][:2]}, parts...)
	}
	pathname := strings.Join(parts, "/")
	if len(parts) < 4 || !pauseID.MatchString(parts[2]) || parts[0] != parts[2][:1] || parts[1] != parts[2][:2] ||
		!filepath.IsLocal(filepath.FromSlash(pathname)) || path.Clean(pathname) != pathname {
		return "", fmt.Errorf("dist %q: want AUTHOR/Dist-1.0.tar.gz", dist)
	}
	if _, _, err := SplitPathname(pathname); err != nil {
		return "", err
	}
	return pathname, nil
}

// archiveURL parses the URL of an archive or a mirror: https, or http to
// the loopback interface. Anything an attacker on the network could change
// would be pinned by the snapshot as it came.
func archiveURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Host == "" || (u.Scheme != "https" && (u.Scheme != "http" || !isLoopback(u.Hostname()))) {
		return nil, fmt.Errorf("%s: want an https URL", raw)
	}
	return u, nil
}

// validateGitURL accepts a repository URL over https or ssh (or http to
// the loopback interface), or git's scp-like ssh form "user@host:path".
func validateGitURL(raw string) error {
	if strings.HasPrefix(raw, "-") {
		return fmt.Errorf("git %q: want an https or ssh URL", raw)
	}
	if !strings.Contains(raw, "://") {
		// git reads "host:path" as ssh when no slash comes before the
		// colon, and "transport::address" as a remote helper's address.
		if host, rest, ok := strings.Cut(raw, ":"); ok && host != "" && !strings.Contains(host, "/") &&
			!(strings.HasPrefix(rest, ":") && transportName.MatchString(host)) {
			return nil
		}
		return fmt.Errorf("git %q: want an https or ssh URL", raw)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme == "ssh" && u.Host != "" {
		return nil
	}
	if _, err := archiveURL(raw); err != nil {
		return fmt.Errorf("git %q: want an https or ssh URL", raw)
	}
	return nil
}

func isLoopback(host string) bool {
	ip := net.ParseIP(host)
	return host == "localhost" || (ip != nil && ip.IsLoopback())
}

// repoName is the name a git repository's distribution goes by: the last
// element of its path without ".git", with the characters a directory or a
// Go import path element may not carry replaced, as cpm names its clones.
func repoName(repo string) (string, error) {
	p := repo
	if u, err := url.Parse(repo); err == nil && u.Scheme != "" {
		p = u.Path
	} else if _, after, ok := strings.Cut(repo, ":"); ok {
		p = after
	}
	name := strings.TrimSuffix(path.Base(p), ".git")
	name = notInName.ReplaceAllString(name, "-")
	if strings.Trim(name, ".") == "" {
		return "", fmt.Errorf("git %q: no repository name in the URL", repo)
	}
	return name, nil
}

// sameSource reports whether the distribution d of a snapshot is the one
// a cpanfile requirement's options name (pin, nil when they name none).
func sameSource(pin, d *Distribution) bool {
	switch {
	case pin == nil:
		return d.Mirror == "" && d.URL == "" && d.Git == ""
	case pin.Git != "":
		return pin.Git == d.Git && pin.Ref == d.Ref
	case pin.URL != "":
		return pin.URL == d.URL
	}
	return pin.Pathname == d.Pathname && pin.Mirror == d.Mirror && d.URL == "" && d.Git == ""
}

// samePin reports whether two cpanfile requirements' options name the same
// distribution (or both name none).
func samePin(a, b *Distribution) bool {
	if a == nil || b == nil {
		return a == b
	}
	return sameSource(a, b)
}

// distName is the name of d's distribution: two releases of one
// distribution cannot both be selected.
func distName(d *Distribution) (string, error) {
	if d.Git != "" {
		return repoName(d.Git)
	}
	name, _, err := SplitPathname(d.Pathname)
	return name, err
}

// validateSource reports a snapshot entry whose source is incomplete, or
// names a URL a cpanfile could not.
func validateSource(d *Distribution) error {
	if d.Git != "" {
		if err := validateGitURL(d.Git); err != nil {
			return err
		}
	}
	for _, u := range []string{d.URL, d.Mirror} {
		if u != "" {
			if _, err := archiveURL(u); err != nil {
				return err
			}
		}
	}
	switch {
	case d.Git != "" && (d.URL != "" || d.Mirror != ""):
		return errors.New("git cannot be combined with url or mirror")
	case d.URL != "" && d.Mirror != "":
		return errors.New("url cannot be combined with mirror")
	case d.Git != "" && !revisionHex.MatchString(d.Revision):
		return fmt.Errorf("revision %q is not a commit id", d.Revision)
	case d.Git == "" && (d.Revision != "" || d.Ref != ""):
		return errors.New("ref and revision need git")
	case (d.URL != "" || d.Mirror != "") && !sha256Hex.MatchString(d.SHA256):
		return fmt.Errorf("sha256 %q is not a sha256 digest", d.SHA256)
	case d.URL == "" && d.Mirror == "" && d.SHA256 != "":
		return errors.New("sha256 needs url or mirror")
	}
	return nil
}
