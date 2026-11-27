package oci

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"
	"time"
)

// The media types of the OCI image specification and their Docker
// equivalents.
const (
	mediaOCIIndex       = "application/vnd.oci.image.index.v1+json"
	mediaOCIManifest    = "application/vnd.oci.image.manifest.v1+json"
	mediaOCILayer       = "application/vnd.oci.image.layer.v1.tar+gzip"
	mediaDockerList     = "application/vnd.docker.distribution.manifest.list.v2+json"
	mediaDockerManifest = "application/vnd.docker.distribution.manifest.v2+json"
	mediaDockerLayer    = "application/vnd.docker.image.rootfs.diff.tar.gzip"
)

// Descriptor names a blob or a manifest by its media type, digest and size.
type Descriptor struct {
	MediaType string    `json:"mediaType"`
	Digest    string    `json:"digest"`
	Size      int64     `json:"size"`
	Platform  *Platform `json:"platform,omitempty"`
}

// Platform is what an image runs on.
type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Variant      string `json:"variant,omitempty"`
}

func (p Platform) String() string {
	s := p.OS + "/" + p.Architecture
	if p.Variant != "" {
		s += "/" + p.Variant
	}
	return s
}

// matches reports whether an image for got runs where p asks: the same
// operating system and architecture.
func (p Platform) matches(got Platform) bool {
	return got.OS == p.OS && got.Architecture == p.Architecture
}

type manifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Config        Descriptor   `json:"config"`
	Layers        []Descriptor `json:"layers"`
}

type index struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Manifests     []Descriptor `json:"manifests"`
}

// Image is an image for one platform: its manifest's media type, its
// config, and its layers.
type Image struct {
	Platform  Platform
	mediaType string
	config    Descriptor
	configRaw []byte
	layers    []layer
}

// layer is a layer of an image and where its content is: in data, or in
// the repository it was pulled from.
type layer struct {
	desc Descriptor
	data []byte
	from Reference
}

// Pull fetches the image ref names for platform p: the manifest ref
// names, or the one an index ref names lists for p. Its layers are read
// from ref when the image is pushed or written.
func (c *Client) Pull(ctx context.Context, ref Reference, p Platform) (*Image, error) {
	body, mediaType, err := c.manifest(ctx, ref, ref.identifier())
	if err != nil {
		return nil, err
	}
	if mediaType == mediaOCIIndex || mediaType == mediaDockerList {
		var idx index
		if err := json.Unmarshal(body, &idx); err != nil {
			return nil, fmt.Errorf("%s: %w", ref, err)
		}
		i := slices.IndexFunc(idx.Manifests, func(d Descriptor) bool { return d.Platform != nil && p.matches(*d.Platform) })
		if i < 0 {
			var have []string
			for _, d := range idx.Manifests {
				if d.Platform != nil {
					have = append(have, d.Platform.String())
				}
			}
			return nil, fmt.Errorf("%s has no image for %s (it has %s)", ref, p, strings.Join(have, ", "))
		}
		// Everything fetched is checked against its digest, which the
		// client checks only in the form it computes.
		if !digestRE.MatchString(idx.Manifests[i].Digest) {
			return nil, fmt.Errorf("%s: unsupported digest %q", ref, idx.Manifests[i].Digest)
		}
		if body, mediaType, err = c.manifest(ctx, ref, idx.Manifests[i].Digest); err != nil {
			return nil, err
		}
	}
	if mediaType != mediaOCIManifest && mediaType != mediaDockerManifest {
		return nil, fmt.Errorf("%s: unsupported manifest type %q", ref, mediaType)
	}
	var m manifest
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", ref, err)
	}
	for _, d := range append([]Descriptor{m.Config}, m.Layers...) {
		if !digestRE.MatchString(d.Digest) {
			return nil, fmt.Errorf("%s: unsupported digest %q", ref, d.Digest)
		}
	}
	rc, err := c.blob(ctx, ref, m.Config)
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(rc)
	if err = errors.Join(err, rc.Close()); err != nil {
		return nil, fmt.Errorf("%s: config: %w", ref, err)
	}
	var got Platform
	if err := json.Unmarshal(raw, &got); err != nil {
		return nil, fmt.Errorf("%s: config: %w", ref, err)
	}
	if !p.matches(got) {
		return nil, fmt.Errorf("%s is an image for %s, not %s", ref, got, p)
	}
	img := &Image{Platform: got, mediaType: mediaType, config: m.Config, configRaw: raw}
	for _, d := range m.Layers {
		img.layers = append(img.layers, layer{desc: d, from: ref})
	}
	return img, nil
}

// epoch is the time an appended layer and its image are dated: a fixed
// one, so that the same input makes the same image.
const epoch = "1970-01-01T00:00:00Z"

// Append returns base with a layer holding an executable file at file
// (an absolute path) with content, which is the image's entrypoint.
func Append(base *Image, file string, content []byte) (*Image, error) {
	if !path.IsAbs(file) {
		return nil, fmt.Errorf("the entrypoint %q is not an absolute path", file)
	}
	tarball, gz, err := entrypointLayer(file, content)
	if err != nil {
		return nil, err
	}
	layerType := mediaOCILayer
	if base.mediaType == mediaDockerManifest {
		layerType = mediaDockerLayer
	}
	desc := Descriptor{MediaType: layerType, Digest: digestOf(gz), Size: int64(len(gz))}

	// Only what an entrypoint changes is rewritten; the rest of the config
	// is kept as the base has it.
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(base.configRaw, &cfg); err != nil {
		return nil, fmt.Errorf("base image config: %w", err)
	}
	run := map[string]json.RawMessage{}
	if raw, ok := cfg["config"]; ok {
		if err := json.Unmarshal(raw, &run); err != nil {
			return nil, fmt.Errorf("base image config: %w", err)
		}
		if run == nil { // "config": null
			run = map[string]json.RawMessage{}
		}
	}
	var rootfs struct {
		Type    string   `json:"type"`
		DiffIDs []string `json:"diff_ids"`
	}
	if err := json.Unmarshal(cfg["rootfs"], &rootfs); err != nil {
		return nil, fmt.Errorf("base image config: rootfs: %w", err)
	}
	rootfs.DiffIDs = append(rootfs.DiffIDs, digestOf(tarball))
	var history []json.RawMessage
	if raw, ok := cfg["history"]; ok {
		if err := json.Unmarshal(raw, &history); err != nil {
			return nil, fmt.Errorf("base image config: history: %w", err)
		}
	}
	entry, err := json.Marshal(map[string]string{"created": epoch, "created_by": "gperl image build", "comment": file})
	if err != nil {
		return nil, err
	}
	history = append(history, entry)
	for key, value := range map[string]any{"Entrypoint": []string{file}, "Cmd": nil} {
		if run[key], err = json.Marshal(value); err != nil {
			return nil, err
		}
	}
	for key, value := range map[string]any{"config": run, "rootfs": rootfs, "history": history, "created": epoch} {
		if cfg[key], err = json.Marshal(value); err != nil {
			return nil, err
		}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	img := &Image{
		Platform:  base.Platform,
		mediaType: base.mediaType,
		config:    Descriptor{MediaType: base.config.MediaType, Digest: digestOf(raw), Size: int64(len(raw))},
		configRaw: raw,
		layers:    append(slices.Clone(base.layers), layer{desc: desc, data: gz}),
	}
	return img, nil
}

// entrypointLayer returns a layer holding an executable file at file with
// content, uncompressed and compressed. Its entries are owned by root and
// dated epoch.
func entrypointLayer(file string, content []byte) (tarball, gz []byte, err error) {
	var tb bytes.Buffer
	tw := tar.NewWriter(&tb)
	var dirs []string
	for d := path.Dir(file); d != "/"; d = path.Dir(d) {
		dirs = append(dirs, d)
	}
	slices.Reverse(dirs)
	for _, d := range dirs {
		if err := tw.WriteHeader(&tar.Header{Typeflag: tar.TypeDir, Name: d[1:] + "/", Mode: 0o755, ModTime: time.Unix(0, 0)}); err != nil {
			return nil, nil, err
		}
	}
	if err := tw.WriteHeader(&tar.Header{Typeflag: tar.TypeReg, Name: file[1:], Mode: 0o755, Size: int64(len(content)), ModTime: time.Unix(0, 0)}); err != nil {
		return nil, nil, err
	}
	if _, err := tw.Write(content); err != nil {
		return nil, nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, nil, err
	}
	var zb bytes.Buffer
	zw := gzip.NewWriter(&zb)
	if _, err := zw.Write(tb.Bytes()); err != nil {
		return nil, nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, nil, err
	}
	return tb.Bytes(), zb.Bytes(), nil
}

// manifest is img's manifest and its descriptor.
func (img *Image) manifest() ([]byte, Descriptor, error) {
	m := manifest{SchemaVersion: 2, MediaType: img.mediaType, Config: img.config}
	for _, l := range img.layers {
		m.Layers = append(m.Layers, l.desc)
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, Descriptor{}, err
	}
	p := img.Platform
	return raw, Descriptor{MediaType: img.mediaType, Digest: digestOf(raw), Size: int64(len(raw)), Platform: &p}, nil
}

// open opens l's content.
func (c *Client) open(ctx context.Context, l layer) (io.ReadCloser, error) {
	if l.data != nil {
		return io.NopCloser(bytes.NewReader(l.data)), nil
	}
	return c.blob(ctx, l.from, l.desc)
}

// Push uploads imgs to ref, which must have a tag: the image, or an index
// of the images when there are several. It returns the digest ref's tag
// then names.
func (c *Client) Push(ctx context.Context, ref Reference, imgs []*Image) (string, error) {
	if ref.Tag == "" || ref.Digest != "" {
		return "", fmt.Errorf("%s: an image is pushed to a tag (repository:tag)", ref)
	}
	if len(imgs) == 0 {
		return "", fmt.Errorf("%s: no image to push", ref)
	}
	// An index lists manifests of its own format.
	for _, img := range imgs[1:] {
		if img.mediaType != imgs[0].mediaType {
			return "", fmt.Errorf("%s: the images for %s and %s are in different manifest formats (%s, %s): use bases of one format", ref, imgs[0].Platform, img.Platform, imgs[0].mediaType, img.mediaType)
		}
	}
	var descs []Descriptor
	for _, img := range imgs {
		blobs := append(slices.Clone(img.layers), layer{desc: img.config, data: img.configRaw})
		for _, l := range blobs {
			has, err := c.hasBlob(ctx, ref, l.desc.Digest)
			if err != nil {
				return "", err
			}
			if has {
				continue
			}
			if err := c.putBlob(ctx, ref, l.desc, func() (io.ReadCloser, error) { return c.open(ctx, l) }); err != nil {
				return "", err
			}
		}
		raw, desc, err := img.manifest()
		if err != nil {
			return "", err
		}
		if len(imgs) == 1 {
			return desc.Digest, c.putManifest(ctx, ref, ref.Tag, desc.MediaType, raw)
		}
		if err := c.putManifest(ctx, ref, desc.Digest, desc.MediaType, raw); err != nil {
			return "", err
		}
		descs = append(descs, desc)
	}
	idx := index{SchemaVersion: 2, MediaType: mediaOCIIndex, Manifests: descs}
	if descs[0].MediaType == mediaDockerManifest {
		idx.MediaType = mediaDockerList
	}
	raw, err := json.Marshal(idx)
	if err != nil {
		return "", err
	}
	return digestOf(raw), c.putManifest(ctx, ref, ref.Tag, idx.MediaType, raw)
}

// WriteTarball writes img as the archive `docker load` reads, tagged tag
// unless tag is "".
func (c *Client) WriteTarball(ctx context.Context, w io.Writer, img *Image, tag string) error {
	tw := tar.NewWriter(w)
	blobPath := func(digest string) string { return "blobs/sha256/" + strings.TrimPrefix(digest, "sha256:") }
	entry := struct {
		Config   string
		RepoTags []string
		Layers   []string
	}{Config: blobPath(img.config.Digest), RepoTags: []string{}}
	if tag != "" {
		entry.RepoTags = append(entry.RepoTags, tag)
	}
	write := func(name string, size int64, r io.Reader) error {
		if err := tw.WriteHeader(&tar.Header{Typeflag: tar.TypeReg, Name: name, Mode: 0o644, Size: size, ModTime: time.Unix(0, 0)}); err != nil {
			return err
		}
		_, err := io.Copy(tw, r)
		return err
	}
	if err := write(entry.Config, int64(len(img.configRaw)), bytes.NewReader(img.configRaw)); err != nil {
		return err
	}
	for _, l := range img.layers {
		rc, err := c.open(ctx, l)
		if err != nil {
			return err
		}
		if err := errors.Join(write(blobPath(l.desc.Digest), l.desc.Size, rc), rc.Close()); err != nil {
			return err
		}
		entry.Layers = append(entry.Layers, blobPath(l.desc.Digest))
	}
	raw, err := json.Marshal([]any{entry})
	if err != nil {
		return err
	}
	if err := write("manifest.json", int64(len(raw)), bytes.NewReader(raw)); err != nil {
		return err
	}
	return tw.Close()
}
