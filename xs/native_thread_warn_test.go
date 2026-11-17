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

// A thread of native code's own, which runs no part of the interpreter,
// that warns (GLib logs from its worker thread through Glib's log handler,
// which calls warn) has the message written to the process's standard
// error, and the program goes on. Before, the SDK entered the interpreter
// without a frame and the process died of SIGSEGV (Glib t/9.t). The message
// leaves the process past the interpreter's STDERR, so the case runs in a
// process of its own.
const threadWarnModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <pthread.h>
static void *warn_from_thread(void *arg) {
    PERL_UNUSED_ARG(arg);
    warn("%s-%s", "a warning", "from a thread of native code");
    return NULL;
}
XS(XS_ThreadWarn_run) {
    dXSARGS;
    pthread_t t;
    PERL_UNUSED_VAR(items);
    if (pthread_create(&t, NULL, warn_from_thread, NULL) != 0) croak("pthread_create failed");
    pthread_join(t, NULL);
    ST(0) = sv_2mortal(newSVpvs("joined"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_ThreadWarn) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("ThreadWarn::run", XS_ThreadWarn_run);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeThreadWarns(t *testing.T) {
	if so := os.Getenv("GOPERL_THREADWARN_CHILD"); so != "" {
		threadWarnChild(t, so)
		return
	}
	so := buildModuleSource(t, "ThreadWarn", threadWarnModule)
	cmd := exec.Command(os.Args[0], "-test.run=^TestNativeThreadWarns$")
	cmd.Env = append(os.Environ(), "GOPERL_THREADWARN_CHILD="+so)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	if err != nil || !strings.Contains(out.String(), "result: joined; after\n") ||
		!strings.Contains(errOut.String(), "a warning-from a thread of native code\n") {
		t.Fatalf("err=%v\nstdout %q\nstderr %q", err, out.String(), errOut.String())
	}
}

func threadWarnChild(t *testing.T, so string) {
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
	auto := filepath.Join(libdir, "auto", "ThreadWarn")
	if err := os.MkdirAll(auto, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(so, filepath.Join(auto, "ThreadWarn.so")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, "ThreadWarn.pm"),
		[]byte("package ThreadWarn;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := p.Eval(context.Background(), `unshift @INC, '`+libdir+`'; require ThreadWarn; ThreadWarn::run() . "; after"`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("Eval died: %v", r.Error)
	}
	os.Stdout.WriteString("result: " + resultStr(r) + "\n")
}
