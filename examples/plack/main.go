// Serve a PSGI Perl web application (Plack + Mojolicious) from Go's
// net/http.
//
// All of the serving machinery lives in the go-perl/psgi package:
// psgi.Server is the http.Handler (a fixed set of warm interpreter
// workers, starman/starlet-style — one request per worker at a time), and
// psgi.New verifies and installs the serving contract, compiles the
// application into an instance. This example prepares ONE prototype — the
// host filesystem, the cpanfile's vendored modules on @INC, the `gperl xs
// build` native XS modules, the compiled app — and psgi.New scales
// it to N workers as copy-on-write clones: the loading work never re-runs
// per worker.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/fs"
	"github.com/goccy/go-perl/psgi"
	"github.com/goccy/go-perl/xs"
)

// newPrototype builds the one instance every worker clones: the host
// filesystem, the vendored module tree on @INC, and the native XS modules
// — the instance's own loading facilities do all of it.
func newPrototype(incDirs []string, xsDir string) (*perl.Perl, error) {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		return nil, err
	}
	p, err := perl.New(perl.Config{
		// The example loads its app and vendored modules from the host
		// working tree.
		FS:        fs.NewHostFS(),
		StdlibDir: stdlib,
		// The PSGI app's STDERR (psgi.errors) surfaces on the process stderr.
		Stderr: os.Stderr,
	})
	if err != nil {
		return nil, err
	}
	if err := p.AddInc(context.Background(), incDirs...); err != nil {
		p.Close()
		return nil, err
	}
	// Register the `gperl xs build` output (the cpanfile's XS
	// distributions) before the app compiles anything that uses it.
	if err := p.AddXSDir(xsDir); err != nil {
		p.Close()
		return nil, fmt.Errorf("load native XS modules: %w", err)
	}
	return p, nil
}

// appHandler assembles the workers for the app.psgi in the working
// directory, with the carton-vendored module tree on @INC.
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
		return nil, nil, fmt.Errorf("%s not found - run `make setup` to vendor the cpanfile modules: %w", libDir, err)
	}
	xsDir, err := filepath.Abs(filepath.Join("local", "xs", xs.ArchTag()))
	if err != nil {
		return nil, nil, err
	}
	if _, err := os.Stat(xsDir); err != nil {
		return nil, nil, fmt.Errorf("%s not found - run `make setup` to build the cpanfile's XS modules: %w", xsDir, err)
	}
	start := time.Now()
	proto, err := newPrototype([]string{libDir}, xsDir)
	if err != nil {
		return nil, nil, err
	}
	server, err := psgi.New(proto, size, appPath)
	if err != nil {
		proto.Close()
		return nil, nil, err
	}
	log.Printf("booted %d Perl worker(s) with Plack + Mojolicious in %s (1 prototype + %d clones)", size, time.Since(start).Round(time.Millisecond), size-1)
	return server, server.Close, nil
}

func main() {
	addr := flag.String("addr", ":8091", "listen address")
	size := flag.Int("workers", runtime.GOMAXPROCS(0), "warm Perl workers (max concurrency)")
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
