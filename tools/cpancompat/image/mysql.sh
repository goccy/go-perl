#!/bin/sh
# Runs once while either image is built, as root: installs Oracle's MySQL 8
# client library and headers (mysql_config, libmysqlclient), which
# DBD-mysql 5 and what depends on it need. Debian bookworm has only
# MariaDB's client library, and Oracle's APT repository publishes no arm64
# packages, so this takes Oracle's Enterprise Linux 9 RPMs for the
# machine's architecture (built for glibc 2.34 and OpenSSL 3, which
# bookworm's glibc 2.36 and libssl3 run), checks their signatures against
# Oracle's key, and unpacks them into /.
set -eu
version=8.4.9
# The key MySQL release packages are signed with since 2023
# (mysql-build@oss.oracle.com), by its id: the last 8 digits of its
# fingerprint BCA4 3417 C3B4 85DD 128E C6D4 B7B3 B788 A8D3 785C.
key_id=a8d3785c
arch=$(uname -m)

apt-get update
apt-get install -y --no-install-recommends rpm2cpio cpio rpm
rm -rf /var/lib/apt/lists/*

tmp=$(mktemp -d)
cd "$tmp"
curl -fsSLO https://repo.mysql.com/RPM-GPG-KEY-mysql-2023
for p in common libs client-plugins devel; do
    curl -fsSLO "https://repo.mysql.com/yum/mysql-8.4-community/el/9/$arch/mysql-community-$p-$version-1.el9.$arch.rpm"
done
rpmkeys --import RPM-GPG-KEY-mysql-2023
test "$(rpm -q gpg-pubkey --qf '%{VERSION}\n')" = "$key_id"
# Every package must carry a signature, and it must verify: rpmkeys reports
# a package that has only digests as OK too, so its report is read.
for f in *.rpm; do
    report=$(rpmkeys --checksig "$f")
    echo "$report"
    case $report in
    *" signatures OK") ;;
    *) echo "mysql.sh: $f is not signed by the MySQL key" >&2; exit 1 ;;
    esac
done
for f in *.rpm; do
    rpm2cpio "$f" | (cd / && cpio -idm --quiet)
done
cd /
rm -rf "$tmp"

# What the packages' install scripts would have done: the library directory
# for the dynamic linker, and mysql_config under its usual name.
echo /usr/lib64/mysql > /etc/ld.so.conf.d/mysql.conf
ldconfig
test -e /usr/bin/mysql_config || ln -s mysql_config-64 /usr/bin/mysql_config
# pkg-config looks in Debian's directories, not in /usr/lib64/pkgconfig.
mkdir -p /usr/lib/pkgconfig
ln -s /usr/lib64/pkgconfig/mysqlclient.pc /usr/lib/pkgconfig/mysqlclient.pc
pkg-config --modversion mysqlclient
mysql_config --version --cflags --libs
