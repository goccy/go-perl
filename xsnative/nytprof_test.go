//go:build darwin || linux

package xsnative_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/xsnative"
)

// TestDevelNYTProf runs Devel::NYTProf — an XS dist that hooks the
// interpreter itself — through the native SDK: PL_ppaddr slot replacement
// routed back as pp hooks, run_original_op through the proxy table,
// OP/COP/context-stack shadows, save-stack destructors for subroutine
// timing, and the profile writer/reader.
//
// It needs a prepared directory (opt-in via GOPERL_NYTPROF_DIR) holding:
//
//	NYTProf.so — NYTProf.c + FileHandle.c (xsubpp output of the dist's two
//	             .xs files) compiled against sdk/include with the host C
//	             compiler, matching this test binary's architecture:
//	               cc -shared -fPIC -O2 -DHAS_GETTIMEOFDAY \
//	                  -DXS_VERSION='"6.14"' -DVERSION='"6.14"' \
//	                  -I <go-perl>/xsnative/sdk/include -I <dist> \
//	                  -o NYTProf.so NYTProf.c FileHandle.c
//	lib/       — the dist's pure-Perl half (Devel-NYTProf-*/lib)
func TestDevelNYTProf(t *testing.T) {
	dir := os.Getenv("GOPERL_NYTPROF_DIR")
	if dir == "" {
		t.Skip("GOPERL_NYTPROF_DIR not set; skipping the Devel::NYTProf acceptance test")
	}
	so := filepath.Join(dir, "NYTProf.so")
	if _, err := os.Stat(so); err != nil {
		t.Fatalf("GOPERL_NYTPROF_DIR is set but %s is missing: %v", so, err)
	}

	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	ctx := context.Background()

	if err := xsnative.Load(p, "Devel::NYTProf", so); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if r, err := p.Eval(ctx, `sub __t_ninc { unshift @INC, @_; 1 } 1;`); err != nil || !r.Ok {
		t.Fatalf("inc helper: err=%v error=%q", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_ninc", filepath.Join(dir, "lib")); err != nil {
		t.Fatalf("add inc: %v", err)
	}

	out := filepath.Join(t.TempDir(), "nytprof.out")

	mustEval := func(what, src, want string) string {
		t.Helper()
		r, err := p.Eval(ctx, src)
		if err != nil || !r.Ok {
			t.Fatalf("%s: err=%v ok=%v error=%q", what, err, r.Ok, r.Error)
		}
		res := fmt.Sprint(r.Result)
		if want != "" && res != want {
			t.Fatalf("%s = %q, want %q", what, res, want)
		}
		return res
	}

	mustEval("arm profiler", `
		$ENV{NYTPROF} = 'file=`+out+`:start=begin:sigexit=0';
		require Devel::NYTProf;
		'armed';
	`, "armed")

	// The profiled workload: statement hooks fire per line, entersub hooks
	// per call (fib(12) is exactly 465 calls), leave hooks per scope exit,
	// and each perl-sub call parks a save-stack destructor for its timing.
	mustEval("workload", `
		sub fib { my $n = shift; $n < 2 ? $n : fib($n-1) + fib($n-2) }
		sub greet { my $who = shift; "hello, $who" }
		my $x = 0;
		for my $i (1 .. 50) { $x += $i }
		my $f = fib(12);
		my $g = greet('gopher');
		"$x|$f|$g";
	`, "1275|144|hello, gopher")

	mustEval("finish", `DB::finish_profile(); 'finished';`, "finished")

	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("profile output missing: %v", err)
	}
	if st.Size() < 1000 {
		t.Fatalf("profile suspiciously small: %d bytes", st.Size())
	}
	head := make([]byte, 12)
	fh, _ := os.Open(out)
	fh.Read(head)
	fh.Close()
	if !strings.HasPrefix(string(head), "NYTProf ") {
		t.Fatalf("profile header = %q, want a NYTProf signature", head)
	}

	// Read the profile back through the reader XS and verify exact call
	// counts survived the whole pipeline.
	subs := mustEval("read profile", `
		require Devel::NYTProf::Data;
		my $data = Devel::NYTProf::Data->new({filename => '`+out+`', quiet => 1});
		my $subs = $data->subname_subinfo_map;
		my @lines;
		for my $name (sort keys %$subs) {
			next unless $name =~ /::(fib|greet)$/;
			push @lines, join(":", $name, $subs->{$name}->calls);
		}
		join("|", @lines);
	`, "")
	if subs != "main::fib:465|main::greet:1" {
		t.Fatalf("profiled subs = %q, want main::fib:465|main::greet:1", subs)
	}
}
