//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// A module that sets a signal aside for a moment — it keeps the handler it
// finds, sets the default action and puts the handler back, as IO::Tty
// does with SIGCHLD around grantpt — leaves the process's handling of the
// signal as it was: the next such signal is delivered to the process, not
// the end of it. The signal is sent to the thread the test runs on where
// the host has a way: a thread that is running the program is where a
// handler that is not the process's own does its damage.
func TestNativeModuleSetsASignalAside(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	delivered := make(chan os.Signal, 4)
	signal.Notify(delivered, syscall.SIGCHLD)
	defer signal.Stop(delivered)
	for _, how := range []string{"sigaction", "signal"} {
		runEvalCases(t, p, []evalCase{
			{"a handler is found and put back with " + how, `PerlAPI::set_signal_aside("` + how + `", ` + strconv.Itoa(int(syscall.SIGCHLD)) + `) ? "handler" : "none"`, "handler"},
		})
		if err := signalThisThread(syscall.SIGCHLD); err != nil {
			t.Fatal(err)
		}
		select {
		case <-delivered:
		case <-time.After(10 * time.Second):
			t.Fatalf("after %s: the signal was not delivered to the process", how)
		}
	}
}

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

// A compiled regular expression is a value native code can ask about and
// match with: its pattern and flags, a match in a string of the
// interpreter's or of the module's own, and where the groups matched.
func TestNativeRegexps(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the pattern, the flags, the offsets", `PerlAPI::rx_info(qr/a(b+)c/i, "xxABBCxx")`, "a(b+)c:fold:2-6:3-5:1"},
		{"no match", `PerlAPI::rx_info(qr/z/, "abc")`, "z:exact:no match"},
		{"a match leaves the capture variables of the program alone", `"keep" =~ /(ee)/; PerlAPI::rx_info(qr/a(b)/, "ab"); $1`, "ee"},
		{"pregcomp, and a string that is not the interpreter's", `PerlAPI::rx_compile("b+", "aabbbcc")`, "2-5"},
		{"a wide pattern and text", `PerlAPI::rx_compile("\x{263A}+", "ab\x{263A}\x{263A}c")`, "2-8"},
	})
}

// mg_find finds the magic the interpreter itself keeps on a value: the back
// references of a referent with weak references to it, the overload table
// of a class.
func TestNativeFindsInterpreterMagic(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"no magic", `my @a; PerlAPI::own_magic(\@a)`, ""},
		{"back references", `require Scalar::Util; my @a; my $w = \@a; Scalar::Util::weaken($w); PerlAPI::own_magic(\@a)`, "backref"},
		{"an overload table with the method", `package WithDeref { use overload '@{}' => sub { [] } } PerlAPI::own_magic(bless {}, "WithDeref")`, "amt:1"},
		{"an overload table without it", `package WithString { use overload '""' => sub { "s" } } PerlAPI::own_magic(bless {}, "WithString")`, "amt:0"},
	})
}

// pv_pretty, pv_display and pv_escape are the interpreter's.
func TestNativeStringsForDisplay(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"plain text", `PerlAPI::show("abc", 20)`, `"abc" "abc"\0 3`},
		{"control characters, cut short", `PerlAPI::show("a\tb\ncdefgh", 6)`, `"a\tb\n"... "a\tb\ncdefgh"\0 4`},
	})
}

// AvARRAY is the body of an array for as long as the code that asked for it
// uses it: one native call reads and stores into the bodies of any number
// of arrays.
func TestNativeArrayBodiesInOneCall(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a few arrays", `my @a = ([1, 2, 3], [4], [], [5, 6]); join "|", PerlAPI::firsts(\@a), map { "@$_" } @a`, "145|1 2 3|4||5 2"},
		{"many arrays", `my @a = map { [$_ % 10, "x", "y"] } 1 .. 1000; my $f = PerlAPI::firsts(\@a); join ",", length($f), substr($f, 0, 12), scalar(grep { $_->[2] == 3 && $_->[1] eq "x" } @a)`, "1000,123456789012,1000"},
		{"an array twice", `my $r = [7, 8]; my @a = ($r, $r, $r); join "|", PerlAPI::firsts(\@a), "@$r"`, "777|7 2"},
	})
}

// SvPVX and SvLEN are lvalues: code that takes the buffer of an SV over
// leaves the SV without one, and the buffer outlives the SV.
func TestNativeStringBufferTakenOver(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the buffer is taken", `PerlAPI::take_buffer("taken over")`, "1:1:taken over"},
		{"many times", `my $n = 0; for my $i (1 .. 300) { $n++ if PerlAPI::take_buffer("s$i" x 20) eq "1:1:" . ("s$i" x 20) } $n`, "300"},
	})
}

// A const op is built from an SV, which is not an op: the constructor
// describes the op, not its operand.
func TestNativeConstOpsFromSVs(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"many const ops", `PerlAPI::const_ops(2000)`, "2000"},
	})
}

// PERL_ASYNC_CHECK is where native code that runs long lets a signal's
// handler run: the handler's die ends the native call.
func TestNativeLoopIsInterruptedByASignalHandler(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"nothing pending", `PerlAPI::spin(5)`, "5"},
		{"the handler dies", `my $t = time; my $n = eval { local $SIG{ALRM} = sub { die "TIMEOUT\n" }; alarm 1; PerlAPI::spin(20000) }; alarm 0; join "|", $n // "undef", $@, time - $t < 10 ? "soon" : "late"`, "undef|TIMEOUT\n|soon"},
		{"the handler returns", `local $SIG{ALRM} = sub { $PerlAPI::stop = 1 }; alarm 1; my $n = PerlAPI::spin(20000); alarm 0; $n < 10000 ? "stopped" : "ran $n"`, "stopped"},
	})
}

// The key of a hash is any bytes: hv_store, hv_fetch, hv_delete and
// hv_exists take it with its length, bytes that are 0 included.
func TestNativeHashKeysAreBytes(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"integers as keys", `PerlAPI::by_number(1000)`, "1000:1000:500:500"},
		{"one key", `PerlAPI::by_number(1)`, "1:1:1:1"},
	})
}

// CvNAME_HEK is the name a sub has by itself, and CvNAME_HEK_set gives it
// one made with share_hek.
func TestNativeSubOwnName(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a lexical sub", `use feature "lexical_subs"; my sub lexy { 1 } PerlAPI::own_name(\&lexy)`, "lexy"},
		{"a sub in a glob", `sub PerlAPITest::in_glob { 1 } PerlAPI::own_name(\&PerlAPITest::in_glob)`, ""},
		{"renamed", `use feature "lexical_subs"; my sub lexy { (caller(0))[3] } join "|", PerlAPI::own_name(\&lexy, "other"), PerlAPI::own_name(\&lexy), lexy()`, "other|other|other"},
		{"a name in UTF-8", `use utf8; use feature "lexical_subs"; my sub lexy { 1 } my $n = PerlAPI::own_name(\&lexy, "n\x{e9}\x{263a}"); join "|", length($n), utf8::is_utf8($n) ? 1 : 0, ord(substr($n, 2))`, "3|1|9786"},
	})
}

// The aux of a signature op is perl's op_multiparam_aux: the counts of the
// parameters, their pad indexes and the named ones.
func TestNativeSignatureParameters(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"no signature", `PerlAPI::signature(sub { 1 })`, "none"},
		{"positional", `use feature "signatures"; PerlAPI::signature(sub ($a, $b) { 1 })`, "2:2:0:-:p:p"},
		{"optional, unnamed and slurpy", `use feature "signatures"; PerlAPI::signature(sub ($a, $, $c = 3, @rest) { 1 })`, "2:3:0:@:p:_:p"},
		{"a hash", `use feature "signatures"; PerlAPI::signature(sub ($a, %h) { 1 })`, "1:1:0:%:p"},
		{"the op kept for a later call", `use feature "signatures"; my $code = sub ($a, $b, $c = 1) { 1 }; my @s = (PerlAPI::signature($code)); my @junk = map { [1 .. 50] } 1 .. 200; push @s, PerlAPI::signature(); "@s"`, "2:3:0:-:p:p:p 2:3:0:-:p:p:p"},
	})
}

// MULTICALL calls a block many times with $_ as the caller sets it, and
// the caller reads what the block returned from the top of the stack.
func TestNativeMulticall(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the first match", `PerlAPI::first_index(sub { $_ > 3 }, 1, 5, 2, 9)`, "1"},
		{"no match", `PerlAPI::first_index(sub { $_ eq "x" }, qw(a b c))`, "-1"},
		{"a block that returns nothing", `PerlAPI::first_index(sub { return }, 1, 2)`, "-1"},
		{"a block that compares $a and $b, called where only multicall_cop is known", `PerlAPI::largest(sub { $a <=> $b }, 3, 11, 7, 2)`, "11,plain"},
		{"the block is called in the context of the call, void too", `our $kept; sub keep { $kept = PerlAPI::largest(sub { $a <=> $b }, 4, 8, 6); 1 } keep(); my @l = PerlAPI::largest(sub { $a <=> $b }, 5, 9, 7); PerlAPI::largest(sub { $a <=> $b }, 1, 2); "$kept|@l"`, "8,plain|9,plain"},
		{"by another order", `PerlAPI::largest(sub { length($a) <=> length($b) }, "ccc", "a", "bbbbb", "dd")`, "bbbbb,plain"},
		{"the caller's $a and $b are as they were", `our ($a, $b) = ("left", "right"); PerlAPI::largest(sub { $a <=> $b }, 1, 2); "$a,$b"`, "left,right"},
		{"what a block returns in list context is on the stack of the calls", `PerlAPI::each_pair(sub { ($a + $b, $a * $b) }, 1, 2, 3, 4)`, "3,2,;5,6,;7,12,;args 1 2 3 4"},
		{"a block that returns nothing, and one that returns more than the stack held", `PerlAPI::each_pair(sub { $a == 1 ? () : (($a) x 3) }, 1, 2, 3) . "|" . length(PerlAPI::each_pair(sub { (1) x 200 }, 1, 2))`, ";2,2,2,;args 1 2 3|409"},
		{"an overloaded <=>", `{ package Rank; use overload "<=>" => sub { ($_[2] ? -1 : 1) * ($_[0]{r} <=> (ref $_[1] ? $_[1]{r} : $_[1])) }, q("") => sub { "rank" . $_[0]{r} }; } PerlAPI::largest(sub { $a <=> $b }, (map { bless { r => $_ }, "Rank" } 2, 9, 4))`, "rank9,overloaded -1"},
		{"many calls", `my $n = 0; join ",", PerlAPI::first_index(sub { $n++; $_ == 4999 }, 0 .. 5000), $n`, "4999,5000"},
		{"$_ is the caller's again", `my @r; for ("outer") { PerlAPI::first_index(sub { 0 }, 1, 2); push @r, $_ } "@r"`, "outer"},
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

// An XSUB may read its arguments with dMARK and dSP and never say dXSARGS
// (Params::Classify's do): the frame is current, and has a catch point for
// croak, before the function's first statement.
func TestNativeXSUBWithoutDXSARGS(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"one argument", `PerlAPI::sum_without_dxsargs(5)`, "105"},
		{"two arguments", `PerlAPI::sum_without_dxsargs(5, 6)`, "111"},
		{"it croaks", `eval { PerlAPI::sum_without_dxsargs() }; $@ =~ /^Usage: PerlAPI::sum_without_dxsargs\(a, \[b\]\)/ ? "usage" : "other: $@"`, "usage"},
		{"and is called again", `PerlAPI::sum_without_dxsargs(1)`, "101"},
		{"the caller's stack is intact", `my @r = (1, PerlAPI::sum_without_dxsargs(2, 3), 9); "@r"`, "1 105 9"},
		{"from a callback of another XSUB", `PerlAPI::call_on_new_stack(sub { PerlAPI::sum_without_dxsargs($_[0]) }, 21)`, "121"},
		{"a croak in a callback leaves the outer XSUB running", `PerlAPI::call_on_new_stack(sub { eval { PerlAPI::sum_without_dxsargs() }; $@ ? "survived" : "no error" }, 1)`, "survived"},
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
