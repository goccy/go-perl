package internal

import "testing"

// The name of the architecture is the host's processor and system as perl's
// Configure spells them there, marked as go-perl's.
func TestArchNameSpellsTheHostAsPerlDoes(t *testing.T) {
	for _, c := range []struct{ goos, goarch, want string }{
		{"linux", "arm64", "aarch64-linux-goperl"},
		{"linux", "amd64", "x86_64-linux-goperl"},
		{"linux", "386", "i686-linux-goperl"},
		{"linux", "ppc64le", "ppc64le-linux-goperl"},
		{"linux", "riscv64", "riscv64-linux-goperl"},
		{"darwin", "arm64", "arm64-darwin-goperl"},
		{"darwin", "amd64", "x86_64-darwin-goperl"},
		{"freebsd", "amd64", "amd64-freebsd-goperl"},
		{"openbsd", "arm64", "arm64-openbsd-goperl"},
		{"netbsd", "amd64", "x86_64-netbsd-goperl"},
		{"illumos", "amd64", "amd64-solaris-goperl"},
		{"windows", "amd64", "MSWin32-x64-goperl"},
	} {
		if got := archName(c.goos, c.goarch); got != c.want {
			t.Errorf("archName(%s, %s) = %q, want %q", c.goos, c.goarch, got, c.want)
		}
	}
}
