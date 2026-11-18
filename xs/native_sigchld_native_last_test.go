//go:build linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// A handler of native code made the action of SIGCHLD after the program set
// its own $SIG{CHLD} has the signal, as in perl, where the last sigaction
// wins: the end of a fork child reaches the native handler (whose loop then
// reaps the child with waitpid(-1), so the program's waitpid finds none),
// and the program's handler no longer runs. Before, both ran. The other
// order, the program's handler set after the native one, is
// TestNativeSigchldActionGivesWayToTheProgramsOwn. What is expected is what
// perl 5.36 on Linux prints with the same module.
const sigchldNativeLastModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <signal.h>
#include <sys/wait.h>
#include <string.h>
static volatile sig_atomic_t chld_signals;
static int chld_reaped;
static void on_chld(int sig) { (void)sig; chld_signals++; }
/* install(): SIGCHLD's action is a handler of the module's (as libev's
 * default loop makes one for its child watchers) */
XS(XS_RevSc_install) {
    dXSARGS;
    struct sigaction sa;
    PERL_UNUSED_VAR(items);
    memset(&sa, 0, sizeof sa);
    sa.sa_handler = on_chld;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = SA_RESTART;
    sigaction(SIGCHLD, &sa, NULL);
    XSRETURN_EMPTY;
}
/* poll(): the loop's turn after the handler ran: every child that ended is
 * reaped, as libev's child callback does */
XS(XS_RevSc_poll) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    if (chld_signals) {
        int status;
        while (waitpid(-1, &status, WNOHANG) > 0) chld_reaped++;
    }
    XSRETURN_EMPTY;
}
XS(XS_RevSc_report) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("native handler %d, native reaped %d", (int)chld_signals, chld_reaped));
    XSRETURN(1);
}
XS(XS_RevSc_uninstall) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    signal(SIGCHLD, SIG_DFL);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_RevSc) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("RevSc::install", XS_RevSc_install, __FILE__);
    newXS("RevSc::poll", XS_RevSc_poll, __FILE__);
    newXS("RevSc::report", XS_RevSc_report, __FILE__);
    newXS("RevSc::uninstall", XS_RevSc_uninstall, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeSigchldActionSetLastHasTheSignal(t *testing.T) {
	so := buildModuleSource(t, "RevSc", sigchldNativeLastModule)
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var out, errOut bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, Fork: true, HostProcess: true,
		NativeModules: true, Stdout: &out, Stderr: &errOut})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	loadXSModule(t, p, "RevSc", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "RevSc.pm"),
		[]byte("package RevSc;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := p.RunProgram(context.Background(), perl.Program{Inc: []string{libdir}, Code: []string{
		`use RevSc;`,
		`require POSIX;`,
		`my $handled = 0;`,
		`$SIG{CHLD} = sub { $handled++ };`,
		`RevSc::install();`,
		`my $pid = fork // die "fork: $!";`,
		`POSIX::_exit(3) if !$pid;`,
		`for (1 .. 500) { RevSc::poll(); last if $handled || RevSc::report() =~ /reaped [1-9]/; select undef, undef, undef, 0.01 }`,
		`select undef, undef, undef, 0.2; RevSc::poll();`,
		`my $got = waitpid($pid, 0);`,
		`my $r = "handled $handled, " . RevSc::report() . ", waitpid " . ($got == $pid ? "the child" : $got);`,
		`RevSc::uninstall();`,
		`print $r, "\n";`,
	}})
	p.Close()
	if err != nil {
		t.Fatalf("RunProgram: %v", err)
	}
	if want := "handled 0, native handler 1, native reaped 1, waitpid -1\n"; status != 0 || out.String() != want {
		t.Fatalf("status %d, got %q, want %q, stderr %q", status, out.String(), want, errOut.String())
	}
}
