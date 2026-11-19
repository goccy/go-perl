//go:build darwin || linux

package xs_test

import "testing"

// chdir and fchdir of native code are the program's, as perl's are the
// process's: after POSIX::2008's chdir, perl's relative paths and getcwd
// are from the new directory, a relative path native code gives is from
// the program's directory, and a directory native code has a descriptor
// of is the one fchdir goes to. Before, native chdir moved the host process
// and the program stayed where it was. The expected string is what perl
// prints with the same module code.
const chdirSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <fcntl.h>

static SV *MyCd_result(pTHX_ int rc) {
    return newSVpv(rc == 0 ? "ok" : errno == ENOENT ? "ENOENT" : errno == ENOTDIR ? "ENOTDIR" : "other", 0);
}
XS(XS_MyCd_chdir_to) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "path");
    ST(0) = sv_2mortal(MyCd_result(aTHX_ chdir(SvPV_nolen(ST(0)))));
    XSRETURN(1);
}
/* a directory by its descriptor */
XS(XS_MyCd_fchdir_to) {
    dXSARGS;
    int fd, rc;
    if (items != 1) croak_xs_usage(cv, "path");
    fd = open(SvPV_nolen(ST(0)), O_RDONLY);
    if (fd < 0) croak("open: %s", Strerror(errno));
    rc = fchdir(fd);
    ST(0) = sv_2mortal(MyCd_result(aTHX_ rc));
    close(fd);
    XSRETURN(1);
}

XS_EXTERNAL(boot_MyCd) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyCd::chdir_to", XS_MyCd_chdir_to);
    newXS_deffile("MyCd::fchdir_to", XS_MyCd_fchdir_to);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeChdirIsTheProgramsDirectory(t *testing.T) {
	so := buildModuleSource(t, "MyCd", chdirSource)
	got := requireModule(t, "MyCd", so, `use Cwd qw(getcwd abs_path);
use File::Temp qw(tempdir);
my $d = abs_path(tempdir(CLEANUP => 1));
mkdir "$d/sub" or die "mkdir: $!";
{ open my $fh, ">", "$d/sub/f" or die "open: $!"; close $fh }
my $start = getcwd();
my @r;
push @r, MyCd::chdir_to("$d/sub"), (getcwd() eq "$d/sub" ? "in sub" : "elsewhere"), (-e "f" ? "f here" : "no f");
push @r, MyCd::chdir_to(".."), (getcwd() eq $d ? "in d" : "elsewhere");
push @r, MyCd::fchdir_to("$d/sub"), (getcwd() eq "$d/sub" ? "in sub" : "elsewhere");
push @r, MyCd::fchdir_to("$d/sub/f"), MyCd::chdir_to("$d/none"), MyCd::chdir_to("");
chdir $start or die "chdir: $!";
join "|", @r`)
	want := "ok|in sub|f here|ok|in d|ok|in sub|ENOTDIR|ENOENT|ENOENT"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
