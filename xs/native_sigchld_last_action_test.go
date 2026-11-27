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

// The program's own $SIG{CHLD}, set after native code made a handler of its
// own the action of SIGCHLD, is the process's action from then on, as in
// perl, where the last sigaction wins: the native handler no longer runs,
// and the end of a fork child is the program's to reap. libev makes such a
// handler when EV loads, and its loop reaps every child with waitpid(-1);
// IO::Async sets $SIG{CHLD} for watch_process after that. Before, the
// native handler stayed the process's action, ran for the fork child's end,
// and the loop took the child from the program: IO::Async::Loop::Mojo's
// t/05loop-process.t under Mojo's EV reactor never saw its child end.
const sigchldLastActionModule = `#include "EXTERN.h"
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
XS(XS_MySc_install) {
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
XS(XS_MySc_poll) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    if (chld_signals) {
        int status;
        while (waitpid(-1, &status, WNOHANG) > 0) chld_reaped++;
    }
    XSRETURN_EMPTY;
}
XS(XS_MySc_report) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("native handler %d, native reaped %d", (int)chld_signals, chld_reaped));
    XSRETURN(1);
}
XS(XS_MySc_uninstall) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    signal(SIGCHLD, SIG_DFL);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MySc) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySc::install", XS_MySc_install);
    newXS_deffile("MySc::poll", XS_MySc_poll);
    newXS_deffile("MySc::report", XS_MySc_report);
    newXS_deffile("MySc::uninstall", XS_MySc_uninstall);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeSigchldActionGivesWayToTheProgramsOwn(t *testing.T) {
	so := buildModuleSource(t, "MySc", sigchldLastActionModule)
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
	loadXSModule(t, p, "MySc", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "MySc.pm"),
		[]byte("package MySc;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := p.RunProgram(context.Background(), perl.Program{Inc: []string{libdir}, Code: []string{
		`use strict; use warnings; use MySc; use POSIX ();`,
		`MySc::install();`,
		`my $handled = 0;`,
		`$SIG{CHLD} = sub { $handled++ };`,
		`my $pid = fork // die "fork: $!";`,
		`POSIX::_exit(3) if !$pid;`,
		`for (1 .. 500) { MySc::poll(); last if $handled; select undef, undef, undef, 0.01 }`,
		`select undef, undef, undef, 0.2; MySc::poll();`,
		`my $got = waitpid($pid, 0);`,
		`print "handled $handled, ", MySc::report(), ", waitpid ", ($got == $pid ? "the child" : $got), ", status ", $? >> 8, "\n";`,
		`MySc::uninstall();`,
	}})
	p.Close()
	if err != nil {
		t.Fatalf("RunProgram: %v", err)
	}
	// perl 5.36 prints the same
	if want := "handled 1, native handler 0, native reaped 0, waitpid the child, status 3\n"; status != 0 || out.String() != want {
		t.Fatalf("status %d, got %q, want %q, stderr %q", status, out.String(), want, errOut.String())
	}
}
