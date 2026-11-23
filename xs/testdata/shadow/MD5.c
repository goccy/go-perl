/*
 * A native extension with the name of one that is linked into the
 * interpreter: Digest::MD5, at a version of its own, whose md5_hex answers
 * with a marker instead of a digest. A program that gets the marker loaded
 * this shared object, not the linked-in extension.
 */
#define PERL_NO_GET_CONTEXT
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

XS_INTERNAL(XS_Digest__MD5_md5_hex); /* prototype to pass -Wmissing-prototypes */
XS_INTERNAL(XS_Digest__MD5_md5_hex)
{
    dVAR; dXSARGS;
    if (items != 1)
        croak_xs_usage(cv, "data");
    ST(0) = sv_2mortal(newSVpvf("shadow:%s", SvPV_nolen(ST(0))));
    XSRETURN(1);
}

#ifdef __cplusplus
extern "C"
#endif
XS_EXTERNAL(boot_Digest__MD5); /* prototype to pass -Wmissing-prototypes */
XS_EXTERNAL(boot_Digest__MD5)
{
    dVAR; dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(cv);
    PERL_UNUSED_VAR(items);
    newXS_deffile("Digest::MD5::md5_hex", XS_Digest__MD5_md5_hex);
    Perl_xs_boot_epilog(aTHX_ ax);
}
