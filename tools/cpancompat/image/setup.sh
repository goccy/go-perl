#!/bin/sh
# Runs once while the image is built: makes gperl the perl command of the
# machine, the way `gperl xs build` arranges it for the build of one
# distribution.
set -e
G=/opt/gperl
mkdir -p $G/proj $G/fakearchlib $G/local/lib/perl5 $G/local/bin

cat > $G/bin/perl <<'EOF'
#!/bin/sh
GPERL_INTERNAL_PERL_CLI=1
export GPERL_INTERNAL_PERL_CLI
exec /opt/gperl/bin/gperl run "$@"
EOF
chmod +x $G/bin/perl

# The XS SDK (the headers native extensions compile against) is unpacked by
# any `gperl xs build`; the distribution named here does not exist.
$G/bin/gperl xs build -C $G/proj /nonexistent >/dev/null 2>&1 || true
test -f $G/proj/local/xs/.sdk/perl.h
ln -s $G/proj/local/xs/.sdk $G/fakearchlib/CORE

# The standard library is unpacked into the user's cache on first use; that
# directory is what %Config has to name, because it is the one in @INC.
stdlib=$($G/bin/perl -e 'require strict; my $p = $INC{"strict.pm"}; $p =~ s{/strict\.pm\z}{}; print $p')
test -f "$stdlib/strict.pm"
# The SDK has a config.h, so the Makefile ExtUtils::MakeMaker writes depends
# on Config.pm in the architecture library.
ln -s "$stdlib/Config.pm" $G/fakearchlib/Config.pm

# The compiler %Config describes is this machine's, in every key that names
# it, as `gperl xs build` describes it (hostToolchain in gperl/xs.go): build
# scripts choose flags by the compiler's identity, pass the preprocessor
# flags and libraries on to the C configure scripts they run, and look for
# libraries where the compiler looks. The interpreter's own values describe
# the wasm32 toolchain it was built with.
ccversion=$(cc -dumpfullversion)
# "libraries: =dir:dir:..." is the format of -print-search-dirs; the list
# keeps the directories that exist, once each, in order.
libpth=$(cc -print-search-dirs | sed -n 's/^libraries: =//p' | tr ':' '\n' |
    while read -r dir; do
        [ -n "$dir" ] && [ -d "$dir" ] && (cd "$dir" && pwd)
    done | awk '!seen[$0]++' | tr '\n' ' ' | sed 's/ $//')
test -n "$libpth"
# A name without a directory is the compiler's way of saying it did not
# find the file.
libc=$(cc -print-file-name=libc.so.6)
case $libc in
/*) test -e "$libc" || libc= ;;
*) libc= ;;
esac
glibc=$(printf '#include <features.h>\n__GLIBC__ __GLIBC_MINOR__\n' | cc -E -P -x c - | tail -1 | tr ' ' .)

sed -e "s|@STDLIB@|$stdlib|g" -e "s|@CCVERSION@|$ccversion|g" \
    -e "s|@LIBPTH@|$libpth|g" -e "s|@LIBC@|$libc|g" -e "s|@GLIBC@|$glibc|g" \
    > $G/config-overlay.pl <<'EOF'
use Config ();
my $ccflags = '-fno-strict-aliasing -fstack-protector-strong -DPERL_USE_SAFE_PUTENV -D_GNU_SOURCE -I/opt/gperl/proj/local/xs/.sdk';
my %over = (
    'perlpath' => '/opt/gperl/bin/perl', 'startperl' => '#!/opt/gperl/bin/perl',
    'cc' => 'cc', 'ld' => 'cc', 'ar' => 'ar', 'full_ar' => 'ar',
    'ranlib' => 'ranlib', 'make' => 'make',
    'ccname' => 'gcc', 'gccversion' => '@CCVERSION@', 'ccversion' => '',
    'gccosandvers' => '',
    'cpp' => 'cc -E', 'cpprun' => 'cc -E', 'cppstdin' => 'cc -E',
    'cppminus' => '-', 'cpplast' => '-', 'cppflags' => $ccflags,
    'ccsymbols' => '', 'cppsymbols' => '', 'cppccsymbols' => '',
    'libc' => '@LIBC@', 'gnulibc_version' => '@GLIBC@',
    'libs' => '-lm', 'perllibs' => '-lm',
    'libpth' => '@LIBPTH@', 'glibpth' => '@LIBPTH@', 'libspath' => '@LIBPTH@',
    'ccflags' => $ccflags,
    'optimize' => '-O2', 'ldflags' => '', 'cccdlflags' => '-fPIC',
    'ccdlflags' => '', 'obj_ext' => '.o', 'exe_ext' => '', 'lib_ext' => '.a',
    'usedl' => 'define', 'dlsrc' => 'dl_dlopen.xs', 'useshrplib' => 'false',
    'osname' => 'linux', 'so' => 'so', 'dlext' => 'so', 'lddlflags' => '-shared',
    'privlibexp' => '@STDLIB@', 'privlib' => '@STDLIB@',
    'archlibexp' => '/opt/gperl/fakearchlib', 'archlib' => '/opt/gperl/fakearchlib',
    'vendorlibexp' => '', 'vendorarchexp' => '',
    'installprivlib' => '@STDLIB@', 'installarchlib' => '/opt/gperl/fakearchlib',
);
# %Config stays the tied, read-only view it is on every perl; the values go
# into the Config object, where FETCH looks first.
my $config = tied %Config::Config;
$config->{$_} = $over{$_} for keys %over;
$^O = 'linux';
1;
EOF

GPERL_PERL_PRELOAD=$G/config-overlay.pl $G/bin/perl -MConfig -e 'print "perl $^V as $^X, cc=$Config{cc}\n"'
