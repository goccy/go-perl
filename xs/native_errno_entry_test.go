//go:build darwin || linux

package xs_test

import "testing"

// An XSUB starts with the interpreter's errno in errno, as on perl, where
// errno and $! are one variable: what one XSUB left there is what the next
// one finds, and so is what the program put in $! (Affix's errno() read
// GOPERL_ERRNO_UNSET, -6778736, after the call that set ENOENT). What is
// expected is what perl 5.36 on Linux gives with the same module.
func TestNativeCodeFindsTheInterpretersErrno(t *testing.T) {
	so := buildModuleSource(t, "ErrnoAt", `#include <errno.h>
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* errno as one XSUB leaves it and as the next one finds it */
XS(XS_ErrnoAt_set_enoent) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    errno = ENOENT;
    XSRETURN_EMPTY;
}
/* errno_name(): the name of errno as this XSUB starts, or its number */
XS(XS_ErrnoAt_errno_name) {
    dXSARGS;
    const int e = errno;
    PERL_UNUSED_VAR(items);
    if (e == ENOENT) ST(0) = sv_2mortal(newSVpvs("ENOENT"));
    else if (e == EACCES) ST(0) = sv_2mortal(newSVpvs("EACCES"));
    else ST(0) = sv_2mortal(newSViv(e));
    XSRETURN(1);
}
XS_EXTERNAL(boot_ErrnoAt) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("ErrnoAt::set_enoent", XS_ErrnoAt_set_enoent, __FILE__);
    newXS("ErrnoAt::errno_name", XS_ErrnoAt_errno_name, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `use POSIX ();
my @r;
ErrnoAt::set_enoent();
push @r, "after set " . ($!{ENOENT} ? "ENOENT" : 0 + $!);
push @r, "next xsub " . ErrnoAt::errno_name();
$! = POSIX::EACCES();
push @r, "after \$! " . ErrnoAt::errno_name();
$! = 0;
push @r, "cleared " . ErrnoAt::errno_name();
push @r, "\$! still " . (0 + $!);
join ", ", @r`
	if got, want := requireModule(t, "ErrnoAt", so, program), "after set ENOENT, next xsub ENOENT, after $! EACCES, cleared 0, $! still 0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
