//go:build !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package internal

import "errors"

// sysctlValue is the raw value of a sysctl variable; this system has none.
func sysctlValue(string) ([]byte, error) {
	return nil, errors.New("no sysctl on this system")
}
