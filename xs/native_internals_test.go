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
		{"package hash", `PerlAPI::mderef_watch(); our %pkg_h = (key => 7); my $v = $pkg_h{key}; PerlAPI::mderef_seen() . $v`, "pkg_h:key:pc;7"},
		{"lexical hash", `PerlAPI::mderef_watch(); my %lex_h = (other => 8); my $v = $lex_h{other}; PerlAPI::mderef_seen() . $v`, "pad:other:pc;8"},
		{"a string turned into a reference", `my $s = "text"; PerlAPI::vivify_into($s); ref($s) . scalar(@$s)`, "ARRAY0"},
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
