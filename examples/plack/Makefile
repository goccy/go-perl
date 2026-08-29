# Vendors the CPAN modules the example needs (Plack + Mojolicious, both pure
# Perl) into ./local with a host perl, then serves the PSGI app from Go.
# cpanm is fetched as the standalone script if it is not installed.
#
# cpanm runs from a scratch directory: on a case-insensitive filesystem
# (macOS) running it here would resolve the module name "Plack" against this
# very directory ("plack") and try to install the directory instead.
CPANM   := $(shell command -v cpanm 2>/dev/null)
WORKDIR := $(shell mktemp -d)

.PHONY: setup run test clean

setup:
ifeq ($(CPANM),)
	cd $(WORKDIR) && curl -fsSL https://cpanmin.us | perl - -L $(CURDIR)/local --notest Plack Mojolicious
else
	cd $(WORKDIR) && $(CPANM) -L $(CURDIR)/local --notest Plack Mojolicious
endif

run:
	go run .

test:
	go test ./...

clean:
	rm -rf local
