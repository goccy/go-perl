// Package psgi serves PSGI Perl web applications from Go's net/http.
//
// PSGI specifies the application: the .psgi file's last evaluated value is
// a code reference that takes the environment hashref (CGI-style keys plus
// psgi.*) and returns [$status, \@headers, $body]. This package supplies
// the SERVER side of that interface — the environment (psgi.streaming is
// false: this server does not implement delayed/streaming responses), the
// request/response plumbing, and the worker model.
//
// The application is exactly what PSGI says it is: New loads the file with
// Plack::Util::load_psgi and holds the resulting code reference; every
// request calls it. The only Perl that runs besides the application is a
// per-request closure wrapped around it at load time, because two things
// cannot cross the Go/Perl boundary as data: psgi.input must be a real
// filehandle over the request body, and response bodies are arbitrary
// bytes while the bridge speaks JSON (bodies cross base64-encoded in both
// directions).
//
// Server is a fixed set of warm interpreter workers (the starman/starlet
// serving model). Prepare ONE instance with the instance's own loading
// facilities, and New compiles the app into it and scales it: every other
// worker is a copy-on-write clone (perl.Clone) that adopts the same loaded
// application (perl.AdoptRef), so the loading work runs once no matter the
// worker count:
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
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	perl "github.com/goccy/go-perl"
	// Serving stacks routinely carry native XS modules; linking the
	// loader here makes (*perl.Perl).AddXSDir work without application
	// code importing it.
	_ "github.com/goccy/go-perl/xs"
)

// maxRequestBody caps how much of a request body a worker reads.
const maxRequestBody = 32 << 20

// requestEnv flattens the request into the CGI-style PSGI environment.
// The psgi.* runtime keys are filled in Perl-side when the application
// runs.
func requestEnv(r *http.Request, bodyLen int) map[string]any {
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

// writeResponse maps the application's (status, header pairs, base64 body)
// return list onto the ResponseWriter.
func writeResponse(w http.ResponseWriter, res []any) error {
	if len(res) != 3 {
		return fmt.Errorf("PSGI response crossed as %d values, want 3", len(res))
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

// worker is one interpreter with its adopted handle to the loaded
// application.
type worker struct {
	p   *perl.Perl
	app *perl.Ref
}

// Server serves PSGI traffic over a fixed set of warm Perl workers — the
// role starman/starlet's preforked workers play, with interpreters instead
// of processes: each request checks a worker out, runs through the
// application, and returns it. Goroutine-per-request on the Go side, never
// more than one request per worker at a time; the worker count is the max
// concurrency. A Server is an http.Handler, so it plugs into any Go HTTP
// stack; ListenAndServe is the shortcut for the plain case.
type Server struct {
	workers chan worker
	multi   bool // more than one worker serves concurrently
}

// loaderProgram is the .psgi program New synthesizes around the real
// application file. Its last evaluated value — the PSGI contract — is a
// code reference closing over the loaded application; the closure carries
// the two conversions that cannot cross the bridge as data (psgi.input as
// a real filehandle, bodies as base64) and the psgi.* environment keys.
const loaderProgram = `use strict;
use warnings;
use MIME::Base64 qw(decode_base64 encode_base64);
use Plack::Util;

my $path = %s;
my $app = Plack::Util::load_psgi($path);
die "$path did not return a PSGI application (a code reference)\n"
    unless ref $app eq 'CODE';

sub {
    my ($env, $body_b64) = @_;

    my $body = decode_base64(defined $body_b64 ? $body_b64 : '');
    open my $in, '<', \$body or die "open request body: $!\n";

    $env->{'psgi.version'}      = [1, 1];
    $env->{'psgi.url_scheme'}   = $env->{'psgi.url_scheme'} || 'http';
    $env->{'psgi.input'}        = $in;
    $env->{'psgi.errors'}       = \*STDERR;
    $env->{'psgi.multithread'}  = Plack::Util::FALSE;
    # the server passes psgi.multiprocess when other workers serve
    # concurrently; a lone instance defaults to false
    $env->{'psgi.multiprocess'} = Plack::Util::FALSE
        unless exists $env->{'psgi.multiprocess'};
    $env->{'psgi.run_once'}     = Plack::Util::FALSE;
    $env->{'psgi.streaming'}    = Plack::Util::FALSE;
    $env->{'psgi.nonblocking'}  = Plack::Util::FALSE;

    my $res = $app->($env);
    die "PSGI app returned a non-arrayref response (psgi.streaming is off)\n"
        unless ref $res eq 'ARRAY';
    my ($status, $headers, $out_body) = @$res;

    my $out = '';
    Plack::Util::foreach($out_body, sub { $out .= $_[0] if defined $_[0] });
    if (ref($out_body) ne 'ARRAY') {
        eval { $out_body->close };
    }
    # encode_base64 wants octets; a character string (wide chars) would die.
    utf8::encode($out) if utf8::is_utf8($out);

    ($status, $headers, encode_base64($out, ''));
};
`

// New loads the PSGI application at appPath into the PREPARED prototype
// (module trees on @INC via p.AddInc, native XS modules via p.AddXSDir —
// anything every worker should share) and scales it to workers serving
// workers: the remaining workers-1 are copy-on-write clones (perl.Clone)
// that adopt the same loaded application (perl.AdoptRef), so none of the
// loading work re-runs per worker and the workers share the read-only bulk
// of their memory. The prototype itself becomes worker 1 and is owned by
// the returned Server (Close closes it). On any failure everything built
// so far is closed.
//
// The file must satisfy the PSGI contract — its last evaluated value is a
// code reference taking $env — and anything else fails New, before any
// worker exists and before anything serves.
func New(prototype *perl.Perl, workers int, appPath string) (*Server, error) {
	if workers < 1 {
		return nil, fmt.Errorf("psgi: New needs at least 1 worker, got %d", workers)
	}
	ctx := context.Background()
	app, err := loadApp(ctx, prototype, appPath)
	if err != nil {
		return nil, err
	}
	ws := &Server{workers: make(chan worker, workers), multi: workers > 1}
	ws.workers <- worker{p: prototype, app: app}
	for i := 1; i < workers; i++ {
		c, err := prototype.Clone()
		if err != nil {
			ws.Close()
			return nil, fmt.Errorf("clone worker %d: %w", i+1, err)
		}
		capp, err := c.AdoptRef(app)
		if err != nil {
			c.Close()
			ws.Close()
			return nil, fmt.Errorf("adopt application in worker %d: %w", i+1, err)
		}
		ws.workers <- worker{p: c, app: capp}
	}
	return ws, nil
}

// loadApp evaluates the application file per the PSGI contract and returns
// the request-serving code reference (the synthesized closure over it).
func loadApp(ctx context.Context, p *perl.Perl, appPath string) (*perl.Ref, error) {
	abs, err := filepath.Abs(appPath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, err
	}
	if r, err := p.Eval(ctx, "require Plack::Util; 1"); err != nil {
		return nil, fmt.Errorf("load Plack::Util: %w", err)
	} else if r.Error != nil {
		return nil, fmt.Errorf("load Plack::Util (is Plack on @INC?): %w", r.Error)
	}
	// The loader is itself a .psgi program: written to a scratch file,
	// loaded the standard way, gone afterwards.
	quoted := "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(abs) + "'"
	tmp, err := os.CreateTemp("", "gperl-psgi-*.psgi")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := fmt.Fprintf(tmp, loaderProgram, quoted); err != nil {
		tmp.Close()
		return nil, err
	}
	tmp.Close()
	res, err := p.Call(ctx, "Plack::Util::load_psgi", tmpPath)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", appPath, err)
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("load %s: %d results, want the application", appPath, len(res))
	}
	app, ok := res[0].(*perl.Ref)
	if !ok || app.Reftype() != "CODE" {
		return nil, fmt.Errorf("load %s: got %#v, want a code reference", appPath, res[0])
	}
	return app, nil
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
	var wk worker
	select {
	case wk = <-ws.workers:
	case <-r.Context().Done():
		http.Error(w, "no worker available", http.StatusServiceUnavailable)
		return
	}
	defer func() { ws.workers <- wk }()

	if err := serve(r.Context(), wk, w, r, ws.multi); err != nil {
		log.Printf("psgi: %s %s: %v", r.Method, r.URL.Path, err)
	}
}

// serve runs one request through the worker's application; multiproc feeds
// the PSGI env's psgi.multiprocess key (true when other workers may serve
// concurrently).
func serve(ctx context.Context, wk worker, w http.ResponseWriter, r *http.Request, multiproc bool) error {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBody))
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return err
	}
	env := requestEnv(r, len(body))
	if multiproc {
		// Workers are isolated interpreters serving at the same time: the
		// PSGI flag for "another copy of the app may run concurrently and
		// shares nothing with you" is multiprocess.
		env["psgi.multiprocess"] = true
	}
	res, err := wk.app.Invoke(ctx, env, base64.StdEncoding.EncodeToString(body))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return err
	}
	return writeResponse(w, res)
}

// Close shuts down every worker. Callers stop serving first; Close does not
// wait for in-flight requests.
func (ws *Server) Close() {
	close(ws.workers)
	for wk := range ws.workers {
		wk.app.Free()
		wk.p.Close()
	}
}
