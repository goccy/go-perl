package oci

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/goccy/go-perl/internal/oci/ocitest"
)

func credentials(string) (Credential, bool, error) { return Credential{"user", "pass"}, true, nil }

// An image is pulled for a platform from an index, gets the entrypoint
// layer, and is pushed as an index of the platforms' images, through
// token authentication; the pushed image keeps the base's config but for
// the entrypoint, command, layers and history.
func TestPullAppendPush(t *testing.T) {
	for _, manifestType := range []string{mediaOCIManifest, mediaDockerManifest} {
		t.Run(manifestType, func(t *testing.T) {
			reg := ocitest.New(t, "tok")
			reg.SeedBase(manifestType)
			c := &Client{Credential: credentials}
			ctx := context.Background()
			base, err := ParseReference(reg.Host() + "/base")
			if err != nil {
				t.Fatal(err)
			}
			var imgs []*Image
			for _, arch := range []string{"amd64", "arm64"} {
				b, err := c.Pull(ctx, base, Platform{OS: "linux", Architecture: arch})
				if err != nil {
					t.Fatal(err)
				}
				img, err := Append(b, "/app/hello", []byte("binary-"+arch))
				if err != nil {
					t.Fatal(err)
				}
				imgs = append(imgs, img)
			}
			dst, err := ParseReference(reg.Host() + "/team/app:v1")
			if err != nil {
				t.Fatal(err)
			}
			digest, err := c.Push(ctx, dst, imgs)
			if err != nil {
				t.Fatal(err)
			}
			top := reg.Manifests["team/app@v1"]
			if digestOf(top.Body) != digest {
				t.Fatalf("Push returned %s, the tag names %s", digest, digestOf(top.Body))
			}
			wantIndex := map[string]string{mediaOCIManifest: mediaOCIIndex, mediaDockerManifest: mediaDockerList}[manifestType]
			var idx index
			if err := json.Unmarshal(top.Body, &idx); err != nil || top.MediaType != wantIndex || idx.MediaType != wantIndex || len(idx.Manifests) != 2 {
				t.Fatalf("pushed index %s %s (%v)", top.MediaType, top.Body, err)
			}
			for i, arch := range []string{"amd64", "arm64"} {
				d := idx.Manifests[i]
				if d.Platform == nil || d.Platform.Architecture != arch || d.MediaType != manifestType {
					t.Fatalf("index entry %d: %+v", i, d)
				}
				var m manifest
				if err := json.Unmarshal(reg.Manifests["team/app@"+d.Digest].Body, &m); err != nil || len(m.Layers) != 2 {
					t.Fatalf("manifest for %s: %+v (%v)", arch, m, err)
				}
				wantLayer := map[string]string{mediaOCIManifest: mediaOCILayer, mediaDockerManifest: mediaDockerLayer}[manifestType]
				if m.Layers[1].MediaType != wantLayer {
					t.Errorf("appended layer type %s, want %s", m.Layers[1].MediaType, wantLayer)
				}
				wantFiles(t, reg.Blobs[m.Layers[1].Digest], map[string]string{"app/": "", "app/hello": "binary-" + arch})
				var cfg struct {
					Architecture string
					Config       map[string]any
					Rootfs       struct {
						DiffIDs []string `json:"diff_ids"`
					}
					History []map[string]string
					Created string
				}
				if err := json.Unmarshal(reg.Blobs[m.Config.Digest], &cfg); err != nil {
					t.Fatal(err)
				}
				tarball, _, err := entrypointLayer("/app/hello", []byte("binary-"+arch))
				if err != nil {
					t.Fatal(err)
				}
				wantCfg := map[string]any{"User": "65532", "Env": []any{"PATH=/usr/bin"}, "Entrypoint": []any{"/app/hello"}, "Cmd": nil}
				if cfg.Architecture != arch || !reflect.DeepEqual(cfg.Config, wantCfg) ||
					len(cfg.Rootfs.DiffIDs) != 2 || cfg.Rootfs.DiffIDs[1] != digestOf(tarball) ||
					len(cfg.History) != 2 || cfg.History[1]["created_by"] != "gperl image build" || cfg.Created != epoch {
					t.Errorf("config for %s: %s", arch, reg.Blobs[m.Config.Digest])
				}
			}
			// Pushing again uploads nothing: every blob is there.
			uploads := reg.Uploads
			if _, err := c.Push(ctx, dst, imgs); err != nil {
				t.Fatal(err)
			}
			if reg.Uploads != uploads {
				t.Errorf("a second push uploaded %d blobs", reg.Uploads-uploads)
			}
		})
	}
}

// wantFiles checks that the gzipped tar layer holds exactly the entries
// named (a directory with "" content), owned by root, dated epoch, files
// executable.
func wantFiles(t *testing.T, gz []byte, want map[string]string) {
	t.Helper()
	zr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(zr)
	got := map[string]string{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		if h.Mode != 0o755 || h.Uid != 0 || h.Gid != 0 || h.ModTime.Unix() != 0 {
			t.Errorf("%s: mode %o uid %d gid %d time %v", h.Name, h.Mode, h.Uid, h.Gid, h.ModTime)
		}
		got[h.Name] = string(data)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("layer holds %v, want %v", got, want)
	}
}

// A blob whose content is not what its digest says fails the pull, and a
// platform the index lacks is reported with the ones it has.
func TestPullRefuses(t *testing.T) {
	reg := ocitest.New(t, "")
	reg.SeedBase(mediaOCIManifest)
	c := &Client{Credential: credentials}
	base, err := ParseReference(reg.Host() + "/base")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Pull(context.Background(), base, Platform{OS: "linux", Architecture: "riscv64"}); err == nil || !strings.Contains(err.Error(), "has no image for linux/riscv64 (it has linux/amd64, linux/arm64)") {
		t.Errorf("Pull for riscv64 = %v", err)
	}
	var m manifest
	if err := json.Unmarshal(reg.Manifests["base@m-arm64"].Body, &m); err != nil {
		t.Fatal(err)
	}
	reg.Corrupt = m.Config.Digest
	if _, err := c.Pull(context.Background(), base, Platform{OS: "linux", Architecture: "arm64"}); err == nil || !strings.Contains(err.Error(), "bytes of digest") {
		t.Errorf("Pull of a corrupt config = %v", err)
	}
	if _, err := c.Push(context.Background(), base, nil); err == nil || !strings.Contains(err.Error(), "no image to push") {
		t.Errorf("Push of no images = %v", err)
	}
	byDigest, err := ParseReference(reg.Host() + "/base@" + m.Config.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Push(context.Background(), byDigest, nil); err == nil || !strings.Contains(err.Error(), "pushed to a tag") {
		t.Errorf("Push to a digest = %v", err)
	}
}

// The archive docker load reads holds the config, the layers by digest,
// and a manifest.json naming them and the tag.
func TestWriteTarball(t *testing.T) {
	reg := ocitest.New(t, "")
	reg.SeedBase(mediaOCIManifest)
	c := &Client{Credential: credentials}
	ctx := context.Background()
	base, err := ParseReference(reg.Host() + "/base")
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.Pull(ctx, base, Platform{OS: "linux", Architecture: "arm64"})
	if err != nil {
		t.Fatal(err)
	}
	img, err := Append(b, "/app/hello", []byte("binary"))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := c.WriteTarball(ctx, &buf, img, "example.com/app:v1"); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	tr := tar.NewReader(&buf)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if files[h.Name], err = io.ReadAll(tr); err != nil {
			t.Fatal(err)
		}
	}
	var entries []struct {
		Config   string
		RepoTags []string
		Layers   []string
	}
	if err := json.Unmarshal(files["manifest.json"], &entries); err != nil || len(entries) != 1 {
		t.Fatalf("manifest.json %s (%v)", files["manifest.json"], err)
	}
	e := entries[0]
	if !reflect.DeepEqual(e.RepoTags, []string{"example.com/app:v1"}) || len(e.Layers) != 2 || len(files) != 4 {
		t.Fatalf("manifest.json %+v, files %d", e, len(files))
	}
	for _, name := range append([]string{e.Config}, e.Layers...) {
		sum := sha256.Sum256(files[name])
		if name != "blobs/sha256/"+hex.EncodeToString(sum[:]) {
			t.Errorf("%s holds content of digest %x", name, sum)
		}
	}
	if !bytes.Equal(files[e.Config], img.configRaw) {
		t.Errorf("config %s", files[e.Config])
	}
}

func TestParseReference(t *testing.T) {
	for in, want := range map[string]Reference{
		"alpine":                           {"docker.io", "library/alpine", "latest", ""},
		"me/app:v1":                        {"docker.io", "me/app", "v1", ""},
		"ghcr.io/me/app":                   {"ghcr.io", "me/app", "latest", ""},
		"localhost:5000/app:dev":           {"localhost:5000", "app", "dev", ""},
		"cgr.dev/chainguard/static:latest": {"cgr.dev", "chainguard/static", "latest", ""},
		"ghcr.io/me/app@sha256:" + strings.Repeat("a", 64): {"ghcr.io", "me/app", "", "sha256:" + strings.Repeat("a", 64)},
	} {
		got, err := ParseReference(in)
		if err != nil || got != want {
			t.Errorf("ParseReference(%q) = %+v, %v; want %+v", in, got, err, want)
		}
	}
	for _, in := range []string{"Upper/app", "ghcr.io/me/app:bad tag", "app@sha256:abc", "ghcr.io/"} {
		if _, err := ParseReference(in); err == nil {
			t.Errorf("ParseReference(%q) succeeded", in)
		}
	}
	for ref, want := range map[string]string{
		"alpine":             "https://registry-1.docker.io/v2/library/alpine",
		"localhost:5000/app": "http://localhost:5000/v2/app",
		"127.0.0.1:5000/app": "http://127.0.0.1:5000/v2/app",
		"ghcr.io/me/app":     "https://ghcr.io/v2/me/app",
	} {
		r, err := ParseReference(ref)
		if err != nil || r.apiBase() != want {
			t.Errorf("apiBase of %s = %s (%v), want %s", ref, r.apiBase(), err, want)
		}
	}
}

func TestParseChallenge(t *testing.T) {
	c, err := parseChallenge(`Bearer realm="https://auth.example/token",service="registry.example",scope="repository:a/b:pull,push"`)
	want := map[string]string{"realm": "https://auth.example/token", "service": "registry.example", "scope": "repository:a/b:pull,push"}
	if err != nil || c.scheme != "bearer" || !reflect.DeepEqual(c.params, want) {
		t.Errorf("parseChallenge = %+v, %v", c, err)
	}
	if c, err := parseChallenge(`Basic realm=registry`); err != nil || c.scheme != "basic" || c.params["realm"] != "registry" {
		t.Errorf("parseChallenge(Basic) = %+v, %v", c, err)
	}
	for _, bad := range []string{`Bearer realm`, `Bearer realm="x`, `Bearer realm="x" service="y"`} {
		if _, err := parseChallenge(bad); err == nil {
			t.Errorf("parseChallenge(%q) succeeded", bad)
		}
	}
}

// Credentials come from the registry's credential helper, else the
// credential store, else config.json's entry; a helper with none for the
// server leaves the registry anonymous.
func TestDockerCredential(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dir)
	bin := t.TempDir()
	t.Setenv("PATH", bin)
	helper := func(name, script string) {
		if err := os.WriteFile(filepath.Join(bin, "docker-credential-"+name), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	helper("fake", `read server; echo "{\"ServerURL\":\"$server\",\"Username\":\"helper-user\",\"Secret\":\"$server\"}"`+"\n")
	helper("empty", "echo '"+credentialsNotFound+"'; exit 1\n")
	config := func(v any) {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok, err := DockerCredential("ghcr.io"); ok || err != nil {
		t.Errorf("no config.json: %v %v", ok, err)
	}
	config(map[string]any{
		"auths": map[string]any{
			"https://index.docker.io/v1/": map[string]string{"auth": base64.StdEncoding.EncodeToString([]byte("hub:secret"))},
			"ghcr.io":                     map[string]string{"auth": base64.StdEncoding.EncodeToString([]byte("gh:token"))},
		},
		"credHelpers": map[string]string{"example.com": "fake"},
	})
	for registry, want := range map[string]Credential{
		"docker.io":   {"hub", "secret"},
		"ghcr.io":     {"gh", "token"},
		"example.com": {"helper-user", "example.com"},
	} {
		if got, ok, err := DockerCredential(registry); !ok || err != nil || got != want {
			t.Errorf("DockerCredential(%s) = %+v %v %v, want %+v", registry, got, ok, err, want)
		}
	}
	if _, ok, err := DockerCredential("quay.io"); ok || err != nil {
		t.Errorf("a registry config.json does not name: %v %v", ok, err)
	}
	config(map[string]any{"credsStore": "empty"})
	if _, ok, err := DockerCredential("ghcr.io"); ok || err != nil {
		t.Errorf("a store without credentials for the server: %v %v", ok, err)
	}
	config(map[string]any{"credsStore": "missing"})
	if _, _, err := DockerCredential("ghcr.io"); err == nil {
		t.Error("a credential store that cannot run gave no error")
	}
}

// A registry asking for basic credentials gets docker's; one whose token
// endpoint refuses them, or asking with no docker credentials, fails with
// what happened.
func TestAuthorization(t *testing.T) {
	ctx := context.Background()
	reg := ocitest.New(t, "")
	reg.Basic = true
	reg.SeedBase(mediaOCIManifest)
	base, err := ParseReference(reg.Host() + "/base")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&Client{Credential: credentials}).Pull(ctx, base, Platform{OS: "linux", Architecture: "amd64"}); err != nil {
		t.Errorf("Pull with basic credentials: %v", err)
	}
	none := func(string) (Credential, bool, error) { return Credential{}, false, nil }
	if _, err := (&Client{Credential: none}).Pull(ctx, base, Platform{OS: "linux", Architecture: "amd64"}); err == nil || !strings.Contains(err.Error(), "docker has none for it") {
		t.Errorf("Pull without credentials from a basic registry = %v", err)
	}

	tokenReg := ocitest.New(t, "tok")
	tokenReg.SeedBase(mediaOCIManifest)
	base, err = ParseReference(tokenReg.Host() + "/base")
	if err != nil {
		t.Fatal(err)
	}
	wrong := func(string) (Credential, bool, error) { return Credential{"user", "wrong"}, true, nil }
	if _, err := (&Client{Credential: wrong}).Pull(ctx, base, Platform{OS: "linux", Architecture: "amd64"}); err == nil || !strings.Contains(err.Error(), "401 Unauthorized") {
		t.Errorf("Pull with refused credentials = %v", err)
	}
	failing := func(string) (Credential, bool, error) { return Credential{}, false, fmt.Errorf("helper broke") }
	if _, err := (&Client{Credential: failing}).Pull(ctx, base, Platform{OS: "linux", Architecture: "amd64"}); err == nil || !strings.Contains(err.Error(), "helper broke") {
		t.Errorf("Pull with a failing credential lookup = %v", err)
	}
}

// Pushing asks for a token to pull and push the repository, and an upload
// location on another host (a storage service's signed URL) gets no
// authorization.
func TestPushAuthorization(t *testing.T) {
	ctx := context.Background()
	storage := ocitest.New(t, "")
	reg := ocitest.New(t, "tok")
	reg.UploadURL = "http://" + storage.Host()
	reg.SeedBase(mediaOCIManifest)
	c := &Client{Credential: credentials}
	base, err := ParseReference(reg.Host() + "/base")
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.Pull(ctx, base, Platform{OS: "linux", Architecture: "amd64"})
	if err != nil {
		t.Fatal(err)
	}
	img, err := Append(b, "/app/hello", []byte("binary"))
	if err != nil {
		t.Fatal(err)
	}
	dst, err := ParseReference(reg.Host() + "/team/app:v1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Push(ctx, dst, []*Image{img}); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(reg.Scopes, "repository:team/app:pull,push") {
		t.Errorf("token scopes %q, want one to pull and push team/app", reg.Scopes)
	}
	if len(storage.UploadAuthorizations) == 0 {
		t.Fatal("nothing was uploaded to the upload location")
	}
	for _, a := range storage.UploadAuthorizations {
		if a != "" {
			t.Errorf("the upload location got Authorization %q", a)
		}
	}
}

// Credentials are not sent to a token realm over plain HTTP off the
// loopback interface; digests other than sha256 are refused before
// anything is fetched by them; an index of mixed manifest formats is
// refused.
func TestRefusals(t *testing.T) {
	ctx := context.Background()
	reg := ocitest.New(t, "tok")
	reg.Realm = "http://auth.example/token"
	reg.SeedBase(mediaOCIManifest)
	base, err := ParseReference(reg.Host() + "/base")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&Client{Credential: credentials}).Pull(ctx, base, Platform{OS: "linux", Architecture: "amd64"}); err == nil || !strings.Contains(err.Error(), "is not HTTPS") {
		t.Errorf("Pull with a plain HTTP realm = %v", err)
	}

	plain := ocitest.New(t, "")
	plain.SeedBase(mediaOCIManifest)
	base, err = ParseReference(plain.Host() + "/base")
	if err != nil {
		t.Fatal(err)
	}
	top := plain.Manifests["base@latest"]
	plain.Manifests["base@latest"] = ocitest.Manifest{MediaType: top.MediaType, Body: bytes.Replace(top.Body, []byte(`"sha256:`), []byte(`"sha512:`), 1)}
	c := &Client{Credential: credentials}
	if _, err := c.Pull(ctx, base, Platform{OS: "linux", Architecture: "amd64"}); err == nil || !strings.Contains(err.Error(), "unsupported digest") {
		t.Errorf("Pull through a sha512 index entry = %v", err)
	}
	plain.Manifests["base@latest"] = top
	m := plain.Manifests["base@m-arm64"]
	plain.Manifests["base@m-arm64"] = ocitest.Manifest{MediaType: m.MediaType, Body: bytes.Replace(m.Body, []byte(`"sha256:`), []byte(`"../../x:`), 1)}
	byTag, err := ParseReference(plain.Host() + "/base:m-arm64")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Pull(ctx, byTag, Platform{OS: "linux", Architecture: "arm64"}); err == nil || !strings.Contains(err.Error(), "unsupported digest") {
		t.Errorf("Pull of a manifest with a malformed digest = %v", err)
	}

	docker := ocitest.New(t, "")
	docker.SeedBase(mediaDockerManifest)
	dbase, err := ParseReference(docker.Host() + "/base")
	if err != nil {
		t.Fatal(err)
	}
	oci, err := c.Pull(ctx, base, Platform{OS: "linux", Architecture: "amd64"})
	if err != nil {
		t.Fatal(err)
	}
	dock, err := c.Pull(ctx, dbase, Platform{OS: "linux", Architecture: "arm64"})
	if err != nil {
		t.Fatal(err)
	}
	dst, err := ParseReference(plain.Host() + "/app:v1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Push(ctx, dst, []*Image{oci, dock}); err == nil || !strings.Contains(err.Error(), "different manifest formats") {
		t.Errorf("Push of mixed formats = %v", err)
	}
}

func TestSameOrigin(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"https://reg.example/v2/app", "https://reg.example/upload/1", true},
		{"https://reg.example/v2/app", "https://REG.example:443/upload/1", true},
		{"http://127.0.0.1:5000/v2/app", "http://127.0.0.1:5000/upload", true},
		{"https://reg.example/v2/app", "http://reg.example/upload", false},
		{"https://reg.example/v2/app", "https://reg.example:8443/upload", false},
		{"https://reg.example/v2/app", "https://storage.example/upload", false},
	} {
		a, err := url.Parse(c.a)
		if err != nil {
			t.Fatal(err)
		}
		b, err := url.Parse(c.b)
		if err != nil {
			t.Fatal(err)
		}
		if got := sameOrigin(a, b); got != c.want {
			t.Errorf("sameOrigin(%s, %s) = %v", c.a, c.b, got)
		}
	}
}
