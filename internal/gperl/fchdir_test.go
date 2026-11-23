//go:build darwin || linux

package gperl_test

import "testing"

// chdir takes a directory handle (fchdir): the handle's directory becomes
// the working directory, wherever the program has gone since it opened it.
func TestChdirToADirectoryHandle(t *testing.T) {
	got := runForkScript(t, `
use Cwd qw(getcwd);
my $dir = $REPORT =~ s{/[^/]*$}{}r;
chdir $dir or die "chdir: $!";
my $base = getcwd();
sub where { getcwd() =~ s/^\Q$base\E//r || '/' }
mkdir $_ or die "mkdir $_: $!" for qw(a a/deep b);
open(my $made, '>', 'file') or die "create: $!"; close $made;

chdir 'a/deep' or die "chdir: $!";
opendir(my $deep, '.') or die "opendir: $!";
opendir(my $up, '..') or die "opendir: $!";
chdir "$base/b" or die "chdir: $!";
report("to a handle opened as '.': ", (chdir($deep) ? "ok" : "failed: $!"), " cwd=", where());
report("to a handle opened as '..': ", (chdir($up) ? "ok" : "failed: $!"), " cwd=", where());

# a file handle on a directory works like a directory handle
chdir $base or die "chdir: $!";
open(my $as_file, '<', 'b') or die "open a directory: $!";
report("to a file handle on a directory: ", (chdir($as_file) ? "ok" : "failed: $!"), " cwd=", where());

chdir $base or die "chdir: $!";
open(my $file, '<', 'file') or die "open: $!";
report("to a handle on a file: ", (chdir($file) ? "ok" : ($!{ENOTDIR} ? "ENOTDIR" : "error $!")), " cwd=", where());
closedir $deep;
{
    no warnings;
    report("to a closed handle: ", (chdir($deep) ? "ok" : ($!{EBADF} ? "EBADF" : "error $!")), " cwd=", where());
}

# The name a handle was opened under may come to mean another directory.
# The handle still means the one it opened, so the program never ends up in
# the other one.
chdir $base or die "chdir: $!";
opendir(my $moved, 'a') or die "opendir: $!";
rename 'a', 'c' or die "rename: $!";
mkdir 'a' or die "mkdir: $!";
my $ok = chdir($moved);
report("after its name was given to another directory: ", ($ok ? "ok" : "refused"), " cwd=", where());
`)
	wantLines(t, got,
		"to a handle opened as '.': ok cwd=/a/deep",
		"to a handle opened as '..': ok cwd=/a",
		"to a file handle on a directory: ok cwd=/b",
		"to a handle on a file: ENOTDIR cwd=/",
		"to a closed handle: EBADF cwd=/",
		"after its name was given to another directory: refused cwd=/",
	)
}
