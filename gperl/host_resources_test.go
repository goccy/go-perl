package gperl_test

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"
)

// A datagram socket that is connected sends to its peer; what an address
// given to send means on it is the host's sendto to say, as it is for
// perl: Linux sends the datagram to that address, the BSDs refuse it
// (EISCONN). A module that keeps a connected socket and names the
// destination on every send (Net::SNMP) relies on the first.
func TestDatagramWithAnAddressOnAConnectedSocket(t *testing.T) {
	got := runForkScript(t, `
use IO::Socket::INET;
use IO::Select;
my $peer = IO::Socket::INET->new(Proto => 'udp', LocalAddr => '127.0.0.1', LocalPort => 0) or die "socket: $!";
my $other = IO::Socket::INET->new(Proto => 'udp', LocalAddr => '127.0.0.1', LocalPort => 0) or die "socket: $!";
my $client = IO::Socket::INET->new(Proto => 'udp', PeerAddr => '127.0.0.1', PeerPort => $peer->sockport) or die "socket: $!";
sub outcome { defined $_[0] ? "sent" : $!{EISCONN} ? "EISCONN" : "$!" }
sub arrived { my $sock = shift; return "nothing" unless IO::Select->new($sock)->can_read(0.5); $sock->recv(my $data, 100); "'$data'" }
report("without an address: ", outcome(send($client, "plain", 0)), ", the peer receives ", arrived($peer));
report("with the peer's address: ", outcome(send($client, "named", 0, $peer->sockname)), ", the peer receives ", arrived($peer));
report("with another address: ", outcome(send($client, "elsewhere", 0, $other->sockname)), ", the other socket receives ", arrived($other));
`)
	want := []string{
		"without an address: sent, the peer receives 'plain'",
		"with the peer's address: sent, the peer receives 'named'",
		"with another address: sent, the other socket receives 'elsewhere'",
	}
	if runtime.GOOS != "linux" {
		want = []string{
			"without an address: sent, the peer receives 'plain'",
			"with the peer's address: EISCONN, the peer receives nothing",
			"with another address: EISCONN, the other socket receives nothing",
		}
	}
	wantLines(t, got, want...)
}

// A command that has ended keeps its pid until the program has waited for
// it, as a process does: the host gives the number to no other process
// meanwhile, so the status the program collects later is that command's
// whatever else it ran in between. (On Linux, where a process that ended
// can be seen in /proc.)
func TestFinishedCommandKeepsItsPidUntilWaitedFor(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the state of a process that ended is read from /proc")
	}
	got := runForkScript(t, `
sub state { open(my $st, "<", "/proc/$_[0]/stat") or return "gone"; my $line = <$st>; $line =~ /\) (\S)/ ? $1 : "?" }
my $pid = open(my $out, "-|", "/bin/sh", "-c", "exit 3") // die "open: $!";
my @none = <$out>;                       # end of output: the command has ended
select(undef, undef, undef, 0.3);
report("ended, not waited for: state ", state($pid));
system("/bin/true") == 0 or die "true: $?" for 1 .. 20;
report("after other commands: state ", state($pid));
close($out);
report("waited for: exit status ", $? >> 8);
select(undef, undef, undef, 0.3);
report("after the wait: ", state($pid));
`)
	wantLines(t, got,
		"ended, not waited for: state Z",
		"after other commands: state Z",
		"waited for: exit status 3",
		"after the wait: gone",
	)
}

// Opening a FIFO waits for the other end, and a signal ends the wait: the
// open fails with EINTR and the handler runs. With the other end arriving,
// the open returns.
func TestOpeningAFIFOIsInterruptible(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("FIFOs are made on linux and darwin")
	}
	got := runForkScript(t, `
use POSIX qw(mkfifo);
use Time::HiRes qw(time alarm);
use File::Temp ();
my $dir = File::Temp::tempdir(CLEANUP => 1);
my $fifo = "$dir/fifo";
mkfifo($fifo, 0600) or die "mkfifo: $!";
for my $mode ("<", ">") {
    my $start = time;
    my $done = eval {
        local $SIG{ALRM} = sub { die "alarm\n" };
        alarm 0.3;
        my $opened = open(my $fh, $mode, $fifo);
        my $error = $opened ? "" : "$!";
        alarm 0;
        $opened ? "opened" : "failed: $error";
    };
    report("open $mode with nobody at the other end: ", ($done // "interrupted: $@" =~ s/\n//r), (time - $start < 5 ? "" : " (late)"));
}
my $writer = fork // die "fork: $!";
if ($writer == 0) {
    select(undef, undef, undef, 0.3);
    open(my $out, ">", $fifo) or die "open for writing: $!";
    print {$out} "through the fifo\n";
    close $out;
    exit 0;
}
my $line = eval {
    local $SIG{ALRM} = sub { die "alarm\n" };
    alarm 10;
    open(my $in, "<", $fifo) or die "open for reading: $!\n";
    my $read = <$in>;
    alarm 0;
    $read;
};
waitpid($writer, 0);
report("with a writer arriving: ", ($line // "failed: $@") =~ s/\n//r, "; writer exit=", $? >> 8);
`)
	wantLines(t, got,
		"open < with nobody at the other end: interrupted: alarm",
		"open > with nobody at the other end: interrupted: alarm",
		"with a writer arriving: through the fifo; writer exit=0",
	)
}

// An open of a FIFO that a signal interrupted is over when the program is
// told: neither end of the FIFO is open any more, so the next open of the
// other end finds nobody — a writer that does not wait is refused (ENXIO),
// and a reader finds the end of input. Opening with O_NONBLOCK does not
// wait for the other end at all.
func TestInterruptedOpenOfAFIFOLeavesNoEndOpen(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("FIFOs are made on linux and darwin")
	}
	got := runForkScript(t, `
use POSIX qw(mkfifo);
use Fcntl qw(O_RDONLY O_WRONLY O_NONBLOCK);
use Time::HiRes qw(alarm);
use File::Temp ();
my $dir = File::Temp::tempdir(CLEANUP => 1);
my $fifo = "$dir/fifo";
mkfifo($fifo, 0600) or die "mkfifo: $!";
my $rounds = 20;
my %seen;
for my $round (1 .. $rounds) {
    for my $mode ("<", ">") {
        local $SIG{ALRM} = sub { };
        alarm 0.02;
        my $opened = open(my $fh, $mode, $fifo);
        my $interrupted = !$opened && $!{EINTR};
        alarm 0;
        my $left;
        if ($mode eq "<") {
            # no reader may be left: a writer that does not wait is refused
            $left = sysopen(my $w, $fifo, O_WRONLY | O_NONBLOCK) ? "a reader is left" : $!{ENXIO} ? "nothing is left" : "error $!";
        } else {
            # no writer may be left: a reader is at the end of input
            sysopen(my $r, $fifo, O_RDONLY | O_NONBLOCK) or die "open for reading without waiting: $!";
            my $n = sysread($r, my $byte, 1);
            $left = defined $n ? "nothing is left" : $!{EAGAIN} ? "a writer is left" : "error $!";
        }
        $seen{"open $mode " . ($interrupted ? "interrupted" : $opened ? "opened" : "failed: $!") . ", $left"}++;
    }
}
report("$_: $seen{$_}") for sort keys %seen;

# without waiting, and with both ends open
sysopen(my $r, $fifo, O_RDONLY | O_NONBLOCK) or die "open for reading without waiting: $!";
sysopen(my $w, $fifo, O_WRONLY | O_NONBLOCK) or die "open for writing without waiting: $!";
my $n = sysread($r, my $buf, 16);
report("read of the empty FIFO: ", defined $n ? "$n bytes" : $!{EAGAIN} ? "EAGAIN" : "error $!");
syswrite($w, "through") or die "write: $!";
$n = sysread($r, $buf, 16);
report("read after a write: ", defined $n ? "'$buf'" : "error $!");
`)
	wantLines(t, got,
		"open < interrupted, nothing is left: 20",
		"open > interrupted, nothing is left: 20",
		"read of the empty FIFO: EAGAIN",
		"read after a write: 'through'",
	)
}

// A program that runs as root changes the ids of its process, for all of
// the process (every thread of it), and has the rights of the user it
// became. The program is the process, so it is run as one of its own.
func TestRootChangesTheIDsOfTheProcess(t *testing.T) {
	if runtime.GOOS != "linux" || os.Geteuid() != 0 {
		t.Skip("needs to run as root on Linux")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	secret := dir + "/secret"
	if err := os.WriteFile(secret, []byte("for root\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := perlProcess("-e", `
use strict; use warnings;
my $secret = shift;
sub threads_as { my $want = shift; my %seen; for my $task (glob "/proc/self/task/*/status") { open(my $st, "<", $task) or next; while (<$st>) { $seen{(split)[2]}++ if /^Uid:/ } } join(",", sort keys %seen) eq $want ? "every thread" : "threads as " . join(",", sort keys %seen) }
print "as root: uid $< euid $>, the file ", (open(my $in, "<", $secret) ? "opens" : "does not open: $!"), "\n";
$) = "65534 65534";
print "group: egid ", (split ' ', $))[0], "\n";
$> = 65534;
print "effective user changed: uid $< euid $>, ", threads_as(65534), ", the file ", (open(my $in2, "<", $secret) ? "opens" : "does not open"), "\n";
$> = 0;
print "back: euid $>, the file ", (open(my $in3, "<", $secret) ? "opens" : "does not open"), "\n";
$< = $> = 65534;
print "both: uid $< euid $>\n";
$> = 0;
print "no way back: euid $>\n";
`, secret)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("the program failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	want := strings.Join([]string{
		"as root: uid 0 euid 0, the file opens",
		"group: egid 65534",
		"effective user changed: uid 0 euid 65534, every thread, the file does not open",
		"back: euid 0, the file opens",
		"both: uid 65534 euid 65534",
		"no way back: euid 65534",
		"",
	}, "\n")
	if stdout.String() != want {
		t.Errorf("stdout:\n%s\nwant:\n%s\nstderr:\n%s", stdout.String(), want, stderr.String())
	}
}

// The monotonic clock of a program that is the host process is the host's:
// it counts from the start of the system, as it does for every process
// there, and a program may take its readings for large numbers (a
// scheduler that tells a delay from a point in time by its size does).
func TestMonotonicClockIsTheHosts(t *testing.T) {
	const systemPerl = "/usr/bin/perl"
	if _, err := os.Stat(systemPerl); err != nil {
		t.Skip("the system's perl says what the host's clock reads")
	}
	got := runForkScriptArgs(t, `
use Time::HiRes qw(clock_gettime CLOCK_MONOTONIC);
my $perl = shift;
my $ours = clock_gettime(CLOCK_MONOTONIC);
my $theirs = qx($perl -MTime::HiRes=clock_gettime,CLOCK_MONOTONIC -e 'print clock_gettime(CLOCK_MONOTONIC)');
die "the system's perl gave no reading" unless $theirs =~ /^[0-9.]+$/;
my $later = clock_gettime(CLOCK_MONOTONIC);
report("the reading is ", ($theirs - $ours > -1 && $theirs - $ours < 30 ? "the host's" : "$ours where the host reads $theirs"),
       ", and ", ($later >= $ours ? "does not go back" : "goes back"));
`, systemPerl)
	wantLines(t, got, "the reading is the host's, and does not go back")
}
