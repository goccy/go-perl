package gperl_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"debug/elf"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-perl/internal/gperl"
	"github.com/goccy/go-perl/internal/oci/ocitest"
)

// Arguments that cannot make an image are refused before anything is
// built.
func TestImageBuildArguments(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"app.pl"}, "give -push, -tarball, or both"},
		{[]string{"-push", "r.example/app:v1"}, "usage: gperl image build"},
		{[]string{"-push", "r.example/app:v1", "a.pl", "b.pl"}, "usage: gperl image build"},
		{[]string{"-tarball", "x.tar", "app.pl"}, "docker load reads an image of one platform"},
		{[]string{"-push", "r.example/app:v1", "-platform", "darwin/arm64", "app.pl"}, `platform "darwin/arm64"`},
		{[]string{"-push", "r.example/app:v1", "-platform", "linux", "app.pl"}, `platform "linux"`},
		{[]string{"-push", "r.example/app:v1", "-platform", "linux/arm", "app.pl"}, `want linux/amd64 or linux/arm64`},
		{[]string{"-push", "r.example/app:v1", "-platform", "linux/arm64/v8", "app.pl"}, `platform "linux/arm64/v8"`},
		{[]string{"-push", "Bad/Name", "app.pl"}, "invalid repository"},
		{[]string{"-push", "r.example/app:v1", "-psgi", "app.psgi", "--", "-o", "out"}, "-o is not an option of an image"},
		{[]string{"-push", "r.example/app:v1", "-psgi", "app.psgi", "--", "extra"}, "usage: gperl psgi build"},
	} {
		if err := gperl.ImageBuild(c.args); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("ImageBuild(%q) = %v, want %q", c.args, err, c.want)
		}
	}
}

// gperl image build builds the script for each platform, puts each binary
// at /app/<name> on the base image as its entrypoint, and pushes an index
// of the platforms' images through the registry's token authentication
// with docker's credentials; for one platform it also writes the archive
// docker load reads, tagged as pushed.
func TestImageBuild(t *testing.T) {
	requireToolchain(t)
	t.Setenv("GPERL_DEV_REPLACE", devReplaces(t))
	t.Setenv("GPERLMODCACHE", t.TempDir())
	reg := ocitest.New(t, "tok")
	reg.SeedBase(ocitest.OCIManifest)
	dockerConfig := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dockerConfig)
	auth, err := json.Marshal(map[string]any{"auths": map[string]any{
		reg.Host(): map[string]string{"auth": base64.StdEncoding.EncodeToString([]byte("user:pass"))},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dockerConfig, "config.json"), auth, 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "hello.pl")
	if err := os.WriteFile(script, []byte("print \"hello\\n\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	base := reg.Host() + "/base"

	if err := gperl.ImageBuild([]string{"-base", base, "-push", reg.Host() + "/app:v1", script}); err != nil {
		t.Fatal(err)
	}
	top, ok := reg.Manifests["app@v1"]
	if !ok {
		t.Fatal("nothing was pushed to app:v1")
	}
	var idx struct {
		MediaType string
		Manifests []struct {
			Digest   string
			Platform struct{ OS, Architecture string }
		}
	}
	if err := json.Unmarshal(top.Body, &idx); err != nil || idx.MediaType != "application/vnd.oci.image.index.v1+json" || len(idx.Manifests) != 2 {
		t.Fatalf("app:v1 is %s (%v)", top.Body, err)
	}
	for i, arch := range []string{"amd64", "arm64"} {
		d := idx.Manifests[i]
		if d.Platform.OS != "linux" || d.Platform.Architecture != arch {
			t.Fatalf("index entry %d is for %+v", i, d.Platform)
		}
		var m struct {
			Config struct{ Digest string }
			Layers []struct{ Digest string }
		}
		if err := json.Unmarshal(reg.Manifests["app@"+d.Digest].Body, &m); err != nil || len(m.Layers) != 2 {
			t.Fatalf("image for %s: %+v (%v)", arch, m, err)
		}
		var cfg struct{ Config map[string]any }
		if err := json.Unmarshal(reg.Blobs[m.Config.Digest], &cfg); err != nil {
			t.Fatal(err)
		}
		if got := cfg.Config["Entrypoint"]; !reflect.DeepEqual(got, []any{"/app/hello"}) || cfg.Config["User"] != "65532" {
			t.Errorf("config for %s: %s", arch, reg.Blobs[m.Config.Digest])
		}
		wantBinary(t, reg.Blobs[m.Layers[1].Digest], "app/hello", map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64}[arch])
	}

	tarball := filepath.Join(t.TempDir(), "app.tar")
	if err := gperl.ImageBuild([]string{"-base", base, "-platform", "linux/arm64", "-push", reg.Host() + "/app:arm64", "-tarball", tarball, script}); err != nil {
		t.Fatal(err)
	}
	if _, ok := reg.Manifests["app@arm64"]; !ok {
		t.Error("nothing was pushed to app:arm64")
	}
	files := readTar(t, tarball)
	var entries []struct{ RepoTags, Layers []string }
	if err := json.Unmarshal(files["manifest.json"], &entries); err != nil || len(entries) != 1 ||
		!reflect.DeepEqual(entries[0].RepoTags, []string{reg.Host() + "/app:arm64"}) || len(entries[0].Layers) != 2 {
		t.Fatalf("manifest.json %s (%v)", files["manifest.json"], err)
	}
	wantBinary(t, files[entries[0].Layers[1]], "app/hello", elf.EM_AARCH64)

	// A PSGI application's server, with gperl psgi build's options after --.
	app := filepath.Join(dir, "app.psgi")
	if err := os.WriteFile(app, []byte("sub { [200, [], ['ok']] };\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tarball = filepath.Join(t.TempDir(), "server.tar")
	if err := gperl.ImageBuild([]string{"-base", base, "-platform", "linux/amd64", "-tarball", tarball, "-psgi", app, "--", "--workers", "2"}); err != nil {
		t.Fatal(err)
	}
	files = readTar(t, tarball)
	if err := json.Unmarshal(files["manifest.json"], &entries); err != nil || len(entries) != 1 || len(entries[0].RepoTags) != 0 {
		t.Fatalf("manifest.json %s (%v)", files["manifest.json"], err)
	}
	wantBinary(t, files[entries[0].Layers[1]], "app/app", elf.EM_X86_64)
}

// wantBinary checks that the gzipped tar layer holds an ELF executable for
// machine at name.
func wantBinary(t *testing.T, layer []byte, name string, machine elf.Machine) {
	t.Helper()
	zr, err := gzip.NewReader(bytes.NewReader(layer))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(zr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			t.Fatalf("the layer has no %s", name)
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Name != name {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		f, err := elf.NewFile(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if f.Machine != machine || h.Mode != 0o755 {
			t.Errorf("%s: machine %v mode %o, want %v 755", name, f.Machine, h.Mode, machine)
		}
		return
	}
}

// readTar reads the tar archive at path into its files' contents.
func readTar(t *testing.T, path string) map[string][]byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	files := map[string][]byte{}
	tr := tar.NewReader(f)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return files
		}
		if err != nil {
			t.Fatal(err)
		}
		if files[h.Name], err = io.ReadAll(tr); err != nil {
			t.Fatal(err)
		}
	}
}
