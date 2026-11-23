/* The C unit of the mixed CxxPlain module: a function of its own that uses
 * the perl API, called by an XSUB of the C++ unit. */
#include "EXTERN.h"
#include "perl.h"

SV *cxxplain_twice(IV n) { return newSViv(2 * n); }
