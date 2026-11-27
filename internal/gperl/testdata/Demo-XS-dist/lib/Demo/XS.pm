package Demo::XS;

use strict;
use warnings;

our $VERSION = '0.01';

use XSLoader;
XSLoader::load(__PACKAGE__, $VERSION);

1;
