#!/bin/bash
# Configures, builds, tests and installs one distribution, the way a CPAN
# client does, and records what happened under /out:
#
#   phases.jsonl   one object per phase that ran, in order: its name, exit
#                  status, duration, and whether it ran out of time
#   tests.jsonl    one object per test file (written by CPANCompat::Harness)
#   unmet.json     the requirements the installed modules do not meet
#   <phase>.log    the output of each phase
#
# The phases up to the build reach the internet (Alien distributions
# download what they build); the tests run with loopback only. Before the
# test phase dist.sh asks the driver to take the container off its network
# by writing {"request":"offline"} to its standard output, the driver's
# channel (everything else it prints goes to its standard error); the
# driver acknowledges by making /run/cpancompat/offline, which only root
# can, and dist.sh checks that no interface but loopback is left. The
# "offline" record before the test phase's says so.
#
# usage: dist.sh <archive> <build|shipped>
#
# "build" is the normal course. "shipped" is for a distribution the
# interpreter itself ships at the same or a newer version: nothing is built
# or installed, and its test files run against the interpreter's own copy.
#
# It runs as the image's unprivileged user, in that user's login
# environment. The distribution is unpacked into a directory of its own
# under /work, the run's build volume, which outlives the container, as
# CPAN clients keep their build directories (~/.cpan/build, ~/.cpanm/work):
# what a distribution installs can point into its build directory (an
# RPATH, a pkg-config file that names it), and the distributions that use
# it later find it there.
#
# The orchestration is a shell script on purpose: it must not depend on the
# perl under measurement behaving.
set -u
archive=$1
mode=$2
out=/out
stage=/stage
dest=/opt/gperl/local
tools=/opt/cpancompat
# The most a phase may write to its log, in bytes.
output_limit=$((256 << 20))

# Standard output is the channel to the driver (fd 3), and no command run
# here gets it.
exec 3>&1 1>&2

# What CPAN clients set for the programs they drive.
export PERL_MM_USE_DEFAULT=1 NONINTERACTIVE_TESTING=1 PERL_USE_UNSAFE_INC=1
export PERL5_CPAN_IS_RUNNING=$$ PERL5_CPANPLUS_IS_RUNNING=$$

# phase <name> <limit in seconds> <command...> runs one phase with its output
# in <name>.log and appends its record to phases.jsonl.
phase() {
    local name=$1 limit=$2
    shift 2
    local start=$SECONDS status seconds timed_out=false output_limited=false watcher
    # The watcher ends a phase whose log passes the output limit (and says
    # so in <name>.output-limit), and shortly before the time limit writes
    # what the processes still running are doing to <name>.stuck.jsonl:
    # whether a phase that runs out of time is blocked, spinning or merely
    # slow is told by that record, not by waiting longer.
    bash $tools/watch.sh $$ "$limit" "$out/$name.log" $output_limit "$out/$name" 3>&- &
    watcher=$!
    # By its full path: distributions install programs into a directory that
    # comes first in PATH, and one of them is named timeout.
    /usr/bin/timeout --kill-after=10 "$limit" "$@" > "$out/$name.log" 2>&1 < /dev/null 3>&-
    status=$?
    # A record of the processes the watcher has begun is kept whole: it is
    # what tells why a phase ran out of time. It takes some seconds.
    while [ -e "$out/$name.stuck.jsonl.partial" ] && [ $((SECONDS - start)) -lt $((limit + 60)) ]; do
        sleep 1
    done
    kill "$watcher" 2> /dev/null
    wait "$watcher" 2> /dev/null
    seconds=$((SECONDS - start))
    # Whether the time limit ended the command is told by its exit status:
    # timeout(1) exits 124 when the command ended after its TERM, and is
    # killed itself (137) when it had to send KILL, 10 seconds past the
    # limit. A command killed by something else (the kernel, for the
    # container's memory limit) ends with 137 too, but before the limit.
    case $status in
        124) timed_out=true ;;
        137) if [ "$seconds" -ge "$limit" ]; then timed_out=true; fi ;;
    esac
    if [ -f "$out/$name.output-limit" ]; then output_limited=true; fi
    printf '{"phase":"%s","status":%d,"seconds":%d,"timed_out":%s,"output_limited":%s}\n' \
        "$name" "$status" "$seconds" "$timed_out" "$output_limited" >> "$out/phases.jsonl"
    return $status
}

# offline takes the container off its network for the test phase and records
# it, status 0 when no interface but loopback is left, 1 when one is, 2 when
# the driver did not acknowledge the request in time.
offline() {
    local start=$SECONDS status=0 i
    printf '{"request":"offline"}\n' >&3
    while [ ! -e /run/cpancompat/offline ] && [ $((SECONDS - start)) -lt 120 ]; do
        sleep 1
    done
    if [ ! -e /run/cpancompat/offline ]; then
        status=2
    else
        for i in /sys/class/net/*; do
            if [ "${i##*/}" != lo ]; then status=1; fi
        done
    fi
    printf '{"phase":"offline","status":%d,"seconds":%d,"timed_out":false}\n' \
        "$status" $((SECONDS - start)) >> "$out/phases.jsonl"
    return $status
}

# The build directory, named after the archive as CPAN.pm names its own.
base=${archive##*/}
for ext in .tar.gz .tgz .tar.bz2 .tbz .tar.xz .tar .zip; do
    base=${base%"$ext"}
done
build=$(mktemp -d "/work/$base-XXXXXX") || exit 1
mkdir -p "$stage$dest"
case "$archive" in
    *.zip) phase unpack 300 unzip -q "$archive" -d "$build" ;;
    *) phase unpack 300 tar -xf "$archive" -C "$build" ;;
esac || exit 0
cd "$build" || exit 0
entries=(*)
if [ ${#entries[@]} -eq 1 ] && [ -d "${entries[0]}" ]; then
    cd "${entries[0]}" || exit 0
fi
chmod -R u+w .

# Configured for where it is installed in the end, the run's install tree:
# what a distribution records about its installation (an Alien library's
# prefix, the path of a program it installs) is then true for the
# distributions that use it. It is installed into the staging directory
# with DESTDIR (--destdir), and published from there.
builder=$(perl $tools/builder.pl 2> "$out/builder.log" 3>&-)
case "$builder" in
    Makefile.PL)
        phase configure 900 perl Makefile.PL INSTALL_BASE=$dest INSTALLMAN1DIR=none INSTALLMAN3DIR=none NO_PERLLOCAL=1
        configured=$?
        made=Makefile
        ;;
    Build.PL)
        phase configure 900 perl Build.PL --install_base $dest
        configured=$?
        made=Build
        ;;
    *)
        printf '{"phase":"configure","status":-1,"seconds":0,"timed_out":false,"no_build_script":true}\n' >> "$out/phases.jsonl"
        exit 0
        ;;
esac
# Configure is judged by whether it wrote the build script, as CPAN clients
# judge it, not by its exit status: a Makefile.PL can write its Makefile and
# still exit non-zero. The phases that follow tell the driver it went on.
if [ ! -f "$made" ]; then
    # Ended successfully without writing it: it declined this platform.
    if [ "$configured" -eq 0 ]; then
        printf '{"phase":"declined","status":0,"seconds":0,"timed_out":false}\n' >> "$out/phases.jsonl"
    fi
    exit 0
fi

# unmet.pl writes the requirements the installed modules do not meet to
# unmet.json. The distribution is built and tested whether or not they are
# all met, as CPAN.pm does (it warns and goes on); a status other than 0 is
# the helper having failed.
phase prerequisites 300 bash -c "perl $tools/unmet.pl > $out/unmet.json" || exit 0

# The test phase has a display, as a smoker's desktop has: xvfb-run starts a
# virtual X server on a free display number for the command. (It waits for
# a signal from the server, which a shell running as PID 1 never gets: the
# container runs with an init process.)
xvfb=(/usr/bin/xvfb-run -a)
export HARNESS_SUBCLASS=CPANCompat::Harness CPANCOMPAT_TESTS=$out/tests.jsonl
if [ "$mode" = shipped ]; then
    if [ "$builder" = Makefile.PL ]; then
        # The test files the distribution declared, without the build that
        # `make test` would start with.
        files=$(make -s --eval='cpancompat-test-files:;@echo $(TEST_FILES)' cpancompat-test-files 2> "$out/test-files.log" 3>&-)
    else
        files='t/*.t'
    fi
    offline || exit 0
    # shellcheck disable=SC2086  # the list is a set of glob patterns
    PERL5LIB=$PERL5LIB:$tools/lib phase test 1200 "${xvfb[@]}" \
        perl -MExtUtils::Command::MM -MTest::Harness -e 'undef *Test::Harness::Switches; test_harness(0)' $files
    exit 0
fi

# An install step may load what it has just installed (XML::SAX registers its
# parser that way). A client that installs into a directory of its own has
# that directory in PERL5LIB, so the staged copy of the install tree is there
# for the step.
stagelib=$stage$dest/lib/perl5
if [ "$builder" = Makefile.PL ]; then
    phase build 1800 make || exit 0
    offline || exit 0
    PERL5LIB=$PERL5LIB:$tools/lib phase test 1200 "${xvfb[@]}" make test
    PERL5LIB=$stagelib:$PERL5LIB phase install 600 make install DESTDIR=$stage || exit 0
else
    phase build 1800 ./Build || exit 0
    offline || exit 0
    PERL5LIB=$PERL5LIB:$tools/lib phase test 1200 "${xvfb[@]}" ./Build test
    PERL5LIB=$stagelib:$PERL5LIB phase install 600 ./Build install --destdir $stage || exit 0
fi

# Publish what was installed for the distributions that depend on this one.
# Other distributions are being tested against the destination right now, so
# every file appears there complete, by a rename.
phase publish 600 flock "$dest/.publish-lock" bash -c '
    set -e
    cd "$1"
    find . -type d -exec mkdir -p "$2/{}" \;
    find . \( -type f -o -type l \) -print0 | while IFS= read -r -d "" f; do
        cp -a "$f" "$2/$f.incoming"
        mv -f "$2/$f.incoming" "$2/$f"
    done
' publish "$stage$dest" "$dest"
exit 0
