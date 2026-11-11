# Prints which build script of the distribution in the current directory to
# run: "Build.PL", "Makefile.PL", or nothing when it has neither.
#
# A distribution that ships both is configured through Build.PL only when
# its metadata asks for Module::Build at configure time, the rule CPAN
# clients apply.
use strict;
use warnings;
use CPAN::Meta ();

my ($mb, $mm) = (-f 'Build.PL', -f 'Makefile.PL');
if ($mb && $mm) {
    my ($file) = grep { -f } qw(META.json META.yml);
    my $meta = $file && eval { CPAN::Meta->load_file($file, { lazy_validation => 1 }) };
    my $asks = $meta && grep { $_ eq 'Module::Build' }
        $meta->effective_prereqs->requirements_for('configure', 'requires')->required_modules;
    $mb = 0 unless $asks;
}
print $mb ? "Build.PL" : $mm ? "Makefile.PL" : "";
