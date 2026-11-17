//go:build darwin || linux

package xs_test

import "testing"

// A structure of native code's own that it keeps where a value would be,
// in a hash or an array of the interpreter's (Text::KnuthPlass keeps its
// list of breakpoints in its object's hash), is not dropped by perl when a
// store replaces it or the container is freed: perl's reference count of
// it is a word of the module's memory, which the interpreter reads outside
// its own memory, and it trapped ("wasm trap: memory access out of
// bounds"). Native code reads back what it stored, and a hash that held one
// is a hash as before to Perl code (keys, exists, each, delete, tied, ref,
// Data::Dumper, no warning). What is expected is what perl 5.36 on Linux
// gives with the same module.
func TestNativeStructuresKeptInContainers(t *testing.T) {
	so := buildModuleSource(t, "Kept", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* A structure of the module's own, which it keeps where a value would be,
 * as Text::KnuthPlass keeps its list of breakpoints in its object's hash.
 * Its second word is what perl reads as a reference count there. */
typedef struct {
    IV n;
    IV seven;
} kept_t;
static kept_t *kept_new(IV n) {
    kept_t *k;
    Newxz(k, 1, kept_t);
    k->n = n;
    k->seven = 7;
    return k;
}
/* keep(obj, n): the object's hash gets a structure under "list" (over the
 * one it had) */
XS(XS_Kept_keep) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    hv_stores((HV *)SvRV(ST(0)), "list", (SV *)kept_new(SvIV(ST(1))));
    XSRETURN_EMPTY;
}
/* value(obj): what the structure under "list" says */
XS(XS_Kept_value) {
    dXSARGS;
    kept_t *k;
    PERL_UNUSED_VAR(items);
    k = (kept_t *)*hv_fetch((HV *)SvRV(ST(0)), "list", 4, 0);
    XSRETURN_IV(k->n * 10 + k->seven);
}
/* keep_in_array(aref, n): a structure pushed, and one stored at index 3,
 * over another stored there before */
XS(XS_Kept_keep_in_array) {
    dXSARGS;
    AV *av = (AV *)SvRV(ST(0));
    IV n = SvIV(ST(1));
    PERL_UNUSED_VAR(items);
    av_push(av, (SV *)kept_new(n));
    av_store(av, 3, (SV *)kept_new(n + 1));
    av_store(av, 3, (SV *)kept_new(n + 2));
    XSRETURN_EMPTY;
}
/* array_value(aref, i): what the structure at index i says */
XS(XS_Kept_array_value) {
    dXSARGS;
    kept_t *k;
    PERL_UNUSED_VAR(items);
    k = (kept_t *)*av_fetch((AV *)SvRV(ST(0)), SvIV(ST(1)), 0);
    XSRETURN_IV(k->n * 10 + k->seven);
}
/* give_back(obj, sv): a value of the interpreter's under "list", over the
 * structure */
XS(XS_Kept_give_back) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    hv_stores((HV *)SvRV(ST(0)), "list", newSVsv(ST(1)));
    XSRETURN_EMPTY;
}
XS_EXTERNAL(boot_Kept) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS("Kept::keep", XS_Kept_keep, __FILE__);
    newXS("Kept::value", XS_Kept_value, __FILE__);
    newXS("Kept::keep_in_array", XS_Kept_keep_in_array, __FILE__);
    newXS("Kept::array_value", XS_Kept_array_value, __FILE__);
    newXS("Kept::give_back", XS_Kept_give_back, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `use warnings;
my @got;
my @warned;
local $SIG{__WARN__} = sub { push @warned, @_ };
for my $n (1 .. 3) {
    my $obj = bless {}, 'Kept';
    Kept::keep($obj, $n);
    Kept::keep($obj, $n + 1);
    push @got, Kept::value($obj);
}
my $aref = [];
Kept::keep_in_array($aref, 5);
push @got, Kept::array_value($aref, 0), Kept::array_value($aref, 3);
undef $aref;
# a hash that held a structure, and holds values of perl's again
my $obj = bless { name => "x", size => 3 }, 'Kept';
Kept::keep($obj, 1);
push @got, join("+", sort keys %$obj), exists $obj->{list} ? "exists" : "absent";
Kept::give_back($obj, "back");
my @pairs;
for my $k (sort keys %$obj) { push @pairs, "$k=$obj->{$k}" }
my $each = 0;
while (my ($k, $v) = each %$obj) { $each++ if defined $v }
push @got, join("+", @pairs), "$each by each", ref($obj), (tied(%$obj) ? "tied" : "untied");
delete $obj->{size};
require Data::Dumper;
local $Data::Dumper::Sortkeys = 1;
local $Data::Dumper::Indent = 0;
push @got, Data::Dumper::Dumper($obj);
push @got, scalar(@warned) . " warnings";
join ",", @got, "done"`
	if got, want := requireModule(t, "Kept", so, program), "27,37,47,57,77,list+name+size,exists,list=back+name=x+size=3,3 by each,Kept,untied,$VAR1 = bless( {'list' => 'back','name' => 'x'}, 'Kept' );,0 warnings,done"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
