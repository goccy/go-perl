//go:build darwin || linux

package gperl_test

import "testing"

// poll reports POLLHUP for a pipe whose writers have all closed it, and not
// while one is there, as perl's poll does: an event loop that waits for a
// hang-up (Danga::Socket's, for a client it has stopped reading) learns of
// it from poll.
func TestPollReportsAHangUp(t *testing.T) {
	out, status := hostPerl(t, "-e", `
use IO::Poll qw(POLLIN POLLHUP);
pipe(my $r, my $w) or die "pipe: $!";
syswrite($w, "x");
my @set = (fileno($r), POLLIN);
IO::Poll::_poll(1000, @set);
print "writer there: ", ($set[1] & POLLIN ? "readable" : "not readable"), ", ", ($set[1] & POLLHUP ? "hung up" : "not hung up"), "\n";
close $w;
@set = (fileno($r), POLLIN);
IO::Poll::_poll(1000, @set);
print "writer gone: ", ($set[1] & POLLHUP ? "hung up" : "not hung up"), "\n";
`)
	want := "writer there: readable, not hung up\nwriter gone: hung up\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}
