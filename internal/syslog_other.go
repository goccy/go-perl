//go:build windows || plan9

package internal

import (
	"errors"

	"github.com/goccy/perlwasm2go/base"
)

// hostSyslogger is the system logger of a host that has none the runtime
// can reach: every call reports that, so the guest's native log mechanism
// is unavailable and nothing is dropped silently.
type hostSyslogger struct{}

func newHostSyslogger() base.Syslogger { return hostSyslogger{} }

func (hostSyslogger) Open(ident string, options, facility int32) error {
	return errors.ErrUnsupported
}

func (hostSyslogger) Log(msg base.SyslogMessage) error { return errors.ErrUnsupported }

func (hostSyslogger) Close() error { return nil }
