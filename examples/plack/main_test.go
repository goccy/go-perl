package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// newTestHandler boots a one-instance pool, skipping when the vendored
// module tree is absent (run `make setup` first).
func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	if _, err := os.Stat("local/lib/perl5"); err != nil {
		t.Skip("local/lib/perl5 not found - run `make setup` to vendor Plack and Mojolicious")
	}
	handler, cleanup, err := appHandler(1)
	if err != nil {
		t.Fatalf("boot PSGI pool: %v", err)
	}
	t.Cleanup(cleanup)
	return handler
}

func TestMojoliciousRoutes(t *testing.T) {
	handler := newTestHandler(t)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("text", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET / = %d, body %q", resp.StatusCode, body)
		}
		if want := "Hello from Mojolicious running on go-perl!\n"; string(body) != want {
			t.Fatalf("GET / body = %q, want %q", body, want)
		}
	})

	t.Run("json", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/json")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /json = %d, body %q", resp.StatusCode, body)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Fatalf("GET /json content-type = %q", ct)
		}
		for _, want := range []string{`"framework":"Mojolicious"`, `"runtime":"go-perl"`} {
			if !strings.Contains(string(body), want) {
				t.Fatalf("GET /json body = %q, want it to contain %q", body, want)
			}
		}
	})

	t.Run("echo", func(t *testing.T) {
		resp, err := http.Post(srv.URL+"/echo", "text/plain", strings.NewReader("shout this"))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST /echo = %d, body %q", resp.StatusCode, body)
		}
		if want := "SHOUT THIS"; string(body) != want {
			t.Fatalf("POST /echo body = %q, want %q", body, want)
		}
	})

	t.Run("not found", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/no/such/route")
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("GET /no/such/route = %d, want 404", resp.StatusCode)
		}
	})
}
