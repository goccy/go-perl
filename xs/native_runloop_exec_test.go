//go:build darwin || linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// A run loop of native code installed when the module is loaded, as
// Devel::Cover and Test::LeakTrace install theirs.
const runopsExecSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static int my_runops(pTHX) {
    while ((PL_op = PL_op->op_ppaddr(aTHX))) {
        PERL_ASYNC_CHECK();
    }
    TAINT_NOT;
    return 0;
}
XS_EXTERNAL(boot_RunopsExec) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    PL_runops = my_runops;
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// runRunopsExec runs code as a program that may exec, with RunopsExec to
// `use`, and answers its status and what it printed.
func runRunopsExec(t *testing.T, so string, code ...string) (int, string) {
	t.Helper()
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var out bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, NativeModules: true,
		Exec: func(string, []string) bool { return true }, Stdout: &out})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	loadXSModule(t, p, "RunopsExec", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "RunopsExec.pm"),
		[]byte("package RunopsExec;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := p.RunProgram(context.Background(), perl.Program{Code: code, Inc: []string{libdir}})
	if err != nil {
		t.Fatalf("RunProgram: %v", err)
	}
	return status, out.String()
}

// A program that execs while a run loop of native code runs it (Devel::
// Cover's t/e2e/aexec.t) ends with the status of the command it execs, as
// perl's does, and runs no END block: exec, from the program's top or from
// an eval in a sub. Before, the exit at the end of the exec came back into
// the run loop as a failed operation: the loop died with $! = EBADF (every
// descriptor was closed by then) and the program ended with 9. What is
// expected is what perl 5.36 on Linux gives with the same module.
func TestNativeRunLoopThroughAnExec(t *testing.T) {
	so := buildModuleSource(t, "RunopsExec", runopsExecSource)
	status, out := runRunopsExec(t, so, `use RunopsExec; $| = 1; print "before\n"; END { print "END ran\n" }`,
		`exec "sh", "-c", "exit 3"; print "not reached\n"`)
	if status != 3 || out != "before\n" {
		t.Errorf("exec from the top: status %d, output %q; want 3, %q", status, out, "before\n")
	}
	status, out = runRunopsExec(t, so, `use RunopsExec; sub f { my $x = eval { exec "sh", "-c", "exit 4"; 1 }; print "not reached\n" } f()`)
	if status != 4 || out != "" {
		t.Errorf("exec from an eval in a sub: status %d, output %q; want 4, %q", status, out, "")
	}
}
