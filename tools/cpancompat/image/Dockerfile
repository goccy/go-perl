# The environment every distribution is configured, built and tested in:
# a Debian with a C toolchain and the development files of the C libraries
# CPAN distributions most often bind to, where the only perl is gperl.
FROM debian:bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential ca-certificates curl patch pkg-config unzip bzip2 xz-utils \
      zlib1g-dev libssl-dev libexpat1-dev libxml2-dev libyaml-dev libsqlite3-dev \
      libdb-dev libgdbm-dev libgmp-dev libmpfr-dev libpng-dev libjpeg-dev \
 && rm -rf /var/lib/apt/lists/*

COPY gperl /opt/gperl/bin/gperl
COPY setup.sh dist.sh builder.pl unmet.pl /opt/cpancompat/
COPY Harness.pm /opt/cpancompat/lib/CPANCompat/Harness.pm
RUN sh /opt/cpancompat/setup.sh

# gperl is the perl command: the shim is what $^X, make rules and #! lines
# name, and the preload describes the C toolchain of this machine to the
# build tools (the interpreter's own %Config describes wasm32).
ENV PATH=/opt/gperl/local/bin:/opt/gperl/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    PERL=/opt/gperl/bin/perl \
    GPERL_PERL_EXE=/opt/gperl/bin/perl \
    GPERL_PERL_PRELOAD=/opt/gperl/config-overlay.pl \
    PERL5LIB=/opt/gperl/local/lib/perl5
