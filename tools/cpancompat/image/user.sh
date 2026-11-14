#!/bin/sh
# Runs once while either image is built, as root: makes the unprivileged
# user distributions are configured, built and tested as. CPAN clients and
# smokers run as a normal user: as root, tests that expect a file they made
# unreadable to be unreadable fail, tests that refuse to run as root fail,
# and tests that only root runs are not skipped.
set -eu
user=cpan
useradd --create-home --uid 1000 --user-group --shell /bin/bash --comment "CPAN tester" $user

# What the harness writes as that user: the run's install tree and the
# build directories (each a volume, which takes the owner of its directory
# in the image when it is first mounted), the staging directory, and the
# records of the run, which the driver copies out of the container.
mkdir -p /opt/gperl/local/lib/perl5 /opt/gperl/local/bin /work /stage /out
chown -R $user:$user /opt/gperl /work /stage /out
# Where the driver acknowledges that it took the container off its network
# (dist.sh): root's, so that no process of the user can do it instead.
mkdir -p /run/cpancompat

# The containers run without capabilities (the code under test may be
# malicious), and a program whose file grants capabilities then cannot be
# run at all: exec fails with EPERM. ping is one, and needs none: without
# CAP_NET_RAW it sends its echo requests on an ICMP datagram socket, which
# Docker lets every group open (net.ipv4.ping_group_range). So no program
# keeps file capabilities.
getcap -r /usr 2> /dev/null | while read -r file caps; do
    echo "user.sh: removing $caps from $file"
    setcap -r "$file"
done

# A git identity, as a login user has one: git cannot make one up from the
# container's host name, which has no domain.
su $user -c 'git config --global user.name "CPAN tester" && git config --global user.email cpan@localhost.localdomain'
