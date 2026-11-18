//go:build darwin || linux

package xs_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// perl.h's suffixes of the mode of a handle on a socket and on a pipe
// (SOCKET_OPEN_MODE, PIPE_OPEN_MODE, PIPESOCK_MODE) and handy.h's size of
// a string's buffer (PERL_STRLEN_ROUNDUP) are there for native code, as
// POSIX::2008 opens a descriptor with "r+" SOCKET_OPEN_MODE and sizes its
// strings with PERL_STRLEN_ROUNDUP. Before, they were not defined, and
// such code did not compile or did not load.
// The expected string is what perl prints with the same module code.
const openModesSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

/* a pipe's two descriptors as handles of the modes perl.h gives, and the
 * sizes of buffers */
XS(XS_MyModes_pipe_handles) {
    dXSARGS;
    int fds[2];
    PerlIO *r, *w;
    PERL_UNUSED_VAR(items);
    if (pipe(fds) != 0) croak("pipe: %s", Strerror(errno));
    r = PerlIO_fdopen(fds[0], "r" PIPE_OPEN_MODE);
    w = PerlIO_fdopen(fds[1], "w" PIPE_OPEN_MODE);
    if (!r || !w) croak("fdopen");
    PerlIO_puts(w, "through");
    PerlIO_close(w);
    {
        char buf[16] = {0};
        SSize_t n = PerlIO_read(r, buf, sizeof buf - 1);
        PerlIO_close(r);
        ST(0) = sv_2mortal(newSVpvf("%s|%s|%s|%.*s|%d %d %d %d", "r+" SOCKET_OPEN_MODE,
                                    "r" PIPE_OPEN_MODE, "w" PIPESOCK_MODE, (int)(n < 0 ? 0 : n), buf,
                                    (int)PERL_STRLEN_ROUNDUP(1), (int)PERL_STRLEN_ROUNDUP(8),
                                    (int)PERL_STRLEN_ROUNDUP(9), (int)PERL_STRLEN_ROUNDUP(17)));
    }
    XSRETURN(1);
}

XS_EXTERNAL(boot_MyModes) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyModes::pipe_handles", XS_MyModes_pipe_handles);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativePerlHNamesOfHandlesAndBuffers(t *testing.T) {
	so := buildModuleSource(t, "MyModes", openModesSource)
	// the descriptors native code makes are the program's, as for perl
	got := requireModuleWith(t, perl.Config{HostProcess: true}, "MyModes", so, `MyModes::pipe_handles()`)
	want := "r+|r|w|through|8 8 16 24"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
