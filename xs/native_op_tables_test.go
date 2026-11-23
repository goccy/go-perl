//go:build darwin || linux

package xs_test

import "testing"

// The entries of PL_check and PL_ppaddr are one for the op types whose
// function in the interpreter is one (ck_null for OP_NULL and OP_STUB,
// pp_null for OP_NULL and OP_SCALAR), as Opcodes compares them, and an
// entry a module replaces leaves the others of the same function as they
// were. Before, every type had a slot of its own, so check(0) and check(1)
// differed. The expected string is what perl prints with the same module
// code.
const opTablesSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

static OP *MyTab_ck(pTHX_ OP *o) { return o; }
XS(XS_MyTab_tables) {
    dXSARGS;
    Perl_check_t was;
    const char *kept;
    PERL_UNUSED_VAR(items);
    was = PL_check[OP_STUB];
    PL_check[OP_STUB] = MyTab_ck;
    kept = PL_check[OP_NULL] == was ? "kept" : "changed";
    PL_check[OP_STUB] = was;
    ST(0) = sv_2mortal(newSVpvf("check null/stub %s, null/entersub %s; ppaddr null/scalar %s, "
                                "null/stub %s; %s",
                                PL_check[OP_NULL] == PL_check[OP_STUB] ? "same" : "other",
                                PL_check[OP_NULL] == PL_check[OP_ENTERSUB] ? "same" : "other",
                                PL_ppaddr[OP_NULL] == PL_ppaddr[OP_SCALAR] ? "same" : "other",
                                PL_ppaddr[OP_NULL] == PL_ppaddr[OP_STUB] ? "same" : "other", kept));
    XSRETURN(1);
}

XS_EXTERNAL(boot_MyTab) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyTab::tables", XS_MyTab_tables);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeEntriesOfTheOpTables(t *testing.T) {
	so := buildModuleSource(t, "MyTab", opTablesSource)
	got := requireModule(t, "MyTab", so, `MyTab::tables()`)
	want := "check null/stub same, null/entersub other; ppaddr null/scalar same, null/stub other; kept"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
