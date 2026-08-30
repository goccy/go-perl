// Package psgi serves PSGI Perl web applications from Go's net/http.
//
// PSGI specifies the application: a code reference that takes the
// environment hashref (CGI-style keys plus psgi.*) and returns
// [$status, \@headers, $body]. This package supplies the SERVER side of
// that interface — the environment (psgi.streaming is false: this server
// does not implement delayed/streaming responses), the request/response
// plumbing, and the worker model.
//
// It owns the two conversions every PSGI-on-go-perl server needs — an
// *http.Request into the CGI-style PSGI environment, and the app's
// (status, headers, body) response back onto an http.ResponseWriter — plus
// the Perl-side adapter (psgi.pl, embedded) that wraps the application with
// Plack::Util — and Server, a fixed set of warm interpreter workers (the
// starman/starlet serving model). Prepare ONE instance with the instance's
// own loading facilities, and New compiles the app into it and scales it:
// every other worker is a copy-on-write clone (perl.Clone), so the loading
// work runs once no matter the worker count:
//
//	stdlib, _ := perl.ExtractStdlib()
//	p, _ := perl.New(perl.Config{FS: fs.NewHostFS(), StdlibDir: stdlib})
//	p.AddInc(ctx, "local/lib/perl5")
//	p.AddXSDir("local/xs")
//	server, _ := psgi.New(p, 4, "app.psgi")
//	server.ListenAndServe(":8091") // or http.ListenAndServe(addr, server)
//
// A *perl.Perl must not run concurrent requests; Server guarantees that.
package psgi

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	perl "github.com/goccy/go-perl"
	// Serving stacks routinely carry native XS modules; linking the
	// loader here makes (*perl.Perl).AddXSDir work without application
	// code importing it.
	_ "github.com/goccy/go-perl/xs"
)

// adapter is the Perl half: psgi_load compiles the app via Plack::Util,
// psgi_handle runs one request. Bodies cross base64-encoded in both
// directions because the bridge speaks JSON, and JSON strings cannot carry
// arbitrary bytes.
//
//go:embed psgi.pl
var adapter string

// MaxRequestBody caps how much of a request body Handle reads.
const MaxRequestBody = 32 << 20

// Handle forwards one HTTP request into p's loaded application and writes
// the PSGI response. The caller guarantees p is not serving another request
// concurrently. Cancelling ctx stops the Perl code at the next opcode.
func Handle(ctx context.Context, p *perl.Perl, w http.ResponseWriter, r *http.Request) error {
	return handle(ctx, p, w, r, false)
}

// handle runs one request; multiproc feeds the PSGI env's
// psgi.multiprocess key (true when other workers may serve concurrently).
func handle(ctx context.Context, p *perl.Perl, w http.ResponseWriter, r *http.Request, multiproc bool) error {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxRequestBody))
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return err
	}
	env := Env(r, len(body))
	if multiproc {
		// Workers are isolated interpreters serving at the same time: the
		// PSGI flag for "another copy of the app may run concurrently and
		// shares nothing with you" is multiprocess.
		env["psgi.multiprocess"] = true
	}
	res, err := p.Call(ctx, "psgi_handle", env,
		base64.StdEncoding.EncodeToString(body))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return err
	}
	return WriteResponse(w, res)
}

// Env flattens the request into the CGI-style PSGI environment. The psgi.*
// runtime keys are filled in Perl-side by the adapter.
func Env(r *http.Request, bodyLen int) map[string]any {
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

// WriteResponse maps psgi_handle's (status, header pairs, base64 body)
// return list onto the ResponseWriter.
func WriteResponse(w http.ResponseWriter, res []any) error {
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

// Server serves PSGI traffic over a fixed set of warm Perl workers — the
// role starman/starlet's preforked workers play, with interpreters instead
// of processes: each request checks a worker out, runs through Handle, and
// returns it. Goroutine-per-request on the Go side, never more than one
// request per worker at a time; the worker count is the max concurrency.
// A Server is an http.Handler, so it plugs into any Go HTTP stack;
// ListenAndServe is the shortcut for the plain case.
type Server struct {
	workers chan *perl.Perl
	multi   bool // more than one worker serves concurrently
}

// New compiles the PSGI application at appPath into the PREPARED prototype
// (module trees on @INC via p.AddInc, native XS modules via xs.LoadDir —
// anything every worker should share) and scales it to workers serving
// workers: the remaining workers-1 are copy-on-write clones (perl.Clone),
// so none of the loading work re-runs per worker and the workers share the
// read-only bulk of their memory. The prototype itself becomes worker 1
// and is owned by the returned Server (Close closes it). On any failure
// everything built so far is closed.
func New(prototype *perl.Perl, workers int, appPath string) (*Server, error) {
	if workers < 1 {
		return nil, fmt.Errorf("psgi: New needs at least 1 worker, got %d", workers)
	}
	ctx := context.Background()
	// Install this package's adapter (idempotent) and compile the
	// application through it: the file must yield a PSGI application — a
	// code reference taking $env and returning [$status, \@headers,
	// $body] — and anything else fails New, before any worker exists and
	// before anything serves.
	installed, err := adapterInstalled(ctx, prototype)
	if err != nil {
		return nil, err
	}
	if !installed {
		if r, err := prototype.Eval(ctx, adapter); err != nil {
			return nil, fmt.Errorf("install PSGI adapter: %w", err)
		} else if r.Error != nil {
			return nil, fmt.Errorf("install PSGI adapter: %w", r.Error)
		}
	}
	if _, err := prototype.Call(ctx, "psgi_load", appPath); err != nil {
		return nil, fmt.Errorf("load %s: %w", appPath, err)
	}
	ws := &Server{workers: make(chan *perl.Perl, workers), multi: workers > 1}
	ws.workers <- prototype
	for i := 1; i < workers; i++ {
		c, err := prototype.Clone()
		if err != nil {
			ws.Close()
			return nil, fmt.Errorf("clone worker %d: %w", i+1, err)
		}
		ws.workers <- c
	}
	return ws, nil
}

// adapterInstalled reports whether this package's adapter is already in
// the instance (so New is idempotent across calls on the same prototype).
func adapterInstalled(ctx context.Context, p *perl.Perl) (bool, error) {
	r, err := p.Eval(ctx, `(defined &main::psgi_load && defined &main::psgi_handle) ? 1 : 0`)
	if err != nil {
		return false, fmt.Errorf("probe PSGI adapter: %w", err)
	}
	if r.Error != nil {
		return false, fmt.Errorf("probe PSGI adapter: %w", r.Error)
	}
	return r.Value.Bool(), nil
}

// ListenAndServe serves the application on addr until the listener fails.
// For anything beyond the plain case (TLS, timeouts, middleware), use the
// Server as the http.Handler it is.
func (ws *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, ws)
}

// ServeHTTP forwards one request into a free worker. Cancelling the request
// context while the app runs stops the Perl code at the next opcode.
func (ws *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var p *perl.Perl
	select {
	case p = <-ws.workers:
	case <-r.Context().Done():
		http.Error(w, "no worker available", http.StatusServiceUnavailable)
		return
	}
	defer func() { ws.workers <- p }()

	if err := handle(r.Context(), p, w, r, ws.multi); err != nil {
		log.Printf("psgi: %s %s: %v", r.Method, r.URL.Path, err)
	}
}

// Close shuts down every worker. Callers stop serving first; Close does not
// wait for in-flight requests.
func (ws *Server) Close() {
	close(ws.workers)
	for p := range ws.workers {
		p.Close()
	}
}
