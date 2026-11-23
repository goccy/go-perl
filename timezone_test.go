package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// An isolated program is in UTC until it names a time zone: the host's own
// zone is not its business. TZ from its environment counts like one it sets
// itself.
func TestSandboxLocalTime(t *testing.T) {
	program := perl.Program{Code: code(
		`use POSIX qw(strftime);`,
		`print strftime("%Y-%m-%d %H:%M:%S %z %Z", localtime(1721044800)), "\n";`,
		`$ENV{TZ} = "EST5EDT,M3.2.0,M11.1.0"; POSIX::tzset();`,
		`print strftime("%H:%M %z %Z", localtime(1721044800)), "\n";`,
	)}
	r := runProgram(t, perl.Config{}, program, "")
	r.want(t, 0, "2024-07-15 12:00:00 +0000 UTC\n08:00 -0400 EDT\n")

	r = runProgram(t, perl.Config{Env: []string{"TZ=JST-9"}}, program, "")
	r.want(t, 0, "2024-07-15 21:00:00 +0900 JST\n08:00 -0400 EDT\n")
}
