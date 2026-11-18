//go:build darwin || linux

package xs_test

import "testing"

// sigprocmask and pthread_sigmask of native code change the program's
// signal mask, as they change the mask of the one thread a process of
// perl's has (Thread::SigMask): SIGUSR1 blocked, a kill of the program's own
// pid held until native code unblocks it, and the mask before as the
// program's. Thread::SigMask's blocked SIGUSR1 ran the handler at once. What
// is expected is what perl 5.36 on Linux gives with the same module.
func TestNativeSignalMaskIsTheProgramsMask(t *testing.T) {
	so := buildModuleSource(t, "SigMask", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <signal.h>
#include <pthread.h>
/* block($how, $use_pthread): SIGUSR1 blocked (how 0) or unblocked (how 1)
 * with pthread_sigmask or sigprocmask, as Thread::SigMask's sigmask does;
 * whether SIGUSR1 was blocked before */
XS(XS_SigMask_mask) {
    dXSARGS;
    sigset_t set, old;
    int how = SvIV(ST(0)) == 0 ? SIG_BLOCK : SIG_UNBLOCK;
    PERL_UNUSED_VAR(items);
    sigemptyset(&set);
    sigaddset(&set, SIGUSR1);
    if (SvTRUE(ST(1))) {
        if (pthread_sigmask(how, &set, &old) != 0) croak("pthread_sigmask failed");
    } else {
        if (sigprocmask(how, &set, &old) != 0) croak("sigprocmask: %s", strerror(errno));
    }
    ST(0) = sv_2mortal(newSViv(sigismember(&old, SIGUSR1) == 1 ? 1 : 0));
    XSRETURN(1);
}
/* blocked(): whether SIGUSR1 is blocked now, as pthread_sigmask says */
XS(XS_SigMask_blocked) {
    dXSARGS;
    sigset_t old;
    PERL_UNUSED_VAR(items);
    sigemptyset(&old);
    pthread_sigmask(SIG_BLOCK, NULL, &old);
    ST(0) = sv_2mortal(newSViv(sigismember(&old, SIGUSR1) == 1 ? 1 : 0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_SigMask) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("SigMask::mask", XS_SigMask_mask, __FILE__);
    newXS("SigMask::blocked", XS_SigMask_blocked, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $got = 0;
local $SIG{USR1} = sub { $got++ };
my @r;
for my $pthread (1, 0) {
    $got = 0;
    my $was = SigMask::mask(0, $pthread);
    my $now = SigMask::blocked();
    kill USR1 => $$;
    my $held = $got;
    my $was2 = SigMask::mask(1, $pthread);
    push @r, "was $was blocked $now held $held was2 $was2 after $got now " . SigMask::blocked();
}
join " | ", @r`
	if got, want := requireModule(t, "SigMask", so, program), "was 0 blocked 1 held 0 was2 1 after 1 now 0 | was 0 blocked 1 held 0 was2 1 after 1 now 0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
