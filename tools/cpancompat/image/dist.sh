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
# usage: dist.sh <archive> <build|shipped>
#
# "build" is the normal course. "shipped" is for a distribution the
# interpreter itself ships at the same or a newer version: nothing is built
# or installed, and its test files run against the interpreter's own copy.
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

# What CPAN clients set for the programs they drive.
export PERL_MM_USE_DEFAULT=1 NONINTERACTIVE_TESTING=1 PERL_USE_UNSAFE_INC=1
export PERL5_CPAN_IS_RUNNING=$$ PERL5_CPANPLUS_IS_RUNNING=$$

# phase <name> <limit in seconds> <command...> runs one phase with its output
# in <name>.log and appends its record to phases.jsonl.
phase() {
    local name=$1 limit=$2
    shift 2
    local start=$SECONDS status seconds timed_out=false
    timeout --kill-after=10 "$limit" "$@" > "$out/$name.log" 2>&1 < /dev/null
    status=$?
    seconds=$((SECONDS - start))
    # A command that outlived its limit was ended by timeout (status 124, or
    # 137 when it had to be killed); the duration tells that apart from a
    # command that returned such a status by itself.
    if [ "$seconds" -ge "$limit" ]; then timed_out=true; fi
    printf '{"phase":"%s","status":%d,"seconds":%d,"timed_out":%s}\n' \
        "$name" "$status" "$seconds" "$timed_out" >> "$out/phases.jsonl"
    return $status
}

mkdir -p /work/src "$stage"
case "$archive" in
    *.zip) phase unpack 300 unzip -q "$archive" -d /work/src ;;
    *) phase unpack 300 tar -xf "$archive" -C /work/src ;;
esac || exit 0
cd /work/src || exit 0
entries=(*)
if [ ${#entries[@]} -eq 1 ] && [ -d "${entries[0]}" ]; then
    cd "${entries[0]}" || exit 0
fi
chmod -R u+w .

builder=$(perl $tools/builder.pl 2> "$out/builder.log")
case "$builder" in
    Makefile.PL)
        phase configure 900 perl Makefile.PL INSTALL_BASE=$stage INSTALLMAN1DIR=none INSTALLMAN3DIR=none NO_PERLLOCAL=1 || exit 0
        made=Makefile
        ;;
    Build.PL)
        phase configure 900 perl Build.PL --install_base $stage || exit 0
        made=Build
        ;;
    *)
        printf '{"phase":"configure","status":-1,"seconds":0,"timed_out":false,"no_build_script":true}\n' >> "$out/phases.jsonl"
        exit 0
        ;;
esac
# A build script that ends successfully without writing its Makefile / Build
# declined this platform.
if [ ! -f "$made" ]; then
    printf '{"phase":"declined","status":0,"seconds":0,"timed_out":false}\n' >> "$out/phases.jsonl"
    exit 0
fi

# unmet.pl exits 3 when a requirement is not met; the details are in
# unmet.json either way.
phase prerequisites 300 bash -c "perl $tools/unmet.pl > $out/unmet.json" || exit 0

export HARNESS_SUBCLASS=CPANCompat::Harness CPANCOMPAT_TESTS=$out/tests.jsonl
if [ "$mode" = shipped ]; then
    if [ "$builder" = Makefile.PL ]; then
        # The test files the distribution declared, without the build that
        # `make test` would start with.
        files=$(make -s --eval='cpancompat-test-files:;@echo $(TEST_FILES)' cpancompat-test-files 2> "$out/test-files.log")
    else
        files='t/*.t'
    fi
    # shellcheck disable=SC2086  # the list is a set of glob patterns
    PERL5LIB=$PERL5LIB:$tools/lib phase test 1200 \
        perl -MExtUtils::Command::MM -MTest::Harness -e 'undef *Test::Harness::Switches; test_harness(0)' $files
    exit 0
fi

# An install step may load what it has just installed (XML::SAX registers its
# parser that way). A client that installs into a directory of its own has
# that directory in PERL5LIB, so the staging directory is there for the step.
stagelib=$stage/lib/perl5
if [ "$builder" = Makefile.PL ]; then
    phase build 1800 make || exit 0
    PERL5LIB=$PERL5LIB:$tools/lib phase test 1200 make test
    PERL5LIB=$stagelib:$PERL5LIB phase install 600 make install || exit 0
else
    phase build 1800 ./Build || exit 0
    PERL5LIB=$PERL5LIB:$tools/lib phase test 1200 ./Build test
    PERL5LIB=$stagelib:$PERL5LIB phase install 600 ./Build install || exit 0
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
' publish "$stage" "$dest"
exit 0
