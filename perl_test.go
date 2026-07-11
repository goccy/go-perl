package perl_test

import (
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// newInterp builds an interpreter backed by the embedded stdlib and closes it
// on test cleanup.
func newInterp(t *testing.T) *perl.Interpreter {
	t.Helper()
	i, err := perl.NewInterpreter(perl.Config{})
	if err != nil {
		t.Fatalf("NewInterpreter: %v", err)
	}
	t.Cleanup(func() { i.Close() })
	return i
}

func TestEvalArithmetic(t *testing.T) {
	i := newInterp(t)
	r, err := i.Eval("1 + 1")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if !r.Ok {
		t.Fatalf("eval not ok: error=%q stderr=%q", r.Error, r.Stderr)
	}
	if r.Result != "2" {
		t.Fatalf("1 + 1 = %q, want %q", r.Result, "2")
	}
}

func TestEvalPrint(t *testing.T) {
	i := newInterp(t)
	r, err := i.Eval(`print "hello\n"; 42`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if !r.Ok {
		t.Fatalf("eval not ok: error=%q stderr=%q", r.Error, r.Stderr)
	}
	if r.Stdout != "hello\n" {
		t.Fatalf("stdout = %q, want %q", r.Stdout, "hello\n")
	}
	if r.Result != "42" {
		t.Fatalf("result = %q, want %q", r.Result, "42")
	}
}

func TestEvalDie(t *testing.T) {
	i := newInterp(t)
	r, err := i.Eval(`die "boom\n"`)
	if err != nil {
		t.Fatalf("Eval (transport): %v", err)
	}
	if r.Ok {
		t.Fatalf("expected eval to fail, got ok with result %q", r.Result)
	}
	if !strings.Contains(r.Error, "boom") {
		t.Fatalf("error = %q, want it to contain %q", r.Error, "boom")
	}
}

func TestEvalUseModule(t *testing.T) {
	i := newInterp(t)
	// strict/warnings + a core module exercise @INC (the embedded stdlib).
	r, err := i.Eval(`use strict; use warnings; use List::Util qw(sum); sum(1,2,3,4)`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if !r.Ok {
		t.Fatalf("eval not ok: error=%q stderr=%q", r.Error, r.Stderr)
	}
	if r.Result != "10" {
		t.Fatalf("sum(1..4) = %q, want %q", r.Result, "10")
	}
}

// TestDeleteStashBackref is a regression test for a Perl→wasm miscompile:
// deleting a stash entry that holds a CV (its only backref) used to die with
// "panic: del_backref, svp=0". Root cause was the wasm build missing
// -fno-strict-aliasing, so TBAA let DSE delete a live store in
// Perl_sv_kill_backrefs's single-backref path (svp = (SV**)&av). Any regression
// in the build flags resurfaces here.
func TestDeleteStashBackref(t *testing.T) {
	i := newInterp(t)
	r, err := i.Eval(`package Bar; sub x { 1 } package main; delete $Bar::{x}; "done"`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if !r.Ok {
		t.Fatalf("eval not ok: error=%q stderr=%q", r.Error, r.Stderr)
	}
	if r.Result != "done" {
		t.Fatalf("result = %q, want %q", r.Result, "done")
	}
}

// TestListUtilFunctions guards against the archive-name collision that dropped
// List::Util's XS (boot_List__Util) from the linked wasm: Hash-Util and
// Scalar-List-Utils both archive to auto/.../Util/Util.a and one clobbered the
// other, leaving sum/max/first/reduce/uniq unresolved.
func TestListUtilFunctions(t *testing.T) {
	i := newInterp(t)
	r, err := i.Eval(`use List::Util qw(sum max first reduce uniq);` +
		`join(",", sum(1..10), max(3,9,2), (first { $_ > 5 } 1..10), (reduce { $a + $b } 1..5), join("", uniq(1,1,2,3,3)))`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if !r.Ok {
		t.Fatalf("eval not ok: error=%q stderr=%q", r.Error, r.Stderr)
	}
	if r.Result != "55,9,6,15,123" {
		t.Fatalf("List::Util result = %q, want %q", r.Result, "55,9,6,15,123")
	}
}

func TestPersistentState(t *testing.T) {
	i := newInterp(t)
	if _, err := i.Eval(`our $x = 40`); err != nil {
		t.Fatalf("Eval set: %v", err)
	}
	r, err := i.Eval(`$x + 2`)
	if err != nil {
		t.Fatalf("Eval get: %v", err)
	}
	if !r.Ok || r.Result != "42" {
		t.Fatalf("persistent $x: ok=%v result=%q (error=%q)", r.Ok, r.Result, r.Error)
	}
}
