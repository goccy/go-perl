//go:build linux

package gperl_test

import "testing"

// A forked process that makes a process group of its own (setpgrp) and
// execs a command is replaced by a command that leads that group: a signal
// to the group (kill -9 => $pid, or SIGTERM, as a timeout wrapper sends
// them) reaches the command and what the command started, as on perl. So
// it does when the command is found on PATH, after the directories before
// the one that has it.
func TestExecdCommandLeadsTheGroupOfTheProcess(t *testing.T) {
	for _, tc := range []struct{ shell, signal string }{{"/bin/sh", "-9 => "}, {"sh", "-9 => "}, {"sh", "'TERM', -"}} {
		t.Run(tc.shell+" "+tc.signal, func(t *testing.T) { execdCommandLeadsTheGroup(t, tc.shell, tc.signal) })
	}
}

func execdCommandLeadsTheGroup(t *testing.T, shell, signal string) {
	out, status := hostPerl(t, "-MFile::Temp=tempdir", "-e", `
my $dir = tempdir(CLEANUP => 1);
my $file = "$dir/pids";
my $pid = fork // die "fork: $!";
if (!$pid) {
    setpgrp(0, 0);
    exec "`+shell+`", "-c", "sleep 30 & echo \$\$ \$! > $file.tmp; mv $file.tmp $file; wait";
}
select(undef, undef, undef, 0.05) until -s $file;
open(my $fh, '<', $file) or die "open: $!";
my ($sh, $sleep) = split ' ', scalar <$fh>;
kill `+signal+`$pid;
waitpid $pid, 0;
print "the process: ", ($? & 127) ? "killed" : "status $?", "\n";
# a process is gone when it is no more, or a zombie nobody reaps
sub gone {
    my $p = shift;
    for (1 .. 50) {
        open(my $st, '<', "/proc/$p/stat") or return 1;
        my @f = split ' ', scalar <$st>;
        return 1 if $f[2] eq 'Z';
        select(undef, undef, undef, 0.05);
    }
    0;
}
print "the command: ", (gone($sh) ? "gone" : "running"), "\n";
print "what it started: ", (gone($sleep) ? "gone" : "running"), "\n";
`)
	want := "the process: killed\nthe command: gone\nwhat it started: gone\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
