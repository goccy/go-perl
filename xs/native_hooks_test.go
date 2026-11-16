//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// A checker that links the execution order of the op it is handed
// (LINKLIST) reads the links the interpreter set: from where the order
// starts they lead through the operands to the op, as a module that
// threads an op of its own in between follows them
// (Syntax::Keyword::Inplace); the op still runs as compiled.
func TestNativeReadsTheLinksTheInterpreterSet(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"links after LINKLIST", `PerlAPI::hooks("linklist", 1); my $f = eval q{ sub { my ($x, $y) = @_; sprintf "%s-%s", $x, $y } }; my $seen = PerlAPI::hooks("linklist", 0); join ",", $seen, $f->("a", "b")`, "4,a-b"},
	})
}

// SvRV reads the slot of a reference, as perl's does, also when the ROK
// flag is off.
func TestNativeReadsTheReferenceSlot(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the slot of a reference without its flag", `my $t = "target"; PerlAPI::rv_slot_after_rok_off($t) ? "the target" : "something else"`, "the target"},
	})
}

// A module puts a function of its own into PL_check (with wrap_op_checker
// or by assignment) and is handed every op of that type the compiler
// builds, with the ops below it; it calls the checker it replaced, and the
// hook is gone again when it puts that back. An op it gives a pp function
// of its own runs that function, which reads and writes the perl stack and
// the mark stack and may run the interpreter's pp function on what it left.
func TestNativeOpBuildHooksAndPPFunctions(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"additions compiled under the checker", `PerlAPI::hooks("add", 1); my $f = eval q{ sub { $_[0] + $_[1] } }; my $n = PerlAPI::hooks("add", 0); my $g = eval q{ sub { $_[0] + $_[1] } }; join ",", $f->(1, 2), $f->(40, 2), $g->(1, 2), $n, PerlAPI::hooks("add", 0)`, "1003,1042,3,1,1"},
		{"the result is an operand like any other", `PerlAPI::hooks("add", 1); my $f = eval q{ sub { my @l = (7, $_[0] + $_[1], 9); "@l" } }; PerlAPI::hooks("add", 0); $f->(1, 2)`, "7 1003 9"},
		{"a pp function that goes on to the interpreter's", `PerlAPI::hooks("join", 1); my $f = eval q{ sub { join "-", @_ } }; PerlAPI::hooks("join", 0); my $g = eval q{ sub { join "-", @_ } }; join ",", $f->(1, 2, 3), PerlAPI::hooks("join", 0), $f->(4 .. 9), PerlAPI::hooks("join", 0), $g->(1, 2)`, "1+2+3,3,4+5+6+7+8+9,6,1-2"},
		{"in a list that is long", `PerlAPI::hooks("join", 1); my $f = eval q{ sub { length join "-", @_ } }; PerlAPI::hooks("join", 0); join ",", $f->(1 .. 3000), PerlAPI::hooks("join", 0)`, "13892,3000"},
	})
}

// The mark pointer a pp function finds is its own frame's from its first
// statement, whatever native call went before: the frames of native calls
// are short-lived, a new one may lie where an earlier one was, and a
// pointer left by that one would be taken for the new frame's (the mark of
// a join around the module's addition was lost that way, rarely). The
// fixture's pp function counts the entries at which the pointer named
// other marks.
func TestNativePPFunctionMarksAfterEarlierCalls(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a join around an addition of the module's", `PerlAPI::hooks("add", 1); my $f = eval q{ sub { join "-", "a", $_[0] + $_[1], "b" } }; PerlAPI::hooks("add", 0); my %got; for (1 .. 100) { PerlAPI::big(); $got{$f->(1, 2)}++ } join ",", (map { "$_=$got{$_}" } sort keys %got), PerlAPI::hooks("add with foreign marks", 0)`, "a-1003-b=100,0"},
	})
}

// A module puts a function of its own into PL_rpeepp and is handed every
// sub the compiler finishes, in execution order; the op it sees there is
// the op it sees when it is handed the same sub again.
func TestNativePeepholeHook(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the joins of the subs compiled under the hook", `PerlAPI::hooks("peep", 1); my $f = eval q{ sub { my $a = join "-", @_; my $b = $a ? join("+", @_) : join("*", @_); $a . $b } }; my $n = PerlAPI::hooks("peep", 0); eval q{ sub { join "", @_ } }; join ",", $n, PerlAPI::hooks("peep", 0), $f->(1, 2)`, "3,3,1-21+2"},
	})
}

// A callback native code runs on a stack of its own (PUSHSTACKi) has no
// context of the code around the native call: a `last` in it finds no loop
// to leave and dies, which the call's G_EVAL catches, and the loop around
// goes on. Before, the stack was not switched, and the `last` left the loop
// through the native call's frame.
func TestNativeCallbackOnAStackOfItsOwn(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"last in the callback", `my @r; for my $i (1 .. 2) { my $e = PerlAPI::on_own_stack(sub { last }); push @r, $e =~ /^Can't "last" outside a loop block/ ? "caught" : "[$e]" } join ",", @r`, "caught,caught"},
		{"a callback that returns", `my $e = PerlAPI::on_own_stack(sub { 1 }); "[$e]"`, "[]"},
	})
}

// A call native code makes with G_EVAL leaves $@ to the caller, as perl's
// call_sv does: an exception object's bool overloading does not run on the
// way (Text::Stencil counts that it does not). Before, the interpreter asked
// for the truth of $@ after every such call.
func TestNativeEvalCallLeavesTheExceptionAlone(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a false exception object", `{ package FalseErr; our $n = 0; use overload 'bool' => sub { $n++; 0 }, fallback => 1 } my $r = PerlAPI::eval_call(sub { die bless {}, 'FalseErr' }); "$r,$FalseErr::n"`, "died with a reference,0"},
		{"a message", `PerlAPI::eval_call(sub { die "no\n" })`, "died"},
		{"a return", `PerlAPI::eval_call(sub { 1 })`, "returned"},
	})
}

// The op_targ of an op is as wide as a pointer, and a module keeps one of
// its own there (Object::Proto's typed setter: op_targ = PTR2IV(data)); its
// pp function reads it back whole. Before, the op the function was given
// had the low 32 bits of it, and Object::Proto faulted on the rest.
func TestNativeOpTargIsWhole(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a value past 32 bits", `sub tg {} PerlAPI::targ_calls(\&tg, 0x123456789); my $f = eval q{ sub { tg() } } or die $@; sprintf "%x", $f->()`, "123456789"},
	})
}

// A sub that was given a call checker has every call of it handed to the
// checker when the call is compiled; cv_get_call_checker says which checker
// a sub has, the interpreter's own being a function like any other.
func TestNativeCallsCompiledByAModule(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"calls compile into what the checker returns", `sub answer { "computed" } my $before = eval q{ sub { answer() } }; my $state = PerlAPI::constant_calls(\&answer, "constant"); my $after = eval q{ sub { answer(1, 2) . answer() } }; join ",", $state, $before->(), $after->()`, "default,ours,constant,computed,constantconstant"},
	})
}

// A destructor native code registered for a variable of its own runs on
// that variable when the code croaks: before its frame is gone, as when
// perl unwinds to the eval that catches the croak.
func TestNativeDestructorAtACroak(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"registered in the scope the function was called in", `eval { PerlAPI::croak_guarded(0) }; ($@ =~ /^(guarded failure)/)[0] . ":" . PerlAPI::destructor_found()`, "guarded failure:4711"},
		{"registered in a scope of its own", `eval { PerlAPI::croak_guarded(1) }; ($@ =~ /^(guarded failure)/)[0] . ":" . PerlAPI::destructor_found()`, "guarded failure:4711"},
		{"scopes are as they were", `my @seen; { local $_ = "kept"; eval { PerlAPI::croak_guarded(1) }; push @seen, $_ } push @seen, defined $_ ? "leaked" : "restored"; "@seen"`, "kept restored"},
	})
}

// A function a module gives to call_atexit runs when the interpreter is
// destroyed.
func TestNativeCallAtExit(t *testing.T) {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	perlAPIInstance(t, p)
	file := filepath.Join(t.TempDir(), "at-exit.txt")
	if r, err := p.Eval(context.Background(), `PerlAPI::at_exit_write('`+file+`'); 1`); err != nil || r.Error != nil {
		p.Close()
		t.Fatalf("register: err=%v error=%v", err, r.Error)
	}
	if _, err := os.Stat(file); err == nil {
		p.Close()
		t.Fatal("the function ran before the interpreter was destroyed")
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("the function did not run: %v", err)
	}
	if string(got) != "the interpreter is gone\n" {
		t.Fatalf("got %q", got)
	}
}

// A module runs the pp function of an operator for an op it made itself
// (`BINOP myop; ... PL_op = (OP *)&myop; PL_ppaddr[myop.op_type](aTHX)`) and
// gets the interpreter's semantics of the operator: the assigning forms
// store into the left operand, an overloaded operator is the class's, and a
// logical operator leaves the value it chose on the stack.
func TestNativePPFunctionForAnOpOfTheModulesOwn(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"+=", `my $t = 40; PerlAPI::pp_on_own_op($t, 2, 0); $t`, "42"},
		{".=", `my $t = "ab"; PerlAPI::pp_on_own_op($t, "cd", 1); $t`, "abcd"},
		{"||= of a false value", `my $t = 0; PerlAPI::pp_on_own_op($t, "default", 2); $t`, "default"},
		{"||= of a true value", `my $t = "set"; PerlAPI::pp_on_own_op($t, "default", 2); $t`, "set"},
		{"an overloaded +", `{ package Num; use overload "+" => sub { Num->new($_[0]{v} + (ref $_[1] ? $_[1]{v} : $_[1])) }, q("") => sub { "Num(" . $_[0]{v} . ")" }; sub new { bless { v => $_[1] }, $_[0] } } my $t = Num->new(40); PerlAPI::pp_on_own_op($t, 2, 0); "$t"`, "Num(42)"},
		{"many times over", `my $t = 0; PerlAPI::pp_on_own_op($t, $_, 0) for 1 .. 100; $t`, "5050"},
	})
}

// Names of the perl API that are other names of it: SAVEDESTRUCTOR (a
// destructor that takes its argument alone), hv_common_key_len, the old
// gv_HVadd and gv_AVadd, AvREAL and AvREIFY, IS_PADCONST and IS_PADGV,
// sv_usepvn_flags.
func TestNativeOtherNames(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"each does what the name it stands for does", `my %h; my $r = PerlAPI::other_names(\%h, "main::made_on_demand"); join "|", $r, join(",", sort keys %h), (defined *main::made_on_demand{HASH} ? "hash" : "no hash")`,
			"destructor 3,fetched value,entry value,made 1,put there,absent 0,exists 10,deleted 1,glob 11,array 10,pad 11,string handed over|made,put|hash"},
	})
}

// scan_version reads a version out of a text, the interpreter's or the
// module's own; newGVgen makes globs under generated names; Drand01 is the
// generator rand() uses, with its seed and the flag that says it is seeded;
// sv_cmp_locale compares; PL_sv_count counts.
func TestNativeVersionsGlobsAndRandomNumbers(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"scan_version", `PerlAPI::scan_a_version("1.2.3;the rest")`, "1.2.3|;the rest|5"},
		{"scan_version of what is no version", `eval { PerlAPI::scan_a_version("1.2.3 and more") }; $@ =~ /^Invalid version format/ ? "refused" : "taken: $@"`, "refused"},
		{"scan_version of a decimal version", `PerlAPI::scan_a_version("0.42;")`, "0.42|;|4"},
		{"newGVgen", `PerlAPI::generated_globs("Some::Pkg")`, "111,Some::Pkg"},
		{"Drand01 is rand's generator", `srand(7); my @want = (rand(), rand(), rand()); srand(7); my @got = PerlAPI::random_numbers(3); join ",", (grep { $got[$_] == $want[$_] } 0 .. 2), @got[3 .. 5]`, "0,1,2,1,-1,1"},
	})
}
