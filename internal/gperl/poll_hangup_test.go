//go:build darwin || linux

package gperl_test

import (
	"runtime"
	"testing"
)

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

// On Linux poll reports POLLERR for the write end of a pipe whose reader has
// closed it, asked or not, with POLLOUT when that was asked (the write
// fails at once); and a socket whose peer is gone is hung up and, when asked,
// readable and writable. An event loop that watches the write end for a
// hang-up (IO::Async's on_hangup) learns of the reader's end from it.
func TestPollReportsTheEndOfTheReader(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the events of poll(2) on Linux")
	}
	out, status := hostPerl(t, "-e", `
use IO::Poll qw(POLLIN POLLOUT POLLHUP POLLERR POLLNVAL);
use Socket;
sub events { my $e = shift; join("|", grep { $e & eval $_ } qw(POLLIN POLLOUT POLLHUP POLLERR POLLNVAL)) || "none" }
sub poll1 { my ($fh, $asked) = @_; my @set = (fileno($fh), $asked); my $n = IO::Poll::_poll(200, @set); "$n " . events($set[1]) }
pipe(my $r, my $w) or die "pipe: $!";
print "reader there: ", poll1($w, POLLHUP), "\n";
close $r;
print "reader gone, hang-up asked: ", poll1($w, POLLHUP), "\n";
print "reader gone, nothing asked: ", poll1($w, 0), "\n";
print "reader gone, writing asked: ", poll1($w, POLLOUT), "\n";
socketpair(my $a, my $b, AF_UNIX, SOCK_STREAM, PF_UNSPEC) or die "socketpair: $!";
close $b;
print "peer gone, reading and writing asked: ", poll1($a, POLLIN|POLLOUT), "\n";
`)
	want := "reader there: 0 none\n" +
		"reader gone, hang-up asked: 1 POLLERR\n" +
		"reader gone, nothing asked: 1 POLLERR\n" +
		"reader gone, writing asked: 1 POLLOUT|POLLERR\n" +
		"peer gone, reading and writing asked: 1 POLLIN|POLLOUT|POLLHUP\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}

// select leaves the time it did not wait in its timeout, as select(2) does
// on Linux, and perl returns it as the time left: none once the whole
// timeout has passed, some when a descriptor ended the wait first.
func TestSelectReportsTheTimeLeft(t *testing.T) {
	out, status := hostPerl(t, "-e", `
my ($nfound, $left) = select(undef, undef, undef, 0.2);
print "timeout: $nfound found, ", ($left == 0 ? "none left" : "$left left"), "\n";
pipe(my $r, my $w) or die "pipe: $!";
syswrite($w, "x");
my $bits = '';
vec($bits, fileno($r), 1) = 1;
($nfound, $left) = select(my $rout = $bits, undef, undef, 5);
print "ready: $nfound found, ", ($left > 4 && $left <= 5 ? "most of it left" : "$left left"), "\n";
`)
	want := "timeout: 0 found, none left\nready: 1 found, most of it left\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
