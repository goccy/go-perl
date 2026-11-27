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

// Under the debugger (-d), native code that calls a sub with call_sv and
// no G_EVAL enters DB::sub for that sub, as on perl, and for nothing else:
// the eval gperl makes around such a call is no sub of the program's. A
// closure made not debuggable (CvNODEBUG, as Sub::Metadata makes
// Devel::DebugHooks' cleanups) enters no DB::sub at all. Before, DB::sub
// was entered for the eval's sub too, at every call: Devel::DebugHooks,
// whose DB::sub establishes a cleanup that Scope::Cleanup runs with such a
// call, recursed until the interpreter's stack ran out.
const plainCallModuleSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_MyPlainCall_call) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "code");
    PUSHMARK(SP);
    PUTBACK;
    call_sv(ST(0), G_VOID | G_DISCARD);
    XSRETURN_EMPTY;
}
XS(XS_MyPlainCall_nodebug) {
    dXSARGS;
    if (items != 1 || !SvROK(ST(0)) || SvTYPE(SvRV(ST(0))) != SVt_PVCV)
        croak_xs_usage(cv, "code");
    CvNODEBUG_on((CV *)SvRV(ST(0)));
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_MyPlainCall) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyPlainCall::call", XS_MyPlainCall_call);
    newXS_deffile("MyPlainCall::nodebug", XS_MyPlainCall_nodebug);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativePlainCallUnderTheDebugger(t *testing.T) {
	so := buildModuleSource(t, "MyPlainCall", plainCallModuleSource)
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	libdir := t.TempDir()
	files := map[string]string{
		"MyPlainCall.pm": "package MyPlainCall;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n",
		// the debugger: DB::sub notes the subs it is entered for while
		// $DB::on is set, by name or, for a sub without one, as "anon"
		"MyDB.pm": "package DB;\nour (@seen, $on);\nsub DB { }\n" +
			"sub sub { push @seen, ref $DB::sub ? 'anon' : $DB::sub if $on; no strict 'refs'; &$DB::sub }\n1;\n",
	}
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(libdir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out, errOut bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, NativeModules: true,
		Stdout: &out, Stderr: &errOut, Env: []string{"PERL5DB=BEGIN { require MyDB }"}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	loadXSModule(t, p, "MyPlainCall", so)
	status, err := p.RunProgram(context.Background(), perl.Program{Inc: []string{libdir}, Debug: true, Code: []string{
		`use MyPlainCall;`,
		`sub named { 1 }`,
		`my $x = 2;`,
		`my $anon = sub { $x };`,
		`my $quiet = sub { $x + 1 };`,
		`MyPlainCall::nodebug($quiet);`,
		`$DB::on = 1;`,
		`MyPlainCall::call(\&named);`,
		`MyPlainCall::call($anon);`,
		`MyPlainCall::call($quiet);`,
		`$DB::on = 0;`,
		`print join(",", @DB::seen), "\n";`,
	}})
	if err != nil {
		t.Fatalf("RunProgram: %v (stderr %q)", err, errOut.String())
	}
	want := "MyPlainCall::call,main::named,MyPlainCall::call,anon,MyPlainCall::call\n"
	if status != 0 || out.String() != want {
		t.Fatalf("status %d, got %q, want %q; stderr %q", status, out.String(), want, errOut.String())
	}
}
