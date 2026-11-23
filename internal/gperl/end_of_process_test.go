package gperl

import (
	"errors"
	"testing"
)

// A run that could not be carried out — the interpreter crashed — does not
// leave the process with the status of success.
func TestRunThatFailedIsNoSuccess(t *testing.T) {
	failure := errors.New("perl: the interpreter crashed")
	if status, err := endOfProcess(0, failure); status == 0 || err != failure {
		t.Fatalf("endOfProcess(0, failure) = (%d, %v), want a failure status and the error", status, err)
	}
	if status, err := endOfProcess(3, failure); status != 3 || err != failure {
		t.Fatalf("endOfProcess(3, failure) = (%d, %v), want (3, failure)", status, err)
	}
	if status, err := endOfProcess(0, nil); status != 0 || err != nil {
		t.Fatalf("endOfProcess(0, nil) = (%d, %v), want (0, nil)", status, err)
	}
}
