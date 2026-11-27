//go:build darwin || linux

package gperl_test

import "testing"

// A descriptor is open for reading, for writing or for both, and that is
// what F_GETFL says, what a read or write in the wrong direction answers,
// and what select waits for.
func TestDescriptorsHaveAnAccessMode(t *testing.T) {
	got := runForkScript(t, `
use Fcntl qw(F_GETFL O_ACCMODE O_RDONLY O_WRONLY O_RDWR);
use IO::Socket::INET;
sub mode {
    my ($fh) = @_;
    my $flags = fcntl($fh, F_GETFL, 0) // return "error $!";
    my $access = $flags & O_ACCMODE;
    return $access == O_RDONLY ? "O_RDONLY" : $access == O_WRONLY ? "O_WRONLY" : $access == O_RDWR ? "O_RDWR" : "mode $access";
}
my $dir = $REPORT =~ s{/[^/]*$}{}r;
open(my $out, '>', "$dir/file") or die "open: $!";
print {$out} "content\n"; close $out;
open(my $rd, '<', "$dir/file") or die "open: $!";
open(my $wr, '>>', "$dir/file") or die "open: $!";
open(my $rw, '+<', "$dir/file") or die "open: $!";
pipe(my $pr, my $pw) or die "pipe: $!";
my $listener = IO::Socket::INET->new(Listen => 1, LocalAddr => "127.0.0.1") or die "listen: $!";
report("files: ", join(" ", map { mode($_) } $rd, $wr, $rw));
report("pipe: ", mode($pr), " ", mode($pw));
report("socket: ", mode($listener));
report("standard streams: ", join(" ", map { mode($_) } \*STDIN, \*STDOUT, \*STDERR));

my $buf;
report("read of a write-only file: ", (defined sysread($wr, $buf, 4) ? "read" : ($!{EBADF} ? "EBADF" : "error $!")));
report("write to a read-only file: ", (defined syswrite($rd, "x") ? "written" : ($!{EBADF} ? "EBADF" : "error $!")));
report("read of a pipe's write end: ", (defined sysread($pw, $buf, 4) ? "read" : ($!{EBADF} ? "EBADF" : "error $!")));
report("write to a pipe's read end: ", (defined syswrite($pr, "x") ? "written" : ($!{EBADF} ? "EBADF" : "error $!")));

sub ready {
    my ($fh, $for_write) = @_;
    my $bits = ""; vec($bits, fileno($fh), 1) = 1;
    my $n = $for_write ? select(undef, $bits, undef, 0.1) : select($bits, undef, undef, 0.1);
    return $n ? "ready" : "not ready";
}
report("an empty pipe: write end for reading ", ready($pw, 0), ", for writing ", ready($pw, 1),
       "; read end for reading ", ready($pr, 0), ", for writing ", ready($pr, 1));
syswrite($pw, "x");
report("with input: read end for reading ", ready($pr, 0), ", write end for reading ", ready($pw, 0));
`)
	wantLines(t, got,
		"files: O_RDONLY O_WRONLY O_RDWR",
		"pipe: O_RDONLY O_WRONLY",
		"socket: O_RDWR",
		"standard streams: O_RDONLY O_WRONLY O_WRONLY",
		"read of a write-only file: EBADF",
		"write to a read-only file: EBADF",
		"read of a pipe's write end: EBADF",
		"write to a pipe's read end: EBADF",
		"an empty pipe: write end for reading not ready, for writing ready; read end for reading not ready, for writing not ready",
		"with input: read end for reading ready, write end for reading not ready",
	)
}
