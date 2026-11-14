# The environment every distribution is configured, built and tested in:
# a Debian with a C toolchain, the tools build scripts run and the
# development files of the C libraries CPAN distributions bind to (the list
# in packages, and MySQL's client library from mysql.sh), where the perl
# command is gperl, as an unprivileged user in a login environment.
# Dockerfile.reference builds the same environment around perl itself.
FROM debian:bookworm-slim

COPY packages mysql.sh user.sh /opt/cpancompat/
RUN apt-get update \
 && apt-get install -y --no-install-recommends $(sed 's/#.*//' /opt/cpancompat/packages) \
 && rm -rf /var/lib/apt/lists/*
RUN sh /opt/cpancompat/mysql.sh

COPY gperl /opt/gperl/bin/gperl
RUN sh /opt/cpancompat/user.sh
COPY setup.sh dist.sh watch.sh builder.pl unmet.pl /opt/cpancompat/
COPY Harness.pm /opt/cpancompat/lib/CPANCompat/Harness.pm

# What a login session of that user has. The C.UTF-8 locale is the one a
# Debian system without a configured locale offers.
ENV USER=cpan LOGNAME=cpan HOME=/home/cpan SHELL=/bin/bash LANG=C.UTF-8
USER cpan
# As the user: gperl unpacks its standard library into the user's cache.
RUN sh /opt/cpancompat/setup.sh

# gperl is the perl command: the shim is what $^X, make rules and #! lines
# name, and the preload describes the C toolchain of this machine to the
# build tools (the interpreter's own %Config describes wasm32).
ENV PATH=/opt/gperl/local/bin:/opt/gperl/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    GPERL_PERL_EXE=/opt/gperl/bin/perl \
    GPERL_PERL_PRELOAD=/opt/gperl/config-overlay.pl \
    PERL5LIB=/opt/gperl/local/lib/perl5
