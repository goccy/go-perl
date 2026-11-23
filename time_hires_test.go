package perl_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// TestTimeHiResRealXS pins that Time::HiRes is the REAL statically linked
// core XS module (not an emulation): gettimeofday returns epoch time
// consistent with core time(), intervals measure real elapsed time, usleep
// actually sleeps, and the deliberately disabled portion (clock_nanosleep)
// fails as a clean feature probe.
func TestTimeHiResRealXS(t *testing.T) {
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	ctx := context.Background()

	r, err := p.Eval(ctx, `
		use Time::HiRes qw(gettimeofday tv_interval usleep sleep time);
		my ($s, $us) = gettimeofday();
		my $drift = abs($s - CORE::time());
		my $t0 = [gettimeofday];
		usleep(50_000); # 50ms
		my $elapsed = tv_interval($t0);
		my $hastime = time() > 0 ? 1 : 0;
		my $probe = eval { Time::HiRes::clock_nanosleep(0, 1); 1 } ? "clock-sleep" : "no-clock-sleep";
		join("|", $drift, ($us >= 0 && $us < 1_000_000 ? "us-ok" : "us-bad:$us"),
			sprintf("%.3f", $elapsed), $hastime, $probe,
			$INC{'Time/HiRes.pm'} ? "loaded" : "unloaded");
	`)
	if err != nil || r.Error != nil {
		t.Fatalf("Time::HiRes eval: err=%v error=%v", err, r.Error)
	}
	parts := strings.Split(resultStr(r), "|")
	if len(parts) != 6 {
		t.Fatalf("unexpected result shape: %q", resultStr(r))
	}
	if drift, _ := strconv.Atoi(parts[0]); drift > 1 {
		t.Errorf("gettimeofday seconds drift from CORE::time = %s (want <= 1)", parts[0])
	}
	if parts[1] != "us-ok" {
		t.Errorf("gettimeofday microseconds out of range: %s", parts[1])
	}
	if el, _ := strconv.ParseFloat(parts[2], 64); el < 0.045 || el > 5 {
		t.Errorf("usleep(50ms) measured %.3fs via tv_interval (want ~0.05)", el)
	}
	if parts[3] != "1" {
		t.Errorf("Time::HiRes::time() not positive")
	}
	if parts[4] != "no-clock-sleep" {
		t.Errorf("clock_nanosleep unexpectedly available: %s (must fail cleanly)", parts[4])
	}

	// The XS half really is linked: the bootstrap resolves and the module
	// reports a numeric version from the compiled unit.
	r, err = p.Eval(ctx, `Time::HiRes->VERSION ? "xs-booted" : "no-version"`)
	if err != nil || r.Error != nil {
		t.Fatalf("VERSION: err=%v error=%v", err, r.Error)
	}
	if resultStr(r) != "xs-booted" {
		t.Fatalf("VERSION probe = %q", resultStr(r))
	}
}

// Time::HiRes has clock_gettime and clock_getres: the real-time clock agrees
// with time, the monotonic clock goes forward by the time that passes, and a
// clock the system does not have is refused. A program that finds no
// clock_gettime falls back to worse time sources (AnyEvent's pure-Perl loop
// to POSIX::times, which a mocked clock does not move).
func TestTimeHiResClocks(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Time::HiRes qw(clock_gettime clock_getres usleep CLOCK_REALTIME CLOCK_MONOTONIC);`,
		`my $real = clock_gettime(CLOCK_REALTIME);`,
		`print "realtime: ", (abs($real - time) <= 2 ? "agrees with time" : "off: $real"), "\n";`,
		`my $m1 = clock_gettime(CLOCK_MONOTONIC); usleep(50_000); my $m2 = clock_gettime(CLOCK_MONOTONIC);`,
		`my $d = $m2 - $m1;`,
		`print "monotonic: ", ($m1 > 0 && $d >= 0.045 && $d < 5 ? "advances" : "m1=$m1 d=$d"), "\n";`,
		`my $res = clock_getres(CLOCK_MONOTONIC);`,
		`print "resolution: ", ($res > 0 && $res <= 1 ? "positive" : $res), "\n";`,
		`print "default clock: ", (abs(clock_gettime() - $real) <= 2 ? "realtime" : "other"), "\n";`,
		`print "unknown clock: ", clock_gettime(99), "\n";`,
	)}, "")
	r.want(t, 0, "realtime: agrees with time\nmonotonic: advances\nresolution: positive\ndefault clock: realtime\nunknown clock: -1\n")
}
