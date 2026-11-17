//go:build darwin || linux

package xs_test

import "testing"

// Native code that takes the interpreter's own magic off a value by hand,
// unlinking the MAGIC from the chain SvMAGIC showed and freeing it and
// what it points at, fails saying gperl cannot do that, before anything is
// freed: the magic stays the interpreter's, and the value works on. That is
// what Variable::Magic does to the uvar magic of a Hash::Util::FieldHash
// hash it casts on (vmg_cast: vmg_mg_del of the other uvar magic, then
// sv_magic of its own). Before, the struct the interpreter's magic points
// at was freed in the interpreter, and the SDK's copy of the MAGIC was
// handed to free() (free(): invalid pointer, abort: Variable-Magic
// t/16-huf.t, t/35-stash.t).
func TestNativeCodeUnlinksTheInterpretersMagic(t *testing.T) {
	so := buildModuleSource(t, "UnlinkU", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_UnlinkU_uvar) {
    dXSARGS;
    SV *sv = SvRV(ST(0));
    MAGIC *prev = NULL, *mg, *more = NULL;
    PERL_UNUSED_VAR(items);
    for (mg = SvMAGIC(sv); mg; prev = mg, mg = more) {
        more = mg->mg_moremagic;
        if (mg->mg_type == PERL_MAGIC_uvar) break;
    }
    if (!mg) croak("no uvar magic");
    if (prev) prev->mg_moremagic = more;
    else SvMAGIC_set(sv, more);
    Safefree(mg->mg_ptr);
    Safefree(mg);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_UnlinkU) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("UnlinkU::uvar", XS_UnlinkU_uvar);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	got := requireModule(t, "UnlinkU", so, `use Hash::Util::FieldHash qw(fieldhash);
fieldhash my %h;
my $key = [];
$h{$key} = "by ref";
my $r = eval { UnlinkU::uvar(\%h); 1 } ? "lived"
    : $@ =~ /\ASvMAGIC: native code unlinked magic 'U' that the interpreter keeps on the value/ ? "refused" : $@;
my $id = Hash::Util::FieldHash::id($key);
"$r; " . (exists $h{$id} ? "still a field hash" : "not") . "; " . $h{$key}`)
	if want := "refused; still a field hash; by ref"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
