PERL_WASM_REPO     ?= goccy/perl-wasm
PERL_WASM_VERSION  ?= v0.3.0
# perl-wasm emits its release attestations from release.yml (the v* tag
# workflow), NOT build.yml — releasing lives only in release.yml there.
PERL_WASM_WORKFLOW ?= goccy/perl-wasm/.github/workflows/release.yml

# Files pulled from the perl-wasm release and verified against its SLSA
# attestation: the wasm2go bridge and the embeddable stdlib zip. Asset name on
# the left, in-tree filename on the right (the rename is cosmetic; gh
# attestation verify matches by content digest).
BRIDGE_ASSET := perl_wasm2go.go
BRIDGE_FILE  := internal/perl.go
STDLIB_ASSET := perl_stdlib.zip
STDLIB_FILE  := stdlib.zip
RELEASE_URL       = https://github.com/$(PERL_WASM_REPO)/releases/download/$(PERL_WASM_VERSION)
ATTESTATION_API   = https://api.github.com/repos/$(PERL_WASM_REPO)/attestations

.PHONY: perl download verify test test-cross bench bench-perl

## perl: refresh the bridge + stdlib from the upstream release and verify
## their GitHub artifact attestations. Run whenever PERL_WASM_VERSION bumps.
perl: download verify

## download: fetch the wasm2go bridge and the stdlib zip from the upstream
## release and drop them in place at $(BRIDGE_FILE) / $(STDLIB_FILE).
download:
	curl -fSL --proto '=https' --tlsv1.2 -o $(BRIDGE_FILE) $(RELEASE_URL)/$(BRIDGE_ASSET)
	curl -fSL --proto '=https' --tlsv1.2 -o $(STDLIB_FILE) $(RELEASE_URL)/$(STDLIB_ASSET)

## verify: confirm each upstream-sourced file carries a valid GitHub artifact
## attestation signed by the upstream release.yml workflow. Each bundle is
## fetched anonymously from the public attestation API and verified offline via
## `gh attestation verify --bundle`. No GH access token is required.
verify:
	@set -eu; \
	root=$$(mktemp -d); \
	trap 'rm -rf $$root' EXIT; \
	for f in $(BRIDGE_FILE) $(STDLIB_FILE); do \
	  bundle=$$root/$$(echo $$f | tr / _).jsonl; \
	  digest=$$(shasum -a 256 $$f | awk '{print $$1}'); \
	  echo "==> fetching attestation bundle for $$f (sha256:$$digest)"; \
	  curl -fsSL --proto '=https' --tlsv1.2 \
	    "$(ATTESTATION_API)/sha256:$$digest" \
	    | jq -c '.attestations[].bundle' > $$bundle; \
	  echo "==> verifying $$f"; \
	  GH_TOKEN= GITHUB_TOKEN= gh attestation verify "$$f" \
	    -R $(PERL_WASM_REPO) \
	    --bundle $$bundle \
	    --signer-workflow $(PERL_WASM_WORKFLOW); \
	done

## test: run the Go test suite.
test:
	go test ./...

## test-cross: build a project with native (XS) dependencies for linux/amd64
## and linux/arm64 with gperl build — cgo with the modules linked in (glibc,
## and fully static on musl), without cgo, and a pure-Perl project with no
## C compiler — checking each binary and running the one for this machine.
## The pinned Go and Zig come from `gperl install toolchain`, and the
## dependency tests run again on them (Zig building for this machine).
test-cross:
	go run ./cmd/gperl install toolchain
	GPERL_TEST_CROSS=1 go test -count=1 -timeout 30m -run '^(TestCrossBuild|TestDependenciesEndToEnd)$$' ./internal/gperl

## bench: benchmark gperl against perl — the PSGI server against Starman
## and Starlet, start-up, and a test suite against prove — with the perl
## side installed into .bench/perl5 by bench-perl.
BENCH_PERL5   := $(CURDIR)/.bench/perl5
BENCH_MODULES := Plack Starman Starlet Server::Starter Net::Server::SS::PreFork Mojolicious Cpanel::JSON::XS
BENCHTIME     ?= 3s
bench: bench-perl
	GPERL_BENCH_PERL5=$(BENCH_PERL5) go test -count=1 -timeout 60m -run '^$$' -bench '^Benchmark(PSGI|Startup|TestSuite)$$' -benchtime $(BENCHTIME) ./internal/gperl

## bench-perl: install the perl side of the benchmarks with cpanm, taken
## from its CPAN release and checked against the hash CPAN records for it.
CPANM_VERSION := 1.7048
CPANM_SHA256  := 59b60907ab9fa4f72ca2004fbe6054911439ae9a906890b4d842a87b25f20f3c
CPANM_DIR     := $(CURDIR)/.bench/App-cpanminus-$(CPANM_VERSION)
bench-perl:
	@set -eu; \
	if [ ! -x $(BENCH_PERL5)/bin/plackup ]; then \
	  if [ ! -f $(CPANM_DIR)/bin/cpanm ]; then \
	    mkdir -p $(CURDIR)/.bench; \
	    tarball=$(CURDIR)/.bench/App-cpanminus-$(CPANM_VERSION).tar.gz; \
	    curl -fsSL --proto '=https' --tlsv1.2 -o $$tarball \
	      https://cpan.metacpan.org/authors/id/M/MI/MIYAGAWA/App-cpanminus-$(CPANM_VERSION).tar.gz; \
	    echo "$(CPANM_SHA256)  $$tarball" | shasum -a 256 -c -; \
	    tar -xzf $$tarball -C $(CURDIR)/.bench; \
	  fi; \
	  perl $(CPANM_DIR)/bin/cpanm -L $(BENCH_PERL5) --notest $(BENCH_MODULES); \
	fi
