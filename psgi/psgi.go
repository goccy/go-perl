// Package psgi serves PSGI Perl web applications from Go's net/http.
//
// PSGI specifies the application: the .psgi file's last evaluated value is
// a code reference that takes the environment hashref (CGI-style keys plus
// psgi.*) and returns [$status, \@headers, $body] — or, for a delayed or
// streaming response, a code reference that is handed a responder. This
// package supplies the SERVER side of that interface: the environment, the
// request/response plumbing, and the worker model.
//
// Server is a fixed set of warm interpreter workers (the starman/starlet
// serving model, with interpreters in place of processes). Prepare ONE
// instance with the instance's own loading facilities; New compiles the
// application into it and every worker is a copy-on-write clone of it
// (perl.Clone), so the loading work runs once no matter the worker count:
//
//	p, _ := perl.New(perl.Config{})
//	server, _ := psgi.New(p, 4, "app.psgi")
//	server.ListenAndServe(":8091") // or http.ListenAndServe(addr, server)
//
// A request holds a worker only while the application runs: the request
// body is read before a worker is taken, and a complete response is
// written after the worker is returned, so slow clients do not occupy
// workers. A streaming response holds its worker until the stream ends.
//
// A *perl.Perl must not run concurrent requests; Server guarantees that.
package psgi

import (
	"net/http"

	perl "github.com/goccy/go-perl"
	impl "github.com/goccy/go-perl/internal/psgi"
	// Serving stacks routinely carry native XS modules; linking the
	// loader here makes (*perl.Perl).AddXSDir work without application
	// code importing it.
	_ "github.com/goccy/go-perl/xs"
)

// Server serves PSGI traffic over a fixed set of warm Perl workers — the
// role starman/starlet's preforked workers play, with interpreters instead
// of processes: each request checks a worker out, runs through the
// application, and returns it. Goroutine-per-request on the Go side, never
// more than one request per worker at a time; the worker count is the max
// concurrency. A Server is an http.Handler, so it plugs into any Go HTTP
// stack; ListenAndServe is the shortcut for the plain case.
type Server struct {
	impl *impl.Server
}

// New loads the PSGI application at appPath into the PREPARED prototype
// (module trees on @INC, native XS modules — anything every worker should
// share) and serves it with workers workers, each a copy-on-write clone of
// the prototype (perl.Clone), so none of the loading work re-runs per
// worker and the workers share the read-only bulk of their memory. The
// prototype is owned by the returned Server (Close closes it). On any
// failure everything built so far is closed.
//
// The file must satisfy the PSGI contract — its last evaluated value is a
// code reference taking $env — and anything else fails New, before any
// worker exists and before anything serves.
func New(prototype *perl.Perl, workers int, appPath string) (*Server, error) {
	s, err := impl.NewServer(prototype, workers, appPath, impl.Options{})
	if err != nil {
		return nil, err
	}
	return &Server{impl: s}, nil
}

// ListenAndServe serves the application on addr until the listener fails.
// For anything beyond the plain case (TLS, timeouts, middleware), use the
// Server as the http.Handler it is.
func (ws *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, ws)
}

// ServeHTTP runs one request through a free worker. Cancelling the request
// context while the application runs stops the Perl code at the next
// opcode.
func (ws *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { ws.impl.ServeHTTP(w, r) }

// Close shuts down every worker and the prototype. Callers stop serving
// first; Close does not wait for in-flight requests.
func (ws *Server) Close() { ws.impl.Close() }
