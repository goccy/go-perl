package perl_test

import (
	"strings"
	"sync"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal"
	"github.com/goccy/perlwasm2go/base"
)

// Sys::Syslog is a core module: it loads and has its constants.
func TestSysSyslogLoads(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Sys::Syslog qw(:standard :macros);`,
		`print "LOG_ERR=", LOG_ERR, " LOG_INFO=", LOG_INFO, " LOG_USER=", LOG_USER, " LOG_LOCAL7=", LOG_LOCAL7, " LOG_PID=", LOG_PID, "\n";`,
		`print "log socket: ", Sys::Syslog::_PATH_LOG(), "\n";`,
	)}, "")
	r.want(t, 0, "LOG_ERR=3 LOG_INFO=6 LOG_USER=8 LOG_LOCAL7=184 LOG_PID=1\nlog socket: /dev/log\n")
}

// Without Config.Syslog the program has no system logger. Sys::Syslog finds
// its native mechanism unavailable and goes on to its other ones; in the
// sandbox none of them can connect, and the program is told so. Nothing is
// reported as logged.
func TestSysSyslogHasNoNativeMechanismByDefault(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Sys::Syslog qw(:standard :macros setlogsock);`,
		`my $logged = eval { openlog("go-perl-test", "ndelay,pid", LOG_USER); syslog(LOG_INFO, "lost?") };`,
		`print defined $logged ? "reported as logged: $logged\n" : "openlog failed\n";`,
		`print $@ =~ /no connection to syslog available/ ? "no mechanism connected\n" : "error: $@";`,
		`print $@ =~ /- native: no system logger is available to this program/ ? "native was tried and is unavailable\n" : "error: $@";`,
		`print $@ =~ m{- getservbyname failed for syslog/tcp.*- /dev/log is not a socket}s ? "other mechanisms were tried\n" : "error: $@";`,
		// asked for by name, it is just as unavailable
		`setlogsock("native");`,
		`$logged = eval { openlog("go-perl-test", "ndelay", LOG_USER); syslog(LOG_INFO, "lost?") };`,
		`print defined $logged ? "reported as logged: $logged\n" : "native alone: openlog failed\n";`,
	)}, "")
	r.want(t, 0, "openlog failed\n"+
		"no mechanism connected\n"+
		"native was tried and is unavailable\n"+
		"other mechanisms were tried\n"+
		"native alone: openlog failed\n")
}

// recordedLog is a system logger that keeps what it is given.
type recordedLog struct {
	mu       sync.Mutex
	opened   []string
	messages []base.SyslogMessage
	closed   int
}

func (l *recordedLog) Open(ident string, options, facility int32) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.opened = append(l.opened, ident)
	return nil
}

func (l *recordedLog) Log(msg base.SyslogMessage) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, msg)
	return nil
}

func (l *recordedLog) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed++
	return nil
}

// With Config.Syslog the native mechanism delivers to the host's logger:
// each message arrives with its level, facility, ident and text as the
// program gave them.
func TestSysSyslogDeliversToTheHostLogger(t *testing.T) {
	log := &recordedLog{}
	hostLogger := internal.NewSyslogger
	internal.NewSyslogger = func() base.Syslogger { return log }
	defer func() { internal.NewSyslogger = hostLogger }()

	r := runProgram(t, perl.Config{Syslog: true}, perl.Program{Code: code(
		`use Sys::Syslog qw(:standard :macros);`,
		`openlog("go-perl-test", "ndelay,pid", LOG_DAEMON);`,
		`print syslog(LOG_INFO, "started %s with %d workers", "pool", 4) ? "logged\n" : "not logged\n";`,
		`print syslog(LOG_ERR | LOG_LOCAL3, "disk 93% full") ? "logged\n" : "not logged\n";`,
		`setlogmask(LOG_UPTO(LOG_WARNING));`,
		`syslog(LOG_DEBUG, "masked out");`,
		`syslog("warning", "by name");`,
		`closelog();`,
		`print "done\n";`,
	)}, "")
	r.want(t, 0, "logged\nlogged\ndone\n")

	const pid, ndelay = 0x01, 0x08
	want := []base.SyslogMessage{
		{Severity: 6, Facility: 3, Options: pid | ndelay, Ident: "go-perl-test", Message: "started pool with 4 workers"},
		{Severity: 3, Facility: 19, Options: pid | ndelay, Ident: "go-perl-test", Message: "disk 93% full"},
		{Severity: 4, Facility: 3, Options: pid | ndelay, Ident: "go-perl-test", Message: "by name"},
	}
	if len(log.opened) == 0 || log.opened[0] != "go-perl-test" || log.closed == 0 {
		t.Errorf("opened=%q closed=%d, want the ident opened and the log closed", log.opened, log.closed)
	}
	if len(log.messages) != len(want) {
		t.Fatalf("messages = %+v, want %d of them", log.messages, len(want))
	}
	for i, w := range want {
		got := log.messages[i]
		// Sys::Syslog ends each message with a newline unless told not to.
		got.Message = strings.TrimSuffix(got.Message, "\n")
		if got != w {
			t.Errorf("message %d = %+v, want %+v", i, got, w)
		}
	}
}
