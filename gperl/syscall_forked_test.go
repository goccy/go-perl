//go:build linux && (arm64 || amd64)

package gperl_test

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
)

// syscall() is the host's system call for the program, and dies in a
// process the program forked: a system call acts on the operating system's
// process, which a forked process of gperl shares with its parent.
func TestSyscallDiesInAForkedProcess(t *testing.T) {
	getpid := map[string]int{"arm64": 172, "amd64": 39}[runtime.GOARCH]
	program := `my $getpid = ` + strconv.Itoa(getpid) + `;
$| = 1;
my $pid = fork // die "fork: $!";
if (!$pid) {
    my $r = eval { syscall($getpid) };
    print "child: ", defined $r ? "returned $r" : $@;
    exit 0;
}
waitpid $pid, 0;
print "parent: ", syscall($getpid) == $$ ? "its own pid" : "another pid", "\n";`
	cmd := exec.Command(os.Args[0], "-e", program)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	want := "child: syscall() cannot be used in a forked process: a system call acts on the operating system's process," +
		" which a process forked by gperl shares with the process it was forked from at -e line 5.\n" +
		"parent: its own pid\n"
	if err != nil || string(out) != want {
		t.Fatalf("err %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}
