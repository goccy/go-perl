//go:build darwin || linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// perlAPIInstance loads the PerlAPI fixture into p and makes it usable.
func perlAPIInstance(t *testing.T, p *perl.Perl) {
	t.Helper()
	so := buildPerlAPI(t)
	loadXSModule(t, p, "PerlAPI", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "PerlAPI.pm"),
		[]byte("package PerlAPI;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r, err := p.Eval(context.Background(), `unshift @INC, '`+libdir+`'; require PerlAPI; 1;`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
}

type evalCase struct{ name, code, want string }

func runEvalCases(t *testing.T, p *perl.Perl, cases []evalCase) {
	t.Helper()
	for _, tc := range cases {
		r, err := p.Eval(context.Background(), tc.code)
		if err != nil {
			t.Errorf("%s: Eval: %v", tc.name, err)
			continue
		}
		if r.Error != nil {
			t.Errorf("%s: died: %v", tc.name, r.Error)
			continue
		}
		if got := resultStr(r); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// The IV and UV of native code are 64 bits wide and the interpreter's 32:
// a pointer or a large integer a module stores with sv_setiv, newSViv or
// sv_setref_pv comes back whole from SvIV, SvUV, SvIVX and INT2PTR, and
// what the Perl program sees is the same number.
func TestNativeIntegersCrossAtFullWidth(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a heap pointer kept by plain casts", `PerlAPI::pointer_roundtrip()`, "same"},
		{"integers made beyond 32 bits", `join ",", PerlAPI::wide_made()`, "1099511627776,-1099511627776,3000000000,2199023255552,-5"},
		{"a small integer stays an integer", `my @v = PerlAPI::wide_made(); PerlAPI::flags($v[4])`, "IOK,"},
		{"SvIV and SvUV of an unsigned value", `PerlAPI::wide_values(4000000000)`, "4000000000/4000000000"},
		{"SvIV and SvUV of a large number", `PerlAPI::wide_values(5000000000000)`, "5000000000000/5000000000000"},
		{"SvIV and SvUV of a large negative number", `PerlAPI::wide_values(-5000000000000)`, "-5000000000000/" + strconv.FormatUint(uint64(1<<64-5000000000000), 10)},
		{"SvIV and SvUV of a numeric string", `PerlAPI::wide_values("123456789012")`, "123456789012/123456789012"},
		{"SvIV truncates a fraction", `PerlAPI::wide_values(-7.9)`, "-7/" + strconv.FormatUint(uint64(1<<64-7), 10)},
		{"SvIV of a small integer", `PerlAPI::wide_values(-3)`, "-3/" + strconv.FormatUint(uint64(1<<64-3), 10)},
		{"SvIV of undef", `no warnings; PerlAPI::wide_values(undef)`, "0/0"},
	})
}

// An XSUB takes and returns lists of any length, and so does a Perl sub it
// calls: the stack of the call grows.
func TestNativeLongLists(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a short argument list", `PerlAPI::sum_args(1 .. 10)`, "55"},
		{"5000 arguments", `PerlAPI::sum_args(1 .. 5000)`, "12502500"},
		{"5000 values after one EXTEND", `my @v = PerlAPI::count_up(5000, 0); scalar(@v) . ":$v[0]:$v[-1]"`, "5000:1:5000"},
		{"5000 values pushed one by one", `my @v = PerlAPI::count_up(5000, 1); scalar(@v) . ":$v[0]:$v[-1]"`, "5000:1:5000"},
		{"MARK after a callback with 3000 arguments", `my $n = 0; PerlAPI::callback_then_mark(sub { $n = @_ }, 7, (0) x 2999) . ",$n"`, "3008,3000"},
		{"3000 arguments to a callback that doubles them", `join ":", PerlAPI::relay_list(sub { map { ($_, $_ * 2) } @_ }, 3000)`, "6000:6000"},
	})
}

// The method-cache generation of a class changes when a method it inherits
// is replaced or its @ISA changes: a module that keeps a method it looked
// up relies on that to look again.
func TestNativeMethodCacheGeneration(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	setup := `package GenParent; sub m { 1 } package GenOther; sub m { 3 }
package GenChild; our @ISA = ("GenParent"); package main; 1;`
	if r, err := p.Eval(context.Background(), setup); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"unchanged while nothing is redefined", `GenChild->m; my $a = PerlAPI::method_cache_gen("GenChild"); GenChild->m; $a == PerlAPI::method_cache_gen("GenChild") ? "same" : "moved"`, "same"},
		{"a method of the parent is replaced", `my $a = PerlAPI::method_cache_gen("GenChild"); { no warnings; *GenParent::m = sub { 2 }; } $a == PerlAPI::method_cache_gen("GenChild") ? "same" : "moved"`, "moved"},
		{"GvCV of an inherited method, from the method cache too", `package GenKid; our @ISA = ("GenOther"); package main; PerlAPI::method_twice("GenKid", "m")`, "same sub"},
		{"@ISA changes", `my $a = PerlAPI::method_cache_gen("GenChild"); @GenChild::ISA = ("GenOther"); $a == PerlAPI::method_cache_gen("GenChild") ? "same" : "moved"`, "moved"},
	})
}

// croak(NULL) and croak_sv die with a reference as it is. The reference is
// not stringified on the way: its overloading is code of the program, and
// this one clears $@, which is the value being thrown.
func TestNativeThrowsReferenceUntouched(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	setup := `package Clears; our $calls = 0;
use overload '""' => sub { $calls++; eval { 1 }; "text" }, bool => sub { 1 }, fallback => 1;
package main; 1;`
	if r, err := p.Eval(context.Background(), setup); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"croak(NULL) with an object in $@", `my $e = bless {}, "Clears"; eval { PerlAPI::rethrow_errsv($e) }; (ref($@) || "not a reference") . ":" . $Clears::calls`, "Clears:0"},
		{"croak_sv with the object", `my $e = bless {}, "Clears"; eval { PerlAPI::throw($e) }; (ref($@) || "not a reference") . ":" . $Clears::calls`, "Clears:0"},
	})
}

// The errno native code leaves when an XSUB returns is the program's $!,
// in the interpreter's numbering; calls into the interpreter in between do
// not disturb it, a cleared errno clears $!, and an XSUB that does not
// touch errno leaves $! alone.
func TestNativeErrnoIsDollarBang(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `use Errno (); 1;`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"ENOENT", `$! = 0; PerlAPI::leave_errno("ENOENT"); $!{ENOENT} ? "ENOENT" : "other: " . ($! + 0)`, "ENOENT"},
		{"EINVAL", `$! = 0; PerlAPI::leave_errno("EINVAL"); $!{EINVAL} ? "EINVAL" : "other: " . ($! + 0)`, "EINVAL"},
		{"ERANGE", `$! = 0; PerlAPI::leave_errno("ERANGE"); $!{ERANGE} ? "ERANGE" : "other: " . ($! + 0)`, "ERANGE"},
		{"errno cleared", `$! = Errno::EPERM(); PerlAPI::leave_errno("zero"); $! + 0`, "0"},
		{"errno not touched", `$! = Errno::EPERM(); PerlAPI::leave_errno("nothing"); $!{EPERM} ? "EPERM" : "other: " . ($! + 0)`, "EPERM"},
	})
}

// A macro that stands in for a function of the perl API evaluates each
// argument once: a reference made in the argument of sv_rvweaken is one
// reference, the weak one, and the object it points at can go away.
func TestNativeMacrosEvaluateArgumentsOnce(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	setup := `package Gone; our $destroyed = 0; sub DESTROY { $destroyed++ } package main; 1;`
	if r, err := p.Eval(context.Background(), setup); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"sv_rvweaken(newRV_inc(sv)) and sv_isa(sv, name())", `my @keep; my $obj = bless {}, "Gone"; my @r = PerlAPI::weak_count(\@keep, $obj); undef $obj; join ",", @r, $Gone::destroyed, defined $keep[0] ? "still referenced" : "weak reference cleared"`, "1,1,1,weak reference cleared"},
	})
}

// SvPVbyte gives the byte form of a string: one held as UTF-8 is converted,
// and one with a character above 255 has none.
func TestNativeByteStrings(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"plain bytes", `PerlAPI::bytes_hex("a\xff")`, "61ff"},
		{"a string held as UTF-8", `my $s = "caf\xe9"; utf8::upgrade($s); PerlAPI::bytes_hex($s)`, "636166e9"},
		{"a wide character", `eval { PerlAPI::bytes_hex("\x{263a}") }; $@ =~ /^Wide character/ ? "refused" : "accepted: $@"`, "refused"},
		{"sv_uni_display and memCHRs", `PerlAPI::display("a\x{263a}\n") . "|" . PerlAPI::display("z")`, `a\x{263a}\n (vowel)|z (other)`},
		{"sv_setpv and sv_catpv", `join ",", PerlAPI::set_from_pointer(0)`, "text,kepttext,2"},
		{"sv_setpv and sv_catpv with a null pointer", `join ",", map { defined $_ ? $_ : "undef" } PerlAPI::set_from_pointer(1)`, "undef,kept,2"},
		{"my_setenv reaches the host's C environment", `PerlAPI::setenv_then_getenv("GOPERL_XS_ENV_PROBE", "set") . "," . (defined PerlAPI::setenv_then_getenv("GOPERL_XS_ENV_PROBE", undef) ? "still set" : "removed")`, "set,removed"},
	})
}

// When the interpreter dies inside SvIV or SvPV (an overloaded conversion
// that throws), the XSUB ends there, as it does for every other API call.
func TestNativeDirectAccessorUnwinds(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	setup := `package Thrower; use overload '0+' => sub { die "no number\n" }, '""' => sub { die "no string\n" };
package main; our $thrower = bless [], "Thrower"; 1;`
	if r, err := p.Eval(context.Background(), setup); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"SvIV", `eval { PerlAPI::read_then_mark($thrower, 0) }; "$@" . $PerlAPI::went_on`, "no number\n0"},
		{"SvPV", `eval { PerlAPI::read_then_mark($thrower, 1) }; "$@" . $PerlAPI::went_on`, "no string\n0"},
		{"a value that converts", `PerlAPI::read_then_mark(5, 0); $PerlAPI::went_on`, "1"},
	})
}

// Body slots and interpreter variables used as lvalues: any number of them
// in one call, a store that comes after other operations (the order C
// leaves open for an assignment), and raw bytes in the NV slot.
func TestNativeLvalueAccessors(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"SvIVX++ on many values", `my @a = (1 .. 500); PerlAPI::increment_all(\@a); "$a[0],$a[249],$a[499]"`, "2,251,501"},
		{"a store after later operations", `my $x = 1; PerlAPI::late_store($x) . ",$x"`, "8,8"},
		{"raw bytes in the NV slot", `my $x; PerlAPI::nv_slot_bits($x)`, "123456789abcdef"},
		{"an interpreter variable stored after later operations", `PerlAPI::late_variable_store()`, "1"},
		{"$@ localized from C", `$@ = "outer"; my $inner = PerlAPI::with_local_errsv(sub { die "inner\n" }); "$inner$@"`, "inner\nouter"},
		{"hv_ksplit", `my %h; PerlAPI::presize(\%h, 100) >= 99 ? "split" : "small"`, "split"},
		{"$_ swapped around a callback", `package Counted; our $live = 0; sub new { $live++; bless {} } sub DESTROY { $live-- }
package main; $_ = Counted->new; my $obj = $_; my $seen = PerlAPI::with_topic(sub { $_ .= "!" }, "ping");
my $same = ($_ == $obj) ? "same object" : "changed"; undef $obj; $_ = undef; "$seen,$same,$Counted::live"`, "ping!,same object,0"},
		{"$_ swapped inside a local", `my $out; { local $_ = "outer"; my $seen = PerlAPI::with_topic(sub { }, "inner"); $out = "$seen,$_" } $out`, "inner,outer"},
		{"PL_tainting is assignable", `my $on = PerlAPI::set_tainting(1); my $t = ${^TAINT}; my $off = PerlAPI::set_tainting(0); "$on,$t,$off,${^TAINT}"`, "1,1,0,0"},
	})
}

// caller_cx finds the calling frames and reports their whole type: a sub
// with its CV, and which kind of eval.
func TestNativeCallerCx(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	setup := `sub inner { join "|", map { PerlAPI::caller_frame($_) } 0 .. 3 }
sub outer { inner(1) } sub via_block { eval { inner() } } sub via_string { eval q{ inner() } } 1;`
	if r, err := p.Eval(context.Background(), setup); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"nested subs", `(split /\|/, outer())[0, 1]`, "sub main::outer with args"},
		{"the innermost frame", `(split /\|/, outer())[0]`, "sub main::inner with args"},
		{"an eval block", `(split /\|/, via_block())[1]`, "eval block"},
		{"an eval string", `(split /\|/, via_string())[1]`, "eval string"},
		{"a sub called with &", `sub amp { &inner } (split /\|/, amp(1))[0]`, "sub main::inner"},
	})
}

// PerlIO handles are the interpreter's: native reads, writes and seeks go
// through its layers and buffers, in sequence with the program's own I/O
// on the handle, whether the handle is a file or a scalar.
func TestNativePerlIOHandles(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "lines.txt")
	if err := os.WriteFile(file, []byte("first\nsecond\nthird line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var stdout, stderr bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, Stdout: &stdout, Stderr: &stderr})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { p.Close() })
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `our $file = '`+file+`'; our $out = '`+filepath.Join(dir, "out.txt")+`'; 1;`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"read after the program read a line", `open my $fh, "<", $file or die $!; my $l = <$fh>; $l . "+" . PerlAPI::pio_slurp($fh)`, "first\n+second\nthird line\n"},
		{"read from a scalar opened as a file", `open my $fh, "<", \"in memory text" or die $!; PerlAPI::pio_slurp($fh)`, "in memory text"},
		{"read through a layer", `open my $fh, "<:crlf", \"a\r\nb\r\n" or die $!; PerlAPI::pio_slurp($fh)`, "a\nb\n"},
		{"write in sequence with print", `my $buf = ""; open my $fh, ">", \$buf or die $!; print $fh "p1 "; my $n = PerlAPI::pio_emit($fh, "native"); print $fh "p2"; close $fh; "$n:$buf"`, "15:p1 native[42:fmt]!\np2"},
		{"write to a file", `open my $fh, ">", $out or die $!; PerlAPI::pio_emit($fh, "to disk"); close $fh; open $fh, "<", $out or die $!; local $/; scalar <$fh>`, "to disk[42:fmt]!\n"},
		{"seek and tell", `open my $fh, "<", $file or die $!; join ",", PerlAPI::pio_at($fh, 6, 0, 6), PerlAPI::pio_at($fh, 1, 1, 5), PerlAPI::pio_at($fh, -5, 2, 4), tell($fh)`, "6:second,13:third,19:line,23"},
		{"the host descriptor of a file", `open my $fh, "<", $file or die $!; PerlAPI::pio_host_size($fh) == -s $file ? "same file" : "other"`, "same file"},
		{"no host descriptor for a scalar", `open my $fh, "<", \"abc" or die $!; PerlAPI::pio_host_size($fh)`, "-1"},
		{"a stdio stream of the handle", `open my $fh, "<", $file or die $!; my $a = PerlAPI::pio_stdio_line($fh); my $b = <$fh>; my $c = PerlAPI::pio_stdio_line($fh); "$a$b$c"`, "first\nsecond\nthird line\n"},
		{"an exported stdio stream", `my $buf = ""; open my $fh, ">", \$buf or die $!; print $fh "a"; PerlAPI::pio_stdio_write($fh, "mid"); print $fh "z"; close $fh; $buf`, "a<mid>z"},
		{"the read buffer of a handle", `open my $fh, "<", $file or die $!; my $seen = PerlAPI::pio_buffer($fh); $seen . "|" . scalar(<$fh>)`, "24:fi|rst\n"},
		{"the standard handles", `PerlAPI::pio_std("x"); 1`, "1"},
	})
	if got, want := stdout.String(), "out:x\n"; got != want {
		t.Errorf("stdout: got %q, want %q", got, want)
	}
	if got, want := stderr.String(), "err:x\n"; got != want {
		t.Errorf("stderr: got %q, want %q", got, want)
	}
}
