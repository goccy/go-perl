//go:build darwin || linux

package gperl_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The perl command's readdir lists a directory in the order the system
// gives its names, with "." and ".." where the system puts them, as perl's
// does: the names other than "." and ".." are in the order Readdirnames
// reads them from the system, and "." and ".." are listed once each. The
// names were sorted, with "." and ".." first.
func TestReaddirIsInTheSystemsOrder(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 64; i++ {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%02d", (i*37)%64)), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	d, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	names, err := d.Readdirnames(-1)
	if err := errors.Join(err, d.Close()); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-e", `opendir my $d, $ARGV[0] or die "opendir: $!"; print "$_\n" for readdir $d`, dir)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	listed := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	others := slices.DeleteFunc(slices.Clone(listed), func(name string) bool { return name == "." || name == ".." })
	if err != nil || !slices.Equal(others, names) || len(listed)-len(others) != 2 || !slices.Contains(listed, ".") || !slices.Contains(listed, "..") {
		t.Fatalf("err %v, readdir lists:\n%s\nthe system lists, besides . and ..:\n%s", err, strings.TrimSpace(string(out)), strings.Join(names, "\n"))
	}
}
