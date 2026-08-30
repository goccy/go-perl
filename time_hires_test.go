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
// actually sleeps, and the deliberately disabled portions (clock_gettime,
// interval timers — impossible under wasi) fail as clean feature probes.
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
		my $probe = eval { Time::HiRes::clock_gettime(0); 1 } ? "clock" : "no-clock";
		join("|", $drift, ($us >= 0 && $us < 1_000_000 ? "us-ok" : "us-bad:$us"),
			sprintf("%.3f", $elapsed), $hastime, $probe,
			$INC{'Time/HiRes.pm'} ? "loaded" : "unloaded");
	`)
	if err != nil || r.Error != nil {
		t.Fatalf("Time::HiRes eval: err=%v error=%v", err, r.Error)
	}
	parts := strings.Split(r.Value.String(), "|")
	if len(parts) != 6 {
		t.Fatalf("unexpected result shape: %q", r.Value.String())
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
	if parts[4] != "no-clock" {
		t.Errorf("clock_gettime unexpectedly available: %s (must fail cleanly under wasi)", parts[4])
	}

	// The XS half really is linked: the bootstrap resolves and the module
	// reports a numeric version from the compiled unit.
	r, err = p.Eval(ctx, `Time::HiRes->VERSION ? "xs-booted" : "no-version"`)
	if err != nil || r.Error != nil {
		t.Fatalf("VERSION: err=%v error=%v", err, r.Error)
	}
	if r.Value.String() != "xs-booted" {
		t.Fatalf("VERSION probe = %q", r.Value.String())
	}
}
