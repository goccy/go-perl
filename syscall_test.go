//go:build linux && (arm64 || amd64)

package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// syscall() is the host's raw system call for a program that is the host
// process: numbers as they are, a string as the address of its buffer
// (getrandom fills one, write sends one), and an error as $!. Rand::Urandom,
// Sys::GetRandom::PP, Linux::Landlock and Sys::Linux::Syscall::Execve call
// it; it died "syscall() unimplemented". What is expected is what perl 5.36
// on Linux prints.
func TestSyscallIsTheHostsSystemCall(t *testing.T) {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("ExtractStdlib: %v", err)
	}
	r := runProgram(t, perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, HostProcess: true}, perl.Program{Code: code(
		`use strict; use warnings;`,
		`use Config;`,
		`my %nr = $Config{archname} =~ /^aarch64/`,
		`    ? (getpid => 172, getrandom => 278, write => 64, close => 57)`,
		`    : (getpid => 39, getrandom => 318, write => 1, close => 3);`,
		`my $file = "/tmp/syscall-test-$$";`,
		`my $pid = syscall($nr{getpid});`,
		`my $buf = "\0" x 16;`,
		`my $n = syscall($nr{getrandom}, $buf, 16, 0);`,
		`open my $fh, ">", $file or die "open: $!";`,
		`my $hello = "hello"; my $w = syscall($nr{write}, fileno($fh), $hello, 5);`,
		`close $fh;`,
		`open $fh, "<", $file or die;`,
		`my $got = <$fh>;`,
		`close $fh;`,
		`unlink $file;`,
		`my $c = syscall($nr{close}, 9999);`,
		`my $err = "$!";`,
		`print "getpid ", ($pid == $$ ? "is \$\$" : "$pid vs $$"), ", getrandom $n, write $w $got, close $c $err\n";`,
	)}, "")
	r.want(t, 0, "getpid is $$, getrandom 16, write 5 hello, close -1 Bad file descriptor\n")
}

// A program that is not the host process has no descriptors or identity of
// the process's for a raw system call to act on: syscall() fails with ENOSYS.
func TestSyscallOfAProgramOfItsOwnIsNotImplemented(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`my $r = syscall(39);`,
		`print "$r ", ($!{ENOSYS} ? "ENOSYS" : "$!"), "\n";`,
	)}, "")
	r.want(t, 0, "-1 ENOSYS\n")
}
