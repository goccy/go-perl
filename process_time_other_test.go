//go:build !unix

package perl_test

// processUserSeconds returns the user time the process has used so far,
// which the test does not know how to ask for on this platform.
func processUserSeconds() (seconds float64, known bool) { return 0, false }
