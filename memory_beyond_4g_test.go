package perl_test

import (
	"runtime"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal"
)

// The interpreter's memory goes beyond 4 GiB: a string longer than 4 GiB
// is made and used, and large values made after it, which do not fit the
// memory below it, live at addresses above 4 GiB (pack "p" gives a
// string's address), where strings, arrays and hashes work as anywhere
// else. The test writes about 4.7 GiB, so it runs where the host has room
// for that and the interpreter runs at its usual speed.
func TestMemoryBeyondFourGigabytes(t *testing.T) {
	if testing.Short() {
		t.Skip("writes 4.5 GiB")
	}
	if raceEnabled {
		t.Skip("writes 4.5 GiB through instrumented code")
	}
	if runtime.GOARCH == "386" || runtime.GOARCH == "arm" {
		t.Skip("a 32-bit host has no room for the memory")
	}
	if mem, ok := internal.HostMemory(); !ok || mem < 16<<30 {
		t.Skipf("the host has %d bytes of memory", mem)
	}
	r := runProgram(t, perl.Config{MaxMemoryBytes: 6 << 30}, perl.Program{Code: code(
		`my $size = 4.5 * 2**30;`,
		`my $big = "\0" x $size;`,
		`substr($big, -5) = "tail!";`,
		`print "length=", length($big), " tail=", substr($big, -5), "\n";`,
		`my $high = "abc" x (32 * 2**20);`,
		`my $at = unpack("Q", pack("p", $high));`,
		`print "string above 4 GiB=", ($at >= 2**32 ? 1 : 0), " index=", index($high, "cab", 2**25), " match=", ($high =~ /(abc){3}\z/ ? 1 : 0), "\n";`,
		`my @values = map { "value $_" } 1 .. 2_000_000;`,
		`my %hash = map { ("key $_" => $values[$_ - 1]) } 1 .. 200_000;`,
		`print "element=$values[1_999_999] hash=$hash{'key 200000'} count=", scalar(keys %hash), "\n";`,
		`print "pack p=", unpack("p", pack("p", $values[999])), "\n";`,
		`undef $big;`,
		`print "freed\n";`,
	)}, "")
	r.want(t, 0, "length=4831838208 tail=tail!\n"+
		"string above 4 GiB=1 index=33554432 match=1\n"+
		"element=value 2000000 hash=value 200000 count=200000\n"+
		"pack p=value 1000\n"+
		"freed\n")
}

// An instance limited to the host's memory runs.
func TestHostMemoryLimitedInstance(t *testing.T) {
	r := runProgram(t, perl.Config{MaxMemoryBytes: perl.HostMemoryLimit}, perl.Program{Code: code(
		`my @a = map { $_ * 2 } 1 .. 100_000; print scalar(@a), " $a[-1]\n";`,
	)}, "")
	r.want(t, 0, "100000 200000\n")
}
