//go:build darwin || linux

package xs_test

import (
	"os"
	"path/filepath"
	"testing"
)

// What native code writes with stdio and leaves in the C library's buffer
// (printf to stdout, fprintf to a stream it opened) is written out when the
// program ends, as perl's exit does. Before, a gperl program ended without
// the C library's exit and the buffered output was lost (Clownfish-CFC and
// Inline-Filters print their TAP with printf).
func TestNativeStdioFlushedAtTheEnd(t *testing.T) {
	so := buildModuleSource(t, "StdioLeft", `#include <stdio.h>
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_StdioLeft_note) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "path");
    {
        FILE *fp = fopen(SvPV_nolen(ST(0)), "w");
        if (!fp) croak("fopen failed");
        fprintf(fp, "native line\n");
        XSRETURN_EMPTY;
    }
}
XS_EXTERNAL(boot_StdioLeft) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("StdioLeft::note", XS_StdioLeft_note);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	path := filepath.Join(t.TempDir(), "out.txt")
	got := requireModule(t, "StdioLeft", so, `StdioLeft::note("`+path+`"); -s "`+path+`" || 0`)
	if got != "0" {
		t.Fatalf("the stream was written out while the program ran (%s bytes): the test needs a buffered stream", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "native line\n" {
		t.Errorf("after the end: got %q, want %q", data, "native line\n")
	}
}
