//go:build darwin || linux

package xs_test

import (
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// A save native code makes of an interpreter variable is the
// interpreter's: the value is put back when the scope it was made in ends —
// the call of an XSUB (perl's entersub opens one around it), the sub a pp
// function runs in, not the pp function's return — and LEAVE_SCOPE undoes
// the saves above an index of the save stack.
func TestNativeSavesOfInterpreterVariables(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the scope of an XSUB's call", `$? = 0; my $in; { PerlAPI::status_saved(5); $in = $?; } join ",", $in, $?`, "0,0"},
		{"the scope a pp function runs in", `PerlAPI::hooks("status", 1); my $f = eval q{ sub { my $n = -$_[0]; ($n, $?) } }; my $seen = PerlAPI::hooks("status", 0); $? = 0; my @r = $f->(4); join ",", $seen, @r, $?`, "1,-4,99,0"},
		{"left to an index of the save stack", `$? = 0; PerlAPI::status_left(7) . "," . $?`, "7000,0"},
	})
}

// The padlist, the list of pad names and the pad names of a sub are the
// interpreter's structures with perl's fields: a module reads them, stores
// a new pad name of the interpreter's into the list in place of one it
// releases, sets the sequence range of a name, and the interpreter has
// what it stored.
func TestNativePadStructures(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"names, a stored name and its range", `my $f = sub { my $x = 1; my @y = (2); $x + @y };
			my $before = PerlAPI::pad_names($f);
			my $a = PerlAPI::pad_names($f, '$x', '$renamed');
			my $b = PerlAPI::pad_names($f);
			join ",", ($before =~ /^\$x\[\d+,\d+\] \@y\[\d+,\d+\]$/ ? "read" : $before),
				($a =~ /^\$renamed\[1,2147483648\] \@y\[\d+,\d+\]$/ ? "renamed" : $a),
				($b eq $a ? "kept" : $b), $f->()`, "read,renamed,kept,2"},
		{"of a sub that is gone", `PerlAPI::names_of_a_sub_that_goes(q{sub { my ($a, $b, $c, $d, $e) = @_; my @f; my %g; $a + $b }})`, "7"},
	})
}

// PL_curpad given values in the module's own memory (Object::Pad's
// constructor makes a pad of two values for the field hooks it runs): they
// are the values of the pad in use for the module and for an XSUB it
// calls, a sub the interpreter runs meanwhile has its own, and the scope
// that saved PL_curpad puts back the pad of the caller.
func TestNativeValuesOfAPadOfItsOwn(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"values of the module's own", `sub inner { my $v = "theirs"; PerlAPI::pad_first() }
			sub outer { my $w = "caller"; PerlAPI::own_pad(\&inner, \&PerlAPI::pad_first) . "," . PerlAPI::pad_first() }
			outer()`, "mine,theirs,mine,mine,theirs,caller"},
	})
}

// The fields of a sub are lvalues (CvFILE, CvFLAGS, CvOUTSIDE_SEQ), and a
// module makes a sub of its own from another's parts the way
// Object::Pad's cv_copy_flags does: the fields of its body (SvANY) with a
// name of its own, the shared ops, a padlist of pad_new with PL_compcv and
// PL_comppad_name saved, the values of its pad through PL_curpad.
func TestNativeSubFields(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"assigned fields", `require B; my $f = sub { 42 };
			my ($n0, $l0) = split /:/, PerlAPI::sub_fields($f), 3;
			my ($n1, $l1, $f1) = split /:/, PerlAPI::sub_fields($f, "elsewhere.pl"), 3;
			join ",", $n1 - $n0, $l0, $l1, $f1, B::svref_2object($f)->FILE, $f->()`,
			"1,-,lexical,elsewhere.pl,elsewhere.pl,42"},
		{"a sub copied from another", `use feature 'lexical_subs'; no warnings; my $k = 10; my sub orig_sub { my ($v) = @_; my @l = (1, 2); $v + @l + $k }
			my $c = PerlAPI::copy_sub(\&orig_sub);
			join ",", $c->(5), orig_sub(5), PerlAPI::own_name($c), $c->(1)`,
			"17,17,orig_sub,13"},
	})
}

// The hints of a COP (%^H when its statement was compiled) are read
// through its hints hash: a value, a key that is not there
// (&PL_sv_placeholder), and whether a key exists.
func TestNativeHintsOfACOP(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"present and absent", `BEGIN { $^H{"goperl/test"} = "on" } join ",", PerlAPI::hint("goperl/test"), PerlAPI::hint("goperl/none")`,
			"on:exists,absent:-"},
	})
}

// CALLRUNOPS runs ops native code built on the stack native code has: what
// they leave above the mark is the native code's afterwards (Object::Pad's
// exec_optree_list).
func TestNativeRunsOps(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a list of constants", `my ($got, $n) = PerlAPI::run_list("a", 2, "c"); join ",", $n, @$got`, "3,a,2,c"},
		{"an empty list", `my ($got, $n) = PerlAPI::run_list(); join ",", $n, scalar @$got`, "0,0"},
	})
}

// The fields of an object (SVt_PVOBJ): a module makes an object with an
// array of its own (Newx(ObjectFIELDS(obj), ...)) as Object::Pad does for
// :repr(pvobj), fills it, and reads the fields back; the object is freed
// as perl frees one.
func TestNativeFieldsOfAnObject(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"made and read", `my $o = PerlAPI::fields_object("Some::Class", 10, "x", 30); join ",", ref($o), PerlAPI::fields_of($o)`,
			"Some::Class,10,x,30"},
		{"many, freed", `my $n = 0; for (1 .. 50) { my $o = PerlAPI::fields_object("Some::Class", $_, $_ * 2); $n++ if PerlAPI::fields_of($o) eq "$_," . ($_ * 2) } $n`, "50"},
	})
}

// newSTATEOP needs a parser: a module that finds none makes one of its own
// (Newxz(PL_parser, ...), as Object::Pad does at run time), which the
// interpreter takes for the scope; a statement op made then has the line of
// the statement running, and the parser is gone with the scope.
func TestNativeParserOfItsOwn(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(t.Context(), "sub probe_line {\n\n  join ',', PerlAPI::statement_line() }\n1"); err != nil || r.Error != nil {
		t.Fatalf("define: err=%v died=%v", err, r.Error)
	}
	got, err := p.Call(t.Context(), "probe_line")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Call returned %d values", len(got))
	}
	sv, err := perl.As[perl.ScalarValue](got[0])
	if err != nil {
		t.Fatalf("As: %v", err)
	}
	// line 3; no parser before the call (the program is compiled), none
	// after its scope
	if s := sv.String(); s != "3,,," && s != "3,,"  {
		t.Fatalf("got %q, want the line 3 and no parser before or after", s)
	}
}

// A shared key goes into a message with %" HEKf ".
func TestNativeSharedKeyInAMessage(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	r, err := p.Eval(t.Context(), `eval { PerlAPI::stash_hek(bless {}, "Named::Class") }; $@`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if got := resultStr(r); !strings.HasPrefix(got, "class Named::Class says hello") {
		t.Fatalf("got %q", got)
	}
}

// HvARRAY is the hash's buckets, each the chain of its entries, or NULL for
// a hash that has none yet; perlio.h is a header a module includes.
func TestNativeBucketsOfAHash(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"none and some", `my %e; my %h = (b => 2, a => 1, c => 3); join ",", PerlAPI::buckets(\%e), PerlAPI::buckets(\%h)`, "none,3:a:b:c"},
		{"many", `my %h = map { $_ => 1 } 1 .. 200; my ($n) = split /:/, PerlAPI::buckets(\%h); $n`, "200"},
	})
}

// custom_op_xop is the description a custom op's pp function was
// registered with, and perl's empty one, whose fields have their defaults,
// for a pp function nothing registered.
func TestNativeCustomOpDescription(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"registered", `PerlAPI::custom_op_description(1)`, "goperl_described|a described op|1"},
		{"not registered", `PerlAPI::custom_op_description(0)`, "custom|unknown custom operator|0"},
	})
}

