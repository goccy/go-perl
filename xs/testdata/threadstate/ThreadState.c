/* A module's own thread-local variable, set in one call and read in a
 * later one, and the thread each call runs on: on perl every call of a
 * module runs on the program's one thread. */
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <pthread.h>

static __thread IV tls_value;

/* tls_set(n): the calling thread's variable gets n; answers the thread */
XS(XS_ThreadState_tls_set) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    tls_value = SvIV(ST(0));
    ST(0) = sv_2mortal(newSVuv((UV)(uintptr_t)pthread_self()));
    XSRETURN(1);
}

/* tls_get(): the calling thread's variable */
XS(XS_ThreadState_tls_get) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(tls_value));
    XSRETURN(1);
}

/* thread_id(): the thread the call runs on */
XS(XS_ThreadState_thread_id) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVuv((UV)(uintptr_t)pthread_self()));
    XSRETURN(1);
}

XS_EXTERNAL(boot_ThreadState) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("ThreadState::tls_set", XS_ThreadState_tls_set);
    newXS_deffile("ThreadState::tls_get", XS_ThreadState_tls_get);
    newXS_deffile("ThreadState::thread_id", XS_ThreadState_thread_id);
    Perl_xs_boot_epilog(aTHX_ ax);
}
