//go:build darwin || linux

package gperl_test

import "testing"

// Timers finer than a second: Time::HiRes's alarm and ualarm, and the
// interval timer behind them, which can repeat and reports what is left of
// it.
func TestSubSecondAndRepeatingTimers(t *testing.T) {
	got := runForkScript(t, `
use Time::HiRes qw(ualarm setitimer getitimer ITIMER_REAL);
my $fired = 0;
$SIG{ALRM} = sub { $fired++ };
sub elapsed { my ($start) = @_; my $took = Time::HiRes::time() - $start; $took >= 0.05 && $took < 3 ? "in time" : "after ${took}s" }

my $start = Time::HiRes::time();
ualarm(100_000);
Time::HiRes::sleep(10);
report("ualarm of 0.1s: fired=$fired ", elapsed($start));

($fired, $start) = (0, Time::HiRes::time());
Time::HiRes::alarm(0.15);
Time::HiRes::sleep(10);
report("alarm of 0.15s: fired=$fired ", elapsed($start));

($fired, $start) = (0, Time::HiRes::time());
setitimer(ITIMER_REAL, 0.1, 0.05);
Time::HiRes::sleep(10) while $fired < 4 && Time::HiRes::time() - $start < 20;
my ($left, $interval) = getitimer(ITIMER_REAL);
report("repeating timer: fired=$fired ", elapsed($start), " interval=", sprintf("%.2f", $interval),
       " time left=", ($left > 0 && $left <= 0.05 ? "within the interval" : $left));
my ($was_left, $was_interval) = setitimer(ITIMER_REAL, 0);
report("disarming reports the interval: ", sprintf("%.2f", $was_interval));
$fired = 0;
Time::HiRes::sleep(0.3);
report("after disarming: fired=$fired left=", scalar(getitimer(ITIMER_REAL)));

# the core alarm is the same timer
setitimer(ITIMER_REAL, 2.5);
report("alarm reports what is left of it in whole seconds: ", alarm(0));
alarm(7);
report("getitimer sees an alarm: ", sprintf("%.0f", scalar(getitimer(ITIMER_REAL))));
alarm(0);

# the only timer is the real-time one
report("ITIMER_VIRTUAL: ", (eval { Time::HiRes::ITIMER_VIRTUAL(); 1 } ? "offered" : "not offered"));
`)
	wantLines(t, got,
		"ualarm of 0.1s: fired=1 in time",
		"alarm of 0.15s: fired=1 in time",
		"repeating timer: fired=4 in time interval=0.05 time left=within the interval",
		"disarming reports the interval: 0.05",
		"after disarming: fired=0 left=0",
		"alarm reports what is left of it in whole seconds: 3",
		"getitimer sees an alarm: 7",
		"ITIMER_VIRTUAL: not offered",
	)
}

// A timer and a sleep of the same length, the timer set first: the timer's
// moment comes first, so its handler has run when the statement after the
// sleep runs, as it has where a kernel keeps both. (Lock::File restores an
// outer alarm around its own and tests that it fired by the end of a sleep
// that ends with it.)
func TestTimerSetBeforeASleepOfTheSameLengthFiresFirst(t *testing.T) {
	got := runForkScript(t, `
use Time::HiRes qw(ualarm usleep);
my ($before, $after) = (0, 0);
for (1 .. 40) {
    my $fired = 0;
    local $SIG{ALRM} = sub { $fired++ };
    ualarm(30_000);
    usleep(30_000);
    if ($fired) { $before++ } else { $after++ }
    usleep(20_000);
}
report("the handler ran before the statement after the sleep $before times, after it $after times");
$SIG{ALRM} = sub { $before = "fired" };
alarm(1);
sleep(1);
report("alarm(1); sleep(1): $before");
`)
	wantLines(t, got,
		"the handler ran before the statement after the sleep 40 times, after it 0 times",
		"alarm(1); sleep(1): fired",
	)
}
