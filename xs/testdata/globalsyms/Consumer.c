/*
 * A native extension that calls a function another extension exports
 * (Provider's goperl_test_provider_twice). It is linked without that
 * extension: the dynamic linker finds the function when Consumer is
 * loaded, among the symbols of the modules loaded as global before it.
 */
#define PERL_NO_GET_CONTEXT
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

int goperl_test_provider_twice(int n);

XS_INTERNAL(XS_Consumer_twice);
XS_INTERNAL(XS_Consumer_twice)
{
    dVAR; dXSARGS;
    if (items != 1)
        croak_xs_usage(cv, "n");
    ST(0) = sv_2mortal(newSViv(goperl_test_provider_twice((int)SvIV(ST(0)))));
    XSRETURN(1);
}

#ifdef __cplusplus
extern "C"
#endif
XS_EXTERNAL(boot_Consumer);
XS_EXTERNAL(boot_Consumer)
{
    dVAR; dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(cv);
    PERL_UNUSED_VAR(items);
    newXS_deffile("Consumer::twice", XS_Consumer_twice);
    Perl_xs_boot_epilog(aTHX_ ax);
}
