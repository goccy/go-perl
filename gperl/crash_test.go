package gperl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// A program whose interpreter crashes has not succeeded, and nothing of it
// runs afterwards: a crashed process runs no END blocks and no destructors.
func TestCrashEndsTheRunAbnormally(t *testing.T) {
	report := filepath.Join(t.TempDir(), "report.txt")
	// The C library's abort is a trap of the guest: the call into it does
	// not return.
	script := writeScript(t, "crash.pl", `
use POSIX ();
my $report = shift;
sub note { open(my $fh, '>>', $report) or die; print {$fh} "$_[0]\n"; close $fh }
package Guard { sub new { bless {}, shift } sub DESTROY { main::note("destructor ran") } }
our $guard = Guard->new;
END { note("END block ran") }
note("before the crash");
POSIX::abort();
note("after the crash");
`)
	status, err := gperl.Run(script, []string{report})
	if err == nil || status == 0 {
		t.Fatalf("Run = (%d, %v), want a failure status and an error", status, err)
	}
	raw, readErr := os.ReadFile(report)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if got, want := string(raw), "before the crash\n"; got != want {
		t.Fatalf("the program reported %q, want %q", got, want)
	}
}
