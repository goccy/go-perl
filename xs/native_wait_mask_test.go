//go:build darwin || linux

package xs_test

import (
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// A wait of native code with a signal mask of its own (pselect, ppoll,
// epoll_pwait) has that mask in place while it waits, as the kernel puts
// it in place: a signal the program blocks with sigprocmask and the wait's
// mask lets in ends the wait with EINTR, and its handler runs, as
// IO::Async::Loop::Epoll waits for SIGCHLD. The mask is POSIX::SigSet's,
// handed on as it is, or a sigset_t of the module's own. Before, the mask
// was the kernel's alone: the program's signal stayed blocked for the
// whole wait, which ran to its timeout (IO-Async-Loop-Epoll, IO::Async::
// Process under the Epoll loop, IO-Ppoll).
func TestNativeWaitWithASignalMask(t *testing.T) {
	so := buildModuleSource(t, "WaitMask", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <poll.h>
#include <sys/select.h>
#include <time.h>
#include <unistd.h>
#ifdef __linux__
#include <sys/epoll.h>
#endif
static double wait_now(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return ts.tv_sec + ts.tv_nsec / 1e9;
}
/* wait(how, set): how is pselect, ppoll or epoll_pwait; set a POSIX::SigSet,
 * or undef for an empty sigset_t of the module's own. Waits up to 5 s for a
 * pipe nobody writes. */
XS(XS_WaitMask_wait) {
    dXSARGS;
    const char *how = SvPV_nolen(ST(0));
    sigset_t own;
    const sigset_t *mask = &own;
    int fds[2], n = -2, e = 0;
    double t0;
    PERL_UNUSED_VAR(items);
    if (SvOK(ST(1))) mask = (const sigset_t *)SvPV_nolen(SvRV(ST(1)));
    else sigemptyset(&own);
    if (pipe(fds) != 0) croak("pipe: %s", strerror(errno));
    t0 = wait_now();
    if (strEQ(how, "pselect")) {
        struct timespec ts = {5, 0};
        fd_set r;
        FD_ZERO(&r);
        FD_SET(fds[0], &r);
        n = pselect(fds[0] + 1, &r, NULL, NULL, &ts, mask);
    }
#ifdef __linux__
    else if (strEQ(how, "ppoll")) {
        struct timespec ts = {5, 0};
        struct pollfd p = {fds[0], POLLIN, 0};
        n = ppoll(&p, 1, &ts, mask);
    } else if (strEQ(how, "epoll_pwait")) {
        struct epoll_event ev, out;
        const int ep = epoll_create1(0);
        ev.events = EPOLLIN;
        ev.data.fd = fds[0];
        epoll_ctl(ep, EPOLL_CTL_ADD, fds[0], &ev);
        n = epoll_pwait(ep, &out, 1, 5000, mask);
        e = errno;
        close(ep);
        errno = e;
    }
#endif
    e = errno;
    close(fds[0]);
    close(fds[1]);
    ST(0) = sv_2mortal(newSVpvf("%d %s %.1f", n, n < 0 && e == EINTR ? "EINTR" : "-",
                                wait_now() - t0));
    XSRETURN(1);
}
XS_EXTERNAL(boot_WaitMask) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("WaitMask::wait", XS_WaitMask_wait);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	hows := []string{"pselect"}
	if runtime.GOOS == "linux" {
		hows = append(hows, "ppoll", "epoll_pwait")
	}
	for _, how := range hows {
		for _, set := range []string{`$old`, `undef`} {
			t.Run(how+" "+set, func(t *testing.T) {
				got := requireModule(t, "WaitMask", so, `use POSIX ();
my $got = 0;
local $SIG{ALRM} = sub { $got++ };
my $old = POSIX::SigSet->new;
POSIX::sigprocmask(POSIX::SIG_BLOCK(), POSIX::SigSet->new(POSIX::SIGALRM()), $old);
alarm 1;
my $r = WaitMask::wait("`+how+`", `+set+`);
my $during = $got;
POSIX::sigprocmask(POSIX::SIG_SETMASK(), $old);
"$r handler:$during"`)
				f := strings.Fields(got)
				if len(f) != 4 || f[0] != "-1" || f[1] != "EINTR" || f[3] != "handler:1" {
					t.Fatalf("got %q, want the wait ended by the signal (-1 EINTR) and its handler run", got)
				}
				if secs, err := strconv.ParseFloat(f[2], 64); err != nil || secs > 3 {
					t.Fatalf("got %q: the wait ran %s s, not ended by the signal after 1 s", got, f[2])
				}
			})
		}
	}
}
