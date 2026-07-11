# go-perl

**Perl 5 in pure Go — embed and run Perl anywhere Go runs. No cgo, no external
`perl`, one static binary.**

The interpreter is
[Perl 5.42.2 compiled to WebAssembly](https://github.com/goccy/perl-wasm) and
then [translated to Go](https://github.com/goccy/wasm2go) — no wasm runtime is
involved at run time — published as the
[`perlwasm2go`](https://github.com/goccy/perlwasm2go) module this package
builds on. The pure-Perl standard library ships inside the package as an
embedded zip, so a `go build` produces a single self-contained binary that can
run Perl.

```go
package main

import (
	"fmt"

	perl "github.com/goccy/go-perl"
)

func main() {
	i, err := perl.NewInterpreter(perl.Config{})
	if err != nil {
		panic(err)
	}
	defer i.Close()

	r, err := i.Eval(`join(",", map { $_ * 2 } 1..5)`)
	if err != nil {
		panic(err)
	}
	fmt.Println(r.Result) // 2,4,6,8,10
}
```

## Features

- **Pure Go**: works anywhere Go compiles; `CGO_ENABLED=0` friendly.
- **Batteries included**: libperl plus every static XS extension (List::Util,
  POSIX, Socket, re, Storable, Encode, ...) and the pure-Perl stdlib
  (embedded zip, unpacked automatically — or served from an in-memory FS via
  `NewStdlibMemFS`).
- **Sandboxed**: each `Interpreter` runs in its own WASI sandbox with a
  pluggable filesystem backend (`Config.FS`), environment, and no ambient
  host access.
- **Multi-interpreter**: independent `Interpreter` instances share nothing.
- **Interruptible**: a running `Eval` can be stopped from another goroutine.

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
- **`stdlib.zip` is not MIT**: it is a repackaged subset of the Perl 5.42.2
  standard library — a derivative work of
  [Perl 5](https://github.com/Perl/perl5) — and keeps Perl's own dual license:
  the GNU General Public License version 1 or (at your option) any later
  version ([`LICENSE-GPL`](./LICENSE-GPL)), **or** the "Artistic License"
  ([`LICENSE-ARTISTIC`](./LICENSE-ARTISTIC)), at your choice. Both texts are
  vendored verbatim from the pinned Perl 5.42.2 sources.
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
