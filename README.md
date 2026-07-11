# go-perl

Embed the Perl 5 interpreter in Go — as **pure Go**. No cgo, no wasm runtime,
no external `perl` binary: the interpreter is
[Perl 5.42.2 compiled to WebAssembly](https://github.com/goccy/perl-wasm) and
then [translated to Go](https://github.com/goccy/wasm2go), published as the
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

Practical note: under the Artistic License, embedding the (complete,
unmodified) interpreter and its library in your application is expressly
permitted as mere aggregation — your own application code does not inherit
Perl's license, and Perl code you run on the interpreter, plus its output,
remain yours.
