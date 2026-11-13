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
// child's file there.
func TestForkedProcessDescriptors(t *testing.T) {
	hostDescriptorsHere(t)
	dir := t.TempDir()
	got := runPerlProcess(t, dir, `
use strict; use warnings;
$| = 1;
sub report { print @_, "\n" }
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
    print $w "after closing it, the child's next file is ", (fileno($own) == $number ? "the parent's number" : "another number"), "\n";
    print $own "the child's own\n";
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
for my $name ("child-out.txt", "child-own.txt") {
    open(my $in, "<", $name) or die "$name: $!";
    report("$name: ", join("|", map { chomp; $_ } <$in>));
}
`)
	wantOutput(t, got,
		"child sees the file as 3",
		"after closing it, the child's next file is another number",
		"child status: 0",
		"the shared file: child,parent",
		"still the parent's standard output",
		"child-out.txt: from the child|from its command",
		"child-own.txt: the child's own",
	)
}
