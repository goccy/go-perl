package CPANCompat::Harness;

# The test harness of a distribution under measurement: TAP::Harness, plus a
# record of every test file's outcome in the file CPANCOMPAT_TESTS names
# (one JSON object per line). Selected through HARNESS_SUBCLASS, which both
# `make test` and `./Build test` honour.

use strict;
use warnings;
use parent 'TAP::Harness';
use JSON::PP ();

sub summary {
    my ($self, $aggregate, @rest) = @_;
    my $path = $ENV{CPANCOMPAT_TESTS};
    if (defined $path && length $path && open(my $out, '>>', $path)) {
        my $json = JSON::PP->new->canonical;
        for my $file ($aggregate->descriptions) {
            my ($parser) = $aggregate->parsers($file);
            next unless $parser;
            print {$out} $json->encode({
                file         => "$file",
                planned      => _count($parser->tests_planned),
                run          => _count($parser->tests_run),
                passed       => scalar(my @p = $parser->passed),
                failed       => scalar(my @f = $parser->failed),
                todo_passed  => scalar(my @t = $parser->todo_passed),
                skipped      => scalar(my @s = $parser->skipped),
                skip_all     => $parser->skip_all ? JSON::PP::true : JSON::PP::false,
                parse_errors => scalar(my @e = $parser->parse_errors),
                exit         => _count($parser->exit),
                wait         => _count($parser->wait),
                problems     => $parser->has_problems ? JSON::PP::true : JSON::PP::false,
            }), "\n";
        }
        close $out;
    }
    return $self->SUPER::summary($aggregate, @rest);
}

sub _count { defined $_[0] ? $_[0] + 0 : 0 }

1;
