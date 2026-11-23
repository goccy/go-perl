//go:build darwin || linux

package xs_test

import (
	"context"
	"strconv"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
	"github.com/goccy/go-perl/internal"
)

// Native code is handed values that live above 4 GiB of the interpreter's
// memory, and reads and writes those strings, arrays and hashes, calls back
// into Perl with them, keeps pointers of its own in them, makes objects and
// throws values, as anywhere else in the memory. A 4.5 GiB string takes
// the memory below 4 GiB, and each value is made until it is above (pack
// "p" and the number of a reference say where): what is free below is
// handed out first, and the values made there are kept, so the next one is
// made elsewhere. The test writes about 4.6 GiB, so it runs where the host
// has room for that and the interpreter runs at its usual speed.
func TestNativeValuesBeyondFourGigabytes(t *testing.T) {
	if testing.Short() {
		t.Skip("writes 4.5 GiB")
	}
	if raceEnabled {
		t.Skip("writes 4.5 GiB through instrumented code")
	}
	if mem, ok := internal.HostMemory(); !ok || mem < 16<<30 {
		t.Skipf("the host has %d bytes of memory", mem)
	}
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, MaxMemoryBytes: 6 << 30})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { p.Close() })
	perlAPIInstance(t, p)

	// fresh makes a value with make until every address where answers for
	// it is above 4 GiB; the values made below are kept, so their memory
	// is not handed out again. str answers where a reference to a string
	// and its buffer are. The SV heads free below 4 GiB are taken first,
	// a thousand at a time, so the values of the cases are seldom made
	// more than once.
	r, err := p.Eval(context.Background(), `
		our @keep;
		sub fresh {
			my ($make, $where) = @_;
			for (1 .. 1_000_000) {
				my $v = $make->();
				return $v unless grep { $_ < 2**32 } $where->($v);
				push @keep, $v;
			}
			die "nothing made above 4 GiB\n";
		}
		sub str { my $s = shift; (0 + $s, unpack("Q", pack("p", $$s))) }
		our $big = "\0" x (4.5 * 2**30);
		fresh(sub { [map { [$_] } 1 .. 1000] }, sub { map { 0 + $_, 0 + \$_->[0] } @{$_[0]} });
		length($big)`)
	if err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	if got, want := resultStr(r), strconv.Itoa(9<<29); got != want {
		t.Fatalf("setup: %s bytes, want %s", got, want)
	}

	runEvalCases(t, p, []evalCase{
		{"a string read by native code", `my $s = fresh(sub { (my $s = "abc") x= 100_000; \$s }, \&str); PerlAPI::string_nomg($$s) eq $$s ? "same" : "different"`, "same"},
		{"a wide string", `my $s = fresh(sub { (my $s = "\x{263a}") x= 1000; \$s }, \&str); join ",", PerlAPI::flags($$s), length(PerlAPI::string_nomg($$s))`, "POK,UTF8,,3000"},
		{"the values of an array changed in place", `my $a = fresh(sub { [1 .. 500] }, sub { 0 + $_[0], map { 0 + \$_ } @{$_[0]} }); join ",", PerlAPI::increment_all($a), "$a->[0] $a->[249] $a->[499]"`, "500,2 251 501"},
		{"an array and a hash", `my ($a, $h) = @{ fresh(sub { [[10, 20, 30], {k => 1, j => 2}] }, sub { map { 0 + $_, map { 0 + \$_ } ref($_) eq "HASH" ? values %$_ : @$_ } @{$_[0]} }) }; join ",", PerlAPI::container_ops($a, $h), scalar(@$a), defined $a->[1] ? "defined" : "undef"`, "20,2,1,3,undef"},
		{"an array stored into by native code", `my ($a, $x) = @{ fresh(sub { [[], \(my $x = "value")] }, sub { 0 + $_[0][0], 0 + $_[0][1] }) }; my $ok = PerlAPI::store_first($a, $$x) ? 1 : 0; join ",", $ok, Internals::SvREFCNT($$x), $a->[0], \$a->[0] == $x ? "same" : "copy"`, "1,2,value,same"},
		{"pointers of native code's own in an array", `my $k = fresh(sub { PerlAPI::kept_array(100) }, sub { 0 + $_[0] }); push our @kept, $k; join ",", PerlAPI::kept_array_sums($k)`, "4950,4950,4950"},
		{"pointers of native code's own in a hash", `my $k = fresh(sub { PerlAPI::kept_hash(100) }, sub { 0 + $_[0] }); push our @kept_h, $k; join ",", PerlAPI::kept_hash_sums($k, 100)`, "4950,4950,4950"},
		{"a callback with arguments and results", `my $code = fresh(sub { my $f = 2; sub { $_[0] * $f } }, sub { 0 + $_[0] }); join ",", PerlAPI::call_on_new_stack($code, 21)`, "42"},
		{"an object made by native code", `my $o = fresh(sub { PerlAPI::new_object("My::Class", "payload" x 1000) }, \&str); join ",", ref($o), length($$o)`, "My::Class,7000"},
		{"a value thrown by native code", `my $e = fresh(sub { bless { code => 7 }, "My::Err" }, sub { 0 + $_[0] }); eval { PerlAPI::throw($e) }; join ",", ref($@), $@->{code}, $@ == $e ? "same" : "copy"`, "My::Err,7,same"},
	})
}
