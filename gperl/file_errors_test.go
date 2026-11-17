//go:build darwin || linux

package gperl_test

import (
	"os"
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
