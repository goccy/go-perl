# Plack / PSGI on go-perl

Serve a real Perl web application — [Mojolicious](https://mojolicious.org/)
running as a [PSGI](https://metacpan.org/pod/PSGI) app — from Go's
`net/http`, with no `perl` installed, as one binary.

```
net/http ──goroutine per request──► warm *perl.Perl workers
                                        │  go-perl/psgi
                                        ▼
                                 app.psgi (Mojolicious)
```

- `main.go` is ordinary go-perl embedding code: nothing in it names a CPAN
  module, a module directory, or XS.
- The application's dependencies are the `cpanfile`: Plack, Mojolicious, and
  the XS distribution Cpanel::JSON::XS. `cpanfile.snapshot` pins the
  releases.
- `gperl mod download` fetches those releases into the module directory
  (`~/gperl/pkg/mod`, shared by every project), builds each once — the XS
  module against go-perl's native XS SDK — and writes `go.work`.
- With that workspace, `go build` links everything in: the pure-Perl modules
  embedded, the XS module statically. The interpreter runs on its default
  sandboxed filesystem; only `app.psgi` is read from disk.
- Every worker is a copy-on-write clone of one interpreter that loaded the
  application, so loading runs once and the workers share memory.

## Run it

```sh
make setup   # gperl mod download: fetch, build, write go.work
make run     # serves on http://localhost:8091
```

```sh
curl http://localhost:8091/
curl http://localhost:8091/json
curl -d 'shout this' http://localhost:8091/echo
curl http://localhost:8091/fast-json   # encoded by the XS module
```

`make test` runs the same routes through httptest. `go.work` holds paths of
this machine and is not committed; `go mod tidy` and the other go commands
work with or without it — without it the binary simply carries no CPAN
modules.

## Without writing Go

The same application as a server binary, built by gperl alone:

```sh
make server                      # gperl psgi build -o plack-server app.psgi
./plack-server --port 8091 --workers 8
```
