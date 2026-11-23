//go:build !darwin && !linux

package gperl_test

import "testing"

// keepStandardDescriptors has nothing to keep where a program's
// descriptors are not the process's.
func keepStandardDescriptors(tb testing.TB) (putBack func()) { return func() {} }
