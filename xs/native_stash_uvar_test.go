//go:build darwin || linux

package xs_test

import "testing"

// uvar magic on a stash, as English::Name puts on %main::: the function
// sees every name looked up in it, in the MAGIC's mg_obj, and may put
// another name there, which is then looked up instead. The function looks
// a name up in the stash itself with HV_DISABLE_UVAR_XKEY, so that it is
// not called again for that lookup. Before, the flag was dropped, the
// function was called again for ever, and the interpreter's stack ran out
// in a trap (English-Name).
func TestNativeUvarMagicOnAStash(t *testing.T) {
	so := buildModuleSource(t, "StashFilter", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static I32 name_filter(pTHX_ IV action, SV *value) {
    MAGIC *magic = mg_find(value, PERL_MAGIC_uvar);
    STRLEN len;
    const char *name;
    PERL_UNUSED_VAR(action);
    if (!magic || !magic->mg_obj) return 0;
    name = SvPV(magic->mg_obj, len);
    if (len == 9 && memEQ(name, "SHORTHAND", 9))
        magic->mg_obj = newSVpvn("longhand", 8);
    /* a name of a punctuation variable, as English::Name aliases one:
     * made first when it is not there */
    if (len == 5 && memEQ(name, "\5RRNO", 5)) {
        if (!cBOOL(hv_common_key_len(PL_defstash, "\5RRNO", 5,
                                     HV_FETCH_ISEXISTS | HV_DISABLE_UVAR_XKEY, NULL, 0)))
            gv_fetchpvn_flags("!", 1, GV_ADD, SVt_PV);
        magic->mg_obj = newSVpvn("!", 1);
    }
    return 0;
}
static const struct ufuncs filter = { name_filter, NULL, 0 };
XS_EXTERNAL(boot_StashFilter) {
    dXSBOOTARGSXSAPIVERCHK;
    sv_magic((SV *)PL_defstash, NULL, PERL_MAGIC_uvar, (const char *)&filter, sizeof filter);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	if got, want := requireModule(t, "StashFilter", so, `our $longhand = "long"; no strict "refs"; my $e = do { local $! = 42; 0 + ${"\cERRNO"} }; ${"main::SHORTHAND"} . " " . ${"main::longhand"} . " $e"`), "long long 42"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
