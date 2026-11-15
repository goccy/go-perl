//go:build darwin || linux

package xs_test

import (
	"os"
	"path/filepath"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// Native code that takes the parser away while it builds an op (Devel::
// CallParser's SAVEVPTR(PL_parser); PL_parser = NULL) takes it away from
// the interpreter too: an op checker of native code finds no parser then,
// and the parser is back when the scope ends.
func TestNativeParserTakenAway(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"while compiling", `my $r; BEGIN { $r = PerlAPI::parser_hidden() } $r`, "before:1 checker:0 after:1"},
	})
}

// A save-stack destructor of native code that dies makes the end of the
// scope die, as a destructor of perl's own does (Scope::Cleanup): the die
// reaches the eval around the scope. The scope of an XSUB ends when it
// returns (perl's entersub runs it in a scope of its own).
func TestNativeDestructorDies(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"at the end of the call", `my @seen; my $r = eval { push @seen, "before"; PerlAPI::die_in_destructor("cleanup died\n"); push @seen, "after"; 1 }; join ",", @seen, defined $r ? "returned" : "died: $@"`, "before,died: cleanup died\n"},
		{"the message gets the place", `eval { PerlAPI::die_in_destructor("no newline"); 1 }; $@ =~ /^no newline at \(eval \d+\) line 1\.$/ ? "placed" : $@`, "placed"},
		{"a scope that ends normally goes on", `my $r = eval { { 1 } "fine" }; $r`, "fine"},
	})
}

// The fields of an IO and the IO of a glob are lvalues, as Linux::FD,
// Tickit and FCGI assign them: a handle made field by field reads the
// descriptor it was given.
func TestNativeIOFieldsAssigned(t *testing.T) {
	// the descriptor is one of this process's, as the program's are
	p := newHostProcessPerl(t)
	perlAPIInstance(t, p)
	file := filepath.Join(t.TempDir(), "lines.txt")
	if err := os.WriteFile(file, []byte("first\nsecond\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runEvalCases(t, p, []evalCase{
		{"made field by field", `open my $fh, "<", "` + file + `" or die $!; my ($fields, $h) = PerlAPI::io_on_fd(fileno($fh), "<"); my $line = <$h>; chomp $line; join ",", $fields, ref($h), $line`, "<1,IO::Handle,first"},
		{"a new IO", `PerlAPI::glob_with_new_io()`, "0:IO::File"},
	})
}

// sv_numcmp (overloading included), SAVECOPSTASH_FREE, and what modules
// copied out of perl's sources use: GvNAME_HEK, FUNCTION__, the argument
// checks of proto.h, and the offsets of a match as RX_OFFSp names them.
func TestNativeComparisonsAndCopiedCode(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"numbers", `join ",", PerlAPI::numcmp(10, 9), PerlAPI::numcmp("2", "10"), PerlAPI::numcmp(3, 3.0)`, "1,-1,0"},
		{"overloaded", `package NumOv { use overload "<=>" => sub { my ($a, $b, $swap) = @_; my $r = $a->{n} <=> (ref $b ? $b->{n} : $b); $swap ? -$r : $r }, fallback => 1 } PerlAPI::numcmp(bless({n => 5}, "NumOv"), 7)`, "-1"},
		{"the stash of the statement", `PerlAPI::cop_stash_scope("Some::Where") . ":" . __PACKAGE__`, "Some::Where:main:main"},
		{"copied code", `our $glob_name; PerlAPI::names_and_offsets(\*glob_name, qr/(b+)/, "abbc")`, "glob_name|XS_PerlAPI_names_and_offsets|0|1:1-3"},
	})
}

// The stdio stream of a handle (PerlIO_findFILE) is one on the handle's
// descriptor, as perl's is, for a program that is the host process: the
// handle gets a :stdio layer, what the C code writes reaches the file in
// order with what the program writes, the same stream is found again, and
// PerlIO_releaseFILE takes the layer off.
func TestNativeStdioStreamOfAHandle(t *testing.T) {
	p := newHostProcessPerl(t)
	perlAPIInstance(t, p)
	file := filepath.Join(t.TempDir(), "out.txt")
	code := `open my $fh, ">", "` + file + `" or die $!; print $fh "perl1,"; my $r = PerlAPI::stdio_of($fh, "c1,"); print $fh "perl2,"; my $layers = join "+", PerlIO::get_layers($fh); PerlAPI::stdio_release($fh); my $after = join "+", PerlIO::get_layers($fh); print $fh "perl3"; close $fh; open my $in, "<", "` + file + `" or die $!; my $text = <$in>; join " ", $r, $layers, $after, $text`
	runEvalCases(t, p, []evalCase{
		{"on the descriptor", code, "fd:same unix+perlio+stdio unix+perlio perl1,c1,perl2,perl3"},
		// a handle with no descriptor (an in-memory file): a stream over
		// the handle's functions, with no descriptor either
		{"over the handle", `open my $fh, ">", \my $buf or die $!; print $fh "perl1,"; my $r = PerlAPI::stdio_of($fh, "c1,"); print $fh "perl2"; close $fh; join " ", $r, $buf`, "nofd:same perl1,c1,perl2"},
	})
}

// pv_uni_display is perl's: with UNI_DISPLAY_QQ a backslash and the
// controls that have a mnemonic are escaped, a double quote is not; the
// flags have utf8.h's values.
func TestNativeUniDisplay(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"flags", `PerlAPI::uni_display_flags()`, "7,3"},
		{"qq", `PerlAPI::pv_display(qq{a"b\\c\n\x01}, 7)`, `a"b\\c\n\x{1}`},
		{"utf-8", `my $s = "\x{263a}!"; utf8::encode($s); PerlAPI::pv_display($s, 7)`, `\x{263a}!`},
	})
}

// newHostPerl's interpreter as the host process: its descriptors are this
// process's.
func newHostProcessPerl(t testing.TB) *perl.Perl {
	t.Helper()
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, HostProcess: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { p.Close() })
	return p
}
