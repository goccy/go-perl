//go:build darwin || linux

package xs_test

import "testing"

// Names of the perl API that distributions use and the SDK lacked:
// scan_vstring (Math::GMP, Sereal), grok_atoUV, is_utf8_invariant_string_loc,
// isinfnan, Perl_error_log (STDERR as the program has it now, which
// Test::LeakTrace's tests capture), do_sv_dump into a handle (which wrote
// a placeholder line), TARGi, sv_inc (which only counted: "aa" became 1,
// not "ab") and sv_dec, and names that only need to compile (Direntry_t,
// SVppv_STATIC, CvISXSUB_on, PERL_FLUSHALL_FOR_CHILD, the pointer table
// types Test::LeakTrace keeps its table in). And a
// MAGIC of a module's own linked onto an empty chain through SvMAGIC, as
// Event attaches its watchers, which mg_find finds in the next call. Before,
// the names did not compile, and the MAGIC was not found ("sv_2thing: can't
// decode SV" in every test of Event).
func TestNativeSmallNamesOfThePerlAPI(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"scan_vstring", `my ($v, $n) = PerlAPI::vstring_of("1.22.333"); join ",", ref(\$v), join(".", map ord, split //, $v), $n`, "VSTRING,1.22.333,8"},
		{"scan_vstring stops at the end of the v-string", `my ($v, $n) = PerlAPI::vstring_of("v5.300x"); join ",", join(".", map ord, split //, $v), $n`, "5.300,6"},
		{"grok_atoUV", `join "|", map { PerlAPI::small_names("atouv", $_) } "123", "0", "01", "18446744073709551615", "18446744073709551616", "", "-1"`, "ok 123|ok 0|fail|ok 18446744073709551615|fail|fail|fail"},
		{"grok_atoUV with an end", `PerlAPI::small_names("atouv_end", "12ab")`, "ok 12 ab"},
		{"is_utf8_invariant_string_loc", `join "|", PerlAPI::small_names("invariant", "abc"), PerlAPI::small_names("invariant", "ab\xC3\xA9")`, "yes|no at 2"},
		{"isinfnan", `my $inf = 9**9**9; join "|", map { PerlAPI::small_names("infnan", $_) } $inf, -$inf, $inf - $inf, 1.5`, "infnan|infnan|infnan|finite"},
		{"Perl_error_log", `my $buf = ""; { open local(*STDERR), '>', \$buf or die "open: $!"; PerlAPI::small_names("error_log", 0); } $buf`, "to the error log\n"},
		{"toLOWER_utf8_safe in place", `my $s = "ABC\x{C9}D Search"; utf8::upgrade($s); PerlAPI::small_names("lower_in_place", $s) eq "abc\x{E9}d search" ? "lowered" : "got " . PerlAPI::small_names("lower_in_place", $s)`, "lowered"},
		{"do_sv_dump into a handle", `my $buf = ""; { open local(*STDERR), '>', \$buf or die "open: $!"; PerlAPI::small_names("dump", 42); } $buf =~ /\ASV = IV\(0x[0-9a-f]+\) at 0x[0-9a-f]+\n.*\n  IV = 42\n\z/s ? "dumped" : $buf`, "dumped"},
		{"sv_inc", `join "|", map { PerlAPI::small_names("inc", $_) } "aa", "Az", "zz", "a9", 9, -1, 1.5, 9223372036854775807, 18446744073709551615`, "ab|Ba|aaa|b0|10|0|2.5|9223372036854775808|1.84467440737096e+19"},
		{"sv_dec", `join "|", map { PerlAPI::small_names("dec", $_) } 1, "aa", 0.5`, "0|-1|-0.5"},
		{"cv_name", `sub Some::Pkg::named { 1 } my $anon = sub { 2 }; join ";", PerlAPI::small_names("cvname", \&Some::Pkg::named), PerlAPI::small_names("cvname", $anon)`, "Some::Pkg::named|named;main::__ANON__|__ANON__"},
		{"PERL_SHORT_MIN and the other limits", `PerlAPI::small_names("limits", 0)`, "-32768 32767 65535 255 1"},
		{"STATUS_EXIT", `$? = 3 << 8; PerlAPI::small_names("status", 0)`, "768"},
		{"PerlIO_openn with :scalar", `PerlAPI::small_names("openn", "the body")`, "the body"},
		{"PerlIO_isutf8", `open my $u, "<:utf8", \ "x" or die; open my $b, "<", \ "x" or die; join ",", PerlAPI::small_names("isutf8", $u), PerlAPI::small_names("isutf8", $b)`, "utf8,bytes"},
		{"the words of an IO", `PerlAPI::small_names("iowords", "Some::Filter")`, "Some::Filter,Some::Filter,1,7,7"},
		{"sv_regex_global_pos_*", `my ($s, $t) = ("abcdef", "abcdef"); pos($t) = 4; join ";", PerlAPI::small_names("pos", $s), PerlAPI::small_names("pos", $t), defined(pos($t)) ? "pos" : "no pos"`, "0,1,2,0;1,1,2,0;no pos"},
		{"PL_hash_seed", `require Hash::Util; my $seed = PerlAPI::small_names("seed", 0); length($seed) . "," . ($seed eq Hash::Util::hash_seed() ? "the interpreter's" : "another")`, "32,the interpreter's"},
		{"Perl_hv_common", `PerlAPI::small_names("hvcommon", "stored")`, "stored"},
		{"LVRET and is_lvalue_sub", `PerlAPI::small_names("lvret", 0)`, "0,0"},
		{"SSNEWt and SSPTRt", `PerlAPI::small_names("ssptrt", 4)`, "7"},
		{"VAL_O_NONBLOCK and VAL_EAGAIN", `PerlAPI::small_names("nonblock", 0)`, "same"},
		{"TARGi", `PerlAPI::small_names("targi", 21)`, "42"},
		{"names that compile", `PerlAPI::small_names("compile", 0)`, "compiled"},
		{"a MAGIC linked onto an empty chain", `my $t = PerlAPI::event_like_thing(2); PerlAPI::event_like_ptr($t)`, "33"},
	})
}
