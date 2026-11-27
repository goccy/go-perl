package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// The strings perl reads as infinity and not-a-number give those values,
// with either sign and in any case.
func TestInfinityAndNaNFromStrings(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`print join(" ", "inf" + 0, "-Inf" + 0, "Infinity" + 0, "nan" + 0, "NaN" + 0), "\n";`,
		`print "inf is the overflow value: ", ("inf" + 0 == 9**9**9 ? "yes" : "no"), "\n";`,
		`my $nan = "nan" + 0;`,
		`print "nan differs from itself: ", ($nan != $nan ? "yes" : "no"), "\n";`,
		`print "inf - inf: ", ("inf" - "inf"), "\n";`,
		`use POSIX ();`,
		`print "strtod: ", join(" ", (POSIX::strtod("inf"))[0], (POSIX::strtod("-inf"))[0]), "\n";`,
	)}, "")
	r.want(t, 0, "Inf -Inf Inf NaN NaN\n"+
		"inf is the overflow value: yes\n"+
		"nan differs from itself: yes\n"+
		"inf - inf: NaN\n"+
		"strtod: Inf -Inf\n")
}

// %Config describes the long double and the 64-bit formats of the machine
// the interpreter runs on, and a format with the long double modifier
// prints.
func TestConfigDescribesLongDouble(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Config;`,
		`print join(" ", map { "$_=$Config{$_}" } qw(longdblsize longdblkind longdblmantbits doublesize nvsize)), "\n";`,
		`print join(" ", map { "$_=$Config{$_}" } qw(sPRIfldbl sPRId64 sPRIu64 st_ino_size st_dev_size)), "\n";`,
		`print sprintf("%Lf %.3Lg %Le", 1.5, 2.25, 1000), "\n";`,
		`print "gnulibc=", (defined $Config{d_gnulibc} ? $Config{d_gnulibc} : "undef"), "\n";`,
	)}, "")
	r.want(t, 0, "longdblsize=16 longdblkind=1 longdblmantbits=112 doublesize=8 nvsize=8\n"+
		`sPRIfldbl="Lf" sPRId64="ld" sPRIu64="lu" st_ino_size=8 st_dev_size=8`+"\n"+
		"1.500000 2.25 1.000000e+03\n"+
		"gnulibc=undef\n")
}

// The C99 math functions POSIX offers are the C library's.
func TestPOSIXMathFunctions(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use POSIX ();`,
		`printf "%.12g %.12g %.12g\n", POSIX::erf(0.5), POSIX::tgamma(4.5), POSIX::cbrt(27);`,
		`printf "%.12g %.12g %.12g\n", POSIX::expm1(1e-10), POSIX::log1p(1e-10), POSIX::hypot(3, 4);`,
		`printf "%s %s %s %s\n", POSIX::lround(2.5), POSIX::trunc(-2.7), POSIX::round(-2.5), POSIX::fma(2, 3, 4);`,
		`printf "%d %d %d\n", (POSIX::isfinite(1) ? 1 : 0), (POSIX::isinf(9**9**9) ? 1 : 0), (POSIX::isnan("nan" + 0) ? 1 : 0);`,
		`my @rq = POSIX::remquo(10, 3); print "@rq\n";`,
	)}, "")
	r.want(t, 0, "0.520499877813 11.6317283966 3\n"+
		"1.00000000005e-10 9.9999999995e-11 5\n"+
		"3 -2 -3 10\n"+
		"1 1 1\n"+
		"1 3\n")
}

// A program started with locale categories that disagree runs: the C
// library names such a locale by position ("C.UTF-8;C;C;C;C;C"), and perl
// reads the name back.
func TestLocaleCategoriesMayDisagree(t *testing.T) {
	cfg := perl.Config{Env: []string{"LC_ALL=", "LANG=C", "LC_CTYPE=C.UTF-8"}}
	r := runProgram(t, cfg, perl.Program{Code: code(
		`use POSIX qw(setlocale LC_ALL LC_CTYPE LC_NUMERIC);`,
		`print "running\n";`,
		`print "ctype=", setlocale(LC_CTYPE), " numeric=", setlocale(LC_NUMERIC), "\n";`,
		`print "all=", setlocale(LC_ALL), "\n";`,
		`print "set all: ", (defined setlocale(LC_ALL, "C") ? "ok" : "failed"), " now ", setlocale(LC_ALL), "\n";`,
		`print "set ctype: ", (defined setlocale(LC_CTYPE, "C.UTF-8") ? "ok" : "failed"), " now ", setlocale(LC_CTYPE), "\n";`,
	)}, "")
	if r.status != 0 || r.stderr != "" {
		t.Fatalf("status=%d stdout=%q stderr=%q, want a clean run", r.status, r.stdout, r.stderr)
	}
	r.want(t, 0, "running\n"+
		"ctype=C.UTF-8 numeric=C\n"+
		"all=C.UTF-8;C;C;C;C;C\n"+
		"set all: ok now C\n"+
		"set ctype: ok now C.UTF-8\n")
}
