// Package ocitest is a registry for tests: a distribution API server that
// holds blobs and manifests in memory.
package ocitest

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// The media types of the images SeedBase makes.
const (
	OCIManifest    = "application/vnd.oci.image.manifest.v1+json"
	DockerManifest = "application/vnd.docker.distribution.manifest.v2+json"
)

// Manifest is a stored manifest.
type Manifest struct {
	MediaType string
	Body      []byte
}

// Registry is a distribution API server. With Token set, every request
// needs it as a bearer token, which its /token endpoint gives for the
// basic credentials user:pass; with Basic set, every request needs those
// credentials themselves.
type Registry struct {
	Token     string
	Basic     bool
	Corrupt   string              // a digest whose blob is served altered
	Blobs     map[string][]byte   // by digest
	Manifests map[string]Manifest // by repository + "@" + tag or digest
	Uploads   int                 // blobs uploaded
	Realm     string              // the token realm the challenge names (default: its /token)
	Scopes    []string            // the scopes tokens were asked for
	UploadURL string              // where uploads go (default: this registry)

	// UploadAuthorizations are the Authorization headers of the uploads'
	// PUTs, which this registry received for another one's UploadURL.
	UploadAuthorizations []string

	t   testing.TB
	srv *httptest.Server
	mu  sync.Mutex
}

// New starts a registry on the loopback interface, stopped when t ends.
func New(t testing.TB, token string) *Registry {
	r := &Registry{t: t, Token: token, Blobs: map[string][]byte{}, Manifests: map[string]Manifest{}}
	r.srv = httptest.NewServer(http.HandlerFunc(r.serve))
	t.Cleanup(r.srv.Close)
	return r
}

// Host is the registry's host:port.
func (r *Registry) Host() string { return strings.TrimPrefix(r.srv.URL, "http://") }

// Digest is data's digest.
func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (r *Registry) serve(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if strings.HasPrefix(req.URL.Path, "/upload/") {
		r.UploadAuthorizations = append(r.UploadAuthorizations, req.Header.Get("Authorization"))
	}
	if req.URL.Path == "/token" {
		r.Scopes = append(r.Scopes, req.URL.Query().Get("scope"))
		if user, pass, ok := req.BasicAuth(); !ok || user != "user" || pass != "pass" {
			http.Error(w, "bad credentials", http.StatusUnauthorized)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]string{"token": r.Token}); err != nil {
			r.t.Error(err)
		}
		return
	}
	if user, pass, ok := req.BasicAuth(); r.Basic && (!ok || user != "user" || pass != "pass") {
		w.Header().Set("WWW-Authenticate", `Basic realm="ocitest"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	upload := strings.HasPrefix(req.URL.Path, "/upload/")
	if r.Token != "" && !upload && req.Header.Get("Authorization") != "Bearer "+r.Token {
		realm := r.Realm
		if realm == "" {
			realm = r.srv.URL + "/token"
		}
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s",service="ocitest",scope="repository:x:pull"`, realm))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	path := strings.TrimPrefix(req.URL.Path, "/v2/")
	switch {
	case strings.HasPrefix(req.URL.Path, "/upload/"):
		body, err := io.ReadAll(req.Body)
		if err != nil {
			r.t.Error(err)
			return
		}
		digest := req.URL.Query().Get("digest")
		if Digest(body) != digest {
			http.Error(w, "digest mismatch", http.StatusBadRequest)
			return
		}
		r.Blobs[digest] = body
		r.Uploads++
		w.WriteHeader(http.StatusCreated)
	case strings.HasSuffix(path, "/blobs/uploads/") && req.Method == http.MethodPost:
		w.Header().Set("Location", r.UploadURL+"/upload/1?session=x")
		w.WriteHeader(http.StatusAccepted)
	case strings.Contains(path, "/blobs/"):
		_, digest, _ := strings.Cut(path, "/blobs/")
		data, ok := r.Blobs[digest]
		if !ok {
			http.NotFound(w, req)
			return
		}
		if digest == r.Corrupt {
			data = append([]byte("x"), data[1:]...)
		}
		if _, err := w.Write(data); err != nil {
			r.t.Error(err)
		}
	case strings.Contains(path, "/manifests/"):
		repo, id, _ := strings.Cut(path, "/manifests/")
		if req.Method == http.MethodPut {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				r.t.Error(err)
				return
			}
			r.putManifest(repo, id, req.Header.Get("Content-Type"), body)
			w.WriteHeader(http.StatusCreated)
			return
		}
		m, ok := r.Manifests[repo+"@"+id]
		if !ok {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", m.MediaType)
		if _, err := w.Write(m.Body); err != nil {
			r.t.Error(err)
		}
	default:
		http.NotFound(w, req)
	}
}

func (r *Registry) putManifest(repo, id, mediaType string, body []byte) {
	m := Manifest{mediaType, body}
	r.Manifests[repo+"@"+id] = m
	r.Manifests[repo+"@"+Digest(body)] = m
}

type descriptor struct {
	MediaType string            `json:"mediaType"`
	Digest    string            `json:"digest"`
	Size      int64             `json:"size"`
	Platform  map[string]string `json:"platform,omitempty"`
}

// SeedBase puts an index of base images for linux/amd64 and linux/arm64
// at base:latest in the manifest format manifestType: each has one layer
// holding /etc/base-<arch>, and a config with a user, an environment and
// a command. The images are at base:m-<arch> too.
func (r *Registry) SeedBase(manifestType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	layerType, indexType := "application/vnd.oci.image.layer.v1.tar+gzip", "application/vnd.oci.image.index.v1+json"
	if manifestType == DockerManifest {
		layerType, indexType = "application/vnd.docker.image.rootfs.diff.tar.gzip", "application/vnd.docker.distribution.manifest.list.v2+json"
	}
	blob := func(mediaType string, data []byte) descriptor {
		r.Blobs[Digest(data)] = data
		return descriptor{MediaType: mediaType, Digest: Digest(data), Size: int64(len(data))}
	}
	marshal := func(v any) []byte {
		data, err := json.Marshal(v)
		if err != nil {
			r.t.Fatal(err)
		}
		return data
	}
	var descs []descriptor
	for _, arch := range []string{"amd64", "arm64"} {
		tarball, gz := r.layer("etc/base-"+arch, "base")
		config := marshal(map[string]any{
			"architecture": arch, "os": "linux",
			"config":  map[string]any{"User": "65532", "Env": []string{"PATH=/usr/bin"}, "Cmd": []string{"/bin/sh"}},
			"rootfs":  map[string]any{"type": "layers", "diff_ids": []string{Digest(tarball)}},
			"history": []map[string]string{{"created_by": "base"}},
		})
		body := marshal(map[string]any{
			"schemaVersion": 2, "mediaType": manifestType,
			"config": blob("application/vnd.oci.image.config.v1+json", config),
			"layers": []descriptor{blob(layerType, gz)},
		})
		r.putManifest("base", "m-"+arch, manifestType, body)
		descs = append(descs, descriptor{MediaType: manifestType, Digest: Digest(body), Size: int64(len(body)),
			Platform: map[string]string{"os": "linux", "architecture": arch}})
	}
	r.putManifest("base", "latest", indexType, marshal(map[string]any{"schemaVersion": 2, "mediaType": indexType, "manifests": descs}))
}

// layer is a layer holding a file name with content, uncompressed and
// gzipped.
func (r *Registry) layer(name, content string) (tarball, gz []byte) {
	var tb bytes.Buffer
	tw := tar.NewWriter(&tb)
	if err := tw.WriteHeader(&tar.Header{Typeflag: tar.TypeReg, Name: name, Mode: 0o644, Size: int64(len(content)), ModTime: time.Unix(0, 0)}); err != nil {
		r.t.Fatal(err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		r.t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		r.t.Fatal(err)
	}
	var zb bytes.Buffer
	zw := gzip.NewWriter(&zb)
	if _, err := zw.Write(tb.Bytes()); err != nil {
		r.t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		r.t.Fatal(err)
	}
	return tb.Bytes(), zb.Bytes()
}
