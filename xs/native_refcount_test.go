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

// newXS without a name makes an anonymous sub: it is installed in no glob
// (nothing is redefined when a second one is made), the caller owns the one
// reference to it, and each has its own XSANY. A module builds its closures
// over native code that way.
func TestNativeAnonymousSubs(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"made, kept and called", `my $warned = ""; local $SIG{__WARN__} = sub { $warned .= $_[0] }; my @subs = PerlAPI::anonymous_subs(4); my @more = map { [ (1) x 8 ] } 1 .. 64; join(",", map { ref($_) . ":" . $_->() } @subs) . "|$warned|" . (defined &{""} ? "a sub without a name in a glob" : "no glob")`, "CODE:0,CODE:1,CODE:2,CODE:3||no glob"},
		{"one reference each", `require B; my @subs = PerlAPI::anonymous_subs(2); join ",", map { B::svref_2object($_)->REFCNT } @subs`, "1,1"},
	})
}
