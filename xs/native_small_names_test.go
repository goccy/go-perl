//go:build darwin || linux

package xs_test

import "testing"

// Names of the perl API that distributions use and the SDK lacked:
// scan_vstring (Math::GMP, Sereal), grok_atoUV, is_utf8_invariant_string_loc,
// isinfnan, Perl_error_log, TARGi, and names that only need to compile
// (Direntry_t, SVppv_STATIC, CvISXSUB_on, PERL_FLUSHALL_FOR_CHILD). And a
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
		{"Perl_error_log", `PerlAPI::small_names("error_log", 0)`, "stderr"},
		{"TARGi", `PerlAPI::small_names("targi", 21)`, "42"},
		{"names that compile", `PerlAPI::small_names("compile", 0)`, "compiled"},
		{"a MAGIC linked onto an empty chain", `my $t = PerlAPI::event_like_thing(2); PerlAPI::event_like_ptr($t)`, "33"},
	})
}
