package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// The Unicode directory of the library holds what an installed perl's does:
// the generated tables, and of the Unicode Character Database's source
// files only the three perl reads at run time. A program that finds another
// source file there takes it for an invitation to parse it.
func TestUnicoreHoldsWhatPerlInstalls(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Config;`,
		`my ($lib) = $INC{"Config.pm"} =~ m{^(.*)/Config\.pm\z}s;`,
		`for my $name (qw(version Blocks.txt SpecialCasing.txt NamedSequences.txt UCD.pl Name.pl Name.pm`,
		`    lib/Age/V11.pl To/Ea.pl EastAsianWidth.txt UnicodeData.txt auxiliary/WordBreakTest.txt`,
		`    extracted/DGeneralCategory.txt TestProp.pl mktables Makefile README.perl)) {`,
		`    print "$name: ", (-e "$lib/unicore/$name" ? "found" : "missing"), "\n";`,
		`}`,
		`my @txt; require File::Find;`,
		`File::Find::find(sub { push @txt, $_ if /\.txt\z/ }, "$lib/unicore");`,
		`print "txt files: ", join(",", sort @txt), "\n";`,
		`use Unicode::UCD qw(charinfo charblock casefold);`,
		`print "charinfo: ", charinfo(0x3042)->{name}, "\n";`,
		`print "charblock: ", charblock(0x3042), "\n";`,
		`print "casefold: ", casefold(0xDF)->{full}, "\n";`,
		`use charnames ();`,
		`print "viacode: ", charnames::viacode(0x263A), " ", sprintf("%04X", charnames::vianame("GREEK SMALL LETTER ALPHA")), "\n";`,
		`print "named sequence: ", length(charnames::string_vianame("LATIN SMALL LETTER A WITH MACRON AND GRAVE")), "\n";`,
		`print "properties: ", join("", map { "\x{3042}" =~ $_ ? 1 : 0 } qr/\p{Hiragana}/, qr/\p{Han}/, qr/\p{East_Asian_Width=Wide}/, qr/\p{Lu}/), "\n";`,
		`print "uc: ", (uc("\x{df}\x{100}") eq "SS\x{100}" ? "SS" : "other"), "\n";`,
	)}, "")
	r.want(t, 0, "version: found\nBlocks.txt: found\nSpecialCasing.txt: found\nNamedSequences.txt: found\n"+
		"UCD.pl: found\nName.pl: found\nName.pm: found\nlib/Age/V11.pl: found\nTo/Ea.pl: found\n"+
		"EastAsianWidth.txt: missing\nUnicodeData.txt: missing\nauxiliary/WordBreakTest.txt: missing\n"+
		"extracted/DGeneralCategory.txt: missing\nTestProp.pl: missing\nmktables: missing\nMakefile: missing\nREADME.perl: missing\n"+
		"txt files: Blocks.txt,NamedSequences.txt,SpecialCasing.txt\n"+
		"charinfo: HIRAGANA LETTER A\n"+
		"charblock: Hiragana\n"+
		"casefold: 0073 0073\n"+
		"viacode: WHITE SMILING FACE 03B1\n"+
		"named sequence: 2\n"+
		"properties: 1010\n"+
		"uc: SS\n")
}
