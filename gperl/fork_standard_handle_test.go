//go:build darwin || linux

package gperl_test

import "testing"

// The child of open(STDERR, "|-") has its STDERR, as on perl: a standard
// handle keeps its file while it is opened anew, and the child's open,
// which returns 0, gives it back. A filter the program forks for its own
// standard error (App::optex) writes what it read there.
func TestChildOfAStandardHandlePipeOpenKeepsTheHandle(t *testing.T) {
	out, status := hostPerl(t, "-e", `
$| = 1;
my $pid = open(STDERR, '|-') // die "fork: $!";
if ($pid == 0) {
    print "child: STDERR is ", (defined fileno(STDERR) ? "open on " . fileno(STDERR) : "closed"), "\n";
    open(my $dup, '>&', \*STDERR) or die "dup STDERR: $!";
    print "child: STDERR duplicated\n";
    while (my $line = <STDIN>) { print "child read: $line" }
    exit 0;
}
print STDERR "through the pipe\n";
close STDERR;
print "parent: the child ended with ", $? >> 8, "\n";
`)
	want := "child: STDERR is open on 2\nchild: STDERR duplicated\nchild read: through the pipe\nparent: the child ended with 0\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
