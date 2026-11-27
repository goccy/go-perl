//go:build linux

package gperl_test

import "testing"

// /dev/fd, /proc/self/fd and /proc/$$/fd list the descriptors of the
// process that reads them, in a forked process too: not the descriptors of
// the host process, which are the runtime's and the other processes' as
// well, and under the pid of a forked process the host has nothing.
func TestOwnDescriptorsAreListed(t *testing.T) {
	const dirs = `"/dev/fd", "/proc/self/fd", "/proc/$$/fd"`
	out, status := hostPerl(t, "-e", `
$| = 1;
sub listing { my $d = shift; opendir(my $h, $d) or return "$d: $!";
              my @n = sort { $a <=> $b } grep /^\d+$/, readdir $h; my $own = fileno($h) // -1;
              join " ", grep { $_ != $own } @n }
sub expect { join " ", sort { $a <=> $b } @_ }
sub label { my $d = shift; $d =~ m{^/proc/\d} ? '/proc/$$/fd' : $d }
open(my $f, "<", "/dev/null") or die;
my @mine = (0, 1, 2, fileno $f);
for my $d (`+dirs+`) {
    my $l = listing($d);
    print label($d), ": ", ($l eq expect(@mine) ? "the process's" : "$l, want " . expect(@mine)), "\n";
}
my $pid = fork // die "fork: $!";
if (!$pid) {
    close $f;
    for my $d (`+dirs+`) {
        my $l = listing($d);
        print "child ", label($d), ": ", ($l eq "0 1 2" ? "the process's" : "$l, want 0 1 2"), "\n";
    }
    exit 0;
}
waitpid $pid, 0;
`)
	want := "/dev/fd: the process's\n/proc/self/fd: the process's\n/proc/$$/fd: the process's\n" +
		"child /dev/fd: the process's\nchild /proc/self/fd: the process's\nchild /proc/$$/fd: the process's\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}

// A command the program starts with its standard input closed has it
// closed, as a command started by perl has (system after close STDIN, or a
// forked process that closes it and execs): it does not get /dev/null.
// (The shell's probe is dash's: bash opens a descriptor of its own there.)
func TestCommandHasTheStandardInputTheProgramClosed(t *testing.T) {
	out, status := hostPerl(t, "-e", `
$| = 1;
my $probe = 'if (true 3<&0) 2>/dev/null; then echo open; else echo closed; fi';
my $pid = fork // die "fork: $!";
if (!$pid) { close STDIN; exec "/bin/sh", "-c", "echo forked and execd: \$($probe)" or die "exec: $!" }
waitpid $pid, 0;
open(my $saved, "<&", \*STDIN) or die "dup: $!";
close STDIN;
system "/bin/sh", "-c", "echo system: \$($probe)";
open(STDIN, "<&", $saved) or die "restore: $!";
system "/bin/sh", "-c", "echo after restoring it: \$($probe)";
`)
	want := "forked and execd: closed\nsystem: closed\nafter restoring it: open\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
