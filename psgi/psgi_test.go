package psgi_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/psgi"
)

// newServer loads app source from a scratch .psgi and wraps it in a test
// HTTP server.
func newServer(t *testing.T, workers int, appSrc string) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	appPath := filepath.Join(dir, "app.psgi")
	if err := os.WriteFile(appPath, []byte(appSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	srv, err := psgi.New(p, workers, appPath)
	if err != nil {
		p.Close()
		t.Fatalf("psgi.New: %v", err)
	}
	ts := httptest.NewServer(srv)
	t.Cleanup(func() {
		ts.Close()
		srv.Close()
	})
	return ts
}

// TestServeSimpleApp: a bare .psgi — its last value is the application, no
// framework required — serves status, headers, and body.
func TestServeSimpleApp(t *testing.T) {
	ts := newServer(t, 1, `
use strict;
sub {
    my ($env) = @_;
    [200,
     ['Content-Type' => 'text/plain', 'X-Powered-By' => 'go-perl'],
     ['hello ', 'from ', $env->{PATH_INFO}]];
};
`)
	res, err := http.Get(ts.URL + "/psgi")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if got := res.Header.Get("X-Powered-By"); got != "go-perl" {
		t.Fatalf("X-Powered-By = %q", got)
	}
	if string(body) != "hello from /psgi" {
		t.Fatalf("body = %q", body)
	}
}

// TestRequestBodyBytes: the request body reaches psgi.input as raw bytes
// (NULs and non-UTF-8 included) and the response carries raw bytes back —
// no serialisation sits in between.
func TestRequestBodyBytes(t *testing.T) {
	ts := newServer(t, 1, `
use strict;
sub {
    my ($env) = @_;
    my $in = $env->{'psgi.input'};
    local $/;
    my $data = <$in>;
    $data = '' unless defined $data;
    [200, ['Content-Type' => 'application/octet-stream'],
     [scalar reverse($data)]];
};
`)
	raw := []byte{0x00, 0x01, 0xFF, 'a', 'b', 0x80, 0x00}
	res, err := http.Post(ts.URL, "application/octet-stream", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	want := make([]byte, len(raw))
	for i, b := range raw {
		want[len(raw)-1-i] = b
	}
	if !bytes.Equal(body, want) {
		t.Fatalf("body = %x, want %x", body, want)
	}
}

// TestFilehandleBody: a filehandle response body drains through the guest.
func TestFilehandleBody(t *testing.T) {
	ts := newServer(t, 1, `
use strict;
sub {
    my $data = "streamed-body";
    open my $fh, '<', \$data or die $!;
    [200, ['Content-Type' => 'text/plain'], $fh];
};
`)
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if string(body) != "streamed-body" {
		t.Fatalf("body = %q", body)
	}
}

// TestEnvKeys: the CGI-style keys and psgi.* keys reach the application.
func TestEnvKeys(t *testing.T) {
	ts := newServer(t, 1, `
use strict;
sub {
    my ($env) = @_;
    my $v = $env->{'psgi.version'};
    [200, ['Content-Type' => 'text/plain'],
     [join '|',
      $env->{REQUEST_METHOD},
      $env->{QUERY_STRING},
      $env->{'psgi.url_scheme'},
      "$v->[0].$v->[1]",
      $env->{HTTP_X_PROBE}]];
};
`)
	req, _ := http.NewRequest("GET", ts.URL+"/x?a=1&b=2", nil)
	req.Header.Set("X-Probe", "probed")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if string(body) != "GET|a=1&b=2|http|1.1|probed" {
		t.Fatalf("env view = %q", body)
	}
}

// TestWorkersServeConcurrently: multiple cloned workers answer in parallel,
// each with its own interpreter state.
func TestWorkersServeConcurrently(t *testing.T) {
	ts := newServer(t, 3, `
use strict;
my $served = 0;
sub {
    $served++;
    [200, ['Content-Type' => 'text/plain'], ["ok:$served"]];
};
`)
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := http.Get(ts.URL)
			if err != nil {
				errs <- err
				return
			}
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			if res.StatusCode != 200 || len(body) < 4 || string(body[:3]) != "ok:" {
				errs <- fmt.Errorf("response = %d %q", res.StatusCode, body)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

// TestNonCodePSGIFails: a file whose last value is not a code reference is
// rejected at New, before anything serves.
func TestNonCodePSGIFails(t *testing.T) {
	dir := t.TempDir()
	appPath := filepath.Join(dir, "bad.psgi")
	if err := os.WriteFile(appPath, []byte(`"just a string";`), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	if _, err := psgi.New(p, 1, appPath); err == nil {
		t.Fatalf("expected a non-code .psgi to fail New")
	}
}
