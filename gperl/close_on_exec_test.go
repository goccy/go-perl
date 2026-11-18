//go:build linux

package gperl_test

import (
	"os"
	"os/exec"
	"testing"
)

// The close-on-exec flag fcntl sets and clears is the kernel's flag of the
// descriptor, as perl's is: what F_GETFD reads back is what the kernel
// shows in /proc/self/fdinfo (Run reads it there), and a dup2 onto the
// descriptor clears it. The kernel kept every descriptor close-on-exec,
// whatever the program set. What is expected is what perl prints.
func TestCloseOnExecIsTheKernels(t *testing.T) {
	program := `use POSIX (); use Fcntl qw(F_GETFD F_SETFD FD_CLOEXEC);
sub fdinfo { open my $i, '<', "/proc/self/fdinfo/$_[0]" or return 'none'; my ($f) = map /^flags:\s*(\d+)/ ? $1 : (), <$i>; $f }
pipe my $r, my $w or die $!;
my $n = fileno $r;
fcntl($r, F_SETFD, FD_CLOEXEC) or die;
printf "set: getfd %d, fdinfo %s\n", fcntl($r, F_GETFD, 0), fdinfo($n);
fcntl($r, F_SETFD, 0) or die;
printf "cleared: getfd %d, fdinfo %s\n", fcntl($r, F_GETFD, 0), fdinfo($n);
fcntl($r, F_SETFD, FD_CLOEXEC) or die;
my $t = POSIX::dup($n) // die;
POSIX::dup2($t, $n) // die;
POSIX::close($t);
printf "after dup2: getfd %d, fdinfo %s\n", fcntl($r, F_GETFD, 0), fdinfo($n);`
	cmd := exec.Command(os.Args[0], "-e", program)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	const want = "set: getfd 1, fdinfo 02000000\ncleared: getfd 0, fdinfo 00\nafter dup2: getfd 0, fdinfo 00\n"
	if err != nil || string(out) != want {
		t.Fatalf("err %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}
