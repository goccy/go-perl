//go:build linux

package xs_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// An alarm goes to the handler native code made the process's action for
// SIGALRM, as an event loop's signal watcher is (libev's, under AnyEvent's
// EV backend), and the program goes on: the runtime's alarm went to the
// interpreter alone, whose default action for SIGALRM ended the program
// (ONE's test of its SIGALRM listener died by the signal). In a program
// that is the host process, and on Linux, where the runtime can tell the
// handler of native code from its own.
// What is expected is what perl 5.36 on Linux gives with the same module.
func TestNativeAlarmGoesToTheHandlerOfNativeCode(t *testing.T) {
	so := buildModuleSource(t, "Alrm", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <signal.h>
#include <string.h>
/* An event loop's signal watcher, as libev keeps one: a handler of the
 * module's own made the process's action for SIGALRM, which counts. */
static volatile sig_atomic_t alarms;
static struct sigaction before;
static void on_alarm(int sig) {
    (void)sig;
    alarms++;
}
/* watch(): the handler is the process's action for SIGALRM */
XS(XS_Alrm_watch) {
    dXSARGS;
    struct sigaction sa;
    PERL_UNUSED_VAR(items);
    memset(&sa, 0, sizeof sa);
    sa.sa_handler = on_alarm;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = SA_RESTART | SA_ONSTACK;
    ST(0) = sigaction(SIGALRM, &sa, &before) == 0 ? &PL_sv_yes : &PL_sv_no;
    XSRETURN(1);
}
/* unwatch(): the action before watch() again */
XS(XS_Alrm_unwatch) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    sigaction(SIGALRM, &before, NULL);
    XSRETURN_EMPTY;
}
/* count(): how many SIGALRMs the handler has had */
XS(XS_Alrm_count) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv((IV)alarms));
    XSRETURN(1);
}
XS_EXTERNAL(boot_Alrm) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("Alrm::watch", XS_Alrm_watch, __FILE__);
    newXS("Alrm::unwatch", XS_Alrm_unwatch, __FILE__);
    newXS("Alrm::count", XS_Alrm_count, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my $w = Alrm::watch() ? "watching" : "not watching";
alarm(1);
my $t = time;
select undef, undef, undef, 0.1 while time - $t < 3 && Alrm::count() == 0;
alarm(0);
Alrm::unwatch();
join ",", $w, "alarms " . Alrm::count(), "alive"`
	if got, want := requireModuleWith(t, perl.Config{HostProcess: true}, "Alrm", so, program), "watching,alarms 1,alive"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
