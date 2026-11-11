//go:build darwin || linux

package xs_test

import "testing"

// SvREFCNT_inc and SvREFCNT_dec of a pointer to a structure of the module's
// own count in that structure, where perl counts (the word an SV head keeps
// its count in), and leave the interpreter alone: the pointer is not one of
// its values.
func TestNativeRefcountOfAModulesOwnStructure(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"counted in the structure", `my ($a, $b) = ("x", "y"); my @r = PerlAPI::count_own_struct($a, $b); join ",", @r[0, 1], ($r[2] == Internals::SvREFCNT($a) ? "same" : "changed"), ($r[3] == Internals::SvREFCNT($b) ? "same" : "changed")`, "1,0,same,same"},
	})
}
