# Prints, as one JSON object, every module the running perl ships in its
# standard library together with its version (null when the file declares
# none). Run on the interpreter under survey:
#
#     perl provided.pl > provided.json
#
# The standard library is the first @INC entry holding strict.pm and the
# first holding Config.pm: on perl itself the second is the architecture
# library, where the modules with compiled parts are, and it lies inside the
# first; on gperl both are one directory.
use strict;
use warnings;
use File::Find ();
use ExtUtils::MakeMaker ();
use JSON::PP ();

my @roots;
for my $file ('strict.pm', 'Config.pm') {
    my ($dir) = grep { -f "$_/$file" } @INC
        or die "no directory holding $file in \@INC\n";
    push @roots, $dir unless grep { $_ eq $dir } @roots;
}

my %provided;
for my $root (@roots) {
    File::Find::find(
        {
            no_chdir => 1,
            wanted   => sub {
                # Another root below this one is scanned as a root of its
                # own, where its modules have their names.
                if (-d $_ && $_ ne $root && grep { $_ eq $File::Find::name } @roots) {
                    $File::Find::prune = 1;
                    return;
                }
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
}

print JSON::PP->new->canonical->pretty->encode(\%provided);
