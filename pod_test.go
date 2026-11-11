package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// The standard library has the documentation files perl installs next to
// its modules, where programs look for them; the manual pages of perl's own
// pod directory are not part of the library.
func TestStandardLibraryShipsItsPodFiles(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Config;`,
		`(my $beside = $INC{"Config.pm"}) =~ s/\.pm\z/.pod/;`,
		`print "Config.pod beside Config.pm: ", (-f $beside ? "found" : "missing"), "\n";`,
		`open(my $pod, "<", $beside) or die "open: $!";`,
		`print "it documents Config: ", (grep(/^=head1 NAME/, <$pod>) ? "yes" : "no"), "\n";`,
		`my ($lib) = $INC{"Config.pm"} =~ m{^(.*)/Config\.pm\z}s;`,
		`for my $name (qw(CORE.pod perlfaq.pod perlfunc.pod pod/perlfunc.pod)) {`,
		`    print "$name: ", (-f "$lib/$name" ? "found" : "missing"), "\n";`,
		`}`,
	)}, "")
	r.want(t, 0, "Config.pod beside Config.pm: found\nit documents Config: yes\n"+
		"CORE.pod: found\nperlfaq.pod: found\nperlfunc.pod: missing\npod/perlfunc.pod: missing\n")
}
