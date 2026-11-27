package gperl_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/goccy/go-perl/internal/gperl"
)

// asPerlEnv makes the test binary behave as the perl command (TestMain):
// the tests that need a perl process of its own — one that is started with
// descriptors, say — start the test binary again with it.
const asPerlEnv = "GPERL_TEST_RUN_AS_PERL"

func TestMain(m *testing.M) {
	if os.Getenv(asPerlEnv) != "" {
		os.Unsetenv(asPerlEnv)
		status, err := gperl.RunCLI(os.Args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(status)
	}
	status := m.Run()
	gperl.RemoveSelfPerl()
	os.Exit(status)
}

// perlProcess prepares the test binary to run as `perl args...`.
func perlProcess(args ...string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	return cmd
}

// A program finds the descriptors above 2 it was started with open, under
// their numbers: a file to read (perl -e 'open(STDIN, "<&5")' 5<file), a
// pipe to write to, a listening socket to accept on. A command the program
// runs inherits them in turn.
func TestProgramHasTheDescriptorsItWasStartedWith(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("descriptors are handed to a new process on linux and darwin")
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(input, []byte("from descriptor five\nsecond line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer pipeRead.Close()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	listening, err := listener.(*net.TCPListener).File()
	if err != nil {
		t.Fatal(err)
	}
	defer listening.Close()

	cmd := perlProcess("-e", `
use strict; use warnings;
use IO::Handle; use IO::Socket::INET;
# before the program opens anything itself
print "descriptor 4 is ", (open(my $none, "<&=", 4) ? "open" : "not open"), "\n";
open(STDIN, "<&5") or die "dup of descriptor 5: $!";
sysread(STDIN, my $line, 21) == 21 or die "read of descriptor 5: $!";   # the first line, and no more
print "read: $line";
open(my $out, ">&=", 6) or die "descriptor 6: $!";
$out->autoflush(1);
print {$out} "written to descriptor six\n";
print "descriptor 6 is ", (-p $out ? "a pipe" : "something else"), "\n";
close $out;
# a command the program runs has descriptor 5 too, and reads on from where
# the program stopped (descriptor 6 the program opened itself: perl marks
# that close-on-exec)
$| = 1;
system("/bin/sh", "-c", "cat <&5") == 0 or die "the command failed: $?";
my $server = IO::Socket::INET->new_from_fd(7, "r+") or die "descriptor 7: $!";
print "descriptor 7 is ", (-S $server ? "a socket" : "something else"), " on port ", $server->sockport, "\n";
my $conn = $server->accept or die "accept: $!";
print "accepted: ", scalar <$conn>;
`)
	// descriptors 3 and 4 are not handed over
	cmd.ExtraFiles = []*os.File{nil, nil, file, pipeWrite, listening}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pipeWrite.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	conn, err := net.Dial("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(conn, "a connection to the inherited socket")
	conn.Close()
	piped, _ := io.ReadAll(pipeRead)
	if err := cmd.Wait(); err != nil {
		t.Fatalf("the program failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	want := strings.Join([]string{
		"descriptor 4 is not open",
		"read: from descriptor five",
		"descriptor 6 is a pipe",
		"second line",
		fmt.Sprintf("descriptor 7 is a socket on port %d", port),
		"accepted: a connection to the inherited socket",
		"",
	}, "\n")
	if stdout.String() != want {
		t.Errorf("stdout:\n%s\nwant:\n%s\nstderr:\n%s", stdout.String(), want, stderr.String())
	}
	if got, want := string(piped), "written to descriptor six\n"; got != want {
		t.Errorf("the pipe at descriptor 6 received %q, want %q", got, want)
	}
}
