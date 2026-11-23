//go:build darwin || linux

package xs_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// A write of native code into a pipe or a socket nobody reads fails with
// EPIPE and raises the program's SIGPIPE, whose handler runs as the call
// returns, as on perl, where the kernel delivers it then; a send with
// MSG_NOSIGNAL raises none, and an ignored SIGPIPE leaves the errno alone.
// Before, the process's copy of the signal was dropped (the runtime takes
// the guest's own writes' SIGPIPE at the write) and the handler never ran
// (Log::Syslog::Fast's send to a unix stream server that went away). What
// is expected is what perl 5.36 on Linux gives with the same module.
func TestNativeWriteIntoAClosedSocketRaisesSIGPIPE(t *testing.T) {
	so := buildModuleSource(t, "NSigPipe", `#include <errno.h>
#include <sys/socket.h>
#include <sys/uio.h>
#include <unistd.h>
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* writes of native code into a socket whose other end is closed, as
 * Log::Syslog::Fast sends to a server that went away: how(1 write, 2
 * writev, 3 send, 4 send with MSG_NOSIGNAL) -> the errno's name */
XS(XS_NSigPipe_write_closed) {
    dXSARGS;
    int sv[2], how;
    ssize_t r = 0;
    char c = 'x';
    if (items != 1) croak_xs_usage(cv, "how");
    how = (int)SvIV(ST(0));
    if (socketpair(AF_UNIX, SOCK_STREAM, 0, sv) != 0) croak("socketpair failed");
    close(sv[1]);
    errno = 0;
    if (how == 1) r = write(sv[0], &c, 1);
    else if (how == 2) {
        struct iovec iov;
        iov.iov_base = &c;
        iov.iov_len = 1;
        r = writev(sv[0], &iov, 1);
    } else if (how == 3) r = send(sv[0], &c, 1, 0);
#ifdef MSG_NOSIGNAL
    else r = send(sv[0], &c, 1, MSG_NOSIGNAL);
#else
    else r = -1, errno = EPIPE;
#endif
    {
        const int e = errno;
        close(sv[0]);
        ST(0) = sv_2mortal(newSVpv(r < 0 && e == EPIPE ? "EPIPE" : "other", 0));
    }
    XSRETURN(1);
}
XS_EXTERNAL(boot_NSigPipe) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("NSigPipe::write_closed", XS_NSigPipe_write_closed, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $piped = 0;
local $SIG{PIPE} = sub { $piped++ };
my @r;
for my $how (1 .. 4) {
    my $e = NSigPipe::write_closed($how);
    push @r, "$how $e piped $piped";
}
{
    local $SIG{PIPE} = 'IGNORE';
    push @r, "ignored " . NSigPipe::write_closed(1) . " piped $piped";
}
join ", ", @r`
	// the program is the host process, as gperl's is: the process's own
	// SIGPIPE does not end it
	if got, want := requireModuleWith(t, perl.Config{HostProcess: true}, "NSigPipe", so, program), "1 EPIPE piped 1, 2 EPIPE piped 2, 3 EPIPE piped 3, 4 EPIPE piped 3, ignored EPIPE piped 3"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
