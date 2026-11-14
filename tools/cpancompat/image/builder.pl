# Prints which build script of the distribution in the current directory to
# run: "Build.PL", "Makefile.PL", or nothing when it has neither.
#
# Build.PL whenever the distribution has one, as cpanm does and as CPAN.pm
# does by default (prefer_installer MB). Of a distribution that ships both,
# Makefile.PL is run while Module::Build is not installed, as CPAN.pm does
# (cpanm installs Module::Build first instead): its configure requirements
# need not name Module::Build when Makefile.PL works without it.
use strict;
use warnings;
use Module::Metadata ();

my ($mb, $mm) = (-f 'Build.PL', -f 'Makefile.PL');
$mb = 0 if $mb && $mm && !Module::Metadata->find_module_by_name('Module::Build');
print $mb ? "Build.PL" : $mm ? "Makefile.PL" : "";
