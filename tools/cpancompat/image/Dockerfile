# The environment every distribution is configured, built and tested in:
# a Debian with a C toolchain and the development files of the C libraries
# CPAN distributions bind to (the list in packages), where the perl command
# is gperl. Dockerfile.reference builds the same environment around perl
# itself.
FROM debian:bookworm-slim

COPY packages /opt/cpancompat/packages
RUN apt-get update \
 && apt-get install -y --no-install-recommends $(sed 's/#.*//' /opt/cpancompat/packages) \
 && rm -rf /var/lib/apt/lists/*

COPY gperl /opt/gperl/bin/gperl
COPY setup.sh dist.sh watch.sh builder.pl unmet.pl /opt/cpancompat/
COPY Harness.pm /opt/cpancompat/lib/CPANCompat/Harness.pm
RUN sh /opt/cpancompat/setup.sh

# gperl is the perl command: the shim is what $^X, make rules and #! lines
# name, and the preload describes the C toolchain of this machine to the
# build tools (the interpreter's own %Config describes wasm32).
ENV PATH=/opt/gperl/local/bin:/opt/gperl/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    GPERL_PERL_EXE=/opt/gperl/bin/perl \
    GPERL_PERL_PRELOAD=/opt/gperl/config-overlay.pl \
    PERL5LIB=/opt/gperl/local/lib/perl5
