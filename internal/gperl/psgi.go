package gperl

// gperl psgi: serve a PSGI application from Go's net/http, one interpreter
// per worker goroutine — the application server and the front end that a
// Perl deployment usually splits between starman and nginx, in one process
// (or, with gperl psgi build, one binary).
//
// The settings follow plackup and starman by name. A setting is taken from
// the command line, else from the environment (GPERL_PSGI_<OPTION>, and
// PORT), else from the defaults a binary was built with.

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	_ "unsafe" // go:linkname

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal/psgi"
)

// multiFlag is a repeatable option. Values given by one source (defaults,
// environment, command line) replace those of the sources before it.
type multiFlag struct {
	values []string
	fresh  bool // the next Set starts a new source
}

func (m *multiFlag) String() string { return strings.Join(m.values, ",") }

func (m *multiFlag) Set(v string) error {
	if m.fresh {
		m.values, m.fresh = nil, false
	}
	m.values = append(m.values, v)
	return nil
}

// psgiConfig is a PSGI server's settings.
type psgiConfig struct {
	listen           multiFlag // host:port, :port, or a unix socket path
	host             string
	port             int
	env              string    // PLACK_ENV
	inc              multiFlag // -I
	modules          multiFlag // -M
	workers          int
	maxRequests      int
	maxRequestBody   int64
	enableSSL        bool
	sslCert          string
	sslKey           string
	readTimeout      int // seconds
	keepaliveTimeout int // seconds
	accessLog        string
	reload           bool
	// starterPorts is SERVER_STARTER_PORT: the sockets a supervisor
	// (Server::Starter's start_server) has opened for the server, as
	// "address=descriptor" entries separated by ';'.
	starterPorts string
}

// flags declares the options on a flag set. plackup's -o (host) is left
// out where the command has its own -o.
func (c *psgiConfig) flags(name string, shortHost bool) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	for _, n := range []string{"l", "listen"} {
		fs.Var(&c.listen, n, "address to listen on: host:port, :port, or a unix socket path (repeatable)")
	}
	hostNames := []string{"host"}
	if shortHost {
		hostNames = append(hostNames, "o")
	}
	for _, n := range hostNames {
		fs.StringVar(&c.host, n, "", "address to bind (all interfaces by default)")
	}
	for _, n := range []string{"p", "port"} {
		fs.IntVar(&c.port, n, 5000, "port to listen on")
	}
	for _, n := range []string{"E", "env"} {
		fs.StringVar(&c.env, n, "deployment", "the application environment (PLACK_ENV)")
	}
	fs.Var(&c.inc, "I", "directory to add to @INC (repeatable)")
	fs.Var(&c.modules, "M", "module to load before the application (repeatable)")
	fs.IntVar(&c.workers, "workers", runtime.GOMAXPROCS(0), "number of interpreter workers (the maximum concurrency)")
	fs.IntVar(&c.maxRequests, "max-requests", 0, "replace a worker after it served this many requests (0: never)")
	fs.Int64Var(&c.maxRequestBody, "max-request-body", 0, "largest request body accepted, in bytes (0: 32 MiB)")
	fs.BoolVar(&c.enableSSL, "enable-ssl", false, "serve TLS (with HTTP/2); needs --ssl-cert and --ssl-key")
	fs.StringVar(&c.sslCert, "ssl-cert", "", "TLS certificate file")
	fs.StringVar(&c.sslKey, "ssl-key", "", "TLS private key file")
	fs.IntVar(&c.readTimeout, "read-timeout", 60, "seconds allowed for reading a request (0: no limit)")
	fs.IntVar(&c.keepaliveTimeout, "keepalive-timeout", 60, "seconds an idle keep-alive connection stays open (0: no limit)")
	fs.StringVar(&c.accessLog, "access-log", "", "file to write the access log to (\"-\" for standard error)")
	for _, n := range []string{"r", "reload"} {
		fs.BoolVar(&c.reload, n, false, "reload the application when its files change (gperl psgi run only)")
	}
	return fs
}

// parsePSGI resolves the settings from, in rising priority, the built-in
// defaults, the defaults a binary was built with, the environment, and the
// command line. It returns the arguments left after the options.
func parsePSGI(name string, defaults, args []string, getenv func(string) string) (*psgiConfig, []string, error) {
	c := &psgiConfig{}
	fs := c.flags(name, true)
	fs.SetOutput(io.Discard)
	newSource := func() { c.listen.fresh, c.inc.fresh, c.modules.fresh = true, true, true }

	newSource()
	if err := fs.Parse(defaults); err != nil {
		return nil, nil, fmt.Errorf("built-in defaults: %w", err)
	}
	newSource()
	if port := getenv("PORT"); port != "" {
		if err := fs.Set("port", port); err != nil {
			return nil, nil, fmt.Errorf("PORT: %w", err)
		}
	}
	var envErr error
	fs.VisitAll(func(f *flag.Flag) {
		key := "GPERL_PSGI_" + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		value := getenv(key)
		if value == "" || len(f.Name) == 1 && f.Name != "I" && f.Name != "M" {
			return
		}
		values := []string{value}
		if _, repeatable := f.Value.(*multiFlag); repeatable {
			values = strings.Split(value, ",")
		}
		for _, v := range values {
			if err := fs.Set(f.Name, v); err != nil && envErr == nil {
				envErr = fmt.Errorf("%s: %w", key, err)
			}
		}
	})
	if envErr != nil {
		return nil, nil, envErr
	}
	newSource()
	if err := fs.Parse(splitShortOptions(args, "r", "IMpolE")); err != nil {
		return nil, nil, err
	}
	if c.workers < 1 {
		return nil, nil, fmt.Errorf("--workers must be at least 1")
	}
	c.starterPorts = getenv("SERVER_STARTER_PORT")
	if c.enableSSL && (c.sslCert == "" || c.sslKey == "") {
		return nil, nil, fmt.Errorf("--enable-ssl needs --ssl-cert and --ssl-key")
	}
	return c, fs.Args(), nil
}

// listeners opens what the settings ask to listen on.
func (c *psgiConfig) listeners() ([]net.Listener, error) {
	if c.starterPorts != "" {
		return starterListeners(c.starterPorts)
	}
	addrs := c.listen.values
	if len(addrs) == 0 {
		addrs = []string{net.JoinHostPort(c.host, fmt.Sprint(c.port))}
	}
	var out []net.Listener
	for _, addr := range addrs {
		network := "tcp"
		if strings.ContainsRune(addr, '/') {
			network = "unix"
			if err := os.Remove(addr); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
		}
		ln, err := net.Listen(network, addr)
		if err != nil {
			for _, open := range out {
				err = errors.Join(err, open.Close())
			}
			return nil, err
		}
		out = append(out, ln)
	}
	return out, nil
}

// starterListeners takes over the listening sockets SERVER_STARTER_PORT
// names, as Starman and Starlet do under Server::Starter.
func starterListeners(ports string) ([]net.Listener, error) {
	var out []net.Listener
	fail := func(err error) ([]net.Listener, error) {
		for _, ln := range out {
			err = errors.Join(err, ln.Close())
		}
		return nil, err
	}
	for _, entry := range strings.Split(ports, ";") {
		addr, fdText, ok := strings.Cut(entry, "=")
		if !ok {
			return fail(fmt.Errorf("SERVER_STARTER_PORT: %q is not address=descriptor", entry))
		}
		fd, err := strconv.Atoi(fdText)
		if err != nil {
			return fail(fmt.Errorf("SERVER_STARTER_PORT: %q: %w", entry, err))
		}
		f := os.NewFile(uintptr(fd), addr)
		ln, err := net.FileListener(f)
		if ln != nil {
			out = append(out, ln)
		}
		// FileListener has a descriptor of its own.
		if err = errors.Join(err, f.Close()); err != nil {
			return fail(fmt.Errorf("SERVER_STARTER_PORT: %s: %w", entry, err))
		}
	}
	return out, nil
}

// psgiApp is where a PSGI application lives and what it loads from.
type psgiApp struct {
	path   string   // the application file
	dir    string   // its working directory
	inc    []string // module directories ahead of the dependencies
	depInc []string // the dependencies' directories (projectDeps.inc)
	// native says whether the dependencies have native modules to load
	// from depInc.
	native bool
}

// newServer loads the application into a fresh prototype and returns the
// server over its workers.
func (a *psgiApp) newServer(c *psgiConfig, accessLog io.Writer) (*psgi.Server, error) {
	cfg, err := HostConfig()
	if err != nil {
		return nil, err
	}
	// The descriptors the process was started with are the server's.
	cfg.InheritDescriptors = false
	// Workers load native modules from @INC only when the dependencies have
	// some, and then from any directory of @INC; the native modules linked
	// into a binary are loaded without it (the xs package's loadLinked).
	cfg.NativeModules = a.native
	cfg.Stdin, cfg.Stdout, cfg.Stderr = os.Stdin, os.Stdout, os.Stderr
	cfg.Env = append(os.Environ(), "PLACK_ENV="+c.env)
	p, err := newWorkerPerl(cfg)
	if err != nil {
		return nil, err
	}
	fail := func(err error) (*psgi.Server, error) { return nil, errors.Join(err, p.Close()) }
	ctx := context.Background()
	if err := p.AddInc(ctx, append(append(append([]string{}, c.inc.values...), a.inc...), a.depInc...)...); err != nil {
		return fail(err)
	}
	// The application's relative paths (templates, static files) are
	// relative to its directory.
	if err := p.Bind("__gperl_psgi_dir", func([]perl.Value) ([]perl.Value, error) {
		return []perl.Value{perl.ValueOf(a.dir)}, nil
	}); err != nil {
		return fail(err)
	}
	// The command line is the server's: the application finds no arguments
	// of its own, as under plackup, which takes its options off @ARGV.
	boot := "@ARGV = ();\nmy $dir = __gperl_psgi_dir(); chdir $dir or die \"chdir $dir: $!\\n\";\n"
	for _, module := range c.modules.values {
		name, args, hasArgs := strings.Cut(module, "=")
		if !perlModuleName.MatchString(name) {
			return fail(fmt.Errorf("-M %s: not a module name", module))
		}
		boot += "use " + name
		if hasArgs {
			boot += " split(/,/, q\x00" + strings.ReplaceAll(args, "\x00", "") + "\x00)"
		}
		boot += ";\n"
	}
	r, err := p.Eval(ctx, boot+"1;")
	if err == nil && r.Error != nil {
		err = r.Error
	}
	if err != nil {
		return fail(err)
	}
	srv, err := psgi.NewServer(p, c.workers, a.path, psgi.Options{
		MaxRequests: c.maxRequests, MaxRequestBody: c.maxRequestBody, AccessLog: accessLog,
	})
	if err != nil {
		return fail(err)
	}
	return srv, nil
}

// servePSGI serves the application until SIGINT or SIGTERM, then lets the
// requests in flight finish.
func servePSGI(c *psgiConfig, app *psgiApp) (err error) {
	var accessLog io.Writer
	switch c.accessLog {
	case "":
	case "-":
		accessLog = os.Stderr
	default:
		f, err := os.OpenFile(c.accessLog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, f.Close()) }()
		accessLog = f
	}
	first, err := app.newServer(c, accessLog)
	if err != nil {
		return err
	}
	var current atomic.Pointer[psgi.Server]
	current.Store(first)
	defer func() { current.Load().Close() }()

	listeners, err := c.listeners()
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { current.Load().ServeHTTP(w, r) }),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       time.Duration(c.readTimeout) * time.Second,
		IdleTimeout:       time.Duration(c.keepaliveTimeout) * time.Second,
	}
	scheme := "http"
	if c.enableSSL {
		scheme = "https"
	}
	failed := make(chan error, len(listeners))
	for _, ln := range listeners {
		fmt.Fprintf(os.Stderr, "gperl: serving %s on %s://%s (%d workers)\n", filepath.Base(app.path), scheme, ln.Addr(), c.workers)
		go func() {
			if c.enableSSL {
				failed <- srv.ServeTLS(ln, c.sslCert, c.sslKey)
			} else {
				failed <- srv.Serve(ln)
			}
		}()
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if c.reload {
		go reloadOnChange(ctx, c, app, accessLog, &current)
	}
	select {
	case err := <-failed:
		return err
	case <-ctx.Done():
	}
	fmt.Fprintln(os.Stderr, "gperl: shutting down")
	shutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return srv.Shutdown(shutdown)
}

// reloadOnChange swaps in a freshly loaded application whenever a file it
// is made of changes. A version that fails to load is reported and the
// running one keeps serving.
func reloadOnChange(ctx context.Context, c *psgiConfig, app *psgiApp, accessLog io.Writer, current *atomic.Pointer[psgi.Server]) {
	watched := append([]string{app.path}, app.inc...)
	seen := sourceStamp(watched)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		now := sourceStamp(watched)
		if now == seen {
			continue
		}
		seen = now
		next, err := app.newServer(c, accessLog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gperl: reload failed, still serving the previous version: %v\n", err)
			continue
		}
		old := current.Swap(next)
		fmt.Fprintln(os.Stderr, "gperl: reloaded", filepath.Base(app.path))
		// Requests already running on the old workers finish first.
		go func() {
			time.Sleep(5 * time.Second)
			old.Close()
		}()
	}
}

// sourceStamp summarizes the files under paths — how many there are and
// the newest modification time — so that any edit, addition, or removal
// changes it. A path that cannot be read contributes its error text, which
// changes the stamp just the same.
func sourceStamp(paths []string) string {
	var (
		newest time.Time
		count  int
		failed []string
	)
	for _, root := range paths {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != root && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			count++
			if info.ModTime().After(newest) {
				newest = info.ModTime()
			}
			return nil
		})
		if err != nil {
			failed = append(failed, err.Error())
		}
	}
	return fmt.Sprintf("%d %d %s", count, newest.UnixNano(), strings.Join(failed, ";"))
}

// perlModuleName matches a Perl package name.
var perlModuleName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(::[A-Za-z0-9_]+)*$`)

// projectApp locates the application file of a project and syncs the
// project's dependencies for this process.
func projectApp(file string) (*psgiApp, error) {
	path, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	app := &psgiApp{path: path, dir: filepath.Dir(path)}
	deps, err := hostDeps(app.dir)
	if err == nil {
		app.depInc, app.native = deps.inc(), len(deps.nativeDirs) > 0
	}
	if lib := filepath.Join(app.dir, "lib"); dirExists(lib) {
		app.inc = append(app.inc, lib)
	}
	return app, err
}

// PSGIRun implements `gperl psgi run [options] app.psgi`: it serves the
// application in this process, on the project's dependencies.
func PSGIRun(args []string) error {
	c, rest, err := parsePSGI("gperl psgi run", nil, args, os.Getenv)
	if err != nil {
		return err
	}
	if len(rest) != 1 {
		return errors.New("usage: gperl psgi run [options] app.psgi")
	}
	app, err := projectApp(rest[0])
	if err != nil {
		return err
	}
	if err := servePSGI(c, app); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// PSGIBuild implements `gperl psgi build [-o out] [options] app.psgi`: it
// builds a server binary that carries the application's directory, its
// dependencies, and the interpreter. The options become the binary's
// defaults.
func PSGIBuild(args []string) error {
	c := &psgiConfig{}
	fs := c.flags("gperl psgi build", false)
	out := fs.String("o", "", "output binary path")
	if err := fs.Parse(splitShortOptions(args, "r", "IMplE")); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: gperl psgi build [-o out] [options] app.psgi")
	}
	// Never nil: no defaults is still a server (buildBinary).
	defaults := []string{}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "o" {
			return
		}
		if m, ok := f.Value.(*multiFlag); ok {
			for _, v := range m.values {
				defaults = append(defaults, "--"+f.Name+"="+v)
			}
			return
		}
		defaults = append(defaults, "--"+f.Name+"="+f.Value.String())
	})
	return buildBinary(fs.Arg(0), *out, defaults)
}

// PSGIMain is the entry point of a server binary gperl psgi build makes:
// it serves the application at appPath, with incDirs ahead of the
// dependencies on @INC, under the settings from defaults (given at build
// time), the environment, and args. It returns the process exit status.
func PSGIMain(appPath string, incDirs, defaults, args []string) int {
	c, rest, err := parsePSGI(filepath.Base(os.Args[0]), defaults, args, os.Getenv)
	if err == nil && len(rest) > 0 {
		err = fmt.Errorf("unexpected argument %q", rest[0])
	}
	if err == nil {
		c.reload = false
		err = servePSGI(c, &psgiApp{path: appPath, dir: filepath.Dir(appPath), inc: incDirs})
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		if errors.Is(err, flag.ErrHelp) {
			c := &psgiConfig{}
			fs := c.flags(filepath.Base(os.Args[0]), true)
			fs.SetOutput(os.Stderr)
			fs.PrintDefaults()
			return 2
		}
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		return 1
	}
	return 0
}

// newWorkerPerl creates the prototype of the workers: a program with the
// host process's identity and numbers whose working directory, environment
// and pid are its own, so that what one request changes of them reaches no
// other worker (its clones are the same).
//
//go:linkname newWorkerPerl github.com/goccy/go-perl.newOwnState
func newWorkerPerl(cfg perl.Config) (*perl.Perl, error)
