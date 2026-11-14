# Prints, as one JSON object, the requirements of the configured
# distribution in the current directory that the installed modules do not
# meet: {"unmet":[{"module":..., "wanted":..., "have":...}, ...]}. It
# exits 0 whenever it could tell: a distribution whose requirements are not
# all met is still built and tested, as CPAN.pm does (it warns and goes on);
# the driver puts a failure down to what it lacked.
#
# The requirements are the ones the distribution computed while it was
# configured (MYMETA), which can differ from the static ones the dependency
# graph was built from. "have" is absent when the module is not installed.
use strict;
use warnings;
use CPAN::Meta ();
use JSON::PP ();
use Module::Metadata ();

my @unmet;
my ($file) = grep { -f } qw(MYMETA.json MYMETA.yml);
if ($file) {
    my $meta = CPAN::Meta->load_file($file, { lazy_validation => 1 });
    my $reqs = $meta->effective_prereqs->merged_requirements(
        [qw(configure build test runtime)], ['requires']);
    for my $module (sort $reqs->required_modules) {
        my $wanted = $reqs->requirements_for_module($module);
        if ($module eq 'perl') {
            push @unmet, { module => 'perl', wanted => "$wanted", have => "$]" }
                unless $reqs->accepts_module(perl => $]);
            next;
        }
        my $info = Module::Metadata->new_from_module($module);
        if (!$info) {
            push @unmet, { module => $module, wanted => "$wanted" };
            next;
        }
        my $have = $info->version($module);
        $have = defined $have ? "$have" : 0;
        push @unmet, { module => $module, wanted => "$wanted", have => "$have" }
            unless eval { $reqs->accepts_module($module => $have) };
    }
}
print JSON::PP->new->canonical->encode({ unmet => \@unmet }), "\n";
