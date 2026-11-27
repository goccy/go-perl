# go-perl

[![CI](https://github.com/goccy/go-perl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/goccy/go-perl/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/goccy/go-perl.svg)](https://pkg.go.dev/github.com/goccy/go-perl)

**Perl 5 in pure Go — embed and run Perl anywhere Go runs. No cgo, no external
`perl`, one static binary.**

The interpreter is
[Perl 5.44.0 compiled to WebAssembly](https://github.com/goccy/perl-wasm) and
then [translated to Go](https://github.com/goccy/wasm2go) — no wasm runtime is
involved at run time — published as the
[`perlwasm2go`](https://github.com/goccy/perlwasm2go) module this package
builds on. The pure-Perl standard library ships inside the package as an
embedded zip, so a `go build` produces a single self-contained binary that can
run Perl.

```go
package main

import (
	"context"
	"fmt"

	perl "github.com/goccy/go-perl"
)

func main() {
	p, err := perl.New(perl.Config{})
	if err != nil {
		panic(err)
	}
	defer p.Close()

	r, err := p.Eval(context.Background(), `join(",", map { $_ * 2 } 1..5)`)
	if err != nil {
		panic(err)
	}
	s, _ := perl.As[perl.ScalarValue](r.Value)
	fmt.Println(s.String()) // 2,4,6,8,10
}
```

## Highlights

- **Pure Go, cross-compile anywhere**: the interpreter is ordinary Go code —
  no cgo, no shared libraries, `CGO_ENABLED=0` friendly. Plain
  `GOOS`/`GOARCH` cross compilation produces a single static binary for any
  target Go supports; CI cross-builds every package for linux, macOS,
  Windows, and even wasip1.
- **Batteries included**: libperl plus every static XS extension (List::Util,
  POSIX, Socket, re, Storable, Encode, ...) and the pure-Perl stdlib are
  embedded in the binary (served from a private in-memory filesystem by
  default) — nothing to install or deploy next to your executable.
- **Native XS support, unmodified modules**: real CPAN XS distributions
  compile against go-perl's XS SDK with their own `Makefile.PL`/`Build.PL`
  — no source changes. Moose, Text::Xslate, Devel::NYTProf, DBD::mysql,
  PerlIO::utf8_strict, and keyword plugins like Syntax::Keyword::Match run
  as-is (exercised by the test suite). The embedded interpreter drives the
  whole build, so no host `perl` is needed even at build time.
- **XS in one binary**: a built XS module is one object whose only global
  symbol is its entry table, so any number of them link *statically* into a
  Go binary — `go build` produces a single file that carries the
  interpreter, the pure-Perl modules, and the XS modules, and on Linux it
  can be a fully static executable. Without cgo the same modules are
  embedded as shared libraries and loaded at run time instead. Native
  modules are available on linux/amd64, linux/arm64, and darwin/arm64.
- **Go and Perl call each other in-process**: because the interpreter *is*
  Go code, the two languages meet at the Go ABI — crossing the boundary is
  a function call, not FFI or IPC. `Call` invokes named Perl subs from Go,
  `Bind`/`BindClass` expose Go functions to Perl as ordinary subs (which may
  call back into Perl, so round trips compose). Every value crosses typed —
  see [Calling between Go and Perl](#calling-between-go-and-perl).
- **PSGI**: the [`psgi`](./psgi) package serves PSGI applications straight
  from `net/http`, so an existing Plack app runs on a Go HTTP server —
  see [`examples/plack`](./examples/plack) for a Mojolicious app carrying
  real traffic over a pool of warm workers.
- **Many interpreters, shared memory**: the first `New` boots one
  interpreter and snapshots its memory; every later `New` maps that snapshot
  copy-on-write, so instances start without re-running interpreter init.
  `Clone` does the same from a *live* instance's current state — modules
  compiled once are inherited by every clone. Writable state is fully
  private per instance; the read-only bulk of memory is shared, so one
  process can hold many interpreters cheaply.
- **Sandboxed by default, capability-controlled**: the translation to Go
  keeps the wasm linear-memory model and the WASI syscall boundary, so every
  filesystem, network, and process access funnels through hooks you control.
  Each `Perl` runs against a PRIVATE in-memory filesystem unless configured
  otherwise — `Config.FS` selects the backend from the [`fs`](./fs) package
  (`fs.NewHostFS()` passes through to the OS, matching `perl`; `fs.DirFS`
  scopes to one directory; `fs.NewMemFS()` or any custom backend plugs in
  the same way). The capability hooks are FAIL-CLOSED: with a zero `Config`,
  outbound connections (`Dial`), name resolution (`Resolve`), subprocess
  spawns (`Exec`), and `fork` (`Fork`) are all denied — each capability is
  granted explicitly.
- **Cancellable**: cancelling the `context.Context` passed to `Eval`/`Call`
  stops a runaway script at the next Perl opcode.
- **Go-style tooling for Perl**: the [`gperl`](./cmd/gperl) command manages
  CPAN dependencies the way `go` manages modules (`gperl get`,
  `gperl mod tidy`), runs scripts (`gperl run`, with `perl(1)` switches)
  and test suites (`gperl test`), builds self-contained binaries
  (`gperl build`), and serves PSGI applications or builds their server
  binary (`gperl psgi run`, `gperl psgi build`).
- **Supply-chain verified**: the Perl-derived artifacts are consumed as
  attested release binaries, verified against SLSA build provenance on every
  refresh and re-verified in CI — see
  [Supply-chain verification](#supply-chain-verification).

## Calling between Go and Perl

```go
p, _ := perl.New(perl.Config{})
defer p.Close()
ctx := context.Background()

// Go -> Perl: call a named sub. Arguments and results are typed Values.
p.Eval(ctx, `sub add { my ($a, $b) = @_; $a + $b } 1;`)
sum, _ := p.Call(ctx, "add", perl.ValueOf(40), perl.ValueOf(2))
n, _ := perl.As[perl.ScalarValue](sum[0])
fmt.Println(n.Int()) // 42

// Perl objects cross as handles, not copies: the same object, its methods,
// and its state remain live on the Go side.
p.Eval(ctx, `package Counter; sub new { bless {n=>0}, shift } sub inc { $_[0]{n}++ }
             package main; sub counter { our $c ||= Counter->new } 1;`)
res, _ := p.Call(ctx, "counter")
obj, _ := perl.As[perl.RefValue](res[0]) // Class() == ("Counter", true)
obj.MethodCall(ctx, "inc")            // mutates the object Perl sees
p.Call(ctx, "counter")                // returns an Equal handle: same object

// Perl -> Go: bind a Go function as a Perl sub.
p.Bind("go_upper", func(args []perl.Value) ([]perl.Value, error) {
	s, err := perl.As[perl.ScalarValue](args[0])
	if err != nil {
		return nil, err
	}
	return []perl.Value{perl.ValueOf(strings.ToUpper(s.String()))}, nil
})
r, _ := p.Eval(ctx, `go_upper("hello")`)
s, _ := perl.As[perl.ScalarValue](r.Value)
fmt.Println(s.String()) // HELLO
```

Every value crosses the boundary typed, following Perl's own structure:
a `Value` is one of the sealed concrete types — `ScalarValue` (SV),
`RefValue` (RV), `ArrayValue` (AV), `HashValue` (HV), `CodeValue` (CV),
`GlobValue`, `IOValue` — inspected either with a Go type switch or with
`perl.As[T]` when the expected type is known up front; `Kind()` reports the
runtime kind (mirroring `reflect.Value.Kind`). Nothing is stringified in
transit and byte strings cross raw (`Bytes()` next to `String()`). Scalars
coerce the way Perl would (`Bool`/`Int`/`Float`/`String`), references
dereference with `RefValue.Deref` (the `reflect.Value.Elem` analog), and
aggregates operate in place through `ArrayValue`/`HashValue`/`CodeValue` —
real element accesses in the interpreter, so ties and overloads behave as
in plain Perl. Perl references — blessed objects, array/hash/code refs —
cross as identity-preserving handles, never serialized, so the same object
stays the same object across any number of round trips.
`NewArray`/`NewHash` materialize Go data as guest aggregates in one
crossing, and an `ArrayValue`/`HashValue` in an argument list flattens
exactly like `f(@a, %h)`. Errors map to `*PerlError` / Perl `die`
respectively.

## Serving PSGI applications

The [`psgi`](./psgi) package supplies the server side of PSGI on top of the
bridge: the `.psgi` file's last evaluated value is the application, exactly
as PSGI specifies, and every request builds the environment as a guest hash
and calls it — no serialisation format sits between Go and the application.
`psgi.Server` is a fixed pool of warm interpreter workers (the
starman/starlet model): prepare ONE instance, and every worker is a
copy-on-write clone of it, so the loading work runs once no matter the
worker count. Delayed and streaming responses are supported, and a request
holds a worker only while the application runs — a slow client does not
occupy one.

```go
p, _ := perl.New(perl.Config{})
server, _ := psgi.New(p, 4, "app.psgi")
server.ListenAndServe(":8091") // or http.ListenAndServe(addr, server)
```

The application's CPAN dependencies need no code: declare them in a
`cpanfile` next to `go.mod` and run `gperl mod tidy`. It writes a `go.work`
through which `go build` links them into the binary — see
[examples/plack](./examples/plack).

## The gperl command

`gperl` is the perl-meets-go toolchain: it handles Perl programs with
go-style workflows on the embedded interpreter.

```sh
gperl get Module[@version]...   # add to the cpanfile and resolve
gperl mod tidy                  # resolve the cpanfile, write cpanfile.snapshot
gperl mod download              # fetch and build what the snapshot names
gperl clean -modcache           # remove the module directory
gperl install toolchain         # install the Go and C toolchains gperl builds with

gperl run script.pl [args...]   # sync dependencies, then run
gperl run [switches] [script]   # the perl(1) command line: every perlrun
                                # switch (-e/-n/-p/-i/-M/-C/-T/...), a
                                # program on stdin, #! line switches
gperl test [-v] [-j N] [paths]  # run t/*.t, report like prove
gperl test -c [-o out]          # the project's test binary (<dir>.test), which
                                # takes gperl test's options and paths
gperl build [-o out] script.pl  # one binary: script + dependencies + interpreter

gperl psgi run [options] app.psgi          # serve a PSGI application
gperl psgi build [-o out] [options] app.psgi   # build its server binary

gperl image build -push ghcr.io/me/app:v1 script.pl   # a container image
gperl image build -push ghcr.io/me/app:v1 -psgi app.psgi [-- options]
```

**Dependencies** follow the Perl conventions on the outside — a `cpanfile`
declares them and `cpanfile.snapshot` (the Carton format) pins the resolved
releases — and the Go model on the inside: the distributions themselves
live in a module directory shared by every project (`~/gperl/pkg/mod`, or
`GPERLMODCACHE`), each release extracted and built once. Modules are
resolved through CPAN MetaDB and MetaCPAN, fetched from a CPAN mirror
(`GPERLMIRROR`; `GPERLMIRRORONLY=1` resolves with the mirror's own index),
and verified against the author's `CHECKSUMS`. Builds ride each
distribution's own `Makefile.PL`/`Build.PL` with every `perl` in the
pipeline being the embedded interpreter, so the only tools required are
`make` and a C compiler.

**`gperl run` and the binaries `gperl build` produces** behave like
`perl`: host filesystem, host environment, no sandbox (the library's zero
`Config` is the opposite — deny by default). The program runs as perl's
main program — `(*Perl).RunProgram` with a `perl.Program`, the parsed form
of a perl command line, which an embedding application can use directly.
`gperl test` runs every test file as a perl process of its own;
`--in-process` runs them inside one process instead, each as an isolated
program with its own working directory and environment, which is not the
host process (it may not listen on the network). A built binary needs
nothing beside it: the XS modules are linked in statically, or with
`GPERLNATIVE=load` embedded and loaded at run time. `gperl test -c` builds
the test run into such a binary, as `go test -c` does: it carries the
whole project directory and its dependencies, and runs the test files from
its own copy of the project, each as a perl process of its own whose `$^X`
runs the binary as perl.

**Toolchain.** `gperl run` and `gperl test` need nothing but gperl for
pure-Perl dependencies. Building a binary needs the Go toolchain, and
building an XS distribution a C compiler; a machine with neither gets both
from `gperl install toolchain`, which puts a pinned Go and Zig (as the C
compiler) into `~/gperl/toolchain` (`GPERLTOOLCHAIN`), each download
checked against a recorded hash. A tool named by the environment (`CC`,
`AR`) is used first, then the installed one, then the one on `PATH`.
Distributions that build with `Makefile.PL` also need `make`.

**Cross compilation**: `GPERLOS` and `GPERLARCH` name the target with Go's
names (`linux`, `darwin`, `windows`, ... and `amd64`, `arm64`, ...); unset,
they name this machine. Without XS dependencies nothing else is needed.
With them, the native modules are built for the target by the C compiler
in `CC` — for example `CC="zig cc -target aarch64-linux-musl"`, which
with the modules linked in yields a fully static binary — or by the
installed Zig when `CC` is unset. Native modules built for musl are kept
apart from glibc's, and a build for musl carries the build tag `musl`:
`gperl build` adds it, and a Go application built with `go build` on the
generated `go.work` passes `-tags musl`. `gperl mod tidy`, `gperl mod
download` and `gperl get` build the native modules for `GPERLOS` and
`GPERLARCH`, so such an application is built for the same platform.

**`gperl psgi`** is the application server and the front end in one
process: one interpreter per worker goroutine, TLS with HTTP/2, graceful
shutdown, access log, and worker recycling. Options follow plackup and
starman (`-p`, `-o`, `-l`, `-E`, `-I`, `-M`, `--workers`, `--max-requests`,
`--enable-ssl`, `--ssl-cert`, `--ssl-key`, `--read-timeout`,
`--keepalive-timeout`, `--access-log`, `-r`). A server binary takes each
setting from its command line, else from the environment
(`GPERL_PSGI_<OPTION>`, and `PORT`), else from the options it was built
with. A server binary's workers load the native (XS) modules linked into
it; under `gperl psgi run` they load native modules from `@INC` only when
the project's dependencies have some, so a native module in a directory
given with `-I` or `PERL5LIB` is found only then.

**`gperl image build`** makes a container image without Docker: it builds
the script — or with `-psgi` the application's server, taking `gperl psgi
build`'s options after `--` — for `linux/amd64`, `linux/arm64` or both
(`-platform`, default both), and puts each binary at `/app/<name>` on the
base image as its entrypoint. The base is `-base`, or
`cgr.dev/chainguard/static` for a static binary and
`cgr.dev/chainguard/glibc-dynamic` for one that needs the dynamic loader.
`-push repository:tag` pushes the image (an index of the platforms' images
for several) with the credentials of `docker login` and prints its digest;
`-tarball file` writes an image of one platform as the archive `docker
load` reads. The image's layer and config carry no build time, so the same
binary makes the same image. The binary unpacks its standard library into
the user cache directory or the temporary directory at startup, so an image
run with a read-only root filesystem needs a writable `/tmp`.

## Supply-chain verification

Two files in this repository are release artifacts of
[perl-wasm](https://github.com/goccy/perl-wasm), not hand-written code:
`internal/perl.go` (the generated bridge) and `stdlib.zip` (the embeddable
stdlib).
They are refreshed with:

```sh
make perl PERL_WASM_VERSION=v0.2.1
```

which downloads both from the perl-wasm release and verifies each against the
release's SLSA build-provenance attestation (`gh attestation verify`). CI
re-verifies them on every push and pull request. The interpreter itself comes
in through the `perlwasm2go` Go module dependency, which applies the same
attestation-verified vendoring on its side.

## License

- **The Go source code of this repository is licensed under [MIT](./LICENSE).**
- **`stdlib.zip` is not MIT**: it is a repackaged subset of the Perl 5.44.0
  standard library — a derivative work of
  [Perl 5](https://github.com/Perl/perl5) — and keeps Perl's own dual license:
  the GNU General Public License version 1 or (at your option) any later
  version ([`LICENSE-GPL`](./LICENSE-GPL)), **or** the "Artistic License"
  ([`LICENSE-ARTISTIC`](./LICENSE-ARTISTIC)), at your choice. Both texts are
  vendored verbatim from the pinned Perl 5.44.0 sources.
- The [`perlwasm2go`](https://github.com/goccy/perlwasm2go) dependency (the
  translated interpreter) is likewise dual-licensed under Perl's terms in its
  own repository.

### Using go-perl in your own project

- **As a library dependency** (source distribution): your repository contains
  no Perl-derived bytes — only an import path and a go.mod entry. License your
  code however you like (MIT, proprietary, ...); no Perl license text needs to
  accompany it. Your users receive go-perl and perlwasm2go from their own
  origins, under their own licenses.
- **Shipping a compiled binary**: the binary embeds the translated interpreter
  and `stdlib.zip`. The Artistic License expressly permits this: linking the
  complete interpreter into your executable is "a mere form of aggregation"
  (§5), and embedded use inside a (commercial) distribution "shall not be
  construed as a distribution of this Package" (§8) — so your own code does not
  inherit Perl's license and no Perl license text is required to accompany the
  binary, provided the interpreter is embedded complete and unmodified, you do
  not overtly expose Perl's interfaces to your end users, and you do not
  advertise Perl as your own product (§5, §9). If your product *does* expose
  Perl to end users (say, a Perl plugin system), add a short third-party notice
  pointing at Perl 5 and its dual license — customary and costless in any case.
- **Perl code you run** on the interpreter, and its output, remain yours (§6).

This summary is not legal advice; the license texts govern.
