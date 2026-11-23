//go:build darwin || linux

package gperl_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// The errors of calls on files are the system's: chdir into a directory the
// user may not search fails with EACCES (it needs search permission, not
// reading permission), and rename of a file onto a directory fails with
// EISDIR, as rename(2) has it; a directory renamed onto an empty one
// replaces it; rmdir of / is the system's answer.
func TestErrorsOfFileCalls(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root may search any directory")
	}
	// rmdir(2) of / is EBUSY on Linux and EISDIR on darwin.
	rootRmdirError := "EBUSY"
	if runtime.GOOS == "darwin" {
		rootRmdirError = "EISDIR"
	}
	out, status := hostPerl(t, "-MErrno", "-MFile::Temp=tempdir", "-e", `
sub errname { my $e = $! + 0; (grep { $!{$_} } sort keys %!)[0] // $e }
my $dir = tempdir(CLEANUP => 1);
mkdir "$dir/unsearchable" or die "mkdir: $!";
chmod 0644, "$dir/unsearchable" or die "chmod: $!";
print "chdir into a directory without search permission: ", (chdir("$dir/unsearchable") ? "done" : errname()), "\n";
chmod 0755, "$dir/unsearchable";
print "chdir into it with search permission: ", (chdir("$dir/unsearchable") ? "done" : errname()), "\n";
chdir "/";
mkdir "$dir/target"; mkdir "$dir/empty"; mkdir "$dir/moved";
open(my $fh, '>', "$dir/file") or die; close $fh;
print "rename of a file onto a directory: ", (rename("$dir/file", "$dir/target") ? "done" : errname()), "\n";
print "rename of a directory onto an empty one: ", (rename("$dir/moved", "$dir/empty") ? "done" : errname()), "\n";
print "rmdir of the root: ", (rmdir("/") ? "done" : errname()), "\n";
`)
	want := "chdir into a directory without search permission: EACCES\n" +
		"chdir into it with search permission: done\n" +
		"rename of a file onto a directory: EISDIR\n" +
		"rename of a directory onto an empty one: done\n" +
		"rmdir of the root: " + rootRmdirError + "\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}

// A command started in a working directory the user may no longer search
// starts there, as the child of fork and exec does (it inherits the
// directory, it does not change into it).
func TestCommandStartsInADirectoryThatCannotBeSearched(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root may search any directory")
	}
	out, status := hostPerl(t, "-MFile::Temp=tempdir", "-e", `
my $dir = tempdir(CLEANUP => 1);
mkdir "$dir/here" or die "mkdir: $!";
chdir "$dir/here" or die "chdir: $!";
chmod 0644, "$dir/here" or die "chmod: $!";
my $out = qx(/bin/echo started);
my $status = $?;
chmod 0755, "$dir/here";
chdir "/";
print "command: ", ($status == 0 ? $out : "status $status\n");
`)
	if want := "command: started\n"; status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}

// A program started in a working directory the user may not search runs
// there, as perl does, and so do the commands it starts: by a pipe open of
// $^X, which is a new perl process started there, and by fork and exec.
func TestProgramStartedInADirectoryThatCannotBeSearched(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root may search any directory")
	}
	if runtime.GOOS == "darwin" {
		t.Skip("darwin's getcwd does not name a directory the process may not search")
	}
	dir := filepath.Join(t.TempDir(), "here")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	program := `
$| = 1;
open(my $fh, '-|', $^X, '-e', 'print "started\n"') or die "pipe open: $!";
print "a new perl process: ", scalar(<$fh>) // "nothing\n";
close $fh;
my $pid = fork // die "fork: $!";
if (!$pid) { exec "/bin/echo", "a forked process's command: started" or die "exec: $!" }
waitpid $pid, 0;
`
	// The shell goes into the directory while it may, and starts perl
	// there once it may not.
	cmd := exec.Command("/bin/sh", "-c", `cd "$1" && chmod 0600 . && exec "$2" -e "$3"`, "sh", dir, os.Args[0], program)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	want := "a new perl process: started\na forked process's command: started\n"
	if err != nil || string(out) != want {
		t.Fatalf("perl: %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}
