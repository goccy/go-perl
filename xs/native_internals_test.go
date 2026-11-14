//go:build darwin || linux

package xs_test

import (
	"os"
	"path/filepath"
	"testing"
)

// What native code stores into a context of the interpreter's reaches it,
// also after an operation came between reading the context and storing
// into it (Wanted's double_return); CxLVAL and the C library's assertions
// are as in code compiled against perl's headers.
func TestNativeContextStores(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"gimme stored into the calling sub's context",
			`sub probe_gimme { PerlAPI::caller_gimme_list(); wantarray ? "list" : "scalar" } my $s = probe_gimme(); $s`, "list"},
		{"assertions are off", `PerlAPI::assertions_off()`, "1"},
	})
}

// The auxiliary data of a multideref op: an SV and a pad offset fill the
// low half of an item; PL_multideref_pc takes the address of an item. A
// string buffer is given back with SvPV_free (autovivification's way of
// turning a scalar into a reference in place), and croak_no_modify is the
// read-only error.
func TestNativeMultiderefItems(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		// the function put in PL_ppaddr is the one of the ops made after
		// (perl's newOP copies PL_ppaddr[type]); the ops made before keep
		// the interpreter's
		{"ops made before and after the replacement",
			`my $before = sub { our %h_before = (a => 1); $h_before{a} }; PerlAPI::mderef_watch(); my $after = eval q{ sub { our %h_after = (b => 2); $h_after{b} } }; my $v = $before->(); my $n1 = PerlAPI::mderef_seen(); $v .= $after->(); "[$n1][" . PerlAPI::mderef_seen() . "]$v"`,
			"[][h_after:b:pc;]12"},
		{"package hash", `PerlAPI::mderef_watch(); my $v = eval q{ our %pkg_h = (key => 7); $pkg_h{key} }; PerlAPI::mderef_seen() . $v`, "pkg_h:key:pc;7"},
		{"lexical hash", `PerlAPI::mderef_watch(); my $v = eval q{ my %lex_h = (other => 8); $lex_h{other} }; PerlAPI::mderef_seen() . $v`, "pad:other:pc;8"},
		{"a string turned into a reference", `my $s = "text"; PerlAPI::vivify_into($s); ref($s) . scalar(@$s)`, "ARRAY0"},
		// the string shares the constant's buffer, which stays the
		// constant's (freed, it was given to other strings while the
		// constant still had it, and freed again with the constant)
		{"a string sharing a constant's buffer", `my $bad = 0; for (1 .. 20) { my $s = "text"; $bad++ if $s ne "text"; PerlAPI::vivify_into($s); my @fill = map { "q" x 5 } 1 .. 8 } $bad`, "0"},
		{"a read-only value", `eval { PerlAPI::vivify_into("const") }; $@ =~ /^Modification of a read-only value attempted/ ? "refused" : "other: $@"`, "refused"},
	})
}

// A compile error queued by native code fails the compilation with its
// message; a keyword plugin put in PL_keyword_plugin directly (the way of
// modules older than wrap_keyword_plugin) gets the words, and the words it
// declines are compiled as before.
func TestNativeParserPlugins(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"keyword plugin in PL_keyword_plugin", `PerlAPI::install_kw42(); eval q{ goperl_kw42 + 1 }`, "43"},
		{"queued compile error", `PerlAPI::install_kw42(); my $r = eval q{ goperl_kwerr; 1 }; ($r // "undef") . ":" . ($@ =~ /^queued by a keyword plugin at / ? "message" : "none: $@")`, "undef:message"},
		{"declined words", `PerlAPI::install_kw42(); eval q{ my $x = 5; $x * 2 }`, "10"},
	})
}

// Perl's public functions have their long names too (Perl_av_push(aTHX_
// ...)), and utf8.h's invariant macros are there.
func TestNativeLongNames(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"long names", `my @a = (1); PerlAPI::long_names(\@a, "12") . ":" . $a[1]`, "1,2,101:12"},
		{"not a number", `my @a; PerlAPI::long_names(\@a, "x")`, "0,1,101"},
		// a module's own copy of a function of the API, under its long name
		{"a module's own definition", `PerlAPI::own_long_name()`, "4242"},
	})
}

// The signal names and numbers native code sees are the system's; what
// native code stores into PL_statcache is what -X _ answers from.
func TestNativeSignalNamesAndStatCache(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	f := filepath.Join(t.TempDir(), "sized")
	if err := os.WriteFile(f, []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}
	runEvalCases(t, p, []evalCase{
		{"signal names", `PerlAPI::signal_names()`, "USR1,1"},
		{"the size the last stat found", `stat "` + f + `"; PerlAPI::statcache(undef)`, "5"},
		{"a size stored by native code", `stat "` + f + `"; PerlAPI::statcache(4242); -s _`, "4242"},
	})
}

// Fields perl's macros make lvalues (SvFLAGS, SvSTASH, GvFLAGS, HvRITER,
// HvEITER) are stored into, AvREAL is turned off and on, and an iterator
// position kept by one call is put back by another.
func TestNativeFieldLvalues(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"flags, stash, glob flags, AvREAL",
			`our $glob_for_flags; my $o = bless {}, "Some::Class"; my @a = (1); my $r = PerlAPI::field_lvalues($o, \*glob_for_flags, \@a); $r . ":" . ref($o)`,
			"1010010:HASH"},
		{"iterator kept and put back",
			`my %h = (a => 1, b => 2, c => 3); my ($k1) = each %h; my $at = PerlAPI::iter_save(\%h); my ($k2) = each %h; PerlAPI::iter_restore(\%h); my ($k3) = each %h; ($at eq $k1 ? "at" : "not at") . "," . ($k3 eq $k2 ? "same" : "other")`,
			"at,same"},
	})
}

// Interpreter variables native code reads and stores into (PL_argvoutgv,
// PL_signals, PL_defstash under save_hptr, the last stat, the stacks, the
// op mask under SAVEVPTR), the engine's exec with flags (CALLREGEXEC),
// sv_peek, despatch_signals and save_re_context.
func TestNativeInterpreterVariables(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	f := filepath.Join(t.TempDir(), "statted")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	runEvalCases(t, p, []evalCase{
		{"variables", `our $gl_vars; stat "` + f + `"; PerlAPI::interp_vars(\*gl_vars)`, "01111|0,1|" + f},
		{"op mask in a scope", `PerlAPI::op_mask_scope("1 + 1")`, "refused|1"},
		{"no op mask", `PerlAPI::op_mask_scope("1 + 1") . "," . (eval "2 + 2")`, "refused|1,4"},
		{"regex, peek, signals, match variables",
			`"keep" =~ /(e+)/; my $r = PerlAPI::regex_api(qr/b+/, "abbc"); "$r:$1"`, "10|IV(42)|0:ee"},
		{"recoded by an encoding object",
			`package Upper::Enc { sub decode { my ($self, $s) = @_; my $r = uc $s; utf8::upgrade($r); $r } } my $s = "abc"; PerlAPI::recode($s, bless {}, "Upper::Enc")`,
			"ABC:1"},
	})
}

// The descriptions and classes of ops, the Latin-1 classes and case tables,
// formats given no argument list, and assorted functions and macros of the
// API (bytes_cmp_utf8, mini_mktime, newSVhek, sortsv_flags, newAV_alloc_x,
// CvAUTOLOAD, ANY's members, xsubpp's INTERFACE).
func TestNativeMoreAPI(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"op descriptions and classes", `PerlAPI::op_info()`, "subroutine entry|2|1|1,1,1|private variable"},
		{"Latin-1 classes", `PerlAPI::chars_l1()`, "10101111|yQ|20,20|97|5.44.0"},
		{"formats without arguments",
			`use warnings; my @w; local $SIG{__WARN__} = sub { push @w, $_[0] }; my $r = PerlAPI::format_none(); $r . "|" . join "", map { /^(warned without arguments|Missing argument in subroutine entry)/ ? "<$1>" : "<other: $_>" } @w`,
			"no args, 100%/[|0]/form %|<Missing argument in subroutine entry><warned without arguments>"},
		{"assorted functions", `PerlAPI::misc_api("caf\xc3\xa9", ["b", "a", "c"])`, "0,-1|2024-03-02 01:6:61|cba|hk|10|7|42"},
	})
}
