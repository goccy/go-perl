//go:build darwin || linux

package gperl_test

import (
	"strings"
	"testing"
)

// A forked process's new descriptor gets the lowest number free for it, as
// a child process's does, also where its parent still has a descriptor
// under that number. Proc::Daemon counts on it: the daemon it forks closes
// every descriptor under perl's handles and opens /dev/null until the
// numbers are open again, so that perl closes those handles without an
// error when the daemon exits.
func TestForkedProcessGetsTheLowestFreeNumber(t *testing.T) {
	cmd := perlProcess("-MPOSIX=", "-MFile::Temp=tempdir", "-e", `
$| = 1;
my $dir = tempdir(CLEANUP => 1);
open(my $a, "<", "/dev/null") or die;
open(my $b, "<", "/dev/null") or die;
my $pid = fork // die "fork: $!";
if (!$pid) {
    my $n = fileno $b;
    POSIX::close($n);
    my $fd = POSIX::open("/dev/null", POSIX::O_RDONLY());
    print "child: the closed number is given out again: ", ($fd == $n ? "yes" : "no, $fd for $n"), "\n";
    POSIX::close($fd);
    my $dup = POSIX::dup(0);
    print "child: so is it by dup: ", ($dup == $n ? "yes" : "no, $dup for $n"), "\n";
    POSIX::close($dup);
    my $above = fcntl(STDIN, POSIX::F_DUPFD(), $n + 3);
    print "child: F_DUPFD gives at least its minimum: ", ($above >= $n + 3 ? "yes" : "no, $above"), "\n";
    POSIX::_exit(0);
}
waitpid $pid, 0;
print "parent: its own descriptor is still open: ", (defined(fileno $b) && sysseek($b, 0, 0) ? "yes" : "no"), "\n";

# The daemon of Proc::Daemon, which execs a command and exits.
$pid = fork // die "fork: $!";
if (!$pid) {
    my $highest;
    for (0 .. 20) {
        if    ($_ == 0) { close STDIN }
        elsif ($_ == 1) { close STDOUT }
        elsif ($_ == 2) { close STDERR }
        else            { $highest = $_ if POSIX::close($_) }
    }
    open(STDIN,  "</dev/null");
    open(STDOUT, ">", "$dir/out");
    open(STDERR, ">", "$dir/err");
    if ($highest) {
        my @fh;
        foreach (3 .. $highest) { open $fh[$_], "</dev/null" }
    }
    my $child = fork // die "fork: $!";
    if (!$child) { exec "sh", "-c", "echo command >&2" or die "exec: $!" }
    waitpid $child, 0;
    exit 0;
}
waitpid $pid, 0;
open(my $err, "<", "$dir/err") or die "open: $!";
print "daemon's standard error: ", <$err>;
`)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("perl: %v, stderr %s", err, stderr.String())
	}
	want := "child: the closed number is given out again: yes\n" +
		"child: so is it by dup: yes\n" +
		"child: F_DUPFD gives at least its minimum: yes\n" +
		"parent: its own descriptor is still open: yes\n" +
		"daemon's standard error: command\n"
	if string(out) != want {
		t.Errorf("output:\n%s\nwant:\n%s\nstderr: %s", out, want, stderr.String())
	}
}
