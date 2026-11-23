package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// Devel::PPPort is part of perl's library: build tools require it, and its
// one job is to write ppport.h.
func TestDevelPPPortWritesItsHeader(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Devel::PPPort;`,
		`print "version: ", (Devel::PPPort->VERSION >= 3.68 ? "current" : "old"), "\n";`,
		`Devel::PPPort::WriteFile("/ppport.h") or die "WriteFile: $!";`,
		`open(my $h, "<", "/ppport.h") or die "open: $!";`,
		`my $text = do { local $/; <$h> };`,
		`print "guard: ", ($text =~ /^#ifndef _P_P_PORTABILITY_H_$/m ? "found" : "missing"), "\n";`,
		`print "names the file: ", ($text =~ /ppport\.h -- Perl\/Pollution\/Portability/ ? "yes" : "no"), "\n";`,
		`print "size: ", (length($text) > 100_000 ? "whole" : "short"), "\n";`,
	)}, "")
	r.want(t, 0, "version: current\nguard: found\nnames the file: yes\nsize: whole\n")
}
