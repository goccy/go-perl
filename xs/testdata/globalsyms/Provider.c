/*
 * A native extension that exports a C function for other extensions to
 * link against (its module file asks for global symbols with
 * dl_load_flags 0x01), and keeps a counter of its own that the function
 * reads, so that a consumer calling it sees this module's state.
 */
#define PERL_NO_GET_CONTEXT
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

static int provider_calls = 0;

/* The function another extension calls. */
int goperl_test_provider_twice(int n);
int goperl_test_provider_twice(int n)
{
    provider_calls++;
    return n * 2;
}

XS_INTERNAL(XS_Provider_calls);
XS_INTERNAL(XS_Provider_calls)
{
    dVAR; dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(provider_calls));
    XSRETURN(1);
}

#ifdef __cplusplus
extern "C"
#endif
XS_EXTERNAL(boot_Provider);
XS_EXTERNAL(boot_Provider)
{
    dVAR; dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(cv);
    PERL_UNUSED_VAR(items);
    newXS_deffile("Provider::calls", XS_Provider_calls);
    Perl_xs_boot_epilog(aTHX_ ax);
}
