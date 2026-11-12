package gperl_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// forkLoop is a program that fills its heap with about the number of
// megabytes named by its second argument, then forks the number of children
// named by its third, each ending at once and waited for; between two forks
// it changes a little of what it holds, as a server forking its workers
// does. It reports the mean time of a fork and wait, in microseconds. One
// more child, forked first, is not timed: the host may make the image of the
// filled heap for it that the following forks share.
const forkLoop = `
use Time::HiRes ();
use POSIX ();
my ($megabytes, $forks) = @ARGV;
my @heap = map { "x" x 1024 } 1 .. 1024 * $megabytes;
my $first = fork // die "fork: $!";
POSIX::_exit(0) if $first == 0;
waitpid($first, 0);
my $start = Time::HiRes::time();
for my $i (1 .. $forks) {
    $heap[$i] = "y" x 1024;
    my $pid = fork // die "fork: $!";
    POSIX::_exit($heap[$i] eq "y" x 1024 && $heap[-1] eq "x" x 1024 ? 0 : 1) if $pid == 0;
    waitpid($pid, 0);
    die "child $i saw another heap\n" if $?;
}
report(int((Time::HiRes::time() - $start) * 1e6 / $forks));
`

// A fork costs what the program changed since its last fork, not what it
// holds: a program with a large heap forks its children about as fast as
// one with a small heap. The host has to be able to tell which pages of the
// program's memory changed for that, which it can on Linux; elsewhere the
// memory is copied whole for every child.
func TestForkCostDoesNotGrowWithTheHeap(t *testing.T) {
	if runtime.GOOS != "linux" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		t.Skip("the host cannot tell which pages a program changed: every fork copies the memory")
	}
	got := runForkScriptArgs(t, forkLoop, "128", "40")
	if len(got) != 1 {
		t.Fatalf("report: %q", got)
	}
	micros, err := strconv.Atoi(got[0])
	if err != nil {
		t.Fatalf("report: %q", got)
	}
	// Copying 128 MB takes tens of milliseconds; a fork that copies only
	// the changed pages takes a fraction of one.
	if micros > 8000 {
		t.Fatalf("a fork with a 128 MB heap takes %d us: the heap is copied for every child", micros)
	}
}

// What a fork child sees is the program's state at the fork, whether the
// program changed little or most of its memory since the previous fork, and
// whatever the children and the program do afterwards.
func TestForkChildrenSeeTheStateAtTheirFork(t *testing.T) {
	got := runForkScript(t, `
my @heap = map { "a" x 4096 } 1 .. 2048;
my @pids;
for my $round (1 .. 6) {
    # odd rounds change one element, even rounds all of them
    if ($round % 2) { $heap[$round] = "r$round" } else { @heap = map { "r$round" x 1024 } 1 .. 2048 }
    my $want = join ",", map { substr($heap[$_], 0, 2) } 0 .. 7;
    my $pid = fork // die "fork: $!";
    if ($pid == 0) {
        select(undef, undef, undef, 0.3);   # the program goes on changing its heap meanwhile
        my $have = join ",", map { substr($heap[$_], 0, 2) } 0 .. 7;
        @heap = ();
        exit($have eq $want ? 0 : 1);
    }
    push @pids, $pid;
}
@heap = map { "end" } 1 .. 2048;
my @status = map { waitpid($_, 0); $? >> 8 } @pids;
report("children: @status; program: $heap[0] $heap[-1]");
`)
	wantLines(t, got, "children: 0 0 0 0 0 0; program: end end")
}

// BenchmarkFork measures one fork of the program and the wait for its
// child, for a program whose heap holds 1 MB of data and for one whose heap
// holds 64 MB.
func BenchmarkFork(b *testing.B) {
	for _, megabytes := range []int{1, 64} {
		b.Run(fmt.Sprintf("heap-%dMB", megabytes), func(b *testing.B) {
			dir := b.TempDir()
			script := filepath.Join(dir, "fork.pl")
			report := filepath.Join(dir, "report.txt")
			src := `my $REPORT = shift @ARGV;
sub report { open(my $fh, '>>', $REPORT) or die "report: $!"; syswrite($fh, join('', @_) . "\n"); close $fh }
` + forkLoop
			if err := os.WriteFile(script, []byte(src), 0o644); err != nil {
				b.Fatal(err)
			}
			status, err := gperl.Run(script, []string{report, strconv.Itoa(megabytes), strconv.Itoa(b.N)})
			if err != nil || status != 0 {
				b.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
			}
			raw, err := os.ReadFile(report)
			if err != nil {
				b.Fatal(err)
			}
			micros, err := strconv.ParseFloat(strings.TrimSpace(string(raw)), 64)
			if err != nil {
				b.Fatal(err)
			}
			// The program times its own loop: starting the interpreter
			// and filling the heap are not part of a fork.
			b.ReportMetric(micros*1000, "ns/fork")
		})
	}
}
