package perl

import (
	"bytes"
	"context"
	"testing"

	goperlfs "github.com/goccy/go-perl/fs"
)

// Workloads that show what the interpreter's configuration costs: each one
// runs as a whole program on a fresh instance (as under the perl command),
// and reports, besides the time, the size its linear memory has grown to —
// the memory never shrinks, so that is its peak.
var workloads = []struct {
	name string
	code string
}{
	{"IntegerLoop", `
my ($sum, $x) = (0, 12345);
for my $i (1 .. 3_000_000) {
    $x = ($x * 1103 + 12345) % 65521;
    $sum += $x ^ ($i & 255);
    $sum -= $i >> 3;
}
print "$sum\n";
`},
	{"Strings", `
my $text = join " ", map { "word$_" } 1 .. 20_000;
my $total = 0;
for my $round (1 .. 100) {
    my $copy = uc $text;
    $copy =~ tr/A-M/n-z/;
    my @parts = split / /, $copy;
    $total += length(join "-", reverse @parts);
    $total += index($copy, "WORD19999");
    $total += length sprintf("%s:%05d:%-10s|", $_, $round, "x") for @parts[0 .. 999];
    substr($copy, 0, 100) = "";
    $total += length $copy . ("ab" x 1000);
}
print "$total\n";
`},
	{"HashesAndArrays", `
my (%count, @list);
for my $i (1 .. 300_000) {
    my $key = "k" . ($i % 50_000);
    $count{$key} += $i;
    push @list, $i * 2;
}
my $total = 0;
$total += $count{$_} for sort keys %count;
my @sorted = sort { $b <=> $a } @list;
$total += $sorted[$_] for 0 .. 999;
my %inverse = map { $list[$_] => $_ } 0 .. 99_999;
$total += scalar keys %inverse;
my @nested = map { [ $_, { id => $_, name => "n$_" } ] } 1 .. 50_000;
$total += $_->[1]{id} for @nested;
print "$total\n";
`},
	{"SubCalls", `
sub fib { my $n = shift; $n < 2 ? $n : fib($n - 1) + fib($n - 2) }
sub add { $_[0] + $_[1] }
package Counter { sub new { bless { n => 0 }, shift } sub step { $_[0]{n} += $_[1]; $_[0] } sub n { $_[0]{n} } }
my $total = fib(27);
$total = add($total, $_) for 1 .. 500_000;
my $counter = Counter->new;
$counter->step($_) for 1 .. 500_000;
print $total + $counter->n, "\n";
`},
	{"Regex", `
my @lines = map { "2026-11-" . sprintf("%02d", $_ % 28 + 1) . " host$_ GET /path/$_/item?id=$_&x=abc HTTP/1.1 " . (200 + $_ % 5) . " " . ($_ * 37 % 5000) } 1 .. 60_000;
my ($hits, $bytes, %by_status) = (0, 0);
for my $line (@lines) {
    if ($line =~ m{^(\d{4})-(\d\d)-(\d\d) (\S+) (GET|POST) (/\S*?)\?id=(\d+)\S* HTTP/1\.\d (\d{3}) (\d+)$}) {
        $hits++;
        $bytes += $9;
        $by_status{$8}++;
    }
    (my $copy = $line) =~ s/(\d+)/<$1>/g;
    $bytes += () = $copy =~ /<\d+>/g;
}
print "$hits $bytes ", join(",", map { "$_=$by_status{$_}" } sort keys %by_status), "\n";
`},
	{"ModuleLoading", `
use strict; use warnings;
use POSIX (); use File::Spec; use File::Temp (); use File::Basename; use File::Path ();
use Data::Dumper; use Getopt::Long; use Text::Wrap; use Text::ParseWords; use Time::Local;
use Scalar::Util (); use List::Util (); use Storable (); use Encode (); use JSON::PP ();
use HTTP::Tiny (); use Pod::Usage (); use Test::More (); use Math::BigInt (); use IO::Socket::INET ();
print scalar(keys %INC), "\n";
`},
}

func BenchmarkWorkload(b *testing.B) {
	stdlib, err := ExtractStdlib()
	if err != nil {
		b.Fatal(err)
	}
	for _, w := range workloads {
		b.Run(w.name, func(b *testing.B) {
			var memory uint64
			for n := 0; n < b.N; n++ {
				var out, errOut bytes.Buffer
				p, err := New(Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, Stdout: &out, Stderr: &errOut})
				if err != nil {
					b.Fatal(err)
				}
				status, err := p.RunProgram(context.Background(), Program{Code: []string{w.code}})
				if err != nil || status != 0 || out.Len() == 0 {
					b.Fatalf("status=%d err=%v stdout=%q stderr=%q", status, err, out.String(), errOut.String())
				}
				memory = p.raw.MemorySize()
				p.Close()
			}
			b.ReportMetric(float64(memory), "linear-memory-bytes")
		})
	}
}
