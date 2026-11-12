//go:build unix

package perl_test

import "syscall"

// processUserSeconds returns the user time the process has used so far.
func processUserSeconds() (seconds float64, known bool) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, false
	}
	return float64(usage.Utime.Sec) + float64(usage.Utime.Usec)/1e6, true
}
