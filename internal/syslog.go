package internal

import (
	"github.com/goccy/perlwasm2go/base"
)

// The system logger behind the guest's syslog(3).
//
// The interpreter's C library has no syslog; its openlog, syslog and
// closelog hand each message to the host as typed values (see perl-wasm's
// perl_syslog.cc). An instance configured with InstanceOptions.Syslog gets a
// base.Syslogger that delivers them to the host's system logger; any other
// instance has none, and the guest sees that its native log mechanism is
// not available.

// NewSyslogger returns the logger an instance with InstanceOptions.Syslog
// delivers to: the host's system logger. It is a variable so that a test
// can observe what a program sends without a system logger at hand.
var NewSyslogger = func() base.Syslogger { return newHostSyslogger() }

// The severities of a message, as syslog.h numbers them.
const (
	syslogEmerg = iota
	syslogAlert
	syslogCrit
	syslogErr
	syslogWarning
	syslogNotice
	syslogInfo
	syslogDebug
)
