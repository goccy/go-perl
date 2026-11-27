// Package oci builds container images and moves them between registries
// and files: the registry HTTP API of the OCI distribution specification,
// the image and index formats of the OCI image specification and their
// Docker equivalents, and the archive `docker load` reads.
package oci

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// Reference names an image: a repository in a registry, and a tag or a
// digest in it.
type Reference struct {
	Registry   string // as written ("docker.io" when none was)
	Repository string
	Tag        string
	Digest     string
}

// dockerHub is the registry of a name without one.
const dockerHub = "docker.io"

var (
	// The reference grammar of github.com/distribution/reference.
	repositoryRE = regexp.MustCompile(`^[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*(?:/[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*)*$`)
	tagRE        = regexp.MustCompile(`^[\w][\w.-]{0,127}$`)
	digestRE     = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
)

// ParseReference parses [registry/]repository[:tag][@digest]. A name
// without a registry is on Docker Hub, a single-component Docker Hub name
// is under library/, and a name with neither tag nor digest is tagged
// latest.
func ParseReference(s string) (Reference, error) {
	name, digest, hasDigest := strings.Cut(s, "@")
	var tag string
	if i := strings.LastIndex(name, ":"); i > strings.LastIndex(name, "/") {
		name, tag = name[:i], name[i+1:]
	}
	r := Reference{Registry: dockerHub, Repository: name, Tag: tag, Digest: digest}
	// The first component names a registry when it cannot be part of a
	// repository: it has a dot or a port, or it is localhost.
	if first, rest, ok := strings.Cut(name, "/"); ok && (strings.ContainsAny(first, ".:") || first == "localhost") {
		r.Registry, r.Repository = first, rest
	}
	if r.Registry == dockerHub && !strings.Contains(r.Repository, "/") {
		r.Repository = "library/" + r.Repository
	}
	switch {
	case !repositoryRE.MatchString(r.Repository):
		return Reference{}, fmt.Errorf("image reference %q: invalid repository %q", s, r.Repository)
	case tag != "" && !tagRE.MatchString(tag):
		return Reference{}, fmt.Errorf("image reference %q: invalid tag %q", s, tag)
	case hasDigest && !digestRE.MatchString(digest):
		return Reference{}, fmt.Errorf("image reference %q: invalid digest %q (want sha256:<64 hex digits>)", s, digest)
	}
	if tag == "" && !hasDigest {
		r.Tag = "latest"
	}
	return r, nil
}

// String writes r back as a reference.
func (r Reference) String() string {
	s := r.Registry + "/" + r.Repository
	if r.Tag != "" {
		s += ":" + r.Tag
	}
	if r.Digest != "" {
		s += "@" + r.Digest
	}
	return s
}

// Name is r without its tag and digest.
func (r Reference) Name() string { return r.Registry + "/" + r.Repository }

// identifier is what r's manifest is fetched by: its digest, else its tag.
func (r Reference) identifier() string {
	if r.Digest != "" {
		return r.Digest
	}
	return r.Tag
}

// apiBase is the URL r's registry serves the distribution API under.
// Docker Hub's API is on its own host, and a registry on the loopback
// interface is spoken to in plain HTTP, as docker itself does.
func (r Reference) apiBase() string {
	host := r.Registry
	if host == dockerHub {
		host = "registry-1.docker.io"
	}
	scheme := "https"
	if isLoopback(host) {
		scheme = "http"
	}
	return scheme + "://" + host + "/v2/" + r.Repository
}

// isLoopback reports whether host ("name" or "name:port") is on the
// loopback interface.
func isLoopback(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	return host == "localhost" || (ip != nil && ip.IsLoopback())
}
