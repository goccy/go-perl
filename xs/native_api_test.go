//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// A callback a module makes without G_EVAL leaves $@ as it was when the
// code it called returns normally: only a call that asks for G_EVAL
// empties it.
func TestNativeCallbackLeavesErrsvAlone(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a message", `eval { die "kept\n" }; PerlAPI::call_plain(sub { 1 }); $@`, "kept\n"},
		{"an object", `my $e = bless [], "My::Err"; eval { die $e }; PerlAPI::call_plain(sub { 1 }); ref($@) . ($@ == $e ? ":same" : ":other")`, "My::Err:same"},
		{"the callee's own eval empties it, as in perl", `eval { die "old\n" }; PerlAPI::call_plain(sub { eval { 1 } }); "[$@]"`, "[]"},
		{"the callee sees it", `eval { die "old\n" }; my $seen; PerlAPI::call_plain(sub { $seen = $@ }); $seen`, "old\n"},
		{"an exception that is false is still an exception", `package Falsy { use overload bool => sub { 0 }, '""' => sub { "falsy" } } my $went_on = 0; eval { PerlAPI::call_plain(sub { die bless [], "Falsy" }); $went_on = 1 }; ref($@) . ":$went_on"`, "Falsy:0"},
		{"a callee that dies replaces it", `eval { die "old\n" }; eval { PerlAPI::call_plain(sub { die "new\n" }) }; $@`, "new\n"},
	})
}

// The string buffer of an SV is the interpreter's memory. A module that
// frees it (Safefree(SvPVX(sv)), then SvPV_set(sv, NULL)) or resizes it
// (Renew) gets the interpreter's allocator to do it.
func TestNativeModuleFreesAndResizesInterpreterBuffers(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `use B (); sub kind { my $f = B::svref_2object(\$_[0])->FLAGS; join "", ($f & B::SVf_IOK ? "i" : ""), ($f & B::SVf_POK ? "p" : "") } 1`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"the string of a converted value is discarded", `my $s = "9" . "9"; PerlAPI::discard_string($s); kind($s) . ":" . ($s + 1)`, "i:100"},
		{"many times over", `for (1 .. 2000) { my $s = "1" . $_; PerlAPI::discard_string($s); $s == "1$_" or return "wrong at $_" } "all"`, "all"},
		{"a buffer resized by the module", `my $s = "abc" . "def"; PerlAPI::resize_buffer($s, 4000); length($s) . ":" . substr($s, 0, 8) . ":" . substr($s, -2)`, "4000:abcdef..:.."},
	})
}

// TAINT_IF and TAINT_PROPER are the interpreter's taint check: with taint
// mode on, tainted data in use ends the call with perl's message.
func TestNativeTaintChecks(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"no taint mode, no check", `PerlAPI::insecure("x", "the test")`, "1"},
		{"clean data passes", `PerlAPI::set_tainting(1); my $r = PerlAPI::insecure("x", "the test"); PerlAPI::set_tainting(0); $r`, "1"},
		{"tainted data dies", `require Scalar::Util; PerlAPI::set_tainting(1); open my $in, "<", \"x" or die $!; my $t = <$in>; my $is = Scalar::Util::tainted($t) ? "tainted" : "clean"; eval { PerlAPI::insecure($t, "the test") }; my $e = $@; PerlAPI::set_tainting(0); $is . ": " . ($e =~ /^(Insecure dependency in the test while running with -T switch) at / ? $1 : $e)`, "tainted: Insecure dependency in the test while running with -T switch"},
	})
}

// $? is a variable a module can save, change and get restored when its
// call ends (SAVEI32(PL_statusvalue), the `local $?` of an object system),
// and a saved int is restored as an int: what lies beside it is untouched.
func TestNativeSavedIntegers(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"$? saved around a callback", `$? = 3; my $seen; PerlAPI::with_local_status(sub { $seen = $?; $? = 9 }); "$seen,$?"`, "0,3"},
		{"an int saved in a scope", `join ",", PerlAPI::save_one_int()`, "1,7"},
	})
}

// do_openn opens a handle the way perl's open does, and the PerlLIO
// functions are the low-level I/O of the C library.
func TestNativeOpenFunctions(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(file, []byte("first line\nsecond\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `our $file = '`+file+`'; 1`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"do_openn reads a file", `PerlAPI::open_and_read("<", $file, 10)`, "first line"},
		{"do_openn with a layer", `PerlAPI::open_and_read("<:raw", $file, 5)`, "first"},
		{"do_openn of a missing file", `defined(PerlAPI::open_and_read("<", "$file.missing", 5)) ? "opened" : "failed: " . ($!{ENOENT} ? "ENOENT" : "$!")`, "failed: ENOENT"},
		{"PerlLIO", `PerlAPI::lio_peek($file)`, "18:f"},
	})
}

// dXSARGS works in a function that was given no interpreter context: the
// glue of Inline::C reads and returns its arguments from a plain C function
// (Inline_Stack_Vars) that its XSUB calls.
func TestNativeStackInAPlainFunction(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"arguments and results", `join ",", PerlAPI::stack_in_a_plain_function(5, 6, 7)`, "3,18,5"},
		{"again, with one argument", `join ",", PerlAPI::stack_in_a_plain_function(40)`, "1,40,40"},
		{"the caller's stack is intact", `my @r = (1, PerlAPI::stack_in_a_plain_function(2, 3), 9); "@r"`, "1 2 5 2 9"},
	})
}

// A native module registered with an instance defines nothing in its
// package until it is loaded, like an extension linked into the
// interpreter; loading it installs its boot function then.
func TestNativeModuleIsNotInItsPackageUntilLoaded(t *testing.T) {
	p := newHostPerl(t)
	so := buildPerlAPI(t)
	loadXSModule(t, p, "PerlAPI", so)
	runEvalCases(t, p, []evalCase{
		{"before the module is loaded", `no strict "refs"; join ",", (defined &PerlAPI::bootstrap ? "bootstrap" : "no bootstrap"), scalar(grep { defined &{"PerlAPI::$_"} } keys %PerlAPI::)`, "no bootstrap,0"},
		{"loading it", `require XSLoader; XSLoader::load("PerlAPI"); "loaded"`, "loaded"},
		{"after", `join ",", (defined &PerlAPI::bootstrap ? "bootstrap" : "no bootstrap"), (defined &PerlAPI::flags ? "functions" : "no functions")`, "bootstrap,functions"},
		{"a module nobody registered", `eval { XSLoader::load("No::Such::Native::Module") }; $@ ? "refused" : "loaded"`, "refused"},
	})
}
