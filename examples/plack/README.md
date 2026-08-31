# Plack / PSGI on go-perl

Serve a real Perl web application — [Mojolicious](https://mojolicious.org/)
running as a [PSGI](https://metacpan.org/pod/PSGI) app under
[Plack](https://metacpan.org/pod/Plack) — from Go's `net/http`, with no
external `perl` at runtime.

```
net/http ──goroutine per request──► pool of warm *perl.Perl instances
                                        │  psgi.Handle(ctx, p, w, r)
                                        ▼
                       go-perl/psgi (env/response conversion + adapter)
                                        │
                                        ▼
                                 app.psgi (Mojolicious)
```

- A fixed pool of Perl instances boots at startup, each with Plack and the
  Mojolicious app compiled. Instances are cheap: go-perl maps every
  interpreter copy-on-write from a process-wide snapshot.
- The request/response conversion (http.Request -> PSGI env, PSGI triple ->
  ResponseWriter) and the Plack::Util adapter live in the reusable
  `github.com/goccy/go-perl/psgi` package; this example only owns the pool.
  Cancelling the request context stops the Perl code mid-run.
- The dependencies are ordinary CPAN modules from the `cpanfile`, vendored
  with carton into `./local`; the pool puts `local/lib/perl5` on `@INC`.
- XS distributions in the `cpanfile` (Cpanel::JSON::XS here) are compiled
  against go-perl's native XS SDK with `gperl xs build` and loaded into every
  pool instance — see `/fast-json`.

## Run it

```sh
make setup   # carton install + carton bundle (host perl; carton is
             # bootstrapped into ./.tools if missing), then `gperl xs build`
             # compiles the cpanfile's XS distributions into ./local/xs
make run     # serves on http://localhost:8091
```

```sh
curl http://localhost:8091/
curl http://localhost:8091/json
curl -d 'shout this' http://localhost:8091/echo
```

`make test` runs the same routes through httptest.

## Limitations

- `psgi.streaming` is off: delayed/streaming PSGI responses are buffered
  guest-side and rejected if the app returns a code ref.
- Response and request bodies cross the bridge base64-encoded (the bridge
  speaks JSON, which cannot carry arbitrary bytes), so very large bodies pay
  an encoding tax.
