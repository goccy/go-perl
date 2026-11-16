/* A module built with the SDK whose layout of the interpreter's entries of
 * hashes disagrees with the interpreter's (wrong.c, linked with it, has
 * the module's __goperl_xs_layout): the loader refuses it. */
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

XS(XS_Layout_hello) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvs("hello"));
    XSRETURN(1);
}

XS_EXTERNAL(boot_Layout) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("Layout::hello", XS_Layout_hello);
    Perl_xs_boot_epilog(aTHX_ ax);
}
