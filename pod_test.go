package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// The standard library has the documentation files perl installs next to
// its modules, where programs look for them; the manual pages of perl's own
// pod directory are in the library's pod directory, where perl installs
// them, not beside the modules.
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
		"CORE.pod: found\nperlfaq.pod: found\nperlfunc.pod: missing\npod/perlfunc.pod: found\n")
}

// perl's manual pages are in the pod directory of the library, as perl
// installs them: perldiag, where the diagnostics pragma reads the
// explanations of perl's messages, perl.pod and the pages it names, and
// the pages perl's build generates (perlapi, perlmodlib, perltoc, ...).
func TestStandardLibraryShipsPerlsManualPages(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Config;`,
		`my ($lib) = $INC{"Config.pm"} =~ m{^(.*)/Config\.pm\z}s;`,
		`my $page = "$lib/pod/perldiag.pod";`,
		`print "pod/perldiag.pod: ", (-f $page ? "found" : "missing"), "\n";`,
		`open(my $pod, "<", $page) or die "open: $!";`,
		`print "it explains a message: ", (grep(/^=item Can't locate %s/, <$pod>) ? "yes" : "no"), "\n";`,
		`for my $name (qw(perl perlfunc perlvar perlre perlapi perlintern perlmodlib perltoc perluniprops perl5440delta perllinux)) {`,
		`    print "$name: ", (-f "$lib/pod/$name.pod" ? "found" : "missing"), "\n";`,
		`}`,
	)}, "")
	r.want(t, 0, "pod/perldiag.pod: found\nit explains a message: yes\n"+
		"perl: found\nperlfunc: found\nperlvar: found\nperlre: found\nperlapi: found\nperlintern: found\n"+
		"perlmodlib: found\nperltoc: found\nperluniprops: found\nperl5440delta: found\nperllinux: found\n")
}
