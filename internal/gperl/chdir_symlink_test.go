//go:build linux || darwin

package gperl_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// chdir finds the directory as the kernel does: through a symbolic link
// the program is in the link's target, which getcwd, Cwd::getcwd and
// Cwd::abs_path name, and ".." from there is the target's parent, not the
// link's (AnyEvent-MCP's worker refuses a working directory a link takes
// outside its root by comparing Cwd::getcwd with the root). What is
// expected is what perl does.
func TestChdirThroughASymbolicLinkIsInItsTarget(t *testing.T) {
	dir := t.TempDir()
	real, err := filepath.EvalSymlinks(dir) // the temporary directory may itself be under a link
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "real", "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	for link, target := range map[string]string{"link": "real", "sublink": "real/sub", "abslink": filepath.Join(real, "real", "sub")} {
		if err := os.Symlink(target, filepath.Join(dir, link)); err != nil {
			t.Fatal(err)
		}
	}
	program := `use Cwd (); use POSIX ();
for my $to (@ARGV) {
    chdir $to or die "chdir $to: $!";
    print join(" ", POSIX::getcwd(), Cwd::getcwd(), Cwd::abs_path(".")), "\n";
    chdir ".." or die "chdir ..: $!";
    print POSIX::getcwd(), "\n";
}`
	cmd := exec.Command(os.Args[0], "-e", program,
		filepath.Join(dir, "link"), filepath.Join(dir, "sublink"), filepath.Join(dir, "abslink"), dir+"/sublink/../sub")
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	r, s := filepath.Join(real, "real"), filepath.Join(real, "real", "sub")
	want := strings.Join([]string{
		r + " " + r + " " + r, real,
		s + " " + s + " " + s, r,
		s + " " + s + " " + s, r,
		s + " " + s + " " + s, r,
	}, "\n") + "\n"
	if err != nil || string(out) != want {
		t.Fatalf("err %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}
