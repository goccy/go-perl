package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// A module that is linked into the interpreter but was never loaded leaves
// no trace: its package has no sub, so nothing takes it for loaded. Loading
// it — through XSLoader or through DynaLoader's bootstrap — installs its
// boot function then, as a dynamically loading perl does.
func TestLinkedInExtensionsAreNotLoadedUntilUsed(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`sub subs_of { my $pkg = shift; no strict "refs"; scalar grep { defined &{"${pkg}::$_"} } keys %{"${pkg}::"} }`,
		`for my $m (qw(Digest::MD5 Digest::SHA Storable SDBM_File)) {`,
		`    print "$m before: subs=", subs_of($m), " bootstrap=", (defined &{"${m}::bootstrap"} ? 1 : 0), "\n";`,
		`}`,
		`require Digest::MD5;   # XSLoader::load`,
		`{ package Digest::SHA; require DynaLoader; our @ISA = ("DynaLoader"); Digest::SHA->bootstrap }`,
		`print "md5: ", Digest::MD5::md5_hex("abc"), "\n";`,
		`print "Digest::MD5 after: bootstrap=", (defined &Digest::MD5::bootstrap ? 1 : 0), "\n";`,
		`print "Digest::SHA after: bootstrap=", (defined &Digest::SHA::bootstrap ? 1 : 0),`,
		`    " sha1_hex=", (defined &Digest::SHA::sha1_hex ? 1 : 0), "\n";`,
		`print "Storable still: subs=", subs_of("Storable"), "\n";`,
		`print "not linked in: ", (eval { require XSLoader; XSLoader::load("No::Such::Extension"); 1 } ? "loaded" : "refused"), "\n";`,
	)}, "")
	r.want(t, 0, "Digest::MD5 before: subs=0 bootstrap=0\n"+
		"Digest::SHA before: subs=0 bootstrap=0\n"+
		"Storable before: subs=0 bootstrap=0\n"+
		"SDBM_File before: subs=0 bootstrap=0\n"+
		"md5: 900150983cd24fb0d6963f7d28e17f72\n"+
		"Digest::MD5 after: bootstrap=1\n"+
		"Digest::SHA after: bootstrap=1 sha1_hex=1\n"+
		"Storable still: subs=0\n"+
		"not linked in: refused\n")
}
