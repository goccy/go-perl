//go:build linux

package gperl_test

import (
	"os"
	"path/filepath"
	"testing"
)

// fcntl duplicates a descriptor under the lowest free number from the one
// given (F_DUPFD), and takes POSIX record locks of a file (F_SETLK,
// F_GETLK), which another process is refused while they are held, as on
// Linux: Run redirects through F_DUPFD, Mail::Box locks a folder with
// F_SETLK.
func TestFcntlDuplicatesAndLocks(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "folder")
	if err := os.WriteFile(file, []byte("first line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, status := hostPerl(t, "-MFcntl=:DEFAULT,:flock,F_DUPFD,F_SETLK,F_GETLK,F_RDLCK,F_WRLCK,F_UNLCK,SEEK_SET", "-e", `
my $file = shift;
open(my $fh, '<', $file) or die "open: $!";
my $n = fcntl($fh, F_DUPFD, 20);
print "F_DUPFD: ", (defined $n ? ($n >= 20 ? "at least 20" : "at $n") : "failed: $!"), "\n";
open(my $dup, '<&=', $n) or die "fdopen: $!";
print "the duplicate reads: ", scalar(<$dup>);
open(my $lk, '+<', $file) or die "open: $!";
my $lock = pack('s s x4 q q i x4', F_WRLCK, SEEK_SET, 0, 0, 0);
print "write lock: ", (fcntl($lk, F_SETLK, $lock) ? "taken" : "refused: $!"), "\n";
my $other = q{
    use Fcntl qw(F_SETLK F_GETLK F_RDLCK F_WRLCK SEEK_SET);
    use Errno qw(EAGAIN EACCES);
    open(my $lk, '+<', $ARGV[0]) or die "open: $!";
    my $lock = pack('s s x4 q q i x4', F_RDLCK, SEEK_SET, 0, 0, 0);
    print "another process: ", (fcntl($lk, F_SETLK, $lock) ? "taken" : ($!{EAGAIN} || $!{EACCES}) ? "refused" : "failed: $!"), "\n";
    my $probe = pack('s s x4 q q i x4', F_RDLCK, SEEK_SET, 0, 0, 0);
    fcntl($lk, F_GETLK, $probe) or die "F_GETLK: $!";
    my ($type, undef, undef, undef, $pid) = unpack('s s x4 q q i x4', $probe);
    print "in the way: ", ($type == F_WRLCK ? "a write lock" : "type $type"), " of ", ($pid == getppid() ? "the parent" : "pid $pid"), "\n";
};
system($^X, '-e', $other, $file) == 0 or die "the other process: $?";
my $unlock = pack('s s x4 q q i x4', F_UNLCK, SEEK_SET, 0, 0, 0);
print "unlock: ", (fcntl($lk, F_SETLK, $unlock) ? "done" : "failed: $!"), "\n";
`, file)
	want := "F_DUPFD: at least 20\n" +
		"the duplicate reads: first line\n" +
		"write lock: taken\n" +
		"another process: refused\n" +
		"in the way: a write lock of the parent\n" +
		"unlock: done\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
