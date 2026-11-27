//go:build darwin || linux

package xs_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

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

// The sets native code hands sigprocmask may be POSIX::SigSet objects' own,
// as Thread::SigMask hands them on: the set's buffer as a sigset_t. Those
// are the interpreter's sets, whose layout is not the host's, and they are
// read, and the set before written, as the interpreter has them: SIGUSR1
// blocked, not SIGUSR2 (the buffer read as a host set blocked SIGUSR2 and
// left SIGUSR1 to run its handler at once). What is expected is what perl
// 5.36 on Linux gives with the same module.
func TestNativeSignalMaskTakesThePOSIXSigSetsOfTheProgram(t *testing.T) {
	so := buildModuleSource(t, "SigMaskSet", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <signal.h>
/* sigprocmask with the sets of POSIX::SigSet objects, as Thread::SigMask
 * calls it: the set's buffer handed over as a sigset_t */
static sigset_t *sigset_of(pTHX_ SV *obj) {
    if (!SvOK(obj)) return NULL;
    return (sigset_t *)SvPV_nolen(SvRV(obj));
}
/* sigmask(how, set[, oldset]): sigprocmask's result */
XS(XS_SigMaskSet_sigmask) {
    dXSARGS;
    int r;
    if (items < 2) croak_xs_usage(cv, "how, set, oldset = undef");
    r = sigprocmask((int)SvIV(ST(0)), sigset_of(aTHX_ ST(1)), items > 2 ? sigset_of(aTHX_ ST(2)) : NULL);
    ST(0) = sv_2mortal(newSViv(r));
    XSRETURN(1);
}
XS_EXTERNAL(boot_SigMaskSet) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("SigMaskSet::sigmask", XS_SigMaskSet_sigmask, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `use POSIX qw(SIG_BLOCK SIG_UNBLOCK SIGUSR1 SIGUSR2);
my ($c1, $c2) = (0, 0);
local $SIG{USR1} = sub { $c1++ };
local $SIG{USR2} = sub { $c2++ };
my @r;
kill USR1 => $$;
push @r, "first $c1 $c2";
my $old = POSIX::SigSet->new;
push @r, "block " . SigMaskSet::sigmask(SIG_BLOCK, POSIX::SigSet->new(SIGUSR1), $old);
push @r, "old has usr1 " . ($old->ismember(SIGUSR1) ? 1 : 0);
kill USR1 => $$;
kill USR2 => $$;
push @r, "blocked $c1 $c2";
my $now = POSIX::SigSet->new;
SigMaskSet::sigmask(SIG_BLOCK, POSIX::SigSet->new, $now);
push @r, "mask has usr1 " . ($now->ismember(SIGUSR1) ? 1 : 0) . " usr2 " . ($now->ismember(SIGUSR2) ? 1 : 0);
push @r, "unblock " . SigMaskSet::sigmask(SIG_UNBLOCK, POSIX::SigSet->new(SIGUSR1));
push @r, "after $c1 $c2";
join ", ", @r`
	// The program is the host process, as gperl's is: POSIX's SIG_BLOCK is
	// the host's number, which native code hands sigprocmask as it is.
	if got, want := requireModuleWith(t, perl.Config{HostProcess: true}, "SigMaskSet", so, program), "first 1 0, block 0, old has usr1 0, blocked 1 1, mask has usr1 1 usr2 0, unblock 0, after 2 1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
