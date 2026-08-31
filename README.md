# go-perl

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

## Features

- **Pure Go**: works anywhere Go compiles; `CGO_ENABLED=0` friendly.
- **Batteries included**: libperl plus every static XS extension (List::Util,
  POSIX, Socket, re, Storable, Encode, ...) and the pure-Perl stdlib
  (embedded zip, served from a private in-memory filesystem by default).
- **Instant start, copy-on-write**: the first `New` boots one interpreter and
  snapshots its memory; every later `New` maps that snapshot copy-on-write, so
  instances start without re-running interpreter init and share the read-only
  bulk of their memory.
- **Sandboxed by default**: each `Perl` runs in its own WASI sandbox with a
  PRIVATE in-memory filesystem — an embedded instance touches no host files
  unless asked to. `Config.FS` selects the backend from the
  [`fs`](./fs) package: `fs.NewHostFS()` passes through to the operating
  system's filesystem (how the `gperl` command behaves, matching `perl`),
  `fs.DirFS` scopes it to one directory, `fs.NewMemFS()` or any custom
  backend plugs in the same way — and environment, network (`Dial`/`Resolve`)
  and subprocess (`Exec`) policy hooks complete the sandbox.
- **Multi-instance**: independent `Perl` instances share nothing (writable
  state, that is — read-only snapshot pages are shared).
- **Cancellable**: cancelling the `context.Context` passed to `Eval` stops a
  runaway script at the next Perl opcode.
- **Bridged, typed**: every value crosses the boundary typed, following
  Perl's own structure — `ScalarValue` (SV), `RefValue` (RV), `ArrayValue`
  (AV), `HashValue` (HV), `CodeValue` (CV) — nothing is stringified in
  transit and byte strings cross raw (`Bytes()` next to `String()`). `Call`
  invokes named Perl subs from Go and `Bind` makes Go functions callable
  from Perl as ordinary subs. Perl references — blessed objects,
  array/hash/code refs — cross as identity-preserving handles, never
  serialized, so the same object stays the same object across any number of
  round trips. Errors map to `*PerlError` / Perl `die` respectively.

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

The value model follows Perl's own type structure. A `Value` is one of the
sealed concrete types — `ScalarValue`, `RefValue`, `ArrayValue`,
`HashValue`, `CodeValue`, `GlobValue`, `IOValue` — inspected either with a
Go type switch or with `perl.As[T]` when the expected type is known up
front; `Kind()` reports the runtime kind (mirroring `reflect.Value.Kind`).
Scalars coerce the way Perl would (`Bool`/`Int`/`Float`/`String`, plus
`Bytes` for the raw byte string), references dereference with
`RefValue.Deref` (the `reflect.Value.Elem` analog), and aggregates operate
in place through `ArrayValue`/`HashValue`/`CodeValue` — real element
accesses in the interpreter, so ties and overloads behave as in plain
Perl. `NewArray`/`NewHash` materialize Go data as guest aggregates in one
crossing, and an `ArrayValue`/`HashValue` in an argument list flattens
exactly like `f(@a, %h)`. A bound Go function may call back into the same
instance, so round trips compose. The [`psgi`](./psgi) package builds on
the bridge to serve PSGI web applications from `net/http` — the `.psgi`
file's last evaluated value
is the application, exactly as PSGI specifies; see
[`examples/plack`](./examples/plack) for it carrying real traffic — a
Mojolicious app over a pool of warm instances.

## Supply-chain verification

Two files in this repository are release artifacts of
[perl-wasm](https://github.com/goccy/perl-wasm), not hand-written code:
`perl.go` (the generated bridge) and `stdlib.zip` (the embeddable stdlib).
They are refreshed with:

```sh
make perl PERL_WASM_VERSION=v0.1.1
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
- **`gperl/cpanm` is not MIT**: it is the fatpacked
  [cpanminus](https://github.com/miyagawa/cpanminus) program, vendored
  verbatim from the App-cpanminus 1.7049 CPAN release and embedded so the
  XS build pipeline can resolve CPAN dependencies without downloading or
  consulting a host installation. It stays under its own terms — the same
  dual license as Perl itself (see the script's POD).
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
