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
