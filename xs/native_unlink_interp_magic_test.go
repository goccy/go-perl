//go:build darwin || linux

package xs_test

import "testing"

// Native code takes the interpreter's own magic off a value by hand, as
// perl lets it: it unlinks the MAGIC from the chain SvMAGIC showed and
// frees it and what it points at with Safefree. The magic is gone from the
// value then, as on perl: a Hash::Util::FieldHash hash whose uvar magic
// was taken off is a plain hash (keys by reference are strings again, and
// the entry stored under the id stays). Before, gperl refused with
// "SvMAGIC: native code unlinked magic 'U' that the interpreter keeps on
// the value" (Variable-Magic t/16-huf.t).
func TestNativeCodeUnlinksTheInterpretersMagic(t *testing.T) {
	so := buildModuleSource(t, "UnlinkU", unlinkUSource)
	got := requireModule(t, "UnlinkU", so, `use Hash::Util::FieldHash qw(fieldhash);
fieldhash my %h;
my $key = [];
$h{$key} = "by ref";
my $r = eval { UnlinkU::uvar(\%h); 1 } ? "lived" : $@;
my $id = Hash::Util::FieldHash::id($key);
$h{$key} = "plain";
join "; ", $r, (exists $h{$id} ? "the entry stays under the id" : "no entry under the id"),
    (exists $h{"$key"} ? "a plain key now" : "still a field hash")`)
	if want := "lived; the entry stays under the id; a plain key now"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

const unlinkUSource = `#include "EXTERN.h"
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
    mg->mg_flags = 0;
    Safefree(mg->mg_ptr);
    Safefree(mg);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_UnlinkU) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("UnlinkU::uvar", XS_UnlinkU_uvar);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// Native code puts its own uvar magic first on a field hash the way
// Variable::Magic casts: it keeps the functions of the uvar magic it finds
// (Hash::Util::FieldHash's), takes that magic off and frees it, attaches
// its own with sv_magic, and calls the functions it kept from its own
// callback, so the hash stays a field hash; its dispell takes its magic off
// and puts the kept functions back with sv_magic, and the hash is a field
// hash of its own again. Before, the cast was refused, and the kept
// functions were the interpreter's, which native code cannot call.
func TestNativeCodeChainsTheInterpretersUvarMagic(t *testing.T) {
	so := buildModuleSource(t, "MyVm", chainUvarSource)
	got := requireModule(t, "MyVm", so, `use Hash::Util::FieldHash qw(fieldhash id);
fieldhash my %h;
my $obj = [];
MyVm::cast(\%h);
$h{$obj} = 1;
my $c1 = MyVm::calls();
my $by_id = exists $h{id($obj)} ? "stored under the id" : "stored under " . join(",", keys %h);
my $v = $h{$obj};
MyVm::dispell(\%h);
my $c2 = MyVm::calls();
$h{$obj} = 2;
my $c3 = MyVm::calls();
join "; ", $by_id, "fetched $v", ($c1 > 0 ? "called" : "not called"),
    ($c2 > $c1 ? "called again" : "not again"), ($c3 == $c2 ? "not after dispell" : "after dispell"),
    ((exists $h{id($obj)} && $h{id($obj)} == 2) ? "a field hash after" : "not a field hash after")`)
	want := "stored under the id; fetched 1; called; called again; not after dispell; a field hash after"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

const chainUvarSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static struct ufuncs kept;
static IV ncalls;
static I32 my_val(pTHX_ IV action, SV *sv) {
    ncalls++;
    if (kept.uf_val) kept.uf_val(aTHX_ action, sv);
    if (kept.uf_set) kept.uf_set(aTHX_ action, sv);
    return 0;
}
/* takes the first uvar magic off sv, and gives its functions */
static int take_uvar(pTHX_ SV *sv, struct ufuncs *was) {
    MAGIC *prev = NULL, *mg, *more = NULL;
    for (mg = SvMAGIC(sv); mg; prev = mg, mg = more) {
        more = mg->mg_moremagic;
        if (mg->mg_type == PERL_MAGIC_uvar) break;
    }
    if (!mg) return 0;
    if (was) *was = *(struct ufuncs *)mg->mg_ptr;
    if (prev) prev->mg_moremagic = more;
    else SvMAGIC_set(sv, more);
    mg->mg_flags = 0;
    Safefree(mg->mg_ptr);
    Safefree(mg);
    return 1;
}
XS(XS_MyVm_cast) {
    dXSARGS;
    SV *sv = SvRV(ST(0));
    struct ufuncs mine;
    PERL_UNUSED_VAR(items);
    Zero(&kept, 1, struct ufuncs);
    (void)take_uvar(aTHX_ sv, &kept);
    mine.uf_val = my_val;
    mine.uf_set = NULL;
    mine.uf_index = 0;
    sv_magic(sv, NULL, PERL_MAGIC_uvar, (const char *)&mine, sizeof mine);
    XSRETURN_EMPTY;
}
XS(XS_MyVm_dispell) {
    dXSARGS;
    SV *sv = SvRV(ST(0));
    PERL_UNUSED_VAR(items);
    if (!take_uvar(aTHX_ sv, NULL)) croak("no uvar magic of mine");
    if (kept.uf_val || kept.uf_set)
        sv_magic(sv, NULL, PERL_MAGIC_uvar, (const char *)&kept, sizeof kept);
    XSRETURN_EMPTY;
}
XS(XS_MyVm_calls) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(ncalls));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyVm) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyVm::cast", XS_MyVm_cast);
    newXS_deffile("MyVm::dispell", XS_MyVm_dispell);
    newXS_deffile("MyVm::calls", XS_MyVm_calls);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// Native code that frees the data of the interpreter's own magic, as SvMAGIC
// showed it (the ufuncs of a field hash's uvar magic), without unlinking
// the MAGIC, fails saying gperl cannot, and frees nothing: the field hash
// works on. (On perl the magic would go on with its data freed.) Before,
// the block was freed in the interpreter, whose magic went on with it
// freed, and the interpreter crashed at the end of Variable-Magic's
// t/16-huf.t in one run out of a few.
func TestNativeCodeFreesTheInterpretersMagicData(t *testing.T) {
	so := buildModuleSource(t, "FreeUData", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
XS(XS_FreeUData_uvar) {
    dXSARGS;
    SV *sv = SvRV(ST(0));
    MAGIC *mg;
    PERL_UNUSED_VAR(items);
    for (mg = SvMAGIC(sv); mg; mg = mg->mg_moremagic)
        if (mg->mg_type == PERL_MAGIC_uvar) break;
    if (!mg) croak("no uvar magic");
    Safefree(mg->mg_ptr);
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_FreeUData) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("FreeUData::uvar", XS_FreeUData_uvar);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	got := requireModule(t, "FreeUData", so, `use Hash::Util::FieldHash qw(fieldhash);
fieldhash my %h;
my $key = [];
$h{$key} = "by ref";
my $r = eval { FreeUData::uvar(\%h); 1 } ? "lived"
    : $@ =~ /\ASafefree: native code freed the data of magic 'U' that the interpreter keeps on the value/ ? "refused" : $@;
my @keys;
for (1 .. 50) { my $k = [$_]; $h{$k} = $_; push @keys, $k }
my $sum = 0; $sum += $h{$_} for @keys;
"$r; " . $h{$key} . "; $sum"`)
	if want := "refused; by ref; 1275"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
