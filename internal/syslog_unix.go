//go:build !windows && !plan9

package internal

import (
	"fmt"
	"log/syslog"
	"os"
	"strconv"
	"sync"

	"github.com/goccy/perlwasm2go/base"
)

// hostSyslogger delivers to the system logger of the host through
// log/syslog, which speaks to the local log daemon. A connection to the
// daemon carries one facility and one tag, so the logger keeps the
// connection for the facility and ident it last logged with and makes a new
// one when a message names others. It is shared by an instance and its fork
// children.
//
// It behaves as the C library's syslog(3), whose functions return nothing:
// when no log daemon takes a message, the message is lost and the program
// goes on as if it had been logged — or, with LOG_CONS, the message goes to
// the system console. The next message tries the daemon again.
type hostSyslogger struct {
	mu       sync.Mutex
	writer   *syslog.Writer
	ident    string
	facility int32
}

func newHostSyslogger() base.Syslogger { return &hostSyslogger{} }

// DialSyslog connects to the host's log daemon, and SyslogConsole is the
// console that takes a message the daemon did not (LOG_CONS). They are
// variables so that a test can stand in a host that has no log daemon.
var (
	DialSyslog    = syslog.New
	SyslogConsole = "/dev/console"
)

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
	w, err := DialSyslog(syslog.Priority(facility<<3)|syslog.LOG_INFO, ident)
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

// Open is openlog: with LOG_NDELAY it connects to the log daemon at once,
// otherwise the first message does. A daemon that cannot be reached is no
// failure of openlog(3).
func (l *hostSyslogger) Open(ident string, options, facility int32) error {
	if options&base.SyslogOptNDelay == 0 {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.connectLocked(ident, facility)
	return nil
}

func (l *hostSyslogger) Log(msg base.SyslogMessage) error {
	if msg.Severity < syslogEmerg || msg.Severity > syslogDebug {
		return fmt.Errorf("perl: syslog severity %d out of range", msg.Severity)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	w, err := l.connectLocked(msg.Ident, msg.Facility)
	if err == nil {
		// log/syslog itself retries a failed write once, on a new
		// connection.
		if err = writeSyslog(w, msg); err == nil {
			return nil
		}
		_ = w.Close()
		l.writer = nil
	}
	if msg.Options&base.SyslogOptCons != 0 {
		writeSyslogConsole(msg)
	}
	return nil
}

// writeSyslog hands msg to the daemon at its severity.
func writeSyslog(w *syslog.Writer, msg base.SyslogMessage) error {
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
	default:
		return w.Debug(msg.Message)
	}
}

// writeSyslogConsole writes a message no daemon took to the console, as
// syslog(3) does for LOG_CONS: the tag (with the process id for LOG_PID)
// and the text, ended by "\r\n". A console that cannot be opened loses the
// message.
func writeSyslogConsole(msg base.SyslogMessage) {
	f, err := os.OpenFile(SyslogConsole, os.O_WRONLY|openNoControllingTerminal, 0)
	if err != nil {
		return
	}
	defer f.Close()
	tag := msg.Ident
	if tag == "" && len(os.Args) > 0 {
		tag = os.Args[0]
	}
	if msg.Options&base.SyslogOptPID != 0 {
		tag += "[" + strconv.Itoa(os.Getpid()) + "]"
	}
	_, _ = f.WriteString(tag + ": " + msg.Message + "\r\n")
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
