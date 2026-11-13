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

// The referent of a reference and the variables of a glob are lvalues of the
// perl API: `SvRV(rv) = sv` (also where the right side comes back to the
// same line, as a recursive copy does), `GvSV(gv) = sv`, `GvAV(gv) = av`,
// `GvHV(gv) = hv`. An address taken of such a field stays good while the
// code that took it goes on, and reading a field of many values costs no
// more for each than for the first.
func TestNativeFieldAssignments(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"SvRV", `my ($a, $b) = (1, 2); my $r = \$a; my $got = PerlAPI::repoint($r, \$b); "$$r,$$got," . ($r == \$b ? "same" : "other")`, "2,2,same"},
		{"SvRV in a recursive copy", `my $v = "deep"; my $c = PerlAPI::clone_chain(\\\\$v); $v = "changed"; ref($c) . "," . ref($$c) . "," . $$$$$c`, "REF,REF,deep"},
		{"SvRV of a value that is no reference", `my $c = PerlAPI::clone_chain("plain"); $c`, "plain"},
		{"addresses kept", `my @n = (1 .. 8); my @refs = map { \my $x } 1 .. 20000; ${$refs[$_]} = 1 for 0 .. 9999; join ",", PerlAPI::kept_addresses(\@refs, @n), @n`, "10000,12,23,4,5,6,7,8,9"},
		{"GvSV, GvAV, GvHV", `our ($slots, @slots, %slots) = ("old"); @slots = (9); join ",", (PerlAPI::glob_slots("main::slots") ? "set" : "not set"), $main::slots, "@main::slots", %main::slots`, "set,scalar,1 2,key,value"},
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

// Characters above U+1FFFFF have perl's extended UTF-8: five bytes and
// more. The encoder writes them, the length macros count them and the
// decoder reads them, as the interpreter does for the same code point.
func TestNativeExtendedUTF8(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `sub own { my $c = chr(shift); utf8::encode($c); unpack("H*", $c) . ":" . length($c) . ":" . length($c) } 1`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"four bytes", `PerlAPI::encode_uv(0x10FFFF)`, "f48fbfbf:4:4"},
		{"five bytes, as the interpreter encodes it", `no warnings; PerlAPI::encode_uv(0x310000) eq own(0x310000) ? PerlAPI::encode_uv(0x310000) : "differs: " . own(0x310000)`, "f88c908080:5:5"},
		{"six bytes", `no warnings; PerlAPI::encode_uv(0x7FFFFFFF) eq own(0x7FFFFFFF) ? "same" : own(0x7FFFFFFF)`, "same"},
		{"decoded", `PerlAPI::decode_first("\xf8\x8c\x90\x80\x80", 0)`, "310000:5"},
	})
}

// gv_name_set renames a glob; a module that names the subs it generates
// does it with that (GvNAME itself is not assignable, on perl neither).
func TestNativeGlobRename(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"gv_name_set", `$Renamed::old = 1; join ",", PerlAPI::rename_glob("Renamed::old", "fresh"), *Renamed::old{NAME}`, "fresh,fresh"},
	})
}
