//go:build linux || darwin

package gperl_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// A program that exits while processes it forked still run ends at once
// with its own status, as on perl, and the forked processes, which run in
// the program's process, end with it. That is said on standard error, with
// the pids fork returned, where on perl they would go on (a daemon, a child
// that writes after its parent exited).
func TestProgramEndSaysWhichForkedProcessesEnd(t *testing.T) {
	program := `$| = 1;
my @pids;
for (1, 2) {
    my $pid = fork // die "fork: $!";
    if (!$pid) { sleep 30; print "child $$ went on\n"; exit 0 }
    push @pids, $pid;
}
print "pids @pids\n";
exit 3;`
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(os.Args[0], "-e", program)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	start := time.Now()
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 3 {
		t.Fatalf("err %v, want exit status 3; stderr %q", err, stderr.String())
	}
	if took := time.Since(start); took > 20*time.Second {
		t.Errorf("the program took %v to end, waiting for its forked processes", took)
	}
	pids := strings.Fields(strings.TrimPrefix(strings.TrimSpace(stdout.String()), "pids "))
	if len(pids) != 2 || strings.Contains(stdout.String(), "went on") {
		t.Fatalf("stdout %q", stdout.String())
	}
	want := "perl: the program exited while processes it forked were still running;" +
		" they run inside its process and end with it: " + strings.Join(pids, ", ") + "\n"
	if stderr.String() != want {
		t.Errorf("stderr %q, want %q", stderr.String(), want)
	}
}
