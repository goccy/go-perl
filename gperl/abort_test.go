//go:build darwin || linux

package gperl_test

import (
	"path/filepath"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// POSIX::abort ends the process by SIGABRT.
func TestAbortRaisesSIGABRT(t *testing.T) {
	shim := perlCommandShim(t)
	report := filepath.Join(t.TempDir(), "report.txt")
	script := writeScript(t, "abort.pl", `
use strict; use warnings;
use POSIX ();
my ($report, $perl) = @ARGV;
sub report { open(my $fh, '>>', $report) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
my $pid = fork // die "fork: $!";
if (!$pid) { POSIX::abort() }
waitpid($pid, 0);
report("forked process: signal=", $? & 127, " exit=", $? >> 8);
system($perl, '-e', 'use POSIX (); END { print "END ran\n" } POSIX::abort()');
report("perl process: signal=", $? & 127, " exit=", $? >> 8);
my $out = `+"`$perl -e 'use POSIX (); POSIX::abort()'; echo \"shell status \\$?\"`"+`;
chomp $out;
report($out);
`)
	status, err := gperl.Run(script, []string{report, shim})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	wantLines(t, readLines(t, report),
		"forked process: signal=6 exit=0",
		"perl process: signal=6 exit=0",
		"shell status 134",
	)
}
