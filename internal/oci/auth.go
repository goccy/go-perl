package oci

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Credential is a user name and password (or token) for a registry.
type Credential struct{ Username, Password string }

// dockerConfig is the part of docker's config.json that names credentials.
type dockerConfig struct {
	Auths map[string]struct {
		Auth          string `json:"auth"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		IdentityToken string `json:"identitytoken"`
	} `json:"auths"`
	CredsStore  string            `json:"credsStore"`
	CredHelpers map[string]string `json:"credHelpers"`
}

// credentialsNotFound is what a docker credential helper writes when it
// has no credentials for a server: the protocol's "not found" reply
// (errCredentialsNotFoundMessage in docker-credential-helpers).
const credentialsNotFound = "credentials not found in native keychain"

// DockerCredential returns the credentials `docker login` stored for
// registry, as docker finds them: the registry's credential helper, else
// the credential store, else config.json's own entry. ok is false when
// there are none, and the registry is then used anonymously.
func DockerCredential(registry string) (cred Credential, ok bool, err error) {
	dir := os.Getenv("DOCKER_CONFIG")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return cred, false, fmt.Errorf("locate docker's config.json (set DOCKER_CONFIG): %w", err)
		}
		dir = filepath.Join(home, ".docker")
	}
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return cred, false, nil
	}
	if err != nil {
		return cred, false, err
	}
	var config dockerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return cred, false, fmt.Errorf("%s: %w", path, err)
	}
	// Docker Hub's credentials are stored under its old index URL.
	server := registry
	if registry == dockerHub {
		server = "https://index.docker.io/v1/"
	}
	host := serverHost(server)
	if helper, found := config.CredHelpers[host]; found {
		return helperCredential(helper, server)
	}
	if config.CredsStore != "" {
		return helperCredential(config.CredsStore, server)
	}
	// Keys naming the same host are taken in order, so that the same
	// config.json always gives the same credentials.
	for _, key := range slices.Sorted(maps.Keys(config.Auths)) {
		entry := config.Auths[key]
		if serverHost(key) != host {
			continue
		}
		switch {
		case entry.IdentityToken != "":
			return cred, false, fmt.Errorf("%s: the credentials for %s are an identity token, which gperl does not support", path, key)
		case entry.Auth != "":
			raw, err := base64.StdEncoding.DecodeString(entry.Auth)
			if err != nil {
				return cred, false, fmt.Errorf("%s: the credentials for %s: %w", path, key, err)
			}
			user, pass, found := strings.Cut(string(raw), ":")
			if !found {
				return cred, false, fmt.Errorf("%s: the credentials for %s are not user:password", path, key)
			}
			return Credential{user, pass}, true, nil
		case entry.Username != "":
			return Credential{entry.Username, entry.Password}, true, nil
		}
	}
	return cred, false, nil
}

// serverHost is the host a config.json key names: keys are written as a
// host or as a URL ("https://index.docker.io/v1/").
func serverHost(key string) string {
	for _, scheme := range []string{"https://", "http://"} {
		key = strings.TrimPrefix(key, scheme)
	}
	host, _, _ := strings.Cut(key, "/")
	if host == "registry-1.docker.io" || host == dockerHub {
		host = "index.docker.io"
	}
	return host
}

// helperCredential asks the docker credential helper
// docker-credential-<helper> for server's credentials.
func helperCredential(helper, server string) (Credential, bool, error) {
	cmd := exec.Command("docker-credential-"+helper, "get")
	cmd.Stdin = strings.NewReader(server)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && strings.TrimSpace(stdout.String()) == credentialsNotFound {
			return Credential{}, false, nil
		}
		return Credential{}, false, fmt.Errorf("docker-credential-%s get %s: %w: %s%s", helper, server, err, stdout.String(), stderr.String())
	}
	var reply struct{ Username, Secret string }
	if err := json.Unmarshal(stdout.Bytes(), &reply); err != nil {
		return Credential{}, false, fmt.Errorf("docker-credential-%s get %s: %w", helper, server, err)
	}
	// "<token>" marks an identity token (an OAuth2 refresh token).
	if reply.Username == "<token>" {
		return Credential{}, false, fmt.Errorf("docker-credential-%s gave an identity token for %s, which gperl does not support", helper, server)
	}
	return Credential{reply.Username, reply.Secret}, true, nil
}

// challenge is a WWW-Authenticate header: a scheme and its parameters.
type challenge struct {
	scheme string
	params map[string]string
}

// parseChallenge parses a WWW-Authenticate header with one challenge,
// scheme key="value",... (RFC 7235); a value may hold commas.
func parseChallenge(h string) (challenge, error) {
	scheme, rest, _ := strings.Cut(strings.TrimSpace(h), " ")
	c := challenge{scheme: strings.ToLower(scheme), params: map[string]string{}}
	for rest = strings.TrimSpace(rest); rest != ""; {
		key, after, found := strings.Cut(rest, "=")
		if !found {
			return c, fmt.Errorf("WWW-Authenticate %q: a parameter without a value", h)
		}
		key = strings.ToLower(strings.TrimSpace(key))
		var value string
		if strings.HasPrefix(after, `"`) {
			var b strings.Builder
			i := 1
			for ; i < len(after) && after[i] != '"'; i++ {
				if after[i] == '\\' && i+1 < len(after) {
					i++
				}
				b.WriteByte(after[i])
			}
			if i == len(after) {
				return c, fmt.Errorf("WWW-Authenticate %q: an unterminated quoted value", h)
			}
			value, after = b.String(), after[i+1:]
		} else {
			value, after, _ = strings.Cut(after, ",")
			after = "," + after
		}
		c.params[key] = strings.TrimSpace(value)
		after = strings.TrimSpace(after)
		if after != "" && !strings.HasPrefix(after, ",") {
			return c, fmt.Errorf("WWW-Authenticate %q: parameters not separated by commas", h)
		}
		rest = strings.TrimSpace(strings.TrimPrefix(after, ","))
	}
	return c, nil
}
