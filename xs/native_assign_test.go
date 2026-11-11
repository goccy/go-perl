//go:build darwin || linux

package xs_test

import (
	"context"
	"testing"
)

// The function of an XSUB and two fields of a glob are lvalues of the perl
// API: `CvXSUB(cv) = fn` makes the sub call another native function, and
// `GvCVGEN(gv) = n` and `GvSTASH(gv) = stash` change the glob.
func TestNativeAccessorAssignments(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"CvXSUB", `join ",", PerlAPI::impl(), (PerlAPI::exchange() ? "second" : "first"), PerlAPI::impl(), (PerlAPI::exchange() ? "second" : "first"), PerlAPI::impl()`, "one,second,two,first,one"},
		{"GvCVGEN and GvSTASH", `sub Some::Pkg::f { 1 } join ",", PerlAPI::glob_fields("Some::Pkg::f", "Other::Pkg"), *Some::Pkg::f{PACKAGE}`, "7,0,Other::Pkg,Other::Pkg"},
	})
}

// A warning of a category is given the way the interpreter gives its own:
// Perl_warner always, Perl_ck_warner where the category is enabled,
// Perl_ck_warner_d unless it is disabled, and as an exception where the
// caller made the category fatal.
func TestNativeCategoryWarnings(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `our @seen; $SIG{__WARN__} = sub { push @seen, $_[0] }; sub seen { my $s = join "", @seen; @seen = (); $s eq "" ? "nothing" : $s } 1`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"ck_warner with the category enabled", `{ use warnings; PerlAPI::category_warning(1, "enabled") } seen()`, "enabled\n"},
		{"ck_warner with the category off", `{ no warnings; PerlAPI::category_warning(1, "off") } seen()`, "nothing"},
		{"ck_warner_d by default", `{ PerlAPI::category_warning(2, "default") } seen()`, "default\n"},
		{"ck_warner_d with the category off", `{ no warnings "utf8"; PerlAPI::category_warning(2, "off") } seen()`, "nothing"},
		{"warner regardless", `{ no warnings; PerlAPI::category_warning(0, "always") } seen()`, "always\n"},
		{"a fatal category", `my $went_on = 0; { use warnings FATAL => "utf8"; eval { PerlAPI::category_warning(1, "fatal"); $went_on = 1 } } "$@$went_on:" . seen()`, "fatal\n0:nothing"},
	})
}

// A macro that stands in for a function of the perl API evaluates each
// argument once, as a call of the function would: an argument with a side
// effect (`isALPHA(*p++)`, `Renew(*slot(), n, char)`) has it once.
func TestNativeMacrosEvaluateEachArgumentOnce(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"none evaluates an argument twice", `my $s = "text" . "1"; "[" . PerlAPI::evaluated_twice($s) . "]"`, "[]"},
	})
}
