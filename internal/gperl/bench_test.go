//go:build unix

package gperl_test

// Benchmarks of gperl against perl: the PSGI server (gperl's against
// Starman and Starlet on the host's perl), an interpreter's start-up, and a
// test suite's run. `make bench` installs the perl side into .bench/perl5
// and runs them; a perl-side benchmark without what it needs is skipped.
//
// Every server is a process of its own serving a listening socket the
// benchmark has made and handed over the way Server::Starter does
// (SERVER_STARTER_PORT and an inherited descriptor), so the first request
// waits in the socket's queue until the server accepts it.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/goccy/go-perl/internal/gperl"
)

// benchPerl5Env names the directory of the perl-side modules (lib/perl5,
// bin/plackup), as cpanm -L lays them out.
const benchPerl5Env = "GPERL_BENCH_PERL5"

// benchApp is a PSGI application both servers run.
type benchApp struct {
	name     string
	modules  []string // what the host's perl needs besides Plack
	cpanfile string
	psgi     string
}

var benchApps = []benchApp{
	{name: "hello", psgi: "sub { [200, ['Content-Type' => 'text/plain'], ['Hello, World']] };\n"},
	{name: "mojolicious", modules: []string{"Mojolicious"}, cpanfile: "requires 'Mojolicious';\n",
		psgi: "use Mojolicious::Lite;\nget '/' => { text => 'Hello, World' };\napp->start;\n"},
	{name: "xs-json", modules: []string{"Cpanel::JSON::XS"}, cpanfile: "requires 'Cpanel::JSON::XS';\n",
		psgi: "use Cpanel::JSON::XS;\nmy $json = Cpanel::JSON::XS->new->canonical;\n" +
			"sub { [200, ['Content-Type' => 'application/json'], [$json->encode({ id => 1, name => 'gperl', list => [1 .. 20] })]] };\n"},
}

// benchConcurrency are the numbers of clients each server is measured
// with, every client on a connection of its own (keep-alive).
var benchConcurrency = []int{8, 32}

// BenchmarkPSGI measures requests per second of each application on
// gperl's server and on Starman and Starlet, with as many workers as the
// machine has CPUs.
func BenchmarkPSGI(b *testing.B) {
	workers := strconv.Itoa(runtime.GOMAXPROCS(0))
	for _, app := range benchApps {
		dir := b.TempDir()
		appPath := filepath.Join(dir, "app.psgi")
		writeBenchFile(b, appPath, app.psgi)
		if app.cpanfile != "" {
			writeBenchFile(b, filepath.Join(dir, "cpanfile"), app.cpanfile)
		}
		b.Run(app.name+"/gperl", func(b *testing.B) {
			b.Setenv("GPERL_DEV_REPLACE", devReplaces(b))
			server := filepath.Join(b.TempDir(), "server")
			if err := gperl.PSGIBuild([]string{"-o", server, appPath}); err != nil {
				b.Fatal(err)
			}
			ln := benchListener(b)
			stop := startServer(b, os.Environ(), ln, server, "--workers", workers, "-E", "deployment")
			defer stop()
			for _, conc := range benchConcurrency {
				b.Run(fmt.Sprintf("c%d", conc), func(b *testing.B) { benchLoad(b, ln.Addr().String(), conc) })
			}
		})
		// Starlet keeps a connection for one request unless told otherwise;
		// every server keeps its clients' connections here.
		for _, srv := range []struct {
			name  string
			flags []string
		}{
			{"starman", []string{"--workers", workers}},
			{"starlet", []string{"--max-workers", workers, "--max-keepalive-reqs", strconv.Itoa(math.MaxInt32)}},
		} {
			b.Run(app.name+"/"+srv.name, func(b *testing.B) {
				plackup, env := benchPerlSide(b, append([]string{"Plack", srvModule(srv.name), "Server::Starter"}, app.modules...)...)
				if srv.name == "starman" {
					// Starman serves a socket from Server::Starter through it.
					requirePerlModules(b, env, "Net::Server::SS::PreFork")
				}
				ln := benchListener(b)
				argv := append([]string{"perl", plackup, "-s", srvModule(srv.name)}, srv.flags...)
				stop := startServer(b, env, ln, append(argv, "-E", "deployment", appPath)...)
				defer stop()
				for _, conc := range benchConcurrency {
					b.Run(fmt.Sprintf("c%d", conc), func(b *testing.B) { benchLoad(b, ln.Addr().String(), conc) })
				}
			})
		}
	}
}

func srvModule(server string) string {
	return map[string]string{"starman": "Starman", "starlet": "Starlet"}[server]
}

// benchLoad sends b.N requests to addr from conc clients and reports the
// requests per second. Every response must be a 200.
func benchLoad(b *testing.B, addr string, conc int) {
	transport := &http.Transport{MaxIdleConnsPerHost: conc, MaxConnsPerHost: conc}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport}
	url := "http://" + addr + "/"
	get := func() error {
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		_, err = io.Copy(io.Discard, resp.Body)
		err = errors.Join(err, resp.Body.Close())
		if err == nil && resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("status %s", resp.Status)
		}
		return err
	}
	run := func(n int64) error {
		var (
			next atomic.Int64
			wg   sync.WaitGroup
			mu   sync.Mutex
			errs error
		)
		for range conc {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for next.Add(1) <= n {
					if err := get(); err != nil {
						mu.Lock()
						errs = errors.Join(errs, err)
						mu.Unlock()
						return
					}
				}
			}()
		}
		wg.Wait()
		return errs
	}
	// Every client's connection is open and every worker has served before
	// the measurement.
	if err := run(int64(conc * 4)); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	if err := run(int64(b.N)); err != nil {
		b.Fatal(err)
	}
	b.StopTimer()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "req/s")
}

// benchListener listens on a port of the loopback interface.
func benchListener(b *testing.B) *net.TCPListener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	return ln.(*net.TCPListener)
}

// benchPerlSide returns plackup and the environment perl runs with for the
// perl-side benchmarks, skipping the benchmark when modules are missing.
func benchPerlSide(b *testing.B, modules ...string) (plackup string, env []string) {
	env = os.Environ()
	plackup, err := exec.LookPath("plackup")
	if dir := os.Getenv(benchPerl5Env); dir != "" {
		plackup, err = filepath.Join(dir, "bin", "plackup"), nil
		env = append(env, "PERL5LIB="+filepath.Join(dir, "lib", "perl5"))
	}
	if err != nil {
		b.Skipf("no plackup (make bench installs it): %v", err)
	}
	requirePerlModules(b, env, modules...)
	return plackup, env
}

// requirePerlModules skips the benchmark unless the host's perl loads every
// module with env.
func requirePerlModules(b *testing.B, env []string, modules ...string) {
	for _, m := range modules {
		cmd := exec.Command("perl", "-M"+m, "-e", "1")
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			b.Skipf("the host's perl has no %s (make bench installs it): %v\n%s", m, err, out)
		}
	}
}

// startServer starts the server argv runs on ln, in a process group of its
// own, and returns the function that ends the group.
func startServer(b *testing.B, env []string, ln *net.TCPListener, argv ...string) (stop func()) {
	f, err := ln.File()
	if err != nil {
		b.Fatal(err)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = append(env, "SERVER_STARTER_PORT="+ln.Addr().String()+"=3")
	cmd.ExtraFiles = []*os.File{f}
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		b.Fatal(err)
	}
	// The server has the socket; this process's copies go.
	if err := errors.Join(f.Close(), ln.Close()); err != nil {
		b.Fatal(err)
	}
	return func() {
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil {
			b.Error(err)
		}
		var exit *exec.ExitError
		if err := cmd.Wait(); err != nil && !errors.As(err, &exit) {
			b.Error(err)
		}
	}
}

func writeBenchFile(b *testing.B, path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatal(err)
	}
}

// benchGperl builds the gperl command into dir.
func benchGperl(b *testing.B, dir string) string {
	out := filepath.Join(dir, "gperl")
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", out, "../../cmd/gperl")
	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("go build: %v\n%s", err, output)
	}
	return out
}

// BenchmarkStartup measures running an empty program: gperl run -e 1
// against perl -e 1.
func BenchmarkStartup(b *testing.B) {
	gperl := benchGperl(b, b.TempDir())
	for _, c := range []struct {
		name string
		argv []string
	}{{"gperl", []string{gperl, "run", "-e", "1"}}, {"perl", []string{"perl", "-e", "1"}}} {
		b.Run(c.name, func(b *testing.B) {
			for b.Loop() {
				if out, err := exec.Command(c.argv[0], c.argv[1:]...).CombinedOutput(); err != nil {
					b.Fatalf("%v: %v\n%s", c.argv, err, out)
				}
			}
		})
	}
}

// benchTestFile is one file of the suite BenchmarkTestSuite runs: a little
// of everything a test does — loading Test::More, computing, matching.
const benchTestFile = `use strict;
use warnings;
use Test::More;
my $sum = 0;
$sum += $_ for 1 .. 200_000;
is($sum, 20_000_100_000, 'sum');
my %seen = map { ($_ => $_ * 2) } 1 .. 2_000;
is(scalar(keys %seen), 2_000, 'hash');
like(join(',', map { $seen{$_} } 1 .. 10), qr/^2,4,6/, 'join');
done_testing;
`

// BenchmarkTestSuite measures running a suite of 16 test files, one at a
// time and as many at once as the machine has CPUs: gperl test (each file
// a process, and --in-process) against prove.
func BenchmarkTestSuite(b *testing.B) {
	gperl := benchGperl(b, b.TempDir())
	dir := b.TempDir()
	for i := range 16 {
		writeBenchFile(b, filepath.Join(dir, "t", fmt.Sprintf("%02d.t", i)), benchTestFile)
	}
	for _, jobs := range []string{"1", strconv.Itoa(runtime.GOMAXPROCS(0))} {
		for _, c := range []struct {
			name string
			argv []string
		}{
			{"gperl", []string{gperl, "test", "-j", jobs, "t"}},
			{"gperl-in-process", []string{gperl, "test", "--in-process", "-j", jobs, "t"}},
			{"prove", []string{"prove", "-j", jobs, "t"}},
		} {
			b.Run(c.name+"/j"+jobs, func(b *testing.B) {
				for b.Loop() {
					cmd := exec.Command(c.argv[0], c.argv[1:]...)
					cmd.Dir = dir
					if out, err := cmd.CombinedOutput(); err != nil {
						b.Fatalf("%v: %v\n%s", c.argv, err, out)
					}
				}
			})
		}
	}
}
