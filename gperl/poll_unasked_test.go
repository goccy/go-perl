//go:build darwin || linux

package gperl_test

import (
	"runtime"
	"testing"
)

// poll takes a descriptor that asks for neither reading nor writing, and
// reports for it only the events poll reports unasked: nothing while its
// peer is there (another descriptor's input ends the wait), nothing for
// input nobody asked about (the wait runs out), and POLLHUP once its peer
// has closed it. Danga::Socket's poll loop (Perlbal) passes such a
// descriptor while it has stopped reading a client; the C library's poll
// refused it (ENOSYS: n=-1), and the loop spun.
//
// perl 5.44.0 on Linux prints all three lines below. macOS's poll(2) does
// not report the hang-up of a descriptor asked for POLLHUP alone (perl
// 5.34 there: "peer gone: n=0"), so the last line is checked on Linux only.
func TestPollTakesADescriptorAskedForNoEvents(t *testing.T) {
	out, status := hostPerl(t, "-e", `
use IO::Poll qw(POLLIN POLLHUP POLLERR POLLNVAL);
use Socket;
socketpair(my $a, my $b, AF_UNIX, SOCK_STREAM, 0) or die "socketpair: $!";
pipe(my $r, my $w) or die "pipe: $!";
syswrite($w, "x");
my @set = (fileno($a), POLLERR|POLLHUP|POLLNVAL, fileno($r), POLLIN);
my $n = IO::Poll::_poll(1000, @set);
print "beside input: n=$n socket=", ($set[1] & POLLHUP ? "hangup" : "nothing"), " pipe=", ($set[3] & POLLIN ? "readable" : "nothing"), "\n";
sysread($r, my $buf, 1);
syswrite($b, "unread");
@set = (fileno($a), POLLHUP, fileno($r), POLLIN);
$n = IO::Poll::_poll(200, @set);
print "input nobody asked about: n=$n\n";
close $b;
@set = (fileno($a), POLLHUP);
$n = IO::Poll::_poll(1000, @set);
print "peer gone: n=$n socket=", ($set[1] & POLLHUP ? "hangup" : "nothing"), "\n";
`)
	want := "beside input: n=1 socket=nothing pipe=readable\ninput nobody asked about: n=0\n"
	if runtime.GOOS == "linux" {
		want += "peer gone: n=1 socket=hangup\n"
	} else if len(out) > len(want) {
		out = out[:len(want)]
	}
	if status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}
