// Package psgi serves PSGI Perl web applications from Go's net/http.
//
// PSGI specifies the application: the .psgi file's last evaluated value is
// a code reference that takes the environment hashref (CGI-style keys plus
// psgi.*) and returns [$status, \@headers, $body]. This package supplies
// the SERVER side of that interface — the environment (psgi.streaming is
// false: this server does not implement delayed/streaming responses), the
// request/response plumbing, and the worker model.
//
// The application is exactly what PSGI says it is: New evaluates the file
// and holds its last value, a code reference; every request builds the
// environment as a guest hash and calls it. Values cross the bridge typed —
// request and response bodies are raw byte strings, the response arrayref
// is walked through the value API — so no serialisation format sits between
// Go and the application.
//
// Server is a fixed set of warm interpreter workers (the starman/starlet
// serving model). Prepare ONE instance with the instance's own loading
// facilities, and New compiles the app into it and scales it: every other
// worker is a copy-on-write clone (perl.Clone) that adopts the same loaded
// application (perl.Adopt), so the loading work runs once no matter the
// worker count:
//
//	stdlib, _ := perl.ExtractStdlib()
//	p, _ := perl.New(perl.Config{FS: fs.NewHostFS(), StdlibDir: stdlib})
//	p.AddInc(ctx, "local/lib/perl5")
//	p.AddXSDir("local/xs/" + xs.ArchTag())
//	server, _ := psgi.New(p, 4, "app.psgi")
//	server.ListenAndServe(":8091") // or http.ListenAndServe(addr, server)
//
// A *perl.Perl must not run concurrent requests; Server guarantees that.
package psgi

import (
	"context"
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

// helperProgram defines the two generic byte<->filehandle conversions a
// PSGI server needs on the Perl side of the boundary (a filehandle cannot
// cross as data): building psgi.input over the raw request body, and
// draining a filehandle response body. Its last value is [$mk_input,
// $read_body, \*STDERR] so one Eval yields all three.
const helperProgram = `
my $mk_input = sub {
    open my $in, '<', \$_[0] or die "open request body: $!\n";
    binmode $in;
    return $in;
};
my $read_body = sub {
    my ($fh) = @_;
    local $/;
    my $data = <$fh>;
    eval { close $fh };
    return defined $data ? $data : '';
};
[$mk_input, $read_body, \*STDERR];
`

// worker is one interpreter with its adopted handles to the loaded
// application and the request helpers.
type worker struct {
	p        *perl.Perl
	app      perl.CodeValue
	mkInput  perl.CodeValue
	readBody perl.CodeValue
	stderr   perl.RefValue // \*STDERR, the psgi.errors stream
	version  perl.RefValue // [1, 1], the psgi.version arrayref
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

// New loads the PSGI application at appPath into the PREPARED prototype
// (module trees on @INC via p.AddInc, native XS modules via p.AddXSDir —
// anything every worker should share) and scales it to workers serving
// workers: the remaining workers-1 are copy-on-write clones (perl.Clone)
// that adopt the same loaded application (perl.Adopt), so none of the
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
	proto, err := loadWorker(ctx, prototype, appPath)
	if err != nil {
		return nil, err
	}
	ws := &Server{workers: make(chan worker, workers), multi: workers > 1}
	ws.workers <- proto
	for i := 1; i < workers; i++ {
		c, err := prototype.Clone()
		if err != nil {
			ws.Close()
			return nil, fmt.Errorf("clone worker %d: %w", i+1, err)
		}
		wk, err := adoptWorker(c, proto)
		if err != nil {
			c.Close()
			ws.Close()
			return nil, fmt.Errorf("adopt application in worker %d: %w", i+1, err)
		}
		ws.workers <- wk
	}
	return ws, nil
}

// loadWorker evaluates the application file per the PSGI contract — the
// file's last value is the application — plus the server helpers, and
// returns the prototype worker.
func loadWorker(ctx context.Context, p *perl.Perl, appPath string) (worker, error) {
	abs, err := filepath.Abs(appPath)
	if err != nil {
		return worker{}, err
	}
	src, err := os.ReadFile(abs)
	if err != nil {
		return worker{}, err
	}
	// The #line directive makes compile errors and warnings name the
	// application file instead of the eval site.
	res, err := p.Eval(ctx, fmt.Sprintf("#line 1 \"%s\"\n%s",
		strings.ReplaceAll(abs, `"`, `\"`), src))
	if err != nil {
		return worker{}, fmt.Errorf("load %s: %w", appPath, err)
	}
	if res.Error != nil {
		return worker{}, fmt.Errorf("load %s: %w", appPath, res.Error)
	}
	app, err := derefCode(ctx, res.Value)
	if err != nil {
		return worker{}, fmt.Errorf("load %s: %w", appPath, err)
	}

	hres, err := p.Eval(ctx, helperProgram)
	if err != nil {
		return worker{}, fmt.Errorf("install request helpers: %w", err)
	}
	if hres.Error != nil {
		return worker{}, fmt.Errorf("install request helpers: %w", hres.Error)
	}
	helpers, err := derefArray(ctx, hres.Value)
	if err != nil {
		return worker{}, fmt.Errorf("install request helpers: %w", err)
	}
	hv, err := helpers.Values(ctx)
	if err != nil || len(hv) != 3 {
		return worker{}, fmt.Errorf("install request helpers: got %d values (%v)", len(hv), err)
	}
	mkInput, err := derefCode(ctx, hv[0])
	if err != nil {
		return worker{}, fmt.Errorf("install request helpers: %w", err)
	}
	readBody, err := derefCode(ctx, hv[1])
	if err != nil {
		return worker{}, fmt.Errorf("install request helpers: %w", err)
	}
	stderrRef, err := perl.As[perl.RefValue](hv[2])
	if err != nil {
		return worker{}, fmt.Errorf("install request helpers: %w", err)
	}

	version, err := p.NewArray(ctx, perl.NewValue(1), perl.NewValue(1))
	if err != nil {
		return worker{}, fmt.Errorf("build psgi.version: %w", err)
	}
	return worker{
		p:        p,
		app:      app,
		mkInput:  mkInput,
		readBody: readBody,
		stderr:   stderrRef,
		version:  version.Ref(),
	}, nil
}

// adoptWorker rebinds the prototype's handles to a fresh clone: the clone's
// guest memory carries the same registry, so every handle designates the
// same (copied) value.
func adoptWorker(c *perl.Perl, proto worker) (worker, error) {
	app, err := perl.Adopt(c, proto.app)
	if err != nil {
		return worker{}, err
	}
	mkInput, err := perl.Adopt(c, proto.mkInput)
	if err != nil {
		return worker{}, err
	}
	readBody, err := perl.Adopt(c, proto.readBody)
	if err != nil {
		return worker{}, err
	}
	stderr, err := perl.Adopt(c, proto.stderr)
	if err != nil {
		return worker{}, err
	}
	version, err := perl.Adopt(c, proto.version)
	if err != nil {
		return worker{}, err
	}
	return worker{p: c, app: app, mkInput: mkInput, readBody: readBody,
		stderr: stderr, version: version}, nil
}

// derefCode expects v to be a code reference and returns the subroutine.
func derefCode(ctx context.Context, v perl.Value) (perl.CodeValue, error) {
	ref, err := perl.As[perl.RefValue](v)
	if err != nil {
		return perl.CodeValue{}, fmt.Errorf("want a code reference: %w", err)
	}
	inner, err := ref.Deref(ctx)
	if err != nil {
		return perl.CodeValue{}, err
	}
	return perl.As[perl.CodeValue](inner)
}

// derefArray expects v to be an array reference and returns the array.
func derefArray(ctx context.Context, v perl.Value) (perl.ArrayValue, error) {
	ref, err := perl.As[perl.RefValue](v)
	if err != nil {
		return perl.ArrayValue{}, fmt.Errorf("want an array reference: %w", err)
	}
	inner, err := ref.Deref(ctx)
	if err != nil {
		return perl.ArrayValue{}, err
	}
	return perl.As[perl.ArrayValue](inner)
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
// concurrently — isolated interpreters sharing nothing, PSGI's multiprocess
// model).
func serve(ctx context.Context, wk worker, w http.ResponseWriter, r *http.Request, multiproc bool) error {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBody))
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return err
	}
	input, err := wk.mkInput.CallScalar(ctx, perl.NewValue(body))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return fmt.Errorf("build psgi.input: %w", err)
	}
	env, err := wk.p.NewHash(ctx, requestEnv(r, wk, len(body), input, multiproc)...)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return fmt.Errorf("build environment: %w", err)
	}
	res, err := wk.app.Call(ctx, env.Ref())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return err
	}
	if err := writeResponse(ctx, wk, w, res); err != nil {
		return err
	}
	return nil
}

// requestEnv flattens the request into the CGI-style PSGI environment plus
// the psgi.* runtime keys.
func requestEnv(r *http.Request, wk worker, bodyLen int, input perl.Value, multiproc bool) []perl.Pair {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		host, port = r.Host, "80"
	}
	remote := r.RemoteAddr
	if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		remote = h
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	pairs := []perl.Pair{
		{K: "REQUEST_METHOD", V: perl.NewValue(r.Method)},
		{K: "SCRIPT_NAME", V: perl.NewValue("")},
		{K: "PATH_INFO", V: perl.NewValue(r.URL.Path)},
		{K: "REQUEST_URI", V: perl.NewValue(r.URL.RequestURI())},
		{K: "QUERY_STRING", V: perl.NewValue(r.URL.RawQuery)},
		{K: "SERVER_NAME", V: perl.NewValue(host)},
		{K: "SERVER_PORT", V: perl.NewValue(port)},
		{K: "SERVER_PROTOCOL", V: perl.NewValue(r.Proto)},
		{K: "REMOTE_ADDR", V: perl.NewValue(remote)},
		{K: "CONTENT_LENGTH", V: perl.NewValue(bodyLen)},
		{K: "psgi.version", V: wk.version},
		{K: "psgi.url_scheme", V: perl.NewValue(scheme)},
		{K: "psgi.input", V: input},
		{K: "psgi.errors", V: wk.stderr},
		{K: "psgi.multithread", V: perl.NewValue(false)},
		{K: "psgi.multiprocess", V: perl.NewValue(multiproc)},
		{K: "psgi.run_once", V: perl.NewValue(false)},
		{K: "psgi.streaming", V: perl.NewValue(false)},
		{K: "psgi.nonblocking", V: perl.NewValue(false)},
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		pairs = append(pairs, perl.Pair{K: "CONTENT_TYPE", V: perl.NewValue(ct)})
	}
	for k, vs := range r.Header {
		switch k {
		case "Content-Type", "Content-Length":
			continue
		}
		pairs = append(pairs, perl.Pair{
			K: "HTTP_" + strings.ReplaceAll(strings.ToUpper(k), "-", "_"),
			V: perl.NewValue(strings.Join(vs, ", ")),
		})
	}
	return pairs
}

// writeResponse maps the application's [$status, \@headers, $body] response
// onto the ResponseWriter. The body is an arrayref of byte strings or a
// filehandle (psgi.streaming is off, so the delayed-response code-ref form
// is out of contract).
func writeResponse(ctx context.Context, wk worker, w http.ResponseWriter, res []perl.Value) error {
	if len(res) < 1 {
		return fmt.Errorf("PSGI app returned no response")
	}
	arr, err := derefArray(ctx, res[0])
	if err != nil {
		return fmt.Errorf("PSGI response: %w", err)
	}
	parts, err := arr.Values(ctx)
	if err != nil {
		return fmt.Errorf("PSGI response: %w", err)
	}
	if len(parts) != 3 {
		return fmt.Errorf("PSGI response has %d elements, want 3", len(parts))
	}
	statusSV, err := perl.As[perl.ScalarValue](parts[0])
	if err != nil {
		return fmt.Errorf("PSGI status: %w", err)
	}
	status := int(statusSV.Int())
	if status < 100 || status > 999 {
		return fmt.Errorf("bad PSGI status %d", status)
	}

	headers, err := derefArray(ctx, parts[1])
	if err != nil {
		return fmt.Errorf("PSGI headers: %w", err)
	}
	hvals, err := headers.Values(ctx)
	if err != nil {
		return fmt.Errorf("PSGI headers: %w", err)
	}
	if len(hvals)%2 != 0 {
		return fmt.Errorf("PSGI header list has odd length %d", len(hvals))
	}
	for i := 0; i+1 < len(hvals); i += 2 {
		k, kerr := perl.As[perl.ScalarValue](hvals[i])
		v, verr := perl.As[perl.ScalarValue](hvals[i+1])
		if kerr != nil || verr != nil || k.String() == "" {
			continue
		}
		w.Header().Add(k.String(), v.String())
	}

	body, err := responseBody(ctx, wk, parts[2])
	if err != nil {
		return fmt.Errorf("PSGI body: %w", err)
	}
	w.WriteHeader(status)
	_, err = w.Write(body)
	return err
}

// responseBody drains the PSGI body: an arrayref of byte strings, or a
// filehandle (read through the guest, where the handle lives).
func responseBody(ctx context.Context, wk worker, v perl.Value) ([]byte, error) {
	ref, err := perl.As[perl.RefValue](v)
	if err != nil {
		return nil, err
	}
	inner, err := ref.Deref(ctx)
	if err != nil {
		return nil, err
	}
	switch x := inner.(type) {
	case perl.ArrayValue:
		chunks, err := x.Values(ctx)
		if err != nil {
			return nil, err
		}
		var out []byte
		for i, c := range chunks {
			sv, err := perl.As[perl.ScalarValue](c)
			if err != nil {
				return nil, fmt.Errorf("body element %d: %w", i, err)
			}
			out = append(out, sv.Bytes()...)
		}
		return out, nil
	case perl.GlobValue, perl.IOValue:
		data, err := wk.readBody.CallScalar(ctx, ref)
		if err != nil {
			return nil, err
		}
		sv, err := perl.As[perl.ScalarValue](data)
		if err != nil {
			return nil, err
		}
		return sv.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported body form %s", inner.Kind())
	}
}

// Close shuts down every worker. Callers stop serving first; Close does not
// wait for in-flight requests.
func (ws *Server) Close() {
	close(ws.workers)
	for wk := range ws.workers {
		wk.p.Close()
	}
}
