//go:build darwin || linux

package xs_test

import "testing"

// load_module runs the module's code on a stack of its own, as perl's does
// (the BEGIN block of the use it makes runs on a PERLSI_REQUIRE stack): the
// caller's stack is where it was afterwards, also when compiling and running
// the module needs far more stack than there is, so that a pointer native
// code took into it before is good after. Cpanel::JSON::XS's decode_bom
// takes dSP, loads Encode, and pushes through that pointer. Before, the
// module ran on the caller's stack (PERLSI_MAIN where perl has
// PERLSI_REQUIRE, 9), which moved when it grew.
const loadModuleStackSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static SV *after_load(pTHX_ SV *name) {
    dSP;
    I32 n;
    SV *r;
    ENTER;
    load_module(PERL_LOADMOD_NOIMPORT, name, NULL);
    LEAVE;
    ENTER;
    PUSHMARK(SP);
    XPUSHs(sv_2mortal(newSViv(20)));
    XPUSHs(sv_2mortal(newSViv(22)));
    PUTBACK;
    n = call_pv("MyLmDeep::add", G_SCALAR);
    SPAGAIN;
    r = n == 1 ? newSVsv(POPs) : newSVpvs("none");
    PUTBACK;
    LEAVE;
    return r;
}
/* sitype(): the type of the stack the code runs on */
XS(XS_MyLm_sitype) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(PL_curstackinfo->si_type));
    XSRETURN(1);
}
XS(XS_MyLm_load) {
    dXSARGS;
    SV *r;
    PERL_UNUSED_VAR(items);
    r = after_load(aTHX_ newSVsv(ST(0)));
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyLm) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyLm::load", XS_MyLm_load);
    newXS_deffile("MyLm::sitype", XS_MyLm_sitype);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeLoadModuleLeavesTheStack(t *testing.T) {
	so := buildModuleSource(t, "MyLm", loadModuleStackSource)
	got := requireModule(t, "MyLm", so, `my $dir = "/tmp/mylm-$$";
mkdir $dir;
open my $fh, ">", "$dir/MyLmDeep.pm" or die "$dir: $!";
print $fh q{package MyLmDeep;
sub r { my $n = shift; $n ? (1, r($n - 1)) : () }
BEGIN { my @x = r(30000) }
my @y = r(30000);
our $sitype = MyLm::sitype();
sub add { $_[0] + $_[1] }
1;
};
close $fh;
unshift @INC, $dir;
my @pad = (1) x 10;
my $r = join ",", MyLm::load("MyLmDeep"), $MyLmDeep::sitype, scalar(@pad), @pad[0 .. 2];
unlink "$dir/MyLmDeep.pm";
rmdir $dir;
$r`)
	if want := "42,9,10,1,1,1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
