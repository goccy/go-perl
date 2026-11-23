//go:build darwin || linux

package gperl_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// initThreadModule calls a Perl sub from a thread of its own while the
// native call that started the thread waits for it, as Affix's callbacks do
// (t/027_thread_safety.t).
const initThreadModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <pthread.h>
struct job {
    SV *cb;
};
static void *run(void *arg) {
    struct job *j = (struct job *)arg;
    dTHX;
    dSP;
    PUSHMARK(SP);
    PUTBACK;
    call_sv(j->cb, G_DISCARD);
    return NULL;
}
XS(XS_InitThread_run_in_thread) {
    dXSARGS;
    struct job j;
    pthread_t t;
    if (items != 1) croak_xs_usage(cv, "cb");
    j.cb = ST(0);
    pthread_create(&t, NULL, run, &j);
    pthread_join(t, NULL);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_InitThread) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("InitThread::run_in_thread", XS_InitThread_run_in_thread, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// A program started as the perl command (gperl's perl link, or the shim a
// build or test harness runs) runs inside the Go program's package
// initialization, where Go holds a call into Go from a thread it did not
// start until the initialization ends: a native module's thread that calls
// Perl there would wait for ever, and the native call waiting for it with
// it. The process ends with a message that says so instead. (Under `gperl
// run` the program runs after initialization, and the callback runs:
// TestNativeThreadCallsPerlWhileItsCallWaits.)
func TestNativeThreadCallingPerlInsideInitializationIsTold(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	sdk, err := filepath.Abs(filepath.Join("..", "xs", "sdk", "include"))
	if err != nil {
		t.Fatal(err)
	}
	lib := t.TempDir()
	auto := filepath.Join(lib, "auto", "InitThread")
	if err := os.MkdirAll(auto, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), "InitThread.c")
	if err := os.WriteFile(src, []byte(initThreadModule), 0o644); err != nil {
		t.Fatal(err)
	}
	args := []string{"-shared", "-fPIC", "-std=gnu99", "-I", sdk, "-o", filepath.Join(auto, "InitThread.so"), src}
	if runtime.GOOS == "darwin" {
		args = append(args, "-undefined", "dynamic_lookup")
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	} else {
		args = append([]string{"-D_GNU_SOURCE", "-pthread"}, args...)
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile InitThread.so: %v\n%s", err, out)
	}
	if err := os.WriteFile(filepath.Join(lib, "InitThread.pm"),
		[]byte("package InitThread;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// the re-entry of the shim (reexec.go): the program runs inside the
	// test binary's package initialization
	cmd := exec.CommandContext(ctx, os.Args[0], "run", "-I", lib, "-MInitThread", "-e",
		`InitThread::run_in_thread(sub { print "called back\n" }); print "returned\n";`)
	cmd.Env = append(os.Environ(), "GPERL_INTERNAL_PERL_CLI=1")
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err = cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("the program waited for ever\nstdout %q\nstderr %q", out.String(), errOut.String())
	}
	const want = "goperl native XS: a thread of native code's own called into Perl while gperl runs the " +
		"program inside the Go program's initialization (a program started as the perl command: " +
		"by gperl's perl link, or as a child of a build or test harness), where Go holds such a " +
		"call until the program ends; run the program with `gperl run`\n"
	if err == nil || strings.Contains(out.String(), "returned") || !strings.Contains(errOut.String(), want) {
		t.Fatalf("err=%v\nstdout %q\nstderr %q", err, out.String(), errOut.String())
	}
}
