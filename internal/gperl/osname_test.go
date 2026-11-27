package gperl_test

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

// $^O and $Config{osname} name the system the program runs on, as perl's
// do there: linux on Linux, darwin on macOS. They named the system the
// interpreter was built for (wasi), and code that picks a path or a
// structure's layout by $^O took the one of no system it runs on
// (IPC::Shareable read a segment's size as macOS lays it out). What is
// expected is what perl prints on the same system.
func TestOSNameIsTheSystemItRunsOn(t *testing.T) {
	program := `use Config; print "$^O $Config{osname}\n"; print Config::myconfig() =~ /osname=\Q$^O\E/ ? "myconfig\n" : "myconfig differs\n"`
	cmd := exec.Command(os.Args[0], "-e", program)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	if want := runtime.GOOS + " " + runtime.GOOS + "\nmyconfig\n"; err != nil || string(out) != want {
		t.Fatalf("err %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}
