//go:build darwin || linux

package gperl_test

import (
	"os"
	"path/filepath"
	"testing"
)

// The host process follows the program's working directory and
// environment, which is where native code that runs for the program looks.
// A forked process is a process of its own in the same host process: what
// it changes stays its own.
func TestHostProcessFollowsTheProgram(t *testing.T) {
	start, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(start)
	t.Setenv("GPERL_FOLLOWED", "before")
	t.Setenv("GPERL_REMOVED", "before")
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(dir, "other")
	if err := os.Mkdir(other, 0o755); err != nil {
		t.Fatal(err)
	}
	got := runForkScriptArgs(t, `
use POSIX ();
my ($dir, $other) = @ARGV;
my $pid = fork // die "fork: $!";
if (!$pid) {
    chdir $other or die "chdir: $!";
    $ENV{GPERL_FOLLOWED} = "in the child";
    POSIX::_exit(0);
}
waitpid($pid, 0);
chdir $dir or die "chdir: $!";
$ENV{GPERL_FOLLOWED} = "after";
delete $ENV{GPERL_REMOVED};
{
    local $ENV{GPERL_LOCAL} = "scoped";
}
report("done");
`, dir, other)
	wantLines(t, got, "done")
	if now, _ := os.Getwd(); now != dir {
		t.Errorf("working directory of the host process = %s, want %s", now, dir)
	}
	if got := os.Getenv("GPERL_FOLLOWED"); got != "after" {
		t.Errorf("GPERL_FOLLOWED in the host process = %q, want %q", got, "after")
	}
	if _, set := os.LookupEnv("GPERL_REMOVED"); set {
		t.Error("a variable the program deleted is still in the host process's environment")
	}
	if _, set := os.LookupEnv("GPERL_LOCAL"); set {
		t.Error("a variable the program set for a scope outlived it in the host process's environment")
	}
}
