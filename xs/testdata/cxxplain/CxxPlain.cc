/* A C++ module that includes the perl headers as Algorithm::LibLinear does:
 * after C++ headers of its own and outside extern "C". With CXXPLAIN_MIXED
 * it is one unit of a module whose other unit is C (helper.c), and an XSUB
 * returns what a function of that unit makes with the perl API. */
#include <string>

#include "EXTERN.h"
#include "XSUB.h"
#include "perl.h"

#ifdef CXXPLAIN_MIXED
extern "C" SV *cxxplain_twice(IV n);
#endif

namespace {
std::string greeting(const char *who) { return std::string("hello, ") + who; }
}

XS(XS_CxxPlain_greet) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "who");
    const std::string s = greeting(SvPV_nolen(ST(0)));
    ST(0) = sv_2mortal(newSVpvn(s.data(), s.size()));
    XSRETURN(1);
}

#ifdef CXXPLAIN_MIXED
XS(XS_CxxPlain_twice) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    ST(0) = sv_2mortal(cxxplain_twice(SvIV(ST(0))));
    XSRETURN(1);
}
#endif

XS_EXTERNAL(boot_CxxPlain) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    newXS_deffile("CxxPlain::greet", XS_CxxPlain_greet);
#ifdef CXXPLAIN_MIXED
    newXS_deffile("CxxPlain::twice", XS_CxxPlain_twice);
#endif
    Perl_xs_boot_epilog(aTHX_ ax);
}
