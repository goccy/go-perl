//go:build !darwin && !linux

package internal

import "os"

// KeepRuntimeDescriptorsHigh does nothing on this platform, where the
// guest's descriptors are not the host's.
func KeepRuntimeDescriptorsHigh() {}

// keepRuntimeHigh has nothing to do on this platform.
func keepRuntimeHigh() error { return nil }

// inheritedFiles returns none: this platform's descriptors are not listed.
func inheritedFiles(opts InstanceOptions) (map[int32]*os.File, error) { return nil, nil }
