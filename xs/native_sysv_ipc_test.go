//go:build darwin || linux

package xs_test

import (
	"runtime"
	"testing"

	perl "github.com/goccy/go-perl"
)

// Native code uses the host's System V IPC as it does with perl: config.h
// says HAS_SEM, HAS_MSG and HAS_SHM, perl.h gives union semun and Semctl,
// and what native code does to a semaphore, a segment or a queue the
// program made is what the program then finds there (IPC::SysV's own XS,
// built from the CPAN dist, croaked "System V semxxx is not implemented on
// this machine" in IPC::Semaphore's stat). The interpreter is the host
// process, whose System V IPC objects the program uses. What is expected
// is what perl 5.36 on Linux gives with the same module.
const sysvNativeModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <sys/msg.h>
#include <sys/shm.h>
/* has(): what config.h says of System V IPC */
XS(XS_SysVNative_has) {
    dXSARGS;
    SV *r = newSVpvs("");
    PERL_UNUSED_VAR(items);
#ifdef HAS_SEM
    sv_catpvs(r, "sem ");
#endif
#ifdef HAS_MSG
    sv_catpvs(r, "msg ");
#endif
#ifdef HAS_SHM
    sv_catpvs(r, "shm");
#endif
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
/* sem_set(id, num, val): SETVAL with a union semun, through Semctl */
XS(XS_SysVNative_sem_set) {
    dXSARGS;
    union semun arg;
    if (items != 3) croak_xs_usage(cv, "id, num, val");
    arg.val = (int)SvIV(ST(2));
    XSRETURN_IV(Semctl((int)SvIV(ST(0)), (int)SvIV(ST(1)), SETVAL, arg));
}
/* shm_read(id, len): the segment's first len bytes, read through an
 * attachment of native code's own */
XS(XS_SysVNative_shm_read) {
    dXSARGS;
    Shmat_t p;
    SV *r;
    if (items != 2) croak_xs_usage(cv, "id, len");
    p = shmat((int)SvIV(ST(0)), NULL, SHM_RDONLY);
    if (p == (Shmat_t)-1) croak("shmat: %s", strerror(errno));
    r = newSVpvn((const char *)p, (STRLEN)SvUV(ST(1)));
    shmdt(p);
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
/* msg_send(id, type, text) */
XS(XS_SysVNative_msg_send) {
    dXSARGS;
    struct { long mtype; char mtext[64]; } m;
    STRLEN len;
    const char *text;
    if (items != 3) croak_xs_usage(cv, "id, type, text");
    m.mtype = (long)SvIV(ST(1));
    text = SvPV(ST(2), len);
    if (len > sizeof m.mtext) len = sizeof m.mtext;
    memcpy(m.mtext, text, len);
    XSRETURN_IV(msgsnd((int)SvIV(ST(0)), &m, len, 0));
}
XS_EXTERNAL(boot_SysVNative) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("SysVNative::has", XS_SysVNative_has, __FILE__);
    newXS("SysVNative::sem_set", XS_SysVNative_sem_set, __FILE__);
    newXS("SysVNative::shm_read", XS_SysVNative_shm_read, __FILE__);
    newXS("SysVNative::msg_send", XS_SysVNative_msg_send, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// On darwin a segment is not attached to the interpreter's memory: perl's
// shmwrite, which attaches it, fails with ENOSYS.
func TestNativeSystemVIPCIsTheHosts(t *testing.T) {
	so := buildModuleSource(t, "SysVNative", sysvNativeModule)
	shm := `shmwrite($shm, "native", 0, 6) or die "shmwrite: $!";
my $read = SysVNative::shm_read($shm, 6);`
	wantShm := "native"
	if runtime.GOOS == "darwin" {
		shm = `my $read = shmwrite($shm, "native", 0, 6) ? "written" : $!{ENOSYS} ? "ENOSYS" : "$!";`
		wantShm = "ENOSYS"
	}
	program := `use IPC::SysV qw(IPC_PRIVATE IPC_RMID GETVAL S_IRUSR S_IWUSR);
my $sem = semget(IPC_PRIVATE, 2, S_IRUSR | S_IWUSR) // die "semget: $!";
my $set = SysVNative::sem_set($sem, 1, 7);
my $val = semctl($sem, 1, GETVAL, 0);
semctl($sem, 0, IPC_RMID, 0);
my $shm = shmget(IPC_PRIVATE, 4096, S_IRUSR | S_IWUSR) // die "shmget: $!";
` + shm + `
shmctl($shm, IPC_RMID, 0);
my $msg = msgget(IPC_PRIVATE, S_IRUSR | S_IWUSR) // die "msgget: $!";
my $sent = SysVNative::msg_send($msg, 3, "hi");
msgrcv($msg, my $buf, 64, 0, 0) or die "msgrcv: $!";
my ($type, $text) = unpack "l! a*", $buf;
msgctl($msg, IPC_RMID, 0);
"has " . SysVNative::has() . "; sem $set $val; shm $read; msg $sent $type $text"`
	if got, want := requireModuleWith(t, perl.Config{HostProcess: true}, "SysVNative", so, program), "has sem msg shm; sem 0 7; shm "+wantShm+"; msg 0 3 hi"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
