//go:build darwin || linux

package xs_test

import (
	"context"
	"fmt"
	"testing"
)

// BenchmarkNativeAccessors measures what the accessor macros of the SDK
// cost native code: one iteration is one call of an XSUB that evaluates an
// accessor many times (accessorEvaluations, or every element of an array
// arrayRounds times over), so the time of an iteration divided by that
// number is the cost of one evaluation, the call itself included.
//
// The accessors that are lvalues (SvPVX, SvLEN, SvIVX, AvARRAY) designate
// storage that is loaded from the interpreter and compared before the next
// operation; SvPVX_const, SvIV and av_fetch are one operation each, and
// SvCUR keeps the last scalar's length. The pairs show what the lvalue form
// costs over the plain one.
func BenchmarkNativeAccessors(b *testing.B) {
	const accessorEvaluations = 10000
	const arrayRounds = 20
	p := newHostPerl(b)
	perlAPIInstance(b, p)
	ctx := context.Background()
	setup := `our $str = "a string of some length"; our $num = 42; our @small = (1 .. 16); our @large = (1 .. 1000); 1`
	if r, err := p.Eval(ctx, setup); err != nil || r.Error != nil {
		b.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	run := func(name, code, want string) {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r, err := p.Eval(ctx, code)
				if err != nil || r.Error != nil {
					b.Fatalf("err=%v error=%v", err, r.Error)
				}
				if got := resultStr(r); got != want {
					b.Fatalf("got %q, want %q", got, want)
				}
			}
		})
	}
	scalar := func(name, variable string, how int, each uint64) {
		run(name, fmt.Sprintf(`PerlAPI::accessor_loop(%s, %d, %d)`, variable, accessorEvaluations, how),
			fmt.Sprint(each*accessorEvaluations))
	}
	scalar("SvPVX_const/one_operation", "$str", 0, 'a')
	scalar("SvPVX/lvalue", "$str", 1, 'a')
	scalar("SvCUR/kept_length", "$str", 3, 23)
	scalar("SvIV/one_operation", "$num", 5, 42)
	scalar("SvIVX/lvalue", "$num", 4, 42)
	for _, arr := range []struct {
		name string
		len  int
	}{{"small", 16}, {"large", 1000}} {
		for how, form := range []string{"av_fetch", "AvARRAY_each_element", "AvARRAY_once_a_round"} {
			run(fmt.Sprintf("%s/%d_elements", form, arr.len),
				fmt.Sprintf(`PerlAPI::array_loop(\@%s, %d, %d)`, arr.name, arrayRounds, how),
				fmt.Sprint(arr.len*arrayRounds))
		}
	}
}

// BenchmarkNativeHashBuild measures what a hash costs that native code
// builds the way YAML::Syck's loader builds one from a mapping: per pair
// nine operations of the API (two newSVpvn, newSVsv, sv_2mortal, SvPV,
// hv_store_ent, SvREFCNT_inc, SvREFCNT_dec twice). One iteration is one
// call that builds a hash of hashKeys pairs; the time divided by hashKeys
// is the cost of a pair.
func BenchmarkNativeHashBuild(b *testing.B) {
	const hashKeys = 10000
	p := newHostPerl(b)
	perlAPIInstance(b, p)
	ctx := context.Background()
	code := fmt.Sprintf(`scalar keys %%{ PerlAPI::build_map(%d) }`, hashKeys)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, err := p.Eval(ctx, code)
		if err != nil || r.Error != nil {
			b.Fatalf("err=%v error=%v", err, r.Error)
		}
		if got := resultStr(r); got != fmt.Sprint(hashKeys) {
			b.Fatalf("got %q, want %d", got, hashKeys)
		}
	}
}
