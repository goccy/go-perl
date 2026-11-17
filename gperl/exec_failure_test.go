//go:build darwin || linux

package gperl_test

import (
	"strings"
	"testing"
)

// A command that cannot be started gets perl's warning, "Can't exec", when
// the exec warnings are on, from system (a string and a list) and a piped
// open alike, and system returns -1 with $! set: perl's forked child warns
// so when its exec fails.
func TestCommandThatCannotBeStartedWarns(t *testing.T) {
	const missing = "gperl-no-such-command"
	program := `
my $r = system("` + missing + `");
print "system: $r ", ($!{ENOENT} ? "ENOENT" : "other"), "\n";
$r = system("` + missing + `", "an argument");
print "system of a list: $r\n";
my $pid = open(my $fh, "-|", "` + missing + `");
print "piped open: ", (defined $pid ? "a pid" : "failed"), "\n";
`
	cmd := perlProcess("-we", "use Errno;"+program)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("perl: %v, stderr %s", err, stderr.String())
	}
	if want := "system: -1 ENOENT\nsystem of a list: -1\npiped open: failed\n"; string(out) != want {
		t.Errorf("output %q, want %q", out, want)
	}
	warning := `Can't exec "` + missing + `": No such file or directory at -e line `
	if n := strings.Count(stderr.String(), warning); n != 3 {
		t.Errorf("%d warnings %q in stderr, want 3:\n%s", n, warning, stderr.String())
	}
	// Without the warnings on, nothing is said.
	quiet := perlProcess("-e", "use Errno;"+program)
	var quietErr strings.Builder
	quiet.Stderr = &quietErr
	if _, err := quiet.Output(); err != nil || quietErr.Len() != 0 {
		t.Errorf("without warnings: err %v, stderr %q", err, quietErr.String())
	}
}

// A forked process's /dev/stdout is its own standard output: the child of
// open($fh, "-|") that writes to /dev/stdout writes into the pipe its
// parent reads, as on perl.
func TestForkedProcessOpensItsOwnStandardOutput(t *testing.T) {
	out, status := hostPerl(t, "-e", `
my $pid = open(my $in, "-|") // die "fork: $!";
if (!$pid) {
    open(my $o, ">", "/dev/stdout") or die "open /dev/stdout: $!";
    print $o "through /dev/stdout\n";
    close $o;
    exit 0;
}
my @got = <$in>;
close $in;
print "the parent read: @got";
`)
	if want := "the parent read: through /dev/stdout\n"; status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}
