# The environment every distribution is configured, built and tested in:
# a Debian with a C toolchain and the development files of the C libraries
# CPAN distributions most often bind to, where the only perl is gperl.
FROM debian:bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential ca-certificates curl git patch pkg-config unzip bzip2 xz-utils \
      zlib1g-dev libssl-dev libexpat1-dev libxml2-dev libyaml-dev libsqlite3-dev \
      libdb-dev libgdbm-dev libgmp-dev libmpfr-dev libpng-dev libjpeg-dev \
 && rm -rf /var/lib/apt/lists/*

# The libraries further distributions bind to, as CPAN Testers machines
# commonly have them: without them a distribution fails for want of a
# library, and every distribution that depends on it is blocked, whatever
# the interpreter does.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      cmake graphviz protobuf-compiler \
      libffi-dev libglib2.0-dev libgirepository1.0-dev libcairo2-dev libpango1.0-dev \
      libgdk-pixbuf-2.0-dev librsvg2-dev libpoppler-glib-dev libharfbuzz-dev libfreetype-dev \
      libfontconfig-dev libgtk-3-dev libgtk2.0-dev libvte-2.91-dev libgoocanvas-2.0-dev libnotify-dev \
      libsoup2.4-dev libwxgtk3.2-dev libx11-dev libxft-dev libxt-dev libxtst-dev \
      libpq-dev libmariadb-dev libmariadb-dev-compat libmemcached-dev libhiredis-dev libgdbm-compat-dev \
      libgd-dev libtiff-dev libgif-dev libwebp-dev libopenjp2-7-dev libraw-dev liblcms2-dev libexif-dev \
      libreadline-dev libncurses-dev libtickit-dev libtermkey-dev libunibilium-dev libvterm-dev \
      libxslt1-dev libmagic-dev libidn11-dev libidn2-dev libicu-dev libyajl-dev libjansson-dev \
      libapr1-dev libaprutil1-dev libcurl4-openssl-dev libpcap-dev libzmq3-dev libevent-dev \
      libkrb5-dev libsasl2-dev libldap2-dev libssh2-1-dev libpam0g-dev libgcrypt20-dev libgnutls28-dev \
      libnss3-dev libsodium-dev libmcrypt-dev libgpgme-dev libkeyutils-dev \
      libpopt-dev libpcre3-dev libpcre2-dev libarchive-dev libzstd-dev libbz2-dev liblzma-dev \
      liblz4-dev libsnappy-dev libmpc-dev libgsl-dev libfftw3-dev libjudy-dev libcdb-dev libtdb-dev \
      libqdbm-dev libxapian-dev libaspell-dev libhunspell-dev libtidy-dev libmaxminddb-dev libgeoip-dev \
      libprotobuf-dev uuid-dev libltdl-dev libtirpc-dev libnsl-dev libdumbnet-dev libsnmp-dev \
      libogg-dev libvorbis-dev libflac-dev libsndfile1-dev libao-dev libmad0-dev libasound2-dev \
      libpulse-dev libopenal-dev libsdl2-dev libusb-dev libudev-dev libsystemd-dev libdbus-1-dev \
      libcap-dev libacl1-dev libattr1-dev libselinux1-dev libblkid-dev libmount-dev libfuse-dev \
      libnl-3-dev libnetfilter-queue-dev libpcsclite-dev libsane-dev \
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
