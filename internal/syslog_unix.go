//go:build !windows && !plan9

package internal

import (
	"fmt"
	"log/syslog"
	"sync"

	"github.com/goccy/perlwasm2go/base"
)

// hostSyslogger delivers to the system logger of the host through
// log/syslog, which speaks to the local log daemon. A connection to the
// daemon carries one facility and one tag, so the logger keeps the
// connection for the facility and ident it last logged with and makes a new
// one when a message names others. It is shared by an instance and its fork
// children.
type hostSyslogger struct {
	mu       sync.Mutex
	writer   *syslog.Writer
	ident    string
	facility int32
}

func newHostSyslogger() base.Syslogger { return &hostSyslogger{} }

// connectLocked returns the connection for ident and facility, making it
// when the one at hand is for something else. An empty ident is the
// program's name, which is what log/syslog takes an empty tag for. Caller
// holds l.mu.
func (l *hostSyslogger) connectLocked(ident string, facility int32) (*syslog.Writer, error) {
	if l.writer != nil && l.ident == ident && l.facility == facility {
		return l.writer, nil
	}
	if facility < 0 || facility > 23 {
		return nil, fmt.Errorf("perl: syslog facility %d out of range", facility)
	}
	w, err := syslog.New(syslog.Priority(facility<<3)|syslog.LOG_INFO, ident)
	if err != nil {
		return nil, err
	}
	if l.writer != nil {
		// The old connection is done with either way; the new one is what
		// the caller asked for.
		_ = l.writer.Close()
	}
	l.writer, l.ident, l.facility = w, ident, facility
	return w, nil
}

// Open connects to the log daemon, so that a host without one is known at
// openlog and not at the first message.
func (l *hostSyslogger) Open(ident string, options, facility int32) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := l.connectLocked(ident, facility)
	return err
}

func (l *hostSyslogger) Log(msg base.SyslogMessage) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	w, err := l.connectLocked(msg.Ident, msg.Facility)
	if err != nil {
		return err
	}
	switch msg.Severity {
	case syslogEmerg:
		return w.Emerg(msg.Message)
	case syslogAlert:
		return w.Alert(msg.Message)
	case syslogCrit:
		return w.Crit(msg.Message)
	case syslogErr:
		return w.Err(msg.Message)
	case syslogWarning:
		return w.Warning(msg.Message)
	case syslogNotice:
		return w.Notice(msg.Message)
	case syslogInfo:
		return w.Info(msg.Message)
	case syslogDebug:
		return w.Debug(msg.Message)
	}
	return fmt.Errorf("perl: syslog severity %d out of range", msg.Severity)
}

func (l *hostSyslogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.writer == nil {
		return nil
	}
	err := l.writer.Close()
	l.writer = nil
	return err
}
