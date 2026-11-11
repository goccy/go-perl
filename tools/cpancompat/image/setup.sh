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

sed "s|@STDLIB@|$stdlib|g" > $G/config-overlay.pl <<'EOF'
use Config ();
my %over = (
    'perlpath' => '/opt/gperl/bin/perl', 'startperl' => '#!/opt/gperl/bin/perl',
    'cc' => 'cc', 'ld' => 'cc', 'ar' => 'ar', 'full_ar' => 'ar',
    'ranlib' => 'ranlib', 'make' => 'make',
    'ccflags' => '-fno-strict-aliasing -fstack-protector-strong -DPERL_USE_SAFE_PUTENV -D_GNU_SOURCE -I/opt/gperl/proj/local/xs/.sdk',
    'optimize' => '-O2', 'ldflags' => '', 'cccdlflags' => '-fPIC',
    'ccdlflags' => '', 'obj_ext' => '.o', 'exe_ext' => '', 'lib_ext' => '.a',
    'usedl' => 'define', 'dlsrc' => 'dl_dlopen.xs', 'useshrplib' => 'false',
    'osname' => 'linux', 'so' => 'so', 'dlext' => 'so', 'lddlflags' => '-shared',
    'privlibexp' => '@STDLIB@', 'privlib' => '@STDLIB@',
    'archlibexp' => '/opt/gperl/fakearchlib', 'archlib' => '/opt/gperl/fakearchlib',
    'sitelibexp' => '', 'sitearchexp' => '', 'vendorlibexp' => '', 'vendorarchexp' => '',
    'installprivlib' => '@STDLIB@', 'installarchlib' => '/opt/gperl/fakearchlib',
);
my %full = %Config::Config;
untie %Config::Config;
%Config::Config = (%full, %over);
$^O = 'linux';
1;
EOF

GPERL_PERL_PRELOAD=$G/config-overlay.pl $G/bin/perl -MConfig -e 'print "perl $^V as $^X, cc=$Config{cc}\n"'
