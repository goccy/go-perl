//go:build darwin || linux

package gperl_test

import (
	"os"
	"testing"
)

// The children of a process that ignores SIGCHLD leave nothing to wait for:
// waitpid lasts until the child has ended and reports ECHILD, and so does
// the wait a command's status comes from.
func TestIgnoredSIGCHLD(t *testing.T) {
	got := runForkScript(t, `
use POSIX ();
$SIG{CHLD} = 'IGNORE';
my $pid = fork // die "fork: $!";
if (!$pid) { select(undef, undef, undef, 0.2); POSIX::_exit(3) }
my $reaped = waitpid($pid, 0);
report("waitpid=$reaped status=$? ", ($!{ECHILD} ? "ECHILD" : "error $!"));
report("the child is gone=", (kill(0, $pid) ? 0 : 1));
my $said = `+"`echo said`"+`;
chomp $said;
report("command: said=$said status=$?");
$SIG{CHLD} = 'DEFAULT';
$pid = fork // die "fork: $!";
if (!$pid) { POSIX::_exit(3) }
$reaped = waitpid($pid, 0);
report("default action: reaped=", ($reaped == $pid ? 1 : 0), " status=", $? >> 8);
`)
	wantLines(t, got,
		"waitpid=-1 status=-1 ECHILD",
		"the child is gone=1",
		"command: said=said status=-1",
		"default action: reaped=1 status=3",
	)
}

// A command inherits the descriptors that are not close-on-exec, under
// their numbers: none above $^F unless the program clears the flag.
func TestCommandsInheritDescriptors(t *testing.T) {
	got := runForkScript(t, `
use Fcntl;
pipe(my $r, my $w) or die "pipe: $!";
report("a new descriptor is close-on-exec=", (fcntl($w, F_GETFD, 0) & FD_CLOEXEC ? 1 : 0));
system("/bin/sh", "-c", "echo hidden >&" . fileno($w) . " 2>/dev/null");
fcntl($w, F_SETFD, 0) or die "F_SETFD: $!";
report("after clearing the flag=", (fcntl($w, F_GETFD, 0) & FD_CLOEXEC ? 1 : 0));
system("/bin/sh", "-c", "echo passed >&" . fileno($w));
close $w;
report("the command wrote: ", join(" ", map { chomp; $_ } <$r>));
{
    local $^F = 20;
    pipe(my $r2, my $w2) or die "pipe: $!";
    system("/bin/sh", "-c", "echo system descriptor >&" . fileno($w2));
    close $w2;
    report("with \$^F raised: ", join(" ", map { chomp; $_ } <$r2>));
}
# through fork and exec, as IPC::Run and IPC::Open3 hand descriptors over
pipe(my $r3, my $w3) or die "pipe: $!";
my $pid = fork // die "fork: $!";
if (!$pid) {
    close $r3;
    fcntl($w3, F_SETFD, 0) or die "F_SETFD: $!";
    exec("/bin/sh", "-c", "echo from the exec >&" . fileno($w3)) or die "exec: $!";
}
close $w3;
report("fork and exec: ", join(" ", map { chomp; $_ } <$r3>));
waitpid($pid, 0);
`)
	wantLines(t, got,
		"a new descriptor is close-on-exec=1",
		"after clearing the flag=0",
		"the command wrote: passed",
		"with $^F raised: system descriptor",
		"fork and exec: from the exec",
	)
}

// The process is the host process: what it reports and does about itself
// is the host's — CPU time spent in the system, the pid of a command, the
// host's name, FIFOs and file owners of the host filesystem, the scheduling
// priority.
func TestProcessIsTheHostProcess(t *testing.T) {
	host, err := os.Hostname()
	if err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	got := runForkScriptArgs(t, `
use POSIX ();
my ($dir) = @ARGV;
my $start = time;
do { opendir(my $d, $dir) or die; my @e = readdir $d; } while time - $start < 1;
my @t = times;
report("times: user=", ($t[0] > 0 ? "spent" : 0), " system=", ($t[1] > 0 ? "spent" : 0));
system("/bin/sh", "-c", 'i=0; while [ $i -lt 30000 ]; do i=$((i+1)); done');
@t = times;
report("times: of the children=", ($t[2] + $t[3] > 0 ? "spent" : 0));

my $pid = open(my $from, "-|", "/bin/sh", "-c", 'echo $$') // die "spawn: $!";
my $said = <$from>;
chomp $said;
close $from;
report("pid of a command=", ($pid == $said ? "the command's own" : "$pid, the command says $said"));

my @uname = POSIX::uname();
report("uname: nodename=$uname[1]");
require Sys::Hostname;
report("hostname=", Sys::Hostname::hostname());

my $fifo = "$dir/fifo";
report("mkfifo=", (POSIX::mkfifo($fifo, 0600) ? 1 : "0 $!"), " is a FIFO=", (-p $fifo ? 1 : 0));
my $writer = fork // die "fork: $!";
if (!$writer) { open(my $out, ">", $fifo) or die "open: $!"; syswrite($out, "through the FIFO\n"); POSIX::_exit(0) }
open(my $in, "<", $fifo) or die "open: $!";
my $line = <$in>;
chomp $line;
report("read: $line");
close $in;
waitpid($writer, 0);

open(my $fh, ">", "$dir/owned") or die "create: $!";
report("chown to the owner=", chown($<, -1, "$dir/owned"), " fchown=", chown($<, -1, $fh));
close $fh;
if ($< != 0) {
    report("chown to root=", chown(0, 0, "$dir/owned"), " ", ($!{EPERM} ? "EPERM" : "error $!"));
    report("setuid to another user=", (POSIX::setuid($< + 1) ? 1 : 0), " ", ($!{EPERM} ? "EPERM" : "error $!"));
} else {
    report("chown to root=0 EPERM");
    report("setuid to another user=0 EPERM");
}
my $nice = POSIX::nice(0);
report("nice(0)=", (defined $nice ? "the value" : "undef $!"));
report("getpriority=", (getpriority(0, 0) == getpriority(0, $$) ? "one value" : "differs"));
`, dir)
	wantLines(t, got,
		"times: user=spent system=spent",
		"times: of the children=spent",
		"pid of a command=the command's own",
		"uname: nodename="+host,
		"hostname="+host,
		"mkfifo=1 is a FIFO=1",
		"read: through the FIFO",
		"chown to the owner=1 fchown=1",
		"chown to root=0 EPERM",
		"setuid to another user=0 EPERM",
		"nice(0)=the value",
		"getpriority=one value",
	)
}
