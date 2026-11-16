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
		// handles opened and closed in turn (the next one at the address
		// of the last, with a :stdio stream where the last one's was):
		// each gets a stream of its own descriptor (GD's newFromPng(file)
		// called twice)
		{"handles in turn", `my @got; for my $n (1 .. 3) { my $f = "` + file + `.$n"; open my $fh, ">", $f or die $!; PerlAPI::stdio_of($fh, "c$n"); close $fh; open my $in, "<", $f or die $!; push @got, scalar <$in> } join ",", @got`, "c1,c2,c3"},
		// a handle with no descriptor (an in-memory file): a stream over
		// the handle's functions, with no descriptor either
		{"over the handle", `open my $fh, ">", \my $buf or die $!; print $fh "perl1,"; my $r = PerlAPI::stdio_of($fh, "c1,"); print $fh "perl2"; close $fh; join " ", $r, $buf`, "nofd:same perl1,c1,perl2"},
	})
}

// An operation's arguments reach the interpreter whole whatever their size
// (a string longer than the operation box goes the loader's way), and an
// operation that runs Perl code which calls native code (a tied hash's
// STORE and FETCH) gets its own result back after the operations of that
// code.
func TestNativeOperationArguments(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a short key", `my %h; PerlAPI::store_fetch(\%h, "k" x 10) . ":" . join(",", map { length } keys %h)`, "stored:10:10"},
		{"a key as long as the box", `my %h; PerlAPI::store_fetch(\%h, "k" x 4096)`, "stored:4096"},
		{"a key longer than the box", `my %h; my $r = PerlAPI::store_fetch(\%h, "k" x 10000); $r . ":" . (exists $h{"k" x 10000} ? "there" : "missing")`, "stored:10000:there"},
		{"native code inside the operation", `package NestTie { sub TIEHASH { bless {}, shift } sub STORE { my ($s, $k, $v) = @_; $s->{$k} = $v . ":" . PerlAPI::numcmp(2, 1) } sub FETCH { my ($s, $k) = @_; $s->{$k} . ":" . PerlAPI::numcmp(1, 2) } } tie my %h, "NestTie"; PerlAPI::store_fetch(\%h, "key")`, "stored:3:1:-1"},
	})
}

// A MAGIC's svt_free runs while the SV still has it, as in perl's mg_free:
// a hook that looks its MAGIC up with mg_findext finds it (Linux::Epoll's
// clears its vtable that way).
func TestNativeFreeHookFindsItsMagic(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"freed scalars", `for (1 .. 3) { my $x = "v"; PerlAPI::attach_free_hook(\$x) } PerlAPI::free_hook_found()`, "3"},
	})
}

// Native code that makes another pad the one in use for a scope, saving
// PL_comppad, setting it, and only then saving PL_curpad (Exporter::
// Lexical), leaves the pad and the values of the code that called it when
// the scope ends: the caller's lexicals are its own again, and a lexical
// sub given to the code being compiled is there.
func TestNativePadMadeTheOneInUse(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the pad of another sub", `my $x = "mine"; my @a = (1, 2); my $s = sub { my ($y, $z) = (3, 4); $y + $z }; my $n = PerlAPI::pad_switch($s); join ",", $n > 2 ? "pad" : $n, $x, @a, $s->()`, "pad,mine,1,2,7"},
		{"a lexical sub at compile time", `sub lexfoo { "global" } my $r = eval q{ my $before = "kept"; BEGIN { PerlAPI::lexical_import("lexfoo", sub { "lexical" }) } 1; join ",", lexfoo(), $before } || $@; join ",", $r, lexfoo()`, "lexical,kept,global"},
	})
}

// looks_like_number answers as perl's does: grok_number's flags for a
// string (Number::Closest::XS takes "1.3" for a fraction by them), the
// flags of the numbers a value holds otherwise.
func TestNativeLooksLikeNumberFlags(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"strings and numbers", `PerlAPI::number_kinds("12", "1.3", "-5", "abc", "Inf", 12, 1.5)`, "u,uf,un,0,fi,x,x"},
	})
}

// SvUTF8 is SVf_UTF8 when set, so that code that passes it on in the flags
// of a call gets characters (Sub::Name, Sub::Metadata), and a name may hold
// NULs: a stash, a glob and an inserted string keep all their bytes.
func TestNativeNamesAsGiven(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"characters", `my $s = "Snow\x{2603}Man"; my ($copy, $stash, $flag) = PerlAPI::utf8_passed_on($s); join ",", $copy eq $s ? "copy" : "copy differs", $stash eq $s ? "stash" : "stash differs", $flag, exists $main::{"Snow\x{2603}Man::"} ? "exists" : "missing"`, "copy,stash,1,exists"},
		{"bytes", `my ($copy, $stash, $flag) = PerlAPI::utf8_passed_on("caf\xe9"); join ",", length $copy, $stash eq "caf\xe9" ? "stash" : "stash differs", $flag`, "4,stash,0"},
		{"with NULs", `my ($gv, $s) = PerlAPI::names_with_nul(); join ",", (exists $main::{"Nul\0Stash::"} ? "stash" : "no stash"), length(*{$gv}{NAME}), *{$gv}{NAME} eq "na\0me" ? "glob" : "glob differs", length $s, $s eq "ax\0yb" ? "insert" : "insert differs"`, "stash,5,glob,5,insert"},
	})
}

// load_module is `use NAME VERSION LIST`: the import gets the arguments,
// and while code is compiled it changes that code's hints (Quote::Code
// loads charnames as it reads "\N{...}"); PERL_LOADMOD_NOIMPORT leaves the
// import out, PERL_LOADMOD_DENY calls unimport, and a version the module
// does not have dies.
func TestNativeLoadModule(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the import's arguments", `BEGIN { PerlAPI::load_with(0, "constant", undef, "LOADED_C", 42) } LOADED_C()`, "42"},
		{"the hints of the code compiled", `my $in = eval q{ BEGIN { PerlAPI::load_with(0, "integer", undef) } 7 / 2 }; my $out = 7 / 2; "$in,$out"`, "3,3.5"},
		{"charnames", `my $r = eval q{ my $h; BEGIN { PerlAPI::load_with(0, "charnames", undef, ":full"); $h = exists $^H{charnames} ? "set" : "unset" } $h }; $r // $@`, "set"},
		{"no import", `my $in = eval q{ BEGIN { PerlAPI::load_with(2, "integer", undef) } 7 / 2 }; $in`, "3.5"},
		{"unimport", `my $in = eval q{ use integer; BEGIN { PerlAPI::load_with(1, "integer", undef) } 7 / 2 }; $in`, "3.5"},
		{"a version", `BEGIN { PerlAPI::load_with(0, "constant", "1.0", "VERSIONED_C", 7) } my $r = eval q{ BEGIN { PerlAPI::load_with(0, "constant", "99", "TOO_NEW_C", 1) } 1 }; join ",", VERSIONED_C(), defined $r ? "loaded" : ($@ =~ /version 99 required/ ? "too new" : $@)`, "7,too new"},
	})
}

// Every SV** result (av_fetch, hv_fetch) holds its value until the native
// call returns, however many the call asks for: code that keeps more than
// the frame's own slots and reads them after the last fetch
// (Number::Closest::XS keeping the closest elements of a long array) reads
// what it fetched.
func TestNativeFetchedSlotsKept(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a few", `PerlAPI::keep_fetched([1 .. 3], {a => "x"}, ["a"])`, "1,2,3|x"},
		{"more than a block", `my @a = map { "e$_" } 1 .. 600; my %h = map { ("k$_" => "v$_") } 1 .. 300; my $r = PerlAPI::keep_fetched(\@a, \%h, [map { "k$_" } 1 .. 300]); $r eq join(",", @a) . "|" . join(",", map { "v$_" } 1 .. 300) ? "all" : substr($r, 0, 80)`, "all"},
		{"in turn", `my $ok = 0; for my $n (1 .. 50) { $ok++ if PerlAPI::keep_fetched([1 .. 40], {}, []) eq join(",", 1 .. 40) } $ok`, "50"},
	})
}

// What a sub called in void context leaves on the stack is left alone, as
// perl leaves it: a callback can leave values the scope of its call has
// freed by the time call_sv returns (Future::XS's callbacks end with a bare
// return after nested calls), and the caller does not read them. The call
// took a reference on each result and dropped it at the next FREETMPS,
// freeing such a value a second time (Future's "weakself retention" case
// died in FREETMPS: "memory access out of bounds").
func TestNativeVoidCallLeavesItsResults(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"values freed by the call's scope", `PerlAPI::void_calls(200)`, "ok"},
	})
}

// AvARRAY is the array's memory as it is, past its last element too: code
// that takes the elements out by lowering AvFILLp in the same expression
// that reads the slot (`AvARRAY(av)[AvFILLp(av)--]`, Future::XS) reads the
// slot it has just left behind, which perl leaves in place. The slots past
// the last element read as NULL, and Future::XS dereferenced it.
func TestNativeSlotsPastTheLastElement(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"values", `my @a = ("a", "b", "c"); my $r = PerlAPI::take_out_by_fill(\@a); "$r|" . scalar(@a)`, "c,b,a|0"},
		{"native pointers", `PerlAPI::native_pointers_by_fill(5)`, "15"},
		{"in turn", `my $ok = 0; for (1 .. 50) { $ok++ if PerlAPI::native_pointers_by_fill(40) == 820 } $ok`, "50"},
	})
}

// A key whose length is negative is UTF-8, for every keyed hash function
// that takes a length: hv_store, hv_fetch and hv_exists find the key of
// characters, and hv_exists with the same bytes as a key of bytes does not.
// hv_exists took the negative length for the length of the key it made, and
// the interpreter panicked ("sv_setpvn_fresh called with negative strlen").
func TestNativeUTF8KeyLength(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"store, fetch, exists", `use utf8; my %h; my $r = PerlAPI::utf8_key_ops(\%h); join ",", $r, map { length($_) . (exists $h{"füü"} ? "found" : "missing") } keys %h`, "e1 b0 f1,3found"},
	})
}

// SvIsUV is the flag itself: a string above UV_MAX made a number keeps its
// integer (UV_MAX) privately, with the flag and without SvIOK, so SvIsUV is
// true and SvUOK false. SvIsUV answered false then, and Cpanel::JSON::XS
// took such a number for -1.
func TestNativeIsUVWithoutIOK(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"above UV_MAX", `PerlAPI::iv_then_isuv("18446744073709551616")`, "isuv=1 uok=0 uv=18446744073709551615"},
		{"above IV_MAX", `PerlAPI::iv_then_isuv("18446744073709551615")`, "isuv=1 uok=1 uv=18446744073709551615"},
		{"negative", `PerlAPI::iv_then_isuv("-5")`, "isuv=0 uok=0 uv=18446744073709551611"},
	})
}

// The interpreter's variables and structures that Text::CSV_XS and
// Package::Stash::XS reach: the output record separator (PL_ors_sv), set
// aside around a print and put back; the name of a stash (hv_name_set),
// given and taken away; the CV of a glob set with GvCV_set, whose value is
// the CV.
func TestNativeStashAndRecordSeparator(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"ors set aside", `my $out = ""; open my $fh, ">", \$out; { local $\ = "!"; PerlAPI::ors_aside(sub { print $fh "a"; }); print $fh "b"; } print $fh "c"; close $fh; $out`, "ab!c"},
		{"ors read", `local $\ = "#"; my $out = ""; open my $fh, ">", \$out; PerlAPI::ors_aside(sub { $\ = "x" }); print $fh "a"; close $fh; $out`, "a#"},
		{"stash named", `my ($obj, $left) = PerlAPI::named_stash("Some::Anon"); join ",", ref($obj), $left`, "Some::Anon,none"},
		{"sub installed", `my $ok = PerlAPI::install_sub("main::installed_by_native", sub { 42 }); require B; join ",", $ok ? "set" : "unset", main::installed_by_native(), (B::svref_2object(\*main::installed_by_native)->GvFLAGS & 0x80) ? "imported" : "own"`, "set,42,imported"},
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
