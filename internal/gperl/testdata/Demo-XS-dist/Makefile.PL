# A minimal ExtUtils::MakeMaker XS distribution used by the gperl xs build
# tests: the standard Makefile.PL + make pipeline, nothing custom.
use strict;
use warnings;
use ExtUtils::MakeMaker;

WriteMakefile(
    NAME         => 'Demo::XS',
    VERSION_FROM => 'lib/Demo/XS.pm',
);
