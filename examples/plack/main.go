// Command plack serves a Perl web application — Mojolicious running as a
// PSGI app — from Go's net/http, with no external perl at runtime.
//
// The program is ordinary go-perl embedding code: nothing in it names a
// CPAN module. The application's dependencies (the cpanfile next to this
// file: Plack, Mojolicious, and the XS distribution Cpanel::JSON::XS) enter
// the build through the go.work that `gperl mod download` writes, so
// `go build` produces one binary that carries them — the pure-Perl modules
// embedded, the XS module linked in statically.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/psgi"
)

// newServer loads app.psgi from the working directory and returns the
// server over size workers. The interpreter runs on its default sandboxed
// filesystem: the dependencies come from the binary, not from disk.
func newServer(size int) (*psgi.Server, error) {
	start := time.Now()
	// The PSGI app's STDERR (psgi.errors) surfaces on the process stderr.
	p, err := perl.New(perl.Config{Stderr: os.Stderr})
	if err != nil {
		return nil, err
	}
	server, err := psgi.New(p, size, "app.psgi")
	if err != nil {
		if cerr := p.Close(); cerr != nil {
			log.Printf("close interpreter: %v", cerr)
		}
		return nil, err
	}
	log.Printf("booted %d Perl worker(s) with Plack + Mojolicious in %s", size, time.Since(start).Round(time.Millisecond))
	return server, nil
}

func main() {
	addr := flag.String("addr", ":8091", "listen address")
	size := flag.Int("workers", runtime.GOMAXPROCS(0), "warm Perl workers (max concurrency)")
	flag.Parse()

	server, err := newServer(*size)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Printf("serving the PSGI app on http://localhost%s", *addr)
	if err := http.ListenAndServe(*addr, server); err != nil {
		log.Fatal(err)
	}
}
