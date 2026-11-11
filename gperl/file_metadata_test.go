package gperl_test

import (
	"testing"
)

// utime sets the access and the modification time of a file separately,
// undef means now, and a filehandle works like a name.
func TestUtimeSetsFileTimes(t *testing.T) {
	got := runForkScript(t, `
my $f = "$REPORT.data";
open(my $fh, '>', $f) or die "open: $!"; print {$fh} "x"; close $fh;
my ($atime, $mtime) = (time - 9000, time - 5000);
my $n = utime($atime, $mtime, $f);
my @st = stat $f;
report("by name: changed=$n atime=", ($st[8] == $atime ? 'set' : "not set ($st[8])"), " mtime=", ($st[9] == $mtime ? 'set' : "not set ($st[9])"));

report("ages: -M=", (-M $f > 0.05 ? 'old' : 'new'), " -A=", (-A $f > 0.1 ? 'older' : 'new'));

$n = utime(undef, undef, $f);
@st = stat $f;
report("undef: changed=$n now=", (abs($st[9] - time) < 60 && abs($st[8] - time) < 60 ? 1 : 0));

open($fh, '+<', $f) or die "open: $!";
$n = utime($atime, $mtime, $fh);
@st = stat $f;
report("by handle: changed=$n atime=", ($st[8] == $atime ? 'set' : 'not set'), " mtime=", ($st[9] == $mtime ? 'set' : 'not set'));
close $fh;
unlink $f;
`)
	wantLines(t, got,
		"by name: changed=1 atime=set mtime=set",
		"ages: -M=old -A=older",
		"undef: changed=1 now=1",
		"by handle: changed=1 atime=set mtime=set",
	)
}

// A created file or directory gets the mode it was created with, less the
// umask; umask returns the mask it replaces; and a forked child starts with
// its parent's mask and changes only its own.
func TestCreationModeFollowsUmask(t *testing.T) {
	got := runForkScript(t, `
use Fcntl qw(O_CREAT O_WRONLY);
my $d = "$REPORT.dir";
sub mode { sprintf "%o", (stat $_[0])[2] & 07777 }
umask(022);
mkdir($d, 0700) or die "mkdir: $!";

report("previous mask: ", sprintf("%o", umask(022)));
mkdir("$d/m750", 0750) or die "mkdir: $!";
mkdir("$d/m777", 0777) or die "mkdir: $!";
report("mkdir under 022: 0750 -> ", mode("$d/m750"), ", 0777 -> ", mode("$d/m777"));

umask(027);
sysopen(my $fh, "$d/s666", O_CREAT|O_WRONLY, 0666) or die "sysopen: $!"; close $fh;
sysopen($fh, "$d/s600", O_CREAT|O_WRONLY, 0600) or die "sysopen: $!"; close $fh;
open($fh, '>', "$d/o") or die "open: $!"; close $fh;
report("files under 027: 0666 -> ", mode("$d/s666"), ", 0600 -> ", mode("$d/s600"), ", open > -> ", mode("$d/o"));

# A mask looser than the one the host process runs under.
umask(0);
mkdir("$d/open", 0777) or die "mkdir: $!";
sysopen($fh, "$d/s666-open", O_CREAT|O_WRONLY, 0666) or die "sysopen: $!"; close $fh;
report("under 0: mkdir 0777 -> ", mode("$d/open"), ", file 0666 -> ", mode("$d/s666-open"));

# Opening an existing file with O_CREAT does not touch its mode.
sysopen($fh, "$d/s600", O_CREAT|O_WRONLY, 0666) or die "sysopen: $!"; close $fh;
report("existing file: ", mode("$d/s600"));

umask(077);
my $pid = fork;
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    my $inherited = umask(002);
    open(my $c, '>', "$d/child") or die "open: $!"; close $c;
    report("child: inherited ", sprintf("%o", $inherited), ", creates ", mode("$d/child"));
    exit 0;
}
waitpid($pid, 0);
open($fh, '>', "$d/parent") or die "open: $!"; close $fh;
report("parent: mask ", sprintf("%o", umask), ", creates ", mode("$d/parent"));

system("rm", "-rf", $d);
`)
	wantLines(t, got,
		"previous mask: 22",
		"mkdir under 022: 0750 -> 750, 0777 -> 755",
		"files under 027: 0666 -> 640, 0600 -> 600, open > -> 640",
		"under 0: mkdir 0777 -> 777, file 0666 -> 666",
		"existing file: 600",
		"child: inherited 77, creates 664",
		"parent: mask 77, creates 600",
	)
}

// chmod changes a mode through a name and through a filehandle.
func TestChmodByNameAndHandle(t *testing.T) {
	got := runForkScript(t, `
my $f = "$REPORT.data";
open(my $fh, '>', $f) or die "open: $!";
sub mode { sprintf "%o", (stat $_[0])[2] & 07777 }
report("by name: changed=", chmod(0640, $f), " mode=", mode($f));
report("by handle: changed=", chmod(0600, $fh), " mode=", mode($f), " handle sees=", mode($fh));
close $fh;
unlink $f;
`)
	wantLines(t, got,
		"by name: changed=1 mode=640",
		"by handle: changed=1 mode=600 handle sees=600",
	)
}

// stat of a name, lstat of it and stat of a handle open on it describe the
// same file in every field, with its real identity and link count — what
// File::Temp's unlink0 relies on to remove a temporary file safely.
func TestStatDescribesTheSameFileByNameAndHandle(t *testing.T) {
	got := runForkScript(t, `
use File::Temp qw(tempfile unlink0);
my $d = "$REPORT.dir";
mkdir($d) or die "mkdir: $!";
my $f = "$d/f";
open(my $fh, '+>', $f) or die "open: $!"; print {$fh} "hello"; $fh->flush;
chmod(0640, $f) or die "chmod: $!";

my @name = stat $f; my @link = lstat $f; my @handle = stat $fh;
report("stat fields: lstat ", ("@link" eq "@name" ? "same" : "differs"), ", handle ", ("@handle" eq "@name" ? "same" : "differs [@handle] [@name]"));
report("identity: dev ", ($name[0] ? "set" : "unset"), ", ino ", ($name[1] ? "set" : "unset"), ", nlink $name[3], size $name[7], -s ", -s $f);

open(my $other, '>', "$d/other") or die "open: $!"; close $other;
report("another file: ino ", ((stat "$d/other")[1] != $name[1] ? "differs" : "same"));

link($f, "$d/hard") or die "link: $!";
report("hard link: nlink ", (stat $f)[3], ", via handle ", (stat $fh)[3], ", ino ", ((stat "$d/hard")[1] == $name[1] ? "same" : "differs"));

open($other, '>', "$d/target") or die "open: $!"; print {$other} "old"; close $other;
rename("$d/hard", "$d/target") or die "rename: $!";
report("rename over a file: size ", -s "$d/target", ", source ", (-e "$d/hard" ? "left" : "gone"), ", nlink ", (stat $f)[3]);

truncate($fh, 2) or die "truncate: $!";
report("truncate by handle: size ", -s $f);

unlink("$d/target", $f) == 2 or die "unlink: $!";
report("unlinked but open: nlink ", (stat $fh)[3]);
close $fh;

my ($tfh, $tname) = tempfile(DIR => $d);
report("unlink0: ", (unlink0($tfh, $tname) ? "ok" : "failed"), ", file ", (-e $tname ? "left" : "gone"));
close $tfh;
system("rm", "-rf", $d);
`)
	wantLines(t, got,
		"stat fields: lstat same, handle same",
		"identity: dev set, ino set, nlink 1, size 5, -s 5",
		"another file: ino differs",
		"hard link: nlink 2, via handle 2, ino same",
		"rename over a file: size 5, source gone, nlink 2",
		"truncate by handle: size 2",
		"unlinked but open: nlink 0",
		"unlink0: ok, file gone",
	)
}

// flock locks belong to the open file: another open of the same file — in
// the same process, in a forked child, anywhere — competes for the lock,
// while a duplicate of the handle and the copy a forked child inherits
// share it.
func TestFlockExcludesOtherOpens(t *testing.T) {
	got := runForkScript(t, `
use Fcntl qw(:flock);
my $f = "$REPORT.lock";
open(my $holder, '>', $f) or die "open: $!";
open(my $other, '<', $f) or die "open: $!";
open(my $dup, '<&', $holder) or die "dup: $!";
sub try { flock($_[0], $_[1] | LOCK_NB) ? "granted" : ($!{EWOULDBLOCK} ? "would block" : "error $!") }

report("shared: holder ", try($holder, LOCK_SH), ", other shared ", try($other, LOCK_SH), ", other exclusive ", try($other, LOCK_EX));
flock($other, LOCK_UN) or die "unlock: $!";
report("exclusive: holder ", try($holder, LOCK_EX), ", its dup ", try($dup, LOCK_EX), ", other shared ", try($other, LOCK_SH));

my $pid = fork;
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    open(my $own, '<', $f) or die "open: $!";
    report("child: own open ", try($own, LOCK_SH), ", inherited handle ", try($holder, LOCK_EX));
    exit 0;
}
waitpid($pid, 0);
report("after the child exited: other ", try($other, LOCK_SH));

# A lock that waits is granted when the holder lets go.
pipe(my $r, my $w) or die "pipe: $!";
$pid = fork;
die "fork: $!" unless defined $pid;
if ($pid == 0) {
    close $r;
    open(my $own, '<', $f) or die "open: $!";
    syswrite($w, "x");
    flock($own, LOCK_SH) or die "flock: $!";
    report("child: waited for the lock");
    exit 0;
}
close $w;
sysread($r, my $ready, 1);
select(undef, undef, undef, 0.2);
report("parent: unlocking");
flock($dup, LOCK_UN) or die "unlock: $!";
waitpid($pid, 0);

report("unlocked through the dup: other ", try($other, LOCK_EX));
close $other;
report("after the other handle closed: holder ", try($holder, LOCK_EX));
close $_ for $holder, $dup;
unlink $f;
`)
	wantLines(t, got,
		"shared: holder granted, other shared granted, other exclusive would block",
		"exclusive: holder granted, its dup granted, other shared would block",
		"child: own open would block, inherited handle granted",
		"after the child exited: other would block",
		"parent: unlocking",
		"child: waited for the lock",
		"unlocked through the dup: other granted",
		"after the other handle closed: holder granted",
	)
}
