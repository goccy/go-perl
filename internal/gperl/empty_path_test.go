//go:build darwin || linux

package gperl_test

import "testing"

// The empty string names no file. Every operation on it fails with ENOENT,
// as path resolution has it; none of them falls back to the current
// directory.
func TestEmptyPathNamesNoFile(t *testing.T) {
	got := runForkScript(t, `
use Fcntl;
no warnings;
my $dir = $REPORT =~ s{/[^/]*$}{}r;
chdir $dir or die "chdir $dir: $!";
open(my $made, '>', 'file') or die "create: $!"; close $made;
sub outcome { my ($ok) = @_; $ok ? "done" : $!{ENOENT} ? "ENOENT" : "error $!" }

report("open for reading: ", outcome(open(my $in, '<', '')));
report("open for writing: ", outcome(open(my $out, '>', '')));
report("open for appending: ", outcome(open(my $app, '>>', '')));
report("sysopen for reading: ", outcome(sysopen(my $r, '', O_RDONLY)));
report("sysopen creating: ", outcome(sysopen(my $c, '', O_WRONLY | O_CREAT, 0600)));
report("stat: ", outcome(scalar(my @s = stat(''))));
report("lstat: ", outcome(scalar(my @l = lstat(''))));
report("-e: ", outcome(-e ''));
report("-d: ", outcome(-d ''));
report("opendir: ", outcome(opendir(my $dh, '')));
report("mkdir: ", outcome(mkdir('')));
report("rmdir: ", outcome(rmdir('')));
report("unlink: ", outcome(unlink('')));
report("rename from: ", outcome(rename('', 'other')));
report("rename to: ", outcome(rename('file', '')));
report("chmod: ", outcome(chmod(0700, '')));
report("chown: ", outcome(chown(-1, -1, '')));
report("utime: ", outcome(utime(undef, undef, '')));
report("readlink: ", outcome(defined readlink('')));
report("symlink at: ", outcome(symlink('file', '')));
report("link to: ", outcome(link('file', '')));
report("truncate: ", outcome(truncate('', 0)));

# nothing above touched the directory or the file in it
report("the current directory by name: ", outcome(opendir(my $here, '.')));
report("the file is still a file: ", outcome(-f 'file'));
`)
	wantLines(t, got,
		"open for reading: ENOENT",
		"open for writing: ENOENT",
		"open for appending: ENOENT",
		"sysopen for reading: ENOENT",
		"sysopen creating: ENOENT",
		"stat: ENOENT",
		"lstat: ENOENT",
		"-e: ENOENT",
		"-d: ENOENT",
		"opendir: ENOENT",
		"mkdir: ENOENT",
		"rmdir: ENOENT",
		"unlink: ENOENT",
		"rename from: ENOENT",
		"rename to: ENOENT",
		"chmod: ENOENT",
		"chown: ENOENT",
		"utime: ENOENT",
		"readlink: ENOENT",
		"symlink at: ENOENT",
		"link to: ENOENT",
		"truncate: ENOENT",
		"the current directory by name: done",
		"the file is still a file: done",
	)
}
