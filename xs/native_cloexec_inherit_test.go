//go:build darwin || linux

package xs_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// A command native code starts (system(3)) inherits the program's
// descriptors as the program's close-on-exec flags say, and native code's
// fcntl reads the flag the program set: they are the kernel's, as in perl's
// process. The kernel kept every descriptor close-on-exec, so such a
// command got none, and native code saw the flag set after the program
// cleared it. The interpreter is the host process, whose descriptors the
// program's are. What is expected is what perl 5.36 on Linux gives with the
// same module.
const cloexecInheritModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <stdlib.h>
#include <sys/wait.h>
/* inherited(fd): whether a command system(3) starts has descriptor fd */
XS(XS_CloexecInherit_inherited) {
    dXSARGS;
    char cmd[64];
    int st;
    if (items != 1) croak_xs_usage(cv, "fd");
    snprintf(cmd, sizeof cmd, "test -e /dev/fd/%d", (int)SvIV(ST(0)));
    st = system(cmd);
    XSRETURN_IV(WIFEXITED(st) && WEXITSTATUS(st) == 0);
}
/* flag(fd): FD_CLOEXEC as native code's fcntl reads it */
XS(XS_CloexecInherit_flag) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fd");
    XSRETURN_IV((fcntl((int)SvIV(ST(0)), F_GETFD) & FD_CLOEXEC) != 0);
}
XS_EXTERNAL(boot_CloexecInherit) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("CloexecInherit::inherited", XS_CloexecInherit_inherited, __FILE__);
    newXS("CloexecInherit::flag", XS_CloexecInherit_flag, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCommandInheritsByTheProgramsFlags(t *testing.T) {
	so := buildModuleSource(t, "CloexecInherit", cloexecInheritModule)
	program := `use Fcntl qw(F_SETFD FD_CLOEXEC);
pipe my $r, my $w or die "pipe: $!";
my $n = fileno $r;
my @got;
fcntl($r, F_SETFD, 0) or die;
push @got, "cleared: flag " . CloexecInherit::flag($n) . ", inherited " . CloexecInherit::inherited($n);
fcntl($r, F_SETFD, FD_CLOEXEC) or die;
push @got, "set: flag " . CloexecInherit::flag($n) . ", inherited " . CloexecInherit::inherited($n);
join "; ", @got`
	if got, want := requireModuleWith(t, perl.Config{HostProcess: true}, "CloexecInherit", so, program), "cleared: flag 0, inherited 1; set: flag 1, inherited 0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
