//go:build darwin || linux

package gperl_test

import (
	"os"
	"path/filepath"
	"testing"
)

// Under `use filetest 'access'`, -r, -w and -x ask access(2), which decides
// by the file's mode (and access control lists) for the process's ids: a
// file of mode 0600 is readable and writable but cannot be run, one of mode
// 0400 cannot be written but by uid 0. The C library's access() looked at
// the rights of the preopened directory alone, which every file shares, and
// never at X_OK: it said yes to all of them (Test-Permissions).
func TestFiletestAccessAsksTheSystem(t *testing.T) {
	dir := t.TempDir()
	rw, ro := filepath.Join(dir, "rw"), filepath.Join(dir, "ro")
	for path, mode := range map[string]os.FileMode{rw: 0o600, ro: 0o400} {
		if err := os.WriteFile(path, nil, mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
	}
	out, status := hostPerl(t, "-e", `
use filetest 'access';
my ($rw, $ro) = @ARGV;
print join(",", map { $_ ? 1 : 0 } -r $rw, -w $rw, -x $rw, -r $ro, -w $ro, -e "$rw.missing"), "\n";
`, rw, ro)
	want := "1,1,0,1,0,0\n"
	if os.Geteuid() == 0 {
		want = "1,1,0,1,1,0\n"
	}
	if status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}
