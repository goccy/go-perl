// Package psgi is the PSGI server behind the public go-perl/psgi package
// and the gperl psgi commands: the worker pool, the request and response
// plumbing, and the settings (Options) only the commands expose so far.
package psgi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	perl "github.com/goccy/go-perl"
)

// defaultMaxRequestBody caps how much of a request body is read when
// Options.MaxRequestBody is zero.
const defaultMaxRequestBody = 32 << 20

// Options tunes a Server.
type Options struct {
	// MaxRequests replaces a worker with a fresh clone after it has
	// served this many requests (the usual guard against applications
	// whose memory grows). Zero keeps workers forever.
	MaxRequests int
	// MaxRequestBody is the largest request body accepted, in bytes.
	// Zero means 32 MiB.
	MaxRequestBody int64
	// AccessLog receives one line per request in the Apache combined
	// format. Nil disables the log.
	AccessLog io.Writer
}

// helperProgram is the Perl half of the server. Its value is a sub that
// takes the application and returns the request handler: the handler
// completes the environment, runs the application, and reports the
// response to Go in one call — respond for a complete response, start
// followed by the writer's methods for a streaming one. Everything that
// crosses is a scalar, so no call nests back into Perl.
const helperProgram = `
package GoPerl::PSGI;
use strict;
use warnings;
sub {
    my ($app) = @_;
    my $flatten = sub {
        my ($body) = @_;
        return join('', map { defined $_ ? $_ : '' } @$body) if ref $body eq 'ARRAY';
        # PSGI's second body form: a filehandle OR any object with
        # getline/close. A blessed handle-like object drains through
        # getline (Plack's own precedence); an unblessed glob reads
        # directly (->can dies on it, so probe under eval).
        my $out = '';
        my $can_getline = do { local $@; eval { $body->can('getline') } };
        if ($can_getline) {
            while (defined(my $chunk = $body->getline)) { $out .= $chunk }
        } else {
            local $/;
            my $data = <$body>;
            $out = $data if defined $data;
        }
        do { local $@; eval { $body->close } };
        do { local $@; eval { close $body } } unless $can_getline;
        return $out;
    };
    my $responder = sub {
        my ($res) = @_;
        die "PSGI response must be an array reference\n" unless ref $res eq 'ARRAY';
        if (@$res >= 3) {
            GoPerl::PSGI::respond($res->[0], $flatten->($res->[2]), @{ $res->[1] });
            return;
        }
        GoPerl::PSGI::start($res->[0], @{ $res->[1] });
        return bless {}, 'GoPerl::PSGI::Writer';
    };
    return sub {
        my ($env, $input) = @_;
        open my $in, '<', \$input or die "open request body: $!\n";
        binmode $in;
        $env->{'psgi.input'}   = $in;
        $env->{'psgi.errors'}  = \*STDERR;
        $env->{'psgi.version'} = [1, 1];
        my $res = $app->($env);
        if (ref $res eq 'CODE') { $res->($responder) } else { $responder->($res) }
        return;
    };
};
`

// worker is one interpreter with its handle to the request handler.
type worker struct {
	p      *perl.Perl
	handle perl.CodeValue
	served int
	// req is the request the worker is running; the bound response
	// functions write to it.
	req *response
}

// response collects what the application reported for one request.
type response struct {
	w         http.ResponseWriter
	status    int
	header    []string // key, value, ...
	body      []byte
	responded bool // a complete response was reported
	started   bool // a streaming response was started (headers sent)
	err       error
}

// Server serves PSGI traffic over a fixed set of warm Perl workers — the
// role starman/starlet's preforked workers play, with interpreters instead
// of processes: each request checks a worker out, runs through the
// application, and returns it. Goroutine-per-request on the Go side, never
// more than one request per worker at a time; the worker count is the max
// concurrency. A Server is an http.Handler, so it plugs into any Go HTTP
// stack.
type Server struct {
	opts    Options
	workers chan *worker
	multi   bool // more than one worker serves concurrently

	// proto holds the loaded application and never serves: every worker,
	// including the replacements MaxRequests asks for, is cloned from it.
	protoMu sync.Mutex
	proto   *worker
	logMu   sync.Mutex
}

// NewServer loads the PSGI application at appPath into the PREPARED
// prototype (module trees on @INC, native XS modules — anything every
// worker should share) and serves it with workers workers, each a
// copy-on-write clone of the prototype (perl.Clone), so none of the loading
// work re-runs per worker and the workers share the read-only bulk of
// their memory. The prototype is owned by the returned Server (Close
// closes it). On any failure everything built so far is closed.
//
// The file must satisfy the PSGI contract — its last evaluated value is a
// code reference taking $env — and anything else fails NewServer, before
// any worker exists and before anything serves.
func NewServer(prototype *perl.Perl, workers int, appPath string, opts Options) (*Server, error) {
	if workers < 1 {
		return nil, fmt.Errorf("psgi: needs at least 1 worker, got %d", workers)
	}
	if opts.MaxRequestBody == 0 {
		opts.MaxRequestBody = defaultMaxRequestBody
	}
	ctx := context.Background()
	handle, err := loadHandler(ctx, prototype, appPath)
	if err != nil {
		return nil, err
	}
	ws := &Server{opts: opts, workers: make(chan *worker, workers), multi: workers > 1,
		proto: &worker{p: prototype, handle: handle}}
	for i := 0; i < workers; i++ {
		wk, err := ws.cloneWorker()
		if err != nil {
			ws.Close()
			return nil, fmt.Errorf("worker %d: %w", i+1, err)
		}
		ws.workers <- wk
	}
	return ws, nil
}

// loadHandler evaluates the application file per the PSGI contract — the
// file's last value is the application — and wraps it in the request
// handler.
func loadHandler(ctx context.Context, p *perl.Perl, appPath string) (perl.CodeValue, error) {
	none := perl.CodeValue{}
	abs, err := filepath.Abs(appPath)
	if err != nil {
		return none, err
	}
	src, err := os.ReadFile(abs)
	if err != nil {
		return none, err
	}
	// The #line directive makes compile errors and warnings name the
	// application file instead of the eval site.
	res, err := p.Eval(ctx, fmt.Sprintf("#line 1 \"%s\"\n%s",
		strings.ReplaceAll(abs, `"`, `\"`), src))
	if err == nil && res.Error != nil {
		err = res.Error
	}
	if err != nil {
		return none, fmt.Errorf("load %s: %w", appPath, err)
	}
	app, err := derefCode(ctx, res.Value)
	if err != nil {
		return none, fmt.Errorf("load %s: %w", appPath, err)
	}
	hres, err := p.Eval(ctx, helperProgram)
	if err == nil && hres.Error != nil {
		err = hres.Error
	}
	if err != nil {
		return none, fmt.Errorf("install the request handler: %w", err)
	}
	wrap, err := derefCode(ctx, hres.Value)
	if err != nil {
		return none, fmt.Errorf("install the request handler: %w", err)
	}
	handler, err := wrap.CallScalar(ctx, app.Ref())
	if err != nil {
		return none, fmt.Errorf("install the request handler: %w", err)
	}
	handle, err := derefCode(ctx, handler)
	if err != nil {
		return none, fmt.Errorf("install the request handler: %w", err)
	}
	return handle, nil
}

// cloneWorker makes a serving worker from the prototype: a clone that
// adopts the handler and binds the response functions to its own state.
func (ws *Server) cloneWorker() (*worker, error) {
	ws.protoMu.Lock()
	defer ws.protoMu.Unlock()
	c, err := ws.proto.p.Clone()
	if err != nil {
		return nil, fmt.Errorf("clone: %w", err)
	}
	handle, err := perl.Adopt(c, ws.proto.handle)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("adopt the application: %w", err), c.Close())
	}
	wk := &worker{p: c, handle: handle}
	for name, fn := range map[string]perl.GoFunc{
		"GoPerl::PSGI::respond":       wk.respond,
		"GoPerl::PSGI::start":         wk.start,
		"GoPerl::PSGI::Writer::write": wk.write,
		"GoPerl::PSGI::Writer::close": func([]perl.Value) ([]perl.Value, error) { return nil, nil },
	} {
		if err := c.Bind(name, fn); err != nil {
			return nil, errors.Join(fmt.Errorf("bind %s: %w", name, err), c.Close())
		}
	}
	return wk, nil
}

// statusAndHeaders decodes a status scalar and a flat header list.
func statusAndHeaders(status perl.Value, headers []perl.Value) (int, []string, error) {
	sv, err := perl.As[perl.ScalarValue](status)
	if err != nil {
		return 0, nil, fmt.Errorf("PSGI status: %w", err)
	}
	code := int(sv.Int())
	if code < 100 || code > 999 {
		return 0, nil, fmt.Errorf("bad PSGI status %d", code)
	}
	if len(headers)%2 != 0 {
		return 0, nil, fmt.Errorf("PSGI header list has odd length %d", len(headers))
	}
	out := make([]string, 0, len(headers))
	for i := 0; i < len(headers); i += 2 {
		k, kerr := perl.As[perl.ScalarValue](headers[i])
		v, verr := perl.As[perl.ScalarValue](headers[i+1])
		if kerr != nil || verr != nil || k.String() == "" {
			continue
		}
		out = append(out, k.String(), v.String())
	}
	return code, out, nil
}

// respond receives a complete response: (status, body, header pairs...).
func (wk *worker) respond(args []perl.Value) ([]perl.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("PSGI response is incomplete")
	}
	status, header, err := statusAndHeaders(args[0], args[2:])
	if err != nil {
		return nil, err
	}
	body, err := perl.As[perl.ScalarValue](args[1])
	if err != nil {
		return nil, fmt.Errorf("PSGI body: %w", err)
	}
	wk.req.status, wk.req.header, wk.req.body, wk.req.responded = status, header, body.Bytes(), true
	return nil, nil
}

// start begins a streaming response: (status, header pairs...). The
// headers go out now; the writer's write sends the body as it comes.
func (wk *worker) start(args []perl.Value) ([]perl.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("PSGI response is incomplete")
	}
	status, header, err := statusAndHeaders(args[0], args[1:])
	if err != nil {
		return nil, err
	}
	wk.req.status, wk.req.started = status, true
	writeHead(wk.req.w, status, header)
	wk.req.err = http.NewResponseController(wk.req.w).Flush()
	return nil, nil
}

// write sends one chunk of a streaming response: (writer, chunk).
func (wk *worker) write(args []perl.Value) ([]perl.Value, error) {
	if len(args) < 2 || !wk.req.started {
		return nil, fmt.Errorf("write on a PSGI writer that was not started")
	}
	chunk, err := perl.As[perl.ScalarValue](args[1])
	if err != nil {
		return nil, fmt.Errorf("PSGI chunk: %w", err)
	}
	if _, err := wk.req.w.Write(chunk.Bytes()); err != nil {
		// The client went away; the application learns by the die.
		return nil, err
	}
	return nil, http.NewResponseController(wk.req.w).Flush()
}

func writeHead(w http.ResponseWriter, status int, header []string) {
	for i := 0; i+1 < len(header); i += 2 {
		w.Header().Add(header[i], header[i+1])
	}
	w.WriteHeader(status)
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

// countingWriter records what a request's response amounted to, for the
// access log.
type countingWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (c *countingWriter) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

func (c *countingWriter) Write(p []byte) (int, error) {
	if c.status == 0 {
		c.status = http.StatusOK
	}
	n, err := c.ResponseWriter.Write(p)
	c.bytes += int64(n)
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (c *countingWriter) Unwrap() http.ResponseWriter { return c.ResponseWriter }

// ServeHTTP runs one request through a free worker. Cancelling the request
// context while the application runs stops the Perl code at the next
// opcode.
func (ws *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ws.opts.AccessLog != nil {
		cw := &countingWriter{ResponseWriter: w}
		start := time.Now()
		defer func() { ws.logAccess(r, cw, start) }()
		w = cw
	}
	// Read the body before taking a worker: a slow upload must not hold
	// one.
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, ws.opts.MaxRequestBody))
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}
	var wk *worker
	select {
	case wk = <-ws.workers:
	case <-r.Context().Done():
		http.Error(w, "no worker available", http.StatusServiceUnavailable)
		return
	}
	res := &response{w: w}
	wk.req = res
	err = ws.run(r, wk, body)
	wk.req = nil
	ws.release(wk)

	switch {
	case res.started:
		// The stream is already on the wire; nothing more can be said.
		err = errors.Join(err, res.err)
	case err == nil && res.responded:
		writeHead(w, res.status, res.header)
		_, err = w.Write(res.body)
	default:
		if err == nil {
			err = errors.New("the application returned without a response")
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	if err != nil {
		log.Printf("psgi: %s %s: %v", r.Method, r.URL.Path, err)
	}
}

// run calls the application for one request on wk.
func (ws *Server) run(r *http.Request, wk *worker, body []byte) error {
	ctx := r.Context()
	env, err := wk.p.NewHash(ctx, requestEnv(r, len(body), ws.multi)...)
	if err != nil {
		return fmt.Errorf("build environment: %w", err)
	}
	_, err = wk.handle.Call(ctx, env.Ref(), perl.ValueOf(body))
	return err
}

// release returns a worker to the pool, or a fresh clone in its place once
// it has served MaxRequests requests.
func (ws *Server) release(wk *worker) {
	wk.served++
	if ws.opts.MaxRequests > 0 && wk.served >= ws.opts.MaxRequests {
		fresh, err := ws.cloneWorker()
		if err != nil {
			// Keep serving on the old worker rather than lose capacity.
			log.Printf("psgi: replace worker after %d requests: %v", wk.served, err)
		} else {
			if err := wk.p.Close(); err != nil {
				log.Printf("psgi: close replaced worker: %v", err)
			}
			wk = fresh
		}
	}
	ws.workers <- wk
}

// logAccess writes one line in the Apache combined format.
func (ws *Server) logAccess(r *http.Request, cw *countingWriter, start time.Time) {
	host := r.RemoteAddr
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	status := cw.status
	if status == 0 {
		status = http.StatusOK
	}
	line := fmt.Sprintf("%s - - [%s] \"%s %s %s\" %d %d %q %q\n", host,
		start.Format("02/Jan/2006:15:04:05 -0700"), r.Method, r.URL.RequestURI(), r.Proto,
		status, cw.bytes, r.Referer(), r.UserAgent())
	ws.logMu.Lock()
	_, err := io.WriteString(ws.opts.AccessLog, line)
	ws.logMu.Unlock()
	if err != nil {
		log.Printf("psgi: access log: %v", err)
	}
}

// requestEnv flattens the request into the CGI-style PSGI environment plus
// the psgi.* runtime keys; the handler adds psgi.input, psgi.errors, and
// psgi.version on the Perl side.
func requestEnv(r *http.Request, bodyLen int, multiproc bool) []perl.Pair {
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
		{K: "REQUEST_METHOD", V: perl.ValueOf(r.Method)},
		{K: "SCRIPT_NAME", V: perl.ValueOf("")},
		{K: "PATH_INFO", V: perl.ValueOf(r.URL.Path)},
		{K: "REQUEST_URI", V: perl.ValueOf(r.URL.RequestURI())},
		{K: "QUERY_STRING", V: perl.ValueOf(r.URL.RawQuery)},
		{K: "SERVER_NAME", V: perl.ValueOf(host)},
		{K: "SERVER_PORT", V: perl.ValueOf(port)},
		{K: "SERVER_PROTOCOL", V: perl.ValueOf(r.Proto)},
		{K: "REMOTE_ADDR", V: perl.ValueOf(remote)},
		{K: "CONTENT_LENGTH", V: perl.ValueOf(bodyLen)},
		{K: "psgi.url_scheme", V: perl.ValueOf(scheme)},
		{K: "psgi.multithread", V: perl.ValueOf(false)},
		{K: "psgi.multiprocess", V: perl.ValueOf(multiproc)},
		{K: "psgi.run_once", V: perl.ValueOf(false)},
		{K: "psgi.streaming", V: perl.ValueOf(true)},
		{K: "psgi.nonblocking", V: perl.ValueOf(false)},
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		pairs = append(pairs, perl.Pair{K: "CONTENT_TYPE", V: perl.ValueOf(ct)})
	}
	for k, vs := range r.Header {
		switch k {
		case "Content-Type", "Content-Length":
			continue
		}
		pairs = append(pairs, perl.Pair{
			K: "HTTP_" + strings.ReplaceAll(strings.ToUpper(k), "-", "_"),
			V: perl.ValueOf(strings.Join(vs, ", ")),
		})
	}
	return pairs
}

// Close shuts down every worker and the prototype. Callers stop serving
// first; Close does not wait for in-flight requests.
func (ws *Server) Close() {
	for {
		select {
		case wk := <-ws.workers:
			if err := wk.p.Close(); err != nil {
				log.Printf("psgi: close worker: %v", err)
			}
			continue
		default:
		}
		break
	}
	ws.protoMu.Lock()
	defer ws.protoMu.Unlock()
	if err := ws.proto.p.Close(); err != nil {
		log.Printf("psgi: close prototype: %v", err)
	}
}
