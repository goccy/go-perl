//go:build darwin || linux

package gperl_test

import (
	"os"
	"testing"
	"time"
)

// Local time follows the TZ variable as the program has it at that moment:
// POSIX zone strings, and the names of the host's zone database. Without TZ
// the program is where the host is.
func TestLocalTimeFollowsTZ(t *testing.T) {
	const script = `
use POSIX qw(strftime mktime tzset tzname);
my ($winter, $summer) = (1705320000, 1721044800); # 2024-01-15 and 2024-07-15, 12:00:00 UTC
sub stamp { my ($t) = @_; strftime("%Y-%m-%d %H:%M:%S %z %Z", localtime($t)) . " dst=" . (localtime($t))[8] }
for my $tz (@ARGV) {
    if ($tz eq 'unset') { delete $ENV{TZ} } else { $ENV{TZ} = $tz }
    tzset();
    report("$tz: ", stamp($winter), " | ", stamp($summer));
    report("$tz: tzname=", join("/", tzname()), " mktime inverts localtime=",
           (mktime(localtime($winter)) == $winter && mktime(localtime($summer)) == $summer ? "yes" : "no"),
           " scalar localtime=", scalar(localtime($summer)));
}
$ENV{TZ} = "JST-9";
report("gmtime is not local time: ", scalar(gmtime($summer)));
`
	// The host's own zone is whatever the test runs in; TZ is taken out of
	// the environment so that the program starts without one.
	if tz, set := os.LookupEnv("TZ"); set {
		os.Unsetenv("TZ")
		defer os.Setenv("TZ", tz)
	}
	zones := []string{"UTC", "JST-9", "EST5EDT,M3.2.0,M11.1.0", "<+0530>-5:30", "No/Such_Zone", "unset"}
	want := []string{
		"UTC: 2024-01-15 12:00:00 +0000 UTC dst=0 | 2024-07-15 12:00:00 +0000 UTC dst=0",
		"UTC: tzname=UTC/UTC mktime inverts localtime=yes scalar localtime=Mon Jul 15 12:00:00 2024",
		"JST-9: 2024-01-15 21:00:00 +0900 JST dst=0 | 2024-07-15 21:00:00 +0900 JST dst=0",
		"JST-9: tzname=JST/JST mktime inverts localtime=yes scalar localtime=Mon Jul 15 21:00:00 2024",
		"EST5EDT,M3.2.0,M11.1.0: 2024-01-15 07:00:00 -0500 EST dst=0 | 2024-07-15 08:00:00 -0400 EDT dst=1",
		"EST5EDT,M3.2.0,M11.1.0: tzname=EST/EDT mktime inverts localtime=yes scalar localtime=Mon Jul 15 08:00:00 2024",
		"<+0530>-5:30: 2024-01-15 17:30:00 +0530 +0530 dst=0 | 2024-07-15 17:30:00 +0530 +0530 dst=0",
		"<+0530>-5:30: tzname=+0530/+0530 mktime inverts localtime=yes scalar localtime=Mon Jul 15 17:30:00 2024",
		"No/Such_Zone: 2024-01-15 12:00:00 +0000 UTC dst=0 | 2024-07-15 12:00:00 +0000 UTC dst=0",
		"No/Such_Zone: tzname=UTC/UTC mktime inverts localtime=yes scalar localtime=Mon Jul 15 12:00:00 2024",
	}
	// Without TZ: the host's zone, as Go sees it.
	winter, summer := time.Unix(1705320000, 0), time.Unix(1721044800, 0)
	stamp := func(at time.Time) string {
		dst := "0"
		if at.IsDST() {
			dst = "1"
		}
		return at.Format("2006-01-02 15:04:05 -0700 MST") + " dst=" + dst
	}
	want = append(want, "unset: "+stamp(winter)+" | "+stamp(summer))
	if _, err := time.LoadLocation("Europe/Berlin"); err == nil {
		zones = append(zones, "Europe/Berlin", ":Asia/Tokyo")
		want = append(want,
			"Europe/Berlin: 2024-01-15 13:00:00 +0100 CET dst=0 | 2024-07-15 14:00:00 +0200 CEST dst=1",
			"Europe/Berlin: tzname=CET/CEST mktime inverts localtime=yes scalar localtime=Mon Jul 15 14:00:00 2024",
			":Asia/Tokyo: 2024-01-15 21:00:00 +0900 JST dst=0 | 2024-07-15 21:00:00 +0900 JST dst=0",
			":Asia/Tokyo: tzname=JST/JST mktime inverts localtime=yes scalar localtime=Mon Jul 15 21:00:00 2024",
		)
	}
	got := runForkScriptArgs(t, script, zones...)
	// The second line of the host's own zone depends on that zone's names;
	// the first one is compared.
	var kept []string
	for _, line := range got {
		if len(line) >= len("unset: tzname=") && line[:len("unset: tzname=")] == "unset: tzname=" {
			continue
		}
		kept = append(kept, line)
	}
	want = append(want, "gmtime is not local time: Mon Jul 15 12:00:00 2024")
	wantLines(t, kept, want...)
}
