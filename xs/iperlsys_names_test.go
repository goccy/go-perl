//go:build darwin || linux

package xs_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// The calls of perl's iperlsys.h for native code, as perl has them without
// PERL_IMPLICIT_SYS: a directory made, listed, rewound and removed through
// PerlDir_*, a link made and read back through PerlLIO_*, the ids and byte
// order calls, and a jump through PerlProc_setjmp and PerlProc_longjmp. Punk
// and Time::Keeper::Natural did not load without them ("undefined symbol:
// PerlDir_mkdir"). What is expected is what perl 5.36 on Linux gives with
// the same module.
// The program is the host process, whose ids native code sees.
func TestNativeCodeCallsTheSystemThroughPerlsNames(t *testing.T) {
	so := buildModuleSource(t, "IperlSys", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* dir(path): makes path, lists it after a file and a link in it, reads
 * the link back, and removes all of it through the PerlDir_ and PerlLIO_
 * calls, as Punk and Time::Keeper::Natural do; the names sorted */
XS(XS_IperlSys_dir) {
    dXSARGS;
    const char *path = SvPV_nolen(ST(0));
    char file[1024], link[1024], buf[1024];
    DIR *d;
    Direntry_t *e;
    AV *names = newAV();
    SV *out;
    SSize_t n;
    PERL_UNUSED_VAR(items);
    if (PerlDir_mkdir(path, 0755) != 0) croak("mkdir: %s", strerror(errno));
    snprintf(file, sizeof file, "%s/a", path);
    PerlLIO_close(PerlLIO_open3(file, O_CREAT | O_WRONLY, 0644));
    snprintf(link, sizeof link, "%s/l", path);
    if (PerlLIO_symlink("a", link) != 0) croak("symlink: %s", strerror(errno));
    n = PerlLIO_readlink(link, buf, sizeof buf - 1);
    buf[n < 0 ? 0 : n] = '\0';
    d = PerlDir_open(path);
    if (!d) croak("opendir: %s", strerror(errno));
    while ((e = PerlDir_read(d)) != NULL)
        av_push(names, newSVpv(e->d_name, 0));
    PerlDir_rewind(d);
    PerlDir_close(d);
    PerlLIO_unlink(link);
    PerlLIO_unlink(file);
    if (PerlDir_rmdir(path) != 0) croak("rmdir: %s", strerror(errno));
    out = newSVpvf("%s|", buf);
    {
        /* sorted by name, to compare */
        SSize_t i, j, c = av_top_index(names) + 1;
        for (i = 0; i < c; i++)
            for (j = i + 1; j < c; j++) {
                SV **a = av_fetch(names, i, 0), **b = av_fetch(names, j, 0);
                if (strcmp(SvPV_nolen(*a), SvPV_nolen(*b)) > 0) {
                    SV *t = SvREFCNT_inc(*a);
                    av_store(names, i, SvREFCNT_inc(*b));
                    av_store(names, j, t);
                }
            }
        for (i = 0; i < c; i++)
            sv_catpvf(out, "%s%s", i ? "," : "", SvPV_nolen(*av_fetch(names, i, 0)));
    }
    SvREFCNT_dec((SV *)names);
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
/* ids(): the ids and the byte order calls, against what Perl says */
XS(XS_IperlSys_ids) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%d,%d,%d,%x,%x", (int)PerlProc_geteuid(), (int)PerlProc_getgid(),
                                (int)PerlProc_getegid(), (unsigned)PerlSock_ntohs(PerlSock_htons(0x1234)),
                                (unsigned)PerlSock_ntohl(PerlSock_htonl(0x01020304))));
    XSRETURN(1);
}
/* jump(): a Sigsetjmp the code jumps back to */
XS(XS_IperlSys_jump) {
    dXSARGS;
    Sigjmp_buf env;
    volatile int n = 0;
    PERL_UNUSED_VAR(items);
    if (PerlProc_setjmp(env, 1) < 3) {
        n++;
        PerlProc_longjmp(env, n);
    }
    ST(0) = sv_2mortal(newSViv(n));
    XSRETURN(1);
}
XS_EXTERNAL(boot_IperlSys) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("IperlSys::dir", XS_IperlSys_dir, __FILE__);
    newXS("IperlSys::ids", XS_IperlSys_ids, __FILE__);
    newXS("IperlSys::jump", XS_IperlSys_jump, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $d = (-d "/tmp" ? "/tmp" : ".") . "/iperlsys-$$";
my $ids = IperlSys::ids();
my $want = sprintf "%d,%d,%d,1234,1020304", $>, $( + 0, $) + 0;
join ";", IperlSys::dir($d), ($ids eq $want ? "ids" : "ids $ids vs $want"), IperlSys::jump(), (-e $d ? "left" : "gone")`
	if got, want := requireModuleWith(t, perl.Config{HostProcess: true}, "IperlSys", so, program), "a|.,..,a,l;ids;3;gone"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
