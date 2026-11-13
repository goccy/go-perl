//go:build darwin || linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// For a program that is the host process, a descriptor is the same number
// for the program and for native code, in both directions: native code
// uses the number of a handle of the program's with the host's system
// calls, finds it in the blocking mode the program set, and a descriptor
// native code made is the program's to open a handle on, to wait for, to
// duplicate and to close.
func TestNativeCodeAndTheProgramShareDescriptors(t *testing.T) {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, HostProcess: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { p.Close() })
	perlAPIInstance(t, p)
	dir := t.TempDir()
	setup := `use strict; use warnings; use IO::Handle; use Fcntl; use POSIX ();
our $dir = '` + dir + `';
our ($fh, $r, $w, $nr, $nw, $nrh, $sa, $sb, $sah);
1;`
	if r, err := p.Eval(context.Background(), setup); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		// from the program to native code
		{"the number of a handle", `open($fh, "+>", "$dir/a.txt") or die "open: $!"; PerlAPI::fd_of($fh) == fileno($fh) ? "the program's" : "another"`, "the program's"},
		{"what native code finds under it", `PerlAPI::fd_state(fileno($fh))`, "file blocking"},
		{"a write by native code", `PerlAPI::fd_write(fileno($fh), "native,"); syswrite($fh, "perl"); sysseek($fh, 0, 0); sysread($fh, my $all, 64); $all`, "native,perl"},
		{"a pipe of the program's", `pipe($r, $w) or die "pipe: $!"; join "; ", PerlAPI::fd_state(fileno($r)), PerlAPI::fd_state(fileno($w))`, "pipe blocking; pipe blocking"},
		{"the mode the program sets", `$r->blocking(0); PerlAPI::fd_state(fileno($r))`, "pipe nonblocking"},
		{"a read by native code that finds nothing", `defined PerlAPI::fd_read(fileno($r), 8) ? "read" : ($!{EAGAIN} ? "EAGAIN" : "error $!")`, "EAGAIN"},
		{"the mode the program sets back", `$r->blocking(1); PerlAPI::fd_state(fileno($r))`, "pipe blocking"},
		{"a read by native code", `syswrite($w, "from perl"); PerlAPI::fd_read(fileno($r), 64)`, "from perl"},
		{"a socket of the program's", `use Socket; socketpair(my $x, my $y, AF_UNIX, SOCK_STREAM, 0) or die "socketpair: $!"; syswrite($x, "over a socket"); join "; ", PerlAPI::fd_state(fileno($y)), PerlAPI::fd_read(fileno($y), 64)`, "socket blocking; over a socket"},
		{"a handle that is no descriptor of the process", `PerlAPI::fd_of(\*STDOUT)`, "-1"},

		// from native code to the program
		{"a handle on a descriptor native code made", `($nr, $nw) = PerlAPI::fd_pair("pipe"); open($nrh, "<&=", $nr) or die "fdopen: $!"; PerlAPI::fd_write($nw, "from native\n"); fileno($nrh) == $nr ? scalar(<$nrh>) : "another number"`, "from native\n"},
		{"IO::Handle->new_from_fd", `my $h = IO::Handle->new_from_fd($nw, "w") or die "new_from_fd: $!"; $h->autoflush(1); print {$h} "through a handle\n"; my $line = <$nrh>; $h->close; $line . PerlAPI::fd_state($nw)`, "through a handle\nclosed"},
		{"select", `($nr, $nw) = PerlAPI::fd_pair("pipe"); my $rin = ""; vec($rin, $nr, 1) = 1; my $before = select(my $rout = $rin, undef, undef, 0); PerlAPI::fd_write($nw, "x"); my $after = select($rout = $rin, undef, undef, 5); "$before $after " . vec($rout, $nr, 1)`, "0 1 1"},
		{"fcntl", `open(my $h, "<&=", $nr) or die "fdopen: $!"; my $flags = fcntl($h, F_GETFL, 0) // die "F_GETFL: $!"; fcntl($h, F_SETFL, $flags | O_NONBLOCK) or die "F_SETFL: $!"; my $state = PerlAPI::fd_state($nr); close($h); "$state; " . PerlAPI::fd_state($nr)`, "pipe nonblocking; closed"},
		{"POSIX::dup2 and POSIX::close", `my $to = POSIX::dup2($nw, 30) // die "dup2: $!"; my $state = PerlAPI::fd_state(30); POSIX::close(30); POSIX::close($nw); "$to: $state; " . PerlAPI::fd_state(30) . "; " . PerlAPI::fd_state($nw)`, "30: pipe blocking; closed; closed"},
		{"a socket native code made", `($sa, $sb) = PerlAPI::fd_pair("socket"); open($sah, "+<&=", $sa) or die "fdopen: $!"; PerlAPI::fd_write($sb, "ping"); sysread($sah, my $got, 4); syswrite($sah, "pong"); $got . PerlAPI::fd_read($sb, 4)`, "pingpong"},
		{"closing the program's handle", `close($sah); my $state = PerlAPI::fd_state($sa); PerlAPI::fd_close($sb); $state`, "closed"},

		// handles native code opens on descriptors
		{"PerlIO_fdopen", `pipe(my $pr, my $pw) or die "pipe: $!"; syswrite($pw, "a line\nand more"); my $n = POSIX::dup(fileno($pr)); my $line = PerlAPI::fd_line($n); "$line; " . PerlAPI::fd_state($n)`, "a line; closed"},
		{"PerlIO_importFILE", `my ($ir, $iw) = PerlAPI::fd_pair("pipe"); my $n = PerlAPI::fd_import_write($iw, "imported"); my $got = PerlAPI::fd_read($ir, 64); PerlAPI::fd_close($ir); "$n $got; " . PerlAPI::fd_state($iw)`, "8 imported; closed"},
	})
}

// A forked process starts with its parent's descriptors under the same
// numbers, for native code as for the program. Where the child puts a file
// of its own under a number its parent still uses, the child's handle has
// no number of the process's — native code is told so, and does not reach
// the parent's file — while the parent's handle keeps its number.
func TestNativeCodeInAForkedProcessSeesItsOwnDescriptors(t *testing.T) {
	so := buildPerlAPI(t)
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var out, errOut bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, Fork: true, HostProcess: true, Stdout: &out, Stderr: &errOut})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	loadXSModule(t, p, "PerlAPI", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "PerlAPI.pm"),
		[]byte("package PerlAPI;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	status, err := p.RunProgram(context.Background(), perl.Program{Inc: []string{libdir}, Code: []string{
		`use strict; use warnings; use PerlAPI; use POSIX (); $| = 1;`,
		`my $dir = '` + dir + `';`,
		`open(my $shared, ">", "$dir/shared.txt") or die "open: $!";`,
		`my $number = fileno($shared);`,
		`pipe(my $r, my $w) or die "pipe: $!";`,
		`my $pid = fork() // die "fork: $!";`,
		`if (!$pid) {`,
		`    close($r);`,
		`    print {$w} "child, inherited: ", (PerlAPI::fd_of($shared) == $number ? "the parent's number" : "another"), ", ", PerlAPI::fd_state($number), "\n";`,
		`    PerlAPI::fd_write($number, "child,");`,
		`    open(my $own, ">", "$dir/own.txt") or die "open: $!";`,
		`    POSIX::dup2(fileno($own), $number) // die "dup2: $!";`,
		`    print {$w} "child, its own file under the number: ", PerlAPI::fd_of($shared), "\n";`,
		`    syswrite($shared, "the child's own");`,
		`    exit 0;`,
		`}`,
		`close($w);`,
		`print while <$r>;`,
		`waitpid($pid, 0);`,
		`print "child status: $?\n";`,
		`print "parent: ", (PerlAPI::fd_of($shared) == $number ? "its number" : "another"), "\n";`,
		`PerlAPI::fd_write($number, "parent");`,
		`for my $name ("shared.txt", "own.txt") { open(my $in, "<", "$dir/$name") or die "$name: $!"; print "$name: ", scalar(<$in>), "\n" }`,
	}})
	if err != nil {
		p.Close()
		t.Fatalf("RunProgram: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	want := strings.Join([]string{
		"child, inherited: the parent's number, file blocking",
		"child, its own file under the number: -1",
		"child status: 0",
		"parent: its number",
		"shared.txt: child,parent",
		"own.txt: the child's own",
	}, "\n") + "\n"
	if status != 0 || out.String() != want || errOut.Len() != 0 {
		t.Fatalf("status %d\nstdout:\n%s\nwant:\n%s\nstderr:\n%s", status, out.String(), want, errOut.String())
	}
}
