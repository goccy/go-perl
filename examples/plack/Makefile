# Prepares the example with a host perl toolchain, then serves from Go —
# no pre-existing Perl environment needed beyond `perl` itself:
#
#   make setup   carton install vendors the cpanfile modules into ./local
#                (carton is bootstrapped into ./.tools if not installed),
#                carton bundle keeps the dist tarballs (./local/cache — NOT
#                Go's ./vendor), and `gperl xs build` compiles the XS
#                distributions from those tarballs against go-perl's native
#                XS SDK into ./local/xs (loaded by the pool at boot).
#   make run     go run .   (serves on http://localhost:8091)
#   make test    go test ./...
PERL   ?= perl
TOOLS  := $(CURDIR)/.tools
CARTON := $(shell command -v carton 2>/dev/null)
ifeq ($(CARTON),)
CARTON := PERL5LIB=$(TOOLS)/lib/perl5 $(TOOLS)/bin/carton
endif
GPERL  := go run ../../cmd/gperl

# The XS distributions from the cpanfile; everything else rides pure Perl.
XS_DISTS := Cpanel-JSON-XS

.PHONY: setup deps xs run test clean

setup: deps xs

# carton runs from a scratch directory: on a case-insensitive filesystem
# (macOS) resolving the module name "Plack" against this very directory
# ("plack") would try to install the directory instead of the dist. The
# vendored tree, the dist tarballs, and the snapshot are synced back here.
deps:
	@test -x $(TOOLS)/bin/carton || command -v carton >/dev/null 2>&1 || \
		{ echo "==> bootstrapping carton into $(TOOLS)"; \
		  curl -fsSL https://cpanmin.us | $(PERL) - -L $(TOOLS) --notest Carton; }
	@WORK=$$(mktemp -d) && \
	cp cpanfile $$WORK/ && \
	{ test ! -f cpanfile.snapshot || cp cpanfile.snapshot $$WORK/; } && \
	( cd $$WORK && $(CARTON) install && $(CARTON) bundle ) && \
	rm -rf local/lib local/cache && \
	mkdir -p local && \
	mv $$WORK/local/lib local/lib && \
	mv $$WORK/vendor/cache local/cache && \
	cp $$WORK/cpanfile.snapshot cpanfile.snapshot && \
	rm -rf $$WORK

xs:
	$(GPERL) xs build $(foreach d,$(XS_DISTS),local/cache/authors/id/*/*/*/$(d)-*.tar.gz)

run:
	go run .

test:
	go test ./...

clean:
	rm -rf local .tools
