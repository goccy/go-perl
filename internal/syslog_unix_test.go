//go:build !windows && !plan9

package internal

import (
	"errors"
	"log/syslog"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/perlwasm2go/base"
)

// noLogDaemon makes the host logger find no log daemon to connect to, and
// returns the file that stands for the console.
func noLogDaemon(t *testing.T) (console string) {
	t.Helper()
	dial, cons := DialSyslog, SyslogConsole
	t.Cleanup(func() { DialSyslog, SyslogConsole = dial, cons })
	DialSyslog = func(syslog.Priority, string) (*syslog.Writer, error) {
		return nil, errors.New("no log daemon")
	}
	SyslogConsole = filepath.Join(t.TempDir(), "console")
	if err := os.WriteFile(SyslogConsole, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return SyslogConsole
}

// A host without a log daemon loses what a program logs, as syslog(3) does:
// neither openlog nor syslog fails. With LOG_CONS the message goes to the
// console instead.
func TestHostSyslogWithoutALogDaemon(t *testing.T) {
	console := noLogDaemon(t)
	l := newHostSyslogger()
	if err := l.Open("prog", base.SyslogOptNDelay|base.SyslogOptPID, 3); err != nil {
		t.Errorf("openlog with ndelay: %v", err)
	}
	if err := l.Log(base.SyslogMessage{Severity: syslogInfo, Facility: 3, Options: base.SyslogOptNDelay, Ident: "prog", Message: "lost"}); err != nil {
		t.Errorf("syslog: %v", err)
	}
	if err := l.Log(base.SyslogMessage{Severity: syslogErr, Facility: 3, Options: base.SyslogOptCons, Ident: "prog", Message: "to the console"}); err != nil {
		t.Errorf("syslog with cons: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Errorf("closelog: %v", err)
	}
	got, err := os.ReadFile(console)
	if err != nil {
		t.Fatal(err)
	}
	if want := "prog: to the console\r\n"; string(got) != want {
		t.Errorf("console has %q, want %q", got, want)
	}
}
