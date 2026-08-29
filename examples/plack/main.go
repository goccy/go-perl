// Serve a PSGI Perl web application (Plack + Mojolicious) from Go's net/http.
//
// Architecture, borrowed from go-spidermonkey's HTTP examples: a fixed pool
// of warm Perl instances - each one a full interpreter with Plack and the
// PSGI app already compiled - and goroutine-per-request handling that checks
// an instance out of the pool, forwards the request over the Go<->Perl
// function bridge (perl.Call / psgi_handle), and writes the PSGI response
// back. Instances are cheap to boot thanks to go-perl's copy-on-write
// snapshot, and requests never share an interpreter concurrently.
package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	perl "github.com/goccy/go-perl"
)

//go:embed psgi.pl
var psgiAdapter string

// pool is the sql.DB-shaped instance pool: checkout on request, return on
// completion. Capacity == warm instances == max concurrency.
type pool struct {
	instances chan *perl.Perl
}

// bootInstance builds one Perl instance with the module search path, the
// PSGI adapter, and the application loaded. Paths cross via the function
// bridge (Bind/Call), so no Perl-source quoting is involved.
func bootInstance(incDirs []string, appPath string) (*perl.Perl, error) {
	p, err := perl.New(perl.Config{
		// The example loads its app and vendored modules from the host
		// working tree.
		HostFS: true,
		// The PSGI app's STDERR (psgi.errors) surfaces on the process stderr.
		Stderr: os.Stderr,
	})
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	fail := func(err error) (*perl.Perl, error) {
		p.Close()
		return nil, err
	}

	dirs := make([]any, len(incDirs))
	for i, d := range incDirs {
		dirs[i] = d
	}
	if err := p.Bind("go_inc_dirs", func([]any) ([]any, error) { return dirs, nil }); err != nil {
		return fail(fmt.Errorf("bind go_inc_dirs: %w", err))
	}
	if r, err := p.Eval(ctx, `unshift @INC, go_inc_dirs(); 1;`); err != nil {
		return fail(fmt.Errorf("extend @INC: %w", err))
	} else if !r.Ok {
		return fail(fmt.Errorf("extend @INC: %s", r.Error))
	}
	if r, err := p.Eval(ctx, psgiAdapter); err != nil {
		return fail(fmt.Errorf("load PSGI adapter: %w", err))
	} else if !r.Ok {
		return fail(fmt.Errorf("load PSGI adapter: %s", r.Error))
	}
	if _, err := p.Call(ctx, "psgi_load", appPath); err != nil {
		return fail(fmt.Errorf("load %s: %w", appPath, err))
	}
	return p, nil
}

func newPool(size int, incDirs []string, appPath string) (*pool, error) {
	pl := &pool{instances: make(chan *perl.Perl, size)}
	for i := 0; i < size; i++ {
		p, err := bootInstance(incDirs, appPath)
		if err != nil {
			pl.Close()
			return nil, err
		}
		pl.instances <- p
	}
	return pl, nil
}

func (pl *pool) Close() {
	close(pl.instances)
	for p := range pl.instances {
		p.Close()
	}
}

// ServeHTTP forwards one request into a pooled interpreter and writes the
// PSGI response. Cancelling the request context while the app runs stops the
// Perl code at the next opcode (perl.Call honours ctx).
func (pl *pool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var p *perl.Perl
	select {
	case p = <-pl.instances:
	case <-r.Context().Done():
		http.Error(w, "no interpreter available", http.StatusServiceUnavailable)
		return
	}
	defer func() { pl.instances <- p }()

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 32<<20))
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	res, err := p.Call(r.Context(), "psgi_handle", psgiEnv(r, len(body)),
		base64.StdEncoding.EncodeToString(body))
	if err != nil {
		log.Printf("psgi_handle %s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := writePSGIResponse(w, res); err != nil {
		log.Printf("write response %s %s: %v", r.Method, r.URL.Path, err)
	}
}

// psgiEnv flattens the request into the CGI-style PSGI environment. The
// psgi.* runtime keys are filled in guest-side by psgi.pl.
func psgiEnv(r *http.Request, bodyLen int) map[string]any {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		host, port = r.Host, "80"
	}
	remote := r.RemoteAddr
	if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		remote = h
	}
	env := map[string]any{
		"REQUEST_METHOD":  r.Method,
		"SCRIPT_NAME":     "",
		"PATH_INFO":       r.URL.Path,
		"REQUEST_URI":     r.URL.RequestURI(),
		"QUERY_STRING":    r.URL.RawQuery,
		"SERVER_NAME":     host,
		"SERVER_PORT":     port,
		"SERVER_PROTOCOL": r.Proto,
		"REMOTE_ADDR":     remote,
		"CONTENT_LENGTH":  fmt.Sprint(bodyLen),
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		env["CONTENT_TYPE"] = ct
	}
	if r.TLS != nil {
		env["psgi.url_scheme"] = "https"
	}
	for k, vs := range r.Header {
		switch k {
		case "Content-Type", "Content-Length":
			continue
		}
		env["HTTP_"+strings.ReplaceAll(strings.ToUpper(k), "-", "_")] = strings.Join(vs, ", ")
	}
	return env
}

// writePSGIResponse maps the (status, header pairs, base64 body) return list
// of psgi_handle onto the ResponseWriter.
func writePSGIResponse(w http.ResponseWriter, res []any) error {
	if len(res) != 3 {
		return fmt.Errorf("psgi_handle returned %d values, want 3", len(res))
	}
	status, ok := res[0].(float64)
	if !ok || status < 100 || status > 999 {
		return fmt.Errorf("bad status %#v", res[0])
	}
	// The PSGI header arrayref crosses as an identity-preserving handle;
	// headers are pure data here, so materialise and release it.
	hdrRef, ok := res[1].(*perl.Ref)
	if !ok {
		return fmt.Errorf("bad header list %#v", res[1])
	}
	defer hdrRef.Free()
	exported, err := hdrRef.Export(context.Background())
	if err != nil {
		return fmt.Errorf("export headers: %w", err)
	}
	pairs, ok := exported.([]any)
	if !ok || len(pairs)%2 != 0 {
		return fmt.Errorf("bad header list %#v", exported)
	}
	bodyB64, ok := res[2].(string)
	if !ok {
		return fmt.Errorf("bad body %#v", res[2])
	}
	body, err := base64.StdEncoding.DecodeString(bodyB64)
	if err != nil {
		return fmt.Errorf("decode body: %w", err)
	}
	for i := 0; i+1 < len(pairs); i += 2 {
		k, _ := pairs[i].(string)
		v, _ := pairs[i+1].(string)
		if k != "" {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(int(status))
	_, err = w.Write(body)
	return err
}

// appHandler assembles the pool for the app.psgi next to the binary's
// working directory, with the cpanm-vendored module tree on @INC.
func appHandler(size int) (http.Handler, func(), error) {
	appPath, err := filepath.Abs("app.psgi")
	if err != nil {
		return nil, nil, err
	}
	if _, err := os.Stat(appPath); err != nil {
		return nil, nil, fmt.Errorf("app.psgi not found (run from examples/plack): %w", err)
	}
	libDir, err := filepath.Abs(filepath.Join("local", "lib", "perl5"))
	if err != nil {
		return nil, nil, err
	}
	if _, err := os.Stat(libDir); err != nil {
		return nil, nil, fmt.Errorf("%s not found - run `make setup` to vendor Plack and Mojolicious: %w", libDir, err)
	}
	start := time.Now()
	pl, err := newPool(size, []string{libDir}, appPath)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("booted %d Perl instance(s) with Plack + Mojolicious in %s", size, time.Since(start).Round(time.Millisecond))
	return pl, pl.Close, nil
}

func main() {
	addr := flag.String("addr", ":8091", "listen address")
	size := flag.Int("pool", runtime.GOMAXPROCS(0), "warm Perl instances (max concurrency)")
	flag.Parse()

	handler, cleanup, err := appHandler(*size)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	log.Printf("serving the PSGI app on http://localhost%s", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal(err)
	}
}
