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

const sigPipeModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <signal.h>
#include <unistd.h>
static const char *sigpipe_action_name(void) {
    struct sigaction old;
    if (sigaction(SIGPIPE, NULL, &old) != 0) return "error";
    return old.sa_handler == SIG_DFL ? "default" : old.sa_handler == SIG_IGN ? "ignored" : "handled";
}
/* what a network library does when it is loaded: ignore SIGPIPE when it
 * finds the default action, then write into a pipe nobody reads */
XS(XS_SigPipe_ignore_and_write) {
    dXSARGS;
    const char *seen = sigpipe_action_name();
    int fds[2];
    ssize_t n;
    int e;
    PERL_UNUSED_VAR(items);
    if (strEQ(seen, "default")) signal(SIGPIPE, SIG_IGN);
    if (pipe(fds) != 0) croak("pipe: %s", strerror(errno));
    close(fds[0]);
    n = write(fds[1], "x", 1);
    e = errno;
    close(fds[1]);
    ST(0) = sv_2mortal(newSVpvf("%s, then %s, write %d %s", seen, sigpipe_action_name(), (int)n,
                                n < 0 && e == EPIPE ? "EPIPE" : n < 0 ? strerror(e) : "written"));
    XSRETURN(1);
}
XS_EXTERNAL(boot_SigPipe) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("SigPipe::ignore_and_write", XS_SigPipe_ignore_and_write);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// Native code of a program that is the host process finds SIGPIPE as the
// program has it: at its default action in a fresh process, which a network
// library then ignores, so that a write into a pipe nobody reads fails with
// EPIPE and no signal ends the program (EV::ClickHouse); ignored when the
// program ignores it; handled when the program has a handler. Before,
// native code found the handler of the process's runtime whatever the
// program had, never ignored the signal, and the write ended the program.
// Each case runs in a process of its own, whose signal actions nothing else
// has changed.
func TestNativeFindsTheProgramsActionOfSIGPIPE(t *testing.T) {
	if os.Getenv("GOPERL_SIGPIPE_CHILD") != "" {
		sigPipeChild(t)
		return
	}
	so := buildModuleSource(t, "SigPipe", sigPipeModule)
	for _, tc := range []struct{ name, program, want string }{
		{"the default action", ``, "default, then ignored, write -1 EPIPE; alive\n"},
		{"ignored by the program", `$SIG{PIPE} = "IGNORE";`, "ignored, then ignored, write -1 EPIPE; alive\n"},
		{"a handler of the program", `my $got = 0; $SIG{PIPE} = sub { $got++ };`, "handled, then handled, write -1 EPIPE; alive\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestNativeFindsTheProgramsActionOfSIGPIPE$")
			cmd.Env = append(os.Environ(), "GOPERL_SIGPIPE_CHILD="+so, "GOPERL_SIGPIPE_PROGRAM="+tc.program)
			var out, errOut bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &errOut
			err := cmd.Run()
			got := out.String()
			if i := strings.Index(got, "result: "); i >= 0 {
				got = got[i+len("result: "):]
				if j := strings.Index(got, "\n"); j >= 0 {
					got = got[:j+1]
				}
			}
			if err != nil || got != tc.want {
				t.Fatalf("err=%v got %q, want %q\nstdout %q\nstderr %q", err, got, tc.want, out.String(), errOut.String())
			}
		})
	}
}

// sigPipeChild is the process of one case: the program that is the host
// process loads the module and calls it.
func sigPipeChild(t *testing.T) {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var out, errOut bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, HostProcess: true,
		NativeModules: true, Stdout: &out, Stderr: &errOut})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	loadXSModule(t, p, "SigPipe", os.Getenv("GOPERL_SIGPIPE_CHILD"))
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "SigPipe.pm"),
		[]byte("package SigPipe;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := p.RunProgram(context.Background(), perl.Program{Inc: []string{libdir}, Code: []string{
		`use SigPipe; $| = 1;`,
		os.Getenv("GOPERL_SIGPIPE_PROGRAM"),
		`print SigPipe::ignore_and_write(), "; alive\n";`,
	}})
	if err != nil || status != 0 {
		t.Fatalf("status=%d err=%v stdout %q stderr %q", status, err, out.String(), errOut.String())
	}
	os.Stdout.WriteString("result: " + out.String())
}
