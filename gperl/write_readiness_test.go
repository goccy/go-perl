//go:build darwin || linux

package gperl_test

import (
	"path/filepath"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// select and poll report a socket or pipe writable only while it can take
// data; a write takes what fits, waits or fails as the descriptor's mode
// says, can be interrupted, and reports a reader that is gone.
func TestWriteReadinessAndStreamWrites(t *testing.T) {
	got := runForkScript(t, `
use Socket; use Fcntl; use IO::Handle; use IO::Poll qw(POLLIN POLLOUT);
sub writable {
    my ($fh) = @_;
    my $w = '';
    vec($w, fileno($fh), 1) = 1;
    my $n = select(undef, $w, undef, 0.05);
    return $n > 0 && vec($w, fileno($fh), 1) ? "writable" : "not writable";
}
sub nonblocking { my ($fh) = @_; my $fl = fcntl($fh, F_GETFL, 0) // die "F_GETFL: $!"; fcntl($fh, F_SETFL, $fl | O_NONBLOCK) or die "F_SETFL: $!" }
sub blocking { my ($fh) = @_; my $fl = fcntl($fh, F_GETFL, 0) // die "F_GETFL: $!"; fcntl($fh, F_SETFL, $fl & ~O_NONBLOCK) or die "F_SETFL: $!" }
sub fill {
    my ($fh) = @_;
    nonblocking($fh);
    my ($total, $chunk) = (0, "x" x 65536);
    while (1) {
        my $n = syswrite($fh, $chunk);
        if (!defined $n) { return ($total, $!{EAGAIN} ? "EAGAIN" : "error $!") }
        $total += $n;
        die "the stream never fills up" if $total > 1 << 30;
    }
}

socket(my $listener, AF_INET, SOCK_STREAM, 0) or die "socket: $!";
bind($listener, pack_sockaddr_in(0, INADDR_LOOPBACK)) or die "bind: $!";
listen($listener, 5) or die "listen: $!";
report("a listening socket: ", writable($listener));
report("poll of a listening socket for output: ", IO::Poll::_poll(50, fileno($listener), my $mask = POLLOUT));
socket(my $client, AF_INET, SOCK_STREAM, 0) or die "socket: $!";
connect($client, getsockname($listener)) or die "connect: $!";
accept(my $server, $listener) or die "accept: $!";
report("a connection: ", writable($client), ", ", writable($server));

pipe(my $r, my $w) or die "pipe: $!";
report("the write end of a pipe: ", writable($w));
report("the read end of a pipe: ", writable($r));
my ($filled, $why) = fill($w);
report("a non-blocking write into a pipe: ", ($filled > 0 ? "took what fits" : "took nothing"), ", then $why");
report("the full pipe: ", writable($w));
sysread($r, my $drain, $filled) == $filled or die "drain: $!";
report("after the reader drained it: ", writable($w));

my ($sent, $sent_why) = fill($client);
report("a non-blocking write into a socket: ", ($sent > 0 ? "took what fits" : "took nothing"), ", then $sent_why");
report("the full socket: ", writable($client));

# a blocking write waits for room; a signal handler ends the wait
fill($w);
blocking($w);
my $alarms = 0;
$SIG{ALRM} = sub { $alarms++ };
alarm 1;
my $n = syswrite($w, "x" x 65536);
alarm 0;
report("a blocking write into a full pipe: ", (defined $n ? "wrote $n" : $!{EINTR} ? "EINTR" : "error $!"), " after $alarms alarm");

# writing where nobody reads any more
for my $disposition ('handler', 'IGNORE') {
    pipe(my $pr, my $pw) or die "pipe: $!";
    close $pr;
    my $got = 0;
    local $SIG{PIPE} = $disposition eq 'handler' ? sub { $got++ } : 'IGNORE';
    my $wrote = syswrite($pw, "x");
    report("write to a pipe without a reader, SIGPIPE $disposition: ",
           (defined $wrote ? "wrote $wrote" : $!{EPIPE} ? "EPIPE" : "error $!"), " signals=$got");
}
my $pid = fork // die "fork: $!";
if (!$pid) {
    $SIG{PIPE} = 'DEFAULT';
    pipe(my $pr, my $pw) or die "pipe: $!";
    close $pr;
    syswrite($pw, "x");
    exit 0;
}
waitpid($pid, 0);
report("a child writing to a pipe without a reader, SIGPIPE default: signal=", $? & 127, " exit=", $? >> 8);
`)
	wantLines(t, got,
		"a listening socket: not writable",
		"poll of a listening socket for output: 0",
		"a connection: writable, writable",
		"the write end of a pipe: writable",
		"the read end of a pipe: not writable",
		"a non-blocking write into a pipe: took what fits, then EAGAIN",
		"the full pipe: not writable",
		"after the reader drained it: writable",
		"a non-blocking write into a socket: took what fits, then EAGAIN",
		"the full socket: not writable",
		"a blocking write into a full pipe: EINTR after 1 alarm",
		"write to a pipe without a reader, SIGPIPE handler: EPIPE signals=1",
		"write to a pipe without a reader, SIGPIPE IGNORE: EPIPE signals=0",
		"a child writing to a pipe without a reader, SIGPIPE default: signal=13 exit=0",
	)
}

// A gperl process that writes to a closed pipe dies of SIGPIPE, like any
// program at the head of a pipeline whose reader has gone.
func TestWritingToAClosedPipeEndsAGperlProcess(t *testing.T) {
	shim := perlCommandShim(t)
	script := writeScript(t, "sigpipe.pl", `
use strict; use warnings;
my ($report, $perl) = @ARGV;
sub report { open(my $fh, '>>', $report) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
open(my $out, '-|', $perl, '-e', '$| = 1; print "line $_\n" for 1 .. 200000; print STDERR "survived\n"') // die "spawn: $!";
my $first = <$out>;
chomp $first;
close $out;
report("read '$first' and closed the pipe: the writer ended with signal=", $? & 127, " exit=", $? >> 8);
`)
	report := filepath.Join(t.TempDir(), "report.txt")
	status, err := gperl.Run(script, []string{report, shim})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	wantLines(t, readLines(t, report),
		"read 'line 1' and closed the pipe: the writer ended with signal=13 exit=0",
	)
}
