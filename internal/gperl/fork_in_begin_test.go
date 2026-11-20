//go:build darwin || linux

package gperl_test

import (
	"strings"
	"testing"
)

// A process forked inside BEGIN (or in the code of a module a use loads)
// is a pseudo-process that cannot go on compiling the file: its parent's
// compiler is not in it (perlfork's "BEGIN blocks"). A child that exits or
// execs inside the block works as on perl; one that falls off the end of
// the block dies saying why, with a failure status, instead of ending as if
// the program had run to its end.
func TestForkInsideBEGIN(t *testing.T) {
	cmd := perlProcess("-e", `
$| = 1;
BEGIN {
    my $pid = fork // die "fork: $!";
    if (!$pid) { print "first child: in BEGIN\n"; exit 3 }
    waitpid $pid, 0;
    print "first child ended with ", $? >> 8, "\n";
}
BEGIN {
    my $pid = fork // die "fork: $!";
    if (!$pid) { print "second child: in BEGIN\n" }
    else {
        waitpid $pid, 0;
        print "second child ended with ", ($? >> 8 ? "a failure" : "success"), "\n";
    }
}
print "the parent runs the program\n";
# a child forked when the program runs goes on to its end, and ends well
my $pid = fork // die "fork: $!";
if ($pid) {
    waitpid $pid, 0;
    print "a child forked at run time ended with ", $? >> 8, "\n";
}
`)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("perl: %v, stderr %s", err, stderr.String())
	}
	want := "first child: in BEGIN\nfirst child ended with 3\nsecond child: in BEGIN\nsecond child ended with a failure\n" +
		"the parent runs the program\na child forked at run time ended with 0\n"
	if string(out) != want {
		t.Errorf("output:\n%s\nwant:\n%s", out, want)
	}
	if !strings.Contains(stderr.String(), "fork inside BEGIN: the child process cannot compile the rest of the file") {
		t.Errorf("stderr does not say why the second child ended: %q", stderr.String())
	}
	if strings.Count(stderr.String(), "fork inside BEGIN") != 1 {
		t.Errorf("stderr: %q, want one message, from the second child", stderr.String())
	}
}
