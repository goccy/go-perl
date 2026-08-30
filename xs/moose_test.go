//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-perl/xs"
)

// TestMoose runs Moose — the metaclass-heavy XS dist Catalyst is built on —
// through the native SDK: the Class::MOP simple readers, the _method_map
// cache keyed on the stash's mro pkg_gen (guest op HV_PKG_GEN), gv_init
// stub expansion, overload flag application on role->instance composition,
// and the Moose::Exporter re-export flag magic whose svt_set hook fires
// through the set-magic anchor upgrade (reserved method -5).
//
// It needs a prepared directory (opt-in via GOPERL_MOOSE_DIR) holding:
//
//	Moose.so — `gperl xs build` output for the Moose dist
//	lib/     — the dist's pure-Perl half (blib/lib, what gperl installs
//	           into local/lib/perl5)
//	deps/    — a cpanm -L style tree (holding lib/perl5) with Moose's
//	           pure-Perl runtime deps (Class::Load, Data::OptList,
//	           Sub::Exporter, Package::Stash + ::PP, Params::Util 1.10+,
//	           Try::Tiny, Module::Runtime, Devel::StackTrace,
//	           Devel::OverloadInfo, Package::DeprecationManager,
//	           Dist::CheckConflicts, Eval::Closure, MRO::Compat, ...)
func TestMoose(t *testing.T) {
	dir := os.Getenv("GOPERL_MOOSE_DIR")
	if dir == "" {
		t.Skip("GOPERL_MOOSE_DIR not set; skipping the Moose acceptance test")
	}
	so := filepath.Join(dir, "Moose.so")
	if _, err := os.Stat(so); err != nil {
		t.Fatalf("GOPERL_MOOSE_DIR is set but %s is missing: %v", so, err)
	}

	p := newHostPerl(t)
	ctx := context.Background()

	if err := xs.Load(p, "Moose", so); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if r, err := p.Eval(ctx, `sub __t_minc { unshift @INC, @_; 1 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("inc helper: err=%v error=%v", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_minc",
		filepath.Join(dir, "lib"),
		filepath.Join(dir, "deps", "lib", "perl5")); err != nil {
		t.Fatalf("add inc: %v", err)
	}

	mustEval := func(what, src, want string) {
		t.Helper()
		r, err := p.Eval(ctx, src)
		if err != nil || r.Error != nil {
			t.Fatalf("%s: err=%v ok=%v error=%v", what, err, (r.Error == nil), r.Error)
		}
		if want != "" && r.Value.String() != want {
			t.Fatalf("%s = %q, want %q", what, r.Value.String(), want)
		}
	}

	mustEval("load", `
		$ENV{PERL_PARAMS_UTIL_PP} = 1;
		$ENV{PACKAGE_STASH_IMPLEMENTATION} = 'PP';
		require Moose;
		defined &Class::MOP::get_code_info ? 'XS' : 'PP';
	`, "XS")

	// Attributes, type constraint, BUILD, immutability (inlined
	// constructor), method modifiers.
	mustEval("class", `
		eval q{
			package Counter;
			use Moose;
			has count => (is => 'rw', isa => 'Int', default => 0);
			sub inc { my $self = shift; $self->count($self->count + 1) }
			before inc => sub { $_[0]->{log} .= 'b' };
			__PACKAGE__->meta->make_immutable;
			1;
		} or die $@;
		my $c = Counter->new;
		$c->inc; $c->inc;
		my $typed = eval { $c->count('x'); 1 } ? 'accepted' : 'rejected';
		join ',', $c->count, $c->{log}, $typed;
	`, "2,bb,rejected")

	// The Class::MOP method map (simple readers + _method_map +
	// get_code_info underneath), and its pkg_gen-driven pruning: deleting
	// the glob must invalidate the cached map.
	mustEval("method map", `
		my @m = sort Counter->meta->get_method_list;
		my $body = Counter->meta->find_method_by_name('inc')->body;
		my @info = Class::MOP::get_code_info($body);
		eval q{ package Dyn; use Moose; 1 } or die $@;
		Dyn->meta->add_method(gone => sub { 1 });
		my $before = (grep { $_ eq 'gone' } Dyn->meta->get_method_list) ? 1 : 0;
		{ no strict 'refs'; delete ${'Dyn::'}{gone}; }
		my $after = (grep { $_ eq 'gone' } Dyn->meta->get_method_list) ? 1 : 0;
		join ',', (grep { $_ eq 'inc' } @m) ? 'inc' : '-', @info, $before, $after;
	`, "inc,Counter,inc,1,0")

	// Role with overloading applied to an INSTANCE — the ToInstance XS
	// (Gv_AMG + SvAMAGIC_on through the guest).
	mustEval("overload role", `
		eval q{
			package Plain;
			use Moose;
			has label => (is => 'ro', default => 'p');
			package Stringy;
			use Moose::Role;
			use overload '""' => sub { 'S:' . $_[0]->label }, fallback => 1;
			1;
		} or die $@;
		my $o = Plain->new;
		my $alias = $o;
		Stringy->meta->apply($o);
		"$o|$alias";
	`, "S:p|S:p")

	// The re-export flag magic: setting it, seeing it, and having the
	// svt_set hook clear it when the glob is overwritten (set-magic
	// anchor upgrade + reserved method -5).
	mustEval("export flag", `
		no strict 'refs';
		*{'FlagT::target'} = sub { 1 };
		my $gv = \*{'FlagT::target'};
		my @s;
		push @s, Moose::Exporter::_export_is_flagged($gv) ? 1 : 0;
		Moose::Exporter::_flag_as_reexport($gv);
		push @s, Moose::Exporter::_export_is_flagged($gv) ? 1 : 0;
		*{'FlagT::target'} = sub { 2 };
		push @s, Moose::Exporter::_export_is_flagged($gv) ? 1 : 0;
		join ',', @s;
	`, "0,1,0")

	// no Moose (Moose::Exporter unimport bookkeeping).
	mustEval("no Moose", `
		eval q{ package CleanMe; use Moose; 1 } or die $@;
		my $had = CleanMe->can('has') ? 1 : 0;
		eval q{ package CleanMe; no Moose; 1 } or die $@;
		join ',', $had, CleanMe->can('has') ? 1 : 0;
	`, "1,0")
}
