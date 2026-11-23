//go:build darwin || freebsd || openbsd || netbsd || dragonfly

package internal

import "syscall"

// sysctlValue is the raw value of a sysctl variable.
func sysctlValue(name string) ([]byte, error) {
	v, err := syscall.Sysctl(name)
	return []byte(v), err
}
