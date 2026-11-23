//go:build darwin || linux

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

// sv_dump's text caught the way Data::Peek's DDump does it without PerlIO:
// native code points descriptor 2 at a pipe for the dump, puts it back and
// reads the pipe to its end, which comes once both ends it wrote through
// are closed. What is expected is what perl 5.36 on Linux gives with the
// same module.
// The program is the host process, with the host's descriptor 2 as its
// standard error, as gperl runs one.
const dumpPipeModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <unistd.h>
#include <stdio.h>
/* dump($sv): sv_dump's text, caught the way Data::Peek's DDump does it
 * without PerlIO: fd 2 pointed at a pipe for the dump, put back, and the
 * pipe read to its end */
XS(XS_DumpPipe_dump) {
    dXSARGS;
    int err[3], n;
    char buf[128];
    SV *out = newSVpvs("");
    PERL_UNUSED_VAR(items);
    if (pipe(err)) croak("pipe: %s", strerror(errno));
    fflush(stderr);
    err[2] = dup(fileno(stderr));
    close(fileno(stderr));
    dup2(err[1], fileno(stderr));
    sv_dump(ST(0));
    fflush(stderr);
    close(err[1]);
    close(fileno(stderr));
    dup2(err[2], fileno(stderr));
    close(err[2]);
    while ((n = read(err[0], buf, sizeof buf)) > 0) sv_catpvn(out, buf, n);
    close(err[0]);
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
XS_EXTERNAL(boot_DumpPipe) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("DumpPipe::dump", XS_DumpPipe_dump, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeCatchesSvDumpThroughAPipe(t *testing.T) {
	so := buildModuleSource(t, "DumpPipe", dumpPipeModule)
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var out bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, HostProcess: true,
		NativeModules: true, Stdout: &out, Stderr: os.Stderr})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	loadXSModule(t, p, "DumpPipe", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "DumpPipe.pm"),
		[]byte("package DumpPipe;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := p.RunProgram(context.Background(), perl.Program{Inc: []string{libdir}, Code: []string{
		`use DumpPipe;`,
		`my $d = DumpPipe::dump(42);`,
		`my $e = DumpPipe::dump("x");`,
		`my $r = join "|", ($d =~ /^SV = IV\(0x[0-9a-f]+\) at 0x[0-9a-f]+\n/ ? "IV header" : "got [$d]"),`,
		`  ($d =~ /^  IV = 42$/m ? "value" : "no value"),`,
		`  ($e =~ /^  PV = 0x[0-9a-f]+ "x"\\0$/m ? "PV" : "got [$e]");`,
		`print STDOUT $r, "\n";`,
	}})
	p.Close()
	if err != nil {
		t.Fatalf("RunProgram: %v", err)
	}
	if want := "IV header|value|PV\n"; status != 0 || out.String() != want {
		t.Fatalf("status %d, got %q, want %q", status, out.String(), want)
	}
}
