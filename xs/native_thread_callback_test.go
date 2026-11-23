//go:build darwin || linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// A thread of native code's own calls a Perl sub while the native call
// that started it waits for it (Affix's callbacks from a C library's worker
// thread, t/027_thread_safety.t): the sub runs, in that call, as it does on
// perl. Before, the thread had no frame and the process died of SIGSEGV.
// The module sets the context on the thread where perl has one to set
// (MULTIPLICITY), as Affix does. What is expected is what perl 5.36 on
// Linux gives with the same module.
const threadCallbackModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <pthread.h>
#include <unistd.h>
struct job {
    SV *cb;
    int val;
    void *interp;
};
static void *run(void *arg) {
    struct job *j = (struct job *)arg;
#ifdef MULTIPLICITY
    PERL_SET_CONTEXT((PerlInterpreter *)j->interp);
#endif
    dTHX;
    dSP;
    ENTER;
    SAVETMPS;
    PUSHMARK(SP);
    XPUSHs(sv_2mortal(newSViv(j->val)));
    PUTBACK;
    call_sv(j->cb, G_DISCARD);
    FREETMPS;
    LEAVE;
    return NULL;
}
/* run_in_thread(cb, val): cb called with val from a thread of the module's
 * own, which this call waits for */
XS(XS_ThreadCb_run_in_thread) {
    dXSARGS;
    struct job j;
    pthread_t t;
    if (items != 2) croak_xs_usage(cv, "cb, val");
    j.cb = ST(0);
    j.val = (int)SvIV(ST(1));
    j.interp = (void *)PERL_GET_CONTEXT;
    if (pthread_create(&t, NULL, run, &j) != 0) croak("pthread_create failed");
    pthread_join(t, NULL);
    XSRETURN_EMPTY;
}
/* run_later(cb, val): cb called with val from a thread of the module's own
 * a fifth of a second later, which nothing waits for */
static struct job later;
static void *run_after(void *arg) {
    usleep(200000);
    return run(arg);
}
XS(XS_ThreadCb_run_later) {
    dXSARGS;
    pthread_t t;
    if (items != 2) croak_xs_usage(cv, "cb, val");
    later.cb = SvREFCNT_inc(ST(0));
    later.val = (int)SvIV(ST(1));
    later.interp = (void *)PERL_GET_CONTEXT;
    if (pthread_create(&t, NULL, run_after, &later) != 0) croak("pthread_create failed");
    pthread_detach(t);
    XSRETURN_EMPTY;
}
/* two_at_once(sv): two threads of the module's own use the interpreter,
 * the second while the first still holds it (it has used it and waits for
 * the second to be done), as an OpenMP loop's threads read Perl values */
struct two {
    SV *sv;
    void *interp;
    pthread_mutex_t mu;
    pthread_cond_t cond;
    int first_in, second_done;
};
static void *first_thread(void *arg) {
    struct two *w = (struct two *)arg;
#ifdef MULTIPLICITY
    PERL_SET_CONTEXT((PerlInterpreter *)w->interp);
#endif
    dTHX;
    (void)SvTRUE(w->sv);
    pthread_mutex_lock(&w->mu);
    w->first_in = 1;
    pthread_cond_broadcast(&w->cond);
    while (!w->second_done) pthread_cond_wait(&w->cond, &w->mu);
    pthread_mutex_unlock(&w->mu);
    return NULL;
}
static void *second_thread(void *arg) {
    struct two *w = (struct two *)arg;
    pthread_mutex_lock(&w->mu);
    while (!w->first_in) pthread_cond_wait(&w->cond, &w->mu);
    pthread_mutex_unlock(&w->mu);
    {
#ifdef MULTIPLICITY
        PERL_SET_CONTEXT((PerlInterpreter *)w->interp);
#endif
        dTHX;
        (void)SvTRUE(w->sv);
    }
    pthread_mutex_lock(&w->mu);
    w->second_done = 1;
    pthread_cond_broadcast(&w->cond);
    pthread_mutex_unlock(&w->mu);
    return NULL;
}
XS(XS_ThreadCb_two_at_once) {
    dXSARGS;
    static struct two w;
    pthread_t a, b;
    if (items != 1) croak_xs_usage(cv, "sv");
    w.sv = ST(0);
    w.interp = (void *)PERL_GET_CONTEXT;
    pthread_mutex_init(&w.mu, NULL);
    pthread_cond_init(&w.cond, NULL);
    w.first_in = w.second_done = 0;
    pthread_create(&a, NULL, first_thread, &w);
    pthread_create(&b, NULL, second_thread, &w);
    pthread_join(a, NULL);
    pthread_join(b, NULL);
    ST(0) = sv_2mortal(newSVpvs("both"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_ThreadCb) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("ThreadCb::run_in_thread", XS_ThreadCb_run_in_thread, __FILE__);
    newXS("ThreadCb::run_later", XS_ThreadCb_run_later, __FILE__);
    newXS("ThreadCb::two_at_once", XS_ThreadCb_two_at_once, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeThreadCallsPerlWhileItsCallWaits(t *testing.T) {
	so := buildModuleSource(t, "ThreadCb", threadCallbackModule)
	program := `my $got = 0;
my $str;
ThreadCb::run_in_thread(sub { $str = "Received: $_[0]"; $got = $_[0] }, 123);
my @inner;
ThreadCb::run_in_thread(sub { ThreadCb::run_in_thread(sub { push @inner, $_[0] }, $_[0] + 1) }, 7);
"got $got, $str, inner @inner"`
	if got, want := requireModule(t, "ThreadCb", so, program), "got 123, Received: 123, inner 8"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A thread of native code's own that calls Perl while no native call of
// the interpreter's waits for it (the call that started it has returned,
// and the program runs Perl code) cannot be let in: the interpreter is
// running on its own thread. The process ends with a message that says so,
// not with SIGSEGV. The process ends, so the case runs in one of its own.
func TestNativeThreadCallingPerlWithNoCallWaitingIsTold(t *testing.T) {
	if so := os.Getenv("GOPERL_THREADCB_CHILD"); so != "" {
		threadCallbackChild(t, so, `ThreadCb::run_later(sub { print "ran\n" }, 1); select undef, undef, undef, 3; "slept"`)
		return
	}
	so := buildModuleSource(t, "ThreadCb", threadCallbackModule)
	cmd := exec.Command(os.Args[0], "-test.run=^TestNativeThreadCallingPerlWithNoCallWaitingIsTold$")
	cmd.Env = append(os.Environ(), "GOPERL_THREADCB_CHILD="+so)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	const want = "goperl native XS: the interpreter was used from a thread of native code while no native call of the interpreter's was waiting: such a thread can use it only while a native call of the interpreter waits for it\n"
	if err == nil || strings.Contains(out.String(), "ran") || strings.Contains(out.String(), "result:") ||
		!strings.Contains(errOut.String(), want) {
		t.Fatalf("err=%v\nstdout %q\nstderr %q", err, out.String(), errOut.String())
	}
}

// Two threads of native code's own that use the interpreter at once (the
// second while the first holds the frame of the call that waits for them)
// would share the frame's operation box, its result slots and the SDK's
// caches, and get one another's answers (an OpenMP loop of av_fetch and
// SvNV summed wrong). Perl survives concurrent reads because they write
// nothing shared; here every use goes through that state, so the process
// ends with a message instead of a wrong answer. It ends, so the case runs
// in a process of its own.
func TestNativeThreadsUsingTheInterpreterAtOnceAreTold(t *testing.T) {
	if so := os.Getenv("GOPERL_THREADCB_CHILD"); so != "" {
		threadCallbackChild(t, so, `ThreadCb::two_at_once(1)`)
		return
	}
	so := buildModuleSource(t, "ThreadCb", threadCallbackModule)
	cmd := exec.Command(os.Args[0], "-test.run=^TestNativeThreadsUsingTheInterpreterAtOnceAreTold$")
	cmd.Env = append(os.Environ(), "GOPERL_THREADCB_CHILD="+so)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	const want = "goperl native XS: two threads used the interpreter at once (threads of native code's own can use it one at a time, while a native call of the interpreter waits for them)\n"
	if err == nil || strings.Contains(out.String(), "result:") || !strings.Contains(errOut.String(), want) {
		t.Fatalf("err=%v\nstdout %q\nstderr %q", err, out.String(), errOut.String())
	}
}

func threadCallbackChild(t *testing.T, so, program string) {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, NativeModules: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	libdir := t.TempDir()
	auto := filepath.Join(libdir, "auto", "ThreadCb")
	if err := os.MkdirAll(auto, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(so, filepath.Join(auto, "ThreadCb.so")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, "ThreadCb.pm"),
		[]byte("package ThreadCb;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := p.Eval(context.Background(), `unshift @INC, '`+libdir+`'; require ThreadCb; `+program)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("Eval died: %v", r.Error)
	}
	os.Stdout.WriteString("result: " + resultStr(r) + "\n")
}
