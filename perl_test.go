package perl_test

import (
	"context"
	"strings"
	"testing"
	"time"

	perl "github.com/goccy/go-perl"
)

// resultStr renders an Eval result's scalar as its Perl string form; a
// non-scalar result renders as its kind (tests asserting on strings then
// fail with a readable value).
func resultStr(r perl.Result) string {
	if s, err := perl.As[perl.ScalarValue](r.Value); err == nil {
		return s.String()
	}
	return r.Value.Kind().String()
}

// newPerl builds an instance backed by the embedded stdlib and closes it on
// test cleanup.
func newPerl(t *testing.T) *perl.Perl {
	t.Helper()
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { p.Close() })
	return p
}

func TestEvalArithmetic(t *testing.T) {
	p := newPerl(t)
	r, err := p.Eval(context.Background(), "1 + 1")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("eval not ok: error=%v stderr=%q", r.Error, r.Stderr)
	}
	if resultStr(r) != "2" {
		t.Fatalf("1 + 1 = %q, want %q", resultStr(r), "2")
	}
}

func TestEvalPrint(t *testing.T) {
	p := newPerl(t)
	r, err := p.Eval(context.Background(), `print "hello\n"; 42`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("eval not ok: error=%v stderr=%q", r.Error, r.Stderr)
	}
	if r.Stdout != "hello\n" {
		t.Fatalf("stdout = %q, want %q", r.Stdout, "hello\n")
	}
	if resultStr(r) != "42" {
		t.Fatalf("result = %q, want %q", resultStr(r), "42")
	}
}

func TestEvalDie(t *testing.T) {
	p := newPerl(t)
	r, err := p.Eval(context.Background(), `die "boom\n"`)
	if err != nil {
		t.Fatalf("Eval (transport): %v", err)
	}
	if r.Error == nil {
		t.Fatalf("expected eval to fail, got ok with result %q", resultStr(r))
	}
	if !strings.Contains(r.Error.Error(), "boom") {
		t.Fatalf("error = %q, want it to contain %q", r.Error, "boom")
	}
}

func TestEvalUseModule(t *testing.T) {
	p := newPerl(t)
	// strict/warnings + a core module exercise @INC (the embedded stdlib).
	r, err := p.Eval(context.Background(), `use strict; use warnings; use List::Util qw(sum); sum(1,2,3,4)`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("eval not ok: error=%v stderr=%q", r.Error, r.Stderr)
	}
	if resultStr(r) != "10" {
		t.Fatalf("sum(1..4) = %q, want %q", resultStr(r), "10")
	}
}

// TestDeleteStashBackref is a regression test for a Perl→wasm miscompile:
// deleting a stash entry that holds a CV (its only backref) used to die with
// "panic: del_backref, svp=0". Root cause was the wasm build missing
// -fno-strict-aliasing, so TBAA let DSE delete a live store in
// Perl_sv_kill_backrefs's single-backref path (svp = (SV**)&av). Any regression
// in the build flags resurfaces here.
func TestDeleteStashBackref(t *testing.T) {
	p := newPerl(t)
	r, err := p.Eval(context.Background(), `package Bar; sub x { 1 } package main; delete $Bar::{x}; "done"`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("eval not ok: error=%v stderr=%q", r.Error, r.Stderr)
	}
	if resultStr(r) != "done" {
		t.Fatalf("result = %q, want %q", resultStr(r), "done")
	}
}

// TestListUtilFunctions guards against the archive-name collision that dropped
// List::Util's XS (boot_List__Util) from the linked wasm: Hash-Util and
// Scalar-List-Utils both archive to auto/.../Util/Util.a and one clobbered the
// other, leaving sum/max/first/reduce/uniq unresolved.
func TestListUtilFunctions(t *testing.T) {
	p := newPerl(t)
	r, err := p.Eval(context.Background(), `use List::Util qw(sum max first reduce uniq);`+
		`join(",", sum(1..10), max(3,9,2), (first { $_ > 5 } 1..10), (reduce { $a + $b } 1..5), join("", uniq(1,1,2,3,3)))`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("eval not ok: error=%v stderr=%q", r.Error, r.Stderr)
	}
	if resultStr(r) != "55,9,6,15,123" {
		t.Fatalf("List::Util result = %q, want %q", resultStr(r), "55,9,6,15,123")
	}
}

func TestPersistentState(t *testing.T) {
	p := newPerl(t)
	if _, err := p.Eval(context.Background(), `our $x = 40`); err != nil {
		t.Fatalf("Eval set: %v", err)
	}
	r, err := p.Eval(context.Background(), `$x + 2`)
	if err != nil {
		t.Fatalf("Eval get: %v", err)
	}
	if r.Error != nil || resultStr(r) != "42" {
		t.Fatalf("persistent $x: ok=%v result=%q (error=%v)", (r.Error == nil), resultStr(r), r.Error)
	}
}

// TestInstanceIsolation guards the copy-on-write snapshot: package-level state
// written in one instance must never be visible in another, including one
// created AFTER the write (both map the same shared image).
func TestInstanceIsolation(t *testing.T) {
	a := newPerl(t)
	if r, err := a.Eval(context.Background(), `our $leak = "from-a"; $leak`); err != nil || r.Error != nil {
		t.Fatalf("Eval in a: err=%v error=%v", err, r.Error)
	}
	b := newPerl(t)
	r, err := b.Eval(context.Background(), `defined $main::leak ? "leaked" : "clean"`)
	if err != nil {
		t.Fatalf("Eval in b: %v", err)
	}
	if r.Error != nil || resultStr(r) != "clean" {
		t.Fatalf("isolation: ok=%v result=%q (error=%v)", (r.Error == nil), resultStr(r), r.Error)
	}
}

// TestEvalContextCancel stops a runaway loop via context cancellation.
func TestEvalContextCancel(t *testing.T) {
	p := newPerl(t)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := p.Eval(ctx, `my $n = 0; while (1) { $n++ }`)
	if err == nil {
		t.Fatalf("expected cancellation error, got nil")
	}
	if ctx.Err() == nil || err != ctx.Err() {
		t.Fatalf("err = %v, want ctx.Err() (%v)", err, ctx.Err())
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("cancellation took %v, expected prompt stop", elapsed)
	}
	// The instance stays usable after a cancelled eval.
	r, err := p.Eval(context.Background(), `1 + 2`)
	if err != nil || r.Error != nil || resultStr(r) != "3" {
		t.Fatalf("post-cancel eval: err=%v ok=%v result=%q error=%v", err, (r.Error == nil), resultStr(r), r.Error)
	}
}
