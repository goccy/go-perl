//go:build darwin || linux

package gperl_test

import (
	"os"
	"os/exec"
	"testing"
)

// perl's System V IPC is the host's: semget, semop and semctl (SETVAL,
// GETVAL, GETALL, IPC_STAT with IPC::Semaphore::stat's layout), msgget,
// msgsnd, msgrcv and msgctl, shmget, shmwrite, shmread (also past the end
// of the segment) and shmctl, and IPC::SysV's ftok. gperl had none of them
// (IPC::SysV was not built: "Your vendor has not defined IPC::SysV macro
// SEM_UNDO", IPC::Shareable and the modules over it). What is expected is
// what perl 5.36 on Linux prints.
func TestSystemVIPCIsTheHosts(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-e", sysvProgram, os.Args[0])
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	want := "sem: val 3, all 0 3, nsems 2, mode 600\n" +
		"msg: queued 1, type 7, text hello\n" +
		"shm: read 0,115,101,103,109,101,110,116,0, segsz 4096\n" +
		"shm: past the end failed: EFAULT\n" +
		"ftok: a key\n"
	if err != nil || string(out) != want {
		t.Fatalf("err %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}

// A segment cannot be attached to the interpreter's memory: IPC::SysV's
// shmat, and shmdt, memread and memwrite, which take what shmat gives, die
// saying so (perl's own shmread and shmwrite copy instead).
func TestSystemVIPCSegmentsAreNotAttached(t *testing.T) {
	program := `use IPC::SysV qw(IPC_PRIVATE IPC_CREAT IPC_RMID shmat);
my $shm = shmget(IPC_PRIVATE, 4096, IPC_CREAT | 0600) // die "shmget: $!";
my $addr = eval { shmat($shm, undef, 0) };
print $@;
shmctl($shm, IPC_RMID, 0) or die "IPC_RMID: $!";`
	cmd := exec.Command(os.Args[0], "-e", program)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	want := "shmat: a shared memory segment cannot be attached to the interpreter's memory under gperl; " +
		"use shmread and shmwrite at -e line 3.\n"
	if err != nil || string(out) != want {
		t.Fatalf("err %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}

const sysvProgram = `use strict;
use warnings;
use IPC::SysV qw(IPC_PRIVATE IPC_CREAT IPC_RMID IPC_STAT SETVAL GETVAL GETALL SEM_UNDO ftok);
use IPC::Semaphore;
use IPC::SharedMem;
use IPC::Msg;

my $sem = semget(IPC_PRIVATE, 2, IPC_CREAT | 0600) // die "semget: $!";
semctl($sem, 1, SETVAL, 5) or die "SETVAL: $!";
semop($sem, pack("s!3", 1, -2, SEM_UNDO)) or die "semop: $!";
my $val = semctl($sem, 1, GETVAL, 0);
my $all = '';
semctl($sem, 0, GETALL, $all) or die "GETALL: $!";
my $data = '';
semctl($sem, 0, IPC_STAT, $data) or die "IPC_STAT: $!";
my $stat = IPC::Semaphore::stat::->new->unpack($data);
print "sem: val $val, all @{[unpack 's!*', $all]}, nsems ", $stat->nsems, ", mode ", sprintf("%o", $stat->mode & 0777), "\n";
semctl($sem, 0, IPC_RMID, 0) or die "IPC_RMID: $!";

my $msg = msgget(IPC_PRIVATE, IPC_CREAT | 0600) // die "msgget: $!";
msgsnd($msg, pack("l! a*", 7, "hello"), 0) or die "msgsnd: $!";
my $mdata = '';
msgctl($msg, IPC_STAT, $mdata) or die "IPC_STAT: $!";
my $qnum = IPC::Msg::stat::->new->unpack($mdata)->qnum;
my $buf;
msgrcv($msg, $buf, 64, 0, 0) or die "msgrcv: $!";
my ($type, $text) = unpack("l! a*", $buf);
print "msg: queued $qnum, type $type, text $text\n";
msgctl($msg, IPC_RMID, 0) or die "IPC_RMID: $!";

my $shm = shmget(IPC_PRIVATE, 4096, IPC_CREAT | 0600) // die "shmget: $!";
shmwrite($shm, "segment", 100, 7) or die "shmwrite: $!";
my $read;
shmread($shm, $read, 99, 9) or die "shmread: $!";
my $sdata = '';
shmctl($shm, IPC_STAT, $sdata) or die "IPC_STAT: $!";
print "shm: read ", join(",", map { ord } split //, $read), ", segsz ", IPC::SharedMem::stat::->new->unpack($sdata)->segsz, "\n";
my $past = shmread($shm, $read, 4090, 7) ? "read" : "failed: " . ($! + 0 == 14 ? "EFAULT" : $!);
print "shm: past the end $past\n";
shmctl($shm, IPC_RMID, 0) or die "IPC_RMID: $!";

my $key = ftok($ARGV[0], 1);
print "ftok: ", (defined $key && $key == ftok($ARGV[0], 1) ? "a key" : "none"), "\n";`
