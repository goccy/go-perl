package psgi_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal/psgi"
)

// newServerWith is newServer with Options.
func newServerWith(t *testing.T, workers int, opts psgi.Options, appSrc string) *httptest.Server {
	t.Helper()
	appPath := filepath.Join(t.TempDir(), "app.psgi")
	if err := os.WriteFile(appPath, []byte(appSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	srv, err := psgi.NewServer(p, workers, appPath, opts)
	if err != nil {
		if cerr := p.Close(); cerr != nil {
			t.Errorf("close: %v", cerr)
		}
		t.Fatalf("psgi.NewServer: %v", err)
	}
	ts := httptest.NewServer(srv)
	t.Cleanup(func() {
		ts.Close()
		srv.Close()
	})
	return ts
}

func get(t *testing.T, url string) (int, http.Header, string) {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := res.Body.Close(); err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, res.Header, string(body)
}

// A delayed response: the application returns a code reference and hands
// the complete response to the responder.
func TestDelayedResponse(t *testing.T) {
	ts := newServerWith(t, 1, psgi.Options{}, `
sub {
    my ($env) = @_;
    die "server does not announce streaming\n" unless $env->{'psgi.streaming'};
    return sub {
        my ($respond) = @_;
        $respond->([201, ['Content-Type' => 'text/plain', 'X-Mode' => 'delayed'], ['later ', 'body']]);
    };
};
`)
	status, header, body := get(t, ts.URL+"/")
	if status != 201 || header.Get("X-Mode") != "delayed" || body != "later body" {
		t.Fatalf("delayed response = %d %q %q", status, header.Get("X-Mode"), body)
	}
}

// A streaming response: the responder returns a writer, and every write
// reaches the client before the application finishes.
func TestStreamingResponse(t *testing.T) {
	ts := newServerWith(t, 1, psgi.Options{}, `
sub {
    return sub {
        my ($respond) = @_;
        my $w = $respond->([200, ['Content-Type' => 'text/event-stream']]);
        $w->write("first\n");
        select(undef, undef, undef, 0.4);
        $w->write("second\n");
        $w->close;
    };
};
`)
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Error(err)
		}
	}()
	if ct := res.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("Content-Type = %q", ct)
	}
	r := bufio.NewReader(res.Body)
	start := time.Now()
	first, err := r.ReadString('\n')
	if err != nil || first != "first\n" {
		t.Fatalf("first chunk = %q, %v", first, err)
	}
	firstAt := time.Since(start)
	second, err := r.ReadString('\n')
	if err != nil || second != "second\n" {
		t.Fatalf("second chunk = %q, %v", second, err)
	}
	if gap := time.Since(start) - firstAt; gap < 200*time.Millisecond {
		t.Fatalf("the first chunk arrived only %v before the second: the response was buffered", gap)
	}
}

// A complete response is written after its worker went back to the pool: a
// client that does not read its response does not block the next request.
func TestSlowClientDoesNotHoldTheWorker(t *testing.T) {
	ts := newServerWith(t, 1, psgi.Options{}, `
my $big = 'x' x (8 * 1024 * 1024);
sub {
    my ($env) = @_;
    return [200, ['Content-Type' => 'text/plain'], [$env->{PATH_INFO} eq '/big' ? $big : 'small']];
};
`)
	// Ask for a response far larger than the socket buffers and never
	// read it.
	conn, err := net.Dial("tcp", strings.TrimPrefix(ts.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	}()
	if _, err := fmt.Fprint(conn, "GET /big HTTP/1.1\r\nHost: x\r\n\r\n"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	done := make(chan string, 1)
	go func() {
		_, _, body := get(t, ts.URL+"/small")
		done <- body
	}()
	select {
	case body := <-done:
		if body != "small" {
			t.Fatalf("body = %q", body)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the only worker is still held by a client that is not reading")
	}
}

// MaxRequests replaces a worker with a fresh clone of the loaded
// application: state a request left in the worker does not survive.
func TestMaxRequestsReplacesWorkers(t *testing.T) {
	app := `
my $count = 0;
sub { [200, ['Content-Type' => 'text/plain'], [++$count]] };
`
	recycled := newServerWith(t, 1, psgi.Options{MaxRequests: 2}, app)
	var got []string
	for i := 0; i < 5; i++ {
		_, _, body := get(t, recycled.URL+"/")
		got = append(got, body)
	}
	if want := "1 2 1 2 1"; strings.Join(got, " ") != want {
		t.Fatalf("with MaxRequests=2 the counter went %v, want %s", got, want)
	}

	kept := newServerWith(t, 1, psgi.Options{}, app)
	for i := 1; i <= 3; i++ {
		if _, _, body := get(t, kept.URL+"/"); body != fmt.Sprint(i) {
			t.Fatalf("request %d without MaxRequests = %q", i, body)
		}
	}
}

// syncBuffer is a bytes.Buffer safe to write from the server and read from
// the test.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func TestAccessLogAndBodyLimit(t *testing.T) {
	var log syncBuffer
	ts := newServerWith(t, 1, psgi.Options{AccessLog: &log, MaxRequestBody: 8}, `
sub {
    my ($env) = @_;
    read($env->{'psgi.input'}, my $body, 100);
    return [404, ['Content-Type' => 'text/plain'], ["got:$body"]] if $env->{PATH_INFO} eq '/missing';
    return [200, ['Content-Type' => 'text/plain'], ["got:$body"]];
};
`)
	post := func(path, body string) (int, string) {
		res, err := http.Post(ts.URL+path, "text/plain", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err := res.Body.Close(); err != nil {
			t.Fatal(err)
		}
		return res.StatusCode, string(data)
	}
	if status, body := post("/ok?q=1", "12345678"); status != 200 || body != "got:12345678" {
		t.Fatalf("body at the limit = %d %q", status, body)
	}
	if status, _ := post("/big", "123456789"); status != http.StatusRequestEntityTooLarge {
		t.Fatalf("body over the limit = %d", status)
	}
	if status, _ := post("/missing", ""); status != 404 {
		t.Fatalf("status = %d", status)
	}
	lines := strings.Split(strings.TrimSpace(log.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("access log has %d lines:\n%s", len(lines), log.String())
	}
	for i, want := range []string{`"POST /ok?q=1 HTTP/1.1" 200 12 `, `"POST /big HTTP/1.1" 413 `, `"POST /missing HTTP/1.1" 404 4 `} {
		if !strings.Contains(lines[i], want) || !strings.HasPrefix(lines[i], "127.0.0.1 - - [") {
			t.Errorf("access log line %d = %q, want it to contain %q", i, lines[i], want)
		}
	}
}

// What an application gets wrong becomes a 500, and the worker serves the
// next request.
func TestBrokenResponses(t *testing.T) {
	ts := newServerWith(t, 1, psgi.Options{}, `
sub {
    my ($env) = @_;
    my $p = $env->{PATH_INFO};
    return [200, ['Content-Type' => 'text/plain'], ['fine']] if $p eq '/ok';
    return 'not a response'                                 if $p eq '/scalar';
    return [200, ['Odd'], ['x']]                            if $p eq '/odd-headers';
    return [42, [], ['x']]                                  if $p eq '/bad-status';
    return sub { 1 }                                        if $p eq '/silent';
    die "boom\n";
};
`)
	for _, path := range []string{"/scalar", "/odd-headers", "/bad-status", "/silent", "/die"} {
		if status, _, _ := get(t, ts.URL+path); status != http.StatusInternalServerError {
			t.Errorf("%s: status = %d, want 500", path, status)
		}
		if status, _, body := get(t, ts.URL+"/ok"); status != 200 || body != "fine" {
			t.Errorf("after %s the worker answered %d %q", path, status, body)
		}
	}
}

func TestNewServerRejectsBadInput(t *testing.T) {
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := p.Close(); err != nil {
			t.Error(err)
		}
	}()
	if _, err := psgi.NewServer(p, 0, "app.psgi", psgi.Options{}); err == nil {
		t.Error("0 workers accepted")
	}
	if _, err := psgi.NewServer(p, 1, filepath.Join(t.TempDir(), "missing.psgi"), psgi.Options{}); err == nil {
		t.Error("a missing application file accepted")
	}
	bad := filepath.Join(t.TempDir(), "bad.psgi")
	if err := os.WriteFile(bad, []byte("this is not perl ("), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := psgi.NewServer(p, 1, bad, psgi.Options{}); err == nil {
		t.Error("an application that does not compile accepted")
	}
}
