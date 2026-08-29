# Plack / PSGI on go-perl

Serve a real Perl web application — [Mojolicious](https://mojolicious.org/)
running as a [PSGI](https://metacpan.org/pod/PSGI) app under
[Plack](https://metacpan.org/pod/Plack) — from Go's `net/http`, with no
external `perl` at runtime.

```
net/http ──goroutine per request──► pool of warm *perl.Perl instances
                                        │  perl.Call(ctx, "psgi_handle", env, body)
                                        ▼
                                    psgi.pl (Plack::Util) ──► app.psgi (Mojolicious)
```

- A fixed pool of Perl instances boots at startup, each with Plack and the
  Mojolicious app compiled. Instances are cheap: go-perl maps every
  interpreter copy-on-write from a process-wide snapshot.
- Each HTTP request checks an instance out of the pool, crosses into Perl
  over the function bridge (`perl.Call`), and gets `(status, headers, body)`
  back. Cancelling the request context stops the Perl code mid-run.
- The dependencies are ordinary CPAN modules vendored with cpanm into
  `./local`; the pool puts `local/lib/perl5` on `@INC`.

## Run it

```sh
make setup   # cpanm -L local --notest Plack Mojolicious (host perl)
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
