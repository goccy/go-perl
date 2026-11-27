//go:build darwin || linux

package gperl_test

import (
	"path/filepath"
	"testing"
)

// The file tests tell a character device, a FIFO or pipe, a socket, a
// directory and a regular file apart, by name and by handle, and a
// standard stream is what it is connected to.
func TestFileTypes(t *testing.T) {
	shim := perlCommandShim(t)
	script := writeScript(t, "types.pl", `
use strict; use warnings;
use Socket;
my ($report, $perl) = @ARGV;
sub report { open(my $fh, '>>', $report) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
my $dir = $report =~ s{/[^/]*$}{}r;
sub types {
    my ($what) = @_;
    return "missing" unless stat($what);
    return join("+", grep { defined } (-f _ ? "file" : undef), (-d _ ? "directory" : undef), (-c _ ? "character device" : undef),
                (-b _ ? "block device" : undef), (-p _ ? "pipe" : undef), (-S _ ? "socket" : undef)) || "no type";
}

report("/dev/null: ", types("/dev/null"));
open(my $null, '<', '/dev/null') or die "open: $!";
report("a handle on /dev/null: ", types($null));
open(my $made, '>', "$dir/plain") or die "create: $!";
report("a regular file and its handle: ", types("$dir/plain"), ", ", types($made));
report("a directory: ", types($dir));
pipe(my $r, my $w) or die "pipe: $!";
report("the ends of a pipe: ", types($r), ", ", types($w));
system("mkfifo", "$dir/fifo") == 0 or die "mkfifo failed";
report("a FIFO: ", types("$dir/fifo"));
socket(my $tcp, AF_INET, SOCK_STREAM, 0) or die "socket: $!";
report("a socket: ", types($tcp));
socketpair(my $a, my $b, AF_UNIX, SOCK_STREAM, 0) or die "socketpair: $!";
report("a socket pair: ", types($a), ", ", types($b));

# the standard streams of a perl process started with them redirected
my $show = 'sub t { stat($_[0]); -p _ ? "pipe" : -c _ ? "character device" : -f _ ? "file" : -S _ ? "socket" : "other" } print t(\*STDIN), ", ", t(\*STDOUT)';
my $said = qx{"$perl" -e '$show' < /dev/null};
report("stdin from /dev/null, stdout into a pipe: $said");
$said = qx{echo input | "$perl" -e '$show'};
report("stdin from a pipe, stdout into a pipe: $said");
system(qq{"$perl" -e '$show' < "$dir/plain" > "$dir/out"}) == 0 or die "child failed";
open(my $out, '<', "$dir/out") or die "open: $!";
report("stdin from a file, stdout into a file: ", scalar(<$out>));
`)
	report := filepath.Join(t.TempDir(), "report.txt")
	status, err := runInThisProcess(t, script, []string{report, shim})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	wantLines(t, readLines(t, report),
		"/dev/null: character device",
		"a handle on /dev/null: character device",
		"a regular file and its handle: file, file",
		"a directory: directory",
		"the ends of a pipe: pipe, pipe",
		"a FIFO: pipe",
		"a socket: socket",
		"a socket pair: socket, socket",
		"stdin from /dev/null, stdout into a pipe: character device, pipe",
		"stdin from a pipe, stdout into a pipe: pipe, pipe",
		"stdin from a file, stdout into a file: file, file",
	)
}
