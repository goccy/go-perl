//go:build !windows && !plan9

package perl_test

import (
	"errors"
	"log/syslog"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal"
)

// With Config.Syslog on a host that has no log daemon, Sys::Syslog's native
// mechanism behaves as on a perl whose C library has syslog(3): openlog and
// syslog succeed and the message is lost. Log::Report's syslog dispatcher,
// for one, opens the log that way and must not die of it.
func TestSysSyslogWithoutALogDaemon(t *testing.T) {
	dial := internal.DialSyslog
	t.Cleanup(func() { internal.DialSyslog = dial })
	internal.DialSyslog = func(syslog.Priority, string) (*syslog.Writer, error) {
		return nil, errors.New("no log daemon")
	}
	r := runProgram(t, perl.Config{Syslog: true}, perl.Program{Code: code(
		`use Sys::Syslog qw(:standard :macros);`,
		`my $ok = eval { openlog("go-perl-test", "ndelay,pid", LOG_USER); 1 };`,
		`print $ok ? "opened\n" : "openlog failed: $@";`,
		`print eval { syslog(LOG_NOTICE, "this is a test") } ? "logged\n" : "syslog failed: $@";`,
		`closelog();`,
	)}, "")
	r.want(t, 0, "opened\nlogged\n")
}
