package gperl_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// hostDescriptorsHere skips a test of the program's descriptors being the
// host process's on a platform where a program has descriptors of its own.
func hostDescriptorsHere(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("a program has the host process's descriptors on linux and darwin")
	}
}

// runPerlProcess runs a program in a perl process of its own and returns
// what it wrote to its standard output.
func runPerlProcess(t *testing.T, dir, program string, args ...string) string {
	t.Helper()
	script := filepath.Join(dir, "program.pl")
	if err := os.WriteFile(script, []byte(program), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := perlProcess(append([]string{script}, args...)...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("perl: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("stderr:\n%s", stderr.String())
	}
	return stdout.String()
}

func wantOutput(t *testing.T, got string, want ...string) {
	t.Helper()
	if expected := strings.Join(want, "\n") + "\n"; got != expected {
		t.Errorf("output:\n%s\nwant:\n%s", got, expected)
	}
}

// The numbers a program gets for its descriptors are the host process's:
// the lowest the process has free, 3 for the first one, and under each
// number the process has the very file the program opened.
func TestDescriptorNumbersAreTheProcesss(t *testing.T) {
	hostDescriptorsHere(t)
	got := runPerlProcess(t, t.TempDir(), `
use strict; use warnings;
sub report { print @_, "\n" }
# Whether the process has, under a number, the file a handle is open on:
# /dev/fd names the descriptors of the process itself, and opening one of
# its names (for what the descriptor is open for) gives the same file.
sub same_as { my ($number, $fh, $mode) = @_;
              open(my $via, $mode, "/dev/fd/$number") or return "not open in the process ($!)";
              my @p = stat($via); my @h = stat($fh);
              @p && @h && $p[0] == $h[0] && $p[1] == $h[1] ? "the process's" : "not the process's" }
sub same { my ($fh, $mode) = @_; same_as(fileno($fh), $fh, $mode // "<") }

report("standard streams: ", join(" ", fileno(STDIN), fileno(STDOUT), fileno(STDERR)), ", ", same(\*STDOUT, ">>"));
open(my $first, "<", $0) or die "open: $!";
report("first file: ", fileno($first), ", ", same($first));
pipe(my $r, my $w) or die "pipe: $!";
report("pipe: ", fileno($r), " ", fileno($w), ", ", same($r), ", ", same($w, ">>"));
opendir(my $dir, ".") or die "opendir: $!";
report("entries of a directory: ", scalar(grep { $_ eq "program.pl" } readdir($dir)));
closedir($dir);
close($first);
open(my $again, "<", "/dev/null") or die "open: $!";
report("after closing it, the next file: ", fileno($again), ", ", same($again), ", ", (-c $again ? "a character device" : "something else"));
open(my $dup, "<&", $again) or die "dup: $!";
report("a duplicate: ", (fileno($dup) != fileno($again) ? "another number" : "the same number"), ", ", same($dup));
open(my $alias, "<&=", fileno($again)) or die "fdopen: $!";
report("a handle on the same descriptor: ", (fileno($alias) == fileno($again) ? "the same number" : "another number"));
require POSIX;
my $to = POSIX::dup2(fileno($r), 20);
report("dup2 onto 20: ", $to, ", ", same_as(20, $r, "<"));
POSIX::close(20);
report("after closing 20: ", (defined POSIX::dup2(20, 21) ? "still open" : "closed"));
`)
	wantOutput(t, got,
		"standard streams: 0 1 2, the process's",
		"first file: 3, the process's",
		"pipe: 4 5, the process's, the process's",
		"entries of a directory: 1",
		"after closing it, the next file: 3, the process's, a character device",
		"a duplicate: another number, the process's",
		"a handle on the same descriptor: the same number",
		"dup2 onto 20: 20, the process's",
		"after closing 20: closed",
	)
}

// A forked process starts with its parent's descriptors, under the same
// numbers. Where it puts a file of its own under a number its parent still
// uses, the parent's stays what it was; a command it then runs has the
// child's file there. After it closes its standard output, the next file it
// opens is descriptor 1, whether or not the parent still has its own there.
// Above the standard descriptors a new one is always the process's under
// its number, which is what a native library is handed: the number of a
// descriptor the parent still has is not given to the child for a file of
// its own, and the parent that closed a connection its child still serves
// accepts the next one under another number.
func TestForkedProcessDescriptors(t *testing.T) {
	hostDescriptorsHere(t)
	dir := t.TempDir()
	got := runPerlProcess(t, dir, `
use strict; use warnings;
$| = 1;
sub report { print @_, "\n" }
# Whether the process has, under the number of a handle, the file the
# handle is open on (/dev/fd names the descriptors of the process itself).
sub same { my ($fh) = @_;
           open(my $via, ">>", "/dev/fd/" . fileno($fh)) or return "not open in the process ($!)";
           my @p = stat($via); my @h = stat($fh);
           @p && @h && $p[0] == $h[0] && $p[1] == $h[1] ? "the process's" : "not the process's" }
open(my $shared, "+>", "shared.txt") or die "open: $!";
my $number = fileno($shared);
pipe(my $r, my $w) or die "pipe: $!";
my $pid = fork() // die "fork: $!";
if (!$pid) {
    close($r);
    print $w "child sees the file as ", fileno($shared), "\n";
    syswrite($shared, "child,");
    # its own standard output, where the parent's still is
    open(STDOUT, ">", "child-out.txt") or die "reopen: $!";
    print "from the child\n";
    system("/bin/sh", "-c", "echo from its command");
    # a descriptor the parent still has
    close($shared);
    open(my $own, ">", "child-own.txt") or die "open: $!";
    print $w "after closing it, the child's next file is ", (fileno($own) == $number ? "the parent's number" : "another number"), ", ", same($own), "\n";
    print $own "the child's own\n";
    # standard output closed and opened again, as a program does that puts
    # a file in its place
    close(STDOUT);
    open(STDOUT, ">>", "child-again.txt") or die "reopen: $!";
    print $w "standard output closed and opened again is ", fileno(STDOUT), "\n";
    print "from the child again\n";
    system("/bin/sh", "-c", "echo from its second command");
    exit 0;
}
close($w);
print while <$r>;
waitpid($pid, 0);
report("child status: $?");
syswrite($shared, "parent");
seek($shared, 0, 0);
report("the shared file: ", scalar(<$shared>));
report("still the parent's standard output");
# a server that forks for each connection: the parent closes what the child
# serves and takes the next one
pipe(my $kept_r, my $kept_w) or die "pipe: $!";
my $served = fileno($kept_w);
pipe(my $done_r, my $done_w) or die "pipe: $!";
my $server = fork() // die "fork: $!";
if (!$server) {
    close($done_w);
    my $end = <$done_r>;
    print $kept_w "serves under ", (fileno($kept_w) == $served ? "the same number" : "another number"), ", ", same($kept_w), "\n";
    exit 0;
}
close($kept_w);
open(my $next, ">", "next.txt") or die "open: $!";
report("the parent's next file is ", (fileno($next) == $served ? "the number its child serves" : "another number"), ", ", same($next));
close($done_w);
report("the child: ", scalar(<$kept_r>) =~ s/\n//r);
waitpid($server, 0);
for my $name ("child-out.txt", "child-own.txt", "child-again.txt") {
    open(my $in, "<", $name) or die "$name: $!";
    report("$name: ", join("|", map { chomp; $_ } <$in>));
}
`)
	wantOutput(t, got,
		"child sees the file as 3",
		"after closing it, the child's next file is another number, the process's",
		"standard output closed and opened again is 1",
		"child status: 0",
		"the shared file: child,parent",
		"still the parent's standard output",
		"the parent's next file is another number, the process's",
		"the child: serves under the same number, the process's",
		"child-out.txt: from the child|from its command",
		"child-own.txt: the child's own",
		"child-again.txt: from the child again|from its second command",
	)
}

// A socket is the process's socket under the number the program sees, and a
// terminal is a terminal to the program: what -t and POSIX::isatty say is
// the host's answer for the descriptor.
func TestSocketsAndTerminalsAreTheProcesss(t *testing.T) {
	hostDescriptorsHere(t)
	got := runPerlProcess(t, t.TempDir(), `
use strict; use warnings;
use Socket; use POSIX ();
sub report { print @_, "\n" }
sub kind { my $n = shift; -S "/dev/fd/$n" ? "a socket of the process's" : -e "/dev/fd/$n" ? "something else there" : "nothing there" }

socket(my $listener, AF_INET, SOCK_STREAM, 0) or die "socket: $!";
report("socket: ", fileno($listener), ", ", kind(fileno($listener)));
bind($listener, pack_sockaddr_in(0, INADDR_LOOPBACK)) or die "bind: $!";
listen($listener, 5) or die "listen: $!";
socket(my $client, AF_INET, SOCK_STREAM, 0) or die "socket: $!";
connect($client, getsockname($listener)) or die "connect: $!";
accept(my $server, $listener) or die "accept: $!";
report("client and accepted connection: ", fileno($client), " ", fileno($server), ", ", kind(fileno($server)));
socketpair(my $a, my $b, AF_UNIX, SOCK_STREAM, 0) or die "socketpair: $!";
report("socket pair: ", fileno($a), " ", fileno($b), ", ", kind(fileno($a)));
close($server);
report("after closing the accepted connection: ", kind(5));

open(my $file, "<", $0) or die "open: $!";
report("a file: ", (-t $file ? "a terminal" : "no terminal"), ", isatty ", (POSIX::isatty(fileno($file)) ? 1 : 0));
if (open(my $pty, "+<", "/dev/ptmx")) {
    report("a pseudo-terminal: ", (-t $pty ? "a terminal" : "no terminal"), ", isatty ", (POSIX::isatty(fileno($pty)) ? 1 : 0));
} else {
    report("no pseudo-terminal: $!");
}
`)
	want := []string{
		"socket: 3, a socket of the process's",
		"client and accepted connection: 4 5, a socket of the process's",
		"socket pair: 6 7, a socket of the process's",
		"after closing the accepted connection: nothing there",
		"a file: no terminal, isatty 0",
	}
	// Where this process can open a pseudo-terminal, the program can.
	if pty, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0); err == nil {
		pty.Close()
		want = append(want, "a pseudo-terminal: a terminal, isatty 1")
	} else {
		lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
		want = append(want, lines[len(lines)-1])
		t.Logf("no pseudo-terminal on this system (%v): what a terminal is to the program is not tested", err)
	}
	wantOutput(t, got, want...)
}
