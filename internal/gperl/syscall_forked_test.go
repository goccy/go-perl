//go:build linux && (arm64 || amd64)

package gperl_test

import (
	"os"
	"os/exec"
	"strconv"
	"testing"
)

// syscall() is the host's system call for the program. A process the
// program forked shares the operating system's process with its parent, so
// there syscall() makes only the calls that depend on no state of that
// process, and the numbers the architecture leaves unassigned, and dies for
// any other. The numbers are the guest's, the Linux ABI's of arm64, on
// every host architecture: getpid, getrandom, and 355, which the ABI
// leaves unassigned (Rand::Urandom's rand() in a forked child calls it on
// aarch64, so that it falls back to /dev/urandom). The child's first two
// lines are what perl 5.36 on aarch64 Linux prints; perl's child gets its
// own pid where gperl's dies.
func TestSyscallInAForkedProcess(t *testing.T) {
	numbers := [3]int{172, 278, 355}
	program := `my ($getpid, $getrandom, $unassigned) = (` + strconv.Itoa(numbers[0]) + `, ` +
		strconv.Itoa(numbers[1]) + `, ` + strconv.Itoa(numbers[2]) + `);
$| = 1;
sub unassigned { local $!; my $r = syscall($unassigned, 0, 0, 0); "$r " . ($! + 0) }
my $parent = unassigned();
my $pid = fork // die "fork: $!";
if (!$pid) {
    my $buf = ' ' x 8;
    my $n = syscall($getrandom, $buf, 8, 0);
    print "child getrandom: $n, ", ($buf =~ /^ +$/ ? "unfilled" : "filled"), "\n";
    my $child = unassigned();
    print "child unassigned: ", ($child eq $parent ? "as the parent" : "$child, the parent $parent"), "\n";
    my $r = eval { syscall($getpid) };
    print "child getpid: ", defined $r ? "returned $r\n" : $@;
    exit 0;
}
waitpid $pid, 0;
print "parent: ", syscall($getpid) == $$ ? "its own pid" : "another pid", "\n";`
	cmd := exec.Command(os.Args[0], "-e", program)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	want := "child getrandom: 8, filled\n" +
		"child unassigned: as the parent\n" +
		"child getpid: syscall(" + strconv.Itoa(numbers[0]) + ") cannot be made in a forked process: the system call" +
		" depends on the operating system's process, which a process forked by gperl shares with the process it was forked from at -e line 12.\n" +
		"parent: its own pid\n"
	if err != nil || string(out) != want {
		t.Fatalf("err %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}
