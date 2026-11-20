package gperl

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func envOf(pairs map[string]string) func(string) string {
	return func(key string) string { return pairs[key] }
}

// A setting comes from the command line, else the environment, else the
// defaults a binary was built with.
func TestParsePSGIPrecedence(t *testing.T) {
	defaults := []string{"--workers=8", "--port=7000", "--listen=:7001", "--listen=:7002", "--env=staging", "-I=buildlib"}

	c, rest, err := parsePSGI("server", defaults, nil, envOf(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.workers != 8 || c.port != 7000 || c.env != "staging" || len(rest) != 0 ||
		!reflect.DeepEqual(c.listen.values, []string{":7001", ":7002"}) || !reflect.DeepEqual(c.inc.values, []string{"buildlib"}) {
		t.Fatalf("defaults only: %+v rest=%v", c, rest)
	}

	c, _, err = parsePSGI("server", defaults, nil, envOf(map[string]string{
		"PORT": "8000", "GPERL_PSGI_WORKERS": "2", "GPERL_PSGI_LISTEN": ":8001,/tmp/app.sock",
		"GPERL_PSGI_MAX_REQUESTS": "500", "GPERL_PSGI_ENABLE_SSL": "true",
		"GPERL_PSGI_SSL_CERT": "c.pem", "GPERL_PSGI_SSL_KEY": "k.pem", "GPERL_PSGI_I": "envlib,more",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.port != 8000 || c.workers != 2 || c.maxRequests != 500 || !c.enableSSL || c.sslCert != "c.pem" || c.env != "staging" ||
		!reflect.DeepEqual(c.listen.values, []string{":8001", "/tmp/app.sock"}) || !reflect.DeepEqual(c.inc.values, []string{"envlib", "more"}) {
		t.Fatalf("environment over defaults: %+v", c)
	}

	// GPERL_PSGI_PORT beats PORT, and the command line beats both.
	env := envOf(map[string]string{"PORT": "8000", "GPERL_PSGI_PORT": "8100", "GPERL_PSGI_WORKERS": "2"})
	if c, _, err = parsePSGI("server", defaults, nil, env); err != nil || c.port != 8100 {
		t.Fatalf("GPERL_PSGI_PORT over PORT: %+v, %v", c, err)
	}
	c, rest, err = parsePSGI("server", defaults, strings.Fields("-p9000 --workers 3 -l :9001 -Ilib -MFoo::Bar=a,b -E production -r app.psgi extra"), env)
	if err != nil {
		t.Fatal(err)
	}
	if c.port != 9000 || c.workers != 3 || c.env != "production" || !c.reload ||
		!reflect.DeepEqual(c.listen.values, []string{":9001"}) || !reflect.DeepEqual(c.inc.values, []string{"lib"}) ||
		!reflect.DeepEqual(c.modules.values, []string{"Foo::Bar=a,b"}) || !reflect.DeepEqual(rest, []string{"app.psgi", "extra"}) {
		t.Fatalf("command line over everything: %+v rest=%v", c, rest)
	}

	// Built-in defaults.
	if c, _, err = parsePSGI("server", nil, nil, envOf(nil)); err != nil || c.port != 5000 || c.env != "deployment" || c.workers < 1 || c.readTimeout != 60 {
		t.Fatalf("built-in defaults: %+v, %v", c, err)
	}
}

func TestParsePSGIRejectsBadSettings(t *testing.T) {
	for name, c := range map[string]struct {
		defaults, args []string
		env            map[string]string
	}{
		"unknown option":       {args: []string{"--no-such-option"}},
		"bad default":          {defaults: []string{"--workers=many"}},
		"bad PORT":             {env: map[string]string{"PORT": "http"}},
		"bad environment":      {env: map[string]string{"GPERL_PSGI_WORKERS": "many"}},
		"zero workers":         {args: []string{"--workers", "0"}},
		"ssl without the keys": {args: []string{"--enable-ssl"}},
	} {
		if _, _, err := parsePSGI("server", c.defaults, c.args, envOf(c.env)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestPSGIListeners(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "app.sock")
	c := &psgiConfig{}
	c.listen.values = []string{"127.0.0.1:0", sock}
	lns, err := c.listeners()
	if err != nil {
		t.Fatal(err)
	}
	if len(lns) != 2 || lns[0].Addr().Network() != "tcp" || lns[1].Addr().Network() != "unix" {
		t.Fatalf("listeners = %v", lns)
	}
	for _, ln := range lns {
		if err := ln.Close(); err != nil {
			t.Error(err)
		}
	}
	// A stale socket file from an earlier run does not block the next.
	if err := os.WriteFile(sock, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if lns, err = c.listeners(); err != nil {
		t.Fatalf("listen over a stale socket file: %v", err)
	}
	for _, ln := range lns {
		if err := ln.Close(); err != nil {
			t.Error(err)
		}
	}

	// host and port apply when nothing is listed, and a failure closes
	// what was already open.
	held, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := held.Close(); err != nil {
			t.Error(err)
		}
	}()
	busy := &psgiConfig{host: "127.0.0.1", port: held.Addr().(*net.TCPAddr).Port}
	if _, err := busy.listeners(); err == nil {
		t.Fatal("listened on a port in use")
	}
	busy.listen.values = []string{"127.0.0.1:0", held.Addr().String()}
	if _, err := busy.listeners(); err == nil {
		t.Fatal("listened on an address in use")
	}
}

func TestSourceStamp(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "app.psgi")
	lib := filepath.Join(dir, "lib")
	write := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(app, "sub { }")
	write(filepath.Join(lib, "My.pm"), "1;")
	write(filepath.Join(lib, ".git", "index"), "x")
	before := sourceStamp([]string{app, lib})
	if again := sourceStamp([]string{app, lib}); again != before {
		t.Fatalf("the stamp of unchanged sources changed: %q -> %q", before, again)
	}
	// A hidden directory is not part of the sources.
	write(filepath.Join(lib, ".git", "other"), "y")
	if now := sourceStamp([]string{app, lib}); now != before {
		t.Fatalf("a file in a hidden directory changed the stamp")
	}
	write(filepath.Join(lib, "New.pm"), "1;")
	added := sourceStamp([]string{app, lib})
	if added == before {
		t.Fatal("a new source file did not change the stamp")
	}
	if err := os.Remove(app); err != nil {
		t.Fatal(err)
	}
	if removed := sourceStamp([]string{app, lib}); removed == added {
		t.Fatal("a removed source file did not change the stamp")
	}
}

// What a request changes of its worker's working directory and environment
// stays in that worker: the other workers and the process keep theirs.
func TestPSGIWorkersAreIsolated(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	appPath := filepath.Join(dir, "app.psgi")
	if err := os.WriteFile(appPath, []byte(`use Cwd ();
sub {
    my $env = shift;
    if ($env->{PATH_INFO} eq '/change') {
        chdir 'sub' or die "chdir: $!";
        $ENV{GPERL_TEST_MARK} = 'set';
    }
    [200, [], [join('|', (Cwd::getcwd() =~ m{/sub$} ? 'sub' : 'top'), $ENV{GPERL_TEST_MARK} // 'unset')]];
};
`), 0o644); err != nil {
		t.Fatal(err)
	}
	c, _, err := parsePSGI("test", nil, []string{"--workers", "2"}, envOf(nil))
	if err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	server, err := (&psgiApp{path: appPath, dir: dir}).newServer(c, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	get := func(path string) string {
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec.Body.String()
	}
	// The two workers take requests in turn.
	for i, want := range []string{"sub|set", "top|unset", "sub|set", "top|unset"} {
		path := "/"
		if i == 0 {
			path = "/change"
		}
		if got := get(path); got != want {
			t.Errorf("request %d (%s) = %q, want %q", i+1, path, got, want)
		}
	}
	if now, err := os.Getwd(); err != nil || now != wd {
		t.Errorf("the process's working directory = %q (%v), want %q", now, err, wd)
	}
	if mark, set := os.LookupEnv("GPERL_TEST_MARK"); set {
		t.Errorf("the process's environment has GPERL_TEST_MARK=%q", mark)
	}
}
