# Prints, as one JSON object, every module the running perl ships in its
# standard library together with its version (null when the file declares
# none). Run on the interpreter under survey:
#
#     perl provided.pl > provided.json
#
# The first @INC entry holding strict.pm is taken as the library root.
use strict;
use warnings;
use File::Find ();
use ExtUtils::MakeMaker ();
use JSON::PP ();

my ($root) = grep { -f "$_/strict.pm" } @INC
    or die "no standard library directory in \@INC\n";

my %provided;
File::Find::find(
    {
        no_chdir => 1,
        wanted   => sub {
            return unless -f $_ && /\.pm\z/;
            my $module = substr($_, length($root) + 1, -3);
            $module =~ s{/}{::}g;
            my $version = eval { MM->parse_version($_) };
            $provided{$module} =
                defined $version && $version ne 'undef' ? "$version" : undef;
        },
    },
    $root,
);

print JSON::PP->new->canonical->pretty->encode(\%provided);
