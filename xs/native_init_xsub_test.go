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

// A module the program loads while it runs, which names an XSUB INIT with
// the warnings of PL_curcop turned off around newXS, as Tk's boot does,
// says no warning; with them all turned on it says "Too late to run INIT
// block", as perl does for an INIT made after the main program started.
func TestNativeInitXSUBWithWarningsOff(t *testing.T) {
	so := buildModuleSource(t, "InitXS", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_InitXS_INIT) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_InitXS) {
    dXSBOOTARGSXSAPIVERCHK;
    {
        STRLEN *old_warn = (STRLEN *)PL_curcop->cop_warnings;
        PL_curcop->cop_warnings = pWARN_ALL;
        newXS("InitXS::Loud::INIT", XS_InitXS_INIT, __FILE__);
        PL_curcop->cop_warnings = pWARN_NONE;
        newXS("InitXS::INIT", XS_InitXS_INIT, __FILE__);
        PL_curcop->cop_warnings = (char *)old_warn;
    }
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var out, errOut bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, NativeModules: true,
		Stdout: &out, Stderr: &errOut})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	loadXSModule(t, p, "InitXS", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "InitXS.pm"),
		[]byte("package InitXS;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := p.RunProgram(context.Background(), perl.Program{Inc: []string{libdir}, Code: []string{
		`my @w; $SIG{__WARN__} = sub { push @w, @_ };`,
		`require InitXS;`,
		`print join("|", map { /^(Too late to run INIT block) at / ? $1 : $_ } @w), "\n";`,
	}})
	if err != nil || status != 0 {
		t.Fatalf("status=%d err=%v stderr %q", status, err, errOut.String())
	}
	if got, want := out.String(), "Too late to run INIT block\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
