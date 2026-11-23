//go:build darwin || linux

package gperl_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The perl command's readdir lists a directory in the order the system
// gives its names, with "." and ".." where the system puts them, as perl's
// does: what is expected is what ls -f lists, which reads the directory
// with readdir(3) as perl does. The names were sorted, with "." and ".."
// first.
func TestReaddirIsInTheSystemsOrder(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 64; i++ {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%02d", (i*37)%64)), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ls, err := exec.Command("ls", "-f", dir).Output()
	if err != nil {
		t.Fatalf("ls -f: %v", err)
	}
	cmd := exec.Command(os.Args[0], "-e", `opendir my $d, $ARGV[0] or die "opendir: $!"; print "$_\n" for readdir $d`, dir)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != string(ls) {
		t.Fatalf("err %v, readdir lists:\n%s\nls -f lists:\n%s", err, strings.TrimSpace(string(out)), strings.TrimSpace(string(ls)))
	}
}
