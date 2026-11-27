package oci

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Client speaks the distribution API to registries.
type Client struct {
	// HTTP sends the requests; http.DefaultClient when nil.
	HTTP *http.Client
	// Credential returns the credentials for a registry, ok false for
	// none; DockerCredential when nil.
	Credential func(registry string) (cred Credential, ok bool, err error)

	// authorization is the Authorization header that last satisfied each
	// registry's challenge.
	authorization map[string]string
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

// do sends a request to ref's registry, answering its authentication
// challenge: a 401 is followed by one more try with the credentials or
// token the challenge asks for (for push, a token to pull and push ref's
// repository, as docker asks). body makes the request's body afresh for
// each try (nil for none). A status other than the wanted ones is an
// error. The authorization goes only to the registry itself: an upload
// location elsewhere (a storage service's signed URL) gets none.
func (c *Client) do(ctx context.Context, ref Reference, push bool, method, u string, header http.Header, body func() (io.Reader, int64, error), want ...int) (*http.Response, error) {
	api, err := url.Parse(ref.apiBase())
	if err != nil {
		return nil, err
	}
	for try := 0; ; try++ {
		var r io.Reader
		size := int64(0)
		if body != nil {
			var err error
			if r, size, err = body(); err != nil {
				return nil, err
			}
		}
		req, err := http.NewRequestWithContext(ctx, method, u, r)
		if err != nil {
			return nil, err
		}
		req.ContentLength = size
		for k, v := range header {
			req.Header[k] = v
		}
		toRegistry := sameOrigin(req.URL, api)
		if a := c.authorization[ref.Name()]; a != "" && toRegistry {
			req.Header.Set("Authorization", a)
		}
		resp, err := c.httpClient().Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusUnauthorized && try == 0 && toRegistry {
			h := resp.Header.Get("WWW-Authenticate")
			if err := resp.Body.Close(); err != nil {
				return nil, err
			}
			if err := c.authorize(ctx, ref, push, h); err != nil {
				return nil, fmt.Errorf("%s %s: %w", method, u, err)
			}
			continue
		}
		for _, w := range want {
			if resp.StatusCode == w {
				return resp, nil
			}
		}
		return nil, fmt.Errorf("%s %s: %w", method, u, statusError(resp))
	}
}

// sameOrigin reports whether a and b are the same scheme, host and port:
// host names compare without case, and an omitted port is the scheme's.
func sameOrigin(a, b *url.URL) bool {
	port := func(u *url.URL) string {
		if p := u.Port(); p != "" {
			return p
		}
		return map[string]string{"http": "80", "https": "443"}[u.Scheme]
	}
	return a.Scheme == b.Scheme && strings.EqualFold(a.Hostname(), b.Hostname()) && port(a) == port(b)
}

// statusError reports resp's unwanted status with the start of its body,
// the registry's explanation.
func statusError(resp *http.Response) error {
	msg, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err = errors.Join(err, resp.Body.Close()); err != nil {
		return fmt.Errorf("%s (reading the reply: %w)", resp.Status, err)
	}
	return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(msg)))
}

// authorize answers ref's registry's challenge h: Basic with the
// credentials, Bearer with a token from the challenge's realm — for push,
// one to pull and push ref's repository. Credentials go to a realm only
// over HTTPS, or plain HTTP on the loopback interface.
func (c *Client) authorize(ctx context.Context, ref Reference, push bool, h string) (err error) {
	ch, err := parseChallenge(h)
	if err != nil {
		return err
	}
	credential := c.Credential
	if credential == nil {
		credential = DockerCredential
	}
	cred, ok, err := credential(ref.Registry)
	if err != nil {
		return err
	}
	if c.authorization == nil {
		c.authorization = map[string]string{}
	}
	switch ch.scheme {
	case "basic":
		if !ok {
			return fmt.Errorf("%s asks for credentials and docker has none for it (docker login %s)", ref.Registry, ref.Registry)
		}
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(cred.Username, cred.Password)
		c.authorization[ref.Name()] = req.Header.Get("Authorization")
		return nil
	case "bearer":
		realm := ch.params["realm"]
		if realm == "" {
			return fmt.Errorf("WWW-Authenticate %q names no realm", h)
		}
		u, err := url.Parse(realm)
		if err != nil {
			return fmt.Errorf("WWW-Authenticate %q: %w", h, err)
		}
		q := u.Query()
		for _, k := range []string{"service", "scope"} {
			if v := ch.params[k]; v != "" {
				q.Set(k, v)
			}
		}
		if push {
			q.Set("scope", "repository:"+ref.Repository+":pull,push")
		}
		u.RawQuery = q.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}
		if ok {
			if u.Scheme != "https" && !(u.Scheme == "http" && isLoopback(u.Host)) {
				return fmt.Errorf("the token realm %s is not HTTPS: the credentials are not sent to it", realm)
			}
			req.SetBasicAuth(cred.Username, cred.Password)
		}
		resp, err := c.httpClient().Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("token from %s: %w", realm, statusError(resp))
		}
		defer func() { err = errors.Join(err, resp.Body.Close()) }()
		var tok struct {
			Token       string `json:"token"`
			AccessToken string `json:"access_token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
			return fmt.Errorf("token from %s: %w", realm, err)
		}
		if tok.Token == "" {
			tok.Token = tok.AccessToken
		}
		if tok.Token == "" {
			return fmt.Errorf("token from %s: the reply has no token", realm)
		}
		c.authorization[ref.Name()] = "Bearer " + tok.Token
		return nil
	default:
		return fmt.Errorf("WWW-Authenticate %q: unsupported scheme %q", h, ch.scheme)
	}
}

// manifestTypes are the manifest formats the client reads.
var manifestTypes = []string{mediaOCIIndex, mediaDockerList, mediaOCIManifest, mediaDockerManifest}

// manifest fetches the manifest ref names (by id, a tag or a digest) and
// its media type. A manifest fetched by digest is checked against it.
func (c *Client) manifest(ctx context.Context, ref Reference, id string) ([]byte, string, error) {
	resp, err := c.do(ctx, ref, false, http.MethodGet, ref.apiBase()+"/manifests/"+id,
		http.Header{"Accept": {strings.Join(manifestTypes, ", ")}}, nil, http.StatusOK)
	if err != nil {
		return nil, "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err = errors.Join(err, resp.Body.Close()); err != nil {
		return nil, "", err
	}
	if digestRE.MatchString(id) && digestOf(body) != id {
		return nil, "", fmt.Errorf("%s@%s: the registry sent a manifest of digest %s", ref.Name(), id, digestOf(body))
	}
	// The media type the manifest declares is what it is; the header is
	// the registry's report of it.
	var head struct {
		MediaType string `json:"mediaType"`
	}
	if err := json.Unmarshal(body, &head); err != nil {
		return nil, "", fmt.Errorf("%s@%s: %w", ref.Name(), id, err)
	}
	if head.MediaType == "" {
		head.MediaType, _, _ = strings.Cut(resp.Header.Get("Content-Type"), ";")
	}
	return body, head.MediaType, nil
}

// blob opens the blob desc names in ref's repository; reading it to the
// end fails if its content is not desc's.
func (c *Client) blob(ctx context.Context, ref Reference, desc Descriptor) (io.ReadCloser, error) {
	resp, err := c.do(ctx, ref, false, http.MethodGet, ref.apiBase()+"/blobs/"+desc.Digest, nil, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &verifyingReader{r: resp.Body, h: sha256.New(), want: desc}, nil
}

// hasBlob reports whether ref's repository, which is being pushed to, has
// the blob digest names.
func (c *Client) hasBlob(ctx context.Context, ref Reference, digest string) (bool, error) {
	resp, err := c.do(ctx, ref, true, http.MethodHead, ref.apiBase()+"/blobs/"+digest, nil, nil, http.StatusOK, http.StatusNotFound)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, resp.Body.Close()
}

// putBlob uploads the blob desc names, whose content open gives, to ref's
// repository in one request (the monolithic upload).
func (c *Client) putBlob(ctx context.Context, ref Reference, desc Descriptor, open func() (io.ReadCloser, error)) (err error) {
	resp, err := c.do(ctx, ref, true, http.MethodPost, ref.apiBase()+"/blobs/uploads/", nil, nil, http.StatusAccepted)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}
	loc, err := resp.Request.URL.Parse(resp.Header.Get("Location"))
	if err != nil {
		return fmt.Errorf("upload of %s: Location: %w", desc.Digest, err)
	}
	q := loc.Query()
	q.Set("digest", desc.Digest)
	loc.RawQuery = q.Encode()
	// The content is opened afresh for each try; the last one opened is
	// closed when the upload ends.
	var opened io.ReadCloser
	defer func() {
		if opened != nil {
			err = errors.Join(err, opened.Close())
		}
	}()
	body := func() (io.Reader, int64, error) {
		if opened != nil {
			if err := opened.Close(); err != nil {
				return nil, 0, err
			}
		}
		var err error
		opened, err = open()
		return opened, desc.Size, err
	}
	resp, err = c.do(ctx, ref, true, http.MethodPut, loc.String(),
		http.Header{"Content-Type": {"application/octet-stream"}}, body, http.StatusCreated)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

// putManifest uploads a manifest of media type mediaType to ref's
// repository under id (a tag or its digest).
func (c *Client) putManifest(ctx context.Context, ref Reference, id, mediaType string, manifest []byte) error {
	body := func() (io.Reader, int64, error) {
		return bytes.NewReader(manifest), int64(len(manifest)), nil
	}
	resp, err := c.do(ctx, ref, true, http.MethodPut, ref.apiBase()+"/manifests/"+id,
		http.Header{"Content-Type": {mediaType}}, body, http.StatusCreated)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

// digestOf is data's digest.
func digestOf(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// verifyingReader reads a blob and fails at its end unless the content
// had the size and digest it should.
type verifyingReader struct {
	r    io.ReadCloser
	h    hash.Hash
	n    int64
	want Descriptor
}

func (v *verifyingReader) Read(p []byte) (int, error) {
	n, err := v.r.Read(p)
	v.h.Write(p[:n])
	v.n += int64(n)
	if errors.Is(err, io.EOF) {
		if got := "sha256:" + hex.EncodeToString(v.h.Sum(nil)); got != v.want.Digest || v.n != v.want.Size {
			return n, fmt.Errorf("blob %s: got %s bytes of digest %s", v.want.Digest, strconv.FormatInt(v.n, 10), got)
		}
	}
	return n, err
}

func (v *verifyingReader) Close() error { return v.r.Close() }
