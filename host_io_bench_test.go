package perl

import (
	"bytes"
	"context"
	"os"
	"runtime"
	"testing"

	goperlfs "github.com/goccy/go-perl/fs"
)

// What descriptors cost a program that is the host process: each workload
// runs as a whole program on a fresh instance, configured the way the perl
// command configures it, and does little besides reading, writing and
// waiting on descriptors of one kind.
var hostIOWorkloads = []struct {
	name string
	code string
}{
	{"PipeRoundTrip", `
pipe(my $r, my $w) or die "pipe: $!";
my ($msg, $buf) = ("x" x 64);
for (1 .. 200_000) {
    syswrite($w, $msg) == 64 or die "write: $!";
    sysread($r, $buf, 64) == 64 or die "read: $!";
}
print "done\n";
`},
	{"SocketPairRoundTrip", `
use Socket;
socketpair(my $a, my $b, AF_UNIX, SOCK_STREAM, 0) or die "socketpair: $!";
my ($msg, $buf) = ("x" x 64);
for (1 .. 100_000) {
    syswrite($a, $msg) == 64 or die "write: $!";
    sysread($b, $buf, 64) == 64 or die "read: $!";
    syswrite($b, $msg) == 64 or die "write: $!";
    sysread($a, $buf, 64) == 64 or die "read: $!";
}
print "done\n";
`},
	{"TCPLoopbackRoundTrip", `
use IO::Socket::INET;
my $listener = IO::Socket::INET->new(Listen => 5, LocalAddr => "127.0.0.1", LocalPort => 0, Proto => "tcp") or die "listen: $!";
my $client = IO::Socket::INET->new(PeerAddr => "127.0.0.1", PeerPort => $listener->sockport, Proto => "tcp") or die "connect: $!";
my $server = $listener->accept or die "accept: $!";
my ($msg, $buf) = ("x" x 64);
for (1 .. 50_000) {
    send($client, $msg, 0) == 64 or die "send: $!";
    defined recv($server, $buf, 64, 0) or die "recv: $!";
    send($server, $msg, 0) == 64 or die "send: $!";
    defined recv($client, $buf, 64, 0) or die "recv: $!";
}
print "done\n";
`},
	{"FileBlocks", `
use File::Temp ();
my ($fh, $path) = File::Temp::tempfile(UNLINK => 1);
my $block = "0123456789abcdef" x 512;
for (1 .. 4000) { syswrite($fh, $block) == 8192 or die "write: $!" }
sysseek($fh, 0, 0);
my ($buf, $total) = ("", 0);
while ((my $n = sysread($fh, $buf, 8192)) > 0) { $total += $n }
sysseek($fh, 0, 0);
while ((my $n = sysread($fh, $buf, 65536)) > 0) { $total += $n }
print "$total\n";
`},
	{"FileLines", `
use File::Temp ();
my ($fh, $path) = File::Temp::tempfile(UNLINK => 1);
print $fh "line number $_ of the file\n" for 1 .. 500_000;
close($fh);
open(my $in, "<", $path) or die "open: $!";
my $total = 0;
$total += length while <$in>;
print "$total\n";
`},
	{"SelectReadyPipes", `
my (@r, @w);
for (0 .. 7) { pipe($r[$_], $w[$_]) or die "pipe: $!"; syswrite($w[$_], "x") }
my $rin = "";
vec($rin, fileno($_), 1) = 1 for @r;
my $found = 0;
for (1 .. 100_000) { $found += select(my $rout = $rin, undef, undef, 0) }
print "$found\n";
`},
	{"ForkedWriterThroughAPipe", `
pipe(my $r, my $w) or die "pipe: $!";
my $pid = fork() // die "fork: $!";
if (!$pid) {
    close($r);
    my $block = "x" x 65536;
    for (1 .. 1000) { my $off = 0; while ($off < 65536) { $off += syswrite($w, $block, 65536 - $off, $off) // die "write: $!" } }
    exit 0;
}
close($w);
my ($buf, $total) = ("", 0);
while ((my $n = sysread($r, $buf, 65536)) > 0) { $total += $n }
waitpid($pid, 0);
print "$total $?\n";
`},
	{"SpawnedCommands", `
for (1 .. 100) { system("/usr/bin/true") == 0 or die "system: $?" }
my $out = "";
$out .= ` + "`/bin/echo $_`" + ` for 1 .. 100;
print length($out), "\n";
`},
}

func BenchmarkHostIO(b *testing.B) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		b.Skip("the workloads use a Unix host")
	}
	stdlib, err := ExtractStdlib()
	if err != nil {
		b.Fatal(err)
	}
	for _, w := range hostIOWorkloads {
		b.Run(w.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				var out, errOut bytes.Buffer
				p, err := New(Config{
					FS:          goperlfs.NewHostFS(),
					StdlibDir:   stdlib,
					Env:         os.Environ(),
					Stdout:      &out,
					Stderr:      &errOut,
					Dial:        func(string, string, string, int) bool { return true },
					Resolve:     func(string) bool { return true },
					Exec:        func(string, []string) bool { return true },
					Fork:        true,
					HostProcess: true,
				})
				if err != nil {
					b.Fatal(err)
				}
				status, err := p.RunProgram(context.Background(), Program{Code: []string{w.code}})
				if err != nil || status != 0 || out.Len() == 0 {
					b.Fatalf("status=%d err=%v stdout=%q stderr=%q", status, err, out.String(), errOut.String())
				}
				p.Close()
			}
		})
	}
}
